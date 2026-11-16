package libmozhi

import (
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/gjson"
)

// ============================================================================
// MICROSOFT TTS — Azure Speech neural voices, part of the Microsoft engine's
// island. It needs its own resource (Speech, regional) besides the
// Translator one. The free tier (F0) allows 500,000 characters a month and
// 20 requests a minute. One voice per language, picked from Microsoft's
// voice list (cached for a day) unless MOZHI_MICROSOFT_TTS_VOICES says
// otherwise ("ja=ja-JP-NanamiNeural,es=es-ES-AlvaroNeural").
// ============================================================================

const microsoftVoicesTTL = 24 * time.Hour

// microsoftDefaultVoices: the usual default voice of the main languages,
// used when the voice list has it.
var microsoftDefaultVoices = map[string]string{
	"en": "en-US-JennyNeural", "es": "es-ES-ElviraNeural", "ja": "ja-JP-NanamiNeural",
	"zh": "zh-CN-XiaoxiaoNeural", "ko": "ko-KR-SunHiNeural", "fr": "fr-FR-DeniseNeural",
	"de": "de-DE-KatjaNeural", "it": "it-IT-ElsaNeural", "pt": "pt-BR-FranciscaNeural",
	"ru": "ru-RU-SvetlanaNeural", "ar": "ar-SA-ZariyahNeural", "hi": "hi-IN-SwaraNeural",
}

var microsoftVoices struct {
	sync.Mutex
	list    []gjson.Result
	fetched time.Time
}

var microsoftTTSHTTP = &http.Client{Timeout: 15 * time.Second}

func microsoftSpeechConfig() (key, region string, err error) {
	key = os.Getenv("MOZHI_MICROSOFT_SPEECH_KEY")
	region = os.Getenv("MOZHI_MICROSOFT_SPEECH_REGION")
	if key == "" || region == "" {
		return "", "", errors.New("microsoft tts requires MOZHI_MICROSOFT_SPEECH_KEY and MOZHI_MICROSOFT_SPEECH_REGION")
	}
	return key, region, nil
}

// microsoftVoiceList returns Microsoft's voice list, fetched at most once a
// day (the list itself costs no characters).
func microsoftVoiceList(key, region string) ([]gjson.Result, error) {
	v := &microsoftVoices
	v.Lock()
	defer v.Unlock()
	if v.list != nil && time.Since(v.fetched) < microsoftVoicesTTL {
		return v.list, nil
	}
	req, err := http.NewRequest("GET", "https://"+region+".tts.speech.microsoft.com/cognitiveservices/voices/list", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Ocp-Apim-Subscription-Key", key)
	res, err := microsoftTTSHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	out, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 || !gjson.ValidBytes(out) {
		return nil, errors.New("microsoft tts: voice list: " + res.Status)
	}
	v.list = gjson.ParseBytes(out).Array()
	v.fetched = time.Now()
	return v.list, nil
}

// microsoftLocale maps Mozhi language codes to Azure locales where they
// differ in more than case.
func microsoftLocale(lang string) string {
	switch strings.ToLower(lang) {
	case "zh", "zh-hans", "zh-cn":
		return "zh-cn"
	case "zh-hant", "zh-tw":
		return "zh-tw"
	case "no", "nb":
		return "nb-no"
	case "tl", "fil":
		return "fil-ph"
	case "sr", "sr-latn", "sr-cyrl":
		return "sr-rs"
	}
	return strings.ToLower(lang)
}

// microsoftPickVoice returns the voice for lang: the configured one, the
// language's usual default, the first neural voice of "<lang>-<LANG>"
// (es-ES, fr-FR...), or the first neural voice of the language.
func microsoftPickVoice(voices []gjson.Result, lang string) (voice, locale string, err error) {
	for _, pair := range strings.Split(os.Getenv("MOZHI_MICROSOFT_TTS_VOICES"), ",") {
		if l, name, ok := strings.Cut(strings.TrimSpace(pair), "="); ok && strings.EqualFold(l, lang) {
			if i := strings.LastIndex(name, "-"); i > 0 {
				return name, name[:i], nil
			}
		}
	}
	want := microsoftLocale(lang)
	base := strings.SplitN(want, "-", 2)[0]
	neural := func(v gjson.Result) bool { return v.Get("VoiceType").String() == "Neural" }
	if def, ok := microsoftDefaultVoices[base]; ok && (!strings.Contains(want, "-") || strings.HasPrefix(strings.ToLower(def), want+"-")) {
		for _, v := range voices {
			if v.Get("ShortName").String() == def {
				return def, v.Get("Locale").String(), nil
			}
		}
	}
	match := func(locale string) (string, string, bool) {
		for _, v := range voices {
			if neural(v) && strings.EqualFold(v.Get("Locale").String(), locale) {
				return v.Get("ShortName").String(), v.Get("Locale").String(), true
			}
		}
		return "", "", false
	}
	if strings.Contains(want, "-") {
		if name, loc, ok := match(want); ok {
			return name, loc, nil
		}
	}
	if name, loc, ok := match(base + "-" + base); ok {
		return name, loc, nil
	}
	for _, v := range voices {
		loc := strings.ToLower(v.Get("Locale").String())
		if neural(v) && (loc == base || strings.HasPrefix(loc, base+"-")) {
			return v.Get("ShortName").String(), v.Get("Locale").String(), nil
		}
	}
	return "", "", errors.New("microsoft tts: unsupported language")
}

func ttsMicrosoft(lang string, text string) ([]byte, error) {
	key, region, err := microsoftSpeechConfig()
	if err != nil {
		return nil, err
	}
	voices, err := microsoftVoiceList(key, region)
	if err != nil {
		return nil, err
	}
	voice, locale, err := microsoftPickVoice(voices, lang)
	if err != nil {
		return nil, err
	}
	var escaped strings.Builder
	if err := xml.EscapeText(&escaped, []byte(text)); err != nil {
		return nil, err
	}
	ssml := `<speak version="1.0" xml:lang="` + locale + `"><voice name="` + voice + `">` + escaped.String() + `</voice></speak>`
	req, err := http.NewRequest("POST", "https://"+region+".tts.speech.microsoft.com/cognitiveservices/v1", strings.NewReader(ssml))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Ocp-Apim-Subscription-Key", key)
	req.Header.Set("Content-Type", "application/ssml+xml")
	req.Header.Set("X-Microsoft-OutputFormat", "audio-24khz-48kbitrate-mono-mp3")
	req.Header.Set("User-Agent", "mozhi")
	res, err := microsoftTTSHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 || len(data) == 0 {
		return nil, errors.New("microsoft tts error: " + res.Status)
	}
	return data, nil
}
