package libmozhi

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

// ============================================================================
// MOJI — the API behind mojidict.com (a Japanese-Chinese dictionary), with
// the public application id of its own web app. It gives the pitch accent
// and romaji, and senses in Chinese with a Japanese gloss. Like Jisho it
// never says "not found" (ぽぽぽぽぽ → ぽぽ様), so only words spelled or
// read exactly as asked are kept.
// ============================================================================

const mojiMaxEntries = 2

var (
	mojiExcerptPOS = regexp.MustCompile(`^\[([^\]]+)\]\s*`)
	mojiJLPT       = regexp.MustCompile(`^N[1-5]$`)
)

func mojiURL(string) string { return "https://www.mojidict.com/" }

func mojiLookup(word string) ([]JaDictEntry, error) {
	payload, _ := json.Marshal(map[string]string{
		"searchText": word, "langEnv": "zh-CN_ja", "_ClientVersion": "js3.4.1",
		"_ApplicationId": "E62VyFVLMiW7kvbtVq3p", "g_os": "PCWeb", "g_ver": "v4.8.8.20240829",
		"_InstallationId": "563b6bd3-e514-46fb-8557-d6138311a06c",
	})
	status, body, err := jaDictDo("POST", "https://api.mojidict.com/parse/functions/word-clickSearchV2", bytes.NewReader(payload), map[string]string{
		"Content-Type": "application/json", "Origin": "https://www.mojidict.com", "Referer": "https://www.mojidict.com/",
	})
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, errors.New("moji: HTTP " + strconv.Itoa(status))
	}
	return mojiParse(body, word)
}

func mojiParse(body []byte, word string) ([]JaDictEntry, error) {
	res := gjson.GetBytes(body, "result.result")
	if !gjson.ValidBytes(body) || !res.Get("word").IsArray() {
		return nil, errors.New("moji: unexpected answer")
	}
	// Senses: subdetails with the same relaId, a Chinese title and a Japanese one.
	type sense struct{ zh, ja string }
	senses := map[string]*sense{}
	for _, s := range res.Get("subdetails").Array() {
		id := s.Get("relaId").String()
		if senses[id] == nil {
			senses[id] = &sense{}
		}
		if s.Get("lang").String() == "ja" {
			senses[id].ja = s.Get("title").String()
		} else {
			senses[id].zh = s.Get("title").String()
		}
	}
	var out []JaDictEntry
	for _, w := range res.Get("word").Array() {
		if len(out) == mojiMaxEntries {
			break
		}
		spell, pron := w.Get("spell").String(), w.Get("pron").String()
		if spell != word && pron != word {
			continue
		}
		e := JaDictEntry{Word: spell, Romaji: w.Get("romaji").String(), Accent: w.Get("accent").String()}
		if pron != spell {
			e.Reading = pron
		}
		seen := map[string]bool{}
		for _, t := range strings.Split(w.Get("tags").String(), "#") {
			if mojiJLPT.MatchString(t) && !seen[t] {
				seen[t] = true
				e.Tags = append(e.Tags, "JLPT "+t)
			}
		}
		excerpt := w.Get("excerpt").String()
		pos := ""
		if m := mojiExcerptPOS.FindStringSubmatch(excerpt); m != nil {
			pos = m[1]
		}
		for _, id := range w.Get("subdetailsIds").Array() {
			if s := senses[id.String()]; s != nil && (s.zh != "" || s.ja != "") {
				var g []string
				for _, t := range []string{s.zh, s.ja} {
					if t != "" {
						g = append(g, t)
					}
				}
				e.Senses = append(e.Senses, JaDictSense{POS: pos, Glosses: g})
				pos = ""
			}
		}
		if len(e.Senses) == 0 { // words without subdetails: the excerpt
			if g := strings.TrimSpace(mojiExcerptPOS.ReplaceAllString(excerpt, "")); g != "" {
				e.Senses = append(e.Senses, JaDictSense{POS: pos, Glosses: strings.Split(g, "；")})
			}
		}
		out = append(out, e)
	}
	return out, nil
}
