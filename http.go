package atomicgo

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

func HttpGet(url string) (result string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	return string(byteArray), nil
}

type HttpReqestType string

const (
	Post HttpReqestType = "POST"
	Get  HttpReqestType = "GET"
)

// 複数headerを送る際は map["A"] = "a;b;c"
func HttpReqest(method HttpReqestType, url string, body string, headers map[string]string) (resp *http.Response, err error) {
	// リクエストの準備
	req, _ := http.NewRequest(string(method), url, bytes.NewBuffer([]byte(body)))
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	// 送信クライアント準備
	client := new(http.Client)
	// Request送信
	resp, err = client.Do(req)
	return
}
