package logbuch

import (
	"io"
	"sync"
	"time"
)

const (
	// LevelDebug log all messages.
	LevelDebug = iota

	// LevelInfo log info, warning and error messages.
	LevelInfo

	// LevelWarning log warning and error messages.
	LevelWarning

	// LevelError log error messages only.
	LevelError
)

// Logger writes messages to different io.Writers depending on the log level by using a Formatter.
type Logger struct {
	m          sync.Mutex
	level      int
	formatter  Formatter
	debugOut   io.Writer
	infoOut    io.Writer
	warningOut io.Writer
	errorOut   io.Writer
	buffer     []byte

	// PanicOnErr enables panics if the logger cannot write to log output.
	PanicOnErr bool
}

// NewLogger creates a new logger using the StandardFormatter for given io.Writers.
func NewLogger(stdout, stderr io.Writer) *Logger {
	return &Logger{formatter: NewStandardFormatter(StandardTimeFormat),
		debugOut:   stdout,
		infoOut:    stdout,
		warningOut: stdout,
		errorOut:   stderr}
}

// SetLevel sets the log level.
func (log *Logger) SetLevel(level int) {
	log.m.Lock()
	defer log.m.Unlock()
	log.level = getValidLevel(level)
}

// GetLevel returns the log level.
func (log *Logger) GetLevel() int {
	return log.level
}

// SetFormatter sets the formatter.
func (log *Logger) SetFormatter(formatter Formatter) {
	log.m.Lock()
	defer log.m.Unlock()
	log.formatter = formatter
}

// GetFormatter returns the formatter.
func (log *Logger) GetFormatter() Formatter {
	return log.formatter
}

// SetOut sets the io.Writer for given level.
func (log *Logger) SetOut(level int, out io.Writer) {
	log.m.Lock()
	defer log.m.Unlock()

	switch level {
	case LevelDebug:
		log.debugOut = out
	case LevelInfo:
		log.infoOut = out
	case LevelWarning:
		log.warningOut = out
	default:
		log.errorOut = out
	}
}

// GetOut returns the io.Writer for given level.
func (log *Logger) GetOut(level int) io.Writer {
	switch level {
	case LevelDebug:
		return log.debugOut
	case LevelInfo:
		return log.infoOut
	case LevelWarning:
		return log.warningOut
	default:
		return log.errorOut
	}
}

// Debug logs a formatted debug message.
func (log *Logger) Debug(msg string, params ...interface{}) {
	if log.level <= LevelDebug {
		log.log(LevelDebug, msg, params)
	}
}

// Info logs a formatted info message.
func (log *Logger) Info(msg string, params ...interface{}) {
	if log.level <= LevelInfo {
		log.log(LevelInfo, msg, params)
	}
}

// Warn logs a formatted warning message.
func (log *Logger) Warn(msg string, params ...interface{}) {
	if log.level <= LevelWarning {
		log.log(LevelWarning, msg, params)
	}
}

// Error logs a formatted error message.
func (log *Logger) Error(msg string, params ...interface{}) {
	// maximum level cannot be disabled
	log.log(LevelError, msg, params)
}

// Fatal logs a formatted error message and panics.
func (log *Logger) Fatal(msg string, params ...interface{}) {
	log.Error(msg, params...)
	log.formatter.Pnc(msg, params)
}

func (log *Logger) log(level int, msg string, params []interface{}) {
	now := time.Now()
	log.m.Lock()
	defer log.m.Unlock()
	log.buffer = log.buffer[:0]
	log.formatter.Fmt(&log.buffer, level, now, msg, params)
	var err error

	switch level {
	case LevelDebug:
		_, err = log.debugOut.Write(log.buffer)
	case LevelInfo:
		_, err = log.infoOut.Write(log.buffer)
	case LevelWarning:
		_, err = log.warningOut.Write(log.buffer)
	default:
		_, err = log.errorOut.Write(log.buffer)
	}

	// panic in case the logger cannot write to the configured io.Writer and panic is enabled
	if err != nil && log.PanicOnErr {
		panic(err)
	}
}

func getValidLevel(level int) int {
	if level < LevelDebug || level > LevelError {
		return LevelDebug
	}

	return level
}
