package cloudformationresources

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	gocf "github.com/crewjam/go-cloudformation"
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

var (
	// HelloWorld is the typename for HelloWorldResource
	HelloWorld = cloudFormationResourceType("HelloWorldResource")
	// S3LambdaEventSource is the typename for S3LambdaEventSourceResource
	S3LambdaEventSource = cloudFormationResourceType("S3LambdaEventSourceResource")
	// SNSLambdaEventSource is the typename for SNSLambdaEventSourceResource
	SNSLambdaEventSource = cloudFormationResourceType("SNSLambdaEventSourceResource")
	// SESLambdaEventSource is the typename for SESLambdaEventSourceResource
	SESLambdaEventSource = cloudFormationResourceType("SESLambdaEventSourceResource")
	// CloudWatchLogsLambdaEventSource is the typename for SESLambdaEventSourceResource
	CloudWatchLogsLambdaEventSource = cloudFormationResourceType("CloudWatchLogsLambdaEventSourceResource")
	// ZipToS3Bucket is the typename for ZipToS3Bucket
	ZipToS3Bucket = cloudFormationResourceType("ZipToS3BucketResource")
)

// CloudFormationLambdaEvent represents the event data sent during a
// Lambda invocation in the context of a CloudFormation operation.
// Ref: http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/crpg-ref-requests.html
type CloudFormationLambdaEvent struct {
	RequestType           string
	ResponseURL           string
	StackID               string `json:"StackId"`
	RequestID             string `json:"RequestId"`
	ResourceType          string
	LogicalResourceID     string `json:"LogicalResourceId"`
	PhysicalResourceID    string `json:"PhysicalResourceId"`
	ResourceProperties    map[string]interface{}
	OldResourceProperties map[string]interface{}
}

// CustomResourceFunction defines a free function that is capable of responding to
// an incoming Lambda-backed CustomResource request and returning either a
// map of outputs or an error.  It is invoked by Run().
type CustomResourceFunction func(requestType string,
	stackID string,
	properties map[string]interface{},
	logger *logrus.Logger) (map[string]interface{}, error)

// AbstractCustomResourceRequest is the base structure used for CustomResourceRequests
type AbstractCustomResourceRequest struct {
	RequestType        string
	ResponseURL        string
	StackID            string `json:"StackId"`
	RequestID          string `json:"RequestId"`
	LogicalResourceID  string `json:"LogicalResourceId"`
	PhysicalResourceID string `json:"PhysicalResourceId"`
	LogGroupName       string `json:"logGroupName"`
	LogStreamName      string `json:"logStreamName"`
	ResourceProperties map[string]interface{}
}

// UserFuncResourceRequest is the go representation of the CloudFormation resource
// request which is handled by a user supplied function. The function result is
// used as the results for the
type UserFuncResourceRequest struct {
	AbstractCustomResourceRequest
	LambdaHandler CustomResourceFunction
}

// CustomResourceRequest is the go representation of a CloudFormation resource
// request for a resource the catalog that had been previously serialized.
type CustomResourceRequest struct {
	AbstractCustomResourceRequest
}

// GoAWSCustomResource is the common embedded struct for all resources defined
// by cloudformationresources
type GoAWSCustomResource struct {
	gocf.CloudFormationCustomResource
	GoAWSType string
}

// CustomResourceCommand defines operations that a CustomResource must implement.
// The return values are either operation outputs or an error value that should
// be used in the response to the CloudFormation AWS Lambda response.
type CustomResourceCommand interface {
	create(session *session.Session,
		logger *logrus.Logger) (map[string]interface{}, error)

	update(session *session.Session,
		logger *logrus.Logger) (map[string]interface{}, error)

	delete(session *session.Session,
		logger *logrus.Logger) (map[string]interface{}, error)
}

// Run manages invoking a user supplied function to perform
// the CloudFormation resource operation. Clients do not need
// to implement anything cloudformationresource related.
func Run(request *UserFuncResourceRequest, logger *logrus.Logger) error {
	logger.WithFields(logrus.Fields{
		"Name":    aws.SDKName,
		"Version": aws.SDKVersion,
	}).Debug("CloudFormation CustomResource AWS SDK info")

	operationOutputs, operationError := request.LambdaHandler(request.RequestType,
		request.StackID,
		request.ResourceProperties,
		logger)

	// Notify CloudFormation of the result
	if "" != request.ResponseURL {
		sendErr := sendCloudFormationResponse(&request.AbstractCustomResourceRequest,
			operationOutputs,
			operationError,
			logger)
		if nil != sendErr {
			logger.WithFields(logrus.Fields{
				"Error": sendErr.Error(),
			}).Info("Failed to notify CloudFormation of result.")
		} else {
			// If the cloudformation notification was complete, then this
			// execution functioned properly and we can clear the Error
			operationError = nil
		}
	}
	return operationError
}

// Handle processes the given CustomResourceRequest value
func Handle(request *CustomResourceRequest, logger *logrus.Logger) error {

	session := awsSession(logger)

	var operationOutputs map[string]interface{}
	var operationError error
	executeOp := false

	logger.WithFields(logrus.Fields{
		"Request": request,
	}).Debug("Incoming request")

	marshaledProperties, marshalError := json.Marshal(request.ResourceProperties)
	if nil != marshalError {
		operationError = marshalError
	}
	if nil == operationError {
		// Don't forward to the CustomAction handler iff we're in CLEANUP mode
		describeStacksInput := &cloudformation.DescribeStacksInput{
			StackName: aws.String(request.StackID),
		}
		cfSvc := cloudformation.New(session)
		describeStacksOutput, describeStacksOutputErr := cfSvc.DescribeStacks(describeStacksInput)
		if nil != describeStacksOutputErr {
			operationError = describeStacksOutputErr
		} else {
			stackDesc := describeStacksOutput.Stacks[0]
			if nil == stackDesc {
				operationError = fmt.Errorf("Failed to describe stack: %s", request.StackID)
			} else {
				executeOp = ("UPDATE_COMPLETE_CLEANUP_IN_PROGRESS" != *stackDesc.StackStatus)
			}
		}
	}
	goAWSResourceType, resourceTypeOK := request.ResourceProperties["GoAWSType"].(string)
	if !resourceTypeOK {
		goAWSResourceType = "Unknown"
	}

	logger.WithFields(logrus.Fields{
		"CustomResource": goAWSResourceType,
		"Operation":      request.RequestType,
		"StackId":        request.StackID,
	}).Info("CustomResource request")

	if nil == operationError && executeOp {
		commandInstance, commandError := customCommandForTypeName(goAWSResourceType, &marshaledProperties)
		if nil != commandError {
			return commandError
		}
		// TODO - lift this into a backoff/retry loop
		customCommandHandler := commandInstance.(CustomResourceCommand)
		switch request.RequestType {
		case CreateOperation:
			operationOutputs, operationError = customCommandHandler.create(session, logger)
		case DeleteOperation:
			operationOutputs, operationError = customCommandHandler.delete(session, logger)
			if operationError != nil {
				logger.WithFields(logrus.Fields{
					"Request": request,
					"Error":   operationError,
				}).Warn("Failed to delete resource during Delete operation")
				operationError = nil
			}
		case UpdateOperation:
			operationOutputs, operationError = customCommandHandler.update(session, logger)
		default:
			operationError = fmt.Errorf("Unsupported operation: %s", request.RequestType)
		}
	}
	if nil != operationError {
		logger.WithFields(logrus.Fields{
			"Operation":    request.RequestType,
			"ResourceType": goAWSResourceType,
			"Error":        operationError,
		}).Error("Failed to execute CustomResource request")
	}
	// Notify CloudFormation of the result
	if "" != request.ResponseURL {
		sendErr := sendCloudFormationResponse(&request.AbstractCustomResourceRequest, operationOutputs, operationError, logger)
		if nil != sendErr {
			logger.WithFields(logrus.Fields{
				"Error": sendErr.Error(),
			}).Info("Failed to notify CloudFormation of result.")
		} else {
			// If the cloudformation notification was complete, then this
			// execution functioned properly and we can clear the Error
			operationError = nil
		}
	}
	return operationError
}

func sendCloudFormationResponse(customResourceRequest *AbstractCustomResourceRequest,
	results map[string]interface{},
	responseErr error,
	logger *logrus.Logger) error {

	parsedURL, parsedURLErr := url.ParseRequestURI(customResourceRequest.ResponseURL)
	if nil != parsedURLErr {
		return parsedURLErr
	}

	status := "FAILED"
	if nil == responseErr {
		status = "SUCCESS"
	}
	reasonText := ""
	if nil != responseErr {
		reasonText = fmt.Sprintf("%s. Details in CloudWatch Logs: %s : %s",
			responseErr.Error(),
			customResourceRequest.LogGroupName,
			customResourceRequest.LogStreamName)
	} else {
		reasonText = fmt.Sprintf("Details in CloudWatch Logs: %s : %s",
			customResourceRequest.LogGroupName,
			customResourceRequest.LogStreamName)
	}

	responseData := map[string]interface{}{
		"Status":             status,
		"Reason":             reasonText,
		"PhysicalResourceId": customResourceRequest.PhysicalResourceID,
		"StackId":            customResourceRequest.StackID,
		"RequestId":          customResourceRequest.RequestID,
		"LogicalResourceId":  customResourceRequest.LogicalResourceID,
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
		return jsonError
	}

	responseBuffer := strings.NewReader(string(jsonData))
	req, httpErr := http.NewRequest("PUT", customResourceRequest.ResponseURL, responseBuffer)

	if nil != httpErr {
		return httpErr
	}
	// Need to use the Opaque field b/c Go will parse inline encoded values
	// which are supposed to be roundtripped to AWS.
	// Ref: https://tools.ietf.org/html/rfc3986#section-2.2
	// Ref: https://golang.org/pkg/net/url/#URL
	req.URL = &url.URL{
		Scheme:   parsedURL.Scheme,
		Host:     parsedURL.Host,
		Opaque:   parsedURL.RawPath,
		RawQuery: parsedURL.RawQuery,
	}
	logger.WithFields(logrus.Fields{
		"URL": req.URL,
	}).Debug("Created URL response")

	// Although it seems reasonable to set the Content-Type to "application/json" - don't.
	// The Content-Type must be an empty string in order for the
	// AWS Signature checker to pass.
	// Ref: http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-lambda-function-code.html
	req.Header.Set("Content-Type", "")

	client := &http.Client{}
	resp, httpErr := client.Do(req)
	if httpErr != nil {
		return httpErr
	}
	logger.WithFields(logrus.Fields{
		"LogicalResourceId":  customResourceRequest.LogicalResourceID,
		"Result":             responseData["Status"],
		"ResponseStatusCode": resp.StatusCode,
	}).Info("Sent CloudFormation response")

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, bodyErr := ioutil.ReadAll(resp.Body)
		if bodyErr != nil {
			logger.Warn("Unable to read body: " + bodyErr.Error())
			body = []byte{}
		}
		return fmt.Errorf("Error sending response: %d. Data: %s", resp.StatusCode, string(body))
	}
	defer resp.Body.Close()
	return nil
}

func customCommandForTypeName(resourceTypeName string, properties *[]byte) (interface{}, error) {
	var unmarshalError error
	var customCommand interface{}
	// ---------------------------------------------------------------------------
	// BEGIN - RESOURCE TYPES
	switch resourceTypeName {
	case HelloWorld:
		command := HelloWorldResource{
			GoAWSCustomResource: GoAWSCustomResource{
				GoAWSType: resourceTypeName,
			},
		}
		if nil != properties {
			unmarshalError = json.Unmarshal([]byte(string(*properties)), &command)
		}
		customCommand = &command
	case S3LambdaEventSource:
		command := S3LambdaEventSourceResource{
			GoAWSCustomResource: GoAWSCustomResource{
				GoAWSType: resourceTypeName,
			},
		}
		if nil != properties {
			unmarshalError = json.Unmarshal([]byte(string(*properties)), &command)
		}
		customCommand = &command
	case SNSLambdaEventSource:
		command := SNSLambdaEventSourceResource{
			GoAWSCustomResource: GoAWSCustomResource{
				GoAWSType: resourceTypeName,
			},
		}
		if nil != properties {
			unmarshalError = json.Unmarshal([]byte(string(*properties)), &command)
		}
		customCommand = &command
	case SESLambdaEventSource:
		command := SESLambdaEventSourceResource{
			GoAWSCustomResource: GoAWSCustomResource{
				GoAWSType: resourceTypeName,
			},
		}
		if nil != properties {
			unmarshalError = json.Unmarshal([]byte(string(*properties)), &command)
		}
		customCommand = &command
	case ZipToS3Bucket:
		command := ZipToS3BucketResource{
			GoAWSCustomResource: GoAWSCustomResource{
				GoAWSType: resourceTypeName,
			},
		}
		if nil != properties {
			unmarshalError = json.Unmarshal([]byte(string(*properties)), &command)
		}
		customCommand = &command
	case CloudWatchLogsLambdaEventSource:
		command := CloudWatchLogsLambdaEventSourceResource{
			GoAWSCustomResource: GoAWSCustomResource{
				GoAWSType: resourceTypeName,
			},
		}
		if nil != properties {
			unmarshalError = json.Unmarshal([]byte(string(*properties)), &command)
		}
		customCommand = &command
	}
	// END - RESOURCE TYPES
	// ---------------------------------------------------------------------------

	if unmarshalError != nil {
		return nil, fmt.Errorf("Failed to unmarshal properties for type: %s", resourceTypeName)
	}
	if nil == customCommand {
		return nil, fmt.Errorf("Failed to create custom command for type: %s", resourceTypeName)
	}
	return customCommand, nil
}

func customTypeProvider(resourceType string) gocf.ResourceProperties {
	commandInstance, commandError := customCommandForTypeName(resourceType, nil)
	if nil != commandError {
		return nil
	}
	resProperties, ok := commandInstance.(gocf.ResourceProperties)
	if !ok {
		return nil
	}
	return resProperties
}

func init() {
	gocf.RegisterCustomResourceProvider(customTypeProvider)
}

type logrusProxy struct {
	logger *logrus.Logger
}

func (proxy *logrusProxy) Log(args ...interface{}) {
	proxy.logger.Info(args...)
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

// cloudFormationResourceType a string for the resource name that represents a
// custom CloudFormation resource typename
func cloudFormationResourceType(resType string) string {
	return fmt.Sprintf("Custom::goAWS::%s", resType)
}
