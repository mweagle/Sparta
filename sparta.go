// Copyright (c) 2015 Matt Weagle <mweagle@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Per https://blog.golang.org/generate, generate the CONSTANTS.go
// file from the source resources
//go:generate go run ./vendor/github.com/mjibson/esc/main.go -o ./CONSTANTS.go -pkg sparta ./resources/index.js ./resources/mermaid/mermaid.css ./resources/mermaid/mermaid.min.js

package sparta

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/voxelbrain/goptions"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"runtime"
)

// RE for sanitizing golang/JS layer
var reSanitize = regexp.MustCompile("[\\.\\-\\s]+")

// Arbitrary JSON key-value object
type ArbitraryJSONObject map[string]interface{}

// Represents the untyped Event data provided via JSON
// to a Lambda handler.  See http://docs.aws.amazon.com/lambda/latest/dg/nodejs-prog-model-handler.html
// for more information
type LambdaEvent map[string]json.RawMessage

// Represents the Lambda Context object provided by the AWS Lambda runtime.
// See http://docs.aws.amazon.com/lambda/latest/dg/nodejs-prog-model-context.html
// for more information on field values.  Note that the golang version doesn't functions
// defined on the Context object.
type LambdaContext struct {
	AWSRequestId       string `json:"awsRequestId"`
	InvokeID           string `json:"invokeid"`
	LogGroupName       string `json:"logGroupName"`
	LogStreamName      string `json:"logStreamName"`
	FunctionName       string `json:"functionName"`
	MemoryLimitInMB    string `json:"memoryLimitInMB"`
	FunctionVersion    string `json:"functionVersion"`
	InvokedFunctionARN string `json:"invokedFunctionArn"`
}

// Package private type to deserialize NodeJS proxied
// Lambda Event and Context information
type lambdaRequest struct {
	Event   LambdaEvent   `json:"event"`
	Context LambdaContext `json:"context"`
}

// golang AWS Lambda handler function signature.  Standard HTTP response codes
// are used to signal AWS Lambda success/failure on the proxied context() object.
// See http://docs.aws.amazon.com/lambda/latest/dg/nodejs-prog-model-context.html for
// more information.
//
// 	200 - 299       : Success
// 	<200 || >= 300  : Failure
//
// Content written to the ResponseWriter will be used as the
// response/Error value provided to AWS Lambda.
type LambdaFunction func(*LambdaEvent, *LambdaContext, *http.ResponseWriter, *logrus.Logger)

// Additional options for lambda execution.  See the AWS Lambda FunctionConfiguration
// (http://docs.aws.amazon.com/lambda/latest/dg/API_FunctionConfiguration.html) docs
// for more information. Note that the "Runtime" field will be automatically set
// to "nodejs" (at least until golang is officially supported)
type LambdaFunctionOptions struct {
	Description string
	MemorySize  int64
	Timeout     int64
}

/*
params := &sns.SubscribeInput{
	Protocol: aws.String("protocol"), // Required
	TopicArn: aws.String("topicARN"), // Required
	Endpoint: aws.String("endpoint"),
}
// so we need two custom resources:
	- 1 to issue the subscription request via http://docs.aws.amazon.com/sdk-for-go/api/service/sns/SNS.html#Subscribe-instance_method
		- Needs to be parameterized with specific params for this function :(
	- 1 to ACK the subscription with http://docs.aws.amazon.com/sdk-for-go/api/service/sns/SNS.html#ConfirmSubscription-instance_method
		- Can be the same ACK for all subscriptions, needs http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-attribute-dependson.html
*/
// type SNSAddPermissionInput struct {
// 	lambda.AddPermissionInput
// 	TopicARN string
// }

// Represents data to provision a golang-based NodeJS AWS Lambda function
type LambdaAWSInfo struct {
	// internal function name, determined by reflection
	lambdaFnName string
	// pointer to lambda function
	lambdaFn LambdaFunction
	// Role name (NOT ARN) to use during AWS Lambda Execution.  See
	// the FunctionConfiguration (http://docs.aws.amazon.com/lambda/latest/dg/API_FunctionConfiguration.html)
	// docs for more info.
	ExecutionRoleName string
	// Additional exeuction options
	Options *LambdaFunctionOptions
	// Permissions to enable push-based Lambda execution.  See the
	// Permission Model docs (http://docs.aws.amazon.com/lambda/latest/dg/intro-permission-model.html)
	// for more information.
	Permissions []*lambda.AddPermissionInput
	// EventSource mappings to enable for pull-based Lambda execution.  See the
	// Event Source docs (http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html)
	// for more information
	EventSourceMappings []*lambda.CreateEventSourceMappingInput
}

// Return a JavaScript compatible function name to proxy a golang reflection-derived
// name
func (info *LambdaAWSInfo) jsHandlerName() string {
	return reSanitize.ReplaceAllString(info.lambdaFnName, "_")
}

// Marshal this object into 1 or more CloudFormation resource definitions that are accumulated
// in the resources map
func (info *LambdaAWSInfo) toCloudFormationResources(S3Bucket string, S3Key string, roleNameMap map[string]string, resources ArbitraryJSONObject) error {
	// Create the primary resource
	arn, exists := roleNameMap[info.ExecutionRoleName]
	if !exists {
		return errors.New("Unable to find ARN for role: " + info.ExecutionRoleName)
	}
	primaryResource := ArbitraryJSONObject{
		"Type": "AWS::Lambda::Function",
		"Properties": ArbitraryJSONObject{
			"Code": ArbitraryJSONObject{
				"S3Bucket": S3Bucket,
				"S3Key":    S3Key,
			},
			"Description": info.Options.Description,
			"Handler":     fmt.Sprintf("index.%s", info.jsHandlerName()),
			"MemorySize":  info.Options.MemorySize,
			"Role":        arn,
			"Runtime":     "nodejs",
			"Timeout":     info.Options.Timeout},
	}

	// Get the resource name we're going to use s.t. we can tie it to the rest of the
	// lambda definition
	hash := sha1.New()
	hash.Write([]byte(info.lambdaFnName))
	resourceName := fmt.Sprintf("Lambda%s", hex.EncodeToString(hash.Sum(nil)))
	resources[resourceName] = primaryResource

	// Create the lambda Ref in case we need a permission or event mapping
	functionAttr := ArbitraryJSONObject{
		"Fn::GetAtt": []string{resourceName, "Arn"},
	}

	// Permissions
	for _, eachPermission := range info.Permissions {

		properties := ArbitraryJSONObject{
			"Action":       eachPermission.Action,
			"FunctionName": functionAttr,
			"Principal":    eachPermission.Principal,
		}
		if nil != eachPermission.SourceAccount {
			properties["SourceAccount"] = *eachPermission.SourceAccount
		}
		if nil != eachPermission.SourceArn {
			properties["SourceArn"] = *eachPermission.SourceArn
		}

		primaryPermission := ArbitraryJSONObject{
			"Type":       "AWS::Lambda::Permission",
			"Properties": properties,
		}
		hash := sha1.New()
		if nil != eachPermission.Principal {
			hash.Write([]byte(*eachPermission.Principal))
		}
		if nil != eachPermission.SourceAccount {
			hash.Write([]byte(*eachPermission.SourceAccount))
		}
		if nil != eachPermission.SourceArn {
			hash.Write([]byte(*eachPermission.SourceArn))
		}
		resourceName := fmt.Sprintf("LambdaPerm%s", hex.EncodeToString(hash.Sum(nil)))
		resources[resourceName] = primaryPermission
	}

	// Event Source Mappings
	for _, eachEventSourceMapping := range info.EventSourceMappings {
		properties := ArbitraryJSONObject{
			"EventSourceArn":   eachEventSourceMapping.EventSourceArn,
			"FunctionName":     functionAttr,
			"StartingPosition": eachEventSourceMapping.StartingPosition,
			"BatchSize":        eachEventSourceMapping.BatchSize,
		}
		if nil != eachEventSourceMapping.Enabled {
			properties["Enabled"] = *eachEventSourceMapping.Enabled
		}

		primaryEventSourceMapping := ArbitraryJSONObject{
			"Type":       "AWS::Lambda::EventSourceMapping",
			"Properties": properties,
		}
		hash := sha1.New()
		hash.Write([]byte(*eachEventSourceMapping.EventSourceArn))
		binary.Write(hash, binary.LittleEndian, *eachEventSourceMapping.BatchSize)
		hash.Write([]byte(*eachEventSourceMapping.StartingPosition))
		resourceName := fmt.Sprintf("LambdaES%s", hex.EncodeToString(hash.Sum(nil)))
		resources[resourceName] = primaryEventSourceMapping
	}
	return nil
}

// Sanitize a function name
func sanitizedName(input string) string {
	return reSanitize.ReplaceAllString(input, "_")
}

// Returns a new LambdaAWSInfo struct capable of being provisioned
func NewLambda(executionRoleName string, fn LambdaFunction, lambdaOptions *LambdaFunctionOptions) *LambdaAWSInfo {
	lambdaPtr := runtime.FuncForPC(reflect.ValueOf(fn).Pointer())
	lambda := &LambdaAWSInfo{
		lambdaFnName:        lambdaPtr.Name(),
		lambdaFn:            fn,
		Options:             lambdaOptions,
		Permissions:         make([]*lambda.AddPermissionInput, 0),
		EventSourceMappings: make([]*lambda.CreateEventSourceMappingInput, 0),
		ExecutionRoleName:   executionRoleName,
	}

	// Defaults
	if nil == lambda.Options {
		lambda.Options = &LambdaFunctionOptions{"", 128, 3}
	}
	if lambda.Options.MemorySize <= 0 {
		lambda.Options.MemorySize = 128
	}
	if lambda.Options.Timeout <= 0 {
		lambda.Options.Timeout = 3
	}
	return lambda
}

// Returns a new logrus.Logger instance. It is the caller's responsibility
// to set the formatter if needed.
func NewLogger(level string) (*logrus.Logger, error) {
	logger := logrus.New()
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, err
	}
	logger.Level = logLevel
	return logger, nil
}

// Primary handler for transforming an application into a Lambda package.  The
// serviceName is used to uniquely identify your service within a region and will
// be used for subsequent updates.  For provisioning, ensure that you've
// properly configured AWS credentials for the golang SDK.
// See http://docs.aws.amazon.com/sdk-for-go/api/aws/defaults.html#DefaultChainCredentials-constant
// for more information.
func Main(serviceName string, serviceDescription string, lambdaAWSInfos []*LambdaAWSInfo) error {

	options := struct {
		LogLevel string        `goptions:"-l, --level, description='Log level [panic, fatal, error, warn, info, debug]'"`
		Help     goptions.Help `goptions:"-h, --help, description='Show this help'"`

		Verb      goptions.Verbs
		Provision struct {
			S3Bucket string `goptions:"-b,--s3Bucket, description='S3 Bucket to use for Lambda source', obligatory"`
		} `goptions:"provision"`
		Delete struct {
		} `goptions:"delete"`
		Execute struct {
			Port            int `goptions:"-p,--port, description='Alternative port for HTTP binding (default=9999)'"`
			SignalParentPID int `goptions:"-s,--signal, description='Process ID to signal with SIGUSR2 once ready'"`
		} `goptions:"execute"`
		Describe struct {
			OutputFile string `goptions:"-o,--out, description='Output file for HTML description', obligatory"`
		} `goptions:"describe"`
		Explore struct {
		} `goptions:"explore"`
	}{ // Default values goes here
		LogLevel: "info",
	}
	goptions.ParseAndFail(&options)
	logger, err := NewLogger(options.LogLevel)
	if err != nil {
		goptions.PrintHelp()
		os.Exit(1)
	}
	switch options.Verb {
	case "provision":
		logger.Formatter = new(logrus.TextFormatter)
		return Provision(serviceName, serviceDescription, lambdaAWSInfos, options.Provision.S3Bucket, logger) //"LambdaExecutor", "weagle")
	case "execute":
		logger.Formatter = new(logrus.JSONFormatter)
		return Execute(lambdaAWSInfos, options.Execute.Port, options.Execute.SignalParentPID, logger)
	case "delete":
		logger.Formatter = new(logrus.TextFormatter)
		return Delete(serviceName, logger)
	case "explore":
		logger.Formatter = new(logrus.TextFormatter)
		return Explore(serviceName, logger)
	case "describe":
		logger.Formatter = new(logrus.TextFormatter)
		fileWriter, err := os.Create(options.Describe.OutputFile)
		if err != nil {
			return errors.New(fmt.Sprintf("Failed to open %s output. Error: %s", options.Describe.OutputFile, err))
		}
		defer fileWriter.Close()
		return Describe(serviceName, serviceDescription, lambdaAWSInfos, fileWriter, logger)
	default:
		goptions.PrintHelp()
		return errors.New("Unsupported subcommand: " + string(options.Verb))
	}
}
