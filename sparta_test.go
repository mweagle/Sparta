package sparta

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	gof "github.com/awslabs/goformation/v5/cloudformation"
	gofs3 "github.com/awslabs/goformation/v5/cloudformation/s3"
	spartaCFResources "github.com/mweagle/Sparta/aws/cloudformation/resources"
	"github.com/rs/zerolog"
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
	lambdaFn1, _ := NewAWSLambda(LambdaName(handler1.handler),
		handler1.handler,
		lambdaTestExecuteARN)
	lambdaFunctions = append(lambdaFunctions, lambdaFn1)

	handler2 := &StructHandler2{}
	lambdaFn2, _ := NewAWSLambda(LambdaName(handler2.handler),
		handler2.handler,
		lambdaTestExecuteARN)
	lambdaFunctions = append(lambdaFunctions, lambdaFn2)

	return lambdaFunctions
}

func testLambdaDoubleStructPtrData() []*LambdaAWSInfo {
	var lambdaFunctions []*LambdaAWSInfo

	handler1 := &StructHandler1{}
	lambdaFn1, _ := NewAWSLambda(LambdaName(handler1.handler),
		handler1.handler,
		lambdaTestExecuteARN)
	lambdaFunctions = append(lambdaFunctions, lambdaFn1)

	handler2 := &StructHandler1{}
	lambdaFn2, _ := NewAWSLambda(LambdaName(handler2.handler),
		handler2.handler,
		lambdaTestExecuteARN)
	lambdaFunctions = append(lambdaFunctions, lambdaFn2)

	return lambdaFunctions
}

func userDefinedCustomResource1(ctx context.Context,
	event spartaCFResources.CloudFormationLambdaEvent) (map[string]interface{}, error) {
	return nil, nil
}

func userDefinedCustomResource2(ctx context.Context,
	event spartaCFResources.CloudFormationLambdaEvent) (map[string]interface{}, error) {
	return nil, nil
}

func TestStruct(t *testing.T) {
	testProvision(t, testLambdaStructData(), nil)
}

func TestDoubleRefStruct(t *testing.T) {
	testProvision(t,
		testLambdaDoubleStructPtrData(),
		assertError("Failed to reject struct exporting duplicate targets"))
}

func TestCustomResource(t *testing.T) {
	lambdaFuncs := testLambdaStructData()
	_, reqErr := lambdaFuncs[0].RequireCustomResource(IAMRoleDefinition{},
		userDefinedCustomResource1,
		nil,
		nil)
	if reqErr != nil {
		t.Fatalf("Failed to include custom resource: %s", reqErr.Error())
	}
	_, reqErr2 := lambdaFuncs[1].RequireCustomResource(IAMRoleDefinition{},
		userDefinedCustomResource2,
		nil,
		nil)
	if reqErr2 != nil {
		t.Fatalf("Failed to include custom resource: %s", reqErr2.Error())
	}
	testProvision(t, lambdaFuncs, nil)
}

func TestDoubleRefCustomResource(t *testing.T) {
	lambdaFuncs := testLambdaStructData()

	for _, eachLambda := range lambdaFuncs {
		_, reqErr := eachLambda.RequireCustomResource(IAMRoleDefinition{},
			userDefinedCustomResource1,
			nil,
			nil)
		if reqErr != nil {
			t.Fatalf("Failed to require custom resource: %s", reqErr.Error())
		}
	}
	testProvision(t,
		lambdaFuncs,
		assertError("Failed to reject multiply included custom resource"))
}

func TestSignatureVersion(t *testing.T) {
	lambdaFunctions := testLambdaDoubleStructPtrData()
	lambdaFunctions[0].Options = &LambdaFunctionOptions{
		ExtendedOptions: &ExtendedOptions{
			Name: "Handler0",
		},
	}
	lambdaFunctions[1].Options = &LambdaFunctionOptions{
		ExtendedOptions: &ExtendedOptions{
			Name: "Handler1",
		},
	}
	testProvision(t,
		lambdaFunctions,
		nil)
}

func TestUserDefinedOverlappingLambdaNames(t *testing.T) {
	lambdaFunctions := testLambdaDoubleStructPtrData()
	for _, eachLambda := range lambdaFunctions {
		eachLambda.Options = &LambdaFunctionOptions{
			ExtendedOptions: &ExtendedOptions{
				Name: "HandlerX",
			},
		}
	}
	testProvision(t,
		lambdaFunctions,
		assertError("Failed to reject duplicate lambdas with overlapping user supplied names"))
}

func invalidFuncSignature(ctx context.Context) string {
	return "Hello World!"
}

func TestInvalidFunctionSignature(t *testing.T) {
	var lambdaFunctions []*LambdaAWSInfo
	invalidSigHandler, _ := NewAWSLambda("InvalidSignature",
		invalidFuncSignature,
		IAMRoleDefinition{})
	lambdaFunctions = append(lambdaFunctions, invalidSigHandler)

	testProvision(t,
		lambdaFunctions,
		assertError("Failed to reject invalid lambda function signature"))
}

func TestNOP(t *testing.T) {
	template := gof.NewTemplate()
	s3Resources := &gofs3.Bucket{
		BucketEncryption: &gofs3.Bucket_BucketEncryption{
			ServerSideEncryptionConfiguration: []gofs3.Bucket_ServerSideEncryptionRule{
				gofs3.Bucket_ServerSideEncryptionRule{
					ServerSideEncryptionByDefault: &gofs3.Bucket_ServerSideEncryptionByDefault{
						KMSMasterKeyID: "SomeKey",
					},
				},
				gofs3.Bucket_ServerSideEncryptionRule{
					ServerSideEncryptionByDefault: &gofs3.Bucket_ServerSideEncryptionByDefault{
						KMSMasterKeyID: "SomeOtherKey",
					},
				},
			},
		},
	}
	template.Resources["S3Bucket"] = s3Resources
	yaml, _ := template.YAML()
	fmt.Printf("\n%s\n", string(yaml))
}

func TestGlobalTransform(t *testing.T) {
	transformName := fmt.Sprintf("Echo%d", time.Now().Unix())
	template := gof.NewTemplate()
	s3Resources := &gofs3.Bucket{
		BucketEncryption: &gofs3.Bucket_BucketEncryption{
			ServerSideEncryptionConfiguration: []gofs3.Bucket_ServerSideEncryptionRule{
				gofs3.Bucket_ServerSideEncryptionRule{
					ServerSideEncryptionByDefault: &gofs3.Bucket_ServerSideEncryptionByDefault{
						KMSMasterKeyID: "SomeKey",
					},
				},
				gofs3.Bucket_ServerSideEncryptionRule{
					ServerSideEncryptionByDefault: &gofs3.Bucket_ServerSideEncryptionByDefault{
						KMSMasterKeyID: "SomeOtherKey",
					},
				},
			},
		},
	}
	template.Resources["S3Bucket"] = s3Resources
	xform := "Transform"
	template.Transform = &gof.Transform{
		String: &xform,
	}
	yaml, _ := template.YAML()
	output := string(yaml)
	fmt.Printf("\n%s\n", string(yaml))

	if !strings.Contains(output, transformName) {
		t.Fatalf("Failed to find global Transform in template")
	}
}

func TestProvisionID(t *testing.T) {
	logger, _ := NewLogger(zerolog.InfoLevel.String())
	testUserValues := []string{
		"",
		"DEFAULT_VALUE",
	}
	for _, eachTestValue := range testUserValues {
		buildID, buildIDErr := computeBuildID(eachTestValue, logger)
		if buildIDErr != nil {
			t.Fatalf("Failed to compute buildID: %s", buildIDErr)
		}
		if eachTestValue == "" && buildID == "" {
			t.Fatalf("Failed to extract buildID. User: %s, Computed: %s", eachTestValue, buildID)
		}
		if eachTestValue != "" &&
			buildID != eachTestValue {
			t.Fatalf("Failed to roundTrip buildID. User: %s, Computed: %s", eachTestValue, buildID)
		}
	}

}
