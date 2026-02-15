// Copyright 2025 Alby Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Download auxiliary models (VAD, embedding, etc) from HuggingFace Hey-Buddy.
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const baseURL = "https://huggingface.co/benjamin-paine/hey-buddy/resolve/main"

var auxiliaryModels = []string{
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

func downloadAuxiliary() error {
	fmt.Println("\nChecking auxiliary models...")

	for _, model := range auxiliaryModels {
		dest := filepath.Join("models", "auxiliary", model)

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

	if err := downloadAuxiliary(); err != nil {
		fmt.Printf("\nFailed to download auxiliary models: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nDone!")
}
