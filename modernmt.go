package libmozhi

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/tidwall/gjson"
)

// ============================================================================
// MODERNMT ENGINE — isla autónoma. No comparte código con otros engines.
// ModernMT, through the endpoint its own website uses (no key; unofficial,
// so it can change without notice: /api/health shows it). The website signs
// each request with md5("webkey_…#<timestamp>#<text>").
// ============================================================================

var modernmtHTTP = &http.Client{Timeout: 15 * time.Second}

const modernmtWebKey = "webkey_E3sTuMjpP8Jez49GcYpDVH7r"

func translateModernMT(to string, from string, text string) (LangOut, error) {
	if err := validateLanguagePair(langListModernMT("sl"), langListModernMT("tl"), from, to); err != nil {
		return LangOut{}, err
	}
	source := from
	if from == "auto" {
		source = ""
	}
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	sum := md5.Sum([]byte(modernmtWebKey + "#" + ts + "#" + text))
	payload, err := json.Marshal(map[string]interface{}{
		"q": text, "source": source, "target": to, "ts": json.Number(ts),
		"verify": hex.EncodeToString(sum[:]), "hints": "", "multiline": "true",
	})
	if err != nil {
		return LangOut{}, err
	}
	req, err := http.NewRequest("POST", "https://webapi.modernmt.com/translate", bytes.NewReader(payload))
	if err != nil {
		return LangOut{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://www.modernmt.com")
	req.Header.Set("Referer", "https://www.modernmt.com/translate")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("X-HTTP-Method-Override", "GET")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:140.0) Gecko/20100101 Firefox/140.0")
	res, err := modernmtHTTP.Do(req)
	if err != nil {
		return LangOut{}, err
	}
	defer res.Body.Close()
	out, err := io.ReadAll(res.Body)
	if err != nil {
		return LangOut{}, err
	}
	translated := gjson.GetBytes(out, "data.translation").String()
	if res.StatusCode != 200 || translated == "" {
		return LangOut{}, errors.New("modernmt error: " + res.Status + " " + gjson.GetBytes(out, "error.message").String())
	}
	var langout LangOut
	langout.Engine = "modernmt"
	langout.SourceLang = from
	langout.TargetLang = to
	langout.OutputText = translated
	if from == "auto" {
		langout.AutoDetect = gjson.GetBytes(out, "data.detectedLanguage").String()
	}
	return langout, nil
}
