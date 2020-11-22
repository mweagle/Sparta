package sparta

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
)

func TestDescribe(t *testing.T) {
	logger, _ := NewLogger(zerolog.InfoLevel.String())
	output, err := os.Create("./graph.html")
	if nil != err {
		t.Fatalf("Failed to create graph: %s", err.Error())
		return
	}
	defer output.Close()
	err = Describe("SampleService",
		"SampleService Description",
		testLambdaData(),
		nil,
		nil,
		"",
		"",
		"",
		output,
		nil,
		logger)
	if nil != err {
		t.Errorf("Failed to describe: %s", err)
	}
}
