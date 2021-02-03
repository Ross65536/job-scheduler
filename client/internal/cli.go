package internal

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	connectionFlag = "-c="
)

func parseFlags(args []string) (*APIClient, []string, error) {
	filteredArgs := []string{}
	url := ""

	for _, arg := range args {
		if strings.HasPrefix(arg, connectionFlag) {
			url = strings.TrimPrefix(arg, connectionFlag)
		} else {
			filteredArgs = append(filteredArgs, arg)
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

	return &api, filteredArgs, nil
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

	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt.Before(jobs[j].CreatedAt)
	})

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

func printHelp() {
	fmt.Println(`Format: client <command> [-c=<connection>] [id/command]
	
	command: list | show | stop | start | help

	Examples:
	- client help
	- client -c=http://user:pass@localhost:80 list
	- client -c=http://user:pass@localhost:80 show d99e3759-bcc8-4573-a267-88709761c67e
	- client -c=http://user:pass@localhost:80 stop d99e3759-bcc8-4573-a267-88709761c67e
	- client -c=http://user:pass@localhost:80 start "ls -l /"
	`)
}

func Start(args []string) error {
	if len(args) < 2 {
		return errors.New("Invalid usage, must specify command")
	}

	task := args[1]
	if task == "help" {
		printHelp()
		return nil
	}

	api, filteredArgs, err := parseFlags(args)
	if err != nil {
		return err
	}

	if len(filteredArgs) < 2 {
		return errors.New("Must specify command")
	}

	argsRest := filteredArgs[2:]
	switch task {
	case "list":
		return listTask(api, argsRest)
	case "start":
		return startTask(api, argsRest)
	case "show":
		return showTask(api, argsRest)
	case "stop":
		return stopTask(api, argsRest)
	default:
		return errors.New("Unknown command " + task)
	}
}
