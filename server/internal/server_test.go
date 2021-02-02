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

type httpBasic struct {
	username string
	password string
}

func buildDefaultUser() httpBasic {
	return httpBasic{
		username: "user1",
		password: "1234",
	}
}

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

func makeRequestWithHttpBasic(t *testing.T, basic httpBasic, method, url, body string, expectedStatus int) *http.Response {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewBuffer([]byte(body))
	}

	req, err := http.NewRequest(method, url, bodyReader)
	assertNotError(t, err)

	req.SetBasicAuth(basic.username, basic.password)

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

func setupTest(basic httpBasic) (*internal.State, *httptest.Server) {
	state := internal.NewState()
	state.AddUser(basic.username, basic.password)
	router := internal.CreateRouter(state)
	return state, httptest.NewServer(router)
}

func teardownTest(state *internal.State, server *httptest.Server) {
	state.ClearUsers()
	server.Close()
}

func TestCanCreateJob(t *testing.T) {
	basic := buildDefaultUser()

	state, server := setupTest(basic)
	defer teardownTest(state, server)

	command := `{"command": ["ls", "/"]}`
	resp := makeRequestWithHttpBasic(t, basic, "POST", server.URL+"/api/jobs", command, 201)
	jsonResponse := parseJsonObj(t, resp)
	assertEquals(t, jsonResponse["status"], "RUNNING")

	id := jsonResponse["id"].(string)

	limitedWait(t, func() bool {
		resp = makeRequestWithHttpBasic(t, basic, "GET", server.URL+"/api/jobs/"+id, "", 200)
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

func TestFailToCreateJob(t *testing.T) {
	basic := buildDefaultUser()

	state, server := setupTest(basic)
	defer teardownTest(state, server)

	command := `{"command": ["ls_invalid_program_1223323"]}` // assumed to be an invalid program
	resp := makeRequestWithHttpBasic(t, basic, "POST", server.URL+"/api/jobs", command, 500)
	jsonResponse := parseJsonObj(t, resp)
	assertEquals(t, jsonResponse["status"], 500.0)
}

func TestInvalidAuth(t *testing.T) {
	basic := buildDefaultUser()

	state, server := setupTest(basic)
	defer teardownTest(state, server)

	invalidAuth := httpBasic{
		username: basic.username,
		password: basic.password + "invalid",
	}

	makeRequestWithHttpBasic(t, invalidAuth, "GET", server.URL+"/api/jobs", "", 401)
}
