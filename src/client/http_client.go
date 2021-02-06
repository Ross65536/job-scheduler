package client

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"
	"path"
)

const (
	jsonMime = "application/json"
)

type HTTPClient struct {
	username   string
	token      string
	apiUrl     url.URL
	httpClient http.Client
}

func parseUrl(apiUrl, expectedScheme string) (*url.URL, error) {
	uri, err := url.ParseRequestURI(apiUrl)
	if err != nil {
		return nil, err
	}

	if !uri.IsAbs() {
		return nil, errors.New("uri must be absolute")
	}

	// use https when supported
	if uri.Scheme != expectedScheme {
		return nil, errors.New("invalid url scheme")
	}

	if uri.User == nil {
		return nil, errors.New("base URI must have an HTTP basic username and password encoded")
	}

	_, ok := uri.User.Password()
	if !ok {
		return nil, errors.New("base URI must have an HTTP basic password encoded")
	}

	return uri, nil
}

func NewHTTPClient(apiUrl string) (*HTTPClient, error) {
	uri, err := parseUrl(apiUrl, "http")
	if err != nil {
		return nil, err
	}

	// ok checked before hand
	password, _ := uri.User.Password()
	return &HTTPClient{
		username:   uri.User.Username(),
		token:      password,
		apiUrl:     *uri,
		httpClient: http.Client{},
	}, nil
}

func (c *HTTPClient) Get(pathSegments ...string) (statusCode int, returnBody []byte, err error) {
	return c.makeJSONRequest(http.MethodGet, nil, pathSegments...)
}

func (c *HTTPClient) Post(requestBody []byte, pathSegments ...string) (statusCode int, returnBody []byte, err error) {
	return c.makeJSONRequest(http.MethodPost, requestBody, pathSegments...)
}

func (c *HTTPClient) Delete(pathSegments ...string) (statusCode int, returnBody []byte, err error) {
	return c.makeJSONRequest(http.MethodDelete, nil, pathSegments...)
}

func (c *HTTPClient) makeJSONRequest(requestMethod string, requestBody []byte, pathSegments ...string) (statusCode int, returnBody []byte, err error) {
	request, err := c.buildRequest(requestMethod, pathSegments, requestBody)
	if err != nil {
		return -1, nil, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return -1, nil, err
	}

	// TODO add checks for Content-Type header, etc
	resp, err := ReadCloseableBuffer(response.Body)
	return response.StatusCode, resp, err
}

func (c *HTTPClient) joinPathFragments(pathSegments []string) string {
	encodedPaths := path.Join(pathSegments...)

	return c.apiUrl.String() + "/" + encodedPaths
}

func (c *HTTPClient) buildRequest(requestMethod string, pathSegments []string, requestBody []byte) (*http.Request, error) {
	requestURI := c.joinPathFragments(pathSegments)

	request, err := http.NewRequest(requestMethod, requestURI, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	//request headers
	request.Header.Set("Host", c.apiUrl.Host)
	request.Header.Set("Accept", jsonMime)
	request.SetBasicAuth(c.username, c.token)
	if requestBody != nil && len(requestBody) != 0 {
		request.Header.Set("Content-Type", jsonMime+"; charset=utf-8")
	}

	return request, nil
}
