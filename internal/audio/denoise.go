package audio

import (
	"fmt"
	"os"
	"os/exec"
)

func DenoiseAudioFile(inputPath string) (string, error) {
	wavPath := inputPath[:len(inputPath)-5] + ".wav"
	if err := convertAIFFToWAV(inputPath, wavPath); err != nil {
		return "", fmt.Errorf("AIFF to WAV conversion failed: %w", err)
	}
	defer os.Remove(wavPath)

	rawInput := wavPath[:len(wavPath)-4] + "_raw.pcm"
	cmd := exec.Command("sox", wavPath, "-t", "raw", "-r", "48000", "-b", "16", "-c", "1", "-e", "signed-integer", rawInput)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("WAV to PCM conversion failed: %w", err)
	}
	defer os.Remove(rawInput)

	rawOutput := wavPath[:len(wavPath)-4] + "_denoised_raw.pcm"
	cmd = exec.Command("./rnnoise/examples/.libs/rnnoise_demo", rawInput, rawOutput)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("RNNoise processing failed: %w", err)
	}
	defer os.Remove(rawOutput)

	denoisedWAV := wavPath[:len(wavPath)-4] + "_denoised.wav"
	cmd = exec.Command("sox", "-t", "raw", "-r", "48000", "-b", "16", "-c", "1", "-e", "signed-integer", rawOutput, denoisedWAV)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("PCM to WAV conversion failed: %w", err)
	}
	defer os.Remove(denoisedWAV)

	denoisedPath := inputPath[:len(inputPath)-5] + "_denoised.aiff"
	if err := convertWAVToAIFF(denoisedWAV, denoisedPath); err != nil {
		return "", fmt.Errorf("WAV to AIFF conversion failed: %w", err)
	}

	return denoisedPath, nil
}

func convertAIFFToWAV(inputPath, outputPath string) error {
	cmd := exec.Command("sox", inputPath, outputPath)
	return cmd.Run()
}

func convertWAVToAIFF(inputPath, outputPath string) error {
	cmd := exec.Command("sox", inputPath, outputPath)
	return cmd.Run()
}

func CheckRNNoiseAvailable() error {
	// Check sox
	if err := exec.Command("sox", "--version").Run(); err != nil {
		return fmt.Errorf("sox not found: %w", err)
	}

	// Check rnnoise_demo
	if err := exec.Command("./rnnoise/examples/.libs/rnnoise_demo").Run(); err != nil {
		return fmt.Errorf("rnnoise_demo not found: %w", err)
	}

	return nil
}
