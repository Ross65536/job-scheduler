package client_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ros-k/job-manager/src/client"
)

const (
	jsonMime = "application/json"
)

func assertNotError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func assertContains(t *testing.T, actual, substr string) {
	if !strings.Contains(actual, substr) {
		t.Fatalf("String %s doesn't contain %s", actual, substr)
	}
}

func assertEquals(t *testing.T, actual interface{}, expected interface{}) {

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Invalid field, expected %s, was %s", expected, actual)
	}
}

func assertNotEquals(t *testing.T, actual interface{}, expected interface{}) {
	if reflect.DeepEqual(actual, expected) {
		t.Fatalf("Invalid field, expected %s to be different from %s", expected, actual)
	}
}

func encodeModel(t *testing.T, model interface{}) []byte {
	returnJson, err := json.Marshal(model)
	assertNotError(t, err)

	return returnJson
}

func setupTestServer(t *testing.T, returnStatusCode int, returnJson []byte, expectedMethod, expectedUriPath, expectedbasicUsername, expectedBasicPassword string) (*httptest.Server, *url.URL) {

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertEquals(t, expectedUriPath, r.URL.Path)
		assertEquals(t, expectedMethod, r.Method)

		user, pass, ok := r.BasicAuth()
		assertEquals(t, ok, true)
		assertEquals(t, user, expectedbasicUsername)
		assertEquals(t, pass, expectedBasicPassword)

		if returnJson != nil {
			w.Header().Set("Content-Type", jsonMime)
			w.WriteHeader(returnStatusCode)
			w.Write(returnJson)
		}
	})

	server := httptest.NewServer(handler)

	uri, err := url.ParseRequestURI(server.URL)
	assertNotError(t, err)

	return server, uri
}

func TestCanShowJob(t *testing.T) {
	id := "123XYZ902"
	job := client.JobViewFull{
		JobViewPartial: client.JobViewPartial{
			JobViewCommand: client.JobViewCommand{
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
	err := client.Start(&buf, []string{"client", "-c=http://user:pass@" + uri.Host, "show", id})
	assertNotError(t, err)

	output, err := ioutil.ReadAll(&buf)
	assertNotError(t, err)

	expected := `ls -l /, RUNNING, 2020-03-02 04:04:04 +0000 UTC -> <nil>, exit_code: -

STDOUT:
STDOUT123

STDERR:
STDERR456
`

	assertEquals(t, string(output), expected)
}

func TestServerError(t *testing.T) {
	returnError := client.ErrorType{
		Status:  401,
		Message: "Invalid creds",
	}

	server, uri := setupTestServer(t, 401, encodeModel(t, returnError), "GET", "/api/jobs", "user", "pass")
	defer server.Close()

	err := client.Start(os.Stdout, []string{"client", "-c=http://user:pass@" + uri.Host, "list"})
	assertNotEquals(t, err, nil)

	assertEquals(t, err.Error(), "an error occurred (HTTP 401): "+returnError.Message)
}
