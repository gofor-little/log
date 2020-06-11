package logger_test

import (
	"os"
	"sync"
	"testing"

	log "github.com/gofor-little/logger"
)

func TestLog(t *testing.T) {
	log.Log = log.NewStandardLogger(os.Stdout, log.Fields{
		"tag": "logTest",
	})

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(3)

	go func(waitGroup *sync.WaitGroup) {
		for i := 0; i < 10; i++ {
			if err := log.Info(log.Fields{
				"string": "test info string",
				"bool":   true,
				"int":    64,
				"float":  3.14159,
			}); err != nil {
				t.Fatalf("failed to write Info log message")
			}
		}

		waitGroup.Done()
	}(waitGroup)

	go func(waitGroup *sync.WaitGroup) {
		for i := 0; i < 10; i++ {
			if err := log.Error(log.Fields{
				"string": "test error string",
				"bool":   true,
				"int":    64,
				"float":  3.14159,
			}); err != nil {
				t.Fatalf("failed to write Error log message")
			}
		}

		waitGroup.Done()
	}(waitGroup)

	go func(waitGroup *sync.WaitGroup) {
		for i := 0; i < 10; i++ {
			if err := log.Debug(log.Fields{
				"string": "test debug string",
				"bool":   true,
				"int":    64,
				"float":  3.14159,
			}); err != nil {
				t.Fatalf("failed to write Debug log message")
			}
		}

		waitGroup.Done()
	}(waitGroup)

	waitGroup.Wait()
}
