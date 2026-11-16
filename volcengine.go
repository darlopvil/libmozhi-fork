package libmozhi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
)

// ============================================================================
// VOLCENGINE ENGINE — isla autónoma. No comparte código con otros engines.
// ByteDance's translator, through the endpoint its browser extension uses
// (no key; unofficial, so it can change without notice: /api/health shows
// it).
// ============================================================================

var volcengineHTTP = &http.Client{Timeout: 15 * time.Second}

func translateVolcengine(to string, from string, text string) (LangOut, error) {
	if err := validateLanguagePair(langListVolcengine("sl"), langListVolcengine("tl"), from, to); err != nil {
		return LangOut{}, err
	}
	body := map[string]interface{}{
		"text": text, "target_language": to,
		"enable_user_glossary": false, "glossary_list": []string{}, "category": "",
	}
	if from != "auto" {
		body["source_language"] = from
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return LangOut{}, err
	}
	req, err := http.NewRequest("POST", "https://translate.volcengine.com/crx/translate/v1/", bytes.NewReader(payload))
	if err != nil {
		return LangOut{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "chrome-extension://klgfhbiooeogdfodpopgppeadghjjemk")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")
	res, err := volcengineHTTP.Do(req)
	if err != nil {
		return LangOut{}, err
	}
	defer res.Body.Close()
	out, err := io.ReadAll(res.Body)
	if err != nil {
		return LangOut{}, err
	}
	translated := gjson.GetBytes(out, "translation").String()
	if res.StatusCode != 200 || translated == "" {
		return LangOut{}, errors.New("volcengine error: " + res.Status + " " + gjson.GetBytes(out, "message").String())
	}
	var langout LangOut
	langout.Engine = "volcengine"
	langout.SourceLang = from
	langout.TargetLang = to
	langout.OutputText = translated
	if from == "auto" {
		langout.AutoDetect = gjson.GetBytes(out, "detected_language").String()
	}
	return langout, nil
}
