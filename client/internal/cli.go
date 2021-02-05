package internal

import (
	"errors"
	"flag"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	connectionFlag = "-c="
	defaultURL     = "http://user2:oAtCvE6Xcu07f2PmjoOjq8kv6X2XTgh3s37suKzKHLo=@localhost:10000"
)

func parseArgs(args []string) (*APIClient, []string, error) {
	flags := flag.NewFlagSet("flags-1", flag.ContinueOnError)

	url := flags.String("c", defaultURL, "the URI to the backend with credentials basic encoded")

	flags.Parse(args)

	filteredArgs := flags.Args()
	if len(filteredArgs) < 1 {
		return nil, nil, errors.New("Invalid usage, must specify command")
	}

	if filteredArgs[0] == "help" {
		printHelp()
		return nil, nil, nil
	}

	httpClient, err := NewHTTPClient(*url)
	if err != nil {
		return nil, nil, err
	}
	api := APIClient{httpClient}

	return &api, filteredArgs, nil
}

func intToStr(num *int) string {
	if num == nil {
		return "-"
	}

	return strconv.Itoa(*num)
}

func displayJobList(jobs []*JobViewPartial) {

	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt.Before(jobs[j].CreatedAt)
	})

	fmt.Println("ID | STATUS | COMMAND | CREATED_AT")
	for _, job := range jobs {
		fmt.Printf("%s | %s | %s | %s \n", job.ID, job.Status, strings.Join(job.Command, " "), job.CreatedAt)
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
		strings.Join(job.Command, " "), job.Status, job.CreatedAt, job.StoppedAt, intToStr(job.ExitCode), job.Stdout, job.Stderr)

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
	if len(commandRest) < 1 {
		return errors.New("Must specify job to start")
	}

	job, err := api.StartJob(commandRest)
	if err != nil {
		return err
	}

	fmt.Printf("ID: %s\n", job.ID)
	return nil
}

func printHelp() {
	fmt.Println(`Format: client [flags] <command> [id/job]
	
	command: list | show | stop | start | help
	`)
	flag.PrintDefaults()
	fmt.Println(`
	Examples:
	- client help
	- client -c=http://user:pass@localhost:80 list
	- client -c=http://user:pass@localhost:80 show d99e3759-bcc8-4573-a267-88709761c67e
	- client -c=http://user:pass@localhost:80 stop d99e3759-bcc8-4573-a267-88709761c67e
	- client -c=http://user:pass@localhost:80 start "ls -l /"
	`)
}

func Start(args []string) error {
	api, filteredArgs, err := parseArgs(args)

	if api == nil && filteredArgs == nil && err == nil {
		return nil
	}
	if err != nil {
		return err
	}

	argsRest := filteredArgs[1:]
	switch task := filteredArgs[0]; task {
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
