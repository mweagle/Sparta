package decorator

import (
	"context"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	gof "github.com/awslabs/goformation/v5/cloudformation"
	goflambda "github.com/awslabs/goformation/v5/cloudformation/lambda"
	sparta "github.com/mweagle/Sparta"
	"github.com/rs/zerolog"
)

// ResourceDecorator is a convenience function to insert a map
// of resources into the template.
func ResourceDecorator(resources map[string]gof.Resource) sparta.ServiceDecoratorHookFunc {
	return func(ctx context.Context,
		serviceName string,
		cfTemplate *gof.Template,
		lambdaFunctionCode *goflambda.Function_Code,
		buildID string,
		awsConfig awsv2.Config,
		noop bool,
		logger *zerolog.Logger) (context.Context, error) {

		for eachName, eachRes := range resources {
			cfTemplate.Resources[eachName] = eachRes
			logger.Debug().
				Str(eachName, eachRes.AWSCloudFormationType()).
				Msg("Inserting resource into template")
		}
		return ctx, nil
	}
}
