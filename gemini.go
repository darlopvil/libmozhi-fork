package libmozhi

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/gjson"
)

// ============================================================================
// GEMINI — chat call and model selection for the Gemini island.
// Uses the native generateContent API: when a model is saturated, the
// OpenAI-compatible endpoint hangs with no answer, while the native one
// replies 503 at once or can be skipped after a timeout (measured nov 2026).
// ============================================================================

// MOZHI_GEMINI_MODEL is a comma-separated preference list. Default order by
// measured latency with minimal thinking: gemini-3.5-flash (~1.2s), the
// flash-lite alias (~0.6s), and the flash alias (it was saturated). The two
// *-latest aliases are maintained by Google, so the list never runs out:
// retired ids (404) are skipped. Last, two lite models for days when every
// Flash model is overloaded: on 19-nov-2026 all of them (Gemma too) answered
// 503 "high demand" or hung, while gemini-3.1-flash-lite answered in 12.5s
// and gemini-3.5-flash-lite in 24.6s (the latter only fits the 15s timeout
// when there is less load).
const geminiDefaultModels = "gemini-3.5-flash,gemini-flash-lite-latest,gemini-flash-latest,gemini-3.1-flash-lite,gemini-3.5-flash-lite"

const (
	// Saturated (503/429) or hanging model: skipped so that only one request
	// pays the timeout, then re-probed by real traffic.
	geminiCoolTTL = 15 * time.Minute
	// Retired model (404): skipped for a day.
	geminiGoneTTL = 24 * time.Hour
)

var geminiState struct {
	sync.Mutex
	skip      map[string]time.Time // model -> skipped until
	why       map[string]string    // model -> why it is skipped ("503", "timeout"...)
	noMinimal map[string]bool      // models that reject thinkingLevel "minimal"
}

// geminiPickModel returns the first preferred model not currently skipped,
// and whether to ask it for minimal thinking.
func geminiPickModel(prefs string) (string, bool, error) {
	if strings.TrimSpace(prefs) == "" {
		prefs = geminiDefaultModels
	}
	s := &geminiState
	s.Lock()
	defer s.Unlock()
	now := time.Now()
	var skipped []string
	for _, id := range strings.Split(prefs, ",") {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if until, ok := s.skip[id]; ok && now.Before(until) {
			skipped = append(skipped, id+": "+s.why[id])
			continue
		}
		return id, !s.noMinimal[id], nil
	}
	return "", false, errors.New("gemini: every configured model is saturated or retired right now (" + strings.Join(skipped, ", ") + ")")
}

// geminiModelCount is the number of models in prefs (the default list when
// empty).
func geminiModelCount(prefs string) int {
	if strings.TrimSpace(prefs) == "" {
		prefs = geminiDefaultModels
	}
	n := 0
	for _, id := range strings.Split(prefs, ",") {
		if strings.TrimSpace(id) != "" {
			n++
		}
	}
	return n
}

// geminiSkipModel skips model for d; why ("503", "timeout", "404"...) goes
// into the error once every model is skipped.
func geminiSkipModel(model string, d time.Duration, why string) {
	s := &geminiState
	s.Lock()
	defer s.Unlock()
	if s.skip == nil {
		s.skip, s.why = map[string]time.Time{}, map[string]string{}
	}
	s.skip[model] = time.Now().Add(d)
	s.why[model] = why
}

func geminiMarkNoMinimal(model string) {
	s := &geminiState
	s.Lock()
	defer s.Unlock()
	if s.noMinimal == nil {
		s.noMinimal = map[string]bool{}
	}
	s.noMinimal[model] = true
}

// geminiChat sends one system+user exchange and returns the answer text.
// prefs is MOZHI_GEMINI_MODEL; on saturation, timeout or retirement it moves
// on to the next model of the list.
func geminiChat(baseURL, key, prefs, systemPrompt, userText string, temperature float64) (string, error) {
	var lastErr error
	// One attempt per model of the list, plus one retry without the
	// thinking setting for models that reject it.
	for attempt := 0; attempt <= geminiModelCount(prefs); attempt++ {
		model, minimal, err := geminiPickModel(prefs)
		if err != nil {
			if lastErr != nil {
				return "", lastErr
			}
			return "", err
		}
		gen := map[string]interface{}{"temperature": temperature}
		if minimal {
			gen["thinkingConfig"] = map[string]string{"thinkingLevel": "minimal"}
		}
		payload, err := json.Marshal(map[string]interface{}{
			"systemInstruction": map[string]interface{}{
				"parts": []map[string]string{{"text": systemPrompt}},
			},
			"contents": []map[string]interface{}{
				{"role": "user", "parts": []map[string]string{{"text": userText}}},
			},
			"generationConfig": gen,
		})
		if err != nil {
			return "", err
		}
		req, err := http.NewRequest("POST", baseURL+"/models/"+model+":generateContent", strings.NewReader(string(payload)))
		if err != nil {
			return "", err
		}
		req.Header.Set("x-goog-api-key", key)
		req.Header.Set("Content-Type", "application/json")

		res, err := geminiHTTP.Do(req)
		if err != nil {
			lastErr = err
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				geminiSkipModel(model, geminiCoolTTL, "timeout") // hanging = saturated
			}
			continue
		}
		out, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		msg := gjson.GetBytes(out, "error.message").String()
		apiErr := errors.New("gemini api error: " + res.Status + " (" + model + "): " + msg)

		switch {
		case res.StatusCode == 200 && gjson.ValidBytes(out):
			var text strings.Builder
			for _, p := range gjson.GetBytes(out, "candidates.0.content.parts").Array() {
				if !p.Get("thought").Bool() {
					text.WriteString(p.Get("text").String())
				}
			}
			return text.String(), nil
		case res.StatusCode == 503 || res.StatusCode == 429 || res.StatusCode == 500:
			geminiSkipModel(model, geminiCoolTTL, strconv.Itoa(res.StatusCode))
		case res.StatusCode == 404:
			geminiSkipModel(model, geminiGoneTTL, "404 retired")
		case res.StatusCode == 400 && minimal && strings.Contains(msg, "hinking level"):
			geminiMarkNoMinimal(model)
		default:
			return "", apiErr
		}
		lastErr = apiErr
	}
	return "", lastErr
}
