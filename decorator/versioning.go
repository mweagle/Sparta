package decorator

import (
	"context"
	"time"

	sparta "github.com/mweagle/Sparta"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/rs/zerolog"
)

// LambdaVersioningDecorator returns a TemplateDecorator
// that is responsible for including a versioning resource
// with the given lambda function
func LambdaVersioningDecorator() sparta.TemplateDecoratorHookFunc {
	return func(ctx context.Context,
		serviceName string,
		lambdaResourceName string,
		lambdaResource gocf.LambdaFunction,
		resourceMetadata map[string]interface{},
		lambdaFunctionCode *gocf.LambdaFunctionCode,
		buildID string,
		template *gocf.Template,
		logger *zerolog.Logger) (context.Context, error) {

		lambdaResName := sparta.CloudFormationResourceName("LambdaVersion",
			buildID,
			time.Now().UTC().String())
		versionResource := &gocf.LambdaVersion{
			FunctionName: gocf.GetAtt(lambdaResourceName, "Arn").String(),
		}
		lambdaVersionRes := template.AddResource(lambdaResName, versionResource)
		lambdaVersionRes.DeletionPolicy = "Retain"
		// That's it...
		return ctx, nil
	}
}
