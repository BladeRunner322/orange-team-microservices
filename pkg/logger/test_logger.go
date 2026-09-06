package logger

import (
	"io"
	"log/slog"
)

// NewTestLogger возвращает логгер, который ничего не выводит (для тестов)
func NewTestLogger() *Logger {
	handler := slog.NewTextHandler(io.Discard, nil)
	return &Logger{Logger: slog.New(handler)}
}
