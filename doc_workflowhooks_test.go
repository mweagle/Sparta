package sparta

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/session"
	"net/http"
	"strings"
)

const userdataResourceContents = `
{
  "Hello" : "World",
}`

// Standard AWS λ function
func helloZipLambda(event *json.RawMessage,
	context *LambdaContext,
	w http.ResponseWriter,
	logger *logrus.Logger) {

	fmt.Fprint(w, "Hello World")
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
		ArchiveHook: archiveHook,
	}

	var lambdaFunctions []*LambdaAWSInfo
	helloWorldLambda := NewLambda("PreexistingAWSLambdaRoleName", mainHelloWorld, nil)
	lambdaFunctions = append(lambdaFunctions, helloWorldLambda)
	MainEx("HelloWorldArchiveHook",
		"Description for Hello World HelloWorldArchiveHook",
		lambdaFunctions,
		nil,
		nil,
		&workflowHooks)
}
