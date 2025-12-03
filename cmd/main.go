package main

import (
	"log"

	"github.com/otis-co-ltd/aihub-recorder/internal/pi"
	"github.com/otis-co-ltd/aihub-recorder/internal/wsclient"
)

func main() {
	piID := pi.GetPiId()
	log.Println("Starting AIHub recorder WebSocket client with Pi ID:", piID)

	wsclient.Start(piID)
}
