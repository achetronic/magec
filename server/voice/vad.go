/*
 * Copyright 2025 Alby Hernández
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package voice

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	ort "github.com/yalue/onnxruntime_go"
)

const (
	// VAD configuration
	VADSampleRate     = 16000
	VADWindowSamples  = 512 // ~32ms at 16kHz (Silero VAD uses 512 samples)
	VADThreshold      = 0.5 // Speech probability threshold
	VADSilenceTimeout = 2000 * time.Millisecond
)

// VADConfig holds configuration for voice activity detection
type VADConfig struct {
	ModelPath      string
	Threshold      float32
	SilenceTimeout time.Duration
}

// VAD implements voice activity detection using Silero VAD ONNX model
type VAD struct {
	config  VADConfig
	logger  *slog.Logger
	session *ort.DynamicAdvancedSession

	// LSTM hidden states (maintained across calls)
	h []float32
	c []float32

	// State tracking
	isSpeaking       bool
	lastSpeechTime   time.Time
	audioBuffer      []float32
	mu               sync.Mutex

	// Callbacks
	onSpeechStart func()
	onSpeechEnd   func()
}

// NewVAD creates a new voice activity detector
func NewVAD(config VADConfig, logger *slog.Logger) *VAD {
	if config.Threshold == 0 {
		config.Threshold = VADThreshold
	}
	if config.SilenceTimeout == 0 {
		config.SilenceTimeout = VADSilenceTimeout
	}

	return &VAD{
		config:      config,
		logger:      logger,
		audioBuffer: make([]float32, 0, VADWindowSamples*10),
		// Initialize LSTM states: [2, 1, 64] flattened = 128 elements
		h: make([]float32, 2*1*64),
		c: make([]float32, 2*1*64),
	}
}

// SetOnSpeechStart sets the callback for when speech starts
func (v *VAD) SetOnSpeechStart(callback func()) {
	v.onSpeechStart = callback
}

// SetOnSpeechEnd sets the callback for when speech ends
func (v *VAD) SetOnSpeechEnd(callback func()) {
	v.onSpeechEnd = callback
}

// Load initializes the ONNX model
func (v *VAD) Load() error {
	var err error
	v.session, err = ort.NewDynamicAdvancedSession(
		v.config.ModelPath,
		[]string{"input", "sr", "h", "c"},
		[]string{"output", "hn", "cn"},
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to load VAD model: %w", err)
	}

	v.logger.Info("VAD model loaded", "path", v.config.ModelPath)
	return nil
}

// ProcessAudio processes audio samples and detects voice activity
// Samples should be float32 in range [-1, 1] at 16kHz
func (v *VAD) ProcessAudio(samples []float32) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Add samples to buffer
	v.audioBuffer = append(v.audioBuffer, samples...)

	// Process in windows of VADWindowSamples
	for len(v.audioBuffer) >= VADWindowSamples {
		window := v.audioBuffer[:VADWindowSamples]
		v.audioBuffer = v.audioBuffer[VADWindowSamples:]

		prob, err := v.runInference(window)
		if err != nil {
			return err
		}

		v.updateState(prob)
	}

	// Check for silence timeout
	if v.isSpeaking && time.Since(v.lastSpeechTime) > v.config.SilenceTimeout {
		v.isSpeaking = false
		v.logger.Debug("VAD: speech ended (silence timeout)")
		if v.onSpeechEnd != nil {
			go v.onSpeechEnd()
		}
	}

	return nil
}

func (v *VAD) runInference(samples []float32) (float32, error) {
	// Create input tensor [1, 512]
	inputShape := ort.NewShape(1, int64(len(samples)))
	inputTensor, err := ort.NewTensor(inputShape, samples)
	if err != nil {
		return 0, fmt.Errorf("failed to create input tensor: %w", err)
	}
	defer inputTensor.Destroy()

	// Create sample rate tensor (scalar int64)
	// Note: Silero VAD expects a scalar but onnxruntime_go needs at least 1 dimension
	srTensor, err := ort.NewTensor(ort.NewShape(1), []int64{VADSampleRate})
	if err != nil {
		return 0, fmt.Errorf("failed to create sr tensor: %w", err)
	}
	defer srTensor.Destroy()

	// Create h tensor [2, 1, 64]
	hShape := ort.NewShape(2, 1, 64)
	hTensor, err := ort.NewTensor(hShape, v.h)
	if err != nil {
		return 0, fmt.Errorf("failed to create h tensor: %w", err)
	}
	defer hTensor.Destroy()

	// Create c tensor [2, 1, 64]
	cShape := ort.NewShape(2, 1, 64)
	cTensor, err := ort.NewTensor(cShape, v.c)
	if err != nil {
		return 0, fmt.Errorf("failed to create c tensor: %w", err)
	}
	defer cTensor.Destroy()

	// Run inference
	outputs := []ort.Value{nil, nil, nil}
	err = v.session.Run([]ort.Value{inputTensor, srTensor, hTensor, cTensor}, outputs)
	if err != nil {
		return 0, fmt.Errorf("VAD inference failed: %w", err)
	}
	defer outputs[0].Destroy()
	defer outputs[1].Destroy()
	defer outputs[2].Destroy()

	// Get output probability
	outputTensor, ok := outputs[0].(*ort.Tensor[float32])
	if !ok {
		return 0, fmt.Errorf("VAD output is not float32 tensor")
	}
	prob := outputTensor.GetData()[0]

	// Update LSTM states for next call
	hnTensor, ok := outputs[1].(*ort.Tensor[float32])
	if ok {
		copy(v.h, hnTensor.GetData())
	}
	cnTensor, ok := outputs[2].(*ort.Tensor[float32])
	if ok {
		copy(v.c, cnTensor.GetData())
	}

	return prob, nil
}

func (v *VAD) updateState(prob float32) {
	wasSpeaking := v.isSpeaking

	// Log every probability for debugging
	v.logger.Debug("VAD inference", "prob", fmt.Sprintf("%.4f", prob), "threshold", v.config.Threshold, "speaking", v.isSpeaking)

	if prob >= v.config.Threshold {
		v.lastSpeechTime = time.Now()
		if !v.isSpeaking {
			v.isSpeaking = true
			v.logger.Debug("VAD: speech started", "prob", prob)
			if v.onSpeechStart != nil {
				go v.onSpeechStart()
			}
		}
	} else if v.isSpeaking && time.Since(v.lastSpeechTime) > v.config.SilenceTimeout {
		v.isSpeaking = false
		v.logger.Debug("VAD: speech ended", "prob", prob, "silence", time.Since(v.lastSpeechTime))
		if v.onSpeechEnd != nil {
			go v.onSpeechEnd()
		}
	}

	if wasSpeaking != v.isSpeaking {
		v.logger.Debug("VAD state changed", "speaking", v.isSpeaking, "prob", prob)
	}
}

// Reset clears the internal state
func (v *VAD) Reset() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.audioBuffer = v.audioBuffer[:0]
	v.isSpeaking = false
	// Reset LSTM states
	for i := range v.h {
		v.h[i] = 0
	}
	for i := range v.c {
		v.c[i] = 0
	}
}

// IsSpeaking returns whether speech is currently detected
func (v *VAD) IsSpeaking() bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.isSpeaking
}

// Close releases all resources
func (v *VAD) Close() {
	if v.session != nil {
		v.session.Destroy()
	}
}

// GetConfig returns the current VAD configuration
func (v *VAD) GetConfig() VADConfig {
	return v.config
}
