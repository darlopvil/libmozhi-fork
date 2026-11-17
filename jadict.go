package libmozhi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

// ============================================================================
// JAPANESE DICTIONARY — word lookups beside the translation, whatever the
// engine: Jisho (readings, English senses, JLPT), jpdb (senses with
// examples, frequency, conjugation), MOJi (pitch accent, romaji, Chinese and
// Japanese senses) and Weblio (a Japanese monolingual dictionary). Each
// source lives in its own file (jadict_<source>.go) and none needs a key.
// Every lookup is recorded as kind "dict" in the health state.
// ============================================================================

// JaDictSense is one sense of a dictionary entry.
type JaDictSense struct {
	POS      string          `json:"pos,omitempty"`
	Glosses  []string        `json:"glosses"`
	Examples []JaDictExample `json:"examples,omitempty"`
}

// JaDictExample is an example sentence and its translation.
type JaDictExample struct {
	Japanese    string `json:"japanese"`
	Translation string `json:"translation,omitempty"`
}

// JaDictEntry is one headword of a source.
type JaDictEntry struct {
	Word    string        `json:"word"`
	Reading string        `json:"reading,omitempty"`
	Romaji  string        `json:"romaji,omitempty"`
	Accent  string        `json:"accent,omitempty"` // MOJi: pitch accent (⓪ flat, ① ② ... downstep after that mora)
	Note    string        `json:"note,omitempty"`   // jpdb: "Conjugated as: past"; Weblio: grammar line
	Tags    []string      `json:"tags,omitempty"`
	Senses  []JaDictSense `json:"senses"`
}

// JaDict is what one source answered for a word.
type JaDict struct {
	Source  string        `json:"source"`
	Name    string        `json:"name"`
	URL     string        `json:"url"`  // the full entry on the source's website
	Lang    string        `json:"lang"` // language of the senses: en, zh or ja
	Entries []JaDictEntry `json:"entries"`
	Error   string        `json:"error,omitempty"`
}

type jaDictSource struct {
	name, lang string
	url        func(word string) string
	lookup     func(word string) ([]JaDictEntry, error)
}

var jaDictSources = map[string]jaDictSource{
	"jisho":  {"Jisho", "en", jishoURL, jishoLookup},
	"jpdb":   {"jpdb", "en", jpdbURL, jpdbLookup},
	"moji":   {"MOJi", "zh", mojiURL, mojiLookup},
	"weblio": {"Weblio", "ja", weblioURL, weblioLookup},
}

// JaDictSources lists the dictionary sources, in display order.
func JaDictSources() []string { return []string{"jisho", "jpdb", "moji", "weblio"} }

// jaDictProbeWord is looked up by the health re-checks: every source has it.
const jaDictProbeWord = "食べる"

// jaDictMaxRunes: longer texts are sentences, not dictionary words.
const jaDictMaxRunes = 15

// JaDictWord returns text trimmed, and whether it looks like one Japanese
// word or short phrase worth looking up: at most 15 characters, at least one
// kana or kanji, and no sentence marks or line breaks.
func JaDictWord(text string) (string, bool) {
	w := strings.TrimSpace(text)
	if w == "" || utf8.RuneCountInString(w) > jaDictMaxRunes || strings.ContainsAny(w, "\n。！？!?") {
		return w, false
	}
	for _, r := range w {
		if unicode.In(r, unicode.Hiragana, unicode.Katakana, unicode.Han) {
			return w, true
		}
	}
	return w, false
}

// JaDictLookup looks word up in one source. Panics are recovered, and the
// outcome is recorded in the health state.
func JaDictLookup(source, word string) (out JaDict, err error) {
	src, ok := jaDictSources[source]
	if !ok {
		return JaDict{}, errors.New("unknown dictionary " + source)
	}
	out = JaDict{Source: source, Name: src.name, URL: src.url(word), Lang: src.lang}
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s: %w: %v", source, errEnginePanic, r)
		}
		if err != nil {
			out.Entries, out.Error = nil, err.Error()
		}
		if out.Entries == nil {
			out.Entries = []JaDictEntry{} // [] in JSON, not null
		}
		recordHealth(source, "dict", start, err)
	}()
	out.Entries, err = src.lookup(word)
	return out, err
}

// JapaneseDictionary looks word up in every given source at once, and
// returns their answers in the same order.
func JapaneseDictionary(word string, sources []string) []JaDict {
	out := make([]JaDict, len(sources))
	var wg sync.WaitGroup
	for i, s := range sources {
		wg.Add(1)
		go func(i int, s string) {
			defer wg.Done()
			out[i], _ = JaDictLookup(s, word)
		}(i, s)
	}
	wg.Wait()
	return out
}

// jaDictProbe is the health re-check of a source: the probe word must have
// entries, so a page whose layout changed shows up as failing.
func jaDictProbe(source string) {
	src, ok := jaDictSources[source]
	if !ok {
		return
	}
	start := time.Now()
	entries, err := func() (e []JaDictEntry, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%s: %w: %v", source, errEnginePanic, r)
			}
		}()
		return src.lookup(jaDictProbeWord)
	}()
	if err == nil && len(entries) == 0 {
		err = errors.New(source + ": no entries for " + jaDictProbeWord + " (page layout changed?)")
	}
	recordHealth(source, "dict", start, err)
}

var jaDictHTTP = &http.Client{Timeout: 10 * time.Second}

const jaDictUA = "Mozilla/5.0 (X11; Linux x86_64; rv:140.0) Gecko/20100101 Firefox/140.0"

// jaDictDo sends a request with a browser User-Agent and returns the status
// and the body (at most 4 MB).
func jaDictDo(method, url string, body io.Reader, headers map[string]string) (int, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("User-Agent", jaDictUA)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := jaDictHTTP.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	return res.StatusCode, b, err
}

// jaDictClean collapses whitespace.
func jaDictClean(s string) string { return strings.Join(strings.Fields(s), " ") }

// jaDictJoin joins the texts of neighbouring nodes, with a space between two
// words that were only split by markup ("family" + "after a surname").
func jaDictJoin(parts []string) string {
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		if b.Len() > 0 {
			prev, _ := utf8.DecodeLastRuneInString(b.String())
			next, _ := utf8.DecodeRuneInString(p)
			if jaDictLatin(prev) && (jaDictLatin(next) || next == '(') {
				b.WriteByte(' ')
			}
		}
		b.WriteString(p)
	}
	return jaDictClean(b.String())
}

func jaDictLatin(r rune) bool { return r < 0x3000 && (unicode.IsLetter(r) || unicode.IsDigit(r)) }

// jaDictCut shortens s to max runes.
func jaDictCut(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max]) + "…"
}
