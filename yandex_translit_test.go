package libmozhi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestYandexTranslit(t *testing.T) {
	answers := map[string]string{
		"こんにちは":      `"kon'nichiwa"`,
		"two\nlines": `"line one\r\nline two"`,
		"hola":       `"hola"`,
		"long":       "<html><title>414 Request-URI Too Large</title></html>",
		"object":     `{"code":500,"message":"error"}`,
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(answers[r.URL.Query().Get("text")]))
	}))
	defer srv.Close()
	old := yandexTranslitURL
	yandexTranslitURL = srv.URL
	defer func() { yandexTranslitURL = old }()

	for text, want := range map[string]string{
		"こんにちは":      "kon'nichiwa",
		"two\nlines": "line one\nline two",
		"hola":       "", // unchanged: nothing to show
		"long":       "", // error page, not a JSON string
		"object":     "", // JSON, but not a string
	} {
		if got := yandexTranslit("ja", text); got != want {
			t.Errorf("yandexTranslit(%q) = %q, want %q", text, got, want)
		}
	}
	if got := yandexTranslit("auto", "こんにちは"); got != "" {
		t.Errorf("auto source: got %q, want \"\"", got)
	}
}
