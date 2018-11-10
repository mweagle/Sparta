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
	// XRayModeCaptureEvent is the flag indicating to capture the input event if there is
	// an error
	XRayModeCaptureEvent XRayInterceptorMode = 1 << iota
	// XRayModeCaptureLogs is the flag indicating to capture all logs
	XRayModeCaptureLogs
	// XRayModeCaptureRequestID is the flag indicating to capture the request ID
	XRayModeCaptureRequestID

	// XRayAll is all options
	XRayAll = XRayModeCaptureEvent |
		XRayModeCaptureLogs |
		XRayModeCaptureRequestID
)

var (
	// devNullLogEntry is the reserved byte value that's returned by the
	// filteringFormatter to instruct the Writer to throw away
	// the serialized version.
	devNullLogEntry = []byte("/dev/null")
)

// So we can turn up the level to max. Which means everything will go to the
// serializer. If that happens, we need to know which entries
// can be thrown away. A custom formatter can be used for that. So we can fake
// this by returning a known string that tells the Out method to discard the data...
// This isn't recommended, but it does tie things together
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
type filteringWriter struct {
	targetOutput io.Writer
}

func (fw *filteringWriter) Write(p []byte) (n int, err error) {
	if bytes.Compare(p, devNullLogEntry) != 0 {
		return fw.targetOutput.Write(p)
	}
	return len(p), nil
}

// xrayInterceptor is an implementation of sparta.LambdaEventInterceptors that
// handles tapping the event handling workflow and publishing a span with optional
// request information on error.
type xrayInterceptor struct {
	mode               XRayInterceptorMode
	filteringFormatter *filteringFormatter
	filteringWriter    *filteringWriter
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
