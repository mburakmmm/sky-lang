package skylib

import (
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel represents log severity
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

var currentLogLevel = LogLevelInfo

// SetLogLevel sets minimum log level
func LogSetLevel(level int) {
	currentLogLevel = LogLevel(level)
}

// Log logs a message with level
func Log(level int, msg string, fields map[string]interface{}) {
	if LogLevel(level) < currentLogLevel {
		return
	}

	levelStr := ""
	switch LogLevel(level) {
	case LogLevelDebug:
		levelStr = "DEBUG"
	case LogLevelInfo:
		levelStr = "INFO"
	case LogLevelWarn:
		levelStr = "WARN"
	case LogLevelError:
		levelStr = "ERROR"
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Format fields
	fieldsStr := ""
	for k, v := range fields {
		fieldsStr += fmt.Sprintf(" %s=%v", k, v)
	}

	fmt.Fprintf(os.Stderr, "[%s] %s: %s%s\n", timestamp, levelStr, msg, fieldsStr)
}

// LogDebug logs debug message
func LogDebug(msg string, fields map[string]interface{}) {
	Log(int(LogLevelDebug), msg, fields)
}

// LogInfo logs info message
func LogInfo(msg string, fields map[string]interface{}) {
	Log(int(LogLevelInfo), msg, fields)
}

// LogWarn logs warning message
func LogWarn(msg string, fields map[string]interface{}) {
	Log(int(LogLevelWarn), msg, fields)
}

// LogError logs error message
func LogError(msg string, fields map[string]interface{}) {
	Log(int(LogLevelError), msg, fields)
}

// LogToFile sets log output to file
func LogToFile(filename string) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	log.SetOutput(file)
	return nil
}
