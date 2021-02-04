package internal_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ros-k/job-manager/client/internal"
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
	if actual != expected {
		t.Fatalf("Invalid field, expected %s, was %s", expected, actual)
	}
}

func setupTestServer(t *testing.T, returnJson []byte, expectedMethod, expectedUriPath, expectedbasicUsername, expectedBasicPassword string) *httptest.Server {
	const jsonMime = "application/json"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertEquals(t, expectedUriPath, r.URL.Path)
		assertEquals(t, expectedMethod, r.Method)

		user, pass, ok := r.BasicAuth()
		assertEquals(t, ok, true)
		assertEquals(t, user, expectedbasicUsername)
		assertEquals(t, pass, expectedBasicPassword)

		if returnJson != nil {
			w.Header().Set("Content-Type", jsonMime)
			w.Write(returnJson)
		}
	})

	return httptest.NewServer(handler)
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
			CreatedAt: time.Now(),
		},
		Stdout: "STDOUT123",
		Stderr: "STDERR456",
	}

	requestJson, err := json.Marshal(job)
	assertNotError(t, err)

	server := setupTestServer(t, requestJson, "GET", "/api/jobs/"+id, "user", "pass")
	defer server.Close()

	uri, err := url.ParseRequestURI(server.URL)
	assertNotError(t, err)

	output := captureOutput(t, func() {
		err = internal.Start([]string{"client", "show", id, "-c=http://user:pass@" + uri.Host})
		assertNotError(t, err)
	})

	assertContains(t, output, "RUNNING")
	assertContains(t, output, "STDOUT123")
	assertContains(t, output, "STDERR456")
}
