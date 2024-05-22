package libmozhi

import (
	"errors"
	"github.com/tidwall/gjson"
	"net/url"
	"strings"
)

type ImgOut struct {
	SourceB64            string
	TranslatedImgB64     string
	SourceLang           string
	TargetLang           string
	SourceTextParsed     string
	TranslatedTextParsed string
}

func ImageGoogle(to string, from string, imgB64 string) (ImgOut, error) {
	ToOrig := to
	FromOrig := from
	// For some reason google uses no for norwegian instead of nb like the rest of the translators. This is for the All function primarily
	if to == "nb" {
		to = "no"
	} else if from == "nb" {
		to = "no"
	}
	var ToValid bool
	var FromValid bool
	for _, v := range langListGoogle("sl") {
		if v.Id == to {
			ToValid = true
		}
		if v.Id == from {
			FromValid = true
		}
		if FromValid == true && ToValid == true {
			break
		}
	}
	if ToValid != true {
		return ImgOut{}, errors.New("Target Language Code invalid")
	}
	if FromValid != true {
		return ImgOut{}, errors.New("Source language code invalid")
	}
	data := url.Values{}
	data.Set("f.req", `[[["WqWDPb","[[\"`+string(imgB64)+`\",\"image/png\"],\"`+from+`\",\"`+to+`\"]",null,"generic"]]]`)
	googleOut, err := postRequest("https://translate.google.com/_/TranslateWebserverUi/data/batchexecute", []byte(data.Encode()), "application/x-www-form-urlencoded")
	if err != nil {
		return ImgOut{}, err
	}
	googleOut = strings.TrimPrefix(googleOut, ")]}'")
	if !gjson.Valid(googleOut) {
		return ImgOut{}, errors.New("invalid json")
	}
	initial := gjson.Get(googleOut, "0.2").String()
	var imgout ImgOut
	// Thanks jsonselector.com
	imgout.SourceB64 = imgB64
	imgout.TranslatedImgB64 = gjson.Get(initial, "0.0").String()
	imgout.SourceLang = FromOrig
	imgout.TargetLang = ToOrig
	imgout.SourceTextParsed = gjson.Get(initial, "1").String()
	imgout.TranslatedTextParsed = gjson.Get(initial, "2").String()
	return imgout, nil
}
