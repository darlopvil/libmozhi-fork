package libmozhi

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"golang.org/x/net/websocket"
)

// ============================================================================
// EDGE TTS — the read-aloud service of Microsoft Edge: the same neural
// voices as Azure Speech, with no key and no quota, over a WebSocket. It is
// the backup of Microsoft's voices: it reads when Azure Speech fails or has
// no key (ttsMicrosoftOrEdge), and it is its own health entry ("edge",
// kind tts). MOZHI_EDGE_ENABLED=false turns it off. Protocol as in
// rany2/edge-tts 7.3.1; the Sec-MS-GEC token is in edge_drm.go.
// ============================================================================

const (
	edgeBaseURL  = "speech.platform.bing.com/consumer/speech/synthesize/readaloud"
	edgeOrigin   = "chrome-extension://jdiccldimpdaibmpdkjnbmckianbfold"
	edgeTimeout  = 20 * time.Second
	edgeVoiceTTL = 24 * time.Hour
)

var edgeUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" +
	strings.SplitN(edgeChromiumVersion, ".", 2)[0] + ".0.0.0 Safari/537.36 Edg/" + strings.SplitN(edgeChromiumVersion, ".", 2)[0] + ".0.0.0"

// EdgeTTSEnabled reports whether Edge's voices may be used
// (MOZHI_EDGE_ENABLED, on by default).
func EdgeTTSEnabled() bool { return os.Getenv("MOZHI_EDGE_ENABLED") != "false" }

// edgeMUID is the random cookie Edge sends with every request.
func edgeMUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return strings.ToUpper(hex.EncodeToString(b))
}

var edgeVoices struct {
	sync.Mutex
	list    []gjson.Result
	fetched time.Time
}

var edgeHTTP = &http.Client{Timeout: edgeTimeout}

// The service's addresses, variables so tests can point them elsewhere.
var (
	edgeWSURL     = "wss://" + edgeBaseURL + "/edge/v1"
	edgeVoicesURL = "https://" + edgeBaseURL + "/voices/list"
	edgeTLSConfig *tls.Config
)

// edgeVoiceList returns Edge's voices, fetched at most once a day, shaped
// like Azure's list (all of them are neural) so the voice is picked the same
// way as for Azure Speech.
func edgeVoiceList() ([]gjson.Result, error) {
	v := &edgeVoices
	v.Lock()
	defer v.Unlock()
	if v.list != nil && time.Since(v.fetched) < edgeVoiceTTL {
		return v.list, nil
	}
	u := edgeVoicesURL + "?trustedclienttoken=" + edgeTrustedClientToken +
		"&Sec-MS-GEC=" + edgeToken() + "&Sec-MS-GEC-Version=" + edgeSecMSGECVersion
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", edgeUserAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cookie", "muid="+edgeMUID()+";")
	res, err := edgeHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	out, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 || !gjson.ValidBytes(out) || !gjson.ParseBytes(out).IsArray() {
		return nil, errors.New("edge tts: voice list: " + res.Status)
	}
	var list []gjson.Result
	for _, e := range gjson.ParseBytes(out).Array() {
		list = append(list, gjson.Parse(`{"VoiceType":"Neural",`+strings.TrimPrefix(strings.TrimSpace(e.Raw), "{")))
	}
	v.list, v.fetched = list, time.Now()
	return list, nil
}

// edgeFrame is one WebSocket message and its type (text or binary).
type edgeFrame struct {
	binary bool
	data   []byte
}

var edgeCodec = websocket.Codec{
	Marshal: func(v interface{}) ([]byte, byte, error) {
		return []byte(v.(string)), websocket.TextFrame, nil
	},
	Unmarshal: func(data []byte, payloadType byte, v interface{}) error {
		f := v.(*edgeFrame)
		f.binary, f.data = payloadType == websocket.BinaryFrame, append([]byte(nil), data...)
		return nil
	},
}

// edgeDate is the JavaScript-style date Edge sends in X-Timestamp.
func edgeDate() string {
	return time.Now().UTC().Format("Mon Jan 02 2006 15:04:05") + " GMT+0000 (Coordinated Universal Time)"
}

func edgeConnID() string { return strings.ReplaceAll(uuid.NewString(), "-", "") }

// ttsEdge reads text aloud with Edge's voice for lang (MP3).
func ttsEdge(lang, text string) ([]byte, error) {
	if !EdgeTTSEnabled() {
		return nil, errors.New("edge tts is disabled (MOZHI_EDGE_ENABLED=false)")
	}
	voices, err := edgeVoiceList()
	if err != nil {
		return nil, err
	}
	voice, locale, err := microsoftPickVoice(voices, lang)
	if err != nil {
		return nil, errors.New("edge tts: unsupported language")
	}
	var escaped strings.Builder
	if err := xml.EscapeText(&escaped, []byte(text)); err != nil {
		return nil, err
	}

	cfg, err := websocket.NewConfig(edgeWSURL+"?TrustedClientToken="+edgeTrustedClientToken+
		"&ConnectionId="+edgeConnID()+"&Sec-MS-GEC="+edgeToken()+"&Sec-MS-GEC-Version="+edgeSecMSGECVersion, edgeOrigin)
	if err != nil {
		return nil, err
	}
	cfg.Header = http.Header{
		"User-Agent":      {edgeUserAgent},
		"Pragma":          {"no-cache"},
		"Cache-Control":   {"no-cache"},
		"Accept-Language": {"en-US,en;q=0.9"},
		"Cookie":          {"muid=" + edgeMUID() + ";"},
	}
	cfg.Dialer = &net.Dialer{Timeout: 10 * time.Second}
	cfg.TlsConfig = edgeTLSConfig
	ws, err := websocket.DialConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("edge tts: connect: %w", err)
	}
	defer ws.Close()
	_ = ws.SetDeadline(time.Now().Add(edgeTimeout))

	date := edgeDate()
	if err := edgeCodec.Send(ws, "X-Timestamp:"+date+"\r\nContent-Type:application/json; charset=utf-8\r\nPath:speech.config\r\n\r\n"+
		`{"context":{"synthesis":{"audio":{"metadataoptions":{"sentenceBoundaryEnabled":"false","wordBoundaryEnabled":"false"},`+
		`"outputFormat":"audio-24khz-48kbitrate-mono-mp3"}}}}`+"\r\n"); err != nil {
		return nil, err
	}
	ssml := "<speak version='1.0' xmlns='http://www.w3.org/2001/10/synthesis' xml:lang='" + locale + "'><voice name='" + voice +
		"'><prosody pitch='+0Hz' rate='+0%' volume='+0%'>" + escaped.String() + "</prosody></voice></speak>"
	// "Z" after the date is not a mistake: Edge itself sends it so.
	if err := edgeCodec.Send(ws, "X-RequestId:"+edgeConnID()+"\r\nContent-Type:application/ssml+xml\r\nX-Timestamp:"+date+
		"Z\r\nPath:ssml\r\n\r\n"+ssml); err != nil {
		return nil, err
	}

	var audio bytes.Buffer
	for {
		var f edgeFrame
		if err := edgeCodec.Receive(ws, &f); err != nil {
			return nil, fmt.Errorf("edge tts: %w", err)
		}
		if !f.binary {
			if bytes.Contains(f.data, []byte("Path:turn.end")) {
				break
			}
			continue
		}
		// Binary: 2-byte header length, headers, then the audio.
		if len(f.data) < 2 {
			return nil, errors.New("edge tts: short binary message")
		}
		hl := int(f.data[0])<<8 | int(f.data[1])
		if 2+hl > len(f.data) {
			return nil, errors.New("edge tts: bad binary header length")
		}
		if bytes.Contains(f.data[2:2+hl], []byte("Path:audio")) {
			audio.Write(f.data[2+hl:])
		}
	}
	if audio.Len() == 0 {
		return nil, errors.New("edge tts: no audio received")
	}
	return audio.Bytes(), nil
}

// ttsMicrosoftOrEdge reads text with Microsoft's voices: Azure Speech when
// its key is set, and Edge's read-aloud service (the same voices, no key)
// when Azure fails or isn't set up. Each one's outcome is recorded under
// its own health entry.
func ttsMicrosoftOrEdge(lang, text string) ([]byte, error) {
	var azureErr error
	if _, _, cerr := microsoftSpeechConfig(); cerr == nil {
		start := time.Now()
		data, err := ttsMicrosoft(lang, text)
		recordHealth("microsoft", "tts", start, err)
		if err == nil || healthIgnored(err) || !EdgeTTSEnabled() {
			return data, err
		}
		azureErr = err
	} else if !EdgeTTSEnabled() {
		return nil, cerr
	}
	start := time.Now()
	data, err := ttsEdge(lang, text)
	recordHealth("edge", "tts", start, err)
	if err != nil && azureErr != nil {
		return nil, fmt.Errorf("%v; %v", azureErr, err)
	}
	return data, err
}
