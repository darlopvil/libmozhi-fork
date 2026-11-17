package libmozhi

import (
	"net/url"
	"strings"

	"github.com/tidwall/gjson"
)

// yandexTranslitURL is Yandex's transliteration service. It answers a JSON
// string with the romanized text.
var yandexTranslitURL = "https://translate.yandex.net/translit/translit"

// yandexTranslit romanizes text written in lang. It answers "" when there is
// nothing useful to show: a failed request, an answer that isn't the JSON
// string it should be (long texts get a "Request-URI Too Large" page), or the
// same text back (languages already written in Latin script).
func yandexTranslit(lang, text string) string {
	if lang == "" || lang == "auto" || strings.TrimSpace(text) == "" {
		return ""
	}
	body, err := getRequest(yandexTranslitURL + "?lang=" + url.QueryEscape(lang) + "&text=" + url.QueryEscape(text))
	if err != nil || !gjson.Valid(body) {
		return ""
	}
	res := gjson.Parse(body)
	if res.Type != gjson.String {
		return ""
	}
	out := strings.ReplaceAll(res.String(), "\r", "")
	if strings.TrimSpace(out) == strings.TrimSpace(text) {
		return ""
	}
	return out
}
