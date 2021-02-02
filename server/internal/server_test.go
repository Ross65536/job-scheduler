package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ros-k/job-manager/server/internal"
)

func assertEquals(t *testing.T, actual interface{}, expected interface{}) {
	if actual != expected {
		t.Fatalf("Invalid field, expected %s, was %s", expected, actual)
	}
}

func assertNotError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

var client = http.Client{}

func makeRequestWithHttpBasic(t *testing.T, basicUsername string, basicPassword string, method string, url string, body string, expectedStatus int) *http.Response {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewBuffer([]byte(body))
	}

	req, err := http.NewRequest(method, url, bodyReader)
	assertNotError(t, err)

	req.SetBasicAuth(basicUsername, basicPassword)

	resp, err := client.Do(req)
	assertNotError(t, err)

	if resp.StatusCode != expectedStatus {
		t.Fatalf("Received non-%d response: %d\n", expectedStatus, resp.StatusCode)
	}

	return resp
}

func parseJsonObj(t *testing.T, resp *http.Response) map[string]interface{} {
	reqBody, err := ioutil.ReadAll(resp.Body)
	assertNotError(t, err)

	var jsonResponse map[string]interface{}
	err = json.Unmarshal(reqBody, &jsonResponse)
	assertNotError(t, err)

	return jsonResponse
}

func limitedWait(t *testing.T, body func() bool) {
	for i := 0; i < 20; i++ {
		time.Sleep(50 * time.Millisecond)
		if body() {
			return
		}
	}

	t.Fatal("Timeout waiting for response")
}

// implicitly also tests authentication
func TestCreateJobHappyPath(t *testing.T) {
	username := "user1"
	token := "1234"
	internal.AddUser(username, token)
	defer internal.ClearUsers()

	router := internal.CreateRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	command := `{"command": ["ls", "/"]}`
	resp := makeRequestWithHttpBasic(t, username, token, "POST", server.URL+"/api/jobs", command, 201)
	jsonResponse := parseJsonObj(t, resp)
	assertEquals(t, jsonResponse["status"], "RUNNING")

	id := jsonResponse["id"].(string)

	limitedWait(t, func() bool {
		resp = makeRequestWithHttpBasic(t, username, token, "GET", server.URL+"/api/jobs/"+id, "", 200)
		jsonResponse = parseJsonObj(t, resp)
		status := jsonResponse["status"].(string)
		if status != "FINISHED" {
			return false
		}

		stdout := jsonResponse["stdout"].(string)
		if !strings.Contains(stdout, "etc") { // assuming all tests environs have this folder
			t.Fatal("Finished job returned unexpected result")
		}

		return true
	})
}

func TestCreateJobUnhappyPath(t *testing.T) {
	username := "user1"
	token := "1234"
	internal.AddUser(username, token)
	defer internal.ClearUsers()

	router := internal.CreateRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	command := `{"command": ["ls_invalid_program_1223323"]}` // assumed to be an invalid program
	resp := makeRequestWithHttpBasic(t, username, token, "POST", server.URL+"/api/jobs", command, 500)
	jsonResponse := parseJsonObj(t, resp)
	assertEquals(t, jsonResponse["status"], 500.0)
}

func TestInvalidAuth(t *testing.T) {
	username := "user1"
	token := "1234"
	internal.AddUser(username, token)
	defer internal.ClearUsers()

	router := internal.CreateRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	makeRequestWithHttpBasic(t, username, token+"invalid", "GET", server.URL+"/api/jobs", "", 401)
}
