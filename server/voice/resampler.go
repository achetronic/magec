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

// Resampler converts audio from one sample rate to another using averaging
// Based on the ricky0123/vad resampler approach
type Resampler struct {
	nativeSampleRate int
	targetSampleRate int
	targetFrameSize  int
	inputBuffer      []float32
}

// NewResampler creates a new resampler
func NewResampler(nativeSampleRate, targetSampleRate, targetFrameSize int) *Resampler {
	return &Resampler{
		nativeSampleRate: nativeSampleRate,
		targetSampleRate: targetSampleRate,
		targetFrameSize:  targetFrameSize,
		inputBuffer:      make([]float32, 0),
	}
}

// Process resamples the input audio and returns frames of targetFrameSize
func (r *Resampler) Process(input []float32) [][]float32 {
	// Add input to buffer
	r.inputBuffer = append(r.inputBuffer, input...)

	// Calculate how many input samples we need for one output frame
	inputSamplesPerFrame := float64(r.nativeSampleRate) / float64(r.targetSampleRate) * float64(r.targetFrameSize)

	var frames [][]float32

	// Process while we have enough samples
	for float64(len(r.inputBuffer)) >= inputSamplesPerFrame {
		frame := r.resampleFrame(int(inputSamplesPerFrame))
		frames = append(frames, frame)

		// Remove processed samples from buffer
		samplesToRemove := int(inputSamplesPerFrame)
		if samplesToRemove > len(r.inputBuffer) {
			samplesToRemove = len(r.inputBuffer)
		}
		r.inputBuffer = r.inputBuffer[samplesToRemove:]
	}

	return frames
}

// resampleFrame resamples inputSamples samples to targetFrameSize using averaging
func (r *Resampler) resampleFrame(inputSamples int) []float32 {
	output := make([]float32, r.targetFrameSize)
	ratio := float64(inputSamples) / float64(r.targetFrameSize)

	for i := 0; i < r.targetFrameSize; i++ {
		// Calculate the range of input samples that contribute to this output sample
		start := float64(i) * ratio
		end := float64(i+1) * ratio

		startIdx := int(start)
		endIdx := int(end)
		if endIdx >= inputSamples {
			endIdx = inputSamples - 1
		}

		// Average the samples in this range
		var sum float32
		count := 0
		for j := startIdx; j <= endIdx && j < len(r.inputBuffer); j++ {
			sum += r.inputBuffer[j]
			count++
		}

		if count > 0 {
			output[i] = sum / float32(count)
		}
	}

	return output
}

// ProcessInt16 processes int16 input and returns int16 output frames
func (r *Resampler) ProcessInt16(input []int16) [][]int16 {
	// Convert int16 to float32
	floatInput := make([]float32, len(input))
	for i, s := range input {
		floatInput[i] = float32(s) / 32768.0
	}

	// Resample
	floatFrames := r.Process(floatInput)

	// Convert back to int16
	int16Frames := make([][]int16, len(floatFrames))
	for i, frame := range floatFrames {
		int16Frame := make([]int16, len(frame))
		for j, s := range frame {
			val := int(s * 32767)
			if val > 32767 {
				val = 32767
			} else if val < -32768 {
				val = -32768
			}
			int16Frame[j] = int16(val)
		}
		int16Frames[i] = int16Frame
	}

	return int16Frames
}

// Reset clears the internal buffer
func (r *Resampler) Reset() {
	r.inputBuffer = r.inputBuffer[:0]
}
