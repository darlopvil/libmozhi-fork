package libmozhi

import (
	"errors"
)

type List struct {
	Name string
	Id   string
}
type LangOut struct {
	Engine      string        `json:"engine"`
	AutoDetect  string        `json:"detected"`
	OutputText  string        `json:"translated-text"`
	SourceLang  string        `json:"source_language"`
	TargetLang  string        `json:"target_language"`
	WordChoices []WordChoices `json:"word_choices"`
	Transliteration string	`json:"transliteration"`
}
type WordChoices struct {
	Word           string   `json:"word"`
	ExamplesSource []string `json:"examples_source"`
	ExamplesTarget []string `json:"examples_target"`
}

func LangList(engine string, listType string) ([]List, error) {
	var data []List
	if listType != "sl" && listType != "tl" {
		return []List{}, errors.New("list type invalid: either give tl for target languages or sl for source languages.")
	}
	if engine == "google" {
		data = langListGoogle(listType)
	} else if engine == "libre" {
		data = langListLibreTranslate(listType)
	} else if engine == "reverso" {
		data = langListReverso(listType)
	} else if engine == "deepl" {
		data = langListDeepl(listType)
	} else if engine == "watson" {
		data = langListWatson(listType)
	} else if engine == "yandex" {
		data = langListYandex(listType)
	} else if engine == "mymemory" {
		data = langListMyMemory(listType)
	} else if engine == "duckduckgo" {
		data = langListDuckDuckGo(listType)
	} else {
		return []List{}, errors.New("Engine does not exist or has been disabled.")
	}
	return data, nil
}

// General function to translate stuff so there is no need for a huge if-block everywhere
func Translate(engine string, to string, from string, text string) (LangOut, error) {
	var err error
	var data LangOut
	if engine == "google" {
		data, err = translateGoogle(to, from, text)
	} else if engine == "libre" {
		data, err = translateLibreTranslate(to, from, text)
	} else if engine == "reverso" {
		data, err = translateReverso(to, from, text)
	} else if engine == "deepl" {
		data, err = translateDeepl(to, from, text)
	} else if engine == "watson" {
		data, err = translateWatson(to, from, text)
	} else if engine == "yandex" {
		data, err = translateYandex(to, from, text)
	} else if engine == "mymemory" {
		data, err = translateMyMemory(to, from, text)
	} else if engine == "duckduckgo" {
		data, err = translateDuckDuckGo(to, from, text)
	} else {
		return LangOut{}, errors.New("Engine does not exist or has been disabled.")
	}
	if err != nil {
		return LangOut{}, err
	}
	return data, nil
}

func TTS(engine string, lang string, text string) ([]byte, error) {
	var err error
	var data []byte
	if engine == "google" {
		data, err = ttsGoogle(lang, text)
	} else if engine == "reverso" {
		data, err = ttsReverso(lang, text)
	} else {
		return []byte(""), errors.New("Engine does not exist and/or doesn't support TTS and/or has been disabled.")
	}
	if err != nil {
		return []byte(""), err
	}
	return data, nil
}
