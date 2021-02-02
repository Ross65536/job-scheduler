package internal

import (
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
)

const (
	bufSize = 256
)

func readPipe(consumer func([]byte), r io.Reader, ch chan<- error) {
	buffer := make([]byte, bufSize)

	for {
		n, err := r.Read(buffer)
		if n != 0 {
			consumer(buffer[:n])
		}

		if err == nil {
			continue
		}

		if err == io.EOF {
			ch <- nil
		} else {
			ch <- err
		}

		return
	}
}

type SpawnJobResult struct {
	Job *Job
	Err error
}

func waitJob(job *Job, cmd *exec.Cmd, stdout, stderr io.Reader) {
	waiter := make(chan error, 1)
	go readPipe(job.UpdateStdout, stdout, waiter)
	go readPipe(job.UpdateStderr, stderr, waiter)

	for i := 0; i < 2; i++ {
		select {
		case pipeErr := <-waiter:
			if pipeErr != nil {
				log.Printf("Something went wrong reading from pipe %s", pipeErr)
			}
		}
	}

	switch waitErr := cmd.Wait(); exitErr := waitErr.(type) {
	case nil:
		job.MarkAsFinished(0)

	case *exec.ExitError:
		code := exitErr.ExitCode()
		if code == -1 { // cannot be -1 because it didn't finish, so it can only be from a signal
			job.MarkAsStopped()
		} else {
			job.MarkAsFinished(code)
		}

	default:
		log.Printf("Something went wrong with process execution or termination %s %s", cmd.Path, exitErr)
		job.MarkAsKilled()
	}
}

func SpawnJob(user *User, command []string) (*Job, error) {
	cmd := exec.Command(command[0], command[1:]...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if startErr := cmd.Start(); startErr != nil {
		return nil, err
	}

	job := CreateJob(command, cmd.Process)
	go waitJob(job, cmd, stdout, stderr)

	return job, nil
}

func StopJob(job *Job) error {
	if !job.IsExecuting() {
		return nil
	}

	var signal os.Signal = syscall.SIGTERM
	if job.IsStopping() {
		signal = os.Kill
	}

	if err := job.GetProcess().Signal(signal); err != nil {
		return err
	}

	job.MarkAsStopping()
	return nil
}
