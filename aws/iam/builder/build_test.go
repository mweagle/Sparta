package iambuilder

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	sparta "github.com/mweagle/Sparta"
)

// Set of iamBuilders whose output is required to match the corresponding
// test{N}.json file in the same directory
var iamBuilders = []sparta.IAMRolePrivilege{
	Allow("ssm:GetParameter").
		ForResource().
		Literal("arn:aws:ssm:").
		Region(":").
		AccountID(":").
		Literal("parameter/SpartaHelloWorld-Discovery").
		ToPrivilege(),
	Allow("ssm:GetParameter").
		ForResource().
		Literal("arn:aws:ssm:").
		Region().
		AccountID().
		Literal("parameter/SpartaHelloWorld-Discovery").
		ToPrivilege(),
	Allow("sts:AssumeRole").
		ForPrincipals("ecs-tasks.amazonaws.com").
		ToPrivilege(),
}

func ExampleIAMResourceBuilder_ssm() {
	Allow("ssm:GetParameter").ForResource().
		Literal("arn:aws:ssm:").
		Region(":").
		AccountID(":").
		Literal("parameter/SpartaHelloWorld-Discovery").
		ToPrivilege()
}

func ExampleIAMResourceBuilder_s3() {
	Allow("s3:GetObject").ForResource().
		Literal("arn:aws:s3:::").
		Ref("MyDynamicS3Bucket").
		Literal("/*").
		ToPrivilege()
}

func ExampleIAMResourceBuilder_lambdaarn() {
	Allow("s3:GetObject").ForResource().
		Literal("arn:aws:s3:::").
		Ref("MyDynamicS3Bucket").
		Literal("/*").
		ToPrivilege()
}

func TestIAMBuilder(t *testing.T) {
	for eachIndex, eachBuilder := range iamBuilders {
		testFile := fmt.Sprintf("test%d.json", eachIndex)
		readFile, readFileErr := ioutil.ReadFile(testFile)
		if readFileErr != nil {
			t.Fatalf("Failed to read file: %s", testFile)
		}
		builderJSON, builderJSONErr := json.Marshal(eachBuilder)
		if builderJSONErr != nil {
			t.Fatalf("Failed to marshal JSON : %s", builderJSONErr)
		}
		var expected map[string]interface{}
		expectedUnmarshalErr := json.Unmarshal(readFile, &expected)
		if expectedUnmarshalErr != nil {
			t.Fatalf("Failed to unmarshal JSON : %s", expectedUnmarshalErr)
		}
		var generated map[string]interface{}
		decodedUnmarshalErr := json.Unmarshal(builderJSON, &generated)
		if decodedUnmarshalErr != nil {
			t.Fatalf("Failed to unmarshal JSON : %s", decodedUnmarshalErr)
		}
		equal := reflect.DeepEqual(expected, generated)
		if !equal {
			t.Fatalf("Failed to verify output for test: %d\nGENERATED:%#v\nEXPECTED: %#v",
				eachIndex,
				generated,
				expected)
		}
	}
}
