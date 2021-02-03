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

func (api *APIClient) StartJob(command []string) (*JobViewFull, error) {
	job := JobViewCommand{Command: command}
	requestJson, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}

	resp, err := api.HTTPClient.MakeJSONRequest(http.MethodPost, requestJson, "api", "jobs")
	if err != nil {
		return nil, err
	}

	respJob := &JobViewFull{}
	err = json.Unmarshal(resp, &respJob)

	return respJob, err
}
