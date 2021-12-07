package cloudwatchlogs

import (
	"context"
	"time"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2CWLogs "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awsv2CWLogsTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/rs/zerolog"
)

func tailParams(logGroupName string, filter string, lastEvent int64) *awsv2CWLogs.FilterLogEventsInput {
	params := &awsv2CWLogs.FilterLogEventsInput{
		LogGroupName: awsv2.String(logGroupName),
	}
	if filter != "" {
		params.FilterPattern = awsv2.String(filter)
	}
	if lastEvent != 0 {
		params.StartTime = awsv2.Int64(lastEvent)
	}
	return params
}

// TailWithContext is a utility function that support tailing the given log stream
// name using the optional filter. It returns a channel for log messages
func TailWithContext(reqContext context.Context,
	closeChan chan bool,
	awsConfig awsv2.Config,
	logGroupName string,
	filter string,
	logger *zerolog.Logger) <-chan *awsv2CWLogsTypes.FilteredLogEvent {

	// Milliseconds...
	lastSeenTimestamp := time.Now().Add(0).Unix() * 1000
	logger.Debug().
		Int64("TS", lastSeenTimestamp).
		Msg("Started polling")

	outputChannel := make(chan *awsv2CWLogsTypes.FilteredLogEvent)

	cwlogsSvc := awsv2CWLogs.NewFromConfig(awsConfig)
	tickerChan := time.NewTicker(time.Millisecond * 333).C //AWS cloudwatch logs limit is 5tx/sec
	go func() {
		for {
			select {
			case <-closeChan:
				logger.Debug().Msg("Exiting polling loop")
				return
			case <-tickerChan:
				logParam := tailParams(logGroupName, filter, lastSeenTimestamp)
				filterEvents, filterEventsErr := cwlogsSvc.FilterLogEvents(reqContext, logParam)
				if filterEventsErr != nil {
					// Just pump the thing back through the channel...
					errorEvent := &awsv2CWLogsTypes.FilteredLogEvent{
						EventId:   awsv2.String("N/A"),
						Message:   awsv2.String(filterEventsErr.Error()),
						Timestamp: awsv2.Int64(time.Now().Unix() * 1000),
					}
					outputChannel <- errorEvent
				} else {
					maxTime := int64(0)
					for _, eachEvent := range filterEvents.Events {
						if maxTime < *eachEvent.Timestamp {
							maxTime = *eachEvent.Timestamp
						}
						logger.Debug().Str("ID", *eachEvent.EventId).Msg("Event")
						outputChannel <- &eachEvent
					}
					if maxTime != 0 {
						lastSeenTimestamp = maxTime + 1
					}
				}
			}
		}
	}()
	return outputChannel
}
