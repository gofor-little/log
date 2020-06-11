package logger

// Logger is a structured logger interface.
type Logger interface {
	Info(fields Fields) error
	Error(fields Fields) error
	Debug(fields Fields) error
}
