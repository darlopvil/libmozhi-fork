package libmozhi

import "os"
import "github.com/tidwall/gjson"

func AutoDetectLibreTranslate(query string) (string, error) {
	json := []byte(`{"q":"` + query + `"}`)
	libreTranslateOut, err := postRequest(os.Getenv("MOZHI_LIBRETRANSLATE_URL")+"/detect", json, "application/json")
	if err != nil {
		return "", err
	}
	gjsonArr := gjson.Get(libreTranslateOut, "0.language").Array()
	answer := gjsonArr[0].String()
	return answer, nil
}
