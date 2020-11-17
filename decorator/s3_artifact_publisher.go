package decorator

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	sparta "github.com/mweagle/Sparta"
	cfCustomResources "github.com/mweagle/Sparta/aws/cloudformation/resources"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/sirupsen/logrus"
)

// S3ArtifactPublisherDecorator returns a ServiceDecoratorHookHandler
// function that publishes the given data to an S3 Bucket
// using the given bucket and key.
func S3ArtifactPublisherDecorator(bucket gocf.Stringable,
	key gocf.Stringable,
	data map[string]interface{}) sparta.ServiceDecoratorHookHandler {

	// Setup the CF distro
	artifactDecorator := func(ctx context.Context,
		serviceName string,
		template *gocf.Template,
		lambdaFunctionCode *gocf.LambdaFunctionCode,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *logrus.Logger) (context.Context, error) {

		// Ensure the custom action handler...
		sourceArnExpr := gocf.Join("",
			gocf.String("arn:aws:s3:::"),
			bucket.String(),
			gocf.String("/*"))

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
		s3PublishResource := &cfCustomResources.S3ArtifactPublisherResource{}
		s3PublishResource.ServiceToken = gocf.GetAtt(configuratorResName, "Arn")
		s3PublishResource.Bucket = bucket.String()
		s3PublishResource.Key = key.String()
		s3PublishResource.Body = data

		// Name?
		resourceInvokerName := sparta.CloudFormationResourceName("ArtifactS3",
			fmt.Sprintf("%v", bucket.String()),
			fmt.Sprintf("%v", key.String()))

		// Add it
		template.AddResource(resourceInvokerName, s3PublishResource)
		return ctx, nil
	}
	return sparta.ServiceDecoratorHookFunc(artifactDecorator)
}
