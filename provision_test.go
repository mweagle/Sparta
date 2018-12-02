package sparta

import (
	"testing"

	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type cloudFormationProvisionTestResource struct {
	gocf.CloudFormationCustomResource
	ServiceToken string
	TestKey      interface{}
}

func customResourceTestProvider(resourceType string) gocf.ResourceProperties {
	switch resourceType {
	case "Custom::ProvisionTestEmpty":
		{
			return &cloudFormationProvisionTestResource{}
		}
	default:
		return nil
	}
}

func init() {
	gocf.RegisterCustomResourceProvider(customResourceTestProvider)
}

func TestProvision(t *testing.T) {
	testProvision(t, testLambdaData(), nil)
}

func templateDecorator(serviceName string,
	lambdaResourceName string,
	lambdaResource gocf.LambdaFunction,
	resourceMetadata map[string]interface{},
	S3Bucket string,
	S3Key string,
	buildID string,
	cfTemplate *gocf.Template,
	context map[string]interface{},
	logger *logrus.Logger) error {

	// Add an empty resource
	newResource, err := newCloudFormationResource("Custom::ProvisionTestEmpty", logger)
	if nil != err {
		return errors.Wrapf(err, "Failed to create test resource")
	}
	customResource := newResource.(*cloudFormationProvisionTestResource)
	customResource.ServiceToken = "arn:aws:sns:us-east-1:84969EXAMPLE:CRTest"
	customResource.TestKey = "Hello World"
	cfTemplate.AddResource("ProvisionTestResource", customResource)

	// Add an output
	cfTemplate.Outputs["OutputDecorationTest"] = &gocf.Output{
		Description: "Information about the value",
		Value:       gocf.String("My key"),
	}
	return nil
}

func TestDecorateProvision(t *testing.T) {

	lambdas := testLambdaData()
	lambdas[0].Decorator = templateDecorator
	testProvision(t, lambdas, nil)
}
