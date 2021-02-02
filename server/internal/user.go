package internal

import (
	"crypto/subtle"
	"log"
	"sync"

	"github.com/google/uuid"
)

type User struct {
	username string          // the username
	token    string          // the API token given to the user to access the API, will be generated using a CSPRNG, stored in base64 format
	jobsLock sync.RWMutex    // synchronizes access to the jobs map
	jobs     map[string]*Job // Index. list of jobs that belong to the user. Index key is the job ID.
}

var usersIndexLock = sync.RWMutex{} // synchronizes access to the 'usersIndex' global state
var usersIndex = map[string]*User{} // maps username to user struct

func (u *User) GetAllJobs() []*Job {
	u.jobsLock.RLock()
	defer u.jobsLock.RUnlock()

	jobsList := make([]*Job, 0, len(u.jobs))

	for _, value := range u.jobs {
		jobsList = append(jobsList, value)
	}

	return jobsList
}

func (u *User) IsTokenMatching(token string) bool {
	// not necessary to synchronize since 'Token' isn't supposed to be modified
	// on the lifetime of the server

	// constant time comparison to avoid oracle attacks
	return subtle.ConstantTimeCompare([]byte(u.token), []byte(token)) == 1
}

func (u *User) GetJob(jobID string) *Job {
	u.jobsLock.RLock()
	defer u.jobsLock.RUnlock()

	if job, ok := u.jobs[jobID]; ok {
		return job
	}

	return nil
}

func generateNewIDLocked(keys map[string]*Job) string {
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

func (u *User) AddJob(jobBuilder func(id string) *Job) *Job {
	u.jobsLock.Lock()

	id := generateNewIDLocked(u.jobs)
	job := jobBuilder(id)
	u.jobs[id] = job

	u.jobsLock.Unlock()

	return job
}

func GetIndexedUser(username string) *User {
	usersIndexLock.RLock()
	defer usersIndexLock.RUnlock()

	if user, ok := usersIndex[username]; ok {
		return user
	}

	return nil
}

func AddUser(username, token string) {
	usersIndexLock.Lock()

	usersIndex[username] = &User{
		username: username,
		token:    token,
		jobs:     map[string]*Job{},
	}

	usersIndexLock.Unlock()
}

func ClearUsers() {
	usersIndexLock.Lock()
	usersIndex = map[string]*User{}
	usersIndexLock.Unlock()
}
