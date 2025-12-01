package audio

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gordonklaus/portaudio"
)

var lastRecordedFile string

type AIFFAudioFormat struct {
	AudioFile       *os.File
	Channel         int16
	BitsPerSample   int16
	SampleRate      float64
	NumberOfSamples int32
	InputBufferSize int
	RecControlSig   *RecondControlSignal
}

func NewAIFFAudioFormat() *AIFFAudioFormat {
	return &AIFFAudioFormat{}
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

func (af *AIFFAudioFormat) GetFileType() string { return "aiff" }

func (af *AIFFAudioFormat) Record() {
	if af.AudioFile == nil {
		panic("audio file not initialized")
	}
	fmt.Println("Starting AIFF recording...")
	defer af.WrapUp()

	if err := portaudio.Initialize(); err != nil {
		panic(err)
	}
	defer portaudio.Terminate()

	in := make([]int32, af.InputBufferSize)

	// Default input device
	inputDevice, err := portaudio.DefaultInputDevice()
	if err != nil {
		panic("no input device found: " + err.Error())
	}
	fmt.Println("Using input device:", inputDevice.Name)

	stream, err := portaudio.OpenDefaultStream(int(af.Channel), 0, af.SampleRate, len(in), in)
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
