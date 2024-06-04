package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	DefaultTimeout = time.Second * 10
)

var (
	DefaultHTTPClient = &http.Client{}
)

type Request struct {
	Method           string
	URL              string
	Params           *url.Values
	Header           *http.Header
	Timeout          time.Duration
	ProxyURL         string
	DisallowRedirect bool

	httpRequest *http.Request
	client      *http.Client
}

func (req *Request) newHTTPRequest() (*http.Request, error) {
	if req.Method == "" {
		req.Method = http.MethodGet
	}
	encodedParams := ""
	if req.Params != nil {
		encodedParams = req.Params.Encode()
	}
	var err error
	var httpRequest *http.Request
	switch req.Method {
	case http.MethodGet:
		httpRequest, err = http.NewRequest(req.Method, req.URL, nil)
		if encodedParams != "" {
			httpRequest.URL.RawQuery = encodedParams
		}
		if req.Header != nil {
			httpRequest.Header = *req.Header
		}
	case http.MethodPost:
		fallthrough
	case http.MethodDelete:
		fallthrough
	case http.MethodPut:
		httpRequest, err = http.NewRequest(req.Method, req.URL, bytes.NewBufferString(encodedParams))
		if req.Header != nil {
			httpRequest.Header = *req.Header
		}
		httpRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		httpRequest.Header.Add("Content-Length", strconv.Itoa(len(encodedParams)))
	default:
		err = errors.New("not yet implemented")
	}

	return httpRequest, err
}

func (req *Request) Build() *Request {
	var err error
	req.httpRequest, err = req.newHTTPRequest()
	if err != nil {
		log.Panic(err)
	}
	if req.Timeout == 0 {
		req.Timeout = DefaultTimeout
	}
	transport := DefaultHTTPClient.Transport
	if req.ProxyURL != "" {
		proxyURL, err := url.Parse(req.ProxyURL)
		if err != nil {
			log.Panic(err)
		}
		transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	}

	req.client = &http.Client{
		Timeout:   req.Timeout,
		Transport: transport,
	}

	if req.DisallowRedirect {
		req.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return req
}

func (req *Request) GetHTTPRequest() *http.Request {
	if req.client == nil {
		req.Build()
	}
	return req.httpRequest
}

func (req *Request) MakeRequest() (*http.Response, error) {
	if req.client == nil || req.httpRequest == nil {
		req.Build()
	}
	return req.client.Do(req.httpRequest)
}

func (req *Request) parseResponse(body io.ReadCloser, target interface{}) error {
	bodyData, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bodyData, target); err != nil {
		if _, ok := err.(*json.SyntaxError); ok {
			log.WithError(err).WithField("Response body", string(bodyData)).Errorf("Response Body cannot be parsed. Body - %s", string(bodyData))
		}
		return err
	}
	return nil
}

func (req *Request) GetResponse(target interface{}) (int, error) {
	httpResponse, err := req.MakeRequest()
	if err != nil {
		// TODO: Change http.StatusInternalServerError to more proper value - zune
		return http.StatusInternalServerError, err
	}

	// https://stackoverflow.com/questions/33238518/what-could-happen-if-i-dont-close-response-body-in-golang
	//noinspection ALL
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode >= http.StatusInternalServerError {
		return httpResponse.StatusCode, errors.New("internal server error")
	}

	if err := req.parseResponse(httpResponse.Body, target); err != nil {
		return httpResponse.StatusCode, err
	}
	return httpResponse.StatusCode, nil
}

func GetHost(urlString string) string {
	urlObj, err := url.Parse(urlString)
	if err != nil {
		panic(err)
	}
	return urlObj.Host
}
