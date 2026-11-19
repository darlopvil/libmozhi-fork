package libmozhi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The whole list is tried, not just its first 3 models, and once every
// model is skipped the error says why for each one.
func TestGeminiRotation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/models/m5:"):
			w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"hola"}]}}]}`))
		case strings.Contains(r.URL.Path, "/models/m4:"):
			w.WriteHeader(404)
			w.Write([]byte(`{"error":{"message":"no longer available"}}`))
		default:
			w.WriteHeader(503)
			w.Write([]byte(`{"error":{"message":"high demand"}}`))
		}
	}))
	defer srv.Close()
	geminiState.skip, geminiState.why = nil, nil
	defer func() { geminiState.skip, geminiState.why = nil, nil }()

	prefs := "m1,m2,m3,m4,m5"
	got, err := geminiChat(srv.URL, "key", prefs, "system", "hello", 0.3)
	if err != nil || got != "hola" {
		t.Fatalf("5th model: got %q, %v", got, err)
	}
	geminiSkipModel("m5", geminiCoolTTL, "timeout")
	_, err = geminiChat(srv.URL, "key", prefs, "system", "hello", 0.3)
	want := "(m1: 503, m2: 503, m3: 503, m4: 404 retired, m5: timeout)"
	if err == nil || !strings.HasSuffix(err.Error(), want) {
		t.Errorf("all skipped: got %v, want ...%s", err, want)
	}
	if n := geminiModelCount(""); n != 5 {
		t.Errorf("default list has %d models, want 5", n)
	}
}
