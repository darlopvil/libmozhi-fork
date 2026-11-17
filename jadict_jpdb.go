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
// JPDB — jpdb.io's search page (HTML; its API needs an account key). An
// exact word gets one "details" result, conjugated forms included ("食べた"
// → 食べる, "Conjugated as: past"); other words get a "search" list, where
// only entries spelled or read exactly as the word are kept (a made-up
// word gets guesses: ぽぽぽぽぽ → ぽい "Conjugated as: stem").
// ============================================================================

const jpdbMaxEntries, jpdbMaxSenses, jpdbMaxExamples = 3, 5, 1

var jpdbSenseNumber = regexp.MustCompile(`^\d+\.\s*`)

func jpdbURL(word string) string {
	return "https://jpdb.io/search?q=" + url.QueryEscape(word) + "&lang=english"
}

func jpdbLookup(word string) ([]JaDictEntry, error) {
	status, body, err := jaDictDo("GET", jpdbURL(word), nil, nil)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, errors.New("jpdb: HTTP " + strconv.Itoa(status))
	}
	return jpdbParse(body, word)
}

func jpdbParse(body []byte, word string) ([]JaDictEntry, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if doc.Find(".results").Length() == 0 {
		return nil, errors.New("jpdb: unexpected page")
	}
	details := doc.Find(".results.details").Length() > 0
	var out []JaDictEntry
	seen := map[string]bool{}
	doc.Find(".results .result.vocabulary").EachWithBreak(func(_ int, r *goquery.Selection) bool {
		var e JaDictEntry
		e.Word, e.Reading = rubyWordReading(r.Find(".primary-spelling ruby"))
		if e.Word == "" || seen[e.Word+"|"+e.Reading] {
			return true
		}
		if !details && e.Word != word && e.Reading != word && !(e.Reading == "" && e.Word == word) {
			return true
		}
		seen[e.Word+"|"+e.Reading] = true
		if note := jaDictClean(r.Children().First().Text()); strings.HasPrefix(note, "Conjugated as") {
			e.Note = strings.TrimSpace(strings.TrimSuffix(note, "?"))
		}
		r.Find(".tags .tag").Each(func(_ int, t *goquery.Selection) {
			if s := jaDictClean(t.Text()); s != "" {
				e.Tags = append(e.Tags, s)
			}
		})
		pos := ""
		r.Find(".subsection-meanings .subsection").First().Children().Each(func(_ int, c *goquery.Selection) {
			switch {
			case c.HasClass("part-of-speech"):
				var parts []string
				c.Children().Each(func(_ int, p *goquery.Selection) { parts = append(parts, jaDictClean(p.Text())) })
				pos = strings.Join(parts, ", ")
			case c.HasClass("description") && len(e.Senses) < jpdbMaxSenses:
				text := c.Clone()
				text.Find(".sense-example").Remove()
				gloss := jpdbSenseNumber.ReplaceAllString(nodeTexts(text), "")
				if gloss == "" {
					return
				}
				sense := JaDictSense{POS: pos, Glosses: strings.Split(gloss, "; ")}
				pos = "" // shown once, on the first sense it applies to
				c.Find(".sense-example-text").Each(func(_ int, x *goquery.Selection) {
					if len(sense.Examples) < jpdbMaxExamples {
						sense.Examples = append(sense.Examples, JaDictExample{
							Japanese:    rubyBaseText(x.Find(".jp")),
							Translation: jaDictClean(x.Find(".en").Text()),
						})
					}
				})
				e.Senses = append(e.Senses, sense)
			}
		})
		out = append(out, e)
		return len(out) < jpdbMaxEntries
	})
	return out, nil
}

// rubyWordReading reads <ruby> elements with furigana: 食<rt>た</rt>べ<rt></rt>る
// → word 食べる, reading たべる. The reading is "" when it equals the word.
func rubyWordReading(ruby *goquery.Selection) (word, reading string) {
	var w, r strings.Builder
	var last string
	ruby.Contents().Each(func(_ int, n *goquery.Selection) {
		node := n.Get(0)
		switch {
		case node.Type == html.TextNode:
			last = node.Data
			w.WriteString(last)
			r.WriteString(last)
		case node.Data == "rt":
			if rt := n.Text(); rt != "" {
				// The furigana replaces the text it annotates.
				s := r.String()
				r.Reset()
				r.WriteString(strings.TrimSuffix(s, last) + rt)
			}
			last = ""
		default:
			last = n.Text()
			w.WriteString(last)
			r.WriteString(last)
		}
	})
	word, reading = strings.TrimSpace(w.String()), strings.TrimSpace(r.String())
	if reading == word {
		reading = ""
	}
	return word, reading
}

// nodeTexts is the text of s, keeping a space between words that only
// markup separates.
func nodeTexts(s *goquery.Selection) string {
	var parts []string
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			parts = append(parts, n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	for _, n := range s.Nodes {
		walk(n)
	}
	return jaDictJoin(parts)
}

// rubyBaseText is the text of s without its furigana.
func rubyBaseText(s *goquery.Selection) string {
	c := s.Clone()
	c.Find("rt").Remove()
	return jaDictClean(c.Text())
}
