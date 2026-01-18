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
