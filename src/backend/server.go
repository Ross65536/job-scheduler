package backend

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ros-k/job-manager/src/core"
)

type Server struct {
	state  *State
	router *mux.Router
}

func (s *Server) GetRouter() http.Handler {
	return s.router
}

func (s *Server) addRoutes() {
	// TODO: add checks/validation for 'Accept', 'Content-Type' client headers

	topRouter := s.router.PathPrefix("/api/jobs").Subrouter()
	topRouter.HandleFunc("", s.authMiddleware(s.getJobs)).Methods("GET")
	topRouter.HandleFunc("", s.authMiddleware(s.createJob)).Methods("POST")

	jobsRouter := s.router.PathPrefix("/api/jobs/{id}").Subrouter()
	jobsRouter.HandleFunc("", s.authMiddleware(s.jobIDMiddleware(s.getJob))).Methods("GET")
	jobsRouter.HandleFunc("", s.authMiddleware(s.jobIDMiddleware(s.stopJob))).Methods("DELETE")
}

func NewServer(state *State) (*Server, error) {
	s := &Server{state: state, router: mux.NewRouter().StrictSlash(true)}
	s.addRoutes()

	return s, nil
}

func (s *Server) Start() error {
	return http.ListenAndServe(":10000", s.router)
}

func (s *Server) checkAuth(r *http.Request) (*User, error) {
	username, password, ok := r.BasicAuth()
	if !ok {
		return nil, errors.New("Request isn't using HTTP Basic")
	}

	user := s.state.GetIndexedUser(username)
	if user == nil {
		return nil, errors.New("Invalid username")
	}

	if !user.IsTokenMatching(password) {
		return user, errors.New("Invalid token")
	}

	return user, nil
}

func (s *Server) authMiddleware(next func(http.ResponseWriter, *http.Request, *User)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if user, err := s.checkAuth(r); err != nil {
			log.Printf("Invalid user tried to access API: %s", err)
			WriteJSONError(w, http.StatusUnauthorized, "Invalid user credentials")
		} else {
			next(w, r, user)
		}
	}
}

func (s *Server) jobIDMiddleware(next func(http.ResponseWriter, *http.Request, *Job)) func(w http.ResponseWriter, r *http.Request, user *User) {
	return func(w http.ResponseWriter, r *http.Request, user *User) {
		id := mux.Vars(r)["id"]
		job := user.GetJob(id)
		if job == nil {
			WriteJSONError(w, http.StatusNotFound, "invalid job ID")
			return
		}

		next(w, r, job)
	}
}

func (s *Server) getJob(w http.ResponseWriter, r *http.Request, job *Job) {
	WriteJSON(w, http.StatusOK, job.AsView())
}

func (s *Server) stopJob(w http.ResponseWriter, r *http.Request, job *Job) {
	if err := job.StopJob(); err != nil {
		log.Printf("Something went wrong stopping job: %s", err)
		WriteJSONError(w, http.StatusInternalServerError, "Failed to stop job")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) getJobs(w http.ResponseWriter, r *http.Request, user *User) {
	jobs := user.GetAllJobs()
	jobViews := make([]core.JobViewPartial, 0, len(jobs))

	for _, v := range jobs {
		jobViews = append(jobViews, v.AsView().JobViewPartial)
	}

	WriteJSON(w, http.StatusOK, jobViews)
}

func isCommandValid(command []string) error {
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

	createJob := core.JobViewCommand{}
	if err := json.Unmarshal(reqBody, &createJob); err != nil {
		return nil, err
	}

	if err := isCommandValid(createJob.Command); err != nil {
		return nil, err
	}

	return createJob.Command, nil
}

func (s *Server) createJob(w http.ResponseWriter, r *http.Request, user *User) {
	command, err := parseJobCreation(r.Body)
	if err != nil {
		WriteJSONError(w, http.StatusUnprocessableEntity, "Invalid or missing 'command' in POST body")
		return
	}

	if job, err := SpawnJob(user, command); err != nil {
		log.Printf("Failed to start job %s, because: %s", command, err)
		WriteJSONError(w, http.StatusInternalServerError, "Failed to start job")
	} else {
		user.AddJob(job)
		WriteJSON(w, http.StatusCreated, job.AsView().JobViewPartial)
	}
}
