package main

import (
	"log"
	"net/http"
)

const (
	addr     = ":24744"
	certFile = "server.crt" // TODO: Generate self-signed certificate
	keyFile  = "server.key" // TODO: Generate self-signed key
)

func main() {
	httpQ := &HTTPQ{}

	log.Printf("Starting httpQ server on https://localhost%s\n", addr)

	// TODO: Use http.ListenAndServeTLS for HTTPS
	// For development, you can generate self-signed certs with:
	//   openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes -subj "/CN=localhost"
	if err := http.ListenAndServeTLS(addr, certFile, keyFile, httpQ.Handler()); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
