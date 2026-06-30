// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

// Package logging provides a simple configurable logger using slog.
package logging

import (
	"log/slog"
	"os"
	"strings"
)

// Setup configures the default slog logger based on level and format.
// level: debug, info, warn, error
// format: console, json
func Setup(level, format string) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{Level: parseLevel(level)}

	if strings.ToLower(format) == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}

// parseLevel converts a level string to the corresponding slog.Level.
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
