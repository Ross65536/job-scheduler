package main

import (
	"log"
	"os"

	"github.com/ros-k/job-manager/client/internal"
)

func main() {
	args := os.Args
	if err := internal.Start(args); err != nil {
		log.Fatalf("Failed to run because: %v", err)
	}
}
