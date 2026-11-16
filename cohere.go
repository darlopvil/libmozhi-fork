package libmozhi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/gjson"
)

// ============================================================================
// COHERE ENGINE — isla autónoma. No comparte código con otros engines.
// Translation with Cohere's translation models, dictionary/paraphrases with
// a general model. A trial key allows 1,000 calls a month and 20 a minute
// (x-endpoint-monthly-call-limit, x-trial-endpoint-call-limit), its own
// quota: not a pool shared with other users.
// ============================================================================

const (
	cohereDefaultModels     = "north-small-translate-1.0,command-a-translate-08-2025"
	cohereDefaultDictModels = "command-a-plus-05-2026,command-a-03-2025"
	cohereGoneTTL           = 24 * time.Hour
)

var cohereHTTP = &http.Client{Timeout: 15 * time.Second}

// cohereGone: models Cohere answered 404 for ("model '...' not found"),
// skipped until the time stored.
var cohereGone struct {
	sync.Mutex
	until map[string]time.Time
}

func cohereSkip(model string) bool {
	cohereGone.Lock()
	defer cohereGone.Unlock()
	until, ok := cohereGone.until[model]
	return ok && time.Now().Before(until)
}

func cohereMarkGone(model string) {
	cohereGone.Lock()
	defer cohereGone.Unlock()
	if cohereGone.until == nil {
		cohereGone.until = map[string]time.Time{}
	}
	cohereGone.until[model] = time.Now().Add(cohereGoneTTL)
}

// cohereChat sends messages to the first model of prefs (comma-separated)
// that isn't known to be gone, moving to the next one on a 404, and retrying
// 429/5xx answers twice.
func cohereChat(key, prefs string, messages []map[string]string, jsonOut bool) (string, error) {
	baseURL := os.Getenv("MOZHI_COHERE_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.cohere.com/v2"
	}
	var models []string
	for _, m := range strings.Split(prefs, ",") {
		if m = strings.TrimSpace(m); m != "" {
			models = append(models, m)
		}
	}
	var lastErr error
	retries := 0
	for i := 0; i < len(models); {
		model := models[i]
		if cohereSkip(model) {
			i++
			continue
		}
		body := map[string]interface{}{"model": model, "messages": messages, "temperature": 0.3}
		if jsonOut {
			body["response_format"] = map[string]string{"type": "json_object"}
		}
		payload, err := json.Marshal(body)
		if err != nil {
			return "", err
		}
		req, err := http.NewRequest("POST", baseURL+"/chat", strings.NewReader(string(payload)))
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", "Bearer "+key)
		req.Header.Set("Content-Type", "application/json")
		res, err := cohereHTTP.Do(req)
		if err != nil {
			return "", err
		}
		out, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return "", err
		}
		if res.StatusCode == 200 {
			// Reasoning models (Command A Plus) send a "thinking" block before
			// the answer: take the first "text" block.
			text := gjson.GetBytes(out, `message.content.#(type=="text").text`).String()
			if strings.TrimSpace(text) == "" {
				return "", errors.New("cohere: empty answer")
			}
			return text, nil
		}
		lastErr = errors.New("cohere api error: " + res.Status + ": " + gjson.GetBytes(out, "message").String())
		switch {
		case res.StatusCode == 404:
			cohereMarkGone(model) // retired or unknown model: next one
			i++
		case (res.StatusCode == 429 || res.StatusCode >= 500) && retries < 2:
			retries++
			time.Sleep(time.Duration(retries) * time.Second)
		default:
			return "", lastErr
		}
	}
	if lastErr == nil {
		lastErr = errors.New("cohere: none of the configured models is available (" + prefs + ")")
	}
	return "", lastErr
}

func cohereLangName(list []List, id string) string {
	for _, l := range list {
		if l.Id == id {
			return l.Name
		}
	}
	return id
}

func translateCohere(to string, from string, text string) (LangOut, error) {
	if err := validateLanguagePair(langListCohere("sl"), langListCohere("tl"), from, to); err != nil {
		return LangOut{}, err
	}
	key := os.Getenv("MOZHI_COHERE_API_KEY")
	if key == "" {
		return LangOut{}, errors.New("cohere engine requires MOZHI_COHERE_API_KEY")
	}
	models := os.Getenv("MOZHI_COHERE_MODEL")
	if strings.TrimSpace(models) == "" {
		models = cohereDefaultModels
	}
	dictModels := os.Getenv("MOZHI_COHERE_DICT_MODEL")
	if strings.TrimSpace(dictModels) == "" {
		dictModels = cohereDefaultDictModels
	}

	targetName := cohereLangName(langListCohere("tl"), to)
	srcName := "the source language"
	instruction := "Translate the following text into " + targetName + ". Output ONLY the translation.\n\n"
	if from != "auto" {
		srcName = cohereLangName(langListCohere("sl"), from)
		instruction = "Translate the following text from " + srcName + " into " + targetName + ". Output ONLY the translation.\n\n"
	}

	// 1) Translation: the translation models take a single user message
	// (Cohere's documented format).
	translated, err := cohereChat(key, models, []map[string]string{
		{"role": "user", "content": instruction + text},
	}, false)
	if err != nil {
		return LangOut{}, err
	}

	var langout LangOut
	langout.Engine = "cohere"
	langout.SourceLang = from
	langout.TargetLang = to
	langout.OutputText = strings.TrimSpace(translated)

	// 2) Dictionary / paraphrases (short text), with a general model.
	if len(strings.Fields(text)) <= 8 {
		dictSystem := "You are a bilingual lexical assistant. Translating from " + srcName + " into " + targetName + ". Analyze the user text and return ONLY a JSON object (no markdown, no preamble) with this exact shape: {\"senses\":[{\"word\":\"<alt translation / sense / paraphrase>\",\"examples\":[{\"src\":\"<example in source lang>\",\"dst\":\"<its target translation>\"}]}],\"synonyms\":[\"<synonyms of the main translation, in target lang>\"],\"antonyms\":[\"<antonyms in target lang>\"]}. Rules: if the text is a single WORD or very short phrase, give its different SENSES (up to 5) with example sentences each, plus synonyms and antonyms. If it is a full SENTENCE, give up to 4 natural PARAPHRASES as senses (word = the paraphrase), with the paraphrase as src and its translation as dst; synonyms/antonyms may be empty. Always output valid JSON only."
		raw, derr := cohereChat(key, dictModels, []map[string]string{
			{"role": "system", "content": dictSystem},
			{"role": "user", "content": text},
		}, true)
		if derr == nil && gjson.Valid(raw) {
			parsed := gjson.Parse(raw)
			for _, sense := range parsed.Get("senses").Array() {
				var wc WordChoices
				wc.Word = sense.Get("word").String()
				for _, ex := range sense.Get("examples").Array() {
					src := ex.Get("src").String()
					dst := ex.Get("dst").String()
					if src == "" && dst == "" {
						continue
					}
					wc.ExamplesSource = append(wc.ExamplesSource, src)
					wc.ExamplesTarget = append(wc.ExamplesTarget, dst)
				}
				if wc.Word != "" {
					langout.WordChoices = append(langout.WordChoices, wc)
				}
			}
			for _, s := range parsed.Get("synonyms").Array() {
				if s.String() != "" {
					langout.TargetSynonyms = append(langout.TargetSynonyms, s.String())
				}
			}
			for _, a := range parsed.Get("antonyms").Array() {
				if a.String() != "" {
					langout.TargetAntonyms = append(langout.TargetAntonyms, a.String())
				}
			}
		}
	}
	return langout, nil
}
