package libmozhi

import (
	"errors"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// ============================================================================
// ENGINE HEALTH — what real traffic says about each engine.
// Every translation and TTS request is recorded as it happens (Translate, TTS
// and TTSStream). Nothing is sent to an engine just to check it while it
// works: several have small quotas (TexTra about 3 requests per 2 minutes for
// the whole account, Groq TTS 100 a day). Only an engine whose last request
// failed is re-checked, at most once per healthProbeGap, so its state turns
// back to ok as soon as it recovers.
// ============================================================================

// healthProbeGap is one TexTra window: a re-check never takes more than one
// of TexTra's 3 slots per window, however often the health is asked for.
const healthProbeGap = textraWindow

// healthKinds lists the engines recorded for each kind of request.
var healthKinds = map[string][]string{
	"translate": {"google", "deepl", "duckduckgo", "mymemory", "yandex", "gemini", "textra", "groq", "gptoss", "gemma", "nemotron", "mistral"},
	"tts":       {"google", "textra", "groq"},
}

// Health is the state of one engine for one kind of request, since start.
type Health struct {
	Status              string     `json:"status"` // ok, failing, or unknown (no request yet)
	LastOK              *time.Time `json:"last_ok,omitempty"`
	LastFailure         *time.Time `json:"last_failure,omitempty"`
	LastError           string     `json:"last_error,omitempty"`
	LatencyMs           int64      `json:"latency_ms,omitempty"` // of the last successful request
	ConsecutiveFailures int        `json:"consecutive_failures"`
	Requests            int64      `json:"requests"`
	Failures            int64      `json:"failures"`
	SlotsFree           *int       `json:"slots_free,omitempty"` // TexTra only
	lastProbe           time.Time
}

var healthState struct {
	sync.Mutex
	m map[string]*Health
}

func healthKnown(engine, kind string) bool {
	for _, e := range healthKinds[kind] {
		if e == engine {
			return true
		}
	}
	return false
}

// HealthKinds returns the kinds of request recorded for engine
// ("translate", "tts"), or none for an unknown engine.
func HealthKinds(engine string) []string {
	kinds := []string{}
	for _, kind := range []string{"translate", "tts"} {
		if healthKnown(engine, kind) {
			kinds = append(kinds, kind)
		}
	}
	return kinds
}

// healthIgnored: errors caused by the request, not by the engine.
func healthIgnored(err error) bool {
	if err == nil {
		return false
	}
	var limit textraLimitError
	return errors.Is(err, errSourceLang) || errors.Is(err, errTargetLang) ||
		errors.As(err, &limit) || strings.HasSuffix(err.Error(), "unsupported language")
}

func recordHealth(engine, kind string, start time.Time, err error) {
	if !healthKnown(engine, kind) || healthIgnored(err) {
		return
	}
	now := time.Now()
	s := &healthState
	s.Lock()
	defer s.Unlock()
	if s.m == nil {
		s.m = map[string]*Health{}
	}
	h := s.m[engine+"/"+kind]
	if h == nil {
		h = &Health{}
		s.m[engine+"/"+kind] = h
	}
	h.Requests++
	if err == nil {
		h.Status = "ok"
		h.LastOK = &now
		h.LatencyMs = now.Sub(start).Milliseconds()
		h.ConsecutiveFailures = 0
		return
	}
	h.Status = "failing"
	h.LastFailure = &now
	h.LastError = healthScrub(err.Error())
	h.ConsecutiveFailures++
	h.Failures++
}

var (
	healthSecretEnv = regexp.MustCompile(`^MOZHI_\w*(KEY|SECRET|TOKEN|LOGIN|EMAIL|PASSWORD)\w*=`)
	healthUserID    = regexp.MustCompile(`\buser_[A-Za-z0-9]{8,}`)
	healthIP        = regexp.MustCompile(`\b\d{1,3}(\.\d{1,3}){3}(:\d+)?\b`)
)

// healthScrub keeps credentials, account ids and addresses out of error
// messages: the health endpoint is public.
func healthScrub(msg string) string {
	for _, kv := range os.Environ() {
		if healthSecretEnv.MatchString(kv) {
			if v := kv[strings.IndexByte(kv, '=')+1:]; len(v) >= 4 {
				msg = strings.ReplaceAll(msg, v, "[redacted]")
			}
		}
	}
	msg = healthUserID.ReplaceAllString(msg, "user_[redacted]")
	msg = healthIP.ReplaceAllString(msg, "[ip]")
	if utf8.RuneCountInString(msg) > 300 {
		msg = string([]rune(msg)[:300]) + "…"
	}
	return msg
}

// HealthOf returns the recorded health of engine for kind.
func HealthOf(engine, kind string) Health {
	s := &healthState
	s.Lock()
	h := Health{Status: "unknown"}
	if p := s.m[engine+"/"+kind]; p != nil {
		h = *p
	}
	s.Unlock()
	if engine == "textra" {
		free := textraFree()
		h.SlotsFree = &free
	}
	return h
}

// EngineHealth returns the recorded health of every kind engine supports.
func EngineHealth(engine string) map[string]Health {
	out := map[string]Health{}
	for _, kind := range HealthKinds(engine) {
		out[kind] = HealthOf(engine, kind)
	}
	return out
}

// HealthCheck returns the health of engine for kind, re-checking it first
// with a short request when its last request failed and it wasn't re-checked
// in the last healthProbeGap. ok is false for an unknown engine or kind.
func HealthCheck(engine, kind string) (h Health, ok bool) {
	if !healthKnown(engine, kind) {
		return Health{}, false
	}
	s := &healthState
	s.Lock()
	probe := false
	if p := s.m[engine+"/"+kind]; p != nil && p.Status == "failing" && time.Since(p.lastProbe) >= healthProbeGap {
		p.lastProbe = time.Now()
		probe = true
	}
	s.Unlock()
	// TexTra: leave at least one slot to real requests.
	if probe && (engine != "textra" || textraFree() >= 2) {
		if kind == "tts" {
			_, _ = TTS(engine, "en", "Good morning")
		} else {
			_, _ = Translate(engine, "es", "en", "Good morning")
		}
	}
	return HealthOf(engine, kind), true
}
