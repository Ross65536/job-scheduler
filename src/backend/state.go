package backend

import "sync"

type State struct {
	usersIndexLock sync.RWMutex     // synchronizes access to the 'usersIndex' global state
	usersIndex     map[string]*User // maps username to user struct
}

func NewState() *State {
	return &State{
		usersIndexLock: sync.RWMutex{},
		usersIndex:     map[string]*User{},
	}
}

func (s *State) GetIndexedUser(username string) *User {
	s.usersIndexLock.RLock()
	defer s.usersIndexLock.RUnlock()

	return s.usersIndex[username]
}

func (s *State) AddUser(username, token string) {
	s.usersIndexLock.Lock()
	defer s.usersIndexLock.Unlock()

	s.usersIndex[username] = &User{
		username: username,
		token:    token,
		jobs:     map[string]*Job{},
	}
}

func (s *State) ClearUsers() {
	s.usersIndexLock.Lock()
	defer s.usersIndexLock.Unlock()

	s.usersIndex = map[string]*User{}
}
