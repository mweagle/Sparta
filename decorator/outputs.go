package decorator

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws/session"
	sparta "github.com/mweagle/Sparta"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/rs/zerolog"
)

var reInvalidOutputChars = regexp.MustCompile("[^A-Za-z0-9]+")

func sanitizedKeyName(userValue string) string {
	return reInvalidOutputChars.ReplaceAllString(userValue, "")
}

// PublishAllResourceOutputs is a utility function to include all Ref and Att
// outputs associated with the given (cfResourceName, cfResource) pair.
func PublishAllResourceOutputs(cfResourceName string,
	cfResource gocf.ResourceProperties) sparta.ServiceDecoratorHookFunc {
	return func(ctx context.Context,
		serviceName string,
		cfTemplate *gocf.Template,
		lambdaFunctionCode *gocf.LambdaFunctionCode,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *zerolog.Logger) (context.Context, error) {

		// Add the Ref
		cfTemplate.Outputs[sanitizedKeyName(fmt.Sprintf("%s_Ref", cfResourceName))] = &gocf.Output{
			Description: fmt.Sprintf("%s (%s) Ref",
				cfResourceName,
				cfResource.CfnResourceType()),
			Value: gocf.Ref(cfResourceName),
		}
		// Get the resource attributes
		for _, eachAttr := range cfResource.CfnResourceAttributes() {
			// Add the function ARN as a stack output
			cfTemplate.Outputs[sanitizedKeyName(fmt.Sprintf("%s_Attr_%s", cfResourceName, eachAttr))] = &gocf.Output{
				Description: fmt.Sprintf("%s (%s) Attribute: %s",
					cfResourceName,
					cfResource.CfnResourceType(),
					eachAttr),
				Value: gocf.GetAtt(cfResourceName, eachAttr),
			}
		}
		return ctx, nil
	}
}

// PublishAttOutputDecorator returns a TemplateDecoratorHookFunc
// that publishes an Att value for a given Lambda
func PublishAttOutputDecorator(keyName string, description string, fieldName string) sparta.TemplateDecoratorHookFunc {
	attrDecorator := func(ctx context.Context,
		serviceName string,
		lambdaResourceName string,
		lambdaResource gocf.LambdaFunction,
		resourceMetadata map[string]interface{},
		lambdaFunctionCode *gocf.LambdaFunctionCode,
		buildID string,
		template *gocf.Template,
		logger *zerolog.Logger) (context.Context, error) {

		// Add the function ARN as a stack output
		template.Outputs[sanitizedKeyName(keyName)] = &gocf.Output{
			Description: description,
			Value:       gocf.GetAtt(lambdaResourceName, fieldName),
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
		lambdaResource gocf.LambdaFunction,
		resourceMetadata map[string]interface{},
		lambdaFunctionCode *gocf.LambdaFunctionCode,
		buildID string,
		template *gocf.Template,
		logger *zerolog.Logger) (context.Context, error) {

		// Add the function ARN as a stack output
		template.Outputs[sanitizedKeyName(keyName)] = &gocf.Output{
			Description: description,
			Value:       gocf.Ref(lambdaResourceName),
		}
		return ctx, nil
	}

	return sparta.TemplateDecoratorHookFunc(attrDecorator)
}
