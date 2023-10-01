package logbuch

import (
	"fmt"
	"time"
)

// DiscardFormatter drops all messages.
type DiscardFormatter struct{}

// NewDiscardFormatter creates a new DiscardFormatter.
func NewDiscardFormatter() *DiscardFormatter {
	return new(DiscardFormatter)
}

// Fmt drops the message.
func (formatter *DiscardFormatter) Fmt(buffer *[]byte, level int, t time.Time, msg string, params []interface{}) {
	// does nothing
}

// Pnc formats the given message and panics.
func (formatter *DiscardFormatter) Pnc(msg string, params []interface{}) {
	if len(params) == 0 {
		panic(msg)
	} else {
		panic(fmt.Sprintf(msg, params...))
	}
}
