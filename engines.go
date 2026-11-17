package libmozhi

import (
	"encoding/json"
	"errors"
	"github.com/gocolly/colly"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	deeplx "github.com/OwO-Network/DeepLX/translate"
	"github.com/google/go-querystring/query"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

// Per-engine HTTP clients. http.DefaultClient has no timeout, so a stalled
// upstream used to hang the request (and the All/Some modes) forever. 15s
// matches the timeout upstream already uses in postRequest (requests.go).
var (
	deeplHTTP  = &http.Client{Timeout: 15 * time.Second}
	geminiHTTP = &http.Client{Timeout: 15 * time.Second}
	textraHTTP = &http.Client{Timeout: 15 * time.Second}
)

func translateGoogle(to string, from string, text string) (LangOut, error) {
	ToOrig := to
	FromOrig := from
	// For some reason google uses no for norwegian instead of nb like the rest of the translators. This is for the All function primarily
	if to == "nb" {
		to = "no"
	} else if from == "nb" {
		from = "no"
	}
	if err := validateLanguagePair(langListGoogle("sl"), langListGoogle("tl"), from, to); err != nil {
		return LangOut{}, err
	}
	text = strings.ReplaceAll(text, "\n", "\\\\n")
	text = strings.ReplaceAll(text, "\r", "\\\\r")
	text = strings.ReplaceAll(text, `"`, `\\\"`)
	//text = strings.ReplaceAll(text, "\r", "")
	// curl -XPOST 'https://translate.google.com/_/TranslateWebserverUi/data/batchexecute' -d 'f.req=[[["MkEWBc", "[[\"Hello World!\",\"auto\",\"fr\",1],[]]",null,"generic"]]]'
	data := `[[["MkEWBc","[[\"` + text + `\",\"` + from + `\",\"` + to + `\",1],[]]",null,"generic"]]]`
	// Form body: QueryEscape, not PathEscape, which left & and + as they were:
	// the request was cut at the first & of the text (empty translation)
	// and every + became a space.
	escapeData := url.QueryEscape(data) + "&"
	//escapeData = strings.ReplaceAll(escapeData, "+", )
	q := `f.req=` + escapeData
	googleOut, err := postRequest("https://translate.google.com/_/TranslateWebserverUi/data/batchexecute", []byte(q), "application/x-www-form-urlencoded")
	if err != nil {
		return LangOut{}, err
	}
	googleOut = strings.TrimPrefix(googleOut, ")]}'")
	googleOut = strings.TrimSuffix(googleOut, "]")
	googleOut = strings.TrimPrefix(googleOut, "[")
	if !gjson.Valid(googleOut + "]") {
		return LangOut{}, errors.New("instance has been rate limited")
	}
	initial := gjson.Get(googleOut, "0.2").String()

	// Thanks jsonselector.com
	textArr := gjson.Get(initial, "1.0.0.5.#.0")
	var textNew string
	for _, text := range textArr.Array() {
		textNew = textNew + text.String() + " "
	}
	textNew = strings.TrimSuffix(textNew, " ")
	langout := dictDataGoogle(text, textNew, from, to)
	langout.OutputText = textNew
	if from == "auto" {
		langout.AutoDetect = gjson.Get(initial, "0.2").String()
	}
	langout.TargetTransliteration = gjson.Get(initial, "1.0.0.1").String()
	langout.SourceTransliteration = gjson.Get(initial, "0.0").String()
	langout.Engine = "google"
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig
	return langout, nil
}

func translateReverso(to string, from string, query string) (LangOut, error) {
	ToOrig := to
	FromOrig := from
	if err := validateLanguagePair(langListReverso("sl"), langListReverso("tl"), from, to); err != nil {
		return LangOut{}, err
	}
	json := []byte(`{ "format": "text", "from": "` + from + `", "to": "` + to + `", "input":"` + query + `", "options": {"sentenceSplitter": false, "origin":"translation.web", contextResults: true, languageDetection: true} }`)
	reversoOut, err := postRequest("https://api.reverso.net/translate/v1/translation", json, "application/json")
	if err != nil {
		return LangOut{}, err
	}
	if !gjson.Valid(reversoOut) {
		return LangOut{}, errors.New("instance has been rate limited")
	}
	gjsonArr := gjson.Get(reversoOut, "translation").Array()
	var langout LangOut
	langout.OutputText = gjsonArr[0].String()
	langout.Engine = "reverso"
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig
	examples := gjson.Get(reversoOut, "contextResults.results")
	langout.SourceTransliteration = gjson.Get(reversoOut, "contextResults.results.0.transliteration").String()
	for _, translation := range examples.Array() {
		var WordChoices WordChoices // WordChoices is the struct as well as var name
		WordChoices.Word = translation.Get("translation").String()
		for _, source := range translation.Get("sourceExamples").Array() {
			WordChoices.ExamplesSource = append(WordChoices.ExamplesSource, source.String())
		}
		for _, target := range translation.Get("targetExamples").Array() {
			WordChoices.ExamplesTarget = append(WordChoices.ExamplesTarget, target.String())
		}
		if len(WordChoices.ExamplesSource) == 0 && len(WordChoices.ExamplesTarget) == 0 {
			continue
		}
		langout.WordChoices = append(langout.WordChoices, WordChoices)
	}
	UserAgent, ok := os.LookupEnv("MOZHI_USER_AGENT")
	if !ok {
		UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"
	}
	sc1 := colly.NewCollector(colly.AllowedDomains("synonyms.reverso.net"), colly.UserAgent(UserAgent))
	sc1.OnHTML("body", func(e *colly.HTMLElement) {
		e.ForEach("div.pannel li a.synonym", func(i int, el *colly.HTMLElement) {
			langout.SourceSynonyms = append(langout.SourceSynonyms, el.Text)
		})
		e.ForEach("div.antonyms-wrapper ul.word-box li a", func(i int, el *colly.HTMLElement) {
			langout.SourceAntonyms = append(langout.SourceAntonyms, el.Text)
		})
	})
	sc1.Visit("https://synonyms.reverso.net/synonym/" + from + "/" + query)
	sc2 := colly.NewCollector(colly.AllowedDomains("synonyms.reverso.net"), colly.UserAgent(UserAgent))
	sc2.OnHTML("body", func(e *colly.HTMLElement) {
		e.ForEach("div.pannel li a.synonym", func(i int, el *colly.HTMLElement) {
			langout.TargetSynonyms = append(langout.TargetSynonyms, el.Text)
		})
		e.ForEach("div.antonyms-wrapper ul.word-box li a", func(i int, el *colly.HTMLElement) {
			langout.TargetAntonyms = append(langout.TargetAntonyms, el.Text)
		})
	})
	sc2.Visit("https://synonyms.reverso.net/synonym/" + to + "/" + langout.OutputText)
	return langout, nil
}

func translateLibreTranslate(to string, from string, query string) (LangOut, error) {
	ToOrig := to
	FromOrig := from
	if err := validateLanguagePair(langListLibreTranslate("sl"), langListLibreTranslate("tl"), from, to); err != nil {
		return LangOut{}, err
	}
	json := []byte(`{"q":"` + query + `","source":"` + from + `","target":"` + to + `"}`)
	// TODO: Make it configurable
	libreTranslateOut, err := postRequest(os.Getenv("MOZHI_LIBRETRANSLATE_URL")+"/translate", json, "application/json")
	if err != nil {
		return LangOut{}, err
	}
	if !gjson.Valid(libreTranslateOut) {
		return LangOut{}, errors.New("instance has been rate limited")
	}
	gjsonArr := gjson.Get(libreTranslateOut, "translatedText").Array()
	var langout LangOut
	langout.OutputText = gjsonArr[0].String()
	langout.Engine = "libre"
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig
	if from == "auto" {
		json := []byte(`{"q":"` + query + `"}`)
		libreTranslateOut, err := postRequest(os.Getenv("MOZHI_LIBRETRANSLATE_URL")+"/detect", json, "application/json")
		if err == nil {
			gjsonArr := gjson.Get(libreTranslateOut, "0.language").Array()
			langout.AutoDetect = gjsonArr[0].String()
		}
	}
	return langout, nil
}

func translateMyMemory(to string, from string, text string) (LangOut, error) {
	FromOrig := from
	ToOrig := to
	if err := validateLanguagePair(langListMyMemory("sl"), langListMyMemory("tl"), from, to); err != nil {
		return LangOut{}, err
	}
	type Options struct {
		Translate string `url:"langpair"`
		Text      string `url:"q"`
		Email     string `url:"de,omitempty"`
	}
	opt := Options{from + "|" + to, text, os.Getenv("MOZHI_MYMEMORY_EMAIL")}
	v, _ := query.Values(opt)
	myMemoryOut, err := postRequest("https://api.mymemory.translated.net/get", []byte(v.Encode()), "application/x-www-form-urlencoded")
	if err != nil {
		return LangOut{}, err
	}
	if !gjson.Valid(myMemoryOut) {
		return LangOut{}, errors.New("instance has been rate limited")
	}
	// MyMemory answers errors and its daily quota warning inside a normal JSON
	// body ("MYMEMORY WARNING: YOU USED ALL AVAILABLE FREE TRANSLATIONS..."),
	// which used to be shown as the translation.
	translated := gjson.Get(myMemoryOut, "responseData.translatedText").String()
	if status := gjson.Get(myMemoryOut, "responseStatus").Int(); status != 200 ||
		gjson.Get(myMemoryOut, "quotaFinished").Bool() || strings.HasPrefix(translated, "MYMEMORY WARNING") {
		detail := gjson.Get(myMemoryOut, "responseDetails").String()
		if detail == "" {
			detail = translated
		}
		return LangOut{}, errors.New("mymemory error " + strconv.FormatInt(status, 10) + ": " + detail)
	}
	if translated == "" {
		return LangOut{}, errors.New("mymemory: no translation in response")
	}
	var langout LangOut
	langout.OutputText = translated
	langout.Engine = "mymemory"
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig

	// Mapear los "matches" (traducciones alternativas de la memoria) a WordChoices.
	// Filtros mínimos: sin duplicados exactos y sin vacíos. Sin ejemplos (MyMemory no los tiene).
	seen := map[string]bool{}
	seen[langout.OutputText] = true // no repetir la traducción principal
	for _, m := range gjson.Get(myMemoryOut, "matches").Array() {
		t := m.Get("translation").String()
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		langout.WordChoices = append(langout.WordChoices, WordChoices{Word: t})
	}

	return langout, nil
}

func translateYandex(to string, from string, text string) (LangOut, error) {
	FromOrig := from
	ToOrig := to
	if err := validateLanguagePair(langListYandex("sl"), langListYandex("tl"), from, to); err != nil {
		return LangOut{}, err
	}
	type Options struct {
		Translate string `url:"lang"`
		Text      string `url:"text"`
		Srv       string `url:"srv"`
		Id        string `url:"sid"`
	}
	uuidWithHyphen := uuid.New()
	uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)
	var opt Options
	if from == "auto" {
		opt = Options{to, text, "android", uuid + "-0-0"}
	} else {
		opt = Options{from + "-" + to, text, "android", uuid + "-0-0"}
	}
	v, _ := query.Values(opt)

	yandexOut, err := postRequest("https://translate.yandex.net/api/v1/tr.json/translate?"+v.Encode(), []byte(""), "application/json")
	if err != nil {
		return LangOut{}, err
	}
	if !gjson.Valid(yandexOut) {
		return LangOut{}, errors.New("instance has been rate limited")
	}
	// Errors come back as {"code":4xx,"message":"..."} instead of a translation.
	if code := gjson.Get(yandexOut, "code").Int(); code != 0 && code != 200 {
		return LangOut{}, errors.New("yandex error " + strconv.FormatInt(code, 10) + ": " + gjson.Get(yandexOut, "message").String())
	}
	gjsonArr := gjson.Get(yandexOut, "text.0").Array()
	if len(gjsonArr) == 0 {
		return LangOut{}, errors.New("yandex: no translation in response")
	}

	var langout LangOut

	if from == "auto" {
		langout.AutoDetect = strings.TrimSuffix(gjson.Get(yandexOut, "lang").String(), "-"+to)
	}

	yandexDictOut, err := getRequest("https://dictionary.yandex.net/dicservice.json/queryCorpus?srv=tr-text&ui=en&src=" + url.QueryEscape(text) + "&lang=" + from + "-" + to)
	if err == nil {
		for _, example := range gjson.Get(yandexDictOut, "result").Array() {
			var WordChoices WordChoices // WordChoices is the struct as well as var name
			WordChoices.Word = example.Get("translation.text").String()
			langout.TargetSynonyms = append(langout.TargetSynonyms, WordChoices.Word)
			if WordChoices.Word == "" {
				WordChoices.Word = "other translations"
			}
			for _, sentence := range example.Get("examples").Array() {
				WordChoices.ExamplesSource = append(WordChoices.ExamplesSource, sentence.Get("src").String())
				WordChoices.ExamplesTarget = append(WordChoices.ExamplesTarget, sentence.Get("dst").String())
			}
			langout.WordChoices = append(langout.WordChoices, WordChoices)
		}
	}
	yandexDict2Out, err := getRequest("https://dictionary.yandex.net/dicservice.json/lookupMultiple?ui=en&srv=tr-text&type=regular%2Csyn%2Cant%2Cderiv&lang=en-fr&dict=" + from + "-" + to + ".regular%2Cen.syn%2Cen.ant%2Cen.deriv&exp_def&text=" + text)
	if err == nil {
		for _, syn := range gjson.Get(yandexDict2Out, "en.syn.0.tr").Array() {
			langout.SourceSynonyms = append(langout.SourceSynonyms, syn.Get("text").String())
			for _, syn2 := range syn.Get("syn").Array() {
				langout.SourceSynonyms = append(langout.SourceSynonyms, syn2.Get("text").String())
			}
		}
	}
	langout.OutputText = gjsonArr[0].String()
	langout.Engine = "yandex"
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig
	srcLang := from
	if from == "auto" {
		srcLang = langout.AutoDetect
	}
	langout.SourceTransliteration = yandexTranslit(srcLang, text)
	langout.TargetTransliteration = yandexTranslit(to, langout.OutputText)
	return langout, nil
}

func translateDeepl(to string, from string, text string) (LangOut, error) {
	FromOrig := from
	ToOrig := to
	if err := validateLanguagePair(langListDeepl("sl"), langListDeepl("tl"), from, to); err != nil {
		return LangOut{}, err
	}

	var langout LangOut
	langout.Engine = "deepl"
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig

	// Si hay key oficial, usa la API REST oficial.
	if key := os.Getenv("MOZHI_DEEPL_API_KEY"); key != "" {
		endpoint := "https://api.deepl.com/v2/translate"
		if strings.HasSuffix(key, ":fx") { // las keys free acaban en ":fx"
			endpoint = "https://api-free.deepl.com/v2/translate"
		}

		payloadMap := map[string]interface{}{
			"text":        []string{text},
			"target_lang": strings.ToUpper(to),
		}
		if from != "auto" { // DeepL autodetecta si omites source_lang
			payloadMap["source_lang"] = strings.ToUpper(from)
		}
		payload, err := json.Marshal(payloadMap)
		if err != nil {
			return LangOut{}, err
		}

		req, err := http.NewRequest("POST", endpoint, strings.NewReader(string(payload)))
		if err != nil {
			return LangOut{}, err
		}
		req.Header.Set("Authorization", "DeepL-Auth-Key "+key)
		req.Header.Set("Content-Type", "application/json")

		res, err := deeplHTTP.Do(req)
		if err != nil {
			return LangOut{}, err
		}
		defer res.Body.Close()

		out, err := io.ReadAll(res.Body)
		if err != nil {
			return LangOut{}, err
		}
		if res.StatusCode != 200 || !gjson.ValidBytes(out) {
			return LangOut{}, errors.New("deepl official api error: " + res.Status)
		}

		langout.OutputText = gjson.GetBytes(out, "translations.0.text").String()
		if from == "auto" {
			langout.AutoDetect = strings.ToLower(gjson.GetBytes(out, "translations.0.detected_source_language").String())
		}
		return langout, nil
	}

	// Fallback: endpoint gratis reverse-engineered (comportamiento original).
	answer, err := deeplx.TranslateByDeepLX(from, to, text, "plaintext", "", "0")
	if err != nil {
		return LangOut{}, err
	}
	langout.OutputText = answer.Data
	langout.TargetSynonyms = answer.Alternatives
	return langout, nil
}

// stripJSONFences quita ```json ... ``` si el modelo envolvió la respuesta.
func stripJSONFences(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func translateGemini(to string, from string, text string) (LangOut, error) {
	FromOrig := from
	ToOrig := to
	if err := validateLanguagePair(langListGemini("sl"), langListGemini("tl"), from, to); err != nil {
		return LangOut{}, err
	}

	var langout LangOut
	langout.Engine = "gemini"
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig

	key := os.Getenv("MOZHI_GEMINI_API_KEY")
	if key == "" {
		return LangOut{}, errors.New("gemini engine requires MOZHI_GEMINI_API_KEY")
	}
	baseURL := os.Getenv("MOZHI_GEMINI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta"
	}
	// Comma-separated preference list (see geminiPickModel in gemini.go).
	model := os.Getenv("MOZHI_GEMINI_MODEL")

	// Nombre legible de los idiomas para los prompts.
	targetName := to
	for _, l := range langListGemini("tl") {
		if l.Id == to {
			targetName = l.Name
			break
		}
	}
	srcName := "the source language"
	if from != "auto" {
		for _, l := range langListGemini("sl") {
			if l.Id == from {
				srcName = l.Name
				break
			}
		}
	}

	// 1) Traducción principal.
	transSystem := "You are a professional translator. Translate the user's text from " + srcName + " into " + targetName + ". Produce a faithful and natural translation: preserve the original meaning, tone and register, do not add, omit or explain anything. Output ONLY the translated text, with no quotes, no notes and no preamble."
	translated, err := geminiChat(baseURL, key, model, transSystem, text, 0.3)
	if err != nil {
		return LangOut{}, err
	}
	langout.OutputText = strings.TrimSpace(translated)

	// 2) Extras (diccionario / paráfrasis) sólo para textos cortos (<=8 palabras).
	if len(strings.Fields(text)) <= 8 {
		dictSystem := "You are a bilingual lexical assistant. Translating from " + srcName + " into " + targetName + ". Analyze the user text and return ONLY a JSON object (no markdown, no preamble) with this exact shape: {\"senses\":[{\"word\":\"<alt translation / sense / paraphrase>\",\"examples\":[{\"src\":\"<example in source lang>\",\"dst\":\"<its target translation>\"}]}],\"synonyms\":[\"<synonyms of the main translation, in target lang>\"],\"antonyms\":[\"<antonyms in target lang>\"]}. Rules: if the text is a single WORD or very short phrase, give its different SENSES (up to 5) with example sentences each, plus synonyms and antonyms. If it is a full SENTENCE, give up to 4 natural PARAPHRASES as senses (word = the paraphrase), with the paraphrase as src and its translation as dst; synonyms/antonyms may be empty. Always output valid JSON only."

		if raw, derr := geminiChat(baseURL, key, model, dictSystem, text, 0.3); derr == nil {
			raw = stripJSONFences(raw)
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

// textraDictPID mapea un par de idiomas al glosario de sistema de TexTra, si existe.
func textraDictPID(from, to string) string {
	pairs := map[string]string{
		"en_ja":    "ld.en_ja",
		"ja_en":    "ld.ja_en",
		"ja_ko":    "kj.ja_ko",
		"ko_ja":    "kj.ko_ja",
		"ja_zh-CN": "er.ja_zh-CN",
		"ja_zh-TW": "er.ja_zh-TW",
		"zh-CN_ja": "er.zh-CN_ja",
		"zh-TW_ja": "er.zh-TW_ja",
		"en_en":    "wn.en_en",
	}
	return pairs[from+"_"+to]
}

func TranslateAll(to string, from string, query string) []LangOut {
	engines := []string{"google", "mymemory", "yandex", "deepl", "duckduckgo", "gemini", "textra", "groq", "qwen", "cohere", "microsoft", "langbly", "papago", "modernmt", "volcengine"}
	langout := runEngines(engines, to, from, query)
	return langout
}

func TranslateSome(engines []string, to string, from string, query string) ([]LangOut, error) {
	enginesFull := []string{"google", "mymemory", "yandex", "deepl", "duckduckgo", "gemini", "textra", "groq", "qwen", "cohere", "microsoft", "langbly", "papago", "modernmt", "volcengine"}
	for i := range engines {
		valid := false
		for j := range enginesFull {
			if engines[i] == enginesFull[j] {
				valid = true
			}
		}
		if valid == false {
			return []LangOut{}, errors.New("Engine " + engines[i] + "not supported or implemented")
		}
	}
	langout := runEngines(engines, to, from, query)
	return langout, nil
}
