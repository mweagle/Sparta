package accessor

import (
	"testing"
)

func TestS3PutObject(t *testing.T) {
	s3Accessor := &S3Accessor{
		testingBucketName: "weagle-sparta-testbucket",
	}
	testPut(t, s3Accessor)
}

func TestS3PutAllObject(t *testing.T) {
	s3Accessor := &S3Accessor{
		testingBucketName: "weagle-sparta-testbucket",
	}
	testPutAll(t, s3Accessor)
}
