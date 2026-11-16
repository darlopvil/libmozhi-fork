package libmozhi

import (
	"errors"
	"log"
	"runtime/debug"
	"time"
)

var (
	errTargetLang = errors.New("Target language code invalid")
	errSourceLang = errors.New("Source language code invalid")
)

func hasLang(list []List, id string) bool {
	for _, entry := range list {
		if entry.Id == id {
			return true
		}
	}

	return false
}

func validateLanguagePair(sourceLangs []List, targetLangs []List, from string, to string) error {
	if !hasLang(targetLangs, to) {
		return errTargetLang
	}
	if !hasLang(sourceLangs, from) {
		return errSourceLang
	}

	return nil
}

type List struct {
	Name string
	Id   string
}
type LangOut struct {
	Engine                     string              `json:"engine"`
	AutoDetect                 string              `json:"detected"`
	OutputText                 string              `json:"translated-text"`
	SourceLang                 string              `json:"source_language"`
	TargetLang                 string              `json:"target_language"`
	WordChoices                []WordChoices       `json:"word_choices"`
	SourceTransliteration      string              `json:"source_transliteration"`
	TargetTransliteration      string              `json:"target_transliteration"`
	SourceSynonyms             []string            `json:"source_synonyms"`
	TargetSynonyms             []string            `json:"target_synonyms"`
	SourceAntonyms             []string            `json:"source_antonyms"`
	TargetAntonyms             []string            `json:"target_antonyms"`
	SourceEquivalentTargetLang map[string][]string `json:"source_equivalent_target_lang"`
	TargetEquivalentSourceLang map[string][]string `json:"target_equivalent_source_lang"`
}
type WordChoices struct {
	Word       string `json:"word"`
	Definition string `json:"definition"`
	// Example is for google alone; since it does not have a linked example for source & target
	Example        string   `json:"example"`
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
	} else if engine == "yandex" {
		data = langListYandex(listType)
	} else if engine == "mymemory" {
		data = langListMyMemory(listType)
	} else if engine == "duckduckgo" {
		data = langListDuckDuckGo(listType)
	} else if engine == "gemini" {
		data = langListGemini(listType)
	} else if engine == "textra" {
		data = langListTexTra(listType)
	} else if engine == "groq" {
		data = langListGroq(listType)
	} else if engine == "gptoss" {
		data = langListOpenrouterGptoss(listType)
	} else if engine == "gemma" {
		data = langListOpenrouterGemma(listType)
	} else if engine == "nemotron" {
		data = langListOpenrouterNemotron(listType)
	} else if engine == "mistral" {
		data = langListMistral(listType)
	} else {
		return []List{}, errors.New("Engine does not exist or has been disabled.")
	}
	return data, nil
}

// Translate translates text with engine. A panic in the engine becomes an
// ordinary error (Fiber's recover middleware only protects the request
// goroutine, not the ones All mode starts), and every result is recorded
// for the engine's health (health.go).
func Translate(engine string, to string, from string, text string) (out LangOut, err error) {
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("libmozhi: engine %q panicked: %v\n%s", engine, r, debug.Stack())
			out, err = LangOut{}, errEnginePanic
		}
		recordHealth(engine, "translate", start, err)
	}()
	return translate(engine, to, from, text)
}

// General function to translate stuff so there is no need for a huge if-block everywhere
func translate(engine string, to string, from string, text string) (LangOut, error) {
	// Accept any spelling of a language this engine knows (es / es-ES,
	// zh-CN / zh-Hans, nb / no): see langs_resolve.go.
	from = resolveLangFor(engine, "sl", from)
	to = resolveLangFor(engine, "tl", to)
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
	} else if engine == "yandex" {
		data, err = translateYandex(to, from, text)
	} else if engine == "mymemory" {
		data, err = translateMyMemory(to, from, text)
	} else if engine == "duckduckgo" {
		data, err = translateDuckDuckGo(to, from, text)
	} else if engine == "gemini" {
		data, err = translateGemini(to, from, text)
	} else if engine == "textra" {
		data, err = translateTexTra(to, from, text)
	} else if engine == "groq" {
		data, err = translateGroq(to, from, text)
	} else if engine == "gptoss" {
		data, err = translateOpenrouterGptoss(to, from, text)
	} else if engine == "gemma" {
		data, err = translateOpenrouterGemma(to, from, text)
	} else if engine == "nemotron" {
		data, err = translateOpenrouterNemotron(to, from, text)
	} else if engine == "mistral" {
		data, err = translateMistral(to, from, text)
	} else {
		return LangOut{}, errors.New("Engine does not exist or has been disabled.")
	}
	if err != nil {
		return LangOut{}, err
	}
	return data, nil
}

// TTS synthesizes text with engine; every result is recorded for the
// engine's health (health.go).
func TTS(engine string, lang string, text string) ([]byte, error) {
	start := time.Now()
	data, err := tts(engine, lang, text)
	// TexTra allows about 3 requests per 2 minutes, shared with its
	// translations: when no slot is free, Google reads the text instead.
	var limit textraLimitError
	if engine == "textra" && errors.As(err, &limit) {
		gl := lang
		if gl == "zh" {
			gl = "zh-CN"
		}
		engine, start = "google", time.Now()
		data, err = tts(engine, gl, text)
	}
	recordHealth(engine, "tts", start, err)
	return data, err
}

func tts(engine string, lang string, text string) ([]byte, error) {
	var err error
	var data []byte
	if engine == "google" {
		data, err = ttsGoogle(lang, text)
	} else if engine == "reverso" {
		data, err = ttsReverso(lang, text)
	} else if engine == "textra" {
		data, err = ttsTexTra(lang, text)
	} else if engine == "groq" {
		data, err = ttsGroq(lang, text)
	} else {
		return []byte(""), errors.New("Engine does not exist and/or doesn't support TTS and/or has been disabled.")
	}
	if err != nil {
		return []byte(""), err
	}
	return data, nil
}
