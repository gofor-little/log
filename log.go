package log

import "errors"

// Log is the package wide Logger. This can be used to have a
// Logger that is globally available within a codebase.
var Log Logger

// Info calls Info for the initialized Logger.
func Info(fields Fields) error {
	if Log == nil {
		return errors.New("log is nil, make sure to initialize the Logger")
	}

	return Log.Info(fields)
}

// Error calls Error for the initialized Logger.
func Error(fields Fields) error {
	if Log == nil {
		return errors.New("log is nil, make sure to initialize the Logger")
	}

	return Log.Error(fields)
}

// Debug calls Debug for the initialized Logger.
func Debug(fields Fields) error {
	if Log == nil {
		return errors.New("log is nil, make sure to initialize the Logger")
	}

	return Log.Debug(fields)
}
