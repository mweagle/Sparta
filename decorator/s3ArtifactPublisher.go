package decorator

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	sparta "github.com/mweagle/Sparta"
	cfCustomResources "github.com/mweagle/Sparta/aws/cloudformation/resources"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/sirupsen/logrus"
)

// S3ArtifactPublisherDecorator returns a ServiceDecoratorHookHandler
// function that publishes the given data to an S3 Bucket
// using the given bucket and key
func S3ArtifactPublisherDecorator(bucket gocf.Stringable,
	key gocf.Stringable,
	data gocf.Stringable) sparta.ServiceDecoratorHookHandler {

	// Setup the CF distro
	artifactDecorator := func(context map[string]interface{},
		serviceName string,
		template *gocf.Template,
		S3Bucket string,
		S3Key string,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *logrus.Logger) error {

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
			S3Bucket,
			S3Key,
			logger)

		if err != nil {
			return err
		}

		// Create the invocation of the custom action...
		s3PublishResource := &cfCustomResources.S3ArtifactPublisherResource{}
		s3PublishResource.ServiceToken = gocf.GetAtt(configuratorResName, "Arn")
		s3PublishResource.Bucket = bucket.String()
		s3PublishResource.Key = key.String()
		s3PublishResource.Body = data.String()

		// Name?
		resourceInvokerName := sparta.CloudFormationResourceName("ArtifactS3",
			fmt.Sprintf("%v", bucket.String()),
			fmt.Sprintf("%v", key.String()))

		// Add it
		template.AddResource(resourceInvokerName, s3PublishResource)
		return nil
	}
	return sparta.ServiceDecoratorHookFunc(artifactDecorator)
}
