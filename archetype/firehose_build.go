// +build !lambdabinary

package archetype

import (
	"context"

	awsEvents "github.com/aws/aws-lambda-go/events"
)

// The core lambda transformation function
func lambdaXForm(ctx context.Context,
	kinesisEvent awsEvents.KinesisFirehoseEvent) (*awsEvents.KinesisFirehoseResponse, error) {

	// NOP...
	return nil, nil
}
