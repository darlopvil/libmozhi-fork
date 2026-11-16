package libmozhi

import "strings"

// ============================================================================
// LANGUAGE CODE MATCHING — every engine names some languages its own way
// (MyMemory: es-ES, en-GB...; Google: zh-CN; DuckDuckGo/DeepL: nb vs no).
// A code chosen for one engine is matched to the equivalent id of another,
// so a choice survives engine switches and All mode can call every engine
// with the code it understands.
// ============================================================================

// langRegionPrefs: preferred variant when an engine only has regional codes
// for a language. Otherwise "<lang>-<LANG>" (es-ES, fr-FR...) is preferred.
var langRegionPrefs = map[string][]string{
	"en": {"en-gb", "en-us"},
	"zh": {"zh-cn", "zh-tw"},
	"sr": {"sr-latn"},
}

// langCanon lower-cases a code and folds known synonyms.
func langCanon(code string) string {
	c := strings.ToLower(code)
	switch c {
	case "zh-hans":
		return "zh-cn"
	case "zh-hant":
		return "zh-tw"
	case "nb":
		return "no"
	}
	return c
}

func langBase(code string) string {
	return strings.SplitN(langCanon(code), "-", 2)[0]
}

// ResolveLang returns the id in list that stands for code: the same id
// (case-insensitive or a known synonym), else the same language under
// another region or script (es-ES -> es, es -> es-ES, zh -> zh-CN), or ""
// when the list doesn't have that language.
func ResolveLang(list []List, code string) string {
	if code == "" {
		return ""
	}
	want := langCanon(code)
	for _, l := range list {
		if langCanon(l.Id) == want {
			return l.Id
		}
	}
	base := langBase(code)
	var candidates []string
	for _, l := range list {
		if langBase(l.Id) == base {
			candidates = append(candidates, l.Id)
		}
	}
	if len(candidates) == 0 {
		return ""
	}
	prefs := append([]string{base}, langRegionPrefs[base]...)
	prefs = append(prefs, base+"-"+base)
	for _, p := range prefs {
		for _, c := range candidates {
			if langCanon(c) == p {
				return c
			}
		}
	}
	return candidates[0]
}

// resolveLangFor matches code to engine's own list ("sl" or "tl"), keeping
// it unchanged when there is no equivalent (the engine then reports it).
func resolveLangFor(engine, listType, code string) string {
	list, err := LangList(engine, listType)
	if err != nil {
		return code
	}
	if id := ResolveLang(list, code); id != "" {
		return id
	}
	return code
}
