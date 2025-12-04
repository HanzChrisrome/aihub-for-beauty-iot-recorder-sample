package audio

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gordonklaus/portaudio"
)

// AudioDevice represents an audio input device
type AudioDevice struct {
	Index             int     `json:"index"`
	Name              string  `json:"name"`
	MaxInputChannels  int     `json:"max_input_channels"`
	DefaultSampleRate float64 `json:"default_sample_rate"`
	HostAPI           string  `json:"host_api"`
}

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

func ListAudioDevices() []AudioDevice {
	if err := portaudio.Initialize(); err != nil {
		log.Printf("Failed to initialize PortAudio: %v", err)
		return []AudioDevice{}
	}
	defer portaudio.Terminate()

	devices, err := portaudio.Devices()
	if err != nil {
		log.Printf("Failed to get devices: %v", err)
		return []AudioDevice{}
	}

	var audioDevices []AudioDevice
	for i, device := range devices {
		if device.MaxInputChannels > 0 && isHardwareDevice(device.Name) {
			hostAPIName := "Unknown"
			if device.HostApi != nil {
				hostAPIName = device.HostApi.Name
			}

			audioDevices = append(audioDevices, AudioDevice{
				Index:             i,
				Name:              device.Name,
				MaxInputChannels:  device.MaxInputChannels,
				DefaultSampleRate: device.DefaultSampleRate,
				HostAPI:           hostAPIName,
			})
		}
	}

	log.Printf("Found %d audio input devices", len(audioDevices))
	return audioDevices
}

func isHardwareDevice(name string) bool {
	n := strings.ToLower(name)
	if strings.Contains(n, "hw:") {
		return true
	}

	virtual := []string{"default", "sysdefault", "spdif"}
	for _, v := range virtual {
		if n == v || strings.Contains(n, v) {
			return false
		}
	}

	return false
}

// GetDeviceByIndex returns device information for a specific index
func GetDeviceByIndex(index int) (*AudioDevice, error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize PortAudio: %w", err)
	}
	defer portaudio.Terminate()

	devices, err := portaudio.Devices()
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	if index < 0 || index >= len(devices) {
		return nil, fmt.Errorf("device index %d out of range", index)
	}

	device := devices[index]
	if device.MaxInputChannels == 0 {
		return nil, fmt.Errorf("device %d has no input channels", index)
	}

	hostAPIName := "Unknown"
	if device.HostApi != nil {
		hostAPIName = device.HostApi.Name
	}

	return &AudioDevice{
		Index:             index,
		Name:              device.Name,
		MaxInputChannels:  device.MaxInputChannels,
		DefaultSampleRate: device.DefaultSampleRate,
		HostAPI:           hostAPIName,
	}, nil
}

func GetDeviceIndexByName(name string) (int, error) {
	if err := portaudio.Initialize(); err != nil {
		return -1, fmt.Errorf("failed to initialize PortAudio: %w", err)
	}

	defer portaudio.Terminate()

	devices, err := portaudio.Devices()
	if err != nil {
		return -1, fmt.Errorf("failed to get devices: %w", err)
	}

	for i, device := range devices {
		if strings.Contains(strings.ToLower(device.Name), strings.ToLower(name)) {
			return i, nil
		}
	}
	return -1, fmt.Errorf("device with name containing %q not found", name)
}

func GetDevicesJSON() ([]byte, error) {
	devices := ListAudioDevices()
	return json.MarshalIndent(devices, "", "  ")
}

func PrintAvailableDevicesJSON() {
	b, err := GetDevicesJSON()
	if err != nil {
		log.Printf("Failed to marshal devices to JSON: %v", err)
		return
	}
	fmt.Println(string(b))
}

// GetDefaultInputDevice returns the default input device
func GetDefaultInputDevice() (*AudioDevice, error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize PortAudio: %w", err)
	}
	defer portaudio.Terminate()

	defaultDevice, err := portaudio.DefaultInputDevice()
	if err != nil {
		return nil, fmt.Errorf("failed to get default input device: %w", err)
	}

	devices, err := portaudio.Devices()
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	// Find the index of the default device
	for i, device := range devices {
		if device == defaultDevice {
			hostAPIName := "Unknown"
			if device.HostApi != nil {
				hostAPIName = device.HostApi.Name
			}

			return &AudioDevice{
				Index:             i,
				Name:              device.Name,
				MaxInputChannels:  device.MaxInputChannels,
				DefaultSampleRate: device.DefaultSampleRate,
				HostAPI:           hostAPIName,
			}, nil
		}
	}

	return nil, fmt.Errorf("default device not found in device list")
}
