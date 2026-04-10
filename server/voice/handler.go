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
	"encoding/binary"
	"encoding/json"
	"log/slog"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// Message types for WebSocket communication
const (
	MsgTypeCapabilities = "capabilities"
	MsgTypeConfig       = "config"
	MsgTypeSetModel     = "setModel"
	MsgTypeWakeword     = "wakeword"
	MsgTypeSpeechStart  = "speech_start"
	MsgTypeSpeechEnd    = "speech_end"
	MsgTypeError        = "error"
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// AudioConfig is sent by the client to configure audio processing
type AudioConfig struct {
	SampleRate int    `json:"sampleRate"`
	Model      string `json:"model"`
}

// WakewordModelInfo represents a wake word model for client
type WakewordModelInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Phrase string `json:"phrase"`
}

// WakewordsCapabilities represents wake word detection capabilities
type WakewordsCapabilities struct {
	Models []WakewordModelInfo `json:"models"`
	Active string              `json:"active"`
}

// VADCapabilities represents voice activity detection capabilities
type VADCapabilities struct {
	Enabled        bool `json:"enabled"`
	SilenceTimeout int  `json:"silenceTimeout"` // milliseconds
}

// Capabilities represents the full capabilities of the voice-events endpoint
type Capabilities struct {
	Wakewords WakewordsCapabilities `json:"wakewords"`
	VAD       VADCapabilities       `json:"vad"`
}

// Handler manages WebSocket connections for voice event detection
type Handler struct {
	logger         *slog.Logger
	detectorConfig DetectorConfig

	// Track active connections
	connections map[*websocket.Conn]*clientState
	mu          sync.Mutex
}

type clientState struct {
	detector   *Detector
	vad        *VAD
	resampler  *Resampler
	sampleRate int
	vadEnabled bool
	connMu     sync.Mutex // Mutex for writing to this connection
}

// NewHandler creates a new WebSocket handler for voice event detection
func NewHandler(detector *Detector, logger *slog.Logger) *Handler {
	return &Handler{
		logger:         logger,
		detectorConfig: detector.config,
		connections:    make(map[*websocket.Conn]*clientState),
	}
}

// ServeHTTP handles WebSocket upgrade and message processing
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("WebSocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	// Create a new detector for this connection
	detector := NewDetector(h.detectorConfig, h.logger)
	if err := detector.Load(); err != nil {
		h.logger.Error("Failed to create detector for connection", "error", err, "remote", r.RemoteAddr)
		conn.WriteJSON(WSMessage{
			Type: MsgTypeError,
			Data: map[string]string{"message": "Failed to initialize voice detector"},
		})
		return
	}

	// Create VAD if model data is configured
	var vad *VAD
	vadEnabled := len(h.detectorConfig.VADModelData) > 0
	if vadEnabled {
		vad = NewVAD(VADConfig{
			ModelData:      h.detectorConfig.VADModelData,
			Threshold:      VADThreshold,
			SilenceTimeout: VADSilenceTimeout,
		}, h.logger)
		if err := vad.Load(); err != nil {
			h.logger.Warn("Failed to load VAD model, continuing without VAD", "error", err)
			vad = nil
			vadEnabled = false
		}
	}

	// Register connection
	h.mu.Lock()
	state := &clientState{
		detector:   detector,
		vad:        vad,
		vadEnabled: vadEnabled,
	}
	h.connections[conn] = state
	h.mu.Unlock()

	h.logger.Info("Voice events WebSocket connected", "remote", r.RemoteAddr, "activeConnections", len(h.connections), "vadEnabled", vadEnabled)

	defer func() {
		h.mu.Lock()
		delete(h.connections, conn)
		h.mu.Unlock()
		detector.Close()
		if vad != nil {
			vad.Close()
		}
		h.logger.Info("Voice events WebSocket disconnected", "remote", r.RemoteAddr)
	}()

	// Set up wake word detection callback
	detector.SetOnDetected(func(modelID string) {
		state.connMu.Lock()
		defer state.connMu.Unlock()
		msg := WSMessage{
			Type: MsgTypeWakeword,
			Data: map[string]string{"model": modelID},
		}
		if err := conn.WriteJSON(msg); err != nil {
			h.logger.Error("Failed to send wakeword message", "error", err)
		}
	})

	// Set up VAD callbacks
	if vad != nil {
		vad.SetOnSpeechStart(func() {
			state.connMu.Lock()
			defer state.connMu.Unlock()
			msg := WSMessage{
				Type: MsgTypeSpeechStart,
			}
			if err := conn.WriteJSON(msg); err != nil {
				h.logger.Error("Failed to send speech_start message", "error", err)
			}
		})
		vad.SetOnSpeechEnd(func() {
			state.connMu.Lock()
			defer state.connMu.Unlock()
			msg := WSMessage{
				Type: MsgTypeSpeechEnd,
			}
			if err := conn.WriteJSON(msg); err != nil {
				h.logger.Error("Failed to send speech_end message", "error", err)
			}
		})
	}

	// Send capabilities on connect
	h.sendCapabilities(conn, detector, vad)

	// Message processing loop
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("WebSocket read error", "error", err)
			}
			break
		}

		switch messageType {
		case websocket.TextMessage:
			h.handleTextMessage(conn, data)
		case websocket.BinaryMessage:
			h.handleBinaryMessage(conn, data)
		}
	}
}

func (h *Handler) sendCapabilities(conn *websocket.Conn, detector *Detector, vad *VAD) {
	models := detector.GetModels()
	modelInfos := make([]WakewordModelInfo, len(models))
	for i, m := range models {
		modelInfos[i] = WakewordModelInfo{
			ID:     m.ID,
			Name:   m.Name,
			Phrase: m.Phrase,
		}
	}

	vadCaps := VADCapabilities{
		Enabled:        vad != nil,
		SilenceTimeout: int(VADSilenceTimeout / time.Millisecond),
	}

	msg := WSMessage{
		Type: MsgTypeCapabilities,
		Data: Capabilities{
			Wakewords: WakewordsCapabilities{
				Models: modelInfos,
				Active: detector.GetActiveModel(),
			},
			VAD: vadCaps,
		},
	}

	if err := conn.WriteJSON(msg); err != nil {
		h.logger.Error("Failed to send capabilities", "error", err)
	}
}

func (h *Handler) handleTextMessage(conn *websocket.Conn, data []byte) {
	var msg WSMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		h.logger.Error("Failed to parse WebSocket message", "error", err)
		return
	}

	switch msg.Type {
	case MsgTypeConfig:
		h.handleConfig(conn, msg.Data)
	case MsgTypeSetModel:
		h.handleSetModel(conn, msg.Data)
	}
}

func (h *Handler) handleConfig(conn *websocket.Conn, data interface{}) {
	// Parse config from data
	configBytes, err := json.Marshal(data)
	if err != nil {
		h.logger.Error("Failed to marshal config", "error", err)
		return
	}

	var config AudioConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		h.logger.Error("Failed to parse audio config", "error", err)
		return
	}

	h.mu.Lock()
	state := h.connections[conn]
	if state != nil {
		state.sampleRate = config.SampleRate
		// Create resampler if needed
		if config.SampleRate != TargetSampleRate {
			state.resampler = NewResampler(config.SampleRate, TargetSampleRate, 1280)
			h.logger.Debug("Created resampler",
				"from", config.SampleRate,
				"to", TargetSampleRate,
			)
		} else {
			state.resampler = nil
		}

		// Set active model if specified
		if config.Model != "" {
			if err := state.detector.SetActiveModel(config.Model); err != nil {
				h.logger.Warn("Failed to set model", "model", config.Model, "error", err)
			}
		}
	}
	h.mu.Unlock()

	h.logger.Debug("Audio config received",
		"sampleRate", config.SampleRate,
		"model", config.Model,
	)
}

func (h *Handler) handleSetModel(conn *websocket.Conn, data interface{}) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	modelID, ok := dataMap["model"].(string)
	if !ok {
		return
	}

	h.mu.Lock()
	state := h.connections[conn]
	h.mu.Unlock()

	if state == nil {
		return
	}

	if err := state.detector.SetActiveModel(modelID); err != nil {
		h.sendError(conn, err.Error())
		return
	}

	// Confirm model change by sending updated capabilities
	h.sendCapabilities(conn, state.detector, state.vad)
}

func (h *Handler) sendError(conn *websocket.Conn, message string) {
	msg := WSMessage{
		Type: MsgTypeError,
		Data: map[string]string{"message": message},
	}
	conn.WriteJSON(msg)
}

func (h *Handler) handleBinaryMessage(conn *websocket.Conn, data []byte) {
	h.mu.Lock()
	state := h.connections[conn]
	h.mu.Unlock()

	if state == nil {
		return
	}

	// Parse audio data as float32 little-endian
	numSamples := len(data) / 4
	samples := make([]float32, numSamples)
	for i := 0; i < numSamples; i++ {
		bits := binary.LittleEndian.Uint32(data[i*4:])
		samples[i] = math.Float32frombits(bits)
	}

	// Process audio (resample if needed)
	if state.resampler != nil {
		frames := state.resampler.Process(samples)
		for _, frame := range frames {
			// Process wake word detection
			int16Samples := floatToInt16(frame)
			if err := state.detector.ProcessAudio(int16Samples); err != nil {
				h.logger.Error("Wake word processing error", "error", err)
			}

			// Process VAD (uses float32)
			if state.vad != nil {
				if err := state.vad.ProcessAudio(frame); err != nil {
					h.logger.Error("VAD processing error", "error", err)
				}
			}
		}
	} else {
		// Already at target sample rate
		if err := state.detector.ProcessAudioFloat32(samples); err != nil {
			h.logger.Error("Wake word processing error", "error", err)
		}

		// Process VAD
		if state.vad != nil {
			if err := state.vad.ProcessAudio(samples); err != nil {
				h.logger.Error("VAD processing error", "error", err)
			}
		}
	}
}

// floatToInt16 converts float32 samples to int16
func floatToInt16(samples []float32) []int16 {
	result := make([]int16, len(samples))
	for i, s := range samples {
		val := int(s * 32767)
		if val > 32767 {
			val = 32767
		} else if val < -32768 {
			val = -32768
		}
		result[i] = int16(val)
	}
	return result
}
