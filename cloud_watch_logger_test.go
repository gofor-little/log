package logger_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	log "github.com/gofor-little/logger"
)

func TestCloudWatchLogger(t *testing.T) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String("ap-southeast-2"),
		},
		Profile: "<PROFILE_NAME>",
	}))

	var err error
	log.Log, err = log.NewCloudWatchLogger(sess, "CloudWatchLoggerTest", log.Fields{
		"tag": "cloudWatchLoggerTest",
	})
	if err != nil {
		t.Fatalf("failed to create new CloudWatchLogger: %v", err)
	}

	for i := 0; i < 10000; i++ {
		if err := log.Info(log.Fields{
			"message": fmt.Sprintf("test info message number: %v", i),
		}); err != nil {
			t.Fatalf("failed to write info log: %v", err)
		}
	}

	time.Sleep(time.Second * 5)
}
