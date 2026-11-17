package libmozhi

import (
	"bytes"
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// ============================================================================
// WEBLIO — weblio.jp's word page (HTML), Japanese monolingual dictionaries.
// One page holds many dictionaries, and one dictionary block can hold
// several headwords (デジタル大辞泉 lists 内, 化, いえ【家】, か【家】... in a
// row), so every headword (h2.midashigo) is read with what follows it up to
// the next one. Only real dictionaries are used (weblioDicts): novels titled
// like the word, names, example sentences or Wikipedia are skipped. An
// unknown word answers 404.
// ============================================================================

const weblioMaxSenses, weblioMaxLines, weblioMaxRunes = 6, 4, 200

// weblioDicts: the dictionaries read, by the class of their content, best
// first: デジタル大辞泉, 実用日本語表現辞典, the conjugation dictionary
// (食べた → 食べる) and the kanji dictionary.
var weblioDicts = []string{"Sgkdj", "Jtnhj", "Nhgkt", "Knjjn"}

var (
	weblioHeadKanji = regexp.MustCompile(`【([^】]*)】`)
	weblioReading   = regexp.MustCompile(`^読み方：\s*(\S+)\s*`)
)

func weblioURL(word string) string { return "https://www.weblio.jp/content/" + url.PathEscape(word) }

func weblioLookup(word string) ([]JaDictEntry, error) {
	status, body, err := jaDictDo("GET", weblioURL(word), nil, nil)
	if err != nil {
		return nil, err
	}
	if status == 404 {
		return nil, nil
	}
	if status != 200 {
		return nil, errors.New("weblio: HTTP " + strconv.Itoa(status))
	}
	return weblioParse(body, word)
}

// weblioHeadMatches: the headword is the word itself, or lists it among its
// spellings (いえ〔いへ〕【家】, け【化／仮／▽花／家／華】).
func weblioHeadMatches(h *goquery.Selection, word string) bool {
	if t, _ := h.Attr("title"); t == word {
		return true
	}
	for _, m := range weblioHeadKanji.FindAllStringSubmatch(h.Text(), -1) {
		for _, s := range strings.Split(m[1], "／") {
			if strings.TrimSpace(strings.NewReplacer("▽", "", "▼", "", "×", "").Replace(s)) == word {
				return true
			}
		}
	}
	return false
}

// weblioDict returns the class of the dictionary whose content follows a
// headword, or "" when it is not one of weblioDicts.
func weblioDict(body *goquery.Selection) string {
	for _, d := range weblioDicts {
		if body.Filter("."+d).Length() > 0 {
			return d
		}
	}
	return ""
}

func weblioParse(body []byte, word string) ([]JaDictEntry, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if doc.Find(".kijiWrp").Length() == 0 {
		return nil, errors.New("weblio: unexpected page")
	}
	type candidate struct {
		head, body *goquery.Selection
		rank       int // index in weblioDicts; numbered 大辞泉 entries come first
	}
	var best *candidate
	doc.Find("h2.midashigo").Each(func(_ int, h *goquery.Selection) {
		if !weblioHeadMatches(h, word) {
			return
		}
		b := h.NextUntil("h2.midashigo")
		d := weblioDict(b)
		if d == "" {
			return
		}
		rank := 1
		for i, x := range weblioDicts {
			if x == d {
				rank = i + 1
			}
		}
		if d == "Sgkdj" && weblioNumbered(b) {
			rank = 0 // a real definition, not a redirect or a novel's summary
		}
		if best == nil || rank < best.rank {
			best = &candidate{h, b, rank}
		}
	})
	if best == nil {
		return nil, nil
	}
	e := JaDictEntry{Word: jaDictClean(best.head.Text())}
	if best.rank <= 1 {
		weblioSenses(&e, best.body, word)
	}
	if len(e.Senses) == 0 {
		// Dictionaries without numbered senses: the entry's text.
		t := blockText(best.body)
		if m := weblioReading.FindStringSubmatch(t); m != nil {
			if m[1] != word {
				e.Reading = m[1]
			}
			t = t[len(m[0]):]
		}
		if t = jaDictCut(strings.TrimSpace(t), weblioMaxRunes); t != "" {
			e.Senses = []JaDictSense{{Glosses: []string{t}}}
		}
		e.Note = ""
	}
	if len(e.Senses) == 0 {
		return nil, nil
	}
	return []JaDictEntry{e}, nil
}

// weblioNumbered: the entry has numbered senses (<p><b>１</b> ...).
func weblioNumbered(body *goquery.Selection) bool {
	n := false
	body.Find("p").EachWithBreak(func(_ int, p *goquery.Selection) bool {
		n = p.Children().First().Is("b")
		return !n
	})
	return n
}

// weblioSenses reads a デジタル大辞泉 entry: reading, grammar line and
// numbered senses with their sub-senses (㋐ ㋑ ...).
func weblioSenses(e *JaDictEntry, body *goquery.Selection, word string) {
	var cur *JaDictSense
	body.Find("p").Each(func(_ int, p *goquery.Selection) {
		line := jaDictCut(jaDictClean(p.Text()), weblioMaxRunes)
		switch {
		case line == "":
		case strings.HasPrefix(line, "読み方："):
			if r := strings.TrimSpace(strings.TrimPrefix(line, "読み方：")); r != word {
				e.Reading = r
			}
		case p.Children().First().Is("b") && strings.HasPrefix(line, jaDictClean(p.Children().First().Text())):
			// A numbered sense: "１ 食物をかんで、のみこむ。"
			if len(e.Senses) < weblioMaxSenses {
				e.Senses = append(e.Senses, JaDictSense{Glosses: []string{line}})
				cur = &e.Senses[len(e.Senses)-1]
			} else {
				cur = nil
			}
		case cur != nil:
			if len(cur.Glosses) == 1 && len([]rune(cur.Glosses[0])) <= 2 {
				cur.Glosses[0] += " " + line // "３" alone, then its first sub-sense
			} else if len(cur.Glosses) < weblioMaxLines {
				cur.Glosses = append(cur.Glosses, line)
			}
		case len(e.Senses) == 0 && e.Note == "":
			e.Note = line // the grammar line before the senses
		}
	})
	if len(e.Senses) == 0 && e.Note != "" {
		// One paragraph and no numbers: that paragraph is the sense.
		e.Senses, e.Note = []JaDictSense{{Glosses: []string{e.Note}}}, ""
	}
}

// weblioBlocks are the elements whose text starts a new line.
var weblioBlocks = map[string]bool{"div": true, "p": true, "br": true, "td": true, "th": true, "tr": true, "li": true, "table": true}

// blockText is the text of s with a space between block elements (Japanese
// text has no spaces of its own to keep them apart).
func blockText(s *goquery.Selection) string {
	var b strings.Builder
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		block := n.Type == html.ElementNode && weblioBlocks[n.Data]
		if block {
			b.WriteByte(' ')
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
		if block {
			b.WriteByte(' ')
		}
	}
	for _, n := range s.Nodes {
		walk(n)
	}
	return jaDictClean(b.String())
}
