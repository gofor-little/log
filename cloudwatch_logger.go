package log

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/gofor-little/ts"
)

// CloudWatchLogger is a structured logger that logs to CloudWatch and is thread safe.
type CloudWatchLogger struct {
	currentDay        int
	cloudWatchLogs    *cloudwatchlogs.CloudWatchLogs
	logEventsList     *ts.LinkedList
	logGroupName      *string
	nextSequenceToken *string
	globalFields      Fields
	mutex             sync.RWMutex
}

// NewCloudWatchLogger initializes a new CloudWatchLogger object and returns it.
// sess is an AWS session. logGroupName is the name of the log group in CloudWatch.
// globalFields are the fields that are written in every log message.
func NewCloudWatchLogger(sess *session.Session, logGroupName string, globalFields Fields) (*CloudWatchLogger, error) {
	log := &CloudWatchLogger{
		currentDay:     time.Now().Day(),
		cloudWatchLogs: cloudwatchlogs.New(sess),
		logEventsList:  &ts.LinkedList{},
		logGroupName:   aws.String(logGroupName),
		globalFields:   globalFields,
	}

	if err := log.checkLogGroup(); err != nil {
		return nil, err
	}

	go func() {
		throttle := time.Tick(time.Second / 5)

		for {
			<-throttle
			if err := log.putLogs(); err != nil {
				fmt.Printf("failed to send logs to CloudWatch: %v", err)
			}
		}
	}()

	return log, nil
}

// Info writes a log message at an info level.
func (c *CloudWatchLogger) Info(fields Fields) error {
	return c.queueLog("info", fields)
}

// Error writes a log message at an error level.
func (c *CloudWatchLogger) Error(fields Fields) error {
	return c.queueLog("error", fields)
}

// Debug writes a log message at a debug level.
func (c *CloudWatchLogger) Debug(fields Fields) error {
	return c.queueLog("debug", fields)
}

// createLogGroup creates a log group in CloudWatch.
func (c *CloudWatchLogger) createLogGroup() error {
	input := &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: c.logGroupName,
	}

	if err := input.Validate(); err != nil {
		return err
	}

	_, err := c.cloudWatchLogs.CreateLogGroup(input)
	return err
}

// queue combines the globalFields and the passed fields, then
// marshals them to JSON and finally adds it to a thread safe queue.
func (c *CloudWatchLogger) queueLog(level string, fields Fields) error {
	for key, value := range c.globalFields {
		fields[key] = value
	}

	fields["level"] = level

	data, err := json.Marshal(fields)
	if err != nil {
		return err
	}

	messages := [][]byte{}

	// Check if the data is larger than the max input log event size.
	// If so, split it into a slice so the data can be added over multiple
	// events. This may break the JSON structure of very large amounts of
	// data as it will be split between multiple log events.
	for {
		if len(data) <= maxInputLogEventSize {
			messages = append(messages, data)
			break
		}

		messages = append(messages, data[:maxBatchInputLogEventSize])
		data = data[maxBatchInputLogEventSize:]
	}

	// Lock the mutex so we can queue our messages.
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Range over the messages and push them to the event list.
	for _, m := range messages {
		var tail *CloudWatchLogEventSlice

		// Fetch the tail from the event list. If the message can be added to the
		// tail add it. Otherwise push to the event list and add to the new tail.
		if !c.logEventsList.IsEmpty() && c.logEventsList.GetTail().(*CloudWatchLogEventSlice).canAdd(m) {
			tail = c.logEventsList.GetTail().(*CloudWatchLogEventSlice)
		} else {
			tail = &CloudWatchLogEventSlice{}
			c.logEventsList.Push(tail)
		}

		tail.add(m)
	}

	return nil
}

// putLogs pops the oldest CloudWatchLogEventList off the queue, then
// writes it to CloudWatch.
func (c *CloudWatchLogger) putLogs() error {
	if c.logEventsList.IsEmpty() {
		return nil
	}

	if err := c.checkLogStream(); err != nil {
		return err
	}

	elements := c.logEventsList.Pop().(*CloudWatchLogEventSlice).logEvents.GetElements()
	inputLogEvents := make([]*cloudwatchlogs.InputLogEvent, len(elements))

	for i, e := range elements {
		inputLogEvents[i] = e.(*cloudwatchlogs.InputLogEvent)
	}

	input := &cloudwatchlogs.PutLogEventsInput{
		LogEvents:     inputLogEvents,
		LogGroupName:  c.logGroupName,
		LogStreamName: aws.String(time.Now().Format("2006-01-02")),
		SequenceToken: c.nextSequenceToken,
	}

	if err := input.Validate(); err != nil {
		return err
	}

	output, err := c.cloudWatchLogs.PutLogEvents(input)

	if err != nil {
		var expectedSequenceToken *string

		if aerr, ok := err.(*cloudwatchlogs.DataAlreadyAcceptedException); ok {
			expectedSequenceToken = aerr.ExpectedSequenceToken
		} else if aerr, ok := err.(*cloudwatchlogs.InvalidSequenceTokenException); ok {
			expectedSequenceToken = aerr.ExpectedSequenceToken
		}

		if expectedSequenceToken != nil {
			input.SequenceToken = expectedSequenceToken

			output, err = c.cloudWatchLogs.PutLogEvents(input)

			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	c.nextSequenceToken = output.NextSequenceToken

	return nil
}

// createLogStream creates a log stream in CloudWatch.
func (c *CloudWatchLogger) createLogStream() error {
	input := &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  c.logGroupName,
		LogStreamName: aws.String(time.Now().Format("2006-01-02")),
	}

	if err := input.Validate(); err != nil {
		return err
	}

	_, err := c.cloudWatchLogs.CreateLogStream(input)
	return err
}

// checkLogGroup checks if the log group exists in CloudWatch.
// If it doesn't it will be created.
func (c *CloudWatchLogger) checkLogGroup() error {
	logGroupExists, err := c.logGroupExists()
	if err != nil {
		return err
	}

	if logGroupExists {
		return nil
	}

	return c.createLogGroup()
}

// checkLogStream checks if the log stream exists in CloudWatch.
// If it doesn't it will be created.
func (c *CloudWatchLogger) checkLogStream() error {
	logStreamExists, err := c.logStreamExists()
	if err != nil {
		return err
	}

	if logStreamExists {
		return nil
	}

	return c.createLogStream()
}

// logGroupExists checks if the log group exists in CloudWatch.
func (c *CloudWatchLogger) logGroupExists() (bool, error) {
	input := &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: c.logGroupName,
	}

	if err := input.Validate(); err != nil {
		return false, err
	}

	output, err := c.cloudWatchLogs.DescribeLogGroups(input)
	if err != nil {
		return false, err
	}

	if output.LogGroups != nil {
		for _, logGroup := range output.LogGroups {
			if *logGroup.LogGroupName == *c.logGroupName {
				return true, nil
			}
		}
	}

	return false, nil
}

// logStreamExists checks if the log stream exists in CloudWatch.
func (c *CloudWatchLogger) logStreamExists() (bool, error) {
	input := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: c.logGroupName,
	}

	if err := input.Validate(); err != nil {
		return false, err
	}

	output, err := c.cloudWatchLogs.DescribeLogStreams(input)
	if err != nil {
		return false, nil
	}

	if output.LogStreams != nil {
		for _, logStream := range output.LogStreams {
			if *logStream.LogStreamName == time.Now().Format("2006-01-02") {
				return true, nil
			}
		}
	}

	return false, nil
}
