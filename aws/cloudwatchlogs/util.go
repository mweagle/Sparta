package cloudwatchlogs

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

func tailParams(logGroupName string, filter string, lastEvent int64) *cloudwatchlogs.FilterLogEventsInput {
	params := &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName: aws.String(logGroupName),
	}
	if filter != "" {
		params.FilterPattern = aws.String(filter)
	}
	if lastEvent != 0 {
		params.StartTime = aws.Int64(lastEvent)
	}
	return params
}

// TailWithContext is a utility function that support tailing the given log stream
// name using the optional filter. It returns a channel for log messages
func TailWithContext(reqContext aws.Context,
	closeChan chan bool,
	awsSession *session.Session,
	logGroupName string,
	filter string,
	logger *logrus.Logger) <-chan *cloudwatchlogs.FilteredLogEvent {

	// Milliseconds...
	lastSeenTimestamp := time.Now().Add(0).Unix() * 1000
	logger.WithField("TS", lastSeenTimestamp).Debug("Started polling")
	outputChannel := make(chan *cloudwatchlogs.FilteredLogEvent)
	tailHandler := func(res *cloudwatchlogs.FilterLogEventsOutput, lastPage bool) bool {
		maxTime := int64(0)
		for _, eachEvent := range res.Events {
			if maxTime < *eachEvent.Timestamp {
				maxTime = *eachEvent.Timestamp
			}
			logger.WithField("ID", *eachEvent.EventId).Debug("Event")
			outputChannel <- eachEvent
		}
		if maxTime != 0 {
			lastSeenTimestamp = maxTime + 1
		}
		return !lastPage
	}

	cwlogsSvc := cloudwatchlogs.New(awsSession)
	tickerChan := time.NewTicker(time.Millisecond * 333).C //AWS cloudwatch logs limit is 5tx/sec
	go func() {
		for {
			select {
			case <-closeChan:
				logger.Debug("Exiting polling loop")
				return
			case <-tickerChan:
				logParam := tailParams(logGroupName, filter, lastSeenTimestamp)
				error := cwlogsSvc.FilterLogEventsPagesWithContext(reqContext, logParam, tailHandler)
				if error != nil {
					// Just pump the thing back through the channel...
					errorEvent := &cloudwatchlogs.FilteredLogEvent{
						EventId:   aws.String("N/A"),
						Message:   aws.String(error.Error()),
						Timestamp: aws.Int64(time.Now().Unix() * 1000),
					}
					outputChannel <- errorEvent
				}
			}
		}
	}()
	return outputChannel
}
