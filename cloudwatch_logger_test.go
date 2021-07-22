package log_test

import (
	"context"
	"testing"
	"time"

	"github.com/gofor-little/env"
	"github.com/matryer/is"

	"github.com/gofor-little/log"
)

func TestCloudWatchLogger(t *testing.T) {
	is := is.New(t)
	is.NoErr(env.Load(".env"))

	var err error
	log.Log, err = log.NewCloudWatchLogger(context.Background(), env.Get("AWS_PROFILE", ""), env.Get("AWS_REGION", ""), "CloudWatchLoggerTest", log.Fields{
		"tag": "cloudWatchLoggerTest",
	})
	is.NoErr(err)

	err = log.Info(log.Fields{
		"string": "test info string",
		"bool":   true,
		"int":    64,
		"float":  3.14159,
	})
	is.NoErr(err)

	err = log.Error(log.Fields{
		"string": "test error string",
		"bool":   true,
		"int":    64,
		"float":  3.14159,
	})
	is.NoErr(err)

	err = log.Debug(log.Fields{
		"string": "test debug string",
		"bool":   true,
		"int":    64,
		"float":  3.14159,
	})
	is.NoErr(err)

	time.Sleep(time.Second)
}
