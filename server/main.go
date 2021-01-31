package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Job struct {
	ID        string    `json:"id"`         // ID exposed to the client (UUID), NOT EMPTY, UNIQUE
	Pid       int       `json:"pid"`        // Unix process ID
	Command   []string  `json:"command"`    // command name + argv, NOT EMPTY
	Status    string    `json:"status"`     // status of job, one of `RUNNING`, `KILLED` or `FINISHED`, NOT EMPTY
	Stdout    string    `json:"stdout"`     // process stdout
	Stderr    string    `json:"stderr"`     // process stderr
	ExitCode  int       `json:"exit_code"`  // process exit code
	CreatedAt time.Time `json:"created_at"` // time when job started, NOT EMPTY
	StoppedAt time.Time `json:"stopped_at"` // time when job is killed or has finished
}

type User struct {
	Username string
	Token    string         // the API token given to the user to access the API, will be generated using a CSPRNG, stored in hex or base64 format
	Jobs     map[string]Job // Index. list of jobs that belong to the user. Index key is the job ID.
	// Password string, // not used, would be stored as hash using BCrypt
}

var usersIndex map[string]User // maps username to user struct

func getJobs(w http.ResponseWriter, r *http.Request) {
	jobs := usersIndex["user1"].Jobs
	fmt.Println("Endpoint Hit: getJobs")

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
			Username: "user1",
			Token:    "1234",
			Jobs: map[string]Job{
				"1": Job{
					"1",
					-1,
					[]string{"ls"},
					"",
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
