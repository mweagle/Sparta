package accessor

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-xray-sdk-go/xray"
)

// Conditionally attach XRay if it seems like the right thing to do
func xrayInit(awsClient *client.Client) {
	if os.Getenv("AWS_XRAY_DAEMON_ADDRESS") != "" &&
		os.Getenv("AWS_EXECUTION_ENV") != "" {
		xray.AWS(awsClient)
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
