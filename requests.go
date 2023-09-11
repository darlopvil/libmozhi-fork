package libmozhi

import (
	"bytes"
	"io"
	"net/http"
)

func postRequest(url string, data []byte, contenttype string) string {
	bodyReader := bytes.NewReader(data)
	r, err := http.NewRequest(http.MethodPost, url, bodyReader)
	if err != nil {
		panic(err)
	}

	UserAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"
	r.Header.Set("Content-Type", contenttype)
	r.Header.Set("User-Agent", UserAgent)

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	return string(body)
}

func getRequest(url string) string {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	UserAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"
	// r.Header.Set("Content-Type", "application/json")
	r.Header.Set("User-Agent", UserAgent)

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	return string(body)
}
