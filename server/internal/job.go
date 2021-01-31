package internal

import "time"

type JobStatus string

const (
	Running JobStatus = "RUNNING"
	Stopped JobStatus = "STOPPED"
	Killed  JobStatus = "KILLED"
)

type Job struct {
	ID        string    `json:"id"`         // ID exposed to the client (UUID), NOT EMPTY, UNIQUE
	Pid       int       `json:"pid"`        // Unix process ID
	Command   []string  `json:"command"`    // command name + argv, NOT EMPTY
	Status    JobStatus `json:"status"`     // status of job, one of `RUNNING`, `KILLED` or `FINISHED`, NOT EMPTY
	Stdout    string    `json:"stdout"`     // process stdout
	Stderr    string    `json:"stderr"`     // process stderr
	ExitCode  int       `json:"exit_code"`  // process exit code
	CreatedAt time.Time `json:"created_at"` // time when job started, NOT EMPTY
	StoppedAt time.Time `json:"stopped_at"` // time when job is killed or has finished
}
