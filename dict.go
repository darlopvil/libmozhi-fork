package libmozhi

import (
	"github.com/tidwall/gjson"
	"net/url"
	"strconv"
	"strings"
)

// TODO: Remove all the repetitive code for source/target
func dictDataGoogle(text string, translated string, from string, to string) LangOut {
	var langout LangOut

	dictData := `[[["rPsWke","[[\"` + text + `\",\"` + from + `\",\"` + to + `\"],1]",null,"generic"]]]`
	escapedDictData := url.PathEscape(dictData)
	dictQuery := `f.req=` + escapedDictData
	dictOut, err := postRequest("https://translate.google.com/_/TranslateWebserverUi/data/batchexecute", []byte(dictQuery), "application/x-www-form-urlencoded")
	if err != nil {
		return LangOut{}
	}
	dictOut = strings.TrimPrefix(dictOut, `)]}'`)
	dictOut = strings.ReplaceAll(dictOut, `\\"`, `"`)
	dictOut = strings.ReplaceAll(dictOut, `\\\"`, `"`)
	dictInit := gjson.Get(dictOut, "0.2").String()
	equivalentArr := gjson.Get(dictInit, "0.5.0.0.1.#.0")
	tmpMap := make(map[string][]string)
	for i, eq := range equivalentArr.Array() {
		tmp := []string{}
		for _, word := range gjson.Get(dictInit, "0.5.0.0.1."+strconv.Itoa(i)+".2").Array() {
			tmp = append(tmp, word.String())
		}
		tmpMap[eq.String()] = tmp
	}
	langout.TargetEquivalentSourceLang = tmpMap
	for _, eg1 := range gjson.Get(dictInit, "0.1.0.#.1.0").Array() {
		var WordChoices WordChoices // WordChoices is the struct as well as var name
		WordChoices.Word = text
		WordChoices.Definition = eg1.Array()[0].String()
		WordChoices.Example = eg1.Array()[1].String()
		langout.WordChoices = append(langout.WordChoices, WordChoices)
		for _, syn := range eg1.Get("5.0.0.#.0").Array() {
			langout.SourceSynonyms = append(langout.SourceSynonyms, syn.String())
		}
	}

	dictData = `[[["rPsWke","[[\"` + translated + `\",\"` + to + `\",\"` + from + `\"],1]",null,"generic"]]]`
	escapedDictData = url.PathEscape(dictData)
	dictQuery = `f.req=` + escapedDictData
	dictOut, err = postRequest("https://translate.google.com/_/TranslateWebserverUi/data/batchexecute", []byte(dictQuery), "application/x-www-form-urlencoded")
	if err != nil {
		return LangOut{}
	}
	dictOut = strings.TrimPrefix(dictOut, `)]}'`)
	dictOut = strings.ReplaceAll(dictOut, `\\"`, `"`)
	dictOut = strings.ReplaceAll(dictOut, `\\\"`, `"`)
	dictInit = gjson.Get(dictOut, "0.2").String()
	equivalentArr = gjson.Get(dictInit, "0.5.0.0.1.#.0")
	tmpMap = make(map[string][]string)
	for i, eq := range equivalentArr.Array() {
		tmp := []string{}
		for _, word := range gjson.Get(dictInit, "0.5.0.0.1."+strconv.Itoa(i)+".2").Array() {
			tmp = append(tmp, word.String())
		}
		tmpMap[eq.String()] = tmp
	}
	langout.SourceEquivalentTargetLang = tmpMap
	for _, eg1 := range gjson.Get(dictInit, "0.1.0.#.1.0").Array() {
		var WordChoices WordChoices // WordChoices is the struct as well as var name
		WordChoices.Word = translated
		WordChoices.Definition = eg1.Array()[0].String()
		WordChoices.Example = eg1.Array()[1].String()
		langout.WordChoices = append(langout.WordChoices, WordChoices)
		for _, syn := range eg1.Get("5.0.0.#.0").Array() {
			langout.TargetSynonyms = append(langout.TargetSynonyms, syn.String())
		}
	}
	return langout
}
