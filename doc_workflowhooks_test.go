package sparta

import (
	"archive/zip"
	"context"
	"io"

	"strings"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sirupsen/logrus"
)

const userdataResourceContents = `
{
  "Hello" : "World",
}`

func helloZipLambda(ctx context.Context,
	props map[string]interface{}) (string, error) {
	lambdaCtx, _ := lambdacontext.FromContext(ctx)
	Logger().WithFields(logrus.Fields{
		"RequestID":  lambdaCtx.AwsRequestID,
		"Properties": props,
	}).Info("Lambda event")
	return "Event processed", nil
}

func archiveHook(context map[string]interface{},
	serviceName string,
	zipWriter *zip.Writer,
	awsSession *session.Session,
	noop bool,
	logger *logrus.Logger) error {

	logger.Info("Adding userResource")
	resourceFileName := "userResource.json"
	binaryWriter, binaryWriterErr := zipWriter.Create(resourceFileName)
	if nil != binaryWriterErr {
		return binaryWriterErr
	}
	userdataReader := strings.NewReader(userdataResourceContents)
	_, copyErr := io.Copy(binaryWriter, userdataReader)
	return copyErr
}

func ExampleWorkflowHooks() {
	workflowHooks := WorkflowHooks{
		Archive: archiveHook,
	}

	var lambdaFunctions []*LambdaAWSInfo
	helloWorldLambda := HandleAWSLambda("PreexistingAWSLambdaRoleName",
		helloZipLambda,
		nil)
	lambdaFunctions = append(lambdaFunctions, helloWorldLambda)
	MainEx("HelloWorldArchiveHook",
		"Description for Hello World HelloWorldArchiveHook",
		lambdaFunctions,
		nil,
		nil,
		&workflowHooks,
		false)
}
