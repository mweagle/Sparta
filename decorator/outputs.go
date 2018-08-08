package decorator

import (
	"regexp"

	"github.com/mweagle/Sparta"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/sirupsen/logrus"
)

var reInvalidOutputChars = regexp.MustCompile("[^A-Za-z0-9]+")

func sanitizedKeyName(userValue string) string {
	return reInvalidOutputChars.ReplaceAllString(userValue, "")
}

// PublishAttOutputDecorator returns an TemplateDecoratorHookFunc
// that publishes an Att value for a given Lambda
func PublishAttOutputDecorator(keyName string, description string, fieldName string) sparta.TemplateDecoratorHookFunc {
	attrDecorator := func(serviceName string,
		lambdaResourceName string,
		lambdaResource gocf.LambdaFunction,
		resourceMetadata map[string]interface{},
		S3Bucket string,
		S3Key string,
		buildID string,
		template *gocf.Template,
		context map[string]interface{},
		logger *logrus.Logger) error {

		// Add the function ARN as a stack output
		template.Outputs[sanitizedKeyName(keyName)] = &gocf.Output{
			Description: description,
			Value:       gocf.GetAtt(lambdaResourceName, fieldName),
		}
		return nil
	}

	return sparta.TemplateDecoratorHookFunc(attrDecorator)
}

// PublishRefOutputDecorator returns an TemplateDecoratorHookFunc
// that publishes the Ref value for a given lambda
func PublishRefOutputDecorator(keyName string, description string) sparta.TemplateDecoratorHookFunc {
	attrDecorator := func(serviceName string,
		lambdaResourceName string,
		lambdaResource gocf.LambdaFunction,
		resourceMetadata map[string]interface{},
		S3Bucket string,
		S3Key string,
		buildID string,
		template *gocf.Template,
		context map[string]interface{},
		logger *logrus.Logger) error {

		// Add the function ARN as a stack output
		template.Outputs[sanitizedKeyName(keyName)] = &gocf.Output{
			Description: description,
			Value:       gocf.Ref(lambdaResourceName),
		}
		return nil
	}

	return sparta.TemplateDecoratorHookFunc(attrDecorator)
}
