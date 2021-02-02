package internal

import (
	"os"
	"sync"
	"time"
)

type JobStatus string

const (
	jobRunning  JobStatus = "RUNNING"
	jobFinished JobStatus = "FINISHED"
	jobStopped  JobStatus = "STOPPED"
	jobStopping JobStatus = "STOPPING"
	jobKilled   JobStatus = "KILLED"
)

type Job struct {
	id        string // ID exposed to the client (UUID), NOT EMPTY, UNIQUE
	proc      *os.Process
	command   []string   // command name + argv, NOT EMPTY
	status    JobStatus  // status of job, NOT EMPTY
	stdout    []byte     // process stdout
	stderr    []byte     // process stderr
	exitCode  *int       // process exit code
	createdAt time.Time  // time when job started, NOT EMPTY
	stoppedAt *time.Time // time when job is stopped, killed or has finished
	lock      sync.Mutex
}

func CreateJob(command []string, proc *os.Process) *Job {
	return &Job{
		id:        "",
		proc:      proc,
		command:   command,
		status:    jobRunning,
		createdAt: time.Now(),
	}
}

func (j *Job) GetProcess() *os.Process {
	j.lock.Lock()
	proc := j.proc
	j.lock.Unlock()

	return proc
}

func (j *Job) SetID(id string) {
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

func (j *Job) IsExecuting() bool {
	j.lock.Lock()

	running := j.status == jobRunning || j.status == jobStopping

	j.lock.Unlock()

	return running
}

func (j *Job) IsStopping() bool {
	j.lock.Lock()

	running := j.status == jobStopping

	j.lock.Unlock()

	return running
}

func (j *Job) StoppingJob() {
	j.lock.Lock()

	if j.status == jobRunning {
		j.status = jobStopping
	}

	j.lock.Unlock()
}

func (j *Job) endJob(status JobStatus) {
	j.status = status
	t := time.Now()
	j.stoppedAt = &t
}

func (j *Job) StopJob(normalStop bool) {
	j.lock.Lock()

	if normalStop && j.status == jobStopping {
		j.endJob(jobStopped)
	} else {
		j.endJob(jobKilled)
	}

	j.lock.Unlock()
}

func (j *Job) FinishJob(exitCode int) {
	j.lock.Lock()

	j.endJob(jobFinished)
	j.exitCode = &exitCode

	j.lock.Unlock()
}

func (j *Job) AsMap() map[string]interface{} {
	j.lock.Lock()

	m := map[string]interface{}{
		"id":     j.id,
		"status": j.status,
		// TODO: remove/escape/replace stdout/stderr characters not corresponding to a utf8 rune
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
