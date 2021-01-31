package internal

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	userCtxKey = "user"
	jobCtxKey  = "job"
)

func CreateRouter() *mux.Router {
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
			if user := GetIndexedUser(username, password); user != nil {
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
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(model)
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

	if ok := StopJob(job); !ok {
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

var createdJobFields = []string{"id", "status", "created_at", "command"}

func createJob(w http.ResponseWriter, r *http.Request) {
	user := getContextUser(r)

	createJob := ParseJobCreation(r.Body)
	if createJob == nil || !createJob.IsValid() {
		writeJSONError(w, http.StatusUnprocessableEntity, "Invalid or missing 'command' in POST body")
		return
	}

	ch := make(chan *Job, 1)
	go SpawnJob(createJob.Command, ch)

	job := <-ch
	if job == nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to start job")
		return
	}

	user.AddJob(job)

	writeJSON(w, http.StatusCreated, MapSubmap(job.AsMap(), createdJobFields...))
}
