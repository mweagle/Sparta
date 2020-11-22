// +build lambdabinary

package interceptor

import (
	"container/ring"
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-xray-sdk-go/xray"
	sparta "github.com/mweagle/Sparta"
	"github.com/sirupsen/logrus"
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
			// So we need a loggerWrapper that has the debug level turned on.
			// This filtering formatter will put everything in a logring and
			// flush iff there is an error
			xri.filteringFormatter = &filteringFormatter{
				targetFormatter: logger.Formatter,
				originalLevel:   logger.Level,
				logRing:         ring.New(logRingSize),
			}
			xri.filteringWriter = &filteringWriter{
				targetOutput: logger.Out,
			}
			logger.SetLevel(logrus.TraceLevel)
			logger.SetFormatter(xri.filteringFormatter)
			logger.SetOutput(xri.filteringWriter)
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

				// Include the error value?
				if xri.mode&XRayModeErrCaptureErrorValue != 0 {
					metadataEventErr := segment.AddMetadataToNamespace(sparta.ProperName, XRayMetadataErrValue, errValue.Error())
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
					logFormatter := &logrus.JSONFormatter{}
					logMessages := make([]map[string]*json.RawMessage, 0)
					xri.filteringFormatter.logRing.Do(func(eachLogEntry interface{}) {
						if eachLogEntry != nil {
							// Format each one to text?
							eachTypedEntry, eachTypedEntryOk := eachLogEntry.(*logrus.Entry)
							if eachTypedEntryOk {
								formatted, formattedErr := logFormatter.Format(eachTypedEntry)
								if formattedErr == nil {
									var jsonData map[string]*json.RawMessage
									unmarshalErr := json.Unmarshal(formatted, &jsonData)
									if unmarshalErr == nil {
										logMessages = append(logMessages, jsonData)
									}
								}
							}
						}
					})
					metadataEventErr = segment.AddMetadataToNamespace(sparta.ProperName, XRayMetadataLogs, logMessages)
					if metadataEventErr != nil {
						log.Printf("Failed to set event %s metadata: %s", XRayMetadataLogs, metadataEventErr)
					}
					// Either way, clear out the ring...
					xri.filteringFormatter.logRing = ring.New(logRingSize)
				}
			}
			segment.Close(errValue)
		}
	}
	return ctx
}
