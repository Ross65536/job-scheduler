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
)

func StartServer() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Use(authMiddleware)

	myRouter.HandleFunc("/api/jobs", getJobs).Methods("GET")
	myRouter.HandleFunc("/api/jobs", createJob).Methods("POST")

	log.Fatal(http.ListenAndServe(":10000", myRouter))
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

func getUser(r *http.Request) *User {
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

func getJobs(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)

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
	user := getUser(r)

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
