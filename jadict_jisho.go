package libmozhi

import (
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

// ============================================================================
// JISHO — jisho.org's public JSON API (JMdict data). It never answers "not
// found": an unknown word gets loose matches (ぽぽぽぽぽ → 波 "Poland"), so
// only entries whose spelling or reading is the word itself are kept.
// ============================================================================

const jishoMaxEntries, jishoMaxSenses = 3, 5

func jishoURL(word string) string { return "https://jisho.org/search/" + url.PathEscape(word) }

func jishoLookup(word string) ([]JaDictEntry, error) {
	status, body, err := jaDictDo("GET", "https://jisho.org/api/v1/search/words?keyword="+url.QueryEscape(word), nil, nil)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, errors.New("jisho: HTTP " + strconv.Itoa(status))
	}
	return jishoParse(body, word)
}

func jishoParse(body []byte, word string) ([]JaDictEntry, error) {
	if !gjson.ValidBytes(body) || !gjson.GetBytes(body, "data").IsArray() {
		return nil, errors.New("jisho: unexpected answer")
	}
	var out []JaDictEntry
	for _, d := range gjson.GetBytes(body, "data").Array() {
		if len(out) == jishoMaxEntries {
			break
		}
		exact := false
		for _, j := range d.Get("japanese").Array() {
			if j.Get("word").String() == word || j.Get("reading").String() == word {
				exact = true
			}
		}
		if !exact {
			continue
		}
		e := JaDictEntry{Word: d.Get("japanese.0.word").String(), Reading: d.Get("japanese.0.reading").String()}
		if e.Word == "" {
			e.Word, e.Reading = e.Reading, ""
		}
		if e.Reading == e.Word {
			e.Reading = ""
		}
		if d.Get("is_common").Bool() {
			e.Tags = append(e.Tags, "common")
		}
		for _, l := range d.Get("jlpt").Array() {
			e.Tags = append(e.Tags, "JLPT "+strings.ToUpper(strings.TrimPrefix(l.String(), "jlpt-")))
		}
		for _, s := range d.Get("senses").Array() {
			if len(e.Senses) == jishoMaxSenses {
				break
			}
			var pos []string
			for _, p := range s.Get("parts_of_speech").Array() {
				pos = append(pos, p.String())
			}
			if len(pos) == 1 && pos[0] == "Wikipedia definition" {
				continue
			}
			sense := JaDictSense{POS: strings.Join(pos, ", ")}
			for _, g := range s.Get("english_definitions").Array() {
				sense.Glosses = append(sense.Glosses, g.String())
			}
			var notes []string
			for _, n := range append(s.Get("tags").Array(), s.Get("info").Array()...) {
				notes = append(notes, n.String())
			}
			if len(notes) > 0 {
				sense.Glosses = append(sense.Glosses, "("+strings.Join(notes, "; ")+")")
			}
			if len(sense.Glosses) > 0 {
				e.Senses = append(e.Senses, sense)
			}
		}
		out = append(out, e)
	}
	return out, nil
}
