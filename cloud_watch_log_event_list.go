package logger

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/gofor-little/ts"
)

const (
	inputLogEventOffset       = 28      // The offset in bytes that is added to each InputLogEvent that is pushed to CloudWatch.
	maxInputLogEventSize      = 256000  // The max size in bytes a single InputLogEvent's message can be. Approx 256KB.
	maxBatchInputLogEventSize = 1048576 // The max size in bytes that a batch of InputLogEvents can be. Approx 1MB.
	maxBatchInputLogEvents    = 10000   // The max amount of InputLogEvents that can be in a single batch.
)

// CloudWatchLogEventList stores a thread safe slice of InputLogEvents.
// size is the current size in bytes of the log messages.
type CloudWatchLogEventList struct {
	logEvents *ts.Slice
	size      int
}

// add adds a new InputLogEvent to the logEvents slice.
func (c *CloudWatchLogEventList) add(message string) error {
	if c.isFull() || c.size+len(message)+inputLogEventOffset > maxBatchInputLogEventSize {
		return fmt.Errorf("max put size of %v exceeded", maxBatchInputLogEventSize)
	}

	if len(message) > maxInputLogEventSize {
		return fmt.Errorf("event size: %v is larger than the max event size: %v", len(message), maxInputLogEventSize)
	}

	c.size += len(message) + inputLogEventOffset

	inputLogEvent := &cloudwatchlogs.InputLogEvent{
		Message:   aws.String(message),
		Timestamp: aws.Int64(time.Now().UnixNano() / int64(time.Millisecond)),
	}

	if err := inputLogEvent.Validate(); err != nil {
		return err
	}

	c.logEvents.Add(inputLogEvent)

	return nil
}

// canAdd checks if message can be added to the logEvents slice
// by first checking if size will still be less than maxBatchInputLogEventSize
// with message being appended.
func (c *CloudWatchLogEventList) canAdd(message []byte) bool {
	if c.isFull() || c.size+len(message)+inputLogEventOffset > maxBatchInputLogEventSize {
		return false
	}

	return true
}

// isFull checks if size is greater than or equal to maxBatchInputLogEventSize.
func (c *CloudWatchLogEventList) isFull() bool {
	if c.logEvents == nil {
		c.logEvents = &ts.Slice{}
	}

	if c.logEvents.Length() >= maxBatchInputLogEvents || c.size >= maxBatchInputLogEventSize {
		return true
	}

	return false
}
