package logs

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"
)

type Logger interface {
	SetEnabled(enabled bool)
	Printf(format string, args ...any)
	Clean()
}

type logger struct {
	enabled atomic.Bool
	file    *os.File
	logger  *log.Logger
}

func New() *logger {
	path := logPath()
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return &logger{
			logger: log.New(io.Discard, "", log.LstdFlags),
		}
	}
	return &logger{
		file:   file,
		logger: log.New(file, "", log.LstdFlags),
	}
}

func logPath() string {
	if dir, err := os.UserCacheDir(); err == nil {
		return filepath.Join(dir, "cadence.log")
	}
	return filepath.Join(os.TempDir(), "cadence.log")
}

func (l *logger) SetEnabled(enabled bool) {
	if l == nil {
		return
	}
	l.enabled.Store(enabled)
	if l.logger == nil {
		return
	}
	if enabled {
		if l.file != nil {
			l.logger.SetOutput(l.file)
		}
		return
	}
	l.logger.SetOutput(io.Discard)
}

func (l *logger) Printf(format string, args ...any) {
	if l == nil || !l.enabled.Load() || l.logger == nil {
		return
	}
	l.logger.Printf(format, args...)
}

func (l *logger) Clean() {
	if l == nil || l.file == nil {
		return
	}
	_ = l.file.Close()
}
