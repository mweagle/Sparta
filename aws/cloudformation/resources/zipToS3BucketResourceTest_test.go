package resources

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	gocf "github.com/mweagle/go-cloudformation"
	"github.com/rs/zerolog"
)

func testEnabled() bool {
	return os.Getenv("TEST_SRC_S3_KEY") != ""
}
func mockZipResourceEvent(t *testing.T) *CloudFormationLambdaEvent {
	props := map[string]interface{}{
		"DestBucket": gocf.String(os.Getenv("TEST_DEST_S3_BUCKET")),
		"SrcBucket":  gocf.String(os.Getenv("TEST_SRC_S3_BUCKET")),
		"SrcKeyName": gocf.String(os.Getenv("TEST_SRC_S3_KEY")),
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
	resUnzip := gocf.NewResourceByType(ZipToS3Bucket)
	zipResource := resUnzip.(*ZipToS3BucketResource)
	event := mockZipResourceEvent(t)

	// Put it
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
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
