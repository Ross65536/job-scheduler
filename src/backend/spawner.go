package backend

import (
	"io"
	"log"
	"os"
	"os/exec"
)

var bufSize = os.Getpagesize()

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

func waitJob(job *Job, cmd *exec.Cmd, stdout, stderr io.Reader) {
	waiter := make(chan error, 2)
	go readPipe(job.UpdateStdout, stdout, waiter)
	go readPipe(job.UpdateStderr, stderr, waiter)

	for i := 0; i < cap(waiter); i++ {
		err := <-waiter
		if err != nil {
			log.Printf("Something went wrong reading from pipe %s", err)
		}
	}

	switch err := cmd.Wait(); exitErr := err.(type) {
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
		log.Printf("Something went wrong with process execution or termination %s %s", cmd.Path, err)
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

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	job := CreateJob(command, cmd.Process)
	go waitJob(job, cmd, stdout, stderr)

	return job, nil
}
