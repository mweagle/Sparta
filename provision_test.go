package sparta

import (
	"context"
	"testing"

	gofcloudformation "github.com/awslabs/goformation/v5/cloudformation/cloudformation"
	cwCustomProvider "github.com/mweagle/Sparta/v3/aws/cloudformation/provider"

	gof "github.com/awslabs/goformation/v5/cloudformation"
	goflambda "github.com/awslabs/goformation/v5/cloudformation/lambda"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type cloudFormationProvisionTestResource struct {
	gofcloudformation.CustomResource
	TestKey interface{}
}

func customResourceTestProvider(resourceType string) gof.Resource {
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
	cwCustomProvider.RegisterCustomResourceProvider(customResourceTestProvider)
}

func TestProvision(t *testing.T) {
	testProvision(t, testLambdaData(), nil)
}

func templateDecorator(ctx context.Context,
	serviceName string,
	lambdaResourceName string,
	lambdaResource *goflambda.Function,
	resourceMetadata map[string]interface{},
	lambdaFunctionCode *goflambda.Function_Code,
	buildID string,
	cfTemplate *gof.Template,
	logger *zerolog.Logger) (context.Context, error) {

	// Add an empty resource
	newResource, err := newCloudFormationResource("Custom::ProvisionTestEmpty", logger)
	if nil != err {
		return ctx, errors.Wrapf(err, "Failed to create test resource")
	}
	customResource := newResource.(*cloudFormationProvisionTestResource)
	customResource.ServiceToken = "arn:aws:sns:us-east-1:84969EXAMPLE:CRTest"
	customResource.TestKey = "Hello World"
	cfTemplate.Resources["ProvisionTestResource"] = customResource

	// Add an output
	cfTemplate.Outputs["OutputDecorationTest"] = gof.Output{
		Description: "Information about the value",
		Value:       "My key",
	}
	return ctx, nil
}

func TestDecorateProvision(t *testing.T) {

	lambdas := testLambdaData()
	lambdas[0].Decorator = templateDecorator
	testProvision(t, lambdas, nil)
}
