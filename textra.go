package libmozhi

import (
	"errors"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/gjson"
)

// ============================================================================
// TEXTRA (NICT) — translation, dictionary lookup and TTS.
// TexTra allows about 3 requests per ~2 minutes for the whole account, shared
// by every API (MT, lookup, TTS): the next one answers code 505 (measured nov
// 2026: a slot frees 98-119s after it was used; token requests don't count).
// Mozhi keeps its own count so it never spends a request it knows will
// bounce, and spends the scarce slots by priority: translation and TTS first,
// the optional dictionary lookup only when a slot is left over.
// ============================================================================

const (
	textraSlots  = 3
	textraWindow = 120 * time.Second
)

// textraLimitError is returned instead of calling TexTra when no request
// slot is free. TTS falls back to Google on it (see TTS in libmozhi.go).
type textraLimitError struct{ wait time.Duration }

func (e textraLimitError) Error() string {
	return "textra: request limit reached (TexTra allows about 3 requests per 2 minutes across all its APIs), try again in " +
		strconv.Itoa(int((e.wait+time.Second-1)/time.Second)) + "s"
}

var textraSlotState struct {
	sync.Mutex
	used []time.Time // when each recent request was sent
}

var textraTokenState struct {
	sync.Mutex
	token   string
	expires time.Time
}

// textraAcquire reserves one request slot, leaving at least keep slots free
// for higher-priority calls.
func textraAcquire(keep int) error { return textraTake(1, keep) }

// textraTake reserves n request slots at once (all or nothing), leaving at
// least keep slots free for higher-priority calls.
func textraTake(n, keep int) error {
	s := &textraSlotState
	s.Lock()
	defer s.Unlock()
	now := time.Now()
	recent := []time.Time{}
	for _, t := range s.used {
		if now.Sub(t) < textraWindow {
			recent = append(recent, t)
		}
	}
	s.used = recent
	if len(s.used)+n+keep > textraSlots {
		wait := textraWindow
		if len(s.used) > 0 {
			wait = textraWindow - now.Sub(s.used[0])
		}
		return textraLimitError{wait}
	}
	for i := 0; i < n; i++ {
		s.used = append(s.used, now)
	}
	return nil
}

// textraFree returns how many request slots are free right now.
func textraFree() int {
	s := &textraSlotState
	s.Lock()
	defer s.Unlock()
	used := 0
	for _, t := range s.used {
		if time.Since(t) < textraWindow {
			used++
		}
	}
	return textraSlots - used
}

// textraLimitHit records that TexTra refused a request for its limits anyway
// (e.g. requests made outside Mozhi): every slot counts as used from now.
func textraLimitHit() {
	s := &textraSlotState
	s.Lock()
	defer s.Unlock()
	now := time.Now()
	s.used = []time.Time{now, now, now}
}

func textraConfig() (key, secret, name, baseURL string, err error) {
	key = os.Getenv("MOZHI_TEXTRA_API_KEY")
	secret = os.Getenv("MOZHI_TEXTRA_API_SECRET")
	name = os.Getenv("MOZHI_TEXTRA_LOGIN_ID")
	if key == "" || secret == "" || name == "" {
		return "", "", "", "", errors.New("textra engine requires MOZHI_TEXTRA_API_KEY, _API_SECRET and _LOGIN_ID")
	}
	baseURL = os.Getenv("MOZHI_TEXTRA_BASE_URL")
	if baseURL == "" {
		baseURL = "https://mt-auto-minhon-mlt.ucri.jgn-x.jp"
	}
	return key, secret, name, baseURL, nil
}

// textraToken returns a cached OAuth2 access token. Tokens last 3600s
// (expires_in, measured); it is renewed a minute before it expires.
func textraToken(baseURL, key, secret string) (string, error) {
	s := &textraTokenState
	s.Lock()
	defer s.Unlock()
	if s.token != "" && time.Now().Before(s.expires) {
		return s.token, nil
	}
	res, err := textraHTTP.PostForm(baseURL+"/oauth2/token.php", url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {key},
		"client_secret": {secret},
	})
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	out, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	token := gjson.GetBytes(out, "access_token").String()
	if token == "" {
		// Say why: maintenance, an outage and a bad key need different
		// fixes. During maintenance, unscheduled ones included, TexTra
		// answers 503 with a page saying "ただいまメンテナンス中です".
		msg := "textra: no access token (token endpoint: " + res.Status
		if strings.Contains(string(out), "メンテナンス") {
			return "", errors.New("textra: under maintenance (token endpoint: " + res.Status + ")")
		}
		if d := gjson.GetBytes(out, "error_description").String(); d != "" {
			msg += ", " + d
		} else if e := gjson.GetBytes(out, "error").String(); e != "" {
			msg += ", " + e
		}
		return "", errors.New(msg + ")")
	}
	life := gjson.GetBytes(out, "expires_in").Int()
	if life <= 0 {
		life = 3600
	}
	s.token = token
	s.expires = time.Now().Add(time.Duration(life)*time.Second - time.Minute)
	return token, nil
}

func textraDropToken() {
	s := &textraTokenState
	s.Lock()
	defer s.Unlock()
	s.token = ""
}

// textraPost sends one API request (already counted by textraAcquire).
func textraPost(endpoint string, values url.Values) ([]byte, error) {
	res, err := textraHTTP.PostForm(endpoint, values)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return io.ReadAll(res.Body)
}

// textraError turns a non-zero resultset into an error, updating the slot
// count (502/504/505 are the request limits) or the token (510/511).
func textraError(prefix string, body []byte) error {
	code := gjson.GetBytes(body, "resultset.code").String()
	switch code {
	case "502", "504", "505":
		textraLimitHit()
		return textraLimitError{textraWindow}
	case "510", "511":
		textraDropToken()
	}
	return errors.New(prefix + " error " + code + ": " + gjson.GetBytes(body, "resultset.message").String())
}

func translateTexTra(to string, from string, text string) (LangOut, error) {
	FromOrig := from
	ToOrig := to
	if err := validateLanguagePair(langListTexTra("sl"), langListTexTra("tl"), from, to); err != nil {
		return LangOut{}, err
	}
	key, secret, name, baseURL, err := textraConfig()
	if err != nil {
		return LangOut{}, err
	}
	token, err := textraToken(baseURL, key, secret)
	if err != nil {
		return LangOut{}, err
	}
	if err := textraAcquire(0); err != nil {
		return LangOut{}, err
	}

	// The language pair goes in the URL: generalNT_<from>_<to>
	out, err := textraPost(baseURL+"/api/mt/generalNT_"+from+"_"+to+"/", url.Values{
		"access_token": {token}, "key": {key}, "name": {name}, "type": {"json"}, "text": {text},
	})
	if err != nil {
		return LangOut{}, err
	}
	if !gjson.ValidBytes(out) {
		return LangOut{}, errors.New("textra api error")
	}
	if gjson.GetBytes(out, "resultset.code").Int() != 0 {
		return LangOut{}, textraError("textra", out)
	}

	var langout LangOut
	langout.Engine = "textra"
	langout.SourceLang = FromOrig
	langout.TargetLang = ToOrig
	langout.OutputText = gjson.GetBytes(out, "resultset.result.text").String()

	// Dictionary: optional. Only when the pair has a system glossary and the
	// lookup still leaves a slot free for the TTS of this translation.
	if pid := textraDictPID(from, to); pid != "" && textraAcquire(1) == nil {
		dout, derr := textraPost(baseURL+"/api/lookup/", url.Values{
			"access_token": {token}, "key": {key}, "name": {name}, "type": {"json"},
			"pid": {pid}, "lang_s": {from}, "text": {text},
		})
		if derr == nil && gjson.ValidBytes(dout) {
			if gjson.GetBytes(dout, "resultset.code").Int() != 0 {
				_ = textraError("textra lookup", dout) // keeps slots and token in sync
			} else {
				seen := map[string]bool{}
				for _, lk := range gjson.GetBytes(dout, "resultset.result.lookup").Array() {
					for _, term := range lk.Get("term").Array() {
						t := term.Get("target").String()
						if t == "" || seen[t] {
							continue
						}
						seen[t] = true
						langout.WordChoices = append(langout.WordChoices, WordChoices{Word: t})
					}
				}
			}
		}
	}
	return langout, nil
}

// textraTTSLocales: the only 4 TexTra TTS voices.
var textraTTSLocales = map[string]string{
	"ja":    "ja-JP",
	"en":    "en-US",
	"zh":    "zh-CN",
	"zh-CN": "zh-CN",
	"ko":    "ko-KR",
}

func ttsTexTra(lang string, text string) ([]byte, error) {
	if _, ok := textraTTSLocales[lang]; !ok {
		return []byte(""), errors.New("textra tts: unsupported language")
	}
	if err := textraAcquire(0); err != nil {
		return []byte(""), err
	}
	return ttsTexTraRaw(lang, text)
}

// ttsTexTraRaw synthesizes text; the caller has already reserved the slot.
func ttsTexTraRaw(lang string, text string) ([]byte, error) {
	locale, ok := textraTTSLocales[lang]
	if !ok {
		return []byte(""), errors.New("textra tts: unsupported language")
	}
	key, secret, name, baseURL, err := textraConfig()
	if err != nil {
		return []byte(""), err
	}
	token, err := textraToken(baseURL, key, secret)
	if err != nil {
		return []byte(""), err
	}
	// Raw WAV (RIFF) on success; errors come back as JSON with type=json.
	data, err := textraPost(baseURL+"/api/tts/", url.Values{
		"access_token": {token}, "key": {key}, "name": {name}, "type": {"json"},
		"locale": {locale}, "text": {text},
	})
	if err != nil {
		return []byte(""), err
	}
	if len(data) < 4 || string(data[0:4]) != "RIFF" {
		if code := gjson.GetBytes(data, "resultset.code").String(); code == "" || code == "0" {
			return []byte(""), errors.New("textra tts: no audio returned")
		}
		return []byte(""), textraError("textra tts", data)
	}
	return data, nil
}
