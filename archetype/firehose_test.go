package archetype

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	awsEvents "github.com/aws/aws-lambda-go/events"
	awsEventsTest "github.com/aws/aws-lambda-go/events/test"
	"github.com/pkg/errors"
)

var firehoseTests = []struct {
	sourceInputPath string
	templatePath    string
	predicate       testPredicate
}{
	{
		"test/records.json",
		"test/testdata.transform",
		okPredicate,
	},
	{
		"test/records-sm.json",
		"test/drop.transform",
		verifyPredicate("Dropped", 3),
	},
	{
		"test/records.json",
		"test/conditional.transform",
		okPredicate,
	},
	{
		"test/records-sm.json",
		"test/conditional.transform",
		verifyPredicate("Dropped", 2),
	},
}

type testPredicate func(t *testing.T, response *awsEvents.KinesisFirehoseResponse) error

func okPredicate(t *testing.T, response *awsEvents.KinesisFirehoseResponse) error {

	for _, eachEntry := range response.Records {
		jsonMap := make(map[string]interface{})
		unmarshalErr := json.Unmarshal(eachEntry.Data, &jsonMap)
		if unmarshalErr != nil {
			return unmarshalErr
		}
		//t.Logf("Record: %#v\n", jsonMap)
	}
	return nil
}

func verifyPredicate(value string, expectedCount int) testPredicate {
	return func(t *testing.T, response *awsEvents.KinesisFirehoseResponse) error {
		counter := 0
		jsonMap := make(map[string]interface{})
		for _, eachEntry := range response.Records {
			//t.Logf("Record: %#v\n", string(eachEntry.Data))
			unmarshalErr := json.Unmarshal(eachEntry.Data, &jsonMap)
			if unmarshalErr != nil {
				continue
			}
			recordVal := fmt.Sprintf("%#v", jsonMap)
			counter += strings.Count(recordVal, value)
		}
		jsonData, _ := json.MarshalIndent(response, "", " ")
		// t.Log(string(jsonData))
		counter += strings.Count(string(jsonData), value)

		if counter != expectedCount {
			return errors.Errorf("Invalid count. Expected: %d, Found: %d for value: %s",
				expectedCount,
				counter,
				value)
		}
		return nil
	}
}

func testData(t *testing.T, inputPath string) *awsEvents.KinesisFirehoseEvent {

	// 1. read JSON from file
	inputJSON := awsEventsTest.ReadJSONFromFile(t, inputPath)

	// 2. de-serialize into Go object
	var inputEvent awsEvents.KinesisFirehoseEvent
	if err := json.Unmarshal(inputJSON, &inputEvent); err != nil {
		t.Fatalf("could not unmarshal event. details: %v", err)
		return nil
	}
	return &inputEvent
}

func testTemplate(t *testing.T, xformPath string) []byte {
	templateBytes, templateBytesErr := ioutil.ReadFile(xformPath)
	if templateBytesErr != nil {
		t.Fatalf("Could not read template data. Error: %v", templateBytesErr)
		return nil
	}
	return templateBytes
}

func TestTransforms(t *testing.T) {

	for _, tt := range firehoseTests {
		t.Run(tt.templatePath, func(t *testing.T) {
			data := testData(t, tt.sourceInputPath)
			xformTemplate := testTemplate(t, tt.templatePath)
			ctx := context.Background()
			response, responseErr := ApplyTransformToKinesisFirehoseEvent(ctx,
				xformTemplate,
				*data)
			if responseErr != nil {
				t.Fatal(responseErr)
				return
			}

			// Unmarshal everything...
			predicateErr := tt.predicate(t, response)
			if predicateErr != nil {
				t.Fatal(predicateErr)
				return
			}
		})
	}
}

func TestLambdaTransform(t *testing.T) {
	lambdaTransform := func(ctx context.Context,
		kinesisRecord *awsEvents.KinesisFirehoseEventRecord) (*awsEvents.KinesisFirehoseResponseRecord, error) {

		return &awsEvents.KinesisFirehoseResponseRecord{
			RecordID: kinesisRecord.RecordID,
			Result:   awsEvents.KinesisFirehoseTransformedStateOk,
			Data:     kinesisRecord.Data,
		}, nil
	}
	ctx := context.Background()

	for _, tt := range firehoseTests {
		t.Run(tt.templatePath, func(t *testing.T) {
			data := testData(t, tt.sourceInputPath)

			for _, eachRecord := range data.Records {
				xformed, xformedErr := lambdaTransform(ctx, &eachRecord)
				if xformedErr != nil {
					t.Fatalf("Failed to transform")
				}
				if xformed.Result != awsEvents.KinesisFirehoseTransformedStateOk {
					t.Fatalf("Failed to successful process record")
				}
			}
		})
	}

}
