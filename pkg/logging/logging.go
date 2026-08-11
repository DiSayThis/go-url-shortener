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
			tint.NewTextHandler(
				os.Stdout,
				&tint.Options{
					Level:      slog.LevelDebug,
					AddSource:  true,
					TimeFormat: time.RFC3339,
					ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
						if a.Value.Kind() == slog.KindAny {
							if _, ok := a.Value.Any().(error); ok {
								return tint.Attr(9, a)
							}
						}
						return a
					},
				},
			),
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
