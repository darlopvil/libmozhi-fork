package libmozhi

import (
	"bytes"
	"io"
	"net/http"
)

func postRequest(url string, data []byte, contenttype string) (string, error) {
	bodyReader := bytes.NewReader(data)
	r, err := http.NewRequest(http.MethodPost, url, bodyReader)
	if err != nil {
		return "", err
	}

	UserAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"
	r.Header.Set("Content-Type", contenttype)
	r.Header.Set("User-Agent", UserAgent)

	client := &http.Client{}
	res, err0 := client.Do(r)
	if err0 != nil {
		return "", err0
	}

	defer res.Body.Close()

	body, err1 := io.ReadAll(res.Body)
	if err1 != nil {
		return "", err1
	}

	return string(body), nil
}

func getRequest(url string) (string, error) {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	UserAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"
	// r.Header.Set("Content-Type", "application/json")
	r.Header.Set("User-Agent", UserAgent)

	client := &http.Client{}
	res, err0 := client.Do(r)
	if err0 != nil {
		return "", err0
	}

	defer res.Body.Close()

	body, err1 := io.ReadAll(res.Body)
	if err1 != nil {
		return "", err1
	}
	return string(body), nil
}
