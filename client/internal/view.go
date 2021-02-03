package internal

import (
	"fmt"
	"strings"
)

func DisplayJobList(jobs []*JobViewPartial) {

	fmt.Println("ID | STATUS | COMMAND | CREATED_AT")
	for _, job := range jobs {
		command := strings.Join(job.Command, " ")
		fmt.Printf("%s | %s | %s | %s \n", job.ID, job.Status, command, job.CreatedAt)
	}

	fmt.Printf("--- %d jobs --- \n", len(jobs))
}
