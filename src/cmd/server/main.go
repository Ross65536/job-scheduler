package main

import (
	"flag"
	"log"
	"os"

	"github.com/Ross65536/job-scheduler/src/backend"
)

const (
	defaultPrivateKeyPath  = "certs/server.key"
	defaultCertificatePath = "certs/server.crt"
)

func parseFlags() (listenPort int, certificatePath string, privateKeyPath string) {
	port := flag.Int("p", 10000, "port to listen on")
	certificate := flag.String("cert", defaultCertificatePath, "path to the server's public certificate")
	privateKey := flag.String("privateKey", defaultPrivateKeyPath, "path to the server's private key, matching the certificate")

	flag.Parse()

	return *port, *certificate, *privateKey
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

	port, certificatePath, privateKeyPath := parseFlags()
	if port < 0 || port > 65535 {
		log.Fatalf("invalid port value")
	}

	log.Printf("Starting server on :%d", port)

	if err := server.StartWithTls(port, certificatePath, privateKeyPath); err != nil {
		log.Printf("An error occurred, the server stopped %s", err)
		os.Exit(1)
	}
}
