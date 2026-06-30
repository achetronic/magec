// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

// Package msgutil exposes shared helpers used by every chat client.
//
// The attachment-related helpers implement the inline-vs-artifact policy
// for user-uploaded files: small files ride inline in the /run_sse request
// body as inlineData parts, while larger files are persisted through the
// ADK artifact service and referenced from the prompt so the LLM can call
// load_artifact on demand. See decision #24 in .agents/DECISIONS.md.

package msgutil

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/adk/artifact"
	"google.golang.org/genai"
)

// StoreAsArtifact persists the file through svc under the given filename and
// returns a single human+machine readable line describing it. The line is
// meant to be collected with others and wrapped by AttachedArtifactsBlock
// before being appended to the user prompt.
//
// Filename must be unique per session; callers typically prefix it with a
// message identifier to avoid clashes. svc must not be nil.
func StoreAsArtifact(
	ctx context.Context,
	svc artifact.Service,
	appName, userID, sessionID, filename, mimeType string,
	data []byte,
) (string, error) {
	if svc == nil {
		return "", fmt.Errorf("artifact service is nil")
	}
	if filename == "" {
		return "", fmt.Errorf("filename is required")
	}
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	part := &genai.Part{
		InlineData: &genai.Blob{
			MIMEType: mimeType,
			Data:     data,
		},
	}
	if _, err := svc.Save(ctx, &artifact.SaveRequest{
		AppName:   appName,
		UserID:    userID,
		SessionID: sessionID,
		FileName:  filename,
		Part:      part,
	}); err != nil {
		return "", fmt.Errorf("save artifact %q: %w", filename, err)
	}

	return fmt.Sprintf("- %s (%s, %s)", filename, mimeType, humanSize(len(data))), nil
}

// AttachedArtifactsBlock wraps a list of per-file descriptor lines in the
// MAGEC_ATTACHED_ARTIFACTS HTML-comment block that the prompt consumes.
// Returns the empty string when lines is empty, so callers can append
// unconditionally.
func AttachedArtifactsBlock(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n\n<!--MAGEC_ATTACHED_ARTIFACTS:\n")
	b.WriteString("The user attached the following files to this message. They were saved to the session artifact store so they do not consume context tokens upfront.\n")
	for _, l := range lines {
		b.WriteString(l)
		b.WriteByte('\n')
	}
	b.WriteString("Call the load_artifact tool with the exact filename when you need to read its contents. The artifact will arrive on the next turn as a native multimodal attachment — do not attempt to decode base64 yourself.\n")
	b.WriteString(":MAGEC_ATTACHED_ARTIFACTS-->")
	return b.String()
}

// humanSize renders a byte count as a compact human-readable string.
func humanSize(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := int64(bytes) / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	suffixes := []string{"KB", "MB", "GB", "TB"}
	if exp >= len(suffixes) {
		exp = len(suffixes) - 1
	}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), suffixes[exp])
}
