package main

import (
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type JobStatus string

const (
	Running JobStatus = "RUNNING"
	Stopped JobStatus = "STOPPED"
	Killed  JobStatus = "KILLED"
)

type Job struct {
	ID        string    `json:"id"`         // ID exposed to the client (UUID), NOT EMPTY, UNIQUE
	Pid       int       `json:"pid"`        // Unix process ID
	Command   []string  `json:"command"`    // command name + argv, NOT EMPTY
	Status    JobStatus `json:"status"`     // status of job, one of `RUNNING`, `KILLED` or `FINISHED`, NOT EMPTY
	Stdout    string    `json:"stdout"`     // process stdout
	Stderr    string    `json:"stderr"`     // process stderr
	ExitCode  int       `json:"exit_code"`  // process exit code
	CreatedAt time.Time `json:"created_at"` // time when job started, NOT EMPTY
	StoppedAt time.Time `json:"stopped_at"` // time when job is killed or has finished
}

type User struct {
	Token string         // the API token given to the user to access the API, will be generated using a CSPRNG, stored in hex or base64 format
	Jobs  map[string]Job // Index. list of jobs that belong to the user. Index key is the job ID.
	// Username string, // not necessary, already stored in the index
	// Password string, // not used, would be stored as hash using BCrypt
}

var usersIndex map[string]User // maps username to user struct

func checkUser(username string, password string) *User {
	user, ok := usersIndex[username]
	if !ok {
		return nil
	}

	// constant time comparison to avoid oracle attacks
	if subtle.ConstantTimeCompare([]byte(user.Token), []byte(password)) != 1 {
		return nil
	}

	return &user
}

func validateUserCredentials(w http.ResponseWriter, r *http.Request) *User {
	if username, password, ok := r.BasicAuth(); ok {
		if user := checkUser(username, password); user != nil {
			return user
		}
	}

	writeHTTPError(w, http.StatusUnauthorized, "Invalid user credentials")
	return nil
}

func writeHTTPError(w http.ResponseWriter, statusCode int, errorMessage string) {
	error := map[string]interface{}{
		"status":  statusCode,
		"message": errorMessage,
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(error)
}

func getJobs(w http.ResponseWriter, r *http.Request) {
	user := validateUserCredentials(w, r)
	if user == nil {
		return
	}

	jobs := user.Jobs
	jobsList := make([]Job, 0, len(jobs))

	for _, value := range jobs {
		jobsList = append(jobsList, value)
	}
	json.NewEncoder(w).Encode(jobsList)
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Use(commonMiddleware)

	myRouter.HandleFunc("/api/jobs", getJobs)

	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func main() {
	usersIndex = map[string]User{
		"user1": User{
			Token: "1234",
			Jobs: map[string]Job{
				"1": Job{
					"1",
					-1,
					[]string{"ls"},
					Running,
					"",
					"",
					0,
					time.Time{},
					time.Time{},
				},
			},
		},
	}

	handleRequests()
}
