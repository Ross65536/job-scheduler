package internal

import (
	"errors"
	"os"
)

func parseFlags([]string) (*APIClient, []string, error) {
	// TODO parse os.Args for flags and filter them out of the command proper

	httpClient, err := NewHTTPClient("user2", "oAtCvE6Xcu07f2PmjoOjq8kv6X2XTgh3s37suKzKHLo=", "http://localhost:10000")
	if err != nil {
		return nil, nil, err
	}

	api := APIClient{httpClient}

	return &api, os.Args, nil
}

func dispatchCommand(api *APIClient, command []string) error {
	if len(command) < 2 {
		return errors.New("Invalid usage, must specify command")
	}

	switch task := command[1]; task {
	case "list":
		if len(command) != 2 {
			return errors.New("Too many args")
		}

		jobs, err := api.ListJobs()
		if err != nil {
			return err
		}

		DisplayJobList(jobs)

		return nil
	default:
		return errors.New("Unknown command " + task)
	}
}

func Start(args []string) error {
	api, command, err := parseFlags(args)
	if err != nil {
		return err
	}

	return dispatchCommand(api, command)
}
