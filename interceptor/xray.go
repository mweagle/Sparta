package interceptor

import (
	"container/ring"
	"os"
	"strings"

	sparta "github.com/mweagle/Sparta/v3"
	"github.com/rs/zerolog"
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

const (
	reservedFieldName string = "xrsparta"
)

// The filteringWriter works together with the filteringFormatter
// to ignore formatted entries that are the /dev/null log entry values
//lint:ignore U1000 because it's used
type zerologXRayHandler struct {
	logRing       *ring.Ring
	originalLevel zerolog.Level
}

func (zrh *zerologXRayHandler) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level >= zrh.originalLevel {
		e.Str(reservedFieldName, "1")
	}
}

func (zrh *zerologXRayHandler) Write(p []byte) (n int, err error) {
	zrh.logRing.Value = string(p)
	zrh.logRing = zrh.logRing.Next()
	if strings.Contains(string(p), reservedFieldName) {
		return os.Stdout.Write(p)
	}
	return len(p), nil
}

// xrayInterceptor is an implementation of sparta.LambdaEventInterceptors that
// handles tapping the event handling workflow and publishing a span with optional
// request information on error.
type xrayInterceptor struct {
	mode XRayInterceptorMode
	//lint:ignore U1000 because it's used
	zerologXRayHandler *zerologXRayHandler
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
