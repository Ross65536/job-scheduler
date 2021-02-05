package main

import (
	"log"
	"os"

	"github.com/ros-k/job-manager/client/internal"
)

func main() {
	args := os.Args
	if err := internal.Start(); err != nil {
		log.Fatalf("Failed to start command: %v, because: %v", args[1:], err)
	}
}
