package internal

import (
	"encoding/json"
	"net/http"
)

type APIClient struct {
	HTTPClient *HTTPClient
}

func (api *APIClient) ListJobs() ([]*JobViewPartial, error) {
	resp, err := api.HTTPClient.MakeJSONRequest(http.MethodGet, nil, "api", "jobs")
	if err != nil {
		return nil, err
	}

	jobs := []*JobViewPartial{}
	err = json.Unmarshal(resp, &jobs)

	return jobs, err
}

func (api *APIClient) ShowJob(id string) (*JobViewFull, error) {
	resp, err := api.HTTPClient.MakeJSONRequest(http.MethodGet, nil, "api", "jobs", id)
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

	resp, err := api.HTTPClient.MakeJSONRequest(http.MethodPost, requestJson, "api", "jobs")
	if err != nil {
		return nil, err
	}

	respJob := JobViewPartial{}
	err = json.Unmarshal(resp, &respJob)

	return &respJob, err
}

func (api *APIClient) StopJob(id string) error {
	_, err := api.HTTPClient.MakeJSONRequest(http.MethodDelete, nil, "api", "jobs", id)
	return err
}
