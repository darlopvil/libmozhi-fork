package libmozhi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// When TexTra gives no token, the error says why: an outage (503) and a
// rejected key (401) need different fixes.
func TestTexTraTokenError(t *testing.T) {
	for _, c := range []struct {
		status int
		body   string
		want   string
	}{
		{503, "<html>Service Unavailable</html>", "textra: no access token (token endpoint: 503 Service Unavailable)"},
		{503, "<html><h1>ただいまメンテナンス中です。</h1></html>", "textra: under maintenance (token endpoint: 503 Service Unavailable)"},
		{401, `{"error":"invalid_client","error_description":"The client credentials are invalid"}`, "(token endpoint: 401 Unauthorized, The client credentials are invalid)"},
		{400, `{"error":"invalid_request"}`, "(token endpoint: 400 Bad Request, invalid_request)"},
	} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(c.status)
			w.Write([]byte(c.body))
		}))
		textraTokenState.token = ""
		_, err := textraToken(srv.URL, "key", "secret")
		srv.Close()
		if err == nil || !strings.HasSuffix(err.Error(), c.want) {
			t.Errorf("HTTP %d: got %v, want ...%s", c.status, err, c.want)
		}
	}
}
