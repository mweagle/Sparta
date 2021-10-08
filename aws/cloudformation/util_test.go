package cloudformation

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	gofintrinsics "github.com/awslabs/goformation/v5/intrinsics"
	"github.com/go-test/deep"

	spartaAWS "github.com/mweagle/Sparta/aws"
	"github.com/rs/zerolog"
)

var conversionParams = map[string]interface{}{
	"Key1": "Value1",
	"Key2": "Value2",
}

var userdataPassingTests = []struct {
	input  string
	output []interface{}
}{
	{
		"HelloWorld",
		[]interface{}{
			`HelloWorld`,
		},
	},
	{
		"Hello {{ .Key1 }}",
		[]interface{}{
			`Hello Value1`,
		},
	},
	{
		`{{ .Key1 }}=={{ .Key2 }}`,
		[]interface{}{
			`Value1==Value2`,
		},
	},
	{
		`A { "Fn::GetAtt" : [ "ResName" , "AttrName" ] }`,
		[]interface{}{
			`A `,
			map[string]interface{}{
				"Fn::GetAtt": []interface{}{"ResName", "AttrName"},
			},
		},
	},
	{
		`A { "Fn::GetAtt" : [ "ResName" , "AttrName" ] } B`,
		[]interface{}{
			`A `,
			map[string]interface{}{
				"Fn::GetAtt": []interface{}{"ResName", "AttrName"},
			},
			` B`,
		},
	},
	{
		`{ "Fn::GetAtt" : [ "ResName" , "AttrName" ] }
A`,
		[]interface{}{
			map[string]interface{}{
				"Fn::GetAtt": []interface{}{"ResName", "AttrName"},
			},
			"\n",
			"A",
		},
	},
	{
		`{"Ref": "AWS::Region"}`,
		[]interface{}{
			map[string]interface{}{
				"Ref": "AWS::Region",
			},
		},
	},
	{
		`A {"Ref" : "AWS::Region"} B`,
		[]interface{}{
			"A ",
			map[string]interface{}{
				"Ref": "AWS::Region",
			},
			" B",
		},
	},
	{
		`A
{"Ref" : "AWS::Region"}
B`,
		[]interface{}{
			"A\n",
			map[string]interface{}{
				"Ref": "AWS::Region",
			},
			"\n",
			"B",
		},
	},
	{
		"{\"Ref\" : \"AWS::Region\"} = {\"Ref\" : \"AWS::AccountId\"}",
		[]interface{}{
			map[string]interface{}{
				"Ref": "AWS::Region",
			},
			" = ",
			map[string]interface{}{
				"Ref": "AWS::AccountId",
			},
		},
	},
}

/*
   "Fn::GetAtt" : []string{"ResName","AttrName"},
*/
func TestExpand(t *testing.T) {
	for _, eachTest := range userdataPassingTests {
		testReader := strings.NewReader(eachTest.input)
		expandResult, expandResultErr := ConvertToTemplateExpression(testReader, conversionParams)
		if nil != expandResultErr {
			t.Errorf("%s (Input: %s)", expandResultErr, eachTest.input)
		} else {
			testOutput := map[string]interface{}{
				"Fn::Join": []interface{}{
					"",
					eachTest.output,
				},
			}
			rawMarshal, rawMarshalErr := json.Marshal(expandResult)
			if nil != rawMarshalErr {
				t.Fatalf("%s (Input: %s)", rawMarshalErr, eachTest.input)
			}
			actualMarshal, actualMarshalErr := gofintrinsics.ProcessJSON(rawMarshal, nil)
			if actualMarshalErr != nil {
				t.Fatalf("%s (Input: %s)", actualMarshalErr, rawMarshal)
			}
			var expandedActual map[string]interface{}
			unmarshalErr := json.Unmarshal(actualMarshal, &expandedActual)
			if unmarshalErr != nil {
				t.Fatalf("%s (Input: %s)", unmarshalErr, actualMarshal)
			}
			diffResult := deep.Equal(testOutput, expandedActual)
			if diffResult != nil {
				t.Errorf("Failed to validate\n")
				t.Errorf("\tEXPECTED: %#v\n", testOutput)
				t.Errorf("\tACTUAL: %#v\n", expandedActual)
				t.Errorf("DIFF: %s", diffResult)
			} else {
				t.Logf("Validated: %v == %v", testOutput, expandedActual)
			}
		}
	}
}

func TestUserScopedStackName(t *testing.T) {
	stackName := UserScopedStackName("TestingService")
	if stackName == "" {
		t.Fatalf("Failed to get `user` scoped name for Stack")
	}
}
func TestPlatformScopedName(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	awsConfig := spartaAWS.NewConfig(&logger)
	stackName, stackNameErr := UserAccountScopedStackName("TestService", awsConfig)
	if stackNameErr != nil {
		t.Fatalf("Failed to create AWS account based stack name: %s", stackNameErr)
	}
	if stackName == "" {
		t.Fatalf("Failed to get `user` AWS account name for Stack")
	}
}
