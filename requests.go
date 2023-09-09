package utils

import (
	"bytes"
	"io"
	"net/http"
	"os"

	"github.com/tidwall/gjson"
)

func postRequest(url string, data []byte) gjson.Result {
	bodyReader := bytes.NewReader(data)
	r, err := http.NewRequest(http.MethodPost, url, bodyReader)
	if err != nil {
		panic(err)
	}

	UserAgent, ok := os.LookupEnv("MOZHI_USER_AGENT")
	if !ok {
		UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"
	}
	// r.Header.Set("Accept", "application/vnd.github.v3+json")
	r.Header.Set("Content-Type", "application/json")
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
	jsonified := gjson.Parse(string(body))

	return jsonified
}

func getRequest(url string) gjson.Result {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	UserAgent, ok := os.LookupEnv("MOZHI_USER_AGENT")
	if !ok {
		UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"
	}
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
	jsonified := gjson.Parse(string(body))

	return jsonified
}
