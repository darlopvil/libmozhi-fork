package libmozhi

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/OwO-Network/gdeeplx"
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
		to = "no"
	}
	var ToValid bool
	var FromValid bool
	for _, v := range langListGoogle("sl") {
		if v.Id == to {
			ToValid = true
		}
		if v.Id == from {
			FromValid = true
		}
		if FromValid == true && ToValid == true {
			break
		}
	}
	if ToValid != true {
		return LangOut{}, errors.New("Target Language Code invalid")
	}
	if FromValid != true {
		return LangOut{}, errors.New("Source language code invalid")
	}
	// curl -XPOST 'https://translate.google.com/_/TranslateWebserverUi/data/batchexecute' -d 'f.req=[[["MkEWBc", "[[\"Hello World!\",\"auto\",\"fr\",1],[]]",null,"generic"]]]'
	escapeText := url.QueryEscape(text)
	data := []byte(`f.req=[[["MkEWBc", "[[\"` + escapeText + `\",\"` + from + `\",\"` + to + `\",1],[]]",null,"generic"]]]`)
	googleOut, err := postRequest("https://translate.google.com/_/TranslateWebserverUi/data/batchexecute", data, "application/x-www-form-urlencoded")
	if err != nil {
		return LangOut{}, err
	}
	googleOut = strings.TrimPrefix(googleOut, ")]}'")
	googleOut = strings.TrimSuffix(googleOut, "]")
	googleOut = strings.TrimPrefix(googleOut, "[")
	initial := gjson.Get(googleOut, "0.2").String()

	var langout LangOut
	// Thanks jsonselector.com
	langout.OutputText = gjson.Get(initial, "1.0.0.5.0.4.0.0").String()
	if from == "auto" {
		langout.AutoDetect = gjson.Get(initial, "0.2").String()
	}
	langout.Engine = "google"
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig
	return langout, nil
}

func translateReverso(to string, from string, query string) (LangOut, error) {
	ToOrig := to
	FromOrig := from
	var ToValid bool
	var FromValid bool
	for _, v := range langListReverso("sl") {
		if v.Id == to {
			ToValid = true
		}
		if v.Id == from {
			FromValid = true
		}
		if FromValid == true && ToValid == true {
			break
		}
	}
	if ToValid != true {
		return LangOut{}, errors.New("Target language code invalid")
	}
	if FromValid != true {
		return LangOut{}, errors.New("Source language code invalid")
	}
	json := []byte(`{ "format": "text", "from": "` + from + `", "to": "` + to + `", "input":"` + query + `", "options": {"sentenceSplitter": false, "origin":"translation.web", contextResults: true, languageDetection: true} }`)
	reversoOut, err := postRequest("https://api.reverso.net/translate/v1/translation", json, "application/json")
	if err != nil {
		return LangOut{}, err
	}
	gjsonArr := gjson.Get(reversoOut, "translation").Array()
	var langout LangOut
	langout.OutputText = gjsonArr[0].String()
	langout.Engine = "reverso"
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig
	examples := gjson.Get(reversoOut, "contextResults.results")
	for _, translation := range examples.Array() {
		var WordChoices WordChoices // WordChoices is the struct as well as var name
		WordChoices.Word = translation.Get("translation").String()
		for _, source := range translation.Get("sourceExamples").Array() {
			WordChoices.ExamplesSource = append(WordChoices.ExamplesSource, source.String())
		}
		for _, target := range translation.Get("targetExamples").Array() {
			WordChoices.ExamplesTarget = append(WordChoices.ExamplesTarget, target.String())
		}
		langout.WordChoices = append(langout.WordChoices, WordChoices)
	}
	return langout, nil
}

func translateLibreTranslate(to string, from string, query string) (LangOut, error) {
	ToOrig := to
	FromOrig := from
	var ToValid bool
	var FromValid bool
	for _, v := range langListLibreTranslate("sl") {
		if v.Id == to {
			ToValid = true
		}
		if v.Id == from {
			FromValid = true
		}
		if FromValid == true && ToValid == true {
			break
		}
	}
	if ToValid != true {
		return LangOut{}, errors.New("Target language code invalid")
	}
	if FromValid != true {
		return LangOut{}, errors.New("Source language code invalid")
	}
	json := []byte(`{"q":"` + query + `","source":"` + from + `","target":"` + to + `"}`)
	// TODO: Make it configurable
	libreTranslateOut, err := postRequest(os.Getenv("MOZHI_LIBRETRANSLATE_URL")+"/translate", json, "application/json")
	if err != nil {
		return LangOut{}, err
	}
	gjsonArr := gjson.Get(libreTranslateOut, "translatedText").Array()
	var langout LangOut
	langout.OutputText = gjsonArr[0].String()
	langout.Engine = "libre"
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig
	if from == "auto" {
		langout.AutoDetect, _ = AutoDetectLibreTranslate(query)
	}
	return langout, nil
}

func translateWatson(to string, from string, query string) (LangOut, error) {
	FromOrig := from
	ToOrig := to
	var ToValid bool
	var FromValid bool
	for _, v := range langListWatson("sl") {
		if v.Id == to {
			ToValid = true
		}
		if v.Id == from {
			FromValid = true
		}
		if FromValid == true && ToValid == true {
			break
		}
	}
	if ToValid != true {
		return LangOut{}, errors.New("Target language code invalid")
	}
	if FromValid != true {
		return LangOut{}, errors.New("Source language code invalid")
	}
	var langout LangOut
	if from == "auto" {
		langout.AutoDetect, _ = AutoDetectWatson(query)
		from = langout.AutoDetect
	}
	query = strings.ReplaceAll(query, "\n", "\\n\\n")
	query = strings.ReplaceAll(query, "\r", "\\r\\r")
	json := []byte(`{"text":"` + query + `","source":"` + from + `","target":"` + to + `"}`)
	watsonOut, err := postRequest("https://www.ibm.com/demos/live/watson-language-translator/api/translate/text", json, "application/json")
	if err != nil {
		return LangOut{}, err
	}
	gjsonArr := gjson.Get(watsonOut, "payload.translations.0.translation").Array()
	text := strings.ReplaceAll(gjsonArr[0].String(), "\n\n", "\n")
	text = strings.ReplaceAll(text, "\r\r", "\r")
	langout.OutputText = text
	langout.Engine = "watson"
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig
	return langout, nil
}

func translateMyMemory(to string, from string, text string) (LangOut, error) {
	FromOrig := from
	ToOrig := to
	var ToValid bool
	var FromValid bool
	for _, v := range langListMyMemory("sl") {
		if v.Id == to {
			ToValid = true
		}
		if v.Id == from {
			FromValid = true
		}
		if FromValid == true && ToValid == true {
			break
		}
	}
	if ToValid != true {
		return LangOut{}, errors.New("Target language code invalid")
	}
	if FromValid != true {
		return LangOut{}, errors.New("Source language code invalid")
	}
	type Options struct {
		Translate string `url:"langpair"`
		Text      string `url:"q"`
	}
	opt := Options{from + "|" + to, text}
	v, _ := query.Values(opt)
	myMemoryOut, err := getRequest("https://api.mymemory.translated.net/get?" + v.Encode())
	if err != nil {
		return LangOut{}, err
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
	var ToValid bool
	var FromValid bool
	for _, v := range langListYandex("sl") {
		if v.Id == to {
			ToValid = true
		}
		if v.Id == from {
			FromValid = true
		}
		if FromValid == true && ToValid == true {
			break
		}
	}
	if ToValid != true {
		return LangOut{}, errors.New("Target language code invalid")
	}
	if FromValid != true {
		return LangOut{}, errors.New("Source language code invalid")
	}
	type Options struct {
		Translate string `url:"lang"`
		Text      string `url:"text"`
		Srv       string `url:"srv"`
		Id        string `url:"sid"`
	}
	uuidWithHyphen := uuid.New()
	uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)
	opt := Options{from + "-" + to, text, "android", uuid + "-0-0"}
	v, _ := query.Values(opt)

	yandexOut, err := postRequest("https://translate.yandex.net/api/v1/tr.json/translate?"+v.Encode(), []byte(""), "application/json")
	if err != nil {
		return LangOut{}, err
	}
	gjsonArr := gjson.Get(yandexOut, "text.0").Array()

	var langout LangOut

	yandexDictOut, err := getRequest("https://dictionary.yandex.net/dicservice.json/queryCorpus?srv=tr-text&ui=en&src=" + text + "&lang=" + from + "-" + to)
	if err == nil {
		for _, example := range gjson.Get(yandexDictOut, "result").Array() {
			var WordChoices WordChoices // WordChoices is the struct as well as var name
			WordChoices.Word = example.Get("translation.text").String()
			if WordChoices.Word == "" { WordChoices.Word = "other translations" }
			for _, sentence := range example.Get("examples").Array() {
				WordChoices.ExamplesSource = append(WordChoices.ExamplesSource, sentence.Get("src").String())
				WordChoices.ExamplesTarget = append(WordChoices.ExamplesTarget, sentence.Get("dst").String())
			}
			langout.WordChoices = append(langout.WordChoices, WordChoices)
		}
	}
	langout.OutputText = gjsonArr[0].String()
	langout.Engine = "yandex"
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig
	return langout, nil
}

func translateDeepl(to string, from string, text string) (LangOut, error) {
	FromOrig := from
	ToOrig := to
	var ToValid bool
	var FromValid bool
	for _, v := range langListDeepl("sl") {
		if v.Id == to {
			ToValid = true
		}
		if v.Id == from {
			FromValid = true
		}
		if FromValid == true && ToValid == true {
			break
		}
	}
	if ToValid != true {
		return LangOut{}, errors.New("Target language code invalid")
	}
	if FromValid != true {
		return LangOut{}, errors.New("Source language code invalid")
	}
	answer, err := gdeeplx.Translate(text, from, to, 0)
	if err != nil {
		return LangOut{}, errors.New("failed")
	}
	answer1 := answer.(map[string]interface{})
	ans := answer1["data"].(string)
	var langout LangOut
	langout.OutputText = ans
	langout.Engine = "deepl"
	if from == "auto" {
		langout.AutoDetect = strings.ToLower(answer1["detected_lang"].(string))
	}
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig
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
	var ToValid bool
	var FromValid bool
	for _, v := range langListDuckDuckGo("sl") {
		if v.Id == to {
			ToValid = true
		}
		if v.Id == from {
			FromValid = true
		}
		if FromValid == true && ToValid == true {
			break
		}
	}
	if ToValid != true {
		return LangOut{}, errors.New("Target language code invalid")
	}
	if FromValid != true {
		return LangOut{}, errors.New("Source language code invalid")
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
	engines := []string{"reverso", "google", "libre", "watson", "mymemory", "yandex", "deepl", "duckduckgo"}
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
