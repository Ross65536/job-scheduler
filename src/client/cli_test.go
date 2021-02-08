package client_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/ros-k/job-manager/src/client"
	"github.com/ros-k/job-manager/src/core"
)

const (
	jsonMime = "application/json"
)

func encodeModel(t *testing.T, model interface{}) []byte {
	returnJson, err := json.Marshal(model)
	core.AssertNotError(t, err)

	return returnJson
}

func setupTestServer(t *testing.T, returnStatusCode int, returnJson []byte, expectedMethod, expectedUriPath, expectedbasicUsername, expectedBasicPassword string) (*httptest.Server, *url.URL) {

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		core.AssertEquals(t, expectedUriPath, r.URL.Path)
		core.AssertEquals(t, expectedMethod, r.Method)

		user, pass, ok := r.BasicAuth()
		core.AssertEquals(t, ok, true)
		core.AssertEquals(t, user, expectedbasicUsername)
		core.AssertEquals(t, pass, expectedBasicPassword)

		if returnJson == nil {
			return
		}

		w.Header().Set("Content-Type", jsonMime)
		w.WriteHeader(returnStatusCode)
		w.Write(returnJson)
	})

	server := httptest.NewServer(handler)

	uri, err := url.ParseRequestURI(server.URL)
	core.AssertNotError(t, err)

	return server, uri
}

func TestCanShowJob(t *testing.T) {
	id := "123XYZ902"
	job := core.JobViewFull{
		JobViewPartial: core.JobViewPartial{
			JobViewCommand: core.JobViewCommand{
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
	err := client.Start(&buf, []string{"client", "-ca=", "-c=http://user:pass@" + uri.Host, "show", id})
	core.AssertNotError(t, err)

	output := buf.String()

	expected := `ls -l /, RUNNING, 2020-03-02 04:04:04 +0000 UTC -> <nil>, exit_code: -

STDOUT:
STDOUT123

STDERR:
STDERR456
`

	core.AssertEquals(t, string(output), expected)
}

func TestServerError(t *testing.T) {
	returnError := client.ErrorType{
		Status:  401,
		Message: "Invalid creds",
	}

	server, uri := setupTestServer(t, 401, encodeModel(t, returnError), "GET", "/api/jobs", "user", "pass")
	defer server.Close()

	err := client.Start(os.Stdout, []string{"client", "-ca=", "-c=http://user:pass@" + uri.Host, "list"})
	core.AssertNotEquals(t, err, nil)

	core.AssertEquals(t, err.Error(), "an error occurred (HTTP 401): "+returnError.Message)
}
