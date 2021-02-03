package internal

import "fmt"

func DisplayJobList(jobs []*JobViewPartial) {

	for _, job := range jobs {
		fmt.Println(*job)
	}
}
