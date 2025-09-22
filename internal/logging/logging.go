package logging

import (
	"fmt"
	"io"
	stdlog "log"
	"os"
	"strings"
	"sync"
)

const (
	InfoLevel = iota
	WarnLevel
	ErrorLevel
	VerboseLevel
	DebugLevel
)

var (
	mu     sync.RWMutex
	level  = InfoLevel
	logger = stdlog.New(os.Stdout, "", stdlog.LstdFlags)
)

// SetLevel sets the current logging level from a string: "info", "warn", "error", "verbose", "debug".
// Unrecognized values default to info.
func SetLevel(s string) {
	mu.Lock()
	defer mu.Unlock()
	switch strings.ToLower(s) {
	case "debug":
		level = DebugLevel
	case "verbose", "v":
		level = VerboseLevel
	case "error":
		level = ErrorLevel
	case "warn", "warning":
		level = WarnLevel
	default:
		level = InfoLevel
	}
}

func levelEnabled(l int) bool {
	mu.RLock()
	defer mu.RUnlock()
	return level >= l
}

func Debugf(format string, v ...interface{}) {
	if levelEnabled(DebugLevel) {
		logger.Printf("[DEBUG] "+format, v...)
	}
}

func Verbosef(format string, v ...interface{}) {
	if levelEnabled(VerboseLevel) {
		logger.Printf("[VERBOSE] "+format, v...)
	}
}

func Infof(format string, v ...interface{}) {
	if levelEnabled(InfoLevel) {
		logger.Printf("[INFO] "+format, v...)
	}
}

func Warnf(format string, v ...interface{}) {
	if levelEnabled(WarnLevel) {
		logger.Printf("[WARN] "+format, v...)
	}
}

func Errorf(format string, v ...interface{}) {
	if levelEnabled(ErrorLevel) {
		logger.Printf("[ERROR] "+format, v...)
	}
}

func Printf(format string, v ...interface{}) {
	Infof(format, v...)
}

func Println(v ...interface{}) {
	Infof(fmt.Sprintln(v...))
}

func Fatalf(format string, v ...interface{}) {
	logger.Fatalf("[FATAL] "+format, v...)
}

func Fatal(v ...interface{}) {
	logger.Fatal(v...)
}

func Panicf(format string, v ...interface{}) {
	logger.Panicf(format, v...)
}

// SetOutput configures the underlying logger output (e.g., MultiWriter).
func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	logger.SetOutput(w)
}

// SetFileOutput sets up logging to both stdout and a file.
// If logFile is empty, only stdout is used.
func SetFileOutput(logFile string) (*os.File, error) {
	if logFile == "" {
		return nil, nil
	}

	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", logFile, err)
	}

	mw := io.MultiWriter(os.Stdout, f)
	SetOutput(mw)
	return f, nil
}

// SetFlags proxies to the underlying logger's SetFlags.
func SetFlags(flags int) {
	mu.Lock()
	defer mu.Unlock()
	logger.SetFlags(flags)
}

// SetPrefix proxies to the underlying logger's SetPrefix.
func SetPrefix(p string) {
	mu.Lock()
	defer mu.Unlock()
	logger.SetPrefix(p)
}
