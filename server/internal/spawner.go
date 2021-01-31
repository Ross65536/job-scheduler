package internal

import (
	"log"
	"os/exec"
)

func SpawnJob(command []string, c chan<- *Job) {
	cmd := exec.Command(command[0], command[1:]...)
	err := cmd.Start()
	if err != nil {
		log.Printf("Failed to start process %s, %s", command, err)
		c <- nil
		return
	}

	_, stdoutErr := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Failed to create stderr pipe %s, %s", command, stdoutErr)
		c <- nil
		return
	}

	_, stderrErr := cmd.StderrPipe()
	if err != nil {
		log.Printf("Failed to create stderr pipe %s, %s", command, stderrErr)
		c <- nil
		return
	}

	job := CreateJob(command, cmd.Process)
	c <- job

	switch exitErr := cmd.Wait(); exitErr.(type) {
	case nil:
		job.FinishJob(JobFinished, 0)

	case *exec.ExitError:
		code := exitErr.(*exec.ExitError).ExitCode()
		if code == -1 { // cannot be -1 because it didn't finish, so it can only be from a signal
			job.StopJob(JobStopped)
		} else {
			job.FinishJob(JobFinished, code)
		}

	default:
		log.Printf("Something went wrong with process execution or termination %s %s", command, exitErr)
		job.StopJob(JobKilled)
	}

}
