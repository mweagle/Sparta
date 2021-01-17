package sparta

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"

	"github.com/rs/zerolog"
)

var discoveryDataNoTags = `
{
	"ResourceID": "mainhelloWorldGETLambda3d6fd4fce31e46927fb59e0cfe2f20461a69490a",
	"Region": "us-west-2",
	"StackID": "arn:aws:cloudformation:us-west-2:123412341234:stack/SpartaDDB-mweagle/c4ada6d0-d697-11e7-9b91-50d5ca789e82",
	"StackName": "SpartaDDB-mweagle",
	"Resources": {
			"DynamoDBad8db2fc80a1af0b5bacfbc66b5ae671301d5e96": {
					"ResourceID": "DynamoDBad8db2fc80a1af0b5bacfbc66b5ae671301d5e96",
					"ResourceRef": "SpartaDDB-mweagle-DynamoDBad8db2fc80a1af0b5bacfbc66b5ae671301d5e96-1EU295I6O4XJH",
					"ResourceType": "AWS::DynamoDB::Table",
					"Properties": {
							"StreamArn": "arn:aws:dynamodb:us-west-2:123412341234:table/SpartaDDB-mweagle-DynamoDBad8db2fc80a1af0b5bacfbc66b5ae671301d5e96-1EU295I6O4XJH/stream/2017-12-03T15:37:38.943"
					}
			}
	}
}
`

func TestDiscoveryInitialized(t *testing.T) {
	// Ensure that sparta.Discover() can only be called from a lambda function
	logger, _ := NewLogger(zerolog.WarnLevel.String())

	// Encode the data, stuff it into the environment variable
	encodedString := base64.StdEncoding.EncodeToString([]byte(discoveryDataNoTags))
	os.Setenv(envVarDiscoveryInformation, encodedString)

	// Initialize the data
	initializeDiscovery(logger)

	configuration, err := Discover()
	t.Logf("Configuration: %#v", configuration)
	t.Logf("Error: %#v", err)
	if err != nil {
		t.Errorf("sparta.Discover() failed to initialize from environment")
	}
	t.Logf("Properly unmarshaled environment data")
}

func TestDiscoveryNotInitialized(t *testing.T) {
	configuration, err := Discover()
	t.Logf("Configuration: %#v", configuration)
	t.Logf("Error: %#v", err)
	if err != nil {
		t.Errorf("sparta.Discover() failed to error when not initialized")
	}
	t.Logf("Properly rejected unintialized discovery data")
}

func TestDiscoveryUnmarshalNoTags(t *testing.T) {
	// Ensure that sparta.Discover() can only be called from a lambda function
	var info DiscoveryInfo
	err := json.Unmarshal([]byte(discoveryDataNoTags), &info)
	if nil != err {
		t.Errorf("Failed to unmarshal discovery data without tags")
	}
	if len(info.Resources) != 1 {
		t.Errorf("Failed to unmarshal single resource")
	}
	t.Logf("Discovery Info: %#v", info)
}

func TestDiscoveryEmptyMetadata(t *testing.T) {
	// Ensure that sparta.Discover() can only be called from a lambda function
	var info DiscoveryInfo
	err := json.Unmarshal([]byte("{}"), &info)
	if nil != err {
		t.Errorf("Failed to unmarshal empty discovery data")
	}
	t.Logf("Discovery Info: %#v", info)
}
