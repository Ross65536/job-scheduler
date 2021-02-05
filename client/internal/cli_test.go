package internal_test

import (
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

	"github.com/ros-k/job-manager/client/internal"
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

func captureOutput(t *testing.T, f func()) string {
	rescueStdout := os.Stdout
	r, w, err := os.Pipe()
	assertNotError(t, err)

	os.Stdout = w

	f()

	w.Close()
	out, err := ioutil.ReadAll(r)
	assertNotError(t, err)
	os.Stdout = rescueStdout

	return string(out)
}

func TestCanShowJob(t *testing.T) {
	id := "123XYZ902"
	job := internal.JobViewFull{
		JobViewPartial: internal.JobViewPartial{
			JobViewCommand: internal.JobViewCommand{
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

	output := captureOutput(t, func() {
		err := internal.Start([]string{"client", "-c=http://user:pass@" + uri.Host, "show", id})
		assertNotError(t, err)
	})

	expected := `ls -l /, RUNNING, 2020-03-02 04:04:04 +0000 UTC -> <nil>, exit_code: -

STDOUT:
STDOUT123

STDERR:
STDERR456
`

	assertEquals(t, output, expected)
}

func TestServerError(t *testing.T) {
	returnError := internal.ErrorType{
		Status:  401,
		Message: "Invalid creds",
	}

	server, uri := setupTestServer(t, 401, encodeModel(t, returnError), "GET", "/api/jobs", "user", "pass")
	defer server.Close()

	err := internal.Start([]string{"client", "-c=http://user:pass@" + uri.Host, "list"})
	assertNotEquals(t, err, nil)

	assertEquals(t, err.Error(), "an error occurred (HTTP 401): "+returnError.Message)
}
