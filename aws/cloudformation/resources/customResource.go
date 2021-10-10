package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2Config "github.com/aws/aws-sdk-go-v2/config"
	awsv2CF "github.com/aws/aws-sdk-go-v2/service/cloudformation"
	awsv2CFTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"

	smithyLogging "github.com/aws/smithy-go/logging"
	cwCustomProvider "github.com/mweagle/Sparta/aws/cloudformation/provider"

	gof "github.com/awslabs/goformation/v5/cloudformation"

	awsLambdaCtx "github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
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
	CustomResourceTypePrefix = "Custom::Sparta"
)

var (
	// HelloWorld is the typename for HelloWorldResource
	HelloWorld = cloudFormationCustomResourceType("HelloWorldResource")
	// S3LambdaEventSource is the typename for S3LambdaEventSourceResource
	S3LambdaEventSource = cloudFormationCustomResourceType("S3EventSource")
	// SNSLambdaEventSource is the typename for SNSLambdaEventSourceResource
	SNSLambdaEventSource = cloudFormationCustomResourceType("SNSEventSource")
	// CodeCommitLambdaEventSource is the type name for CodeCommitEventSourceResource
	CodeCommitLambdaEventSource = cloudFormationCustomResourceType("CodeCommitEventSource")
	// SESLambdaEventSource is the typename for SESLambdaEventSourceResource
	SESLambdaEventSource = cloudFormationCustomResourceType("SESEventSource")
	// CloudWatchLogsLambdaEventSource is the typename for SESLambdaEventSourceResource
	CloudWatchLogsLambdaEventSource = cloudFormationCustomResourceType("CloudWatchLogsEventSource")
	// ZipToS3Bucket is the typename for ZipToS3Bucket
	ZipToS3Bucket = cloudFormationCustomResourceType("ZipToS3Bucket")
	// S3ArtifactPublisher is the typename for publishing an S3Artifact
	S3ArtifactPublisher = cloudFormationCustomResourceType("S3ArtifactPublisher")
)

// CustomResourceRequest is the default type for all
// requests that support ServiceToken
type CustomResourceRequest struct {
	ServiceToken string
}

func ToCustomResourceProperties(crr interface{}) map[string]interface{} {
	var props map[string]interface{}
	jsonData, jsonDataErr := json.Marshal(crr)
	if jsonDataErr == nil {
		_ = json.Unmarshal(jsonData, &props)
	}
	return props
}

//  customTypeProvider returns a gof.Resource instance if one has been defined
func customTypeProvider(resourceType string) gof.Resource {
	var entry gof.Resource
	switch resourceType {
	case HelloWorld:
		return &HelloWorldResource{
			CustomResource: gof.CustomResource{
				Type: resourceType,
				Properties: map[string]interface{}{
					"io.sparta.restype": resourceType,
				},
			},
		}
	case S3LambdaEventSource:
		return &S3LambdaEventSourceResource{
			CustomResource: gof.CustomResource{
				Type: resourceType,
				Properties: map[string]interface{}{
					"io.sparta.restype": resourceType,
				},
			},
		}
	case CloudWatchLogsLambdaEventSource:
		return &CloudWatchLogsLambdaEventSourceResource{
			CustomResource: gof.CustomResource{
				Type: resourceType,
				Properties: map[string]interface{}{
					"io.sparta.restype": resourceType,
				},
			},
		}
	case CodeCommitLambdaEventSource:
		return &CodeCommitLambdaEventSourceResource{
			CustomResource: gof.CustomResource{
				Type: resourceType,
				Properties: map[string]interface{}{
					"io.sparta.restype": resourceType,
				},
			},
		}
	case SNSLambdaEventSource:
		return &SNSLambdaEventSourceResource{
			CustomResource: gof.CustomResource{
				Type: resourceType,
				Properties: map[string]interface{}{
					"io.sparta.restype": resourceType,
				},
			},
		}
	case SESLambdaEventSource:
		return &SESLambdaEventSourceResource{
			CustomResource: gof.CustomResource{
				Type: resourceType,
				Properties: map[string]interface{}{
					"io.sparta.restype": resourceType,
				},
			},
		}
	case ZipToS3Bucket:
		return &ZipToS3BucketResource{
			CustomResource: gof.CustomResource{
				Type: resourceType,
				Properties: map[string]interface{}{
					"io.sparta.restype": resourceType,
				},
			},
		}
	case S3ArtifactPublisher:
		return &S3ArtifactPublisherResource{
			CustomResource: gof.CustomResource{
				Type: resourceType,
				Properties: map[string]interface{}{
					"io.sparta.restype": resourceType,
				},
			},
		}
	}
	return entry
}

func init() {
	cwCustomProvider.RegisterCustomResourceProvider(customTypeProvider)
}

// CustomResourceCommand defines operations that a CustomResource must implement.
type CustomResourceCommand interface {
	Create(ctx context.Context, awsConfig awsv2.Config,
		event *CloudFormationLambdaEvent,
		logger *zerolog.Logger) (map[string]interface{}, error)

	Update(ctx context.Context, awsConfig awsv2.Config,
		event *CloudFormationLambdaEvent,
		logger *zerolog.Logger) (map[string]interface{}, error)

	Delete(ctx context.Context, awsConfig awsv2.Config,
		event *CloudFormationLambdaEvent,
		logger *zerolog.Logger) (map[string]interface{}, error)
}

// CustomResourcePrivilegedCommand is a command that also has IAM privileges
// which implies there must be an ARN associated with the command
type CustomResourcePrivilegedCommand interface {
	// The IAMPrivileges this command requires of the IAM role
	IAMPrivileges() []string
}

// cloudFormationCustomResourceType a string for the resource name that represents a
// custom CloudFormation resource typename
func cloudFormationCustomResourceType(resType string) string {
	// Ref: https://docs.awsv2.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources.html
	// Use the AWS::CloudFormation::CustomResource or Custom::MyCustomResourceTypeName
	return fmt.Sprintf("%s%s", CustomResourceTypePrefix, resType)
}

type zerologProxy struct {
	logger *zerolog.Logger
}

// Log is a utility function to comply with the AWS signature
func (proxy *zerologProxy) Logf(classification smithyLogging.Classification,
	format string,
	args ...interface{}) {
	proxy.logger.Debug().Msg(fmt.Sprintf(format, args...))
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
	logger *zerolog.Logger) error {

	status := "FAILED"
	if nil == responseErr {
		status = "SUCCESS"
	}
	// Env vars:
	// https://docs.awsv2.amazon.com/lambda/latest/dg/current-supported-versions.html
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
	// Ref: https://docs.awsv2.amazon.com/AWSCloudFormation/latest/UserGuide/crpg-ref-requesttypes-create.html
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

	logger.Debug().
		Interface("ResponsePayload", responseData).
		Msg("Response Info")

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

	logger.Debug().
		Str("RawURL", event.ResponseURL).
		Stringer("URL", req.URL).
		Fields(responseData).
		Msg("Created URL response")

	// Although it seems reasonable to set the Content-Type to "application/json" - don't.
	// The Content-Type must be an empty string in order for the
	// AWS Signature checker to pass.
	// Ref: http://docs.awsv2.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-lambda-function-code.html
	req.Header.Set("content-type", "")

	client := &http.Client{}
	resp, httpErr := client.Do(req)
	if httpErr != nil {
		return errors.Wrapf(httpErr, "Sending CloudFormation response")
	}
	logger.Debug().
		Str("LogicalResourceId", event.LogicalResourceID).
		Interface("Result", responseData["Status"]).
		Int("ResponseStatusCode", resp.StatusCode).
		Msg("Sent CloudFormation response")

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, bodyErr := ioutil.ReadAll(resp.Body)
		if bodyErr != nil {
			logger.Warn().
				Err(bodyErr).
				Msg("Unable to read body")
			body = []byte{}
		}
		return errors.Errorf("Error sending response: %d. Data: %s", resp.StatusCode, string(body))
	}
	defer resp.Body.Close()
	return nil
}

// Returns an AWS Config (https://github.com/aws/aws-sdk-go-v2/blob/main/config/doc.go)
// object that attaches a debug level handler to all AWS requests from services
// sharing the session value.
func newAWSConfig(logger *zerolog.Logger) awsv2.Config {

	logger.Debug().
		Str("Name", awsv2.SDKName).
		Str("Version", awsv2.SDKVersion).
		Msg("AWS SDK Info.")

	awsConfig, awsConfigErr := awsv2Config.LoadDefaultConfig(context.Background())
	if awsConfigErr != nil {
		panic("WAT")
	}
	// Log AWS calls if needed
	switch logger.GetLevel() {
	case zerolog.DebugLevel:
		awsConfig.ClientLogMode = awsv2.LogRequest | awsv2.LogResponse | awsv2.LogRetries
	}
	awsConfig.Logger = &zerologProxy{logger}
	return awsConfig
}

// CloudFormationLambdaCustomResourceHandler is an adapter
// function that transforms an implementing CustomResourceCommand
// into something that that can respond to the lambda custom
// resource lifecycle
func CloudFormationLambdaCustomResourceHandler(command CustomResourceCommand,
	logger *zerolog.Logger) interface{} {
	return func(ctx context.Context,
		event CloudFormationLambdaEvent) error {
		lambdaCtx, lambdaCtxOk := awsLambdaCtx.FromContext(ctx)
		if !lambdaCtxOk {
			return errors.Errorf("Failed to access AWS Lambda Context from ctx argument")
		}
		customResourceConfig := newAWSConfig(logger)
		var opResults map[string]interface{}
		var opErr error
		executeOperation := false
		// If we're in cleanup mode, then skip it...
		// Don't forward to the CustomAction handler iff we're in CLEANUP mode
		describeStacksInput := &awsv2CF.DescribeStacksInput{
			StackName: awsv2.String(event.StackID),
		}
		cfSvc := awsv2CF.NewFromConfig(customResourceConfig)
		describeStacksOutput, describeStacksOutputErr := cfSvc.DescribeStacks(context.Background(), describeStacksInput)
		if nil != describeStacksOutputErr {
			opErr = describeStacksOutputErr
		} else if len(describeStacksOutput.Stacks) <= 0 {
			opErr = errors.Errorf("DescribeStack failed: %s", event.StackID)
		} else {
			stackDesc := describeStacksOutput.Stacks[0]
			executeOperation = (stackDesc.StackStatus != awsv2CFTypes.StackStatusUpdateCompleteCleanupInProgress)
		}

		logger.Debug().
			Str("ExecuteOperation", event.LogicalResourceID).
			Str("Stacks", fmt.Sprintf("%#+v", describeStacksOutput)).
			Str("RequestType", event.RequestType).
			Msg("CustomResource Request")

		if opErr == nil && executeOperation {
			switch event.RequestType {
			case CreateOperation:
				opResults, opErr = command.Create(ctx, customResourceConfig, &event, logger)
			case DeleteOperation:
				opResults, opErr = command.Delete(ctx, customResourceConfig, &event, logger)
			case UpdateOperation:
				opResults, opErr = command.Update(ctx, customResourceConfig, &event, logger)
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
				logger.Info().
					Err(sendErr).
					Str("URL", event.ResponseURL).
					Msg("Failed to ACK status to CloudFormation")
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
func NewCustomResourceLambdaHandler(resourceType string,
	logger *zerolog.Logger) interface{} {

	// TODO - eliminate this factory stuff and just register
	// the custom resources as normal lambda handlers...
	var lambdaCmd CustomResourceCommand
	cfResource, _ := cwCustomProvider.NewCloudFormationCustomResource(resourceType, logger)
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
