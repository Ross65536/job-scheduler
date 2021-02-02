package internal

import (
	"crypto/subtle"
	"sync"
)

type User struct {
	username string          // the username
	token    string          // the API token given to the user to access the API, will be generated using a CSPRNG, stored in base64 format
	jobsLock sync.RWMutex    // synchronizes access to the jobs map
	jobs     map[string]*Job // Index. list of jobs that belong to the user. Index key is the job ID.
}

var state = struct {
	usersIndexLock sync.RWMutex     // synchronizes access to the 'usersIndex' global state
	usersIndex     map[string]*User // maps username to user struct
}{
	sync.RWMutex{},
	map[string]*User{},
}

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

	return u.jobs[jobID]
}

func (u *User) AddJob(job *Job) {
	u.jobsLock.Lock()
	defer u.jobsLock.Unlock()

	u.jobs[job.GetId()] = job
}

func GetIndexedUser(username string) *User {
	state.usersIndexLock.RLock()
	defer state.usersIndexLock.RUnlock()

	return state.usersIndex[username]
}

func AddUser(username, token string) {
	state.usersIndexLock.Lock()
	defer state.usersIndexLock.Unlock()

	state.usersIndex[username] = &User{
		username: username,
		token:    token,
		jobs:     map[string]*Job{},
	}
}

func ClearUsers() {
	state.usersIndexLock.Lock()
	defer state.usersIndexLock.Unlock()

	state.usersIndex = map[string]*User{}
}
