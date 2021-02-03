package internal

import (
	"errors"
	"fmt"
	"strings"
)

const (
	connectionFlag = "-c="
)

func parseFlags(args []string) (*APIClient, []string, error) {
	command := []string{}
	url := ""

	for _, arg := range args {
		if strings.HasPrefix(arg, connectionFlag) {
			url = strings.TrimPrefix(arg, connectionFlag)
		} else {
			command = append(command, arg)
		}
	}

	if url == "" {
		return nil, nil, fmt.Errorf("Must specify connection flag '%s'", connectionFlag)
	}

	// TODO parse os.Args for flags and filter them out of the command proper
	httpClient, err := NewHTTPClient(url)
	if err != nil {
		return nil, nil, err
	}
	api := APIClient{httpClient}

	return &api, command, nil
}

func joinString(command []string) string {
	return strings.Join(command, " ")
}

func intToStr(num *int) string {
	if num == nil {
		return "-"
	}

	return fmt.Sprintf("%d", *num)
}

func displayJobList(jobs []*JobViewPartial) {

	fmt.Println("ID | STATUS | COMMAND | CREATED_AT")
	for _, job := range jobs {
		fmt.Printf("%s | %s | %s | %s \n", job.ID, job.Status, joinString(job.Command), job.CreatedAt)
	}

	fmt.Printf("--- %d jobs --- \n", len(jobs))
}

func listTask(api *APIClient, commandRest []string) error {
	if len(commandRest) != 0 {
		return errors.New("Too many args")
	}

	jobs, err := api.ListJobs()
	if err != nil {
		return err
	}

	displayJobList(jobs)
	return nil
}

func showTask(api *APIClient, commandRest []string) error {
	if len(commandRest) != 1 {
		return errors.New("Must specify ID")
	}

	id := commandRest[0]

	job, err := api.ShowJob(id)
	if err != nil {
		return err
	}

	fmt.Printf("%s, %s, %s -> %s, exit_code: %s\n\nSTDOUT:\n%s\n\nSTDERR:\n%s\n",
		joinString(job.Command), job.Status, job.CreatedAt, job.StoppedAt, intToStr(job.ExitCode), job.Stdout, job.Stderr)

	return nil
}

func stopTask(api *APIClient, commandRest []string) error {
	if len(commandRest) != 1 {
		return errors.New("Must specify ID")
	}

	id := commandRest[0]

	err := api.StopJob(id)
	if err != nil {
		return err
	}

	fmt.Println("Stopping")
	return nil
}

func startTask(api *APIClient, commandRest []string) error {
	if len(commandRest) != 1 {
		return errors.New("Must specify job to start")
	}

	jobArgs := strings.Fields(commandRest[0])

	job, err := api.StartJob(jobArgs)
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
	case "show":
		return showTask(api, commandRest)
	case "stop":
		return stopTask(api, commandRest)
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
