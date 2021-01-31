package internal

import (
	"context"
	"encoding/json"
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

func StartServer() {
	router := mux.NewRouter().StrictSlash(true)
	router.Use(authMiddleware)

	topRouter := router.PathPrefix("/api/jobs").Subrouter()
	topRouter.HandleFunc("", getJobs).Methods("GET")
	topRouter.HandleFunc("", createJob).Methods("POST")

	jobsRouter := router.PathPrefix("/api/jobs/{id}").Subrouter()
	jobsRouter.Use(jobIdMiddleware)
	jobsRouter.HandleFunc("", getJob).Methods("GET")

	log.Fatal(http.ListenAndServe(":10000", router))
}

func validateUserCredentials(w http.ResponseWriter, r *http.Request) *User {
	if username, password, ok := r.BasicAuth(); ok {
		if user := GetIndexedUser(username, password); user != nil {
			return user
		}

		log.Print("Client supplied invalid username or password")
	} else {
		log.Print("Client didn't provide Basic auth")
	}

	writeHTTPError(w, http.StatusUnauthorized, "Invalid user credentials")
	return nil
}

func jobIdMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getContextUser(r)
		id := mux.Vars(r)["id"]

		job := user.GetJob(id)
		if job == nil {
			writeHTTPError(w, http.StatusNotFound, "invalid ID")
			return
		}

		ctx := context.WithValue(r.Context(), jobCtxKey, job)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := validateUserCredentials(w, r)
		if user == nil {
			return
		}

		ctx := context.WithValue(r.Context(), userCtxKey, user)
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
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(model)
}

func writeHTTPError(w http.ResponseWriter, statusCode int, errorMessage string) {
	error := map[string]interface{}{
		"status":  statusCode,
		"message": errorMessage,
	}

	writeJSON(w, statusCode, error)
}

func getJob(w http.ResponseWriter, r *http.Request) {
	job := getContextJob(r)

	writeJSON(w, http.StatusOK, job)
}

func getJobs(w http.ResponseWriter, r *http.Request) {
	user := getContextUser(r)

	jobs := user.GetAllJobs()
	writeJSON(w, http.StatusOK, jobs)
}

type jobCreateModel struct {
	Command []string
}

func (j *jobCreateModel) isValid() bool {
	if j.Command == nil || len(j.Command) < 1 {
		return false
	}

	if program := j.Command[0]; len(strings.TrimSpace(program)) == 0 {
		return false
	}

	// TODO: add more sanity checks like invalid characters, etc

	return true
}

func createJob(w http.ResponseWriter, r *http.Request) {
	user := getContextUser(r)

	reqBody, _ := ioutil.ReadAll(r.Body)
	var createJob jobCreateModel
	json.Unmarshal(reqBody, &createJob)
	if !createJob.isValid() {
		writeHTTPError(w, http.StatusUnprocessableEntity, "Invalid or missing 'command'")
		return
	}

	job := user.CreateJob(createJob.Command, -1)

	writeJSON(w, http.StatusCreated, job)
}
