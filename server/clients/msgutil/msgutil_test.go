// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package msgutil

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestValidateInputLength_Short(t *testing.T) {
	text := "hello world"
	result, truncated := ValidateInputLength(text, 100)
	if truncated {
		t.Error("expected no truncation")
	}
	if result != text {
		t.Errorf("expected %q, got %q", text, result)
	}
}

func TestValidateInputLength_Exact(t *testing.T) {
	text := strings.Repeat("a", 100)
	result, truncated := ValidateInputLength(text, 100)
	if truncated {
		t.Error("expected no truncation")
	}
	if result != text {
		t.Errorf("expected same text")
	}
}

func TestValidateInputLength_Truncated(t *testing.T) {
	text := strings.Repeat("a", 200)
	result, truncated := ValidateInputLength(text, 100)
	if !truncated {
		t.Error("expected truncation")
	}
	if !strings.HasSuffix(result, "[message truncated]") {
		t.Error("expected truncation suffix")
	}
	runes := []rune(result)
	if len(runes) > 100+len("\n\n[message truncated]") {
		t.Errorf("truncated result too long: %d runes", len(runes))
	}
}

func TestValidateInputLength_Unicode(t *testing.T) {
	text := strings.Repeat("日", 50)
	result, truncated := ValidateInputLength(text, 30)
	if !truncated {
		t.Error("expected truncation")
	}
	if !strings.HasPrefix(result, strings.Repeat("日", 30)) {
		t.Error("expected unicode-safe truncation")
	}
}

func TestSplitMessage_Short(t *testing.T) {
	text := "hello"
	chunks := SplitMessage(text, 100)
	if len(chunks) != 1 {
		t.Errorf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0] != text {
		t.Errorf("expected %q, got %q", text, chunks[0])
	}
}

func TestSplitMessage_ParagraphBoundary(t *testing.T) {
	part1 := strings.Repeat("a", 40)
	part2 := strings.Repeat("b", 40)
	text := part1 + "\n\n" + part2
	chunks := SplitMessage(text, 50)
	if len(chunks) != 2 {
		t.Errorf("expected 2 chunks, got %d: %v", len(chunks), chunks)
	}
	if chunks[0] != part1 {
		t.Errorf("first chunk: expected %q, got %q", part1, chunks[0])
	}
	if chunks[1] != part2 {
		t.Errorf("second chunk: expected %q, got %q", part2, chunks[1])
	}
}

func TestSplitMessage_LineBoundary(t *testing.T) {
	part1 := strings.Repeat("a", 40)
	part2 := strings.Repeat("b", 40)
	text := part1 + "\n" + part2
	chunks := SplitMessage(text, 50)
	if len(chunks) != 2 {
		t.Errorf("expected 2 chunks, got %d: %v", len(chunks), chunks)
	}
}

func TestSplitMessage_WordBoundary(t *testing.T) {
	text := "hello world this is a test"
	chunks := SplitMessage(text, 15)
	for _, chunk := range chunks {
		if utf8.RuneCountInString(chunk) > 15 {
			t.Errorf("chunk exceeds max: %q (%d runes)", chunk, utf8.RuneCountInString(chunk))
		}
	}
	joined := strings.Join(chunks, " ")
	if !strings.Contains(joined, "hello") || !strings.Contains(joined, "test") {
		t.Error("content lost during split")
	}
}

func TestSplitMessage_HardCut(t *testing.T) {
	text := strings.Repeat("a", 100)
	chunks := SplitMessage(text, 30)
	for _, chunk := range chunks {
		if utf8.RuneCountInString(chunk) > 30 {
			t.Errorf("chunk exceeds max: %d runes", utf8.RuneCountInString(chunk))
		}
	}
	total := 0
	for _, c := range chunks {
		total += utf8.RuneCountInString(c)
	}
	if total != 100 {
		t.Errorf("content lost: expected 100 runes, got %d", total)
	}
}

func TestSplitMessage_TelegramLimit(t *testing.T) {
	text := strings.Repeat("word ", 2000) // ~10000 chars
	chunks := SplitMessage(text, TelegramMaxMessageLength)
	for i, chunk := range chunks {
		if utf8.RuneCountInString(chunk) > TelegramMaxMessageLength {
			t.Errorf("chunk %d exceeds Telegram limit: %d runes", i, utf8.RuneCountInString(chunk))
		}
	}
}

func TestSplitMessage_EmptyString(t *testing.T) {
	chunks := SplitMessage("", 100)
	if len(chunks) != 1 || chunks[0] != "" {
		t.Errorf("expected single empty chunk, got %v", chunks)
	}
}

func TestSplitMessage_Unicode(t *testing.T) {
	text := strings.Repeat("日本語", 100) // 300 unicode chars
	chunks := SplitMessage(text, 50)
	for i, chunk := range chunks {
		if utf8.RuneCountInString(chunk) > 50 {
			t.Errorf("chunk %d exceeds max: %d runes", i, utf8.RuneCountInString(chunk))
		}
	}
}

func TestSplitMessage_NoContentLoss(t *testing.T) {
	part1 := "El comercio atlántico transformó las islas."
	part2 := "Los aborígenes amazigh dejaron su huella."
	text := part1 + "\n\n" + part2
	chunks := SplitMessage(text, 50)

	reconstructed := strings.Join(chunks, "\n\n")
	if reconstructed != text {
		t.Errorf("content mismatch:\noriginal:      %q\nreconstructed: %q", text, reconstructed)
	}
}

func TestSplitMessage_LongParagraphsNoContentLoss(t *testing.T) {
	var paragraphs []string
	for i := 0; i < 5; i++ {
		paragraphs = append(paragraphs, strings.Repeat("abcdefghij ", 400)) // ~4400 chars each
	}
	text := strings.Join(paragraphs, "\n\n")

	chunks := SplitMessage(text, TelegramMaxMessageLength)

	var totalRunes int
	for i, chunk := range chunks {
		runeCount := utf8.RuneCountInString(chunk)
		if runeCount > TelegramMaxMessageLength {
			t.Errorf("chunk %d exceeds limit: %d runes", i, runeCount)
		}
		totalRunes += runeCount
	}

	originalRunes := utf8.RuneCountInString(text)
	separatorRunes := 0
	for i := 0; i < len(chunks)-1; i++ {
		separatorRunes += utf8.RuneCountInString("\n\n")
	}

	if totalRunes+separatorRunes < originalRunes-10 {
		t.Errorf("significant content loss: original %d runes, reconstructed %d runes (with %d separator runes)",
			originalRunes, totalRunes, separatorRunes)
	}
}

func TestSplitMessage_SpanishTextNoTruncation(t *testing.T) {
	text := `PRIMERA PARTE: LOS ORÍGENES Y LOS PUEBLOS ABORÍGENES

Las Islas Canarias, ese archipiélago perdido en el Atlántico frente a las costas africanas, tienen una historia que se remonta a miles de años antes de que los europeos pusieran sus sucias botas conquistadoras en estas tierras.

SEGUNDA PARTE: LA LLEGADA DE LOS EUROPEOS Y LA CONQUISTA

Los europeos "redescubrieron" las Canarias en el siglo XIV, aunque hay referencias anteriores en textos clásicos.

TERCERA PARTE: LA CANARIAS COLONIAL Y EL COMERCIO ATLÁNTICO

Tras la conquista, las islas se reorganizaron completamente. El comercio atlántico fue clave para su desarrollo económico durante siglos.`

	chunks := SplitMessage(text, 200)

	reconstructed := ""
	for i, chunk := range chunks {
		if i > 0 {
			reconstructed += "\n\n"
		}
		reconstructed += chunk
	}

	if !strings.Contains(reconstructed, "comercio atlántico") {
		t.Error("lost 'comercio atlántico' during split")
	}
	if !strings.Contains(reconstructed, "PRIMERA") {
		t.Error("lost 'PRIMERA' during split")
	}
	if !strings.Contains(reconstructed, "TERCERA") {
		t.Error("lost 'TERCERA' during split")
	}

	for i, chunk := range chunks {
		if utf8.RuneCountInString(chunk) > 200 {
			t.Errorf("chunk %d exceeds limit: %d runes: %q", i, utf8.RuneCountInString(chunk), chunk[:50])
		}
		if strings.HasPrefix(chunk, "omercio") || strings.HasPrefix(chunk, "s de los") {
			t.Errorf("chunk %d starts with truncated word: %q", i, chunk[:20])
		}
	}
}
