package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Level represents logging severity levels.
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging with file output.
type Logger struct {
	mu      sync.Mutex
	level   Level
	service string
	logger  *log.Logger
	file    *os.File
	isMulti bool
}

// Config holds logger configuration.
type Config struct {
	ServiceName string
	LogDir      string
	LogFileName string
	Level       Level
	ToStdout    bool
}

// DefaultConfig returns a default logger configuration.
func DefaultConfig(serviceName string) Config {
	return Config{
		ServiceName: serviceName,
		LogDir:      ".",
		LogFileName: "access.log",
		Level:       INFO,
		ToStdout:    true,
	}
}

// New creates a new Logger instance with the given configuration.
func New(cfg Config) (*Logger, error) {
	if cfg.LogFileName == "" {
		cfg.LogFileName = "access.log"
	}
	if cfg.LogDir == "" {
		cfg.LogDir = "."
	}

	// Ensure log directory exists
	if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory %s: %w", cfg.LogDir, err)
	}

	logPath := filepath.Join(cfg.LogDir, cfg.LogFileName)
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", logPath, err)
	}

	var writer io.Writer
	isMulti := false
	if cfg.ToStdout {
		writer = io.MultiWriter(os.Stdout, file)
		isMulti = true
	} else {
		writer = file
	}

	l := &Logger{
		level:   cfg.Level,
		service: cfg.ServiceName,
		logger:  log.New(writer, "", 0),
		file:    file,
		isMulti: isMulti,
	}

	return l, nil
}

// Close closes the underlying log file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// log writes a formatted log entry.
func (l *Logger) log(level Level, msg string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z07:00")

	// Get caller info
	_, file, line, ok := runtime.Caller(2)
	caller := "unknown"
	if ok {
		caller = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	formatted := fmt.Sprintf(msg, args...)
	entry := fmt.Sprintf("[%s] [%s] [%s] [%s] %s",
		timestamp, level.String(), l.service, caller, formatted)

	l.logger.Println(entry)
}

// Debug logs a debug-level message.
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.log(DEBUG, msg, args...)
}

// Info logs an info-level message.
func (l *Logger) Info(msg string, args ...interface{}) {
	l.log(INFO, msg, args...)
}

// Warn logs a warning-level message.
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.log(WARN, msg, args...)
}

// Error logs an error-level message.
func (l *Logger) Error(msg string, args ...interface{}) {
	l.log(ERROR, msg, args...)
}

// Fatal logs a fatal-level message and exits.
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.log(FATAL, msg, args...)
	os.Exit(1)
}

// GRPCAccessLog logs a gRPC request in a structured format.
func (l *Logger) GRPCAccessLog(method string, duration time.Duration, err error) {
	status := "OK"
	if err != nil {
		status = fmt.Sprintf("ERROR: %v", err)
	}
	l.Info("gRPC | method=%s | duration=%s | status=%s", method, duration.String(), status)
}
