package libmozhi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

// ============================================================================
// QWEN ENGINE — Alibaba's Qwen through Groq. Groq rate-limits every model on
// its own (qwen3.8-27b, nov 2026: 1,000 requests a day, 8,000 tokens a
// minute), so this engine doesn't spend the Groq engine's quota. It uses the
// same Groq key and Groq's model list helpers (groqPickModel, groqModelGone
// in groq.go); the chat call is its own, because Qwen needs its reasoning
// turned off.
// ============================================================================

const qwenDefaultModels = "qwen/qwen3.8-27b"

// qwenThink strips a reasoning block, in case a model sends one anyway.
var qwenThink = regexp.MustCompile(`(?s)^\s*<think>.*?</think>`)

func qwenChat(baseURL, key, prefs, systemPrompt, userText string) (string, error) {
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
			"temperature":      0.3,
			"reasoning_effort": "none", // measured: 0.4-1.3s per translation
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
			lastErr = errors.New("qwen (groq) api error: " + res.Status)
			continue
		}
		if res.StatusCode != 200 || !gjson.ValidBytes(out) {
			apiErr := errors.New("qwen (groq) api error: " + res.Status + ": " + gjson.GetBytes(out, "error.message").String())
			// Decommissioned or unknown model: retry once with the next one.
			if !refreshed && groqModelGone(model, res.StatusCode, out) {
				refreshed = true
				lastErr = apiErr
				continue
			}
			return "", apiErr
		}
		content := gjson.GetBytes(out, "choices.0.message.content").String()
		return strings.TrimSpace(qwenThink.ReplaceAllString(content, "")), nil
	}
	return "", lastErr
}

func translateQwen(to string, from string, text string) (LangOut, error) {
	if err := validateLanguagePair(langListQwen("sl"), langListQwen("tl"), from, to); err != nil {
		return LangOut{}, err
	}
	key := os.Getenv("MOZHI_GROQ_API_KEY")
	if key == "" {
		return LangOut{}, errors.New("qwen engine requires MOZHI_GROQ_API_KEY")
	}
	baseURL := os.Getenv("MOZHI_GROQ_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.groq.com/openai/v1"
	}
	// Comma-separated preference list; the first model Groq lists as active
	// is used.
	models := os.Getenv("MOZHI_QWEN_MODEL")
	if strings.TrimSpace(models) == "" {
		models = qwenDefaultModels
	}

	targetName := to
	for _, l := range langListQwen("tl") {
		if l.Id == to {
			targetName = l.Name
			break
		}
	}
	srcName := "the source language"
	if from != "auto" {
		for _, l := range langListQwen("sl") {
			if l.Id == from {
				srcName = l.Name
				break
			}
		}
	}

	// 1) Translation.
	transSystem := "You are a professional translator. Translate the user's text from " + srcName + " into " + targetName + ". Produce a faithful and natural translation: preserve the original meaning, tone and register, do not add, omit or explain anything. Output ONLY the translated text, with no quotes, no notes and no preamble."
	translated, err := qwenChat(baseURL, key, models, transSystem, text)
	if err != nil {
		return LangOut{}, err
	}
	if translated == "" {
		return LangOut{}, errors.New("qwen: empty answer")
	}

	var langout LangOut
	langout.Engine = "qwen"
	langout.SourceLang = from
	langout.TargetLang = to
	langout.OutputText = translated

	// 2) Dictionary / paraphrases (short text).
	if len(strings.Fields(text)) <= 8 {
		dictSystem := "You are a bilingual lexical assistant. Translating from " + srcName + " into " + targetName + ". Analyze the user text and return ONLY a JSON object (no markdown, no preamble) with this exact shape: {\"senses\":[{\"word\":\"<alt translation / sense / paraphrase>\",\"examples\":[{\"src\":\"<example in source lang>\",\"dst\":\"<its target translation>\"}]}],\"synonyms\":[\"<synonyms of the main translation, in target lang>\"],\"antonyms\":[\"<antonyms in target lang>\"]}. Rules: if the text is a single WORD or very short phrase, give its different SENSES (up to 5) with example sentences each, plus synonyms and antonyms. If it is a full SENTENCE, give up to 4 natural PARAPHRASES as senses (word = the paraphrase), with the paraphrase as src and its translation as dst; synonyms/antonyms may be empty. Always output valid JSON only."
		if raw, derr := qwenChat(baseURL, key, models, dictSystem, text); derr == nil {
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
