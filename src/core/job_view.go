package core

import "time"

type JobViewCommand struct {
	Command []string `json:"command"`
}

type JobViewPartial struct {
	JobViewCommand
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type JobViewFull struct {
	JobViewPartial
	Stdout    string     `json:"stdout,omitempty"`
	Stderr    string     `json:"stderr,omitempty"`
	ExitCode  *int       `json:"exit_code,omitempty"`
	StoppedAt *time.Time `json:"stopped_at,omitempty"`
}
