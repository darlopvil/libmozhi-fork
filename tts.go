package libmozhi

import (
	"context"
	"encoding/base64"

	"github.com/carlmjohnson/requests"
	"github.com/google/go-querystring/query"
)

type reversoTTS struct {
	Id    string
	Voice string
}

func ttsGoogle(lang string, text string) ([]byte, error) {
	type Options struct {
		IE     string `url:"ie"`
		Lang   string `url:"tl"`
		Text   string `url:"q"`
		Client string `url:"client"`
	}
	opt := Options{"UTF-8", lang, text, "tw-ob"}
	v, _ := query.Values(opt)
	var file string
	url := "https://translate.google.com/translate_tts?" + v.Encode()
	UserAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"
	err := requests.
		URL(url).
		ToString(&file).
		UserAgent(UserAgent).
		Fetch(context.Background())
	if err != nil {
		return []byte(""), err
	}
	return []byte(file), nil
}

func ttsReverso(lang string, text string) ([]byte, error) {
	TTSData := []reversoTTS{
		// http://voice.reverso.net/RestPronunciation.svc/v1/output=json/GetAvailableVoices with randomized deduplication
		{
			Id:    "ar",
			Voice: "Mehdi22k",
		},
		{
			Id:    "zh",
			Voice: "Lulu22k",
		},
		{
			Id:    "cz",
			Voice: "Eliska22k",
		},
		{
			Id:    "dk",
			Voice: "Mette22k",
		},
		{
			Id:    "nl",
			Voice: "Daan22k",
		},
		{
			Id:    "en",
			Voice: "Will22k",
		},
		{
			Id:    "fr",
			Voice: "Margaux22k",
		},
		{
			Id:    "de",
			Voice: "Andreas22k",
		},
		{
			Id:    "gr",
			Voice: "Dimitris22k",
		},
		{
			Id:    "heb",
			Voice: "he-IL-Asaf",
		},
		{
			Id:    "it",
			Voice: "Chiara22k",
		},
		{
			Id:    "jp",
			Voice: "Sakura22k",
		},
		{
			Id:    "kr",
			Voice: "Minji22k",
		},
		{
			Id:    "pl",
			Voice: "Monika22k",
		},
		{
			Id:    "pt",
			Voice: "Celia22k",
		},
		{
			Id:    "ro",
			Voice: "ro-RO-Andrei",
		},
		{
			Id:    "ru",
			Voice: "Alyona22k",
		},
		{
			Id:    "es",
			Voice: "Antonio22k",
		},
		{
			Id:    "se",
			Voice: "Erik22k",
		},
		{
			Id:    "tr",
			Voice: "Ipek22k",
		},
	}
	// https://voice.reverso.net/RestPronunciation.svc/v1/output=json/GetVoiceStream/voiceName=Lulu22k?voiceSpeed=80&inputText=6K%20V6aqM Base64 input text
	text2 := base64.StdEncoding.EncodeToString([]byte(text))
	var voice string
	for _, s := range TTSData {
		if s.Id == lang {
			voice = s.Voice
			break
		}
	}

	type Options struct {
		VoiceSpeed int    `url:"voiceSpeed"`
		Text       string `url:"inputText"`
	}
	opt := Options{80, text2}
	v, _ := query.Values(opt)

	var file string
	url := "https://voice.reverso.net/RestPronunciation.svc/v1/output=json/GetVoiceStream/voiceName=" + voice + "?" + v.Encode()

	UserAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"
	err := requests.
		URL(url).
		ToString(&file).
		UserAgent(UserAgent).
		Fetch(context.Background())
	if err != nil {
		return []byte(""), err
	}
	return []byte(file), nil
}
