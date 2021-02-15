## A package for structured logging

![GitHub tag (latest SemVer pre-release)](https://img.shields.io/github/v/tag/gofor-little/log?include_prereleases)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gofor-little/log)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://raw.githubusercontent.com/gofor-little/log/main/LICENSE)
![GitHub Workflow Status](https://img.shields.io/github/workflow/status/gofor-little/log/CI)
[![Go Report Card](https://goreportcard.com/badge/github.com/gofor-little/log)](https://goreportcard.com/report/github.com/gofor-little/log)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/gofor-little/log)](https://pkg.go.dev/github.com/gofor-little/log)

### Introduction
* Structured logging
* Supports AWS CloudWatch
* Simple interface for implementing your own

### Example
```go
package main

import (
    "fmt"

    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/gofor-little/log"
)

func main() {
    // Standard logger writes to a user defined io.Writer.
	log.Log = log.NewStandardLogger(os.Stdout, log.Fields{
		"tag": "standard_logger",
    })

    // CloudWatch logger writes to an AWS CloudWatch log group.
    sess, err = session.NewSession()
    log.Log, err = log.NewCloudWatchLogger(sess, "CloudWatchLoggerTest", log.Fields{
		"tag": "cloudWatchLoggerTest",
	})
	if err != nil {
		t.Fatalf("failed to create new CloudWatchLogger: %v", err)
	}

    // Log at info, error and debug levels.
    log.Info(log.Fields{
		"message": "info message",
		"bool":   true,
		"int":    64,
		"float":  3.14159,
    })

	log.Error(log.Fields{
		"string": "error message",
		"bool":   true,
		"int":    64,
		"float":  3.14159,
	})

	log.Debug(log.Fields{
		"string": "debug message",
		"bool":   true,
		"int":    64,
		"float":  3.14159,
	})
}
```

### Testing
Ensure the following environment variables are set, usually with a .env file.
* ```AWS_PROFILE``` (an AWS CLI profile name)
* ```AWS_REGION``` (a valid AWS region)

Run ```go test -v -race ./...``` in the root directory.
