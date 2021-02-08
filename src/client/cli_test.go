package client_test

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/ros-k/job-manager/src/client"
	"github.com/ros-k/job-manager/src/core/testutil"
	"github.com/ros-k/job-manager/src/core/view"
)

const (
	jsonMime = "application/json"
)

func encodeModel(t *testing.T, model interface{}) []byte {
	returnJson, err := json.Marshal(model)
	testutil.AssertNotError(t, err)

	return returnJson
}

func TestMain(m *testing.M) {
	// ignore TLS verification from client
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	code := m.Run()
	os.Exit(code)
}

func setupTestServer(t *testing.T, returnStatusCode int, returnJson []byte, expectedMethod, expectedUriPath, expectedbasicUsername, expectedBasicPassword string) (*httptest.Server, *url.URL) {

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testutil.AssertEquals(t, expectedUriPath, r.URL.Path)
		testutil.AssertEquals(t, expectedMethod, r.Method)

		user, pass, ok := r.BasicAuth()
		testutil.AssertEquals(t, ok, true)
		testutil.AssertEquals(t, user, expectedbasicUsername)
		testutil.AssertEquals(t, pass, expectedBasicPassword)

		if returnJson == nil {
			return
		}

		w.Header().Set("Content-Type", jsonMime)
		w.WriteHeader(returnStatusCode)
		w.Write(returnJson)
	})

	server := httptest.NewTLSServer(handler)

	uri, err := url.ParseRequestURI(server.URL)
	testutil.AssertNotError(t, err)

	return server, uri
}

func TestCanShowJob(t *testing.T) {
	id := "123XYZ902"
	job := view.JobViewFull{
		JobViewPartial: view.JobViewPartial{
			JobViewCommand: view.JobViewCommand{
				Command: []string{"ls", "-l", "/"},
			},
			ID:        id,
			Status:    "RUNNING",
			CreatedAt: time.Date(2020, time.March, 2, 4, 4, 4, 0, time.UTC),
		},
		Stdout: "STDOUT123",
		Stderr: "STDERR456",
	}

	server, uri := setupTestServer(t, 200, encodeModel(t, job), "GET", "/api/jobs/"+id, "user", "pass")
	defer server.Close()

	buf := bytes.Buffer{}
	err := client.Start(&buf, []string{"client", "-ca=", "-c=https://user:pass@" + uri.Host, "show", id})
	testutil.AssertNotError(t, err)

	output := buf.String()

	expected := `ls -l /, RUNNING, 2020-03-02 04:04:04 +0000 UTC -> <nil>, exit_code: -

STDOUT:
STDOUT123

STDERR:
STDERR456
`

	testutil.AssertEquals(t, string(output), expected)
}

func TestServerError(t *testing.T) {
	returnError := client.ErrorType{
		Status:  401,
		Message: "Invalid creds",
	}

	server, uri := setupTestServer(t, 401, encodeModel(t, returnError), "GET", "/api/jobs", "user", "pass")
	defer server.Close()

	err := client.Start(os.Stdout, []string{"client", "-ca=", "-c=https://user:pass@" + uri.Host, "list"})
	testutil.AssertNotEquals(t, err, nil)

	testutil.AssertEquals(t, err.Error(), "an error occurred (HTTP 401): "+returnError.Message)
}
