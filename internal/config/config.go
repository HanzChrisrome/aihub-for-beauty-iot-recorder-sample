package config

import (
	"fmt"
	"os"
	"strconv"
)

var BackendHost = "aeronsarondo.site"
var WebSocketPath = "/ws"
var ReconnectSeconds = 5

type Config struct {
	SYS_TCP_PORT                uint8
	SYS_RECORD_PATH             string
	SYS_AUDIO_TYPE              uint8
	SYS_AUDIO_CHANNEL           int16
	SYS_AUDIO_SAMPLE_RATE       float64
	SYS_AUDIO_INPUT_BUFFER_SIZE int
}

func Load() *Config {

	cfgRecordPath := loadEnv("SYS_RECORD_PATH", "./recordings")
	cfgAudioType := loadEnv("SYS_AUDIO_TYPE", "1")
	cfgAudioChannel := loadEnv("SYS_AUDIO_CHANNEL", "1")
	cfgAudioSampleRate := loadEnv("SYS_AUDIO_SAMPLE_RATE", "48000")
	cfgAudioInputBufferSize := loadEnv("SYS_AUDIO_INPUT_BUFFER_SIZE", "64")

	audioType, err := strconv.Atoi(cfgAudioType)
	must(err)
	sysAudioType := uint8(audioType)

	audioChannel, err := strconv.Atoi(cfgAudioChannel)
	must(err)
	sysAudioChannel := int16(audioChannel)

	sysAudioSampleRate, err := strconv.ParseFloat(cfgAudioSampleRate, 64)
	must(err)

	audioInputBufferSize, err := strconv.Atoi(cfgAudioInputBufferSize)
	must(err)
	sysAudioInputBufferSize := audioInputBufferSize // int now

	return &Config{
		SYS_RECORD_PATH:             cfgRecordPath,
		SYS_AUDIO_TYPE:              sysAudioType,
		SYS_AUDIO_CHANNEL:           sysAudioChannel,
		SYS_AUDIO_SAMPLE_RATE:       sysAudioSampleRate,
		SYS_AUDIO_INPUT_BUFFER_SIZE: sysAudioInputBufferSize,
	}
}

func loadEnv(key, defaultValue string) string {
	cfg := os.Getenv(key)
	if cfg == "" {
		if defaultValue == "" {
			panic(fmt.Sprintf("env var \"%s\" is required", key))
		}
		return defaultValue
	}
	return cfg
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
