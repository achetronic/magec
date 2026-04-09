package voice

import (
	"context"
	"io"

	"github.com/achetronic/magec/server/store"
)

// Schema is a JSON Schema object represented as a plain map so providers can
// define arbitrarily complex schemas without the framework imposing structural
// limits. Same type alias used by clients and memory registries.
type Schema = map[string]interface{}

// TTSRequest holds the parameters for a text-to-speech synthesis call.
// Common fields (Model, Voice) come from the agent's TTSRef in the store.
// TTSConfig holds the full typed config; each provider reads its own section.
type TTSRequest struct {
	Input          string
	Model          string
	Voice          string
	ResponseFormat string
	Config         store.TTSConfig
}

// TTSResponse holds the result of a TTS synthesis call.
type TTSResponse struct {
	Audio       io.ReadCloser
	ContentType string
}

// STTRequest holds the parameters for a speech-to-text transcription call.
type STTRequest struct {
	Audio       io.Reader
	ContentType string
	Model       string
	Config      store.STTConfig
}

// Provider defines what every voice backend type must implement.
// To add a new voice provider (e.g. xAI):
//  1. Create a new package under server/voice/<name>/
//  2. Implement the Provider interface
//  3. Call voice.Register() in an init() function
//  4. Add a blank import in main.go
//  5. Add typed config structs to store/types.go (TTSConfig, STTConfig)
type Provider interface {
	// Type returns the backend type this provider handles (e.g. "openai", "gemini").
	Type() string

	// DisplayName returns a human-readable name for the admin UI.
	DisplayName() string

	// SupportsTTS returns true if this provider can synthesize speech.
	SupportsTTS() bool

	// SupportsSTT returns true if this provider can transcribe speech.
	SupportsSTT() bool

	// TTSConfigSchema returns the JSON Schema for provider-specific TTS
	// configuration fields. These are rendered dynamically in the admin UI
	// alongside the common fields (model, voice).
	// Return nil if no extra config is needed.
	TTSConfigSchema() Schema

	// STTConfigSchema returns the JSON Schema for provider-specific STT
	// configuration fields. Return nil if no extra config is needed.
	STTConfigSchema() Schema

	// SynthesizeSpeech converts text to audio.
	SynthesizeSpeech(ctx context.Context, req TTSRequest, backend store.BackendDefinition) (*TTSResponse, error)

	// TranscribeAudio converts audio to text.
	TranscribeAudio(ctx context.Context, req STTRequest, backend store.BackendDefinition) (string, error)
}
