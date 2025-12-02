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

	fmt.Println("\n Starting recording from BOTH microphones for 10 seconds...")
	err := recorder.StartSession("mic1", 0)
	if err != nil {
		log.Fatal("Failed to start recording session for mic1:", err)
	}
	fmt.Println("Mic 1 started!")

	err = recorder.StartSession("mic2", 1)
	if err != nil {
		log.Fatal("Failed to start recording session for mic2:", err)
	}
	fmt.Println("Mic 2 started!")

	fmt.Println("\n?? Recording from both microphones...")
	for i := 10; i > 0; i-- {
		fmt.Printf("  ??  %d seconds remaining...\n", i)
		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n??  Stopping recordings...")

	file1, err := recorder.StopSession("mic1")
	if err != nil {
		log.Fatal("? Failed to stop mic1:", err)
	}
	fmt.Printf("? Mic 1 saved: %s\n", file1)

	file2, err := recorder.StopSession("mic2")
	if err != nil {
		log.Fatal("? Failed to stop mic2:", err)
	}
	fmt.Printf("? Mic 2 saved: %s\n", file2)

	// Summary
	fmt.Println("\nTest Complete!")
	fmt.Println("================")
	fmt.Println("  1. recordings/mic1/device_0_*.aiff (USB MIC PRO)")
	fmt.Println("  2. recordings/mic2/device_1_*.aiff (USB Condenser Microphone)")

}
