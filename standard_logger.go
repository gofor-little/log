package log

import (
	"encoding/json"
	"io"
	"sync"
	"time"
)

// StandardLogger is a structured logger that is thread safe.
type StandardLogger struct {
	output       io.Writer
	globalFields Fields
	mutex        sync.Mutex
}

// NewStandardLogger initializes a new StandardLogger object and returns it.
// output is the output that the log messages are written to.
// globalFields are the fields that are written in every log message.
func NewStandardLogger(output io.Writer, globalFields Fields) *StandardLogger {
	return &StandardLogger{
		output:       output,
		globalFields: globalFields,
	}
}

// Info writes a log message at an info level.
func (l *StandardLogger) Info(fields Fields) error {
	fields["level"] = "info"
	return l.write(fields)
}

// Error writes a log message at an error level.
func (l *StandardLogger) Error(fields Fields) error {
	fields["level"] = "error"
	return l.write(fields)
}

// Debug writes a log message at a debug level.
func (l *StandardLogger) Debug(fields Fields) error {
	fields["level"] = "debug"
	return l.write(fields)
}

// write combines the globalFields and the passed fields, then
// marshals them to JSON. A new line character is then added to the end
// of the message and finally written to the output.
func (l *StandardLogger) write(fields Fields) error {
	for k, v := range l.globalFields {
		fields[k] = v
	}

	for k := range fields {
		if err, ok := fields[k].(error); ok {
			fields[k] = err.Error()
		}
	}

	fields["time"] = time.Now().Format("2006-01-02 15:04:05")

	data, err := json.Marshal(fields)
	if err != nil {
		return err
	}

	data = append(data, []byte("\n")...)

	l.mutex.Lock()
	defer l.mutex.Unlock()

	_, err = l.output.Write(data)
	return err
}
