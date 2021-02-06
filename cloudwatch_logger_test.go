package log_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gofor-little/env"
	"github.com/stretchr/testify/require"

	"github.com/gofor-little/log"
)

func TestCloudWatchLogger(t *testing.T) {
	var sess *session.Session
	var err error

	if err := env.Load(".env"); err != nil {
		t.Log(".env file not found, ignore this if running in CI/CD Pipeline")
	}

	if env.Get("ENVIRONMENT", "ci/cd") == "development" {
		sess, err = session.NewSessionWithOptions(session.Options{
			Config: aws.Config{
				Region: aws.String(env.Get("AWS_REGION", "ap-southeast-2")),
			},
			Profile: env.Get("AWS_PROFILE", "default"),
		})
	} else {
		sess, err = session.NewSession()
	}
	require.NoError(t, err)

	log.Log, err = log.NewCloudWatchLogger(sess, "CloudWatchLoggerTest", log.Fields{
		"tag": "cloudWatchLoggerTest",
	})
	require.NoError(t, err)

	err = log.Info(log.Fields{
		"string": "test info string",
		"bool":   true,
		"int":    64,
		"float":  3.14159,
	})
	require.NoError(t, err)

	err = log.Error(log.Fields{
		"string": "test error string",
		"bool":   true,
		"int":    64,
		"float":  3.14159,
	})
	require.NoError(t, err)

	err = log.Debug(log.Fields{
		"string": "test debug string",
		"bool":   true,
		"int":    64,
		"float":  3.14159,
	})
	require.NoError(t, err)

	time.Sleep(time.Second)
}
