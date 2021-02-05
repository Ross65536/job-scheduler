package client

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

const (
	connectionFlag = "-c="
	defaultURL     = "http://user2:oAtCvE6Xcu07f2PmjoOjq8kv6X2XTgh3s37suKzKHLo=@localhost:10000"
)

func parseArgs(out io.Writer, args []string) (*APIClient, []string, error) {
	if len(args) < 2 {
		return nil, nil, errors.New("invalid usage, must specify command")
	}

	flags := flag.NewFlagSet("flags-1", flag.ContinueOnError)

	url := flags.String("c", defaultURL, "the URI to the backend with credentials basic encoded")

	flags.Parse(args[1:])

	filteredArgs := flags.Args()

	if filteredArgs[0] == "help" {
		printHelp(out, flags)
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

func displayJobList(out io.Writer, jobs []*JobViewPartial) {

	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt.Before(jobs[j].CreatedAt)
	})

	w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATUS\tCOMMAND\tCREATED_AT\t")
	for _, job := range jobs {
		str := fmt.Sprintf("%s\t%s\t%s\t%s\t", job.ID, job.Status, strings.Join(job.Command, " "), job.CreatedAt)
		fmt.Fprintln(w, str)
	}
	w.Flush()

	fmt.Fprintf(out, "--- %d jobs --- \n", len(jobs))
}

func listTask(out io.Writer, api *APIClient, commandRest []string) error {
	if len(commandRest) != 0 {
		return errors.New("too many args")
	}

	jobs, err := api.ListJobs()
	if err != nil {
		return err
	}

	displayJobList(out, jobs)
	return nil
}

func showTask(out io.Writer, api *APIClient, commandRest []string) error {
	if len(commandRest) != 1 {
		return errors.New("must specify ID")
	}

	id := commandRest[0]

	job, err := api.ShowJob(id)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "%s, %s, %s -> %s, exit_code: %s\n\nSTDOUT:\n%s\n\nSTDERR:\n%s\n",
		strings.Join(job.Command, " "), job.Status, job.CreatedAt, job.StoppedAt, intToStr(job.ExitCode), job.Stdout, job.Stderr)

	return nil
}

func stopTask(out io.Writer, api *APIClient, commandRest []string) error {
	if len(commandRest) != 1 {
		return errors.New("must specify ID")
	}

	id := commandRest[0]

	err := api.StopJob(id)
	if err != nil {
		return err
	}

	fmt.Fprintln(out, "OK")
	return nil
}

func startTask(out io.Writer, api *APIClient, commandRest []string) error {
	if len(commandRest) < 1 {
		return errors.New("must specify job to start")
	}

	job, err := api.StartJob(commandRest)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "ID: %s\n", job.ID)
	return nil
}

func printHelp(out io.Writer, flags *flag.FlagSet) {
	fmt.Fprintf(out, `Format: client [flags] <command> [id/job]
	
	command: list | show | stop | start | help
	`)

	flags.SetOutput(out)
	flags.PrintDefaults()
	fmt.Fprintf(out, `
	Examples:
	- client help
	- client -c=http://user:pass@localhost:80 list
	- client -c=http://user:pass@localhost:80 show d99e3759-bcc8-4573-a267-88709761c67e
	- client -c=http://user:pass@localhost:80 stop d99e3759-bcc8-4573-a267-88709761c67e
	- client -c=http://user:pass@localhost:80 start "ls -l /"
	`)
}

func Start(out io.Writer, args []string) error {
	api, filteredArgs, err := parseArgs(out, args)

	if api == nil && filteredArgs == nil && err == nil {
		return nil
	}
	if err != nil {
		return err
	}

	argsRest := filteredArgs[1:]
	switch task := filteredArgs[0]; task {
	case "list":
		return listTask(out, api, argsRest)
	case "start":
		return startTask(out, api, argsRest)
	case "show":
		return showTask(out, api, argsRest)
	case "stop":
		return stopTask(out, api, argsRest)
	default:
		return errors.New("unknown command " + task)
	}
}
