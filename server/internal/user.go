package internal

import (
	"crypto/subtle"
	"log"

	"github.com/google/uuid"
)

type User struct {
	token string         // the API token given to the user to access the API, will be generated using a CSPRNG, stored in hex or base64 format
	jobs  map[string]Job // Index. list of jobs that belong to the user. Index key is the job ID.
	// Username string, // not necessary, already stored in the index
	// Password string, // not used, would be stored as hash using BCrypt
}

var usersIndex map[string]User // maps username to user struct

func (u *User) GetAllJobs() []Job {
	jobsList := make([]Job, 0, len(u.jobs))

	for _, value := range u.jobs {
		jobsList = append(jobsList, value)
	}

	return jobsList
}

func generateNewID(keys map[string]Job) string {
	for i := 0; i < 1000; i++ {
		id := uuid.NewString()
		if _, ok := keys[id]; !ok {
			return id
		}
	}

	// in practice it shouldn't fail unless there is some logic error
	log.Panic("Failed to generate UUID")
	return "" // not reached
}

func (u *User) CreateJob(command []string, pid int) *Job {
	id := generateNewID(u.jobs)

	job := MakeJob(id, command, pid)
	u.jobs[id] = job

	return &job
}

func GetIndexedUser(username string, password string) *User {
	user, ok := usersIndex[username]
	if !ok {
		return nil
	}

	// constant time comparison to avoid oracle attacks
	if subtle.ConstantTimeCompare([]byte(user.token), []byte(password)) != 1 {
		return nil
	}

	return &user
}

func InitializeUsers() {
	usersIndex = map[string]User{
		"user1": User{
			token: "1234",
			jobs:  map[string]Job{},
		},
		"user2": User{
			token: "1234",
			jobs:  map[string]Job{},
		},
	}
}
