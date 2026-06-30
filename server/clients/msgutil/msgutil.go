// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package msgutil

import (
	"strings"
	"unicode/utf8"
)

const (
	TelegramMaxMessageLength = 4096
	DiscordMaxMessageLength  = 2000
	SlackMaxMessageLength    = 39000

	DefaultMaxInputLength = 16000
)

// ValidateInputLength checks whether the input message exceeds maxLen.
// Returns the (possibly truncated) message and true if it was truncated.
func ValidateInputLength(text string, maxLen int) (string, bool) {
	if utf8.RuneCountInString(text) <= maxLen {
		return text, false
	}
	runes := []rune(text)
	return string(runes[:maxLen]) + "\n\n[message truncated]", true
}

// SplitMessage splits a message into chunks that respect maxLen per chunk.
// It tries to split at paragraph boundaries (\n\n), then line boundaries (\n),
// then word boundaries (space), falling back to hard cuts.
// All chunks are guaranteed to be non-empty and within maxLen runes.
func SplitMessage(text string, maxLen int) []string {
	if maxLen <= 0 {
		maxLen = TelegramMaxMessageLength
	}

	if utf8.RuneCountInString(text) <= maxLen {
		return []string{text}
	}

	var chunks []string
	remaining := text

	for remaining != "" {
		if utf8.RuneCountInString(remaining) <= maxLen {
			chunks = append(chunks, remaining)
			break
		}

		runes := []rune(remaining)
		candidate := string(runes[:maxLen])

		splitByteIdx := findSplitPoint(candidate)
		chunk := strings.TrimRight(candidate[:splitByteIdx], " \n")
		if chunk == "" {
			chunk = candidate
			splitByteIdx = len(candidate)
		}

		chunks = append(chunks, chunk)

		remaining = remaining[splitByteIdx:]
		remaining = strings.TrimLeft(remaining, "\n")
	}

	return chunks
}

// findSplitPoint finds the best byte index within candidate to split at.
// Prefers paragraph breaks, then line breaks, then word boundaries.
func findSplitPoint(candidate string) int {
	if idx := strings.LastIndex(candidate, "\n\n"); idx > 0 {
		return idx
	}
	if idx := strings.LastIndex(candidate, "\n"); idx > 0 {
		return idx
	}
	if idx := strings.LastIndex(candidate, " "); idx > 0 {
		return idx
	}
	return len(candidate)
}
