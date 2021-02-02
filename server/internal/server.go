package internal

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

const (
	userCtxKey = "user"
	jobCtxKey  = "job"
)

func CreateRouter() http.Handler {
	router := mux.NewRouter().StrictSlash(true)
	router.Use(authMiddleware)
	// TODO: add checks/validation for 'Accept', 'Content-Type' client headers

	topRouter := router.PathPrefix("/api/jobs").Subrouter()
	topRouter.HandleFunc("", getJobs).Methods("GET")
	topRouter.HandleFunc("", createJob).Methods("POST")

	jobsRouter := router.PathPrefix("/api/jobs/{id}").Subrouter()
	jobsRouter.Use(jobIDMiddleware)
	jobsRouter.HandleFunc("", getJob).Methods("GET")
	jobsRouter.HandleFunc("", stopJob).Methods("DELETE")

	return router
}

func StartServer() {
	router := CreateRouter()

	log.Fatal(http.ListenAndServe(":10000", router))
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if username, password, ok := r.BasicAuth(); ok {
			if user := GetIndexedUser(username); user != nil && user.IsTokenMatching(password) {
				ctx := context.WithValue(r.Context(), userCtxKey, user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			log.Print("Client supplied invalid username or password")
		} else {
			log.Print("Client didn't provide Basic auth")
		}

		writeJSONError(w, http.StatusUnauthorized, "Invalid user credentials")
	})
}

func jobIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getContextUser(r)
		id := mux.Vars(r)["id"]

		job := user.GetJob(id)
		if job == nil {
			writeJSONError(w, http.StatusNotFound, "invalid ID")
			return
		}

		ctx := context.WithValue(r.Context(), jobCtxKey, job)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getContextUser(r *http.Request) *User {
	v := r.Context().Value(userCtxKey)
	if v == nil {
		log.Panic("Logic error")
	}

	user, ok := v.(*User)
	if !ok {
		log.Panic("Logic error")
	}

	return user
}

func getContextJob(r *http.Request) *Job {
	v := r.Context().Value(jobCtxKey)
	if v == nil {
		log.Panic("Logic error")
	}

	job, ok := v.(*Job)
	if !ok {
		log.Panic("Logic error")
	}

	return job
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

var showJobFields = []string{"id", "status", "created_at", "exit_code", "command", "stopped_at", "stdout", "stderr"}

func getJob(w http.ResponseWriter, r *http.Request) {
	job := getContextJob(r)

	writeJSON(w, http.StatusOK, MapSubmap(job.AsMap(), showJobFields...))
}

func stopJob(w http.ResponseWriter, r *http.Request) {
	job := getContextJob(r)

	if err := StopJob(job); err != nil {
		log.Printf("Something went wrong stopping job: %s", err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to stop job")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

var listJobFields = []string{"id", "status", "created_at", "exit_code", "command", "stopped_at"}

func getJobs(w http.ResponseWriter, r *http.Request) {
	user := getContextUser(r)

	jobs := user.GetAllJobs()
	jobViews := make([]map[string]interface{}, 0, len(jobs))

	for _, v := range jobs {
		view := MapSubmap(v.AsMap(), listJobFields...)
		jobViews = append(jobViews, view)
	}

	writeJSON(w, http.StatusOK, jobViews)
}

func isCommandValid(command []string) bool {
	if len(command) == 0 {
		return false
	}

	if program := command[0]; len(strings.TrimSpace(program)) == 0 {
		return false
	}

	// TODO: add more sanity checks like invalid characters, etc

	return true
}

func parseJobCreation(r io.Reader) ([]string, error) {
	reqBody, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	createJob := struct {
		Command []string
	}{}
	// var createJob JobCreateModel
	if err := json.Unmarshal(reqBody, &createJob); err != nil {
		return nil, err
	}

	if !isCommandValid(createJob.Command) {
		return nil, errors.New("Invalid JSON schema")
	}

	return createJob.Command, nil
}

var createdJobFields = []string{"id", "status", "created_at", "command"}

func createJob(w http.ResponseWriter, r *http.Request) {
	user := getContextUser(r)

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
		writeJSON(w, http.StatusCreated, MapSubmap(job.AsMap(), createdJobFields...))
	}
}
