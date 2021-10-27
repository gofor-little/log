package log_test

import (
	"errors"
	"os"
	"testing"

	"github.com/gofor-little/log"
)

func TestLog(t *testing.T) {
	log.Log = log.NewStandardLogger(os.Stdout, log.Fields{
		"tag": "logTest",
	})

	if err := log.Info(log.Fields{
		"string": "test info string",
		"bool":   true,
		"int":    64,
		"float":  3.14159,
	}); err != nil {
		t.Fatalf("failed to write Info log message")
	}

	if err := log.Error(log.Fields{
		"string": "test error string",
		"bool":   true,
		"int":    64,
		"float":  3.14159,
		"error":  errors.New("test-error"),
	}); err != nil {
		t.Fatalf("failed to write Error log message")
	}

	if err := log.Debug(log.Fields{
		"string": "test debug string",
		"bool":   true,
		"int":    64,
		"float":  3.14159,
	}); err != nil {
		t.Fatalf("failed to write Debug log message")
	}
}
