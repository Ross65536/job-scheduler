package internal

import (
	"errors"
	"fmt"
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

func listTask(api *APIClient, commandRest []string) error {
	if len(commandRest) != 0 {
		return errors.New("Too many args")
	}

	jobs, err := api.ListJobs()
	if err != nil {
		return err
	}

	DisplayJobList(jobs)
	return nil
}

func startTask(api *APIClient, commandRest []string) error {
	if len(commandRest) == 0 {
		return errors.New("Must specify job to start")
	}

	job, err := api.StartJob(commandRest)
	if err != nil {
		return err
	}

	fmt.Printf("ID: %s\n", job.ID)
	return nil
}

func dispatchCommand(api *APIClient, command []string) error {
	if len(command) < 2 {
		return errors.New("Invalid usage, must specify command")
	}

	task := command[1]
	commandRest := command[2:]
	switch task {
	case "list":
		return listTask(api, commandRest)
	case "start":
		return startTask(api, commandRest)
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
