package logging

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

func Set() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}
