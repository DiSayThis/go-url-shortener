package logging

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func NewLogger(environment string) *slog.Logger {
	if environment == "local" {
		return slog.New(
			tint.NewHandler(
				os.Stdout,
				&tint.Options{
					Level:      slog.LevelDebug,
					AddSource:  true,
					TimeFormat: time.RFC3339,
				},
			),
		).With(
			"service", "go-api",
		)
	}

	// В production оставляем JSON без цветов.
	return slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level:     slog.LevelInfo,
				AddSource: true,
			},
		),
	)
}
