package backend

import (
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/ros-k/job-manager/src/core"
)

type JobStatus string

const (
	JobRunning  JobStatus = "RUNNING"
	JobFinished JobStatus = "FINISHED"
	JobStopped  JobStatus = "STOPPED"
	JobStopping JobStatus = "STOPPING"
	JobKilled   JobStatus = "KILLED"
)

type Job struct {
	lock      sync.RWMutex // synchronizes access to all of the fields of the struct
	id        string       // ID exposed to the client (UUID), NOT EMPTY, UNIQUE
	proc      *os.Process
	command   []string  // command name + argv, NOT EMPTY
	status    JobStatus // status of job, NOT EMPTY
	stdout    []byte    // process stdout
	stderr    []byte    // process stderr
	exitCode  *int      // process exit code
	createdAt time.Time // time when job started, NOT EMPTY
	stoppedAt time.Time // time when job is stopped, killed or has finished
}

func CreateJob(command []string, proc *os.Process) *Job {
	return &Job{
		id:        uuid.NewString(),
		proc:      proc,
		command:   command,
		status:    JobRunning,
		createdAt: time.Now(),
	}
}

func (j *Job) GetID() string {
	j.lock.RLock()
	defer j.lock.RUnlock()

	return j.id
}

func (j *Job) UpdateStdout(bytes []byte) {
	j.lock.Lock()
	defer j.lock.Unlock()

	j.stdout = append(j.stdout, bytes...)
}

func (j *Job) UpdateStderr(bytes []byte) {
	j.lock.Lock()
	defer j.lock.Unlock()

	j.stderr = append(j.stderr, bytes...)
}

func (j *Job) isExecutingLocked() bool {
	return j.status == JobRunning || j.status == JobStopping
}

func (j *Job) endJobLocked(status JobStatus) {
	j.status = status
	j.stoppedAt = time.Now()
}

func (j *Job) MarkAsStopped() {
	j.lock.Lock()
	defer j.lock.Unlock()

	if j.status == JobStopping {
		j.endJobLocked(JobStopped)
	} else {
		j.endJobLocked(JobKilled)
	}
}

func (j *Job) MarkAsKilled() {
	j.lock.Lock()
	defer j.lock.Unlock()

	j.endJobLocked(JobKilled)
}

func (j *Job) MarkAsFinished(exitCode int) {
	j.lock.Lock()
	defer j.lock.Unlock()

	j.endJobLocked(JobFinished)
	j.exitCode = &exitCode
}

func (j *Job) StopJob() error {
	j.lock.Lock()
	defer j.lock.Unlock()

	if !j.isExecutingLocked() {
		return nil
	}

	var signal os.Signal = syscall.SIGTERM
	if j.status == JobStopping {
		signal = os.Kill
	}

	if err := j.proc.Signal(signal); err != nil {
		return err
	}

	if j.status == JobRunning {
		j.status = JobStopping
	}

	return nil
}

func (j *Job) AsView() core.JobViewFull {
	j.lock.RLock()
	defer j.lock.RUnlock()

	commandDup := make([]string, len(j.command))
	copy(commandDup, j.command)

	m := core.JobViewFull{
		JobViewPartial: core.JobViewPartial{
			JobViewCommand: core.JobViewCommand{
				Command: commandDup,
			},
			ID:        j.id,
			Status:    string(j.status),
			CreatedAt: j.createdAt,
		},
		Stdout:   string(j.stdout),
		Stderr:   string(j.stderr),
		ExitCode: j.exitCode,
	}

	if !j.stoppedAt.IsZero() {
		copy := j.stoppedAt
		m.StoppedAt = &copy
	}

	return m
}
