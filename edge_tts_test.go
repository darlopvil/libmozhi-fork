package libmozhi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/websocket"
)

// A fake Edge service: the voice list, and a WebSocket that checks the two
// requests and answers like Edge (text frames, binary audio frames with a
// 2-byte header length, then turn.end).
func TestTTSEdge(t *testing.T) {
	var gotConfig, gotSSML, gotQuery, gotOrigin string
	mux := http.NewServeMux()
	mux.HandleFunc("/voices/list", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("Sec-MS-GEC") == "" {
			http.Error(w, "no token", 403)
			return
		}
		w.Write([]byte(`[{"ShortName":"en-US-JennyNeural","Locale":"en-US"},{"ShortName":"ja-JP-NanamiNeural","Locale":"ja-JP"}]`))
	})
	mux.Handle("/edge/v1", websocket.Server{Handler: func(ws *websocket.Conn) {
		gotQuery, gotOrigin = ws.Request().URL.RawQuery, ws.Request().Header.Get("Origin")
		websocket.Message.Receive(ws, &gotConfig)
		websocket.Message.Receive(ws, &gotSSML)
		websocket.Message.Send(ws, "X-RequestId:x\r\nPath:turn.start\r\n\r\n{}")
		for _, part := range []string{"ID3", "AUDIO"} {
			h := []byte("X-RequestId:x\r\nContent-Type:audio/mpeg\r\nPath:audio\r\n")
			websocket.Message.Send(ws, append(append([]byte{byte(len(h) >> 8), byte(len(h))}, h...), part...))
		}
		end := []byte("X-RequestId:x\r\nPath:audio\r\n")
		websocket.Message.Send(ws, append([]byte{0, byte(len(end))}, end...))
		websocket.Message.Send(ws, "X-RequestId:x\r\nPath:turn.end\r\n\r\n{}")
	}, Handshake: func(*websocket.Config, *http.Request) error { return nil }})
	srv := httptest.NewTLSServer(mux)
	defer srv.Close()

	oldWS, oldVoices, oldTLS := edgeWSURL, edgeVoicesURL, edgeTLSConfig
	edgeWSURL = "wss://" + strings.TrimPrefix(srv.URL, "https://") + "/edge/v1"
	edgeVoicesURL = srv.URL + "/voices/list"
	edgeTLSConfig = srv.Client().Transport.(*http.Transport).TLSClientConfig
	oldHTTP := edgeHTTP
	edgeHTTP = srv.Client()
	edgeVoices.list = nil
	defer func() {
		edgeWSURL, edgeVoicesURL, edgeTLSConfig, edgeHTTP = oldWS, oldVoices, oldTLS, oldHTTP
		edgeVoices.list = nil
	}()

	audio, err := ttsEdge("ja", "おはよう & <こんにちは>")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(audio, []byte("ID3AUDIO")) {
		t.Errorf("audio = %q", audio)
	}
	for _, want := range []string{"TrustedClientToken=" + edgeTrustedClientToken, "Sec-MS-GEC=", "Sec-MS-GEC-Version=" + edgeSecMSGECVersion, "ConnectionId="} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("query %q lacks %q", gotQuery, want)
		}
	}
	if gotOrigin != edgeOrigin {
		t.Errorf("origin = %q", gotOrigin)
	}
	if !strings.Contains(gotConfig, "Path:speech.config") || !strings.Contains(gotConfig, "audio-24khz-48kbitrate-mono-mp3") {
		t.Errorf("speech.config = %q", gotConfig)
	}
	for _, want := range []string{"Path:ssml", "<voice name='ja-JP-NanamiNeural'>", "xml:lang='ja-JP'", "おはよう &amp; &lt;こんにちは&gt;"} {
		if !strings.Contains(gotSSML, want) {
			t.Errorf("ssml %q lacks %q", gotSSML, want)
		}
	}
	if _, err := ttsEdge("xx", "hi"); err == nil || !strings.HasSuffix(err.Error(), "unsupported language") {
		t.Errorf("unknown language: %v", err)
	}
}
