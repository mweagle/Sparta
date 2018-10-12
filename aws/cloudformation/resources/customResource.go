package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	awsLambdaCtx "github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	// CreateOperation is a request to create a resource
	// @enum CloudFormationOperation
	CreateOperation = "Create"
	// DeleteOperation is a request to delete a resource
	// @enum CloudFormationOperation
	DeleteOperation = "Delete"
	// UpdateOperation is a request to update a resource
	// @enum CloudFormationOperation
	UpdateOperation = "Update"
)
const (
	// CustomResourceTypePrefix is the known custom resource
	// type prefix
	CustomResourceTypePrefix = "Custom::goAWS"
)

var (
	// HelloWorld is the typename for HelloWorldResource
	HelloWorld = cloudFormationResourceType("HelloWorldResource")
	// S3LambdaEventSource is the typename for S3LambdaEventSourceResource
	S3LambdaEventSource = cloudFormationResourceType("S3EventSource")
	// SNSLambdaEventSource is the typename for SNSLambdaEventSourceResource
	SNSLambdaEventSource = cloudFormationResourceType("SNSEventSource")
	// SESLambdaEventSource is the typename for SESLambdaEventSourceResource
	SESLambdaEventSource = cloudFormationResourceType("SESEventSource")
	// CloudWatchLogsLambdaEventSource is the typename for SESLambdaEventSourceResource
	CloudWatchLogsLambdaEventSource = cloudFormationResourceType("CloudWatchLogsEventSource")
	// ZipToS3Bucket is the typename for ZipToS3Bucket
	ZipToS3Bucket = cloudFormationResourceType("ZipToS3Bucket")
	// S3ArtifactPublisher is the typename for publishing an S3Artifact
	S3ArtifactPublisher = cloudFormationResourceType("S3ArtifactPublisher")
)

func customTypeProvider(resourceType string) gocf.ResourceProperties {
	switch resourceType {
	case HelloWorld:
		return &HelloWorldResource{}
	case S3LambdaEventSource:
		return &S3LambdaEventSourceResource{}
	case CloudWatchLogsLambdaEventSource:
		return &CloudWatchLogsLambdaEventSourceResource{}
	case SNSLambdaEventSource:
		return &SNSLambdaEventSourceResource{}
	case SESLambdaEventSource:
		return &SESLambdaEventSourceResource{}
	case ZipToS3Bucket:
		return &ZipToS3BucketResource{}
	case S3ArtifactPublisher:
		return &S3ArtifactPublisherResource{}
	}
	return nil
}

func init() {
	gocf.RegisterCustomResourceProvider(customTypeProvider)
}

// CustomResourceCommand defines operations that a CustomResource must implement.
type CustomResourceCommand interface {
	Create(session *session.Session,
		event *CloudFormationLambdaEvent,
		logger *logrus.Logger) (map[string]interface{}, error)

	Update(session *session.Session,
		event *CloudFormationLambdaEvent,
		logger *logrus.Logger) (map[string]interface{}, error)

	Delete(session *session.Session,
		event *CloudFormationLambdaEvent,
		logger *logrus.Logger) (map[string]interface{}, error)

	IAMPrivileges() []string
}

// cloudFormationResourceType a string for the resource name that represents a
// custom CloudFormation resource typename
func cloudFormationResourceType(resType string) string {
	return fmt.Sprintf("%s::%s", CustomResourceTypePrefix, resType)
}

type logrusProxy struct {
	logger *logrus.Logger
}

func (proxy *logrusProxy) Log(args ...interface{}) {
	proxy.logger.Info(args...)
}

// CloudFormationLambdaEvent is the event to a resource
type CloudFormationLambdaEvent struct {
	RequestType           string
	RequestID             string `json:"RequestId"`
	ResponseURL           string
	ResourceType          string
	StackID               string `json:"StackId"`
	LogicalResourceID     string `json:"LogicalResourceId"`
	ResourceProperties    json.RawMessage
	OldResourceProperties json.RawMessage
}

// SendCloudFormationResponse sends the given response
// to the CloudFormation URL that was submitted together
// with this event
func SendCloudFormationResponse(lambdaCtx *awsLambdaCtx.LambdaContext,
	event *CloudFormationLambdaEvent,
	results map[string]interface{},
	responseErr error,
	logger *logrus.Logger) error {

	status := "FAILED"
	if nil == responseErr {
		status = "SUCCESS"
	}
	// Env vars:
	// https://docs.aws.amazon.com/lambda/latest/dg/current-supported-versions.html
	logGroupName := os.Getenv("AWS_LAMBDA_LOG_GROUP_NAME")
	logStreamName := os.Getenv("AWS_LAMBDA_LOG_STREAM_NAME")
	reasonText := ""
	if nil != responseErr {
		reasonText = fmt.Sprintf("%s. Details in CloudWatch Logs: %s : %s",
			responseErr.Error(),
			logGroupName,
			logStreamName)
	} else {
		reasonText = fmt.Sprintf("Details in CloudWatch Logs: %s : %s",
			logGroupName,
			logStreamName)
	}
	// PhysicalResourceId
	// This value should be an identifier unique to the custom resource vendor,
	// and can be up to 1 Kb in size. The value must be a non-empty string and
	// must be identical for all responses for the same resource.
	// Ref: https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/crpg-ref-requesttypes-create.html
	physicalResourceID := fmt.Sprintf("LogStreamName: %s", logStreamName)
	responseData := map[string]interface{}{
		"Status":             status,
		"Reason":             reasonText,
		"PhysicalResourceId": physicalResourceID,
		"StackId":            event.StackID,
		"RequestId":          event.RequestID,
		"LogicalResourceId":  event.LogicalResourceID,
	}
	if nil != responseErr {
		responseData["Data"] = map[string]interface{}{
			"Error": responseErr,
		}
	} else if nil != results {
		responseData["Data"] = results
	} else {
		responseData["Data"] = map[string]interface{}{}
	}

	logger.WithFields(logrus.Fields{
		"ResponsePayload": responseData,
	}).Debug("Response Info")

	jsonData, jsonError := json.Marshal(responseData)
	if nil != jsonError {
		return errors.Wrap(jsonError, "Attempting to marshal Cloudformation response")
	}

	responseBuffer := strings.NewReader(string(jsonData))
	req, httpErr := http.NewRequest("PUT",
		event.ResponseURL,
		responseBuffer)

	if nil != httpErr {
		return httpErr
	}
	// Need to use the Opaque field b/c Go will parse inline encoded values
	// which are supposed to be roundtripped to AWS.
	// Ref: https://tools.ietf.org/html/rfc3986#section-2.2
	// Ref: https://golang.org/pkg/net/url/#URL
	// Ref: https://github.com/aws/aws-sdk-go/issues/337
	// parsedURL, parsedURLErr := url.ParseRequestURI(event.ResponseURL)

	// https://en.wikipedia.org/wiki/Percent-encoding
	mapReplace := map[string]string{
		":": "%3A",
		"|": "%7C",
	}
	req.URL.Opaque = req.URL.Path
	for eachKey, eachValue := range mapReplace {
		req.URL.Opaque = strings.Replace(req.URL.Opaque, eachKey, eachValue, -1)
	}
	req.URL.Path = ""
	req.URL.RawPath = ""

	logger.WithFields(logrus.Fields{
		"RawURL": event.ResponseURL,
		"URL":    req.URL,
		"Body":   responseData,
	}).Debug("Created URL response")

	// Although it seems reasonable to set the Content-Type to "application/json" - don't.
	// The Content-Type must be an empty string in order for the
	// AWS Signature checker to pass.
	// Ref: http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-lambda-function-code.html
	req.Header.Set("content-type", "")

	client := &http.Client{}
	resp, httpErr := client.Do(req)
	if httpErr != nil {
		return errors.Wrapf(httpErr, "Sending CloudFormation response")
	}
	logger.WithFields(logrus.Fields{
		"LogicalResourceId":  event.LogicalResourceID,
		"Result":             responseData["Status"],
		"ResponseStatusCode": resp.StatusCode,
	}).Debug("Sent CloudFormation response")

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, bodyErr := ioutil.ReadAll(resp.Body)
		if bodyErr != nil {
			logger.Warn("Unable to read body: " + bodyErr.Error())
			body = []byte{}
		}
		return errors.Errorf("Error sending response: %d. Data: %s", resp.StatusCode, string(body))
	}
	defer resp.Body.Close()
	return nil
}

// Returns an AWS Session (https://github.com/aws/aws-sdk-go/wiki/Getting-Started-Configuration)
// object that attaches a debug level handler to all AWS requests from services
// sharing the session value.
func awsSession(logger *logrus.Logger) *session.Session {
	awsConfig := &aws.Config{
		CredentialsChainVerboseErrors: aws.Bool(true),
	}

	// Log AWS calls if needed
	switch logger.Level {
	case logrus.DebugLevel:
		awsConfig.LogLevel = aws.LogLevel(aws.LogDebugWithHTTPBody)
	}
	awsConfig.Logger = &logrusProxy{logger}
	sess := session.New(awsConfig)
	sess.Handlers.Send.PushFront(func(r *request.Request) {
		logger.WithFields(logrus.Fields{
			"Service":   r.ClientInfo.ServiceName,
			"Operation": r.Operation.Name,
			"Method":    r.Operation.HTTPMethod,
			"Path":      r.Operation.HTTPPath,
			"Payload":   r.Params,
		}).Debug("AWS Request")
	})
	return sess
}

// CloudFormationLambdaCustomResourceHandler is an adapter
// function that transforms an implementing CustomResourceCommand
// into something that that can respond to the lambda custom
// resource lifecycle
func CloudFormationLambdaCustomResourceHandler(command CustomResourceCommand, logger *logrus.Logger) interface{} {
	return func(ctx context.Context,
		event CloudFormationLambdaEvent) error {
		lambdaCtx, lambdaCtxOk := awsLambdaCtx.FromContext(ctx)
		if !lambdaCtxOk {
			return errors.Errorf("Failed to access AWS Lambda Context from ctx argument")
		}
		customResourceSession := awsSession(logger)
		var opResults map[string]interface{}
		var opErr error
		executeOperation := false
		// If we're in cleanup mode, then skip it...
		// Don't forward to the CustomAction handler iff we're in CLEANUP mode
		describeStacksInput := &cloudformation.DescribeStacksInput{
			StackName: aws.String(event.StackID),
		}
		cfSvc := cloudformation.New(customResourceSession)
		describeStacksOutput, describeStacksOutputErr := cfSvc.DescribeStacks(describeStacksInput)
		if nil != describeStacksOutputErr {
			opErr = describeStacksOutputErr
		} else {
			stackDesc := describeStacksOutput.Stacks[0]
			if nil == stackDesc {
				opErr = errors.Errorf("DescribeStack failed: %s", event.StackID)
			} else {
				executeOperation = ("UPDATE_COMPLETE_CLEANUP_IN_PROGRESS" != *stackDesc.StackStatus)
			}
		}

		logger.WithFields(logrus.Fields{
			"ExecuteOperation": event.LogicalResourceID,
			"Stacks":           fmt.Sprintf("%#+v", describeStacksOutput),
			"RequestType":      event.RequestType,
		}).Debug("CustomResource Request")

		if opErr == nil && executeOperation {
			switch event.RequestType {
			case CreateOperation:
				opResults, opErr = command.Create(customResourceSession, &event, logger)
			case DeleteOperation:
				opResults, opErr = command.Delete(customResourceSession, &event, logger)
			case UpdateOperation:
				opResults, opErr = command.Update(customResourceSession, &event, logger)
			}
		}
		// Notify CloudFormation of the result
		if event.ResponseURL != "" {
			sendErr := SendCloudFormationResponse(lambdaCtx,
				&event,
				opResults,
				opErr,
				logger)
			if nil != sendErr {
				logger.WithFields(logrus.Fields{
					"Error": sendErr.Error(),
					"URL":   event.ResponseURL,
				}).Info("Failed to ACK status to CloudFormation")
			} else {
				// If the cloudformation notification was complete, then this
				// execution functioned properly and we can clear the Error
				opErr = nil
			}
		}
		return opErr
	}
}

// NewCustomResourceLambdaHandler returns a handler for the given
// type
func NewCustomResourceLambdaHandler(resourceType string, logger *logrus.Logger) interface{} {

	// TODO - eliminate this factory stuff and just register
	// the custom resources as normal lambda handlers...
	var lambdaCmd CustomResourceCommand
	cfResource := customTypeProvider(resourceType)
	if cfResource != nil {
		cmd, cmdOK := cfResource.(CustomResourceCommand)
		if cmdOK {
			lambdaCmd = cmd
		}
	}
	if lambdaCmd == nil {
		return errors.Errorf("Custom resource handler not found for type: %s", resourceType)
	}
	return CloudFormationLambdaCustomResourceHandler(lambdaCmd, logger)
}
