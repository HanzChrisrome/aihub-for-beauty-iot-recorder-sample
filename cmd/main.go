package main

import (
	"log"
	"os"

	"github.com/otis-co-ltd/aihub-recorder/internal/wsclient"
)

func main() {
	piID := os.Getenv("PI_ID")
	if piID == "" {
		piID = "pi01"
	}

	log.Println("Starting AIHub recorder WebSocket client with Pi ID:", piID)

	wsclient.Start(piID)
}
