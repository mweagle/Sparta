package sparta

import (
	"bytes"
	"testing"

	"github.com/Sirupsen/logrus"
)

func TestProvision(t *testing.T) {

	logger, err := NewLogger("info")
	var templateWriter bytes.Buffer
	err = Provision(true, "SampleProvision", "", testLambdaData(), nil, "S3Bucket", &templateWriter, logger)
	if nil != err {
		t.Fatal(err.Error())
	}
}

func templateDecorator(lambdaResourceName string,
	lambdaResourceDefinition ArbitraryJSONObject,
	resources ArbitraryJSONObject,
	outputs ArbitraryJSONObject,
	logger *logrus.Logger) error {

	// Add a resource
	resources["OutputResourceTest"] = ArbitraryJSONObject{
		"Type":    "AWS::CloudFormation::CustomResource",
		"Version": "1.0",
		"Properties": ArbitraryJSONObject{
			"ServiceToken": "arn:aws:sns:us-east-1:84969EXAMPLE:CRTest",
			"key1":         "string",
		},
	}

	// Add an output
	outputs["OutputDecorationTest"] = ArbitraryJSONObject{
		"Description": "Information about the value",
		"Value":       "Value to return",
	}
	return nil
}

func TestDecorateProvision(t *testing.T) {

	lambdas := testLambdaData()
	lambdas[0].Decorator = templateDecorator

	logger, err := NewLogger("info")
	var templateWriter bytes.Buffer
	err = Provision(true, "SampleProvision", "", lambdas, nil, "S3Bucket", &templateWriter, logger)
	if nil != err {
		t.Fatal(err.Error())
	}
}
