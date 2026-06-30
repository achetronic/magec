// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package voice

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"time"

	ort "github.com/yalue/onnxruntime_go"
)

// MelStep is the number of new audio samples (at 16 kHz) between successive
// mel-spectrogram frames.  OpenWakeWord uses a 10 ms step, i.e. 160 samples.
const melStepSamples = 160

// Global ONNX Runtime initialization (must only happen once)
var (
	ortInitOnce sync.Once
	ortInitErr  error
)

// newLightSessionOptions creates ONNX session options optimised for the small
// models used in voice detection.  By default ONNX Runtime spawns as many
// intra-op threads as there are CPU cores **for every session**.  With 4-5
// concurrent sessions (mel, embedding, wakeword, VAD) this saturates the CPU.
// These models are tiny, so a single thread per session is more than enough.
func newLightSessionOptions() (*ort.SessionOptions, error) {
	opts, err := ort.NewSessionOptions()
	if err != nil {
		return nil, err
	}
	if err := opts.SetIntraOpNumThreads(1); err != nil {
		opts.Destroy()
		return nil, err
	}
	if err := opts.SetInterOpNumThreads(1); err != nil {
		opts.Destroy()
		return nil, err
	}
	return opts, nil
}

const (
	TargetSampleRate  = 16000
	MelBins           = 32
	EmbeddingSize     = 96
	MelWindowSize     = 76
	MelWindowStep     = 8
	WakeWordFeatures  = 16
	DefaultCooldownMs = 2000
)

// ModelConfig represents a wake word model configuration
type ModelConfig struct {
	ID        string
	Name      string
	Data      []byte
	Phrase    string
	Threshold float32
}

// DetectorConfig holds configuration for the wake word detector
type DetectorConfig struct {
	MelspecModelData   []byte
	EmbeddingModelData []byte
	VADModelData       []byte
	Models             []ModelConfig
	OnnxLibraryPath    string
}

// Detector implements OpenWakeWord detection using ONNX Runtime
type Detector struct {
	config DetectorConfig
	logger *slog.Logger

	melspecSession   *ort.DynamicAdvancedSession
	embeddingSession *ort.DynamicAdvancedSession

	// Wake word models (can have multiple loaded)
	wakeWordSessions map[string]*wakeWordModel
	activeModelID    string

	audioBuffer       []int16
	processedSamples  int
	lastDetectionTime time.Time
	mu                sync.Mutex

	// Callbacks
	onDetected func(modelID string)
}

type wakeWordModel struct {
	config  ModelConfig
	session *ort.DynamicAdvancedSession
}

// NewDetector creates a new wake word detector
// NewDetector creates a new instance of the wake word and VAD detector.
func NewDetector(config DetectorConfig, logger *slog.Logger) *Detector {
	return &Detector{
		config:           config,
		logger:           logger,
		audioBuffer:      make([]int16, 0, TargetSampleRate*5),
		wakeWordSessions: make(map[string]*wakeWordModel),
	}
}

// SetOnDetected sets the callback for when wake word is detected
func (d *Detector) SetOnDetected(callback func(modelID string)) {
	d.onDetected = callback
}

// GetModels returns the list of available wake word models
func (d *Detector) GetModels() []ModelConfig {
	return d.config.Models
}

// SetActiveModel sets which wake word model to use for detection
func (d *Detector) SetActiveModel(modelID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.wakeWordSessions[modelID]; !ok {
		return fmt.Errorf("model %q not loaded", modelID)
	}
	d.activeModelID = modelID
	d.logger.Info("Active wake word model changed", "model", modelID)
	return nil
}

// GetActiveModel returns the currently active model ID
func (d *Detector) GetActiveModel() string {
	return d.activeModelID
}

// Load initializes the ONNX models
// Load initializes the ONNX runtime and loads the configured models into memory.
func (d *Detector) Load() error {
	d.logger.Info("Loading wake word models",
		"models", len(d.config.Models),
	)

	// Initialize ONNX Runtime environment (once globally)
	ortInitOnce.Do(func() {
		if d.config.OnnxLibraryPath != "" {
			ort.SetSharedLibraryPath(d.config.OnnxLibraryPath)
		}
		ortInitErr = ort.InitializeEnvironment()
	})
	if ortInitErr != nil {
		return fmt.Errorf("failed to initialize ONNX Runtime: %w", ortInitErr)
	}

	// Create lightweight session options (1 thread per session)
	opts, err := newLightSessionOptions()
	if err != nil {
		return fmt.Errorf("failed to create session options: %w", err)
	}
	defer opts.Destroy()

	// Load melspectrogram model
	d.melspecSession, err = ort.NewDynamicAdvancedSessionWithONNXData(
		d.config.MelspecModelData,
		[]string{"input"},
		[]string{"output"},
		opts,
	)
	if err != nil {
		return fmt.Errorf("failed to load melspec model: %w", err)
	}

	// Load embedding model
	d.embeddingSession, err = ort.NewDynamicAdvancedSessionWithONNXData(
		d.config.EmbeddingModelData,
		[]string{"input_1"},
		[]string{"conv2d_19"},
		opts,
	)
	if err != nil {
		return fmt.Errorf("failed to load embedding model: %w", err)
	}

	// Load all wake word models
	for _, modelCfg := range d.config.Models {
		session, err := ort.NewDynamicAdvancedSessionWithONNXData(
			modelCfg.Data,
			[]string{"onnx::Flatten_0"},
			[]string{"39"},
			opts,
		)
		if err != nil {
			d.logger.Warn("Failed to load wake word model", "model", modelCfg.ID, "error", err)
			continue
		}

		d.wakeWordSessions[modelCfg.ID] = &wakeWordModel{
			config:  modelCfg,
			session: session,
		}
		d.logger.Info("Loaded wake word model", "model", modelCfg.ID, "phrase", modelCfg.Phrase)
	}

	if len(d.wakeWordSessions) == 0 {
		return fmt.Errorf("no wake word models loaded")
	}

	// Set first model as active by default
	for id := range d.wakeWordSessions {
		d.activeModelID = id
		break
	}

	d.logger.Info("Wake word models loaded successfully", "active", d.activeModelID)
	return nil
}

// ProcessAudio processes incoming audio samples (expected to be 16kHz int16)
func (d *Detector) ProcessAudio(samples []int16) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Add to buffer
	d.audioBuffer = append(d.audioBuffer, samples...)

	// Keep max ~5 seconds of audio
	maxAudioLength := TargetSampleRate * 5
	if len(d.audioBuffer) > maxAudioLength {
		excess := len(d.audioBuffer) - maxAudioLength
		d.audioBuffer = d.audioBuffer[excess:]
		d.processedSamples -= excess
		if d.processedSamples < 0 {
			d.processedSamples = 0
		}
	}

	// Run inference whenever there are enough new samples to produce at
	// least one new mel-spectrogram frame (160 samples = 10 ms at 16 kHz).
	// This keeps detection latency as low as the model allows while still
	// avoiding redundant re-computation on identical audio.
	minSamples := TargetSampleRate * 2
	newSamples := len(d.audioBuffer) - d.processedSamples
	if len(d.audioBuffer) >= minSamples && newSamples >= melStepSamples {
		return d.processBuffer()
	}

	return nil
}

// ProcessAudioFloat32 processes incoming audio samples as float32 [-1, 1] and converts to int16
func (d *Detector) ProcessAudioFloat32(samples []float32) error {
	// Convert float32 to int16
	int16Samples := make([]int16, len(samples))
	for i, s := range samples {
		val := int(s * 32767)
		if val > 32767 {
			val = 32767
		} else if val < -32768 {
			val = -32768
		}
		int16Samples[i] = int16(val)
	}
	return d.ProcessAudio(int16Samples)
}

func (d *Detector) processBuffer() error {
	d.processedSamples = len(d.audioBuffer)

	startTime := time.Now()

	// Get active model
	activeModel, ok := d.wakeWordSessions[d.activeModelID]
	if !ok {
		return fmt.Errorf("no active model")
	}

	if d.logger.Enabled(context.Background(), slog.LevelDebug) {
		var sum float64
		var maxVal int16
		for _, s := range d.audioBuffer {
			sum += float64(s) * float64(s)
			abs := s
			if abs < 0 {
				abs = -abs
			}
			if abs > maxVal {
				maxVal = abs
			}
		}
		rms := math.Sqrt(sum / float64(len(d.audioBuffer)))

		d.logger.Debug("Processing buffer",
			"samples", len(d.audioBuffer),
			"rms", int(rms),
			"max", maxVal,
		)
	}

	// Step 1: Compute melspectrogram
	t1 := time.Now()
	melspec, err := d.getMelspectrogram(d.audioBuffer)
	if err != nil {
		return fmt.Errorf("melspectrogram failed: %w", err)
	}
	melTime := time.Since(t1)

	if len(melspec) < MelWindowSize {
		d.logger.Debug("Not enough melspec frames", "frames", len(melspec), "required", MelWindowSize)
		return nil
	}

	// Step 2: Compute embeddings
	t2 := time.Now()
	embeddings, err := d.getEmbeddings(melspec)
	if err != nil {
		return fmt.Errorf("embeddings failed: %w", err)
	}
	embTime := time.Since(t2)

	if len(embeddings) < WakeWordFeatures {
		d.logger.Debug("Not enough embeddings", "count", len(embeddings), "required", WakeWordFeatures)
		return nil
	}

	// Step 3: Run wake word detection on last 16 embeddings
	t3 := time.Now()
	features := embeddings[len(embeddings)-WakeWordFeatures:]
	score, err := d.runWakeWordModel(activeModel.session, features)
	if err != nil {
		return fmt.Errorf("wake word model failed: %w", err)
	}
	wwTime := time.Since(t3)

	totalTime := time.Since(startTime)

	d.logger.Debug("Wake word inference",
		"model", d.activeModelID,
		"score", fmt.Sprintf("%.3f", score),
		"threshold", activeModel.config.Threshold,
		"mel_ms", melTime.Milliseconds(),
		"emb_ms", embTime.Milliseconds(),
		"ww_ms", wwTime.Milliseconds(),
		"total_ms", totalTime.Milliseconds(),
	)

	// Check for detection
	if score >= activeModel.config.Threshold {
		now := time.Now()
		cooldown := time.Duration(DefaultCooldownMs) * time.Millisecond
		if now.Sub(d.lastDetectionTime) >= cooldown {
			d.logger.Info("Wake word DETECTED!", "model", d.activeModelID, "score", fmt.Sprintf("%.3f", score))
			d.lastDetectionTime = now
			d.audioBuffer = d.audioBuffer[:0] // Clear buffer

			if d.onDetected != nil {
				go d.onDetected(d.activeModelID)
			}
		}
	}

	// Trim buffer - keep last 2 seconds
	keepSamples := TargetSampleRate * 2
	if len(d.audioBuffer) > keepSamples {
		d.audioBuffer = d.audioBuffer[len(d.audioBuffer)-keepSamples:]
	}
	d.processedSamples = len(d.audioBuffer)

	return nil
}

func (d *Detector) getMelspectrogram(audioInt16 []int16) ([][]float32, error) {
	// Convert int16 to float32 for ONNX (NOT normalized - raw int16 values as float)
	audioFloat := make([]float32, len(audioInt16))
	for i, s := range audioInt16 {
		audioFloat[i] = float32(s)
	}

	// Create input tensor [1, N]
	inputShape := ort.NewShape(1, int64(len(audioFloat)))
	inputTensor, err := ort.NewTensor(inputShape, audioFloat)
	if err != nil {
		return nil, fmt.Errorf("failed to create input tensor: %w", err)
	}
	defer inputTensor.Destroy()

	// Run inference with nil output to let onnxruntime allocate it
	outputs := []ort.Value{nil}
	err = d.melspecSession.Run([]ort.Value{inputTensor}, outputs)
	if err != nil {
		return nil, fmt.Errorf("melspec inference failed: %w", err)
	}
	defer outputs[0].Destroy()

	// Get output shape and data
	outputShape := outputs[0].GetShape()
	// Output shape is [1, 1, frames, 32]
	if len(outputShape) != 4 {
		return nil, fmt.Errorf("unexpected melspec output shape: %v", outputShape)
	}
	numFrames := int(outputShape[2])

	// Get the data - need to cast to typed tensor
	outputTensor, ok := outputs[0].(*ort.Tensor[float32])
	if !ok {
		return nil, fmt.Errorf("melspec output is not float32 tensor")
	}
	output := outputTensor.GetData()

	// Reshape to array of [32] frames and apply transform: melspec / 10 + 2
	melspec := make([][]float32, numFrames)
	for i := 0; i < numFrames; i++ {
		frame := make([]float32, MelBins)
		for j := 0; j < MelBins; j++ {
			frame[j] = output[i*MelBins+j]/10 + 2
		}
		melspec[i] = frame
	}

	return melspec, nil
}

func (d *Detector) getEmbeddings(melspec [][]float32) ([][]float32, error) {
	// Create windows of 76 frames with step size 8
	var windows [][][]float32
	for i := 0; i <= len(melspec)-MelWindowSize; i += MelWindowStep {
		windows = append(windows, melspec[i:i+MelWindowSize])
	}

	if len(windows) == 0 {
		return nil, nil
	}

	batchSize := len(windows)

	// Flatten data for input tensor [batch, 76, 32, 1]
	flatData := make([]float32, batchSize*MelWindowSize*MelBins)
	idx := 0
	for b := 0; b < batchSize; b++ {
		for i := 0; i < MelWindowSize; i++ {
			for j := 0; j < MelBins; j++ {
				flatData[idx] = windows[b][i][j]
				idx++
			}
		}
	}

	// Create input tensor
	inputShape := ort.NewShape(int64(batchSize), MelWindowSize, MelBins, 1)
	inputTensor, err := ort.NewTensor(inputShape, flatData)
	if err != nil {
		return nil, fmt.Errorf("failed to create input tensor: %w", err)
	}
	defer inputTensor.Destroy()

	// Run inference with nil output to let onnxruntime allocate it
	outputs := []ort.Value{nil}
	err = d.embeddingSession.Run([]ort.Value{inputTensor}, outputs)
	if err != nil {
		return nil, fmt.Errorf("embedding inference failed: %w", err)
	}
	defer outputs[0].Destroy()

	// Get the data - need to cast to typed tensor
	outputTensor, ok := outputs[0].(*ort.Tensor[float32])
	if !ok {
		return nil, fmt.Errorf("embedding output is not float32 tensor")
	}
	output := outputTensor.GetData()
	outputShape := outputTensor.GetShape()

	// Output shape should be [batch, 96] or similar
	embSize := int(outputShape[len(outputShape)-1])

	// Reshape output to [batch, embSize]
	embeddings := make([][]float32, batchSize)
	for b := 0; b < batchSize; b++ {
		emb := make([]float32, embSize)
		for j := 0; j < embSize; j++ {
			emb[j] = output[b*embSize+j]
		}
		embeddings[b] = emb
	}

	return embeddings, nil
}

func (d *Detector) runWakeWordModel(session *ort.DynamicAdvancedSession, features [][]float32) (float32, error) {
	// Flatten data for input tensor [1, 16, 96]
	flatData := make([]float32, WakeWordFeatures*EmbeddingSize)
	idx := 0
	for i := 0; i < WakeWordFeatures; i++ {
		for j := 0; j < EmbeddingSize; j++ {
			flatData[idx] = features[i][j]
			idx++
		}
	}

	// Create input tensor
	inputShape := ort.NewShape(1, WakeWordFeatures, EmbeddingSize)
	inputTensor, err := ort.NewTensor(inputShape, flatData)
	if err != nil {
		return 0, fmt.Errorf("failed to create input tensor: %w", err)
	}
	defer inputTensor.Destroy()

	// Run inference with nil output to let onnxruntime allocate it
	outputs := []ort.Value{nil}
	err = session.Run([]ort.Value{inputTensor}, outputs)
	if err != nil {
		return 0, fmt.Errorf("wake word inference failed: %w", err)
	}
	defer outputs[0].Destroy()

	// Get the data
	outputTensor, ok := outputs[0].(*ort.Tensor[float32])
	if !ok {
		return 0, fmt.Errorf("wake word output is not float32 tensor")
	}

	return outputTensor.GetData()[0], nil
}

// Close releases all resources
// Close shuts down the ONNX sessions and frees allocated C memory.
func (d *Detector) Close() {
	if d.melspecSession != nil {
		d.melspecSession.Destroy()
	}
	if d.embeddingSession != nil {
		d.embeddingSession.Destroy()
	}
	for _, model := range d.wakeWordSessions {
		if model.session != nil {
			model.session.Destroy()
		}
	}
	// NOTE: Do NOT call ort.DestroyEnvironment() here - it's global and shared
}
