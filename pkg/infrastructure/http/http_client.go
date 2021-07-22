package httptrp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flagsense-go-sdk/pkg/util"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

func NewFlagSenseHttpClient() *http.Client {
	return &http.Client{
		Timeout: 5000 * time.Millisecond,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 120 * time.Second,
			}).DialContext,
			MaxIdleConnsPerHost:   100,
			MaxIdleConns:          1000,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

func MakeHttpRequest(ctx context.Context, method string, url string, client *http.Client, requestBody *bytes.Buffer, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest(method, url, requestBody)

	if err != nil {
		return nil, fmt.Errorf("error while making http request. err=%+v, url=%s", err, url)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	response, err := client.Do(req.WithContext(ctx))
	if err != nil || response == nil {
		return nil, fmt.Errorf("error in sending request to API endpoint. err=%+v, url=%s", err, url)
	}

	// Close the connection to reuse it
	defer response.Body.Close()
	if response.StatusCode != 200 {
		var target interface{}
		body := json.NewDecoder(response.Body).Decode(target)
		return nil, errors.New(fmt.Sprintf("non 200 response received. statusCode=%d, url=%s, err=%v", response.StatusCode, url, body))
	}

	// Let's check if the work actually is done
	// We have seen inconsistencies even when we get 200 OK response
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		bodyStr := string(body)
		truncatedBody := bodyStr[0:util.Min(100, len(bodyStr))]
		return nil, fmt.Errorf("error in reading all body, body=%+v, err=%+v, url=%s", truncatedBody, err, url)
	}

	return body, nil
}

func makeRequest(ctx context.Context, client *http.Client, req *http.Request) ([]byte, error) {
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	return body, nil
}
