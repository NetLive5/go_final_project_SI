package main

import (
	"log"
	"os"

	"go_final_project/pkg/server"
)

func main() {
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	log.Printf("Starting task scheduler on :%s", port)
	if err := server.Run(port, "web"); err != nil {
		log.Fatal(err)
	}
}
