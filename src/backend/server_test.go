package backend_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ros-k/job-manager/src/backend"
	"github.com/ros-k/job-manager/src/core/testutil"
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

var client = http.Client{}

func makeRequestWithHttpBasic(t *testing.T, basic httpBasic, method, url, body string, expectedStatus int) *http.Response {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewBuffer([]byte(body))
	}

	req, err := http.NewRequest(method, url, bodyReader)
	testutil.AssertNotError(t, err)

	req.SetBasicAuth(basic.username, basic.password)

	resp, err := client.Do(req)
	testutil.AssertNotError(t, err)

	testutil.AssertEquals(t, resp.StatusCode, expectedStatus)

	return resp
}

func parseJsonObj(t *testing.T, resp *http.Response) map[string]interface{} {
	reqBody, err := ioutil.ReadAll(resp.Body)
	testutil.AssertNotError(t, err)

	var jsonResponse map[string]interface{}
	err = json.Unmarshal(reqBody, &jsonResponse)
	testutil.AssertNotError(t, err)

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

func setupTest(t *testing.T, basic httpBasic) (*backend.State, *httptest.Server) {
	state := backend.NewState()
	state.AddUser(basic.username, basic.password)
	server, err := backend.NewServer(state)
	testutil.AssertNotError(t, err)

	router := server.GetRouter()
	return state, httptest.NewServer(router)
}

func teardownTest(state *backend.State, server *httptest.Server) {
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
	testutil.AssertEquals(t, jsonResponse["status"], "RUNNING")

	id := jsonResponse["id"].(string)

	limitedWait(t, func() bool {
		resp = makeRequestWithHttpBasic(t, basic, "GET", server.URL+"/api/jobs/"+id, "", 200)
		jsonResponse = parseJsonObj(t, resp)
		status := jsonResponse["status"].(string)

		if status != string(backend.JobFinished) {
			return false
		}

		stdout := jsonResponse["stdout"].(string)
		testutil.AssertEquals(t, stdout, targetStr+"\n")

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
	testutil.AssertEquals(t, jsonResponse["status"], 500.0)
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
