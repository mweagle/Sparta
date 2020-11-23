// +build lambdabinary

package interceptor

import (
	"container/ring"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-xray-sdk-go/xray"
	sparta "github.com/mweagle/Sparta"
	"github.com/rs/zerolog"
)

const (
	logRingSize = 1024
)

func (xri *xrayInterceptor) Begin(ctx context.Context, msg json.RawMessage) context.Context {

	// Put this into the context...
	segmentCtx, segment := xray.BeginSubsegment(ctx, "Sparta")

	// Add some xray annotations to help track this version
	// of the service
	// https://docs.aws.amazon.com/xray/latest/devguide/xray-sdk-go-segment.html
	if segment != nil {
		errAddAnnotation := segment.AddAnnotation(XRayAttrBuildID, sparta.StampedBuildID)
		if errAddAnnotation != nil {
			log.Printf("Failed to update segment context: %s", errAddAnnotation)
		}
		segmentCtx = context.WithValue(segmentCtx, contextKeySegment, segment)
	}
	return segmentCtx
}

func (xri *xrayInterceptor) BeforeSetup(ctx context.Context, msg json.RawMessage) context.Context {
	return ctx
}

func (xri *xrayInterceptor) AfterSetup(ctx context.Context, msg json.RawMessage) context.Context {
	if xri.mode&XRayModeErrCaptureLogs != 0 {
		logger, loggerOk := ctx.Value(sparta.ContextKeyLogger).(*zerolog.Logger)
		if loggerOk {
			// How to have a single writer with different levels? We can keep the old logger,
			// We need a ring buffer to store the messages, and then we send it
			// to the original logger
			xri.zerologXRayHandler = &zerologXRayHandler{
				logRing:       ring.New(logRingSize),
				originalLevel: logger.GetLevel(),
			}
			// So this is the new logger, going to the same place, but
			// tapped for the ring...
			newLogger := zerolog.New(xri.zerologXRayHandler).
				With().
				Timestamp().
				Logger().
				Level(zerolog.TraceLevel).Hook(xri.zerologXRayHandler)
			ctx = context.WithValue(ctx, sparta.ContextKeyLogger, &newLogger)
		} else {
			log.Printf("WARNING: Failed to get logger from context\n")
		}
	}
	return ctx
}

func (xri *xrayInterceptor) BeforeDispatch(ctx context.Context, msg json.RawMessage) context.Context {
	return ctx
}
func (xri *xrayInterceptor) AfterDispatch(ctx context.Context, msg json.RawMessage) context.Context {
	return ctx
}

func (xri *xrayInterceptor) Complete(ctx context.Context, msg json.RawMessage) context.Context {
	segmentVal := ctx.Value(contextKeySegment)
	if segmentVal != nil {
		segment, segmentOk := segmentVal.(*xray.Segment)
		if segmentOk {
			errValue, errValueOk := ctx.Value(sparta.ContextKeyLambdaError).(error)
			if errValueOk && errValue != nil {
				var metadataEventErr error

				// Include the error value?
				if xri.mode&XRayModeErrCaptureErrorValue != 0 {
					metadataEventErr = segment.AddMetadataToNamespace(sparta.ProperName, XRayMetadataErrValue, errValue.Error())
					if metadataEventErr != nil {
						log.Printf("Failed to set event %s metadata: %s", XRayMetadataErrValue, metadataEventErr)
					}
				}

				// Include the event?
				if xri.mode&XRayModeErrCaptureEvent != 0 {
					metadataEventErr = segment.AddMetadataToNamespace(sparta.ProperName, XRayMetadataErrEvent, msg)
					if metadataEventErr != nil {
						log.Printf("Failed to set event %s metadata: %s", XRayMetadataErrEvent, metadataEventErr)
					}
				}

				// Include the request ID?
				if xri.mode&XRayModeErrCaptureRequestID != 0 {
					awsContext, _ := lambdacontext.FromContext(ctx)
					if awsContext != nil {
						metadataEventErr = segment.AddMetadataToNamespace(sparta.ProperName, XRayMetadataRequestID, awsContext.AwsRequestID)
						if metadataEventErr != nil {
							log.Printf("Failed to set event %s metadata: %s", XRayMetadataRequestID, metadataEventErr)
						}
					}
				}

				// Include the cached log info?
				if xri.mode&XRayModeErrCaptureLogs != 0 {
					logMessages := make([]interface{}, 0)
					xri.zerologXRayHandler.logRing.Do(func(eachLogEntry interface{}) {
						if eachLogEntry != nil {
							stringVal, stringOk := eachLogEntry.(string)
							if stringOk {
								var unmarshalMap map[string]interface{}
								unmarshalErr := json.Unmarshal([]byte(stringVal), &unmarshalMap)
								if unmarshalErr == nil {
									logMessages = append(logMessages, unmarshalMap)
								} else {
									logMessages = append(logMessages, fmt.Sprintf("%s", stringVal))
								}
							} else {
								logMessages = append(logMessages, fmt.Sprintf("%v", eachLogEntry))
							}
						}
					})

					metadataEventErr = segment.AddMetadataToNamespace(sparta.ProperName, XRayMetadataLogs, logMessages)
					if metadataEventErr != nil {
						log.Printf("Failed to set event %s metadata: %s", XRayMetadataLogs, metadataEventErr)
					}
					// Either way, clear out the ring...
					xri.zerologXRayHandler.logRing = ring.New(logRingSize)
				}
			}
			segment.Close(errValue)
		}
	}
	return ctx
}
