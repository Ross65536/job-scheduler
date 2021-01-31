package internal

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func StartServer() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Use(commonMiddleware)

	myRouter.HandleFunc("/api/jobs", getJobs)

	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func validateUserCredentials(w http.ResponseWriter, r *http.Request) *User {
	if username, password, ok := r.BasicAuth(); ok {
		if user := GetIndexedUser(username, password); user != nil {
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

	jobs := user.GetAllJobs()
	json.NewEncoder(w).Encode(jobs)
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
