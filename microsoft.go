package libmozhi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

// ============================================================================
// MICROSOFT TRANSLATOR ENGINE — isla autónoma. No comparte código con otros
// engines. Translation, romanization of non-Latin scripts and, for pairs
// with English, its bilingual dictionary with examples. Needs an Azure
// Translator resource key; the free tier (F0) allows 2 million characters a
// month and 10 requests a second, its own quota.
// ============================================================================

var microsoftHTTP = &http.Client{Timeout: 15 * time.Second}

// microsoftScripts: languages written in a non-Latin script that Microsoft
// romanizes (to Latn), with the script code its transliterate API takes.
var microsoftScripts = map[string]string{
	"ja": "Jpan", "ko": "Kore", "zh-Hans": "Hans", "zh-Hant": "Hant",
	"ru": "Cyrl", "uk": "Cyrl", "el": "Grek", "ar": "Arab", "he": "Hebr",
	"hi": "Deva", "th": "Thai",
}

// microsoftCall POSTs body to path (with query) and returns the JSON
// answer, retrying 429/5xx twice.
func microsoftCall(path string, query url.Values, body interface{}) ([]byte, error) {
	key := os.Getenv("MOZHI_MICROSOFT_API_KEY")
	if key == "" {
		return nil, errors.New("microsoft engine requires MOZHI_MICROSOFT_API_KEY")
	}
	baseURL := os.Getenv("MOZHI_MICROSOFT_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.cognitive.microsofttranslator.com"
	}
	query.Set("api-version", "3.0")
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		req, err := http.NewRequest("POST", baseURL+path+"?"+query.Encode(), bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Ocp-Apim-Subscription-Key", key)
		// Only regional resources need their region; global ones don't.
		if region := os.Getenv("MOZHI_MICROSOFT_REGION"); region != "" {
			req.Header.Set("Ocp-Apim-Subscription-Region", region)
		}
		req.Header.Set("Content-Type", "application/json")
		res, err := microsoftHTTP.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		out, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		if res.StatusCode == 200 && gjson.ValidBytes(out) {
			return out, nil
		}
		lastErr = errors.New("microsoft api error: " + res.Status + ": " + gjson.GetBytes(out, "error.message").String())
		if res.StatusCode != 429 && res.StatusCode < 500 {
			break
		}
	}
	return nil, lastErr
}

// microsoftRomanize returns text in Latin script, or "" when the language
// isn't in microsoftScripts or the call fails (romanization is optional).
func microsoftRomanize(lang, text string) string {
	script, ok := microsoftScripts[lang]
	if !ok || strings.TrimSpace(text) == "" {
		return ""
	}
	out, err := microsoftCall("/transliterate", url.Values{
		"language": {lang}, "fromScript": {script}, "toScript": {"Latn"},
	}, []map[string]string{{"Text": text}})
	if err != nil {
		return ""
	}
	return gjson.GetBytes(out, "0.text").String()
}

func translateMicrosoft(to string, from string, text string) (LangOut, error) {
	if err := validateLanguagePair(langListMicrosoft("sl"), langListMicrosoft("tl"), from, to); err != nil {
		return LangOut{}, err
	}
	query := url.Values{"to": {to}}
	if from != "auto" {
		query.Set("from", from)
	}
	// Romanization of the translation comes in the same request.
	if _, ok := microsoftScripts[to]; ok {
		query.Set("toScript", "Latn")
	}
	out, err := microsoftCall("/translate", query, []map[string]string{{"Text": text}})
	if err != nil && query.Get("toScript") != "" {
		// Romanization is optional: never lose the translation over it.
		query.Del("toScript")
		out, err = microsoftCall("/translate", query, []map[string]string{{"Text": text}})
	}
	if err != nil {
		return LangOut{}, err
	}
	translated := gjson.GetBytes(out, "0.translations.0.text").String()
	if translated == "" {
		return LangOut{}, errors.New("microsoft: empty translation")
	}

	var langout LangOut
	langout.Engine = "microsoft"
	langout.SourceLang = from
	langout.TargetLang = to
	langout.OutputText = translated
	langout.TargetTransliteration = gjson.GetBytes(out, "0.translations.0.transliteration.text").String()
	src := from
	if from == "auto" {
		src = gjson.GetBytes(out, "0.detectedLanguage.language").String()
		langout.AutoDetect = src
	}
	langout.SourceTransliteration = microsoftRomanize(src, text)

	// Dictionary: only for pairs with English, and for a word or a short
	// expression (longer texts have no dictionary entry).
	if (src == "en" || to == "en") && src != to && len(strings.Fields(text)) <= 3 {
		microsoftDictionary(&langout, src, to, text)
	}
	return langout, nil
}

// microsoftDictionary adds the dictionary's alternative translations (with
// part of speech and up to 2 examples each) and the source words they
// translate back to. Optional: on any error the translation stays as is.
func microsoftDictionary(langout *LangOut, from, to, text string) {
	pair := url.Values{"from": {from}, "to": {to}}
	out, err := microsoftCall("/dictionary/lookup", pair, []map[string]string{{"Text": text}})
	if err != nil {
		return
	}
	entry := gjson.GetBytes(out, "0")
	source := entry.Get("normalizedSource").String()
	var pairs []map[string]string
	var targets []gjson.Result
	seen := map[string]bool{strings.ToLower(text): true}
	for _, tr := range entry.Get("translations").Array() {
		if len(targets) < 5 {
			targets = append(targets, tr)
			pairs = append(pairs, map[string]string{"Text": source, "Translation": tr.Get("normalizedTarget").String()})
		}
		for _, back := range tr.Get("backTranslations").Array() {
			word := back.Get("displayText").String()
			if word != "" && !seen[strings.ToLower(word)] {
				seen[strings.ToLower(word)] = true
				langout.SourceSynonyms = append(langout.SourceSynonyms, word)
			}
		}
	}
	if len(targets) == 0 {
		return
	}
	var examples []gjson.Result
	if exOut, err := microsoftCall("/dictionary/examples", pair, pairs); err == nil {
		examples = gjson.ParseBytes(exOut).Array()
	}
	for i, tr := range targets {
		wc := WordChoices{
			Word:       tr.Get("displayTarget").String(),
			Definition: strings.ToLower(tr.Get("posTag").String()),
		}
		if i < len(examples) {
			for j, ex := range examples[i].Get("examples").Array() {
				if j == 2 {
					break
				}
				wc.ExamplesSource = append(wc.ExamplesSource, ex.Get("sourcePrefix").String()+ex.Get("sourceTerm").String()+ex.Get("sourceSuffix").String())
				wc.ExamplesTarget = append(wc.ExamplesTarget, ex.Get("targetPrefix").String()+ex.Get("targetTerm").String()+ex.Get("targetSuffix").String())
			}
		}
		langout.WordChoices = append(langout.WordChoices, wc)
	}
}
