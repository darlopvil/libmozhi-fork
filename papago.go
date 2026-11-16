package libmozhi

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

// ============================================================================
// PAPAGO ENGINE — isla autónoma. No comparte código con otros engines.
// Naver's translator, through the endpoint its own website uses (no key;
// unofficial, so it can change without notice: /api/health shows it).
// ============================================================================

var papagoHTTP = &http.Client{Timeout: 15 * time.Second}

func translatePapago(to string, from string, text string) (LangOut, error) {
	if err := validateLanguagePair(langListPapago("sl"), langListPapago("tl"), from, to); err != nil {
		return LangOut{}, err
	}
	form := url.Values{
		"source": {from}, "target": {to}, "text": {text}, "honorific": {"false"},
	}
	req, err := http.NewRequest("POST", "https://papago.naver.com/api/text/translation", strings.NewReader(form.Encode()))
	if err != nil {
		return LangOut{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:140.0) Gecko/20100101 Firefox/140.0")
	res, err := papagoHTTP.Do(req)
	if err != nil {
		return LangOut{}, err
	}
	defer res.Body.Close()
	out, err := io.ReadAll(res.Body)
	if err != nil {
		return LangOut{}, err
	}
	translated := gjson.GetBytes(out, "translatedText").String()
	if res.StatusCode != 200 || translated == "" {
		return LangOut{}, errors.New("papago error: " + res.Status + " " + gjson.GetBytes(out, "errorCode").String())
	}
	var langout LangOut
	langout.Engine = "papago"
	langout.SourceLang = from
	langout.TargetLang = to
	langout.OutputText = translated
	if from == "auto" {
		langout.AutoDetect = gjson.GetBytes(out, "srcLangType").String()
	}
	return langout, nil
}
