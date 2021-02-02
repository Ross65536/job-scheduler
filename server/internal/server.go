package internal

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func CreateRouter(state *State) http.Handler {
	router := mux.NewRouter().StrictSlash(true)
	// TODO: add checks/validation for 'Accept', 'Content-Type' client headers

	topRouter := router.PathPrefix("/api/jobs").Subrouter()
	topRouter.HandleFunc("", authMiddleware(state, getJobs)).Methods("GET")
	topRouter.HandleFunc("", authMiddleware(state, createJob)).Methods("POST")

	jobsRouter := router.PathPrefix("/api/jobs/{id}").Subrouter()
	jobsRouter.HandleFunc("", authMiddleware(state, jobIDMiddleware(getJob))).Methods("GET")
	jobsRouter.HandleFunc("", authMiddleware(state, jobIDMiddleware(stopJob))).Methods("DELETE")

	return router
}

func StartServer(state *State) {
	router := CreateRouter(state)

	log.Fatal(http.ListenAndServe(":10000", router))
}

func checkAuth(state *State, r *http.Request) (*User, error) {
	username, password, ok := r.BasicAuth()
	if !ok {
		return nil, errors.New("Request isn't using HTTP Basic")
	}

	user := state.GetIndexedUser(username)
	if user == nil {
		return nil, errors.New("Invalid username")
	}

	if !user.IsTokenMatching(password) {
		return user, errors.New("Invalid token")
	}

	return user, nil
}

func authMiddleware(state *State, next func(http.ResponseWriter, *http.Request, *User)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if user, err := checkAuth(state, r); err != nil {
			log.Printf("Invalid user tried to access API: %s", err)
			writeJSONError(w, http.StatusUnauthorized, "Invalid user credentials")
		} else {
			next(w, r, user)
		}
	}
}

func jobIDMiddleware(next func(http.ResponseWriter, *http.Request, *Job)) func(w http.ResponseWriter, r *http.Request, user *User) {
	return func(w http.ResponseWriter, r *http.Request, user *User) {
		id := mux.Vars(r)["id"]
		job := user.GetJob(id)
		if job == nil {
			writeJSONError(w, http.StatusNotFound, "invalid job ID")
			return
		}

		next(w, r, job)
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, model interface{}) {
	if json, err := json.Marshal(model); err != nil {
		log.Printf("Something went wrong returning a JSON response to the user %s", err)
		writeJSONError(w, http.StatusInternalServerError, "Something went wrong") // this shouldn't fail
	} else {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		w.Write(json)
	}
}

func writeJSONError(w http.ResponseWriter, statusCode int, errorMessage string) {
	error := map[string]interface{}{
		"status":  statusCode,
		"message": errorMessage,
	}

	writeJSON(w, statusCode, error)
}

func getJob(w http.ResponseWriter, r *http.Request, job *Job) {
	writeJSON(w, http.StatusOK, job.AsView())
}

func stopJob(w http.ResponseWriter, r *http.Request, job *Job) {
	if err := StopJob(job); err != nil {
		log.Printf("Something went wrong stopping job: %s", err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to stop job")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getJobs(w http.ResponseWriter, r *http.Request, user *User) {
	jobs := user.GetAllJobs()
	jobViews := make([]JobViewPartial, 0, len(jobs))

	for _, v := range jobs {
		jobViews = append(jobViews, v.AsView().JobViewPartial)
	}

	writeJSON(w, http.StatusOK, jobViews)
}

func isCommandValid(command []string) error {
	if command == nil {
		return errors.New("Invalid JSON schema: command not present")
	}

	if len(command) == 0 {
		return errors.New("Invalid JSON schema: command must have at least 1 element")
	}

	if program := command[0]; len(strings.TrimSpace(program)) == 0 {
		return errors.New("Invalid JSON schema: command must have at least 1 non-empty element")
	}

	// TODO: add more sanity checks like invalid characters, etc

	return nil
}

func parseJobCreation(r io.Reader) ([]string, error) {
	reqBody, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	createJob := JobViewCommand{}
	// var createJob JobCreateModel
	if err := json.Unmarshal(reqBody, &createJob); err != nil {
		return nil, err
	}

	if err := isCommandValid(createJob.Command); err != nil {
		return nil, err
	}

	return createJob.Command, nil
}

func createJob(w http.ResponseWriter, r *http.Request, user *User) {
	command, err := parseJobCreation(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusUnprocessableEntity, "Invalid or missing 'command' in POST body")
		return
	}

	if job, err := SpawnJob(user, command); err != nil {
		log.Printf("Failed to start job %s, because: %s", command, err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to start job")
	} else {
		user.AddJob(job)
		writeJSON(w, http.StatusCreated, job.AsView().JobViewPartial)
	}
}
