package accessor

import (
	"testing"
)

func s3Accessor() KevValueAccessor {
	return &S3Accessor{
		testingBucketName: "weagle-sparta-testbucket",
	}
}

func TestS3PutObject(t *testing.T) {
	testPut(t, s3Accessor())
}

func TestS3PutAllObject(t *testing.T) {
	testPutAll(t, s3Accessor())
}
