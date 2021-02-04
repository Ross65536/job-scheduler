package internal

// TODO: Move this file into a common module/open package to re-use it between the client and server
// This file is the same file as 'server/internal/job_view.go

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
	Stdout    string     `json:"stdout"`
	Stderr    string     `json:"stderr"`
	ExitCode  *int       `json:"exit_code,omitempty"`
	StoppedAt *time.Time `json:"stopped_at,omitempty"`
}
