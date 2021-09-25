package resources

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	cwCustomProvider "github.com/mweagle/Sparta/aws/cloudformation/provider"

	"github.com/rs/zerolog"
)

func testEnabled() bool {
	return os.Getenv("TEST_SRC_S3_KEY") != ""
}
func mockZipResourceEvent(t *testing.T) *CloudFormationLambdaEvent {
	props := map[string]interface{}{
		"DestBucket": os.Getenv("TEST_DEST_S3_BUCKET"),
		"SrcBucket":  os.Getenv("TEST_SRC_S3_BUCKET"),
		"SrcKeyName": os.Getenv("TEST_SRC_S3_KEY"),
		"Manifest": map[string]interface{}{
			"Some": "Data",
		},
	}
	bytes, bytesErr := json.Marshal(props)
	if bytesErr != nil {
		t.Fatalf("Failed to serialize mock custom resource event")
	}

	return &CloudFormationLambdaEvent{
		RequestType:        CreateOperation,
		RequestID:          time.Now().String(),
		StackID:            "1234567890",
		LogicalResourceID:  "logicalID",
		ResourceProperties: json.RawMessage(bytes),
	}
}

func TestUnzip(t *testing.T) {
	if !testEnabled() {
		return
	}
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	resUnzip, _ := cwCustomProvider.NewCloudFormationCustomResource(ZipToS3Bucket, &logger)
	zipResource := resUnzip.(*ZipToS3BucketResource)
	event := mockZipResourceEvent(t)

	// Put it
	awsSession := awsSession(&logger)
	createOutputs, createError := zipResource.Create(awsSession,
		event,
		&logger)
	if nil != createError {
		t.Errorf("Failed to create Unzip resource: %s", createError)
	}
	t.Logf("TestUnzip outputs: %#v", createOutputs)

	deleteOutputs, deleteError := zipResource.Delete(awsSession,
		event,
		&logger)
	if nil != deleteError {
		t.Errorf("Failed to create Unzip resource: %s", createError)
	}
	t.Logf("TestUnzip outputs: %#v", deleteOutputs)
}
