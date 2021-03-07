package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Ross65536/job-scheduler/src/core/view"
)

type APIClient struct {
	HTTPClient *HTTPClient
}

type ErrorType struct {
	Status  int
	Message string
}

func buildResponseError(code int, body []byte) error {
	// best case JSON parsing, return raw HTTP body otherwise
	parsed := ErrorType{}
	err := json.Unmarshal(body, &parsed)

	if err == nil && parsed.Message != "" {
		return fmt.Errorf("an error occurred (HTTP %d): %s", code, parsed.Message)
	}

	return fmt.Errorf("an error occurred (HTTP %d): %s", code, string(body))
}

func (api *APIClient) ListJobs() ([]*view.JobViewPartial, error) {
	status, resp, err := api.HTTPClient.Get("api", "jobs")
	if err != nil {
		return nil, err
	}

	if http.StatusOK != status {
		return nil, buildResponseError(status, resp)
	}

	jobs := []*view.JobViewPartial{}
	err = json.Unmarshal(resp, &jobs)

	return jobs, err
}

func (api *APIClient) ShowJob(id string) (*view.JobViewFull, error) {
	status, resp, err := api.HTTPClient.Get("api", "jobs", url.PathEscape(id))
	if err != nil {
		return nil, err
	}

	if http.StatusOK != status {
		return nil, buildResponseError(status, resp)
	}

	job := view.JobViewFull{}
	err = json.Unmarshal(resp, &job)

	return &job, err
}

func (api *APIClient) StartJob(command []string) (*view.JobViewPartial, error) {
	job := view.JobViewCommand{Command: command}
	requestJson, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}

	status, resp, err := api.HTTPClient.Post(requestJson, "api", "jobs")
	if err != nil {
		return nil, err
	}

	if http.StatusCreated != status {
		return nil, buildResponseError(status, resp)
	}

	respJob := view.JobViewPartial{}
	err = json.Unmarshal(resp, &respJob)

	return &respJob, err
}

func (api *APIClient) StopJob(id string) error {
	status, resp, err := api.HTTPClient.Delete("api", "jobs", url.PathEscape(id))
	if err != nil {
		return err
	}

	if http.StatusNoContent != status {
		return buildResponseError(status, resp)
	}

	return nil
}
