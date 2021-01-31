package internal

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"
)

type JobCreateModel struct {
	Command []string
}

func (j *JobCreateModel) IsValid() bool {
	if j.Command == nil || len(j.Command) < 1 {
		return false
	}

	if program := j.Command[0]; len(strings.TrimSpace(program)) == 0 {
		return false
	}

	// TODO: add more sanity checks like invalid characters, etc

	return true
}

func ParseJobCreation(r io.Reader) *JobCreateModel {
	reqBody, err := ioutil.ReadAll(r)
	if err != nil {
		return nil
	}

	var createJob JobCreateModel
	if err := json.Unmarshal(reqBody, &createJob); err != nil {
		return nil
	}

	return &createJob
}
