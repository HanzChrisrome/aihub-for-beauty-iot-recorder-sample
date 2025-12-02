package recorder

import (
	"errors"

	"github.com/otis-co-ltd/aihub-recorder/internal/audio"
	"github.com/otis-co-ltd/aihub-recorder/internal/config"
)

var currentRecorder audio.IAudioFormat
var recControl *audio.RecondControlSignal

func init() {
	cfg = config.Load()
}

// Start starts a single-mic recording
func Start() error {
    if currentRecorder != nil {
        return errors.New("recording already in progress")
    }

    return StartSession("default_session", 0)
}

// Stop stops the recording
func Stop() error {
    if currentRecorder == nil || recControl == nil {
        return errors.New("no recording in progress")
    }

    _, err := StopSession("default_session")
    return err
}
