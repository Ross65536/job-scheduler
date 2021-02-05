package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	jsonMime = "application/json"
)

type HTTPClient struct {
	username string
	token    string
	apiUrl   url.URL
}

// it's thread-safe, recommended to reuse the client
var httpClient = http.Client{}

func NewHTTPClient(apiUrl string) (*HTTPClient, error) {
	uri, err := url.ParseRequestURI(apiUrl)
	if err != nil {
		return nil, err
	}

	if !uri.IsAbs() {
		return nil, errors.New("uri must be absolute")
	}

	// use https when supported
	if uri.Scheme != "http" {
		return nil, errors.New("invalid url scheme")
	}

	if uri.User == nil {
		return nil, errors.New("base URI must have an HTTP basic username and password encoded")
	}

	password, ok := uri.User.Password()
	if !ok {
		return nil, errors.New("base URI must have an HTTP basic password encoded")
	}

	return &HTTPClient{
		username: uri.User.Username(),
		token:    password,
		apiUrl:   *uri,
	}, nil
}

func (c *HTTPClient) Get(pathSegments ...string) (returnBody []byte, err error) {
	return c.makeJSONRequest(http.MethodGet, nil, pathSegments...)
}

func (c *HTTPClient) Post(requestBody []byte, pathSegments ...string) (returnBody []byte, err error) {
	return c.makeJSONRequest(http.MethodPost, requestBody, pathSegments...)
}

func (c *HTTPClient) Delete(pathSegments ...string) (returnBody []byte, err error) {
	return c.makeJSONRequest(http.MethodDelete, nil, pathSegments...)
}

func (c *HTTPClient) makeJSONRequest(requestMethod string, requestBody []byte, pathSegments ...string) (returnBody []byte, err error) {
	request, err := c.buildRequest(requestMethod, pathSegments, requestBody)
	if err != nil {
		return nil, err
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode/100 != 2 {
		return nil, c.buildProtocolErrorMessage(response)
	}

	// TODO add checks for Content-Type header, etc

	return ReadCloseableBuffer(response.Body)
}

func (c *HTTPClient) joinPathFragments(pathSegments []string) (string, error) {
	encodedPaths := make([]string, len(pathSegments))
	for i, pathSegment := range pathSegments {
		if IsWhitespaceString(pathSegment) {
			return "", errors.New("invalid uri path segment: " + pathSegment)
		}

		encodedPaths[i] = url.PathEscape(pathSegment)
	}

	return c.apiUrl.String() + "/" + strings.Join(encodedPaths, "/"), nil
}

func (c *HTTPClient) buildRequest(requestMethod string, pathSegments []string, requestBody []byte) (*http.Request, error) {
	if requestMethod != http.MethodGet && requestMethod != http.MethodPost && requestMethod != http.MethodPut && requestMethod != http.MethodDelete {
		return nil, errors.New("unsupported HTTP method")
	}

	requestURI, err := c.joinPathFragments(pathSegments)
	if err != nil {
		return nil, err
	}

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

type ErrorType struct {
	Status  int
	Message string
}

func (api *HTTPClient) buildProtocolErrorMessage(response *http.Response) error {
	code := response.StatusCode

	body, err := ReadCloseableBuffer(response.Body)
	if err != nil {
		return fmt.Errorf("An error occurred (HTTP %d), failed reading HTTP response: %s", code, err)
	}

	// best case JSON parsing, return raw HTTP body otherwise
	parsed := ErrorType{}
	err = json.Unmarshal(body, &parsed)

	if err == nil && parsed.Message != "" {
		return fmt.Errorf("An error occurred (HTTP %d): %s", code, parsed.Message)
	} else {
		return fmt.Errorf("An error occurred (HTTP %d): %s", code, string(body))
	}
}
