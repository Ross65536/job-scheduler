package view

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

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
	Stdout    string     `json:"stdout,omitempty"`
	Stderr    string     `json:"stderr,omitempty"`
	ExitCode  *int       `json:"exit_code,omitempty"`
	StoppedAt *time.Time `json:"stopped_at,omitempty"`
}

func intToStr(num *int) string {
	if num == nil {
		return "-"
	}

	return strconv.Itoa(*num)
}

func (job *JobViewFull) String() string {
	return fmt.Sprintf("%s, %s, %s -> %s, exit_code: %s\n\nSTDOUT:\n%s\n\nSTDERR:\n%s",
		strings.Join(job.Command, " "), job.Status, job.CreatedAt, job.StoppedAt, intToStr(job.ExitCode), job.Stdout, job.Stderr)
}
