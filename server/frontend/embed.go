// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package frontend

import (
	"embed"
	"io/fs"
)

//go:embed admin-ui
var adminUI embed.FS

//go:embed voice-ui
var voiceUI embed.FS

// AdminUI returns a filesystem rooted at the admin-ui build output.
func AdminUI() (fs.FS, error) {
	return fs.Sub(adminUI, "admin-ui")
}

// VoiceUI returns a filesystem rooted at the voice-ui build output.
func VoiceUI() (fs.FS, error) {
	return fs.Sub(voiceUI, "voice-ui")
}
