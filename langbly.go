package libmozhi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/tidwall/gjson"
)

// ============================================================================
// LANGBLY ENGINE — isla autónoma. No comparte código con otros engines.
// Official API in Google Translate v2 format, on its EU endpoint. The free
// allowance is 500,000 characters a month, its own quota. Its server gives up
// after 15 s, so the client waits a little longer.
// ============================================================================

var langblyHTTP = &http.Client{Timeout: 20 * time.Second}

func translateLangbly(to string, from string, text string) (LangOut, error) {
	if err := validateLanguagePair(langListLangbly("sl"), langListLangbly("tl"), from, to); err != nil {
		return LangOut{}, err
	}
	key := os.Getenv("MOZHI_LANGBLY_API_KEY")
	if key == "" {
		return LangOut{}, errors.New("langbly engine requires MOZHI_LANGBLY_API_KEY")
	}
	baseURL := os.Getenv("MOZHI_LANGBLY_BASE_URL")
	if baseURL == "" {
		baseURL = "https://eu.langbly.com"
	}
	body := map[string]interface{}{"q": []string{text}, "target": to, "format": "text"}
	if from != "auto" {
		body["source"] = from
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return LangOut{}, err
	}
	var out []byte
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		req, err := http.NewRequest("POST", baseURL+"/language/translate/v2", bytes.NewReader(payload))
		if err != nil {
			return LangOut{}, err
		}
		req.Header.Set("Authorization", "Bearer "+key)
		req.Header.Set("Content-Type", "application/json")
		res, err := langblyHTTP.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		data, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		if res.StatusCode == 200 && gjson.ValidBytes(data) {
			out, lastErr = data, nil
			break
		}
		lastErr = errors.New("langbly api error: " + res.Status + ": " + gjson.GetBytes(data, "error.message").String())
		if res.StatusCode != 429 && res.StatusCode < 500 {
			break
		}
	}
	if lastErr != nil {
		return LangOut{}, lastErr
	}
	translated := gjson.GetBytes(out, "data.translations.0.translatedText").String()
	if translated == "" {
		return LangOut{}, errors.New("langbly: empty translation")
	}
	var langout LangOut
	langout.Engine = "langbly"
	langout.SourceLang = from
	langout.TargetLang = to
	langout.OutputText = translated
	if from == "auto" {
		langout.AutoDetect = gjson.GetBytes(out, "data.translations.0.detectedSourceLanguage").String()
	}
	return langout, nil
}
