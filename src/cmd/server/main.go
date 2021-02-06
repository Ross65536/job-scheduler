package main

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/ros-k/job-manager/src/backend"
)

const (
	defaultPrivateKeyPath  = "certs/server.key"
	defaultCertificatePath = "certs/server.crt"
)

func parsePort() (int, string, string, error) {
	port := flag.Int("p", 10000, "port to listen on")
	certificate := flag.String("cert", defaultCertificatePath, "path to the server's public certificate")
	privateKey := flag.String("privateKey", defaultPrivateKeyPath, "path to the server's private key, matching the certificate")

	flag.Parse()

	if *port < 0 || *port > 65535 {
		return 0, "", "", errors.New("invalid port value")
	}

	return *port, *certificate, *privateKey, nil
}

func start(server *backend.Server, port int, certificatePath, privateKeyPath string) error {
	if certificatePath == "" && privateKeyPath == "" {
		return server.Start(port)
	}

	return server.StartWithTls(port, certificatePath, privateKeyPath)
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

	port, certificatePath, privateKeyPath, err := parsePort()
	if err != nil {
		log.Fatalf("Failed to parse port: %s", err)
	}

	log.Printf("Starting server on :%d", port)

	if err := start(server, port, certificatePath, privateKeyPath); err != nil {
		log.Printf("An error occurred, the server stopped %s", err)
		os.Exit(1)
	}
}
