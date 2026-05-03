package conf

import (
	"io"
	"log/slog"
	"os"
)

func ConfigureLogging(s *Settings) {
	var level slog.Level
	switch s.Logging.Level {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN", "WARNING":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: level}

	switch s.Logging.Format {
	case "json":
		handler = slog.NewJSONHandler(resolveOutput(s.Logging.OutputPath), opts)
	default:
		handler = slog.NewTextHandler(resolveOutput(s.Logging.OutputPath), opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func resolveOutput(path string) io.Writer {
	switch path {
	case "stdout":
		return os.Stdout
	case "stderr", "":
		return os.Stderr
	default:
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return os.Stderr
		}
		return f
	}
}