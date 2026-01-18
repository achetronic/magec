// Download pretrained models (VAD, embedding, etc) from HuggingFace Hey-Buddy.
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const baseURL = "https://huggingface.co/benjamin-paine/hey-buddy/resolve/main"

var pretrainedModels = []string{
	"silero-vad.onnx",
	"mel-spectrogram.onnx",
	"speech-embedding.onnx",
}

func downloadFile(url, dest string) error {
	fmt.Printf("  Downloading %s... ", filepath.Base(dest))

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("FAILED")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("FAILED")
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		fmt.Println("FAILED")
		return err
	}

	out, err := os.Create(dest)
	if err != nil {
		fmt.Println("FAILED")
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		fmt.Println("FAILED")
		return err
	}

	fmt.Println("OK")
	return nil
}

func downloadPretrained() error {
	fmt.Println("\nChecking pretrained models...")

	for _, model := range pretrainedModels {
		dest := filepath.Join("gui", "pretrained", model)

		if _, err := os.Stat(dest); err == nil {
			fmt.Printf("  %s already exists\n", model)
			continue
		}

		url := fmt.Sprintf("%s/pretrained/%s", baseURL, model)
		if err := downloadFile(url, dest); err != nil {
			return fmt.Errorf("failed to download %s: %w", model, err)
		}
	}

	return nil
}

func main() {

	if err := downloadPretrained(); err != nil {
		fmt.Printf("\nFailed to download pretrained models: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nDone!")
}
