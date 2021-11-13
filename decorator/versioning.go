package decorator

import (
	"context"
	"time"

	gof "github.com/awslabs/goformation/v5/cloudformation"
	goflambda "github.com/awslabs/goformation/v5/cloudformation/lambda"
	sparta "github.com/mweagle/Sparta/v3"
	"github.com/rs/zerolog"
)

// LambdaVersioningDecorator returns a TemplateDecorator
// that is responsible for including a versioning resource
// with the given lambda function
func LambdaVersioningDecorator() sparta.TemplateDecoratorHookFunc {
	return func(ctx context.Context,
		serviceName string,
		lambdaResourceName string,
		lambdaResource *goflambda.Function,
		resourceMetadata map[string]interface{},
		lambdaFunctionCode *goflambda.Function_Code,
		buildID string,
		template *gof.Template,
		logger *zerolog.Logger) (context.Context, error) {

		lambdaResName := sparta.CloudFormationResourceName("LambdaVersion",
			buildID,
			time.Now().UTC().String())
		versionResource := &goflambda.Version{
			FunctionName: gof.GetAtt(lambdaResourceName, "Arn"),
		}
		versionResource.AWSCloudFormationDeletionPolicy = "Retain"
		template.Resources[lambdaResName] = versionResource

		// That's it...
		return ctx, nil
	}
}
