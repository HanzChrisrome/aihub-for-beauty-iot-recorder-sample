package main

import (
	"fmt"
	"log"
	"time"

	"github.com/otis-co-ltd/aihub-recorder/internal/audio"
	"github.com/otis-co-ltd/aihub-recorder/internal/recorder"
)

func main() {
	fmt.Println("??? Starting simple recording test...")

	if err := audio.InitPortAudio(); err != nil {
		log.Fatal("Failed to initialize PortAudio:", err)
	}

	fmt.Println("\n Available Devices:")
	audio.PrintAvailableDevices()

	idx1, err := audio.GetDeviceIndexByName("USB Condenser Microphone: Audio (hw:2,0)")
	if err != nil {
		log.Fatalf("Failed to resolve device name for mic2: %v", err)
	}

	if err := recorder.StartSession("mic1", idx1); err != nil {
		log.Fatalf("Failed to start recording for mic1: %v", err)
	}

	fmt.Println("Recording for mic1... Press Enter to stop.")
	for i := 10; i > 0; i-- {
		fmt.Printf("  ??  %d seconds remaining...\n", i)
		time.Sleep(1 * time.Second)
	}

	file1, err := recorder.StopSession("mic1")
	if err != nil {
		log.Fatal("? Failed to stop mic1:", err)
	}
	fmt.Printf("? Mic 1 saved: %s\n", file1)

	fmt.Println("\nTest Complete!")

}
