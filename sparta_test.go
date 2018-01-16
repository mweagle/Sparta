package sparta

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

type StructHandler1 struct {
}

func (handler *StructHandler1) handler(ctx context.Context,
	props map[string]interface{}) (string, error) {
	return "StructHandler1 handler", nil
}

type StructHandler2 struct {
}

func (handler *StructHandler2) handler(ctx context.Context,
	props map[string]interface{}) (string, error) {
	return "StructHandler2 handler", nil
}

func testLambdaStructData() []*LambdaAWSInfo {
	var lambdaFunctions []*LambdaAWSInfo

	handler1 := &StructHandler1{}
	lambdaFn1 := HandleAWSLambda(LambdaName(handler1.handler),
		handler1.handler,
		LambdaExecuteARN)
	lambdaFunctions = append(lambdaFunctions, lambdaFn1)

	handler2 := &StructHandler2{}
	lambdaFn2 := HandleAWSLambda(LambdaName(handler2.handler),
		handler2.handler,
		LambdaExecuteARN)
	lambdaFunctions = append(lambdaFunctions, lambdaFn2)

	return lambdaFunctions
}

func testLambdaDoubleStructPtrData() []*LambdaAWSInfo {
	var lambdaFunctions []*LambdaAWSInfo

	handler1 := &StructHandler1{}
	lambdaFn1 := HandleAWSLambda(LambdaName(handler1.handler),
		handler1.handler,
		LambdaExecuteARN)
	lambdaFunctions = append(lambdaFunctions, lambdaFn1)

	handler2 := &StructHandler1{}
	lambdaFn2 := HandleAWSLambda(LambdaName(handler2.handler),
		handler2.handler,
		LambdaExecuteARN)
	lambdaFunctions = append(lambdaFunctions, lambdaFn2)

	return lambdaFunctions
}

func userDefinedCustomResource1(requestType string,
	stackID string,
	properties map[string]interface{},
	logger *logrus.Logger) (map[string]interface{}, error) {
	return nil, nil
}

func userDefinedCustomResource2(requestType string,
	stackID string,
	properties map[string]interface{},
	logger *logrus.Logger) (map[string]interface{}, error) {
	return nil, nil
}

func TestStruct(t *testing.T) {
	logger, _ := NewLogger("info")
	var templateWriter bytes.Buffer
	err := Provision(true,
		"SampleProvision",
		"",
		testLambdaStructData(),
		nil,
		nil,
		os.Getenv("S3_BUCKET"),
		false,
		false,
		"testBuildID",
		"",
		"",
		"",
		&templateWriter,
		nil,
		logger)
	if nil != err {
		t.Fatal(err.Error())
	}
}

func TestDoubleRefStruct(t *testing.T) {
	logger, _ := NewLogger("info")
	var templateWriter bytes.Buffer
	err := Provision(true,
		"SampleProvision",
		"",
		testLambdaDoubleStructPtrData(),
		nil,
		nil,
		os.Getenv("S3_BUCKET"),
		false,
		false,
		"testBuildID",
		"",
		"",
		"",
		&templateWriter,
		nil,
		logger)

	if nil == err {
		t.Fatal("Failed to enforce lambda function uniqueness")
	}
}

func TestCustomResource(t *testing.T) {
	logger, _ := NewLogger("info")
	lambdaFuncs := testLambdaStructData()
	lambdaFuncs[0].RequireCustomResource(IAMRoleDefinition{},
		userDefinedCustomResource1,
		nil,
		nil)

	lambdaFuncs[1].RequireCustomResource(IAMRoleDefinition{},
		userDefinedCustomResource2,
		nil,
		nil)

	var templateWriter bytes.Buffer
	err := Provision(true,
		"SampleProvision",
		"",
		lambdaFuncs,
		nil,
		nil,
		os.Getenv("S3_BUCKET"),
		false,
		false,
		"testBuildID",
		"",
		"",
		"",
		&templateWriter,
		nil,
		logger)

	if nil != err {
		t.Fatal("Failed to accept unique user CustomResource functions")
	}
}

func TestDoubleRefCustomResource(t *testing.T) {
	logger, _ := NewLogger("info")
	lambdaFuncs := testLambdaStructData()

	for _, eachLambda := range lambdaFuncs {
		eachLambda.RequireCustomResource(IAMRoleDefinition{},
			userDefinedCustomResource1,
			nil,
			nil)
	}
	var templateWriter bytes.Buffer
	err := Provision(true,
		"SampleProvision",
		"",
		lambdaFuncs,
		nil,
		nil,
		os.Getenv("S3_BUCKET"),
		false,
		false,
		"testBuildID",
		"",
		"",
		"",
		&templateWriter,
		nil,
		logger)

	if nil == err {
		t.Fatal("Failed to reject duplicate user CustomResource functions")
	}
}

func TestSignatureVersion(t *testing.T) {
	logger, _ := NewLogger("info")

	lambdaFunctions := testLambdaDoubleStructPtrData()
	lambdaFunctions[0].Options = &LambdaFunctionOptions{
		SpartaOptions: &SpartaOptions{
			Name: fmt.Sprintf("Handler0"),
		},
	}
	lambdaFunctions[1].Options = &LambdaFunctionOptions{
		SpartaOptions: &SpartaOptions{
			Name: fmt.Sprintf("Handler1"),
		},
	}
	var templateWriter bytes.Buffer
	err := Provision(true,
		"TestOverlappingLambdas",
		"",
		lambdaFunctions,
		nil,
		nil,
		os.Getenv("S3_BUCKET"),
		false,
		false,
		"testBuildID",
		"",
		"",
		"",
		&templateWriter,
		nil,
		logger)

	if nil != err {
		t.Fatal("Failed to respect duplicate lambdas with user supplied names")
	} else {
		t.Logf("Rejected duplicate lambdas")
	}
}

func TestUserDefinedOverlappingLambdaNames(t *testing.T) {
	logger, _ := NewLogger("info")

	lambdaFunctions := testLambdaDoubleStructPtrData()
	for _, eachLambda := range lambdaFunctions {
		eachLambda.Options = &LambdaFunctionOptions{
			SpartaOptions: &SpartaOptions{
				Name: fmt.Sprintf("HandlerX"),
			},
		}
	}

	var templateWriter bytes.Buffer
	err := Provision(true,
		"TestOverlappingLambdas",
		"",
		lambdaFunctions,
		nil,
		nil,
		os.Getenv("S3_BUCKET"),
		false,
		false,
		"testBuildID",
		"",
		"",
		"",
		&templateWriter,
		nil,
		logger)

	if nil == err {
		t.Fatal("Failed to reject duplicate lambdas with overlapping user supplied names")
	} else {
		t.Logf("Rejected overlapping user supplied names")
	}
}

func invalidFuncSignature(ctx context.Context) string {
	return "Hello World!"
}

func TestInvalidFunctionSignature(t *testing.T) {
	logger, _ := NewLogger("info")

	var lambdaFunctions []*LambdaAWSInfo
	invalidSigHandler := HandleAWSLambda("InvalidSignature",
		invalidFuncSignature,
		IAMRoleDefinition{})
	lambdaFunctions = append(lambdaFunctions, invalidSigHandler)

	var templateWriter bytes.Buffer
	err := Provision(true,
		"TestInvalidFunctionSignatuure",
		"",
		lambdaFunctions,
		nil,
		nil,
		os.Getenv("S3_BUCKET"),
		false,
		false,
		"testBuildID",
		"",
		"",
		"",
		&templateWriter,
		nil,
		logger)

	if err == nil {
		t.Fatal("Failed to reject invalid lambda function signature")
	} else {
		t.Log("Properly rejected invalid function signature")
	}
}
