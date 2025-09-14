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
	VerboseLevel
	DebugLevel
)

var (
	mu     sync.RWMutex
	level  = InfoLevel
	logger = stdlog.New(os.Stdout, "", stdlog.LstdFlags)
)

// SetLevel sets the current logging level from a string: "info", "verbose", "debug".
// Unrecognized values default to info.
func SetLevel(s string) {
	mu.Lock()
	defer mu.Unlock()
	switch strings.ToLower(s) {
	case "debug":
		level = DebugLevel
	case "verbose", "v":
		level = VerboseLevel
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
