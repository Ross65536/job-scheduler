package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
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
	httpClient *http.Client
}

func buildClientWithCA(caPath string) (*http.Client, error) {
	caCert, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, err
	}

	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		caCertPool = x509.NewCertPool()
	}

	caCertPool.AppendCertsFromPEM(caCert)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	return client, nil
}

func newHTTPClientCore(apiUrl string, client *http.Client) (*HTTPClient, error) {
	uri, err := url.ParseRequestURI(apiUrl)
	if err != nil {
		return nil, err
	}

	if !uri.IsAbs() {
		return nil, errors.New("uri must be absolute")
	}

	// TODO remove plain http support once tests connect with HTTPS
	if uri.Scheme != "http" && uri.Scheme != "https" {
		return nil, errors.New("invalid url scheme, must be HTTPS")
	}

	if uri.User == nil {
		return nil, errors.New("base URI must have an HTTP basic username and password encoded")
	}

	password, ok := uri.User.Password()
	if !ok {
		return nil, errors.New("base URI must have an HTTP basic password encoded")
	}

	return &HTTPClient{
		username:   uri.User.Username(),
		token:      password,
		apiUrl:     *uri,
		httpClient: client,
	}, nil
}

func NewHTTPClient(apiUrl string) (*HTTPClient, error) {
	return newHTTPClientCore(apiUrl, http.DefaultClient)
}

func NewHTTPClientWithCA(apiUrl, caPath string) (*HTTPClient, error) {
	client, err := buildClientWithCA(caPath)
	if err != nil {
		return nil, err
	}

	return newHTTPClientCore(apiUrl, client)
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
	resp, err := ioutil.ReadAll(response.Body)
	// close errors are ignored, since all data should be available
	response.Body.Close()

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
	if len(requestBody) != 0 {
		request.Header.Set("Content-Type", jsonMime+"; charset=utf-8")
	}

	return request, nil
}
