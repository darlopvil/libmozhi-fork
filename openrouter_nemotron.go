package libmozhi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

// ============================================================================
// OPENROUTER Nemotron — isla autónoma. Modelo fijo: nvidia/nemotron-3-super-120b-a12b:free
// ============================================================================

func openrouterNemotronChat(key, model, systemPrompt, userText string, temperature float64) (string, error) {
	baseURL := os.Getenv("MOZHI_OPENROUTER_BASE_URL")
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	payloadMap := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userText},
		},
		"temperature": temperature,
	}
	payload, err := json.Marshal(payloadMap)
	if err != nil {
		return "", err
	}
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		req, err := http.NewRequest("POST", baseURL+"/chat/completions", strings.NewReader(string(payload)))
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", "Bearer "+key)
		req.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
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
			lastErr = errors.New("openrouter api error: " + res.Status)
			continue
		}
		if res.StatusCode != 200 || !gjson.ValidBytes(out) {
			return "", errors.New("openrouter api error: " + res.Status)
		}
		return gjson.GetBytes(out, "choices.0.message.content").String(), nil
	}
	return "", lastErr
}

func openrouterNemotronStripFences(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func translateOpenrouterNemotron(to string, from string, text string) (LangOut, error) {
	FromOrig := from
	ToOrig := to
	if err := validateLanguagePair(langListOpenrouterNemotron("sl"), langListOpenrouterNemotron("tl"), from, to); err != nil {
		return LangOut{}, err
	}
	var langout LangOut
	langout.Engine = "nemotron"
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig

	key := os.Getenv("MOZHI_OPENROUTER_API_KEY")
	if key == "" {
		return LangOut{}, errors.New("openrouter engine requires MOZHI_OPENROUTER_API_KEY")
	}
	model := os.Getenv("MOZHI_OPENROUTER_NEMOTRON_MODEL")
	if model == "" {
		model = "nvidia/nemotron-3-super-120b-a12b:free"
	}

	targetName := to
	for _, l := range langListOpenrouterNemotron("tl") {
		if l.Id == to {
			targetName = l.Name
			break
		}
	}
	srcName := "the source language"
	if from != "auto" {
		for _, l := range langListOpenrouterNemotron("sl") {
			if l.Id == from {
				srcName = l.Name
				break
			}
		}
	}

	transSystem := "You are a professional translator. Translate the user's text from " + srcName + " into " + targetName + ". Produce a faithful and natural translation: preserve the original meaning, tone and register, do not add, omit or explain anything. Output ONLY the translated text, with no quotes, no notes and no preamble."
	translated, err := openrouterNemotronChat(key, model, transSystem, text, 0.3)
	if err != nil {
		return LangOut{}, err
	}
	langout.OutputText = strings.TrimSpace(translated)

	if len(strings.Fields(text)) <= 8 {
		dictSystem := "You are a bilingual lexical assistant. Translating from " + srcName + " into " + targetName + ". Analyze the user text and return ONLY a JSON object (no markdown, no preamble) with this exact shape: {\"senses\":[{\"word\":\"<alt translation / sense / paraphrase>\",\"examples\":[{\"src\":\"<example in source lang>\",\"dst\":\"<its target translation>\"}]}],\"synonyms\":[\"<synonyms of the main translation, in target lang>\"],\"antonyms\":[\"<antonyms in target lang>\"]}. Rules: if the text is a single WORD or very short phrase, give its different SENSES (up to 5) with example sentences each, plus synonyms and antonyms. If it is a full SENTENCE, give up to 4 natural PARAPHRASES as senses (word = the paraphrase), with the paraphrase as src and its translation as dst; synonyms/antonyms may be empty. Always output valid JSON only."
		if raw, derr := openrouterNemotronChat(key, model, dictSystem, text, 0.3); derr == nil {
			raw = openrouterNemotronStripFences(raw)
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