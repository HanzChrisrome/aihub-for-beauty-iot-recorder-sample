package audio

import (
	"fmt"
	"log"

	"github.com/gordonklaus/portaudio"
)

func SampleRateToByte(sampleRate float64) []byte {
	switch sampleRate {
	case 16000:
		return []byte{0x40, 0x0c, 0x7a, 0, 0, 0, 0, 0, 0, 0}
	case 44100:
		return []byte{0x40, 0x0e, 0xac, 0x44, 0, 0, 0, 0, 0, 0}
	case 48000:
		return []byte{0x40, 0x0e, 0xbb, 0x80, 0, 0, 0, 0, 0, 0}
	default:
		panic("unsupported sample rate")
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func ListDevices() {
	err := portaudio.Initialize()
	if err != nil {
		log.Fatal("Failed to init PortAudio:", err)
	}
	defer portaudio.Terminate()

	devices, err := portaudio.Devices()
	if err != nil {
		log.Fatal("Failed to list devices:", err)
	}

	fmt.Println("=== PortAudio Devices ===")
	for i, dev := range devices {
		fmt.Printf("\nDevice %d:\n", i)
		fmt.Printf("  Name: %s\n", dev.Name)
		fmt.Printf("  Max Input Channels: %d\n", dev.MaxInputChannels)
		fmt.Printf("  Max Output Channels: %d\n", dev.MaxOutputChannels)
		fmt.Printf("  Default Sample Rate: %.0f\n", dev.DefaultSampleRate)
		fmt.Printf("  Host API: %s\n", dev.HostApi.Name)
	}
}
