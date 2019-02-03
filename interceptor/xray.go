package interceptor

import (
	"bytes"
	"container/ring"
	"io"

	sparta "github.com/mweagle/Sparta"
	"github.com/sirupsen/logrus"
)

// XRay attributes
const (
	// XRayAttrBuildID is the XRay attribute associated with this
	// service instance. See the official AWS docs at
	// https://docs.aws.amazon.com/xray/latest/devguide/xray-sdk-go-segment.html#xray-sdk-go-segment-annotations
	// for more information on XRay attributes
	XRayAttrBuildID = "buildID"
)

// XRay metadata
const (
	// XRayMetadataErrValue is the metadata kayname used to store
	// the error value when processing an event
	XRayMetadataErrValue = "error"

	// XRayMetadataErrEvent is the event associated with a lambda
	// function that errors out
	XRayMetadataErrEvent = "event"

	// XRayMetadataRequestID is the AWS request ID that came along with the request
	XRayMetadataRequestID = "reqID"

	// XRayMetadataLogs is the key associated with the logfile entries. All log
	// entries regardless of level will be included in the errLog value
	XRayMetadataLogs = "log"
)

// XRayInterceptorMode represents the mode to use for the XRay interceptor
type XRayInterceptorMode uint32

const (
	// XRayModeErrCaptureErrorValue = is the flag indicating to capture the error
	// value iff it's non-empty
	XRayModeErrCaptureErrorValue XRayInterceptorMode = 1 << iota
	// XRayModeErrCaptureEvent is the flag indicating to capture the input event iff
	// there was an error
	XRayModeErrCaptureEvent
	// XRayModeErrCaptureLogs is the flag indicating to capture all logs iff there
	// was an error
	XRayModeErrCaptureLogs
	// XRayModeErrCaptureRequestID is the flag indicating to capture the request ID iff there
	// was an error
	XRayModeErrCaptureRequestID

	// XRayAll is all options
	XRayAll = XRayModeErrCaptureErrorValue |
		XRayModeErrCaptureEvent |
		XRayModeErrCaptureLogs |
		XRayModeErrCaptureRequestID
)

var (
	// devNullLogEntry is the reserved byte value that's returned by the
	// filteringFormatter to instruct the Writer to throw away
	// the serialized version.
	//lint:ignore U1000 because it's actually used
	devNullLogEntry = []byte("/dev/null")
)

// So we can turn up the level to max. Which means everything will go to the
// serializer. If that happens, we need to know which entries
// can be thrown away. A custom formatter can be used for that. So we can fake
// this by returning a known string that tells the Out method to discard the data...
// This isn't recommended, but it does tie things together
//lint:ignore U1000 because it's used
type filteringFormatter struct {
	targetFormatter logrus.Formatter
	originalLevel   logrus.Level
	logRing         *ring.Ring
}

func (ff *filteringFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	ff.logRing.Value = entry
	ff.logRing = ff.logRing.Next()

	if ff.originalLevel >= entry.Level {
		return ff.targetFormatter.Format(entry)
	}
	return devNullLogEntry, nil
}

// The filteringWriter works together with the filteringFormatter
// to ignore formatted entries that are the /dev/null log entry values
//lint:ignore U1000 because it's used
type filteringWriter struct {
	targetOutput io.Writer
}

func (fw *filteringWriter) Write(p []byte) (n int, err error) {
	if !bytes.Equal(p, devNullLogEntry) {
		return fw.targetOutput.Write(p)
	}
	return len(p), nil
}

// xrayInterceptor is an implementation of sparta.LambdaEventInterceptors that
// handles tapping the event handling workflow and publishing a span with optional
// request information on error.
type xrayInterceptor struct {
	mode XRayInterceptorMode
	//lint:ignore U1000 because it's used
	filteringFormatter *filteringFormatter
	//lint:ignore U1000 because it's used
	filteringWriter *filteringWriter
}

// RegisterXRayInterceptor handles pushing the tracing information into XRay
func RegisterXRayInterceptor(handler *sparta.LambdaEventInterceptors,
	mode XRayInterceptorMode) *sparta.LambdaEventInterceptors {
	interceptor := &xrayInterceptor{
		mode: mode,
	}
	if handler == nil {
		handler = &sparta.LambdaEventInterceptors{}
	}
	return handler.Register(interceptor)
}
