package sparta

import (
	"archive/zip"
	"context"
	"io"
	"strings"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/rs/zerolog"
)

const userdataResourceContents = `
{
  "Hello" : "World",
}`

func helloZipLambda(ctx context.Context,
	props map[string]interface{}) (string, error) {
	lambdaCtx, _ := lambdacontext.FromContext(ctx)
	Logger().Info().
		Str("RequestID", lambdaCtx.AwsRequestID).
		Interface("Properties", props).
		Msg("Lambda event")
	return "Event processed", nil
}

func archiveHook(ctx context.Context,
	serviceName string,
	zipWriter *zip.Writer,
	awsSession *session.Session,
	noop bool,
	logger *zerolog.Logger) (context.Context, error) {

	logger.Info().Msg("Adding userResource")
	resourceFileName := "userResource.json"
	binaryWriter, binaryWriterErr := zipWriter.Create(resourceFileName)
	if nil != binaryWriterErr {
		return ctx, binaryWriterErr
	}
	userdataReader := strings.NewReader(userdataResourceContents)
	_, copyErr := io.Copy(binaryWriter, userdataReader)
	return ctx, copyErr
}

func ExampleWorkflowHooks() {
	workflowHooks := WorkflowHooks{
		Archives: []ArchiveHookHandler{ArchiveHookFunc(archiveHook)},
	}
	var lambdaFunctions []*LambdaAWSInfo
	helloWorldLambda, _ := NewAWSLambda("PreexistingAWSLambdaRoleName",
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
