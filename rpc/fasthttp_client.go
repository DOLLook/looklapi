package rpc

import (
	"crypto/tls"
	"encoding/json"
	"github.com/valyala/fasthttp"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var _client *fasthttp.Client

func init() {
	_client = &fasthttp.Client{
		Name:                     "looklapi",
		NoDefaultUserAgentHeader: false,
		TLSConfig:                &tls.Config{InsecureSkipVerify: true},
		MaxConnsPerHost:          2000,
		MaxIdleConnDuration:      10 * time.Second,
		MaxConnDuration:          10 * time.Second,
		ReadTimeout:              10 * time.Second,
		WriteTimeout:             10 * time.Second,
		MaxConnWaitTimeout:       10 * time.Second,
	}
}

func DoRequest(reqMethod string, url string, header *http.Header, body interface{}, urlParams []string) ([]byte, error) {

	if len(urlParams) > 0 {
		urlparam := strings.Join(urlParams, "&")
		url = url + "?" + urlparam
		//loggers.GetLogger().Debug(url)
	}

	var bodyStream *[]byte
	if body != nil {
		if stream, err := json.Marshal(body); err != nil {
			return nil, err
		} else {
			bodyStream = &stream
		}
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)
	if bodyStream != nil && len(*bodyStream) > 0 {
		req.SetBody(*bodyStream)
	}

	req.Header.SetMethod(reqMethod)
	if reqMethod == http.MethodPost {
		req.Header.SetContentType("application/json")
	}

	if header != nil {
		//loggers.GetLogger().Debug(utils.StructToJson(header))
		for k, v := range *header {
			req.Header.Set(k, v[0])
		}
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err := _client.Do(req, resp); err != nil {
		return nil, err
	}

	respBytes := resp.Body()

	copyRespBytes := make([]byte, len(respBytes))
	if len(respBytes) > 0 {
		copy(copyRespBytes, respBytes)
	}

	return copyRespBytes, nil
}

func DoXmlRequest(url string, header *http.Header, xml string) ([]byte, error) {

	var bodyStream []byte
	if len(xml) > 0 {
		stream := []byte(xml)
		bodyStream = stream
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)
	if bodyStream != nil && len(bodyStream) > 0 {
		req.SetBody(bodyStream)
	}

	req.Header.SetMethod(http.MethodPost)
	req.Header.SetContentType("application/xml")

	if header != nil {
		//loggers.GetLogger().Debug(utils.StructToJson(header))
		for k, v := range *header {
			req.Header.Set(k, v[0])
		}
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err := _client.Do(req, resp); err != nil {
		return nil, err
	}

	respBytes := resp.Body()

	copyRespBytes := make([]byte, len(respBytes))
	if len(respBytes) > 0 {
		copy(copyRespBytes, respBytes)
	}

	return copyRespBytes, nil
}

func DoFormRequest(url string, header *http.Header, data url.Values) ([]byte, error) {

	var bodyStream *[]byte
	if len(data) > 0 {
		formBytes := []byte(data.Encode())
		bodyStream = &formBytes
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)
	if bodyStream != nil && len(*bodyStream) > 0 {
		req.SetBody(*bodyStream)
	}

	req.Header.SetMethod(http.MethodPost)
	req.Header.SetContentType("application/x-www-form-urlencoded")

	if header != nil {
		//loggers.GetLogger().Debug(utils.StructToJson(header))
		for k, v := range *header {
			req.Header.Set(k, v[0])
		}
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err := _client.Do(req, resp); err != nil {
		return nil, err
	}

	respBytes := resp.Body()

	copyRespBytes := make([]byte, len(respBytes))
	if len(respBytes) > 0 {
		copy(copyRespBytes, respBytes)
	}

	return copyRespBytes, nil
}
