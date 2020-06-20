package log_test

import (
	"os"
	"testing"

	"github.com/gofor-little/log"
)

func TestLog(t *testing.T) {
	log.Log = log.NewStandardLogger(os.Stdout, log.Fields{
		"tag": "logTest",
	})

	err := log.Info(log.Fields{
		"string": "test info string",
		"bool":   true,
		"int":    64,
		"float":  3.14159,
	})
	if err != nil {
		t.Fatalf("failed to write Info log message")
	}

	err = log.Error(log.Fields{
		"string": "test error string",
		"bool":   true,
		"int":    64,
		"float":  3.14159,
	})
	if err != nil {
		t.Fatalf("failed to write Error log message")
	}

	err = log.Debug(log.Fields{
		"string": "test debug string",
		"bool":   true,
		"int":    64,
		"float":  3.14159,
	})
	if err != nil {
		t.Fatalf("failed to write Debug log message")
	}
}
