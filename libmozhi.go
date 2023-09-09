package utils

import (
	"errors"
	"os"
)

type List struct {
	Name string
	Id   string
}
type LangOut struct {
	Engine     string `json:"engine"`
	AutoDetect string `json:"detected"`
	OutputText string `json:"translated-text"`
	SourceLang string `json:"source_language"`
	TargetLang string `json:"target_language"`
}

func envTrueNoExist(env string) bool {
	if _, ok := os.LookupEnv(env); ok == false || os.Getenv(env) == "true" {
		return true
	}
	return false
}

func LangList(engine string, listType string) ([]List, error) {
	var data []List
	if listType != "sl" && listType != "tl" {
		return []List{}, errors.New("list type invalid: either give tl for target languages or sl for source languages.")
	}
	if engine == "google" && envTrueNoExist("MOZHI_GOOGLE_ENABLED") {
		data = langListGoogle(listType)
	} else if engine == "libre" && envTrueNoExist("MOZHI_LIBRETRANSLATE_ENABLED") {
		if envTrueNoExist("MOZHI_LIBRETRANSLATE_URL") {
			return []List{}, errors.New("Please set MOZHI_LIBRETRANSLATE_URL if you want to use libretranslate. Example: MOZHI_LIBRETRANSLATE_URL=https://lt.psf.lt")
		}
		data = langListLibreTranslate(listType)
	} else if engine == "reverso" && envTrueNoExist("MOZHI_REVERSO_ENABLED") {
		data = langListReverso(listType)
	} else if engine == "deepl" && envTrueNoExist("MOZHI_DEEPL_ENABLED") {
		data = langListDeepl(listType)
	} else if engine == "watson" && envTrueNoExist("MOZHI_WATSON_ENABLED") {
		data = langListWatson(listType)
	} else if engine == "yandex" && envTrueNoExist("MOZHI_YANDEX_ENABLED") {
		data = langListYandex(listType)
	} else if engine == "mymemory" && envTrueNoExist("MOZHI_MYMEMORY_ENABLED") {
		data = langListMyMemory(listType)
	} else if engine == "duckduckgo" && envTrueNoExist("MOZHI_DUCKDUCKGO_ENABLED") {
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
	if engine == "google" && envTrueNoExist("MOZHI_GOOGLE_ENABLED") {
		data, err = translateGoogle(to, from, text)
	} else if engine == "libre" && envTrueNoExist("MOZHI_LIBRETRANSLATE_ENABLED") {
		if os.Getenv("MOZHI_LIBRETRANSLATE_URL") == "" {
			return LangOut{}, errors.New("Please set MOZHI_LIBRETRANSLATE_URL if you want to use libretranslate. Example: MOZHI_LIBRETRANSLATE_URL=https://lt.psf.lt")
		}
		data, err = translateLibreTranslate(to, from, text)
	} else if engine == "reverso" && envTrueNoExist("MOZHI_REVERSO_ENABLED") {
		data, err = translateReverso(to, from, text)
	} else if engine == "deepl" && envTrueNoExist("MOZHI_DEEPL_ENABLED") {
		data, err = translateDeepl(to, from, text)
	} else if engine == "watson" && envTrueNoExist("MOZHI_WATSON_ENABLED") {
		data, err = translateWatson(to, from, text)
	} else if engine == "yandex" && envTrueNoExist("MOZHI_YANDEX_ENABLED") {
		data, err = translateYandex(to, from, text)
	} else if engine == "mymemory" && envTrueNoExist("MOZHI_MYMEMORY_ENABLED") {
		data, err = translateMyMemory(to, from, text)
	} else if engine == "duckduckgo" && envTrueNoExist("MOZHI_DUCKDUCKGO_ENABLED") {
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
	if engine == "google" && envTrueNoExist("MOZHI_GOOGLE_ENABLED") {
		data, err = ttsGoogle(lang, text)
	} else if engine == "reverso" && envTrueNoExist("MOZHI_REVERSO_ENABLED") {
		data, err = ttsReverso(lang, text)
	} else {
		return []byte(""), errors.New("Engine does not exist and/or doesn't support TTS and/or has been disabled.")
	}
	if err != nil {
		return []byte(""), err
	}
	return data, nil
}
