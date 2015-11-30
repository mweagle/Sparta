package sparta

import (
	"bytes"
	"testing"
)

func TestProvision(t *testing.T) {

	logger, err := NewLogger("info")
	var templateWriter bytes.Buffer
	err = Provision(true, "SampleProvision", "", testLambdaData(), nil, "S3Bucket", &templateWriter, logger)
	if nil != err {
		t.Fatal(err.Error())
	}
}
