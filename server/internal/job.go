package internal

import (
	"os"
	"sync"
	"time"
)

type JobStatus string

const (
	JobRunning  JobStatus = "RUNNING"
	JobFinished JobStatus = "FINISHED"
	JobStopped  JobStatus = "STOPPED"
	JobKilled   JobStatus = "KILLED"
)

type Job struct {
	id        string      // ID exposed to the client (UUID), NOT EMPTY, UNIQUE
	proc      *os.Process // Unix process ID
	command   []string    // command name + argv, NOT EMPTY
	status    JobStatus   // status of job, one of `RUNNING`, `KILLED` or `FINISHED`, NOT EMPTY
	stdout    []byte      // process stdout
	stderr    []byte      // process stderr
	exitCode  *int        // process exit code
	createdAt time.Time   // time when job started, NOT EMPTY
	stoppedAt *time.Time  // time when job is killed or has finished
	lock      sync.Mutex
}

func CreateJob(command []string, proc *os.Process) *Job {
	return &Job{
		"",
		proc,
		command,
		JobRunning,
		[]byte{},
		[]byte{},
		nil,
		time.Now(),
		nil,
		sync.Mutex{},
	}
}

func (j *Job) SetId(id string) {
	j.lock.Lock()
	j.id = id
	j.lock.Unlock()
}

func (j *Job) UpdateStdout(bytes []byte) {
	j.lock.Lock()

	j.stdout = append(j.stdout, bytes...)

	j.lock.Unlock()
}

func (j *Job) UpdateStderr(bytes []byte) {
	j.lock.Lock()

	j.stderr = append(j.stderr, bytes...)

	j.lock.Unlock()
}

func (j *Job) StopJob(status JobStatus) {
	j.lock.Lock()

	j.status = status

	j.lock.Unlock()
}

func (j *Job) FinishJob(status JobStatus, exitCode int) {
	j.lock.Lock()

	j.status = status
	j.exitCode = &exitCode

	j.lock.Unlock()
}

func (j *Job) AsMap() map[string]interface{} {
	j.lock.Lock()

	m := map[string]interface{}{
		"id":         j.id,
		"status":     j.status,
		"stdout":     string(j.stdout),
		"stderr":     string(j.stderr),
		"created_at": j.createdAt,
	}

	commandDup := make([]string, len(j.command))
	copy(commandDup, j.command)
	m["command"] = commandDup

	if j.exitCode != nil {
		m["exit_code"] = *j.exitCode
	}

	if j.stoppedAt != nil {
		m["stopped_at"] = *j.stoppedAt
	}

	j.lock.Unlock()

	return m
}
