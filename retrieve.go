package utils

//import (
//	"github.com/gocolly/colly"
//	"os"
//	"strings"
//)
//
//type List struct {
//	Name string
//	Id   string
//}
//// To get new language lists for google yandex and libertranslate.
//func LangListGoogle(listType string) []List {
//	UserAgent, ok := os.LookupEnv("MOZHI_USER_AGENT")
//	if !ok {
//		UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"
//	}
//	sc := colly.NewCollector(colly.AllowedDomains("translate.google.com"), colly.UserAgent(UserAgent))
//	var ListData []List
//	sc.OnHTML("div.root-container", func(e *colly.HTMLElement) {
//		e.ForEach("div.language-item", func(i int, el *colly.HTMLElement) {
//			var ListInfo List
//			ListInfo.Name = el.ChildText("a")
//			if listType == "tl" {
//				ListInfo.Id = strings.TrimPrefix(strings.TrimSuffix(el.ChildAttr("a", "href"), "&hl=en-US"), "./m?sl=en&tl=")
//			} else {
//				ListInfo.Id = strings.TrimPrefix(strings.TrimSuffix(el.ChildAttr("a", "href"), "&tl=en&hl=en-US"), "./m?sl=")
//			}
//			ListData = append(ListData, ListInfo)
//		})
//	})
//	url := "https://translate.google.com/m?sl=en&tl=en&hl=en-US&mui=" + listType
//	sc.Visit(url)
//	return ListData
//}
//func LangListLibreTranslate(listType string) []List {
//	// TODO: Make it configurable
//	libreTranslateOut := PostRequest("https://translate.argosopentech.com/languages", []byte(""))
//	gjsonArr := libreTranslateOut.Array()
//	var ListData []List
//	for _, r := range gjsonArr {
//		code := r.Get("code").String()
//		name := r.Get("name").String()
//
//		ListData = append(ListData, List{Id: code, Name: name})
//	}
//	if listType == "sl" {
//		auto := List{
//			Id:   "auto",
//			Name: "Detect Language",
//		}
//		ListData = append(ListData, auto)
//	}
//	return ListData
//}
//func LangListYandex(listType string) []List {
//	UserAgent, ok := os.LookupEnv("MOZHI_USER_AGENT")
//	if !ok {
//		UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"
//	}
//	sc := colly.NewCollector(colly.AllowedDomains("localhost:8000"), colly.UserAgent(UserAgent))
//	var ListData []List
//	sc.OnHTML("table.langs", func(e *colly.HTMLElement) {
//		if listType == "sl" {
//		e.ForEach("select[name='from'] option", func(i int, el *colly.HTMLElement) {
//			var ListInfo List
//			ListInfo.Name = el.Text
//			ListInfo.Id = el.Attr("value")
//			ListData = append(ListData, ListInfo)
//		})
//	} else {
//				e.ForEach("select[name='to'] option", func(i int, el *colly.HTMLElement) {
//			var ListInfo List
//			ListInfo.Name = el.Text
//			ListInfo.Id = el.Attr("value")
//			ListData = append(ListData, ListInfo)
//		})
//
//		}
//	})
//	// Static version of translate.yandex.net nojs/mobile version
//	url := "http://localhost:8000/index.html"
//	sc.Visit(url)
//	return ListData
//}
//
