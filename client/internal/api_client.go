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
