package internal

import (
	"encoding/json"
	"net/url"
)

type APIClient struct {
	HTTPClient *HTTPClient
}

func (api *APIClient) ListJobs() ([]*JobViewPartial, error) {
	resp, err := api.HTTPClient.Get("api", "jobs")
	if err != nil {
		return nil, err
	}

	jobs := []*JobViewPartial{}
	err = json.Unmarshal(resp, &jobs)

	return jobs, err
}

func (api *APIClient) ShowJob(id string) (*JobViewFull, error) {
	resp, err := api.HTTPClient.Get("api", "jobs", url.PathEscape(id))
	if err != nil {
		return nil, err
	}

	job := JobViewFull{}
	err = json.Unmarshal(resp, &job)

	return &job, err
}

func (api *APIClient) StartJob(command []string) (*JobViewPartial, error) {
	job := JobViewCommand{Command: command}
	requestJson, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}

	resp, err := api.HTTPClient.Post(requestJson, "api", "jobs")
	if err != nil {
		return nil, err
	}

	respJob := JobViewPartial{}
	err = json.Unmarshal(resp, &respJob)

	return &respJob, err
}

func (api *APIClient) StopJob(id string) error {
	_, err := api.HTTPClient.Delete("api", "jobs", url.PathEscape(id))
	return err
}
