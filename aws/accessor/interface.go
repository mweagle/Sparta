package accessor

import (
	"context"
	"os"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	xrayv2 "github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"
)

// Conditionally attach XRay if it seems like the right thing to do
func xrayInit(awsConfig *awsv2.Config) {
	if os.Getenv("AWS_XRAY_DAEMON_ADDRESS") != "" &&
		os.Getenv("AWS_EXECUTION_ENV") != "" {
		xrayv2.AWSV2Instrumentor(&awsConfig.APIOptions)
	}
}

// NewObjectConstructor returns a fresh instance
// of the type that's stored in the KV store
type NewObjectConstructor func() interface{}

// KevValueAccessor represents a simple KV store
type KevValueAccessor interface {
	Delete(ctx context.Context, keyPath string) error
	DeleteAll(ctx context.Context) error
	Put(ctx context.Context, keyPath string, object interface{}) error
	Get(ctx context.Context, keyPath string, object interface{}) error
	GetAll(ctx context.Context, ctor NewObjectConstructor) ([]interface{}, error)
}
