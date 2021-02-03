package main

import (
	"log"
	"os"

	"github.com/ros-k/job-manager/server/internal"
)

func main() {
	state := internal.NewState()
	// TODO: place this into a config file or equivalent
	state.AddUser("user1", "XlG15tRINdWTAm5oZ/mhikbEiwf75w0LJUVek0ROhY4=")
	state.AddUser("user2", "oAtCvE6Xcu07f2PmjoOjq8kv6X2XTgh3s37suKzKHLo=")

	server, err := internal.NewServer(state)
	if err != nil {
		log.Fatalf("Failed to start server %s", err)
	}

	if err := server.Start(); err != nil {
		log.Printf("An error occurred, the server stopped %s", err)
		os.Exit(1)
	}
}
