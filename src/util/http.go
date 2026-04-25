package util

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"net/url"
	"time"
)

func MakeHTTPRequest(method string, url string, headers map[string]string, values *url.Values, data []byte, timeoutSeconds uint32) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	if headers != nil {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}

	if values != nil {
		req.URL.RawQuery = values.Encode()
	}

	tr := &http.Transport{
		//TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := http.Client{Transport: tr}
	client.Timeout = time.Second * time.Duration(timeoutSeconds)
	return client.Do(req)
}

func AddBasicAuth(headers map[string]string, username string, password string) map[string]string {
	headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
	return headers
}
func AddBearerToken(headers map[string]string, bearerToken string) map[string]string {
	headers["Authorization"] = "Bearer " + bearerToken
	return headers
}
