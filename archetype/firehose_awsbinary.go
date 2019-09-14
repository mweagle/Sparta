// +build lambdabinary

package archetype

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	awsEvents "github.com/aws/aws-lambda-go/events"
)

var (
	xformTemplateBytes []byte
	xformError error
)

func init() {
	xformPath := os.Getenv(envVarKinesisFirehoseTransformName)
	fmt.Printf("Reading template file: %s\n", xformPath)
	templateBytes, templateBytesErr := ioutil.ReadFile(xformPath)
	if templateBytesErr != nil {
		xformError = templateBytesErr
		return
	}
	xformTemplateBytes = templateBytes
}

// Great, transform everything
func lambdaXForm(ctx context.Context, kinesisEvent awsEvents.KinesisFirehoseEvent) (*awsEvents.KinesisFirehoseResponse, error) {

	if xformError != nil {
		return nil, xformError
	}
	return ApplyTransformToKinesisFirehoseEvent(ctx, xformTemplateBytes, kinesisEvent)
}
