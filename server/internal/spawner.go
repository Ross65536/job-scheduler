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

func readPipe(consumer func([]byte), r io.ReadCloser, ch chan<- bool) {
	buffer := make([]byte, bufSize)

	for {
		n, err := r.Read(buffer)
		if n != 0 {
			consumer(buffer[:n])
		}

		if err == nil {
			continue
		}

		r.Close()
		if err == io.EOF {
			ch <- true
		} else {
			log.Printf("Something went wrong reading from pipe %s", err)
			ch <- false
		}

		return
	}
}

func SpawnJob(user *User, command []string, c chan<- *Job) {
	cmd := exec.Command(command[0], command[1:]...)

	stdout, stdoutErr := cmd.StdoutPipe()
	if stdoutErr != nil {
		log.Printf("Failed to create stderr pipe %s, %s", command, stdoutErr)
		c <- nil
		return
	}

	stderr, stderrErr := cmd.StderrPipe()
	if stderrErr != nil {
		log.Printf("Failed to create stderr pipe %s, %s", command, stderrErr)
		c <- nil
		return
	}

	err := cmd.Start()
	if err != nil {
		log.Printf("Failed to start process %s, %s", command, err)
		c <- nil
		return
	}

	job := user.AddJob(func(id string) *Job { return CreateJob(id, command, cmd.Process) })
	c <- job

	waiter := make(chan bool, 1)
	go readPipe(job.UpdateStdout, stdout, waiter)
	go readPipe(job.UpdateStderr, stderr, waiter)
	<-waiter
	<-waiter

	switch exitErr := cmd.Wait(); exitErr.(type) {
	case nil:
		job.MarkJobAsFinished(0)

	case *exec.ExitError:
		code := exitErr.(*exec.ExitError).ExitCode()
		if code == -1 { // cannot be -1 because it didn't finish, so it can only be from a signal
			job.MarkJobAsStopped()
		} else {
			job.MarkJobAsFinished(code)
		}

	default:
		log.Printf("Something went wrong with process execution or termination %s %s", command, exitErr)
		job.MarkJobAsKilled()
	}
}

func StopJob(job *Job) bool {
	if !job.IsExecuting() {
		return true
	}

	var signal os.Signal = syscall.SIGTERM
	if job.IsStopping() {
		signal = os.Kill
	}

	if err := job.GetProcess().Signal(signal); err != nil {
		log.Printf("Something went wrong sending a signal %s", err)
		return false
	}

	job.MarkJobAsStopping()
	return true
}
