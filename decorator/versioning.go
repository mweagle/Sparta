package decorator

import (
	sparta "github.com/mweagle/Sparta"
	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/sirupsen/logrus"
)

// LambdaVersioningDecorator returns a TemplateDecorator
// that is responsible for including a versioning resource
// with the given lambda function
func LambdaVersioningDecorator() sparta.TemplateDecorator {
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
		incrementer, incrementerErr :=
			spartaCF.AddAutoIncrementingLambdaVersionResource(serviceName,
				lambdaResourceName,
				template,
				logger)
		if incrementerErr != nil {
			return nil
		}
		versionsMap, versionsMapExists := context[sparta.ContextKeyLambdaVersions].(map[string]*spartaCF.AutoIncrementingLambdaVersionInfo)
		if !versionsMapExists {
			versionsMap = make(map[string]*spartaCF.AutoIncrementingLambdaVersionInfo)
		}
		versionsMap[lambdaResourceName] = incrementer
		context[sparta.ContextKeyLambdaVersions] = versionsMap
		return nil
	}
}
