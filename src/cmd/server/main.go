package main

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/ros-k/job-manager/src/backend"
)

func parsePort() (int, error) {
	port := flag.Int("p", 10000, "port to listen on")

	flag.Parse()

	if *port < 0 || *port > 65535 {
		return 0, errors.New("invalid port value")
	}

	return *port, nil
}

func main() {
	state := backend.NewState()
	// TODO: place this into a config file or equivalent
	state.AddUser("user1", "XlG15tRINdWTAm5oZ/mhikbEiwf75w0LJUVek0ROhY4=")
	state.AddUser("user2", "oAtCvE6Xcu07f2PmjoOjq8kv6X2XTgh3s37suKzKHLo=")

	server, err := backend.NewServer(state)
	if err != nil {
		log.Fatalf("Failed to create server %s", err)
	}

	port, err := parsePort()
	if err != nil {
		log.Fatalf("Failed to parse port: %s", err)
	}

	log.Printf("Starting server on :%d", port)

	if err := server.Start(port); err != nil {
		log.Printf("An error occurred, the server stopped %s", err)
		os.Exit(1)
	}
}
