package internal

import (
	"bytes"
	"errors"
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

func NewHTTPClient(username, token, apiUrl string) (*HTTPClient, error) {
	uri, err := url.Parse(apiUrl)
	if err != nil {
		return nil, err
	}

	if !uri.IsAbs() {
		return nil, errors.New("uri must be absolute")
	}

	if uri.Scheme != "http" && uri.Scheme != "https" {
		return nil, errors.New("Invalid url scheme")
	}

	if IsWhitespaceString(username) || IsWhitespaceString(token) {
		return nil, errors.New("Invalid HTTP basic credentials")
	}

	return &HTTPClient{
		username: username,
		token:    token,
		apiUrl:   *uri,
	}, nil
}

func (c *HTTPClient) MakeJSONRequest(requestMethod string, requestBody []byte, pathSegments ...string) (returnBody []byte, err error) {
	request, err := c.buildRequest(requestMethod, pathSegments, requestBody)
	if err != nil {
		return nil, err
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode/100 != 2 {
		return nil, errors.New("Something went wrong")
		// return nil, c.buildProtocolErrorMessage(response)
	}

	if !JsonContentType(&response.Header) {
		return nil, errors.New("returned content-type isn't json")
	}

	return ReadCloseableBuffer(response.Body)
}

func (c *HTTPClient) joinPathFragments(pathSegments []string) (string, error) {
	encodedPaths := make([]string, len(pathSegments))
	for i, pathSegment := range pathSegments {
		if IsWhitespaceString(pathSegment) {
			return "", errors.New("Invalid uri path segment: " + pathSegment)
		}

		encodedPaths[i] = url.PathEscape(pathSegment)
	}

	return c.apiUrl.String() + "/" + strings.Join(encodedPaths, "/"), nil
}

func (c *HTTPClient) buildRequest(requestMethod string, pathSegments []string, requestBody []byte) (*http.Request, error) {
	if requestMethod != http.MethodGet && requestMethod != http.MethodPost && requestMethod != http.MethodPut && requestMethod != http.MethodDelete {
		return nil, errors.New("Unsupported HTTP method")
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

// func (api *HTTPClient) buildProtocolErrorMessage(response *http.Response) error {
// 	var httpError error = HttpStatusCodeError(response.StatusCode)

// 	if JsonContentType(&response.Header) {
// 		if m, err := BufferToMap(response.Body); err == nil {
// 			statusCode := MapGetValue(m, "code")
// 			msg := MapGetValue(m, "message")
// 			id := SliceFirstNonNil(MapGetValue(m, "id"), MapGetValue(m, "gid"))
// 			if msg != nil && statusCode != nil {
// 				errMsg := fmt.Sprintf("%v: %v", statusCode, msg)
// 				if id != nil && response.StatusCode == 409 {
// 					return ConflictError{errMsg, id.(string)}
// 				} else {
// 					return ServerError(errMsg)
// 				}
// 			}
// 		}
// 	}

// 	return httpError
// }
