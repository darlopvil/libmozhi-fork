package libmozhi

import "os"
import "github.com/tidwall/gjson"

func AutoDetectWatson(query string) (string, error) {
	json := []byte(`{"text":"` + query + `"}`)
	watsonOut := postRequest("https://www.ibm.com/demos/live/watson-language-translator/api/translate/detect", json, "application/json")
	gjsonArr := gjson.Get(watsonOut, "payload.languages.0.language.language").Array()
	answer := gjsonArr[0].String()
	return answer, nil
}

func AutoDetectLibreTranslate(query string) (string, error) {
	json := []byte(`{"q":"` + query + `"}`)
	libreTranslateOut := postRequest(os.Getenv("MOZHI_LIBRETRANSLATE_URL")+"/detect", json, "application/json")
	gjsonArr := gjson.Get(libreTranslateOut, "0.language").Array()
	answer := gjsonArr[0].String()
	return answer, nil
}
