package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
		t.Fatalf("Invalid field, expected %v, was %v", expected, actual)
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

	assertEquals(t, resp.StatusCode, expectedStatus)

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

func setupTest(t *testing.T, basic httpBasic) (*internal.State, *httptest.Server) {
	state := internal.NewState()
	state.AddUser(basic.username, basic.password)
	server, err := internal.NewServer(state)
	assertNotError(t, err)

	router := server.GetRouter()
	return state, httptest.NewServer(router)
}

func teardownTest(state *internal.State, server *httptest.Server) {
	state.ClearUsers()
	server.Close()
}

func TestCanCreateJob(t *testing.T) {
	basic := buildDefaultUser()

	state, server := setupTest(t, basic)
	defer teardownTest(state, server)

	targetStr := "foobaz"

	command := `{"command": ["sh", "-c", "echo '` + targetStr + `'"]}`
	resp := makeRequestWithHttpBasic(t, basic, "POST", server.URL+"/api/jobs", command, 201)
	jsonResponse := parseJsonObj(t, resp)
	assertEquals(t, jsonResponse["status"], "RUNNING")

	id := jsonResponse["id"].(string)

	limitedWait(t, func() bool {
		resp = makeRequestWithHttpBasic(t, basic, "GET", server.URL+"/api/jobs/"+id, "", 200)
		jsonResponse = parseJsonObj(t, resp)
		status := jsonResponse["status"].(string)

		if status != string(internal.JobFinished) {
			return false
		}

		stdout := jsonResponse["stdout"].(string)
		assertEquals(t, stdout, targetStr+"\n")
		// if !strings.Contains(stdout, "etc") { // assuming all tests environs have this folder
		// 	t.Fatal("Finished job returned unexpected result")
		// }

		return true
	})
}

func TestFailToCreateJob(t *testing.T) {
	basic := buildDefaultUser()

	state, server := setupTest(t, basic)
	defer teardownTest(state, server)

	command := `{"command": ["ls_invalid_program_1223323"]}` // assumed to be an invalid program
	resp := makeRequestWithHttpBasic(t, basic, "POST", server.URL+"/api/jobs", command, 500)
	jsonResponse := parseJsonObj(t, resp)
	assertEquals(t, jsonResponse["status"], 500.0)
}

func TestInvalidAuth(t *testing.T) {
	basic := buildDefaultUser()

	state, server := setupTest(t, basic)
	defer teardownTest(state, server)

	invalidAuth := httpBasic{
		username: basic.username,
		password: basic.password + "invalid",
	}

	makeRequestWithHttpBasic(t, invalidAuth, "GET", server.URL+"/api/jobs", "", 401)
}
