package internal

import (
	"sync"
	"time"
)

type JobStatus string

const (
	running JobStatus = "RUNNING"
	stopped JobStatus = "STOPPED"
	killed  JobStatus = "KILLED"
)

type Job struct {
	id        string     // ID exposed to the client (UUID), NOT EMPTY, UNIQUE
	pid       int        // Unix process ID
	command   []string   // command name + argv, NOT EMPTY
	status    JobStatus  // status of job, one of `RUNNING`, `KILLED` or `FINISHED`, NOT EMPTY
	stdout    string     // process stdout
	stderr    string     // process stderr
	exitCode  *int       // process exit code
	createdAt time.Time  // time when job started, NOT EMPTY
	stoppedAt *time.Time // time when job is killed or has finished
	lock      sync.Mutex
}

func CreateJob(id string, command []string, pid int) *Job {
	return &Job{
		id,
		pid,
		command,
		running,
		"",
		"",
		nil,
		time.Now(),
		nil,
		sync.Mutex{},
	}
}

func (j *Job) AsMap() map[string]interface{} {
	j.lock.Lock()

	m := map[string]interface{}{
		"id":         j.id,
		"command":    j.command,
		"status":     j.status,
		"stdout":     j.stdout,
		"stderr":     j.stderr,
		"exit_code":  j.exitCode,
		"stopped_at": j.stoppedAt,
		"created_at": j.createdAt,
	}

	j.lock.Unlock()

	return m
}
