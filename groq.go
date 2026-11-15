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
// GROQ ENGINE — isla autónoma. No comparte código con otros engines.
// Traducción + diccionario/paráfrasis (chat) + TTS (en/ar).
// ============================================================================

// groqChat: llamada al endpoint chat OpenAI-compatible de Groq, con reintentos 503/429.
func groqChat(baseURL, key, prefs, systemPrompt, userText string, temperature float64) (string, error) {
	var lastErr error
	refreshed := false
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		model, err := groqPickModel(baseURL, key, prefs)
		if err != nil {
			return "", err
		}
		payload, err := json.Marshal(map[string]interface{}{
			"model": model,
			"messages": []map[string]string{
				{"role": "system", "content": systemPrompt},
				{"role": "user", "content": userText},
			},
			"temperature": temperature,
		})
		if err != nil {
			return "", err
		}
		req, err := http.NewRequest("POST", baseURL+"/chat/completions", strings.NewReader(string(payload)))
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", "Bearer "+key)
		req.Header.Set("Content-Type", "application/json")

		res, err := groqHTTP.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		out, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if res.StatusCode == 503 || res.StatusCode == 429 {
			lastErr = errors.New("groq api error: " + res.Status)
			continue
		}
		if res.StatusCode != 200 || !gjson.ValidBytes(out) {
			apiErr := errors.New("groq api error: " + res.Status)
			// Decommissioned or unknown model: retry once with the next one.
			if !refreshed && groqModelGone(model, res.StatusCode, out) {
				refreshed = true
				lastErr = apiErr
				continue
			}
			return "", apiErr
		}
		return gjson.GetBytes(out, "choices.0.message.content").String(), nil
	}
	return "", lastErr
}

// groqStripJSONFences: quita ```json ... ``` si el modelo envolvió la respuesta.
func groqStripJSONFences(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func translateGroq(to string, from string, text string) (LangOut, error) {
	FromOrig := from
	ToOrig := to
	if err := validateLanguagePair(langListGroq("sl"), langListGroq("tl"), from, to); err != nil {
		return LangOut{}, err
	}

	var langout LangOut
	langout.Engine = "groq"
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig

	key := os.Getenv("MOZHI_GROQ_API_KEY")
	if key == "" {
		return LangOut{}, errors.New("groq engine requires MOZHI_GROQ_API_KEY")
	}
	baseURL := os.Getenv("MOZHI_GROQ_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.groq.com/openai/v1"
	}
	// Comma-separated preference list; the first model Groq still lists as
	// active is used (see groqPickModel).
	model := os.Getenv("MOZHI_GROQ_MODEL")

	targetName := to
	for _, l := range langListGroq("tl") {
		if l.Id == to {
			targetName = l.Name
			break
		}
	}
	srcName := "the source language"
	if from != "auto" {
		for _, l := range langListGroq("sl") {
			if l.Id == from {
				srcName = l.Name
				break
			}
		}
	}

	// 1) Traducción.
	transSystem := "You are a professional translator. Translate the user's text from " + srcName + " into " + targetName + ". Produce a faithful and natural translation: preserve the original meaning, tone and register, do not add, omit or explain anything. Output ONLY the translated text, with no quotes, no notes and no preamble."
	translated, err := groqChat(baseURL, key, model, transSystem, text, 0.3)
	if err != nil {
		return LangOut{}, err
	}
	langout.OutputText = strings.TrimSpace(translated)

	// 2) Diccionario / paráfrasis (texto corto).
	if len(strings.Fields(text)) <= 8 {
		dictSystem := "You are a bilingual lexical assistant. Translating from " + srcName + " into " + targetName + ". Analyze the user text and return ONLY a JSON object (no markdown, no preamble) with this exact shape: {\"senses\":[{\"word\":\"<alt translation / sense / paraphrase>\",\"examples\":[{\"src\":\"<example in source lang>\",\"dst\":\"<its target translation>\"}]}],\"synonyms\":[\"<synonyms of the main translation, in target lang>\"],\"antonyms\":[\"<antonyms in target lang>\"]}. Rules: if the text is a single WORD or very short phrase, give its different SENSES (up to 5) with example sentences each, plus synonyms and antonyms. If it is a full SENTENCE, give up to 4 natural PARAPHRASES as senses (word = the paraphrase), with the paraphrase as src and its translation as dst; synonyms/antonyms may be empty. Always output valid JSON only."
		if raw, derr := groqChat(baseURL, key, model, dictSystem, text, 0.3); derr == nil {
			raw = groqStripJSONFences(raw)
			if gjson.Valid(raw) {
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
	}

	return langout, nil
}

// ttsGroq: síntesis de voz. Groq TTS (Orpheus) sólo soporta inglés y árabe.
func ttsGroq(lang string, text string) ([]byte, error) {
	models := map[string]string{
		"en": "canopylabs/orpheus-v1-english",
		"ar": "canopylabs/orpheus-arabic-saudi",
	}
	model, ok := models[lang]
	if !ok {
		return []byte(""), errors.New("groq tts: unsupported language")
	}

	key := os.Getenv("MOZHI_GROQ_API_KEY")
	if key == "" {
		return []byte(""), errors.New("groq tts requires MOZHI_GROQ_API_KEY")
	}
	baseURL := os.Getenv("MOZHI_GROQ_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.groq.com/openai/v1"
	}
	voice := os.Getenv("MOZHI_GROQ_TTS_VOICE")
	if voice == "" {
		voice = "diana" // autumn diana hannah austin daniel troy
	}

	payloadMap := map[string]interface{}{
		"model":           model,
		"input":           text,
		"voice":           voice,
		"response_format": "wav",
	}
	payload, err := json.Marshal(payloadMap)
	if err != nil {
		return []byte(""), err
	}

	req, err := http.NewRequest("POST", baseURL+"/audio/speech", strings.NewReader(string(payload)))
	if err != nil {
		return []byte(""), err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	res, err := groqHTTP.Do(req)
	if err != nil {
		return []byte(""), err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return []byte(""), err
	}
	if len(data) < 4 || string(data[0:4]) != "RIFF" {
		return []byte(""), errors.New("groq tts: no audio returned")
	}
	return data, nil
}
// groqHTTP: own HTTP client with a timeout (http.DefaultClient never gives up).
var groqHTTP = &http.Client{Timeout: 15 * time.Second}

// ---------------------------------------------------------------------------
// Groq model selection. MOZHI_GROQ_MODEL is a comma-separated preference
// list; the first id Groq still lists as active is used, so a decommissioned
// model falls through to the next one without touching the .env.
// ---------------------------------------------------------------------------

// Both currently active on Groq's free tier. gpt-oss-safeguard is left out
// on purpose: it is a safety classifier, not a chat model.
const groqDefaultModels = "openai/gpt-oss-20b,openai/gpt-oss-120b"

const (
	groqCatalogTTL = 24 * time.Hour
	groqRefreshGap = time.Minute
	groqGoneTTL    = 24 * time.Hour
)

var groqCatalog struct {
	sync.Mutex
	active  map[string]bool
	fetched time.Time
	tried   time.Time
	gone    map[string]time.Time // id -> skipped until
}

func groqFetchModels(baseURL, key string) (map[string]bool, error) {
	req, err := http.NewRequest("GET", baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	res, err := groqHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 || !gjson.ValidBytes(body) {
		return nil, errors.New("groq models list error: " + res.Status)
	}
	active := map[string]bool{}
	for _, m := range gjson.GetBytes(body, "data").Array() {
		if a := m.Get("active"); a.Exists() && !a.Bool() {
			continue
		}
		active[m.Get("id").String()] = true
	}
	return active, nil
}

// groqPickModel returns the first preferred model that is currently usable.
// If Groq's list can't be fetched, preferences are tried in order as is.
func groqPickModel(baseURL, key, prefs string) (string, error) {
	if strings.TrimSpace(prefs) == "" {
		prefs = groqDefaultModels
	}
	c := &groqCatalog
	c.Lock()
	defer c.Unlock()

	now := time.Now()
	stale := c.active == nil || now.Sub(c.fetched) > groqCatalogTTL
	if stale && now.Sub(c.tried) > groqRefreshGap {
		c.tried = now
		if active, err := groqFetchModels(baseURL, key); err == nil {
			c.active, c.fetched = active, now
		}
	}
	for _, id := range strings.Split(prefs, ",") {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if until, ok := c.gone[id]; ok && now.Before(until) {
			continue
		}
		if c.active == nil || c.active[id] {
			return id, nil
		}
	}
	return "", errors.New("groq: none of the configured models is available (" + prefs + ")")
}

// groqModelGone reports whether Groq rejected the model itself (decommissioned
// or unknown). If so the model is skipped and Groq's list is refreshed.
func groqModelGone(model string, status int, body []byte) bool {
	code := gjson.GetBytes(body, "error.code").String()
	if status != 404 && code != "model_decommissioned" && code != "model_not_found" {
		return false
	}
	c := &groqCatalog
	c.Lock()
	defer c.Unlock()
	if c.gone == nil {
		c.gone = map[string]time.Time{}
	}
	c.gone[model] = time.Now().Add(groqGoneTTL)
	c.fetched = time.Time{}
	return true
}
