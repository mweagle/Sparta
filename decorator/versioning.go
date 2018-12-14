package decorator

import (
	"time"

	sparta "github.com/mweagle/Sparta"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/sirupsen/logrus"
)

// LambdaVersioningDecorator returns a TemplateDecorator
// that is responsible for including a versioning resource
// with the given lambda function
func LambdaVersioningDecorator() sparta.TemplateDecoratorHookFunc {
	return func(serviceName string,
		lambdaResourceName string,
		lambdaResource gocf.LambdaFunction,
		resourceMetadata map[string]interface{},
		S3Bucket string,
		S3Key string,
		buildID string,
		template *gocf.Template,
		context map[string]interface{},
		logger *logrus.Logger) error {

		lambdaResName := sparta.CloudFormationResourceName("LambdaVersion",
			buildID,
			time.Now().UTC().String())
		versionResource := &gocf.LambdaVersion{
			FunctionName: gocf.GetAtt(lambdaResourceName, "Arn").String(),
		}
		lambdaVersionRes := template.AddResource(lambdaResName, versionResource)
		lambdaVersionRes.DeletionPolicy = "Retain"
		// That's it...
		return nil
	}
}
