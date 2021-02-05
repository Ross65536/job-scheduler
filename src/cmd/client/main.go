package main

import (
	"log"
	"os"

	"github.com/ros-k/job-manager/src/client"
)

func main() {
	args := os.Args
	if err := client.Start(os.Stdout, args); err != nil {
		log.Fatalf("Failed to run because: %v", err)
	}
}
