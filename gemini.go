package libmozhi

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
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
// flash-lite alias (~0.6s), and the flash alias last (it was saturated).
// The two *-latest aliases are maintained by Google, so the list never runs
// out: retired ids (404) are skipped.
const geminiDefaultModels = "gemini-3.5-flash,gemini-flash-lite-latest,gemini-flash-latest"

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
	for _, id := range strings.Split(prefs, ",") {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if until, ok := s.skip[id]; ok && now.Before(until) {
			continue
		}
		return id, !s.noMinimal[id], nil
	}
	return "", false, errors.New("gemini: every configured model is saturated or retired right now (" + prefs + ")")
}

func geminiSkipModel(model string, d time.Duration) {
	s := &geminiState
	s.Lock()
	defer s.Unlock()
	if s.skip == nil {
		s.skip = map[string]time.Time{}
	}
	s.skip[model] = time.Now().Add(d)
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
	// One attempt per model of a 3-model list, plus one retry without the
	// thinking setting for models that reject it.
	for attempt := 0; attempt < 4; attempt++ {
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
				geminiSkipModel(model, geminiCoolTTL) // hanging = saturated
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
			geminiSkipModel(model, geminiCoolTTL)
		case res.StatusCode == 404:
			geminiSkipModel(model, geminiGoneTTL)
		case res.StatusCode == 400 && minimal && strings.Contains(msg, "hinking level"):
			geminiMarkNoMinimal(model)
		default:
			return "", apiErr
		}
		lastErr = apiErr
	}
	return "", lastErr
}
