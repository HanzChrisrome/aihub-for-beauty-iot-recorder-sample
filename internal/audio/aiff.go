package audio

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gordonklaus/portaudio"
)

var lastRecordedFile string
var portaudioInitialized bool
var portaudioMutex sync.Mutex

type AIFFAudioFormat struct {
	AudioFile       *os.File
	Channel         int16
	BitsPerSample   int16
	SampleRate      float64
	NumberOfSamples int32
	InputBufferSize int
	RecControlSig   *RecondControlSignal
	DeviceIndex     int
}

func NewAIFFAudioFormat() *AIFFAudioFormat {
	return &AIFFAudioFormat{
		DeviceIndex: -1,
	}
}

func InitPortAudio() error {
	portaudioMutex.Lock()
	defer portaudioMutex.Unlock()

	if !portaudioInitialized {
		if err := portaudio.Initialize(); err != nil {
			return err
		}
		portaudioInitialized = true
		log.Println("? PortAudio initialized globally")
	}
	return nil
}

func (af *AIFFAudioFormat) CreateFilePath(sysPath, filename string) string {
	return filepath.Join(sysPath, fmt.Sprintf("%s.%s", filename, af.GetFileType()))
}

func (af *AIFFAudioFormat) Init(recordControlSig *RecondControlSignal, sysPath, filename string, channel int16, sampleRate float64, inputBufSize int) {
	af.RecControlSig = recordControlSig
	af.Channel = channel
	af.SampleRate = sampleRate
	af.BitsPerSample = 32
	af.InputBufferSize = inputBufSize
	af.NumberOfSamples = 0

	filePath := af.CreateFilePath(sysPath, filename)
	lastRecordedFile = filePath
	if sysPath != "" {
		os.MkdirAll(sysPath, os.ModePerm)
	}

	file, err := os.Create(filePath)
	must(err)

	file.WriteString("FORM")
	must(binary.Write(file, binary.BigEndian, int32(0)))
	file.WriteString(strings.ToUpper(af.GetFileType()))
	file.WriteString("COMM")
	must(binary.Write(file, binary.BigEndian, int32(18)))
	must(binary.Write(file, binary.BigEndian, af.Channel))
	must(binary.Write(file, binary.BigEndian, af.NumberOfSamples))
	must(binary.Write(file, binary.BigEndian, af.BitsPerSample))
	file.Write(SampleRateToByte(af.SampleRate))
	file.WriteString("SSND")
	must(binary.Write(file, binary.BigEndian, int32(0)))
	must(binary.Write(file, binary.BigEndian, int32(0)))
	must(binary.Write(file, binary.BigEndian, int32(0)))

	af.AudioFile = file
}

func (af *AIFFAudioFormat) SetDeviceIndex(deviceIndex int) {
	af.DeviceIndex = deviceIndex
}

func (af *AIFFAudioFormat) GetFileType() string { return "aiff" }

func (af *AIFFAudioFormat) Record() {
	if af.AudioFile == nil {
		panic("audio file not initialized")
	}

	log.Printf("🔧 DEBUG: Starting recording with DeviceIndex=%d", af.DeviceIndex)

	fmt.Println("Starting AIFF recording...")
	defer af.WrapUp()

	if err := portaudio.Initialize(); err != nil {
		panic(err)
	}

	in := make([]int32, af.InputBufferSize)

	// Get the specific device by index
	var inputDevice *portaudio.DeviceInfo
	var err error

	if af.DeviceIndex >= 0 {
		// Use specific device index
		devices, err := portaudio.Devices()
		if err != nil {
			panic("failed to get devices: " + err.Error())
		}

		if af.DeviceIndex >= len(devices) {
			panic(fmt.Sprintf("device index %d out of range (max %d)", af.DeviceIndex, len(devices)-1))
		}

		inputDevice = devices[af.DeviceIndex]
		fmt.Printf("Using input device [%d]: %s\n", af.DeviceIndex, inputDevice.Name)
	} else {
		// Fallback to default input device
		inputDevice, err = portaudio.DefaultInputDevice()
		if err != nil {
			panic("no input device found: " + err.Error())
		}
		fmt.Println("Using default input device:", inputDevice.Name)
	}

	// Open stream with specific device
	streamParams := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   inputDevice,
			Channels: int(af.Channel),
			Latency:  inputDevice.DefaultLowInputLatency,
		},
		SampleRate:      af.SampleRate,
		FramesPerBuffer: len(in),
	}

	stream, err := portaudio.OpenStream(streamParams, in)
	must(err)
	defer stream.Close()
	must(stream.Start())

	for {
		must(stream.Read())
		must(binary.Write(af.AudioFile, binary.BigEndian, in))
		af.NumberOfSamples += int32(len(in))

		select {
		case ctl := <-af.RecControlSig.Sig:
			if ctl == AUDIO_CTL_STOP_REC {
				must(stream.Stop())
				af.RecControlSig.Sig <- AUDIO_CTL_REC_FULLY_STOPPED
				return
			}
			if ctl == AUDIO_GRACE_KILL_SIG_REQ {
				must(stream.Stop())
				af.WrapUp()
				af.RecControlSig.Sig <- AUDIO_GRACE_KILL_SIG_PROC
				return
			}
		default:
		}
	}
}

func (af *AIFFAudioFormat) WrapUp() {
	if af.AudioFile == nil {
		log.Fatal("audio file empty")
	}

	totalBytes := 4 + 8 + 18 + 8 + 8 + 4*af.NumberOfSamples
	_, err := af.AudioFile.Seek(4, 0)
	must(err)
	must(binary.Write(af.AudioFile, binary.BigEndian, totalBytes))

	_, err = af.AudioFile.Seek(22, 0)
	must(err)
	must(binary.Write(af.AudioFile, binary.BigEndian, af.NumberOfSamples))

	_, err = af.AudioFile.Seek(42, 0)
	must(err)
	must(binary.Write(af.AudioFile, binary.BigEndian, int32(4*af.NumberOfSamples+8)))

	must(af.AudioFile.Close())
	fmt.Println("AIFF recording finished")
}

func GetLastFilePath() string {
	return lastRecordedFile
}
