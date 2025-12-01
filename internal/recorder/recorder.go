package recorder

import (
	"errors"

	"github.com/otis-co-ltd/aihub-recorder/internal/audio"
	"github.com/otis-co-ltd/aihub-recorder/internal/config"
)

var currentRecorder audio.IAudioFormat
var recControl *audio.RecondControlSignal
var cfg *config.Config

func init() {
	cfg = config.Load()
}

// Start starts a single-mic recording
func Start() error {
	if currentRecorder != nil {
		return errors.New("recording already in progress")
	}

	recControl = audio.NewRecControlSig()
	currentRecorder = audio.NewAudioInstance(cfg.SYS_AUDIO_TYPE)

	// Note: input buffer size is int now
	currentRecorder.Init(
		recControl,
		cfg.SYS_RECORD_PATH,
		"pi_recording",
		int16(cfg.SYS_AUDIO_CHANNEL),         // number of channels
		float64(cfg.SYS_AUDIO_SAMPLE_RATE),   // sample rate
		int(cfg.SYS_AUDIO_INPUT_BUFFER_SIZE), // buffer size
	)

	go currentRecorder.Record()
	return nil
}

// Stop stops the recording
func Stop() error {
	if currentRecorder == nil || recControl == nil {
		return errors.New("no recording in progress")
	}

	// send stop signal
	recControl.Sig <- audio.AUDIO_CTL_STOP_REC

	// wait until fully stopped
	<-recControl.Sig

	currentRecorder = nil
	recControl = nil
	return nil
}
