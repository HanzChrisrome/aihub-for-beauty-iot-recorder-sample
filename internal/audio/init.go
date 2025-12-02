package audio

type IAudioFormat interface {
	Init(recordControlSig *RecondControlSignal, sysPath, filename string, targetChannel int16, sampleRate float64, inputBufSize int)
	Record()
	GetFileType() string
	SetDeviceIndex(deviceIndex int)
}

const (
	AUDIO_CTL_STOP_REC          = 0
	AUDIO_CTL_REC_FULLY_STOPPED = 1
	AUDIO_GRACE_KILL_SIG_REQ    = 2
	AUDIO_GRACE_KILL_SIG_PROC   = 3
)

type RecondControlSignal struct {
	Sig chan int
}

func NewRecControlSig() *RecondControlSignal {
	return &RecondControlSignal{
		Sig: make(chan int),
	}
}

// NewAudioInstance creates a new audio format instance based on type
func NewAudioInstance(audioType string) IAudioFormat {
	switch audioType {
	case "aiff":
		return NewAIFFAudioFormat()

	default:
		return NewAIFFAudioFormat()
	}
}
