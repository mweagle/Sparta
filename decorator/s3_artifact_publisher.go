package decorator

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	gof "github.com/awslabs/goformation/v5/cloudformation"
	goflambda "github.com/awslabs/goformation/v5/cloudformation/lambda"
	sparta "github.com/mweagle/Sparta"
	cfCustomResources "github.com/mweagle/Sparta/aws/cloudformation/resources"
	"github.com/rs/zerolog"
)

// S3ArtifactPublisherDecorator returns a ServiceDecoratorHookHandler
// function that publishes the given data to an S3 Bucket
// using the given bucket and key.
func S3ArtifactPublisherDecorator(bucket string,
	key string,
	data map[string]interface{}) sparta.ServiceDecoratorHookHandler {

	// Setup the CF distro
	artifactDecorator := func(ctx context.Context,
		serviceName string,
		template *gof.Template,
		lambdaFunctionCode *goflambda.Function_Code,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *zerolog.Logger) (context.Context, error) {

		// Ensure the custom action handler...
		sourceArnExpr := gof.Join("", []string{
			"arn:aws:s3:::",
			bucket,
			"/*",
		})

		configuratorResName, err := sparta.EnsureCustomResourceHandler(serviceName,
			cfCustomResources.S3ArtifactPublisher,
			sourceArnExpr,
			[]string{},
			template,
			lambdaFunctionCode,
			logger)

		if err != nil {
			return ctx, err
		}

		// Create the invocation of the custom action...
		s3PublishRequest := &cfCustomResources.S3ArtifactPublisherResourceRequest{
			CustomResourceRequest: cfCustomResources.CustomResourceRequest{
				ServiceToken: gof.GetAtt(configuratorResName, "Arn"),
			},
			Bucket: bucket,
			Key:    key,
			Body:   data,
		}
		s3PublishResource := &cfCustomResources.S3ArtifactPublisherResource{
			CustomResource: gof.CustomResource{
				Properties: cfCustomResources.ToCustomResourceProperties(s3PublishRequest),
			},
		}
		// Name?
		resourceInvokerName := sparta.CloudFormationResourceName("ArtifactS3",
			fmt.Sprintf("%v", bucket),
			fmt.Sprintf("%v", key))

		// Add it
		template.Resources[resourceInvokerName] = s3PublishResource
		return ctx, nil
	}
	return sparta.ServiceDecoratorHookFunc(artifactDecorator)
}
