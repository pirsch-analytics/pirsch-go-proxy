package logbuch

import (
	"fmt"
	"time"
)

const (
	// StandardTimeFormat is a synonym for time.RFC3339Nano.
	StandardTimeFormat = time.RFC3339Nano
)

// StandardFormatter is the default formatter.
// It prints log messages starting with the timestamp, followed by the log level and the formatted message.
type StandardFormatter struct {
	timeFormat  string
	disableTime bool
}

// NewStandardFormatter creates a new StandardFormatter with given timestamp format.
// The timestamp can be disabled by passing an empty string.
func NewStandardFormatter(timeFormat string) *StandardFormatter {
	return &StandardFormatter{timeFormat: timeFormat, disableTime: timeFormat == ""}
}

// Fmt formats the message as described for the StandardFormatter.
func (formatter *StandardFormatter) Fmt(buffer *[]byte, level int, t time.Time, msg string, params []interface{}) {
	if !formatter.disableTime {
		*buffer = append(*buffer, t.Format(formatter.timeFormat)+" "...)
	}

	switch level {
	case LevelDebug:
		*buffer = append(*buffer, "[DEBUG] "...)
	case LevelInfo:
		*buffer = append(*buffer, "[INFO ] "...)
	case LevelWarning:
		*buffer = append(*buffer, "[WARN ] "...)
	case LevelError:
		*buffer = append(*buffer, "[ERROR] "...)
	}

	if len(params) == 0 {
		*buffer = append(*buffer, msg...)
	} else {
		*buffer = append(*buffer, fmt.Sprintf(msg, params...)...)
	}

	if len(*buffer) == 0 || (*buffer)[len(*buffer)-1] != '\n' {
		*buffer = append(*buffer, '\n')
	}
}

// Pnc formats the given message and panics.
func (formatter *StandardFormatter) Pnc(msg string, params []interface{}) {
	if len(params) == 0 {
		panic(msg)
	} else {
		panic(fmt.Sprintf(msg, params...))
	}
}
