package libmozhi

import (
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

// ============================================================================
// DUCKDUCKGO — translation through duckduckgo.com/translation.js.
// In blocking windows DuckDuckGo resets or stalls the TLS handshake on some
// connections and not on others, for any client (measured nov 2026: in the
// same minute one connection was reset and the next one answered 200). A
// healthy handshake takes ~0.1-0.2s, so connections that stall are dropped
// after 3s and every request is retried on a fresh connection.
// ============================================================================

var ddgTransport = func() *http.Transport {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.TLSHandshakeTimeout = 3 * time.Second
	return t
}()

var ddgHTTP = &http.Client{Timeout: 15 * time.Second, Transport: ddgTransport}

var ddgVqdRe = regexp.MustCompile(`vqd="([^"]*)"`)

// ddgDo sends a request built by newReq, retrying connection-level failures
// (reset, stalled handshake) up to 3 times on a fresh connection. newReq is
// called for every attempt so that a POST body can be sent again.
func ddgDo(newReq func() (*http.Request, error)) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(300 * time.Millisecond)
		}
		req, err := newReq()
		if err != nil {
			return nil, err
		}
		res, err := ddgHTTP.Do(req)
		if err == nil {
			return res, nil
		}
		lastErr = err
		ddgTransport.CloseIdleConnections()
	}
	return nil, lastErr
}

// ddgVqd fetches the vqd token DuckDuckGo requires on every translation call.
func ddgVqd() (string, error) {
	res, err := ddgDo(func() (*http.Request, error) {
		r, err := http.NewRequest("GET", "https://duckduckgo.com/?q=translate", nil)
		if err != nil {
			return nil, err
		}
		r.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36")
		return r, nil
	})
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", errors.New("duckduckgo vqd error: " + res.Status)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	match := ddgVqdRe.FindStringSubmatch(string(body))
	if len(match) < 2 {
		return "", errors.New("duckduckgo vqd token not found")
	}
	return match[1], nil
}

func translateDuckDuckGo(to string, from string, query string) (LangOut, error) {
	FromOrig := from
	ToOrig := to
	if to == "zh" {
		to = "zh-Hans"
	} else if from == "zh" {
		from = "zh-Hans"
	} else if from == "zh-TW" {
		from = "zh-Hant"
	} else if to == "zh-TW" {
		to = "zh-Hant"
	}
	// Validate against the original user-facing IDs for now. Once DDG lang lists
	// are fetched from upstream automatically, this should validate `from`/`to`
	// directly instead of the pre-normalized aliases.
	if err := validateLanguagePair(langListDuckDuckGo("sl"), langListDuckDuckGo("tl"), FromOrig, ToOrig); err != nil {
		return LangOut{}, err
	}
	vqd, err := ddgVqd()
	if err != nil {
		return LangOut{}, err
	}
	url := "https://duckduckgo.com/translation.js?vqd=" + vqd + "&query=translate&to=" + to
	if from != "auto" {
		url += "&from=" + from
	}
	res, err := ddgDo(func() (*http.Request, error) {
		r, err := http.NewRequest("POST", url, strings.NewReader(query))
		if err != nil {
			return nil, err
		}
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0")
		r.Header.Set("Accept", "application/json, text/plain, */*")
		return r, nil
	})
	if err != nil {
		return LangOut{}, err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return LangOut{}, err
	}
	duckDuckGoOut := string(body)
	if !gjson.Valid(duckDuckGoOut) {
		return LangOut{}, errors.New("instance has been rate limited")
	}
	gjsonArr := gjson.Get(duckDuckGoOut, "translated").Array()
	if len(gjsonArr) == 0 {
		return LangOut{}, errors.New("duckduckgo: no translation in response")
	}
	var langout LangOut
	langout.OutputText = gjsonArr[0].String()
	langout.Engine = "duckduckgo"
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig
	if from == "auto" {
		langout.AutoDetect = gjson.Get(duckDuckGoOut, "detected_language").String()
	}
	return langout, nil
}
