package sparta

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	spartaCFResources "github.com/mweagle/Sparta/aws/cloudformation/resources"
	gocf "github.com/mweagle/go-cloudformation"
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
	lambdaFuncs[0].RequireCustomResource(IAMRoleDefinition{},
		userDefinedCustomResource1,
		nil,
		nil)

	lambdaFuncs[1].RequireCustomResource(IAMRoleDefinition{},
		userDefinedCustomResource2,
		nil,
		nil)
	testProvision(t, lambdaFuncs, nil)
}

func TestDoubleRefCustomResource(t *testing.T) {
	lambdaFuncs := testLambdaStructData()

	for _, eachLambda := range lambdaFuncs {
		eachLambda.RequireCustomResource(IAMRoleDefinition{},
			userDefinedCustomResource1,
			nil,
			nil)
	}
	testProvision(t,
		lambdaFuncs,
		assertError("Failed to reject multiply included custom resource"))
}

func TestSignatureVersion(t *testing.T) {
	lambdaFunctions := testLambdaDoubleStructPtrData()
	lambdaFunctions[0].Options = &LambdaFunctionOptions{
		SpartaOptions: &SpartaOptions{
			Name: "Handler0",
		},
	}
	lambdaFunctions[1].Options = &LambdaFunctionOptions{
		SpartaOptions: &SpartaOptions{
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
			SpartaOptions: &SpartaOptions{
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
	template := gocf.NewTemplate()
	s3Resources := gocf.S3Bucket{
		BucketEncryption: &gocf.S3BucketBucketEncryption{
			ServerSideEncryptionConfiguration: &gocf.S3BucketServerSideEncryptionRuleList{
				gocf.S3BucketServerSideEncryptionRule{
					ServerSideEncryptionByDefault: &gocf.S3BucketServerSideEncryptionByDefault{
						KMSMasterKeyID: gocf.String("SomeKey"),
					},
				},
				gocf.S3BucketServerSideEncryptionRule{
					ServerSideEncryptionByDefault: &gocf.S3BucketServerSideEncryptionByDefault{
						KMSMasterKeyID: gocf.String("SomeOtherKey"),
					},
				},
			},
		},
	}
	template.AddResource("S3Bucket", s3Resources)
	json, _ := json.MarshalIndent(template, "", " ")
	fmt.Printf("\n%s\n", string(json))
}

func TestGlobalTransform(t *testing.T) {
	transformName := fmt.Sprintf("Echo%d", time.Now().Unix())
	template := gocf.NewTemplate()
	s3Resources := gocf.S3Bucket{
		BucketEncryption: &gocf.S3BucketBucketEncryption{
			ServerSideEncryptionConfiguration: &gocf.S3BucketServerSideEncryptionRuleList{
				gocf.S3BucketServerSideEncryptionRule{
					ServerSideEncryptionByDefault: &gocf.S3BucketServerSideEncryptionByDefault{
						KMSMasterKeyID: gocf.String("SomeKey"),
					},
				},
				gocf.S3BucketServerSideEncryptionRule{
					ServerSideEncryptionByDefault: &gocf.S3BucketServerSideEncryptionByDefault{
						KMSMasterKeyID: gocf.String("SomeOtherKey"),
					},
				},
			},
		},
	}
	template.AddResource("S3Bucket", s3Resources)
	template.Transform = []string{transformName}
	json, _ := json.MarshalIndent(template, "", " ")
	output := string(json)
	fmt.Printf("\n%s\n", output)

	if !strings.Contains(output, transformName) {
		t.Fatalf("Failed to find global Transform in template")
	}
}

func TestResourceTransform(t *testing.T) {
	transformName := fmt.Sprintf("Echo%d", time.Now().Unix())
	template := gocf.NewTemplate()
	s3Resources := gocf.S3Bucket{
		BucketEncryption: &gocf.S3BucketBucketEncryption{
			ServerSideEncryptionConfiguration: &gocf.S3BucketServerSideEncryptionRuleList{
				gocf.S3BucketServerSideEncryptionRule{
					ServerSideEncryptionByDefault: &gocf.S3BucketServerSideEncryptionByDefault{
						KMSMasterKeyID: gocf.String("SomeKey"),
					},
				},
				gocf.S3BucketServerSideEncryptionRule{
					ServerSideEncryptionByDefault: &gocf.S3BucketServerSideEncryptionByDefault{
						KMSMasterKeyID: gocf.String("SomeOtherKey"),
					},
				},
			},
		},
	}
	bucketResource := template.AddResource("S3Bucket", s3Resources)
	bucketResource.Transform = &gocf.FnTransform{
		Name: gocf.String(transformName),
		Parameters: map[string]interface{}{
			"SomeValue": gocf.Integer(42),
		},
	}

	template.Transform = []string{transformName}
	json, _ := json.MarshalIndent(template, "", " ")
	output := string(json)
	fmt.Printf("\n%s\n", output)

	if !strings.Contains(output, transformName) {
		t.Fatalf("Failed to find resource Transform in template")
	}
	if !strings.Contains(output, "SomeValue") {
		t.Fatalf("Failed to find resource Parameters in template")
	}
}
func TestProvisionID(t *testing.T) {
	logger, _ := NewLogger("info")
	testUserValues := []string{
		"",
		"DEFAULT_VALUE",
	}
	for _, eachTestValue := range testUserValues {
		buildID, buildIDErr := provisionBuildID(eachTestValue, logger)
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
