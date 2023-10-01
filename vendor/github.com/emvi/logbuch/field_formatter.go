package logbuch

import (
	"fmt"
	"strings"
	"time"
)

// Fields is used together with the FieldFormatter.
type Fields map[string]interface{}

// FieldFormatter adds fields to the output as key value pairs. The message won't be formatted.
// It prints log messages starting with the timestamp, followed by the log level, the message and key value pairs.
// To make this work the first and only parameter must be of type Fields.
//
// Example:
//  logbuch.Debug("Hello World!", logbuch.Fields{"integer": 123, "string": "test"})
//
// If there is more than one parameter or the type of the parameter is different,
// all parameters will be appended after the message.
type FieldFormatter struct {
	timeFormat  string
	disableTime bool
	separator   string
}

// NewFieldFormatter creates a new FieldFormatter with given timestamp format and separator between message and key value pairs.
// The timestamp can be disabled by passing an empty string.
func NewFieldFormatter(timeFormat, separator string) *FieldFormatter {
	return &FieldFormatter{timeFormat: timeFormat, disableTime: timeFormat == "", separator: separator}
}

// Fmt formats the message as described for the FieldFormatter.
func (formatter *FieldFormatter) Fmt(buffer *[]byte, level int, t time.Time, msg string, params []interface{}) {
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

	*buffer = append(*buffer, msg...)

	if len(params) > 0 {
		fields, ok := params[0].(Fields)

		if len(params) == 1 && ok {
			*buffer = append(*buffer, formatter.separator...)

			for k, v := range fields {
				*buffer = append(*buffer, fmt.Sprintf(" %s=%v", k, v)...)
			}
		} else {
			*buffer = append(*buffer, formatter.separator...)

			for _, v := range params {
				*buffer = append(*buffer, fmt.Sprintf(" %v", v)...)
			}
		}
	}

	*buffer = append(*buffer, '\n')
}

// Pnc formats the given message and panics.
func (formatter *FieldFormatter) Pnc(msg string, params []interface{}) {
	if len(params) > 0 {
		fields, ok := params[0].(Fields)

		if len(params) == 1 && ok {
			var builder strings.Builder
			builder.WriteString(msg)

			for k, v := range fields {
				builder.WriteString(fmt.Sprintf(" %s=%v", k, v))
			}

			panic(builder.String())
		} else {
			// it's not possible to format the message right here due to a bug in go vet...
			// so we pass it on to another function which takes an array as its argument
			formatter.panicWithFmt(msg, params)
		}
	} else {
		panic(msg)
	}
}

func (formatter *FieldFormatter) panicWithFmt(msg string, params []interface{}) {
	panic(fmt.Sprintf(msg, params...))
}
