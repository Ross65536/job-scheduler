package internal

import (
	"crypto/subtle"
	"log"
	"sync"

	"github.com/google/uuid"
)

// User's username is already stored in the 'usersIndex' keys
type User struct {
	token    string          // the API token given to the user to access the API, will be generated using a CSPRNG, stored in hex or base64 format
	jobs     map[string]*Job // Index. list of jobs that belong to the user. Index key is the job ID.
	jobsLock sync.Mutex
}

// there is no need to synchronize access with a mutex to this index,
// since it is never modified by the backend, only pre-initialized
var usersIndex map[string]User // maps username to user struct

func (u *User) GetAllJobs() []*Job {
	u.jobsLock.Lock()
	jobsList := make([]*Job, 0, len(u.jobs))

	for _, value := range u.jobs {
		jobsList = append(jobsList, value)
	}
	u.jobsLock.Unlock()

	return jobsList
}

func (u *User) GetJob(jobID string) *Job {
	u.jobsLock.Lock()
	job, ok := u.jobs[jobID]
	u.jobsLock.Unlock()

	if ok {
		return job
	}

	return nil
}

func generateNewID(keys map[string]*Job) string {
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

// will modify job argument and return it's ID
func (u *User) AddJob(job *Job) {
	u.jobsLock.Lock()
	id := generateNewID(u.jobs)

	job.SetID(id)
	u.jobs[id] = job
	u.jobsLock.Unlock()
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
			token: "XlG15tRINdWTAm5oZ/mhikbEiwf75w0LJUVek0ROhY4=",
			jobs:  map[string]*Job{},
		},
		"user2": User{
			token: "oAtCvE6Xcu07f2PmjoOjq8kv6X2XTgh3s37suKzKHLo=",
			jobs:  map[string]*Job{},
		},
	}
}
