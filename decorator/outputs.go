package decorator

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws/session"
	gof "github.com/awslabs/goformation/v5/cloudformation"
	goflambda "github.com/awslabs/goformation/v5/cloudformation/lambda"
	sparta "github.com/mweagle/Sparta"
	"github.com/rs/zerolog"
)

var reInvalidOutputChars = regexp.MustCompile("[^A-Za-z0-9]+")

func sanitizedKeyName(userValue string) string {
	return reInvalidOutputChars.ReplaceAllString(userValue, "")
}

// PublishAllResourceOutputs is a utility function to include all Ref and Att
// outputs associated with the given (cfResourceName, cfResource) pair.
func PublishAllResourceOutputs(cfResourceName string,
	cfResource gof.Resource) sparta.ServiceDecoratorHookFunc {
	return func(ctx context.Context,
		serviceName string,
		cfTemplate *gof.Template,
		lambdaFunctionCode *goflambda.Function_Code,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *zerolog.Logger) (context.Context, error) {

		// Add the Ref
		cfTemplate.Outputs[sanitizedKeyName(fmt.Sprintf("%s_Ref", cfResourceName))] = gof.Output{
			Description: fmt.Sprintf("%s (%s) Ref",
				cfResourceName,
				cfResource.AWSCloudFormationType()),
			Value: gof.Ref(cfResourceName),
		}
		// Get the resource attributes
		// TODO - implement resource outputs...
		/*
			for _, eachAttr := range cfResource.CfnResourceAttributes() {
				// Add the function ARN as a stack output
				cfTemplate.Outputs[sanitizedKeyName(fmt.Sprintf("%s_Attr_%s", cfResourceName, eachAttr))] = &gof.Output{
					Description: fmt.Sprintf("%s (%s) Attribute: %s",
						cfResourceName,
						cfResource.CfnResourceType(),
						eachAttr),
					Value: gof.GetAtt(cfResourceName, eachAttr),
				}
			}
		*/
		return ctx, nil
	}
}

// PublishAttOutputDecorator returns a TemplateDecoratorHookFunc
// that publishes an Att value for a given Lambda
func PublishAttOutputDecorator(keyName string, description string, fieldName string) sparta.TemplateDecoratorHookFunc {
	attrDecorator := func(ctx context.Context,
		serviceName string,
		lambdaResourceName string,
		lambdaResource *goflambda.Function,
		resourceMetadata map[string]interface{},
		lambdaFunctionCode *goflambda.Function_Code,
		buildID string,
		template *gof.Template,
		logger *zerolog.Logger) (context.Context, error) {

		// Add the function ARN as a stack output
		template.Outputs[sanitizedKeyName(keyName)] = gof.Output{
			Description: description,
			Value:       gof.GetAtt(lambdaResourceName, fieldName),
		}
		return ctx, nil
	}
	return sparta.TemplateDecoratorHookFunc(attrDecorator)
}

// PublishRefOutputDecorator returns an TemplateDecoratorHookFunc
// that publishes the Ref value for a given lambda
func PublishRefOutputDecorator(keyName string, description string) sparta.TemplateDecoratorHookFunc {
	attrDecorator := func(ctx context.Context,
		serviceName string,
		lambdaResourceName string,
		lambdaResource *goflambda.Function,
		resourceMetadata map[string]interface{},
		lambdaFunctionCode *goflambda.Function_Code,
		buildID string,
		template *gof.Template,
		logger *zerolog.Logger) (context.Context, error) {

		// Add the function ARN as a stack output
		template.Outputs[sanitizedKeyName(keyName)] = gof.Output{
			Description: description,
			Value:       gof.Ref(lambdaResourceName),
		}
		return ctx, nil
	}

	return sparta.TemplateDecoratorHookFunc(attrDecorator)
}
