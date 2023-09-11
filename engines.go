package libmozhi

import (
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/OwO-Network/gdeeplx"
	"github.com/gocolly/colly"
	"github.com/google/go-querystring/query"
	"github.com/google/uuid"
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

	UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"
	sc := colly.NewCollector(colly.AllowedDomains("translate.google.com"), colly.UserAgent(UserAgent))
	var answer string
	sc.OnHTML("div.result-container", func(e *colly.HTMLElement) {
		answer = e.Text
	})
	type Options struct {
		To   string `url:"tl"`
		UI   string `url:"hl"`
		From string `url:"sl"`
		Text string `url:"q"`
	}
	opt := Options{to, to, from, text}
	v, _ := query.Values(opt)
	url := "https://translate.google.com/m?" + v.Encode()
	sc.Visit(url)
	var langout LangOut
	langout.OutputText = answer
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
	json := []byte(`{ "format": "text", "from": "` + from + `", "to": "` + to + `", "input":"` + query + `", "options": {"sentenceSplitter": false, "origin":"translation.web", contextResults: false, languageDetection: true} }`)
	reversoOut := postRequest("https://api.reverso.net/translate/v1/translation", json)
	gjsonArr := reversoOut.Get("translation").Array()
	var langout LangOut
	langout.OutputText = gjsonArr[0].String()
	langout.Engine = "reverso"
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig
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
	libreTranslateOut := postRequest(os.Getenv("MOZHI_LIBRETRANSLATE_URL")+"/translate", json)
	gjsonArr := libreTranslateOut.Get("translatedText").Array()
	var langout LangOut
	langout.OutputText = gjsonArr[0].String()
	langout.Engine = "libretranslate"
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
	json := []byte(`{"text":"` + query + `","source":"` + from + `","target":"` + to + `"}`)
	watsonOut := postRequest("https://www.ibm.com/demos/live/watson-language-translator/api/translate/text", json)
	gjsonArr := watsonOut.Get("payload.translations.0.translation").Array()
	langout.OutputText = gjsonArr[0].String()
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
	myMemoryOut := getRequest("https://api.mymemory.translated.net/get?" + v.Encode())
	gjsonArr := myMemoryOut.Get("responseData.translatedText").Array()
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

	yandexOut := postRequest("https://translate.yandex.net/api/v1/tr.json/translate?"+v.Encode(), []byte(""))
	gjsonArr := yandexOut.Get("text.0").Array()
	var langout LangOut
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
	r, err := http.NewRequest("GET", "https://duckduckgo.com/?q=translate", nil)
	if err != nil {
		panic(err)
	}

	UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"
	r.Header.Set("User-Agent", UserAgent)

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	re := regexp.MustCompile(`vqd="([^"]*)"`)
	match := re.FindStringSubmatch(string(body))
	ddgVqd = match[1]
}

func translateDuckDuckGo(to string, from string, query string) (LangOut, error) {
	FromOrig := from
	ToOrig := to
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
	duckDuckGoOut := postRequest(url, []byte(query))
	gjsonArr := duckDuckGoOut.Get("translated").Array()
	langout.OutputText = gjsonArr[0].String()
	langout.Engine = "duckduckgo"
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig
	if from == "auto" {
		langout.AutoDetect = duckDuckGoOut.Get("detected_language").String()
	}
	return langout, nil
}

func TranslateAll(to string, from string, query string) []LangOut {
	engines := []string{"reverso", "google", "libre", "watson", "mymemory", "yandex", "deepl", "duckduckgo"}
	langout := []LangOut{}
	for i := 0; i < len(engines); i++ {
		data, err := Translate(engines[i], to, from, query)
		if err == nil {
			langout = append(langout, data)
		}
	}
	return langout
}
