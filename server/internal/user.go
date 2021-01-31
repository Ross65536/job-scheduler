package internal

import (
	"crypto/subtle"
	"time"
)

type User struct {
	token string         // the API token given to the user to access the API, will be generated using a CSPRNG, stored in hex or base64 format
	jobs  map[string]Job // Index. list of jobs that belong to the user. Index key is the job ID.
	// Username string, // not necessary, already stored in the index
	// Password string, // not used, would be stored as hash using BCrypt
}

func (u *User) GetAllJobs() []Job {
	jobsList := make([]Job, 0, len(u.jobs))

	for _, value := range u.jobs {
		jobsList = append(jobsList, value)
	}

	return jobsList
}

var usersIndex map[string]User // maps username to user struct

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
			jobs: map[string]Job{
				"1": Job{
					"1",
					-1,
					[]string{"ls"},
					Running,
					"",
					"",
					0,
					time.Time{},
					time.Time{},
				},
			},
		},
	}
}
