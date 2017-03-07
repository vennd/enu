package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	env := os.Getenv("ENV")
	hostname, err := os.Hostname()

	if err != nil {
		hostname = "unknown environment"
	}

	if env == "" {
		env = "unknown host"
	}

	router := NewRouter()

	log.Printf("Enu %s API server started on %s", env, hostname)
	log.Fatal(http.ListenAndServe("localhost:8080", router))
}
