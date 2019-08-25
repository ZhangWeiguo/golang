package easyserver

import (
	"io"
	"net/http"
	"time"
)

type EasyHttpClient struct {
	client  *http.Client
	Timeout time.Duration
	cookie  http.Cookie
}

func (ehc *EasyHttpClient) Init() {
	ehc.client = &http.Client{}
	if ehc.Timeout <= 0 {
		ehc.Timeout = DEFAULT_TIMEOUT
	}
	ehc.client.Timeout = ehc.Timeout
}

func (ehc *EasyHttpClient) Get(url string) (result string, err error) {
	var p []byte
	request, err := http.NewRequest("GET", url, nil)
	if err == nil {
		response, err := ehc.client.Do(request)
		if err == nil {
			_, err = response.Body.Read(p)
		}
	}
	return string(p), err
}

func (ehc *EasyHttpClient) Post(url string, contentType string, body io.Reader) (s string, err error) {
	var p []byte
	request, err := http.NewRequest("POST", url, body)
	if err == nil {
		request.Header.Set("Content-Type", contentType)
		response, err := ehc.client.Do(request)
		if err == nil {
			_, err = response.Body.Read(p)
		}
	}
	return string(p), err
}
