package libmozhi

import (
	"errors"
	"github.com/gocolly/colly"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"encoding/json"

	deeplx "github.com/OwO-Network/DeepLX/translate"
	"github.com/google/go-querystring/query"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

var ddgVqd string

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
	data := `[[["MkEWBc","[[\"` + text + `\",\"` + from + `\",\"` + to + `\",1],[]]",null,"generic"]]]&`
	escapeData := url.PathEscape(data)
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
	gjsonArr := gjson.Get(myMemoryOut, "responseData.translatedText").Array()
	var langout LangOut
	langout.OutputText = gjsonArr[0].String()
	langout.Engine = "mymemory"
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig
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
	gjsonArr := gjson.Get(yandexOut, "text.0").Array()

	var langout LangOut

	if from == "auto" {
		langout.AutoDetect = strings.TrimSuffix(gjson.Get(yandexOut, "lang").String(), "-"+to)
	}

	yandexDictOut, err := getRequest("https://dictionary.yandex.net/dicservice.json/queryCorpus?srv=tr-text&ui=en&src=" + text + "&lang=" + from + "-" + to)
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
	translit, _ := getRequest("https://translate.yandex.net/translit/translit?lang=" + from + "&text=" + url.PathEscape(langout.OutputText))
	langout.SourceTransliteration = strings.TrimSuffix(strings.TrimPrefix(strings.ReplaceAll(strings.ReplaceAll(translit, "\\n", "\n"), "\\r", ""), `"`), `"`)
	translit2, _ := getRequest("https://translate.yandex.net/translit/translit?lang=" + to + "&text=" + url.PathEscape(text))
	langout.TargetTransliteration = strings.TrimSuffix(strings.TrimPrefix(strings.ReplaceAll(strings.ReplaceAll(translit2, "\\n", "\n"), "\\r", ""), `"`), `"`)
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

		res, err := http.DefaultClient.Do(req)
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
		baseURL = "https://generativelanguage.googleapis.com/v1beta/openai"
	}
	model := os.Getenv("MOZHI_GEMINI_MODEL")
	if model == "" {
		model = "gemini-flash-latest"
	}

	// Nombre legible del idioma destino (mejor que el código para el prompt)
	targetName := to
	for _, l := range langListGemini("tl") {
		if l.Id == to {
			targetName = l.Name
			break
		}
	}
	srcInstruction := "the source language"
	if from != "auto" {
		for _, l := range langListGemini("sl") {
			if l.Id == from {
				srcInstruction = l.Name
				break
			}
		}
	}

	systemPrompt := "You are a professional translator. Translate the user's text from " + srcInstruction + " into " + targetName + ". Produce a faithful and natural translation: preserve the original meaning, tone and register, do not add, omit or explain anything. Output ONLY the translated text, with no quotes, no notes and no preamble."

	payloadMap := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": text},
		},
		"temperature": 0.3,
	}
	payload, err := json.Marshal(payloadMap)
	if err != nil {
		return LangOut{}, err
	}

	req, err := http.NewRequest("POST", baseURL+"/chat/completions", strings.NewReader(string(payload)))
	if err != nil {
		return LangOut{}, err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return LangOut{}, err
	}
	defer res.Body.Close()

	out, err := io.ReadAll(res.Body)
	if err != nil {
		return LangOut{}, err
	}
	if res.StatusCode != 200 || !gjson.ValidBytes(out) {
		return LangOut{}, errors.New("gemini api error: " + res.Status)
	}

	langout.OutputText = strings.TrimSpace(gjson.GetBytes(out, "choices.0.message.content").String())
	return langout, nil
}

func translateTexTra(to string, from string, text string) (LangOut, error) {
	FromOrig := from
	ToOrig := to
	if err := validateLanguagePair(langListTexTra("sl"), langListTexTra("tl"), from, to); err != nil {
		return LangOut{}, err
	}

	var langout LangOut
	langout.Engine = "textra"
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig

	key := os.Getenv("MOZHI_TEXTRA_API_KEY")
	secret := os.Getenv("MOZHI_TEXTRA_API_SECRET")
	name := os.Getenv("MOZHI_TEXTRA_LOGIN_ID")
	if key == "" || secret == "" || name == "" {
		return LangOut{}, errors.New("textra engine requires MOZHI_TEXTRA_API_KEY, _API_SECRET and _LOGIN_ID")
	}
	baseURL := os.Getenv("MOZHI_TEXTRA_BASE_URL")
	if baseURL == "" {
		baseURL = "https://mt-auto-minhon-mlt.ucri.jgn-x.jp"
	}

	// 1) Conseguir access_token (OAuth2 client-credentials)
	tokenBody := url.Values{}
	tokenBody.Set("grant_type", "client_credentials")
	tokenBody.Set("client_id", key)
	tokenBody.Set("client_secret", secret)
	tokenRes, err := http.PostForm(baseURL+"/oauth2/token.php", tokenBody)
	if err != nil {
		return LangOut{}, err
	}
	defer tokenRes.Body.Close()
	tokenOut, err := io.ReadAll(tokenRes.Body)
	if err != nil {
		return LangOut{}, err
	}
	if !gjson.ValidBytes(tokenOut) {
		return LangOut{}, errors.New("textra token error")
	}
	token := gjson.GetBytes(tokenOut, "access_token").String()
	if token == "" {
		return LangOut{}, errors.New("textra: no access token")
	}

	// 2) Traducir. El par de idiomas va en la URL: generalNT_<from>_<to>
	apiParam := "generalNT_" + from + "_" + to
	transBody := url.Values{}
	transBody.Set("access_token", token)
	transBody.Set("key", key)
	transBody.Set("name", name)
	transBody.Set("type", "json")
	transBody.Set("text", text)

	transRes, err := http.PostForm(baseURL+"/api/mt/"+apiParam+"/", transBody)
	if err != nil {
		return LangOut{}, err
	}
	defer transRes.Body.Close()
	out, err := io.ReadAll(transRes.Body)
	if err != nil {
		return LangOut{}, err
	}
	if !gjson.ValidBytes(out) {
		return LangOut{}, errors.New("textra api error")
	}
	if gjson.GetBytes(out, "resultset.code").Int() != 0 {
		return LangOut{}, errors.New("textra: " + gjson.GetBytes(out, "resultset.message").String())
	}

	langout.OutputText = gjson.GetBytes(out, "resultset.result.text").String()
	return langout, nil
}


func ddgVqdUpdate() {
	r, _ := http.NewRequest("GET", "https://duckduckgo.com/?q=translate", nil)

	UserAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"
	r.Header.Set("User-Agent", UserAgent)

	client := &http.Client{}
	res, _ := client.Do(r)

	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	re := regexp.MustCompile(`vqd="([^"]*)"`)
	match := re.FindStringSubmatch(string(body))
	ddgVqd = match[1]
}

func translateDuckDuckGo(to string, from string, query string) (LangOut, error) {
	FromOrig := from
	ToOrig := to
	if to == "zh" {
		to = "zh-Hans"
	} else if from == "zh" {
		from = "zh-Hans"
	} else if from == "zh-TW" {
		from = "zh-Hant"
	} else if to == "zh-TW" {
		to = "zh-Hant"
	}
	// Validate against the original user-facing IDs for now. Once DDG lang lists
	// are fetched from upstream automatically, this should validate `from`/`to`
	// directly instead of the pre-normalized aliases.
	if err := validateLanguagePair(langListDuckDuckGo("sl"), langListDuckDuckGo("tl"), FromOrig, ToOrig); err != nil {
		return LangOut{}, err
	}
	var url string
	var langout LangOut
	ddgVqdUpdate()
	if from == "auto" {
		url = "https://duckduckgo.com/translation.js?vqd=" + ddgVqd + "&query=translate&to=" + to
	} else {
		url = "https://duckduckgo.com/translation.js?vqd=" + ddgVqd + "&query=translate&to=" + to + "&from=" + from
	}
	duckDuckGoOut, err := postRequest(url, []byte(query), "application/json")
	if err != nil {
		return LangOut{}, err
	}
	if !gjson.Valid(duckDuckGoOut) {
		return LangOut{}, errors.New("instance has been rate limited")
	}
	gjsonArr := gjson.Get(duckDuckGoOut, "translated").Array()
	langout.OutputText = gjsonArr[0].String()
	langout.Engine = "duckduckgo"
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig
	if from == "auto" {
		langout.AutoDetect = gjson.Get(duckDuckGoOut, "detected_language").String()
	}
	return langout, nil
}

func TranslateAll(to string, from string, query string) []LangOut {
	engines := []string{"google", "mymemory", "yandex", "deepl", "duckduckgo", "gemini", "textra"}
	langout := []LangOut{}
	var wg sync.WaitGroup
	for i := 0; i < len(engines); i++ {
		wg.Add(1)
		go func(i int) {
			data, err := Translate(engines[i], to, from, query)
			if err == nil {
				langout = append(langout, data)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	return langout
}

func TranslateSome(engines []string, to string, from string, query string) ([]LangOut, error) {
	enginesFull := []string{"google", "mymemory", "yandex", "deepl", "duckduckgo", "gemini", "textra"}
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
	langout := []LangOut{}
	var wg sync.WaitGroup
	for i := 0; i < len(engines); i++ {
		wg.Add(1)
		go func(i int) {
			data, err := Translate(engines[i], to, from, query)
			if err == nil {
				langout = append(langout, data)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	return langout, nil
}
