package logbuch

import (
	"time"
)

// Formatter is an interface to format log messages.
type Formatter interface {
	// Fmt formats a logger message and writes the result into the buffer.
	Fmt(*[]byte, int, time.Time, string, []interface{})

	// Pnc formats the given message and panics.
	Pnc(string, []interface{})
}
