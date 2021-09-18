package log

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/gofor-little/ts"
)

// CloudWatchLogger is a structured logger that logs to CloudWatch and is thread safe.
type CloudWatchLogger struct {
	currentDay        int
	cloudWatchLogs    *cloudwatchlogs.Client
	logEventsList     *ts.LinkedList
	logGroupName      *string
	nextSequenceToken *string
	globalFields      Fields
	mutex             sync.RWMutex
}

// NewCloudWatchLogger initializes a new CloudWatchLogger object and returns it.
// The profile and region parameters are optional if authentication with CloudWatch
// can be provided in other ways, such as IAM roles. logGroupName is the name of
// the log group in CloudWatch. globalFields are the fields that are written in every log message.
func NewCloudWatchLogger(ctx context.Context, profile string, region string, logGroupName string, globalFields Fields) (*CloudWatchLogger, error) {
	var cfg aws.Config
	var err error

	if profile != "" && region != "" {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile), config.WithRegion(region))
	} else {
		cfg, err = config.LoadDefaultConfig(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load default config: %w", err)
	}

	logger := &CloudWatchLogger{
		currentDay:     time.Now().Day(),
		cloudWatchLogs: cloudwatchlogs.NewFromConfig(cfg),
		logEventsList:  &ts.LinkedList{},
		logGroupName:   aws.String(logGroupName),
		globalFields:   globalFields,
	}

	if err := logger.checkLogGroup(ctx); err != nil {
		return nil, err
	}

	go func() {
		//lint:ignore SA1015 This is an endless function.
		//nolint:staticcheck // GolangCI Lint.
		throttle := time.Tick(time.Second / 5)

		for {
			<-throttle
			if err := logger.putLogs(ctx); err != nil {
				log.Fatalf("failed to send logs to CloudWatch: %v", err)
			}
		}
	}()

	return logger, nil
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
func (c *CloudWatchLogger) createLogGroup(ctx context.Context) error {
	input := &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: c.logGroupName,
	}

	_, err := c.cloudWatchLogs.CreateLogGroup(ctx, input)
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

		if err := tail.add(m); err != nil {
			return err
		}
	}

	return nil
}

// putLogs pops the oldest CloudWatchLogEventList off the queue, then
// writes it to CloudWatch.
func (c *CloudWatchLogger) putLogs(ctx context.Context) error {
	if c.logEventsList.IsEmpty() {
		return nil
	}

	if err := c.checkLogStream(ctx); err != nil {
		return err
	}

	elements := c.logEventsList.Pop().(*CloudWatchLogEventSlice).logEvents.GetElements()
	inputLogEvents := make([]types.InputLogEvent, len(elements))

	for i, e := range elements {
		inputLogEvents[i] = e.(types.InputLogEvent)
	}

	input := &cloudwatchlogs.PutLogEventsInput{
		LogEvents:     inputLogEvents,
		LogGroupName:  c.logGroupName,
		LogStreamName: aws.String(time.Now().Format("2006-01-02")),
		SequenceToken: c.nextSequenceToken,
	}

	output, err := c.cloudWatchLogs.PutLogEvents(ctx, input)
	if err != nil {
		var dataAlreadyAccepted *types.DataAlreadyAcceptedException
		var invalidSequenceToken *types.InvalidSequenceTokenException

		if errors.As(err, &dataAlreadyAccepted) {
			input.SequenceToken = dataAlreadyAccepted.ExpectedSequenceToken
			output, err = c.cloudWatchLogs.PutLogEvents(ctx, input)
			if err != nil {
				return err
			}
		} else if errors.As(err, &invalidSequenceToken) {
			input.SequenceToken = dataAlreadyAccepted.ExpectedSequenceToken
			output, err = c.cloudWatchLogs.PutLogEvents(ctx, input)
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
func (c *CloudWatchLogger) createLogStream(ctx context.Context) error {
	input := &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  c.logGroupName,
		LogStreamName: aws.String(time.Now().Format("2006-01-02")),
	}

	_, err := c.cloudWatchLogs.CreateLogStream(ctx, input)
	return err
}

// checkLogGroup checks if the log group exists in CloudWatch.
// If it doesn't it will be created.
func (c *CloudWatchLogger) checkLogGroup(ctx context.Context) error {
	logGroupExists, err := c.logGroupExists(ctx)
	if err != nil {
		return err
	}

	if logGroupExists {
		return nil
	}

	return c.createLogGroup(ctx)
}

// checkLogStream checks if the log stream exists in CloudWatch.
// If it doesn't it will be created.
func (c *CloudWatchLogger) checkLogStream(ctx context.Context) error {
	logStreamExists, err := c.logStreamExists(ctx)
	if err != nil {
		return err
	}

	if logStreamExists {
		return nil
	}

	return c.createLogStream(ctx)
}

// logGroupExists checks if the log group exists in CloudWatch.
func (c *CloudWatchLogger) logGroupExists(ctx context.Context) (bool, error) {
	input := &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: c.logGroupName,
	}

	output, err := c.cloudWatchLogs.DescribeLogGroups(ctx, input)
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
func (c *CloudWatchLogger) logStreamExists(ctx context.Context) (bool, error) {
	input := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: c.logGroupName,
	}

	output, err := c.cloudWatchLogs.DescribeLogStreams(ctx, input)
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
