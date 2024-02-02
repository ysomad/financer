package slogx

import (
	"log/slog"
	"os"
)

func Fatal(msg string, err error) {
	slog.Error(msg, "err", err.Error())
	os.Exit(1)
}
