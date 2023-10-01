package logbuch

import (
	"io"
	"os"
)

var (
	logger = NewLogger(os.Stdout, os.Stderr)
)

// SetOutput sets the output channels for the default logger.
// The first parameter is used for debug, info and warning levels.
// The second one for error level.
func SetOutput(stdout, stderr io.Writer) {
	logger.SetOut(LevelDebug, stdout)
	logger.SetOut(LevelInfo, stdout)
	logger.SetOut(LevelWarning, stdout)
	logger.SetOut(LevelError, stderr)
}

// SetLevel sets the logging level.
func SetLevel(level int) {
	logger.SetLevel(level)
}

// SetFormatter sets the formatter of the default logger.
func SetFormatter(formatter Formatter) {
	logger.SetFormatter(formatter)
}

// Debug logs a formatted debug message.
func Debug(msg string, params ...interface{}) {
	logger.Debug(msg, params...)
}

// Info logs a formatted info message.
func Info(msg string, params ...interface{}) {
	logger.Info(msg, params...)
}

// Warn logs a formatted warning message.
func Warn(msg string, params ...interface{}) {
	logger.Warn(msg, params...)
}

// Error logs a formatted error message.
func Error(msg string, params ...interface{}) {
	// maximum level cannot be disabled
	logger.Error(msg, params...)
}

// Fatal logs a formatted error message and panics.
func Fatal(msg string, params ...interface{}) {
	logger.Fatal(msg, params...)
}
