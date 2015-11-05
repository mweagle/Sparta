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

package sparta

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/voxelbrain/goptions"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var AssumePolicyDocument = ArbitraryJSONObject{
	"Version": "2012-10-17",
	"Statement": []ArbitraryJSONObject{
		{
			"Effect": "Allow",
			"Principal": ArbitraryJSONObject{
				"Service": []string{"lambda.amazonaws.com"},
			},
			"Action": []string{"sts:AssumeRole"},
		},
		{
			"Effect": "Allow",
			"Principal": ArbitraryJSONObject{
				"Service": []string{"ec2.amazonaws.com"},
			},
			"Action": []string{"sts:AssumeRole"},
		},
	},
}

var CommonIAMStatements = []ArbitraryJSONObject{
	{
		"Action":   []string{"logs:CreateLogGroup", "logs:CreateLogStream", "logs:PutLogEvents"},
		"Effect":   "Allow",
		"Resource": "arn:aws:logs:*:*:*",
	},
}

// RE for sanitizing golang/JS layer
var reSanitize = regexp.MustCompile("[\\.\\-\\s]+")

// Arbitrary JSON key-value object
type ArbitraryJSONObject map[string]interface{}

// Represents the untyped Event data provided via JSON
// to a Lambda handler.  See http://docs.aws.amazon.com/lambda/latest/dg/nodejs-prog-model-handler.html
// for more information
type LambdaEvent interface{}

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

type TemplateDecorator func(roleNameMap map[string]string, template ArbitraryJSONObject, logger *logrus.Logger)

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

////////////////////////////////////////////////////////////////////////////////
// Types to handle permissions & push source configuration
type LambdaPermissionExporter interface {
	export(targetLambdaFuncRef interface{},
		resources ArbitraryJSONObject,
		logger *logrus.Logger) (string, error)
	descriptionInfo() (string, string)
}

////////////////////////////////////////////////////////////////////////////////
// BasePermission
////////////////////////////////////////////////////////////////////////////////
///
type BasePermission struct {
	StatementId   string `json:"StatementId,omitempty"`
	SourceAccount string `json:"SourceAccount,omitempty"`
	SourceArn     string `json:"SourceArn,omitempty"`
	Qualifier     string `json:"Qualifier,omitempty"`
}

func (perm BasePermission) export(principal string, targetLambdaFuncRef interface{}, resources ArbitraryJSONObject, logger *logrus.Logger) (string, error) {
	properties := ArbitraryJSONObject{
		"Action":       "lambda:InvokeFunction",
		"FunctionName": targetLambdaFuncRef,
		"Principal":    principal,
	}
	if "" != perm.SourceAccount {
		properties["SourceAccount"] = perm.SourceAccount
	}
	if "" != perm.SourceArn {
		properties["SourceArn"] = perm.SourceArn
	}

	primaryPermission := ArbitraryJSONObject{
		"Type":       "AWS::Lambda::Permission",
		"Properties": properties,
	}
	hash := sha1.New()
	hash.Write([]byte(principal))

	if "" != perm.SourceAccount {
		hash.Write([]byte(perm.SourceAccount))
	}
	if "" != perm.SourceArn {
		hash.Write([]byte(perm.SourceArn))
	}
	resourceName := fmt.Sprintf("LambdaPerm%s", hex.EncodeToString(hash.Sum(nil)))
	resources[resourceName] = primaryPermission
	return resourceName, nil
}

func (perm BasePermission) descriptionInfo(b *bytes.Buffer, logger *logrus.Logger) error {
	return errors.New("Describe not implemented")
}

type LambdaPermission struct {
	BasePermission
	Principal string
}

func (perm LambdaPermission) export(targetLambdaFuncRef interface{},
	resources ArbitraryJSONObject,
	logger *logrus.Logger) (string, error) {
	return perm.BasePermission.export(perm.Principal, targetLambdaFuncRef, resources, logger)
}

func (perm LambdaPermission) descriptionInfo() (string, string) {
	return "Source", perm.BasePermission.SourceArn
}

////////////////////////////////////////////////////////////////////////////////
// S3Permission
////////////////////////////////////////////////////////////////////////////////
///
type S3Permission struct {
	BasePermission
	Events []string                           `json:"Events,omitempty"`
	Filter s3.NotificationConfigurationFilter `json:"Filter,omitempty"`
}

func (perm S3Permission) bucketName() string {
	bucketParts := strings.Split(perm.BasePermission.SourceArn, ":")
	return bucketParts[len(bucketParts)-1]
}

func (perm S3Permission) export(targetLambdaFuncRef interface{}, resources ArbitraryJSONObject, logger *logrus.Logger) (string, error) {
	targetLambdaResourceName, err := perm.BasePermission.export("s3.amazonaws.com", targetLambdaFuncRef, resources, logger)
	if nil != err {
		return "", err
	}

	// Make sure the custom lambda that manages s3 notifications is provisioned.
	configuratorResName, err := ensureConfiguratorLambdaResource("s3.amazonaws.com", perm.SourceArn, resources, logger)
	if nil != err {
		return "", err
	}
	permissionData := ArbitraryJSONObject{
		"Events": perm.Events,
	}
	if nil != perm.Filter.Key {
		permissionData["Filter"] = perm.Filter
	}

	// Add a custom resource invocation for this configuration
	//////////////////////////////////////////////////////////////////////////////
	// And finally the custom resource forwarder

	customResourceInvoker := ArbitraryJSONObject{
		"Type":    "AWS::CloudFormation::CustomResource",
		"Version": "1.0",
		"Properties": ArbitraryJSONObject{
			"ServiceToken": ArbitraryJSONObject{
				"Fn::GetAtt": []string{configuratorResName, "Arn"},
			},
			"Permission": permissionData,
			// Use the LambdaTarget value in the JS custom resoruce
			// handler to create the ID used to manage S3 notifications
			"LambdaTarget": targetLambdaFuncRef,
			"Bucket":       perm.bucketName(),
		},
		"DependsOn": []string{targetLambdaResourceName, configuratorResName},
	}
	// Save it
	resourceInvokerName := cloudFormationResourceName(fmt.Sprintf("ConfigS3%s", targetLambdaResourceName))
	resources[resourceInvokerName] = customResourceInvoker
	return "", nil
}

func (perm S3Permission) descriptionInfo() (string, string) {
	return perm.BasePermission.SourceArn, fmt.Sprintf("%s", perm.Events)
}

////////////////////////////////////////////////////////////////////////////////
// SNSPermission
////////////////////////////////////////////////////////////////////////////////
///
type SNSPermission struct {
	BasePermission
}

func (perm SNSPermission) topicName() string {
	topicParts := strings.Split(perm.BasePermission.SourceArn, ":")
	return topicParts[len(topicParts)-1]
}

func (perm SNSPermission) export(targetLambdaFuncRef interface{}, resources ArbitraryJSONObject, logger *logrus.Logger) (string, error) {
	targetLambdaResourceName, err := perm.BasePermission.export("sns.amazonaws.com", targetLambdaFuncRef, resources, logger)
	if nil != err {
		return "", err
	}

	// Make sure the custom lambda that manages SNS notifications is provisioned.
	configuratorResName, err := ensureConfiguratorLambdaResource("sns.amazonaws.com", perm.SourceArn, resources, logger)
	if nil != err {
		return "", err
	}

	// Add a custom resource invocation for this configuration
	//////////////////////////////////////////////////////////////////////////////
	// And the custom resource forwarder
	customResourceSubscriber := ArbitraryJSONObject{
		"Type":    "AWS::CloudFormation::CustomResource",
		"Version": "1.0",
		"Properties": ArbitraryJSONObject{
			"ServiceToken": ArbitraryJSONObject{
				"Fn::GetAtt": []string{configuratorResName, "Arn"},
			},
			"Mode":     "Subscribe",
			"TopicArn": perm.BasePermission.SourceArn,
			// Use the LambdaTarget value in the JS custom resoruce
			// handler to create the ID used to manage S3 notifications
			"LambdaTarget": targetLambdaFuncRef,
		},
		"DependsOn": []string{targetLambdaResourceName, configuratorResName},
	}
	// Save it
	subscriberResourceName := cloudFormationResourceName(fmt.Sprintf("SubscriberSNS%s", targetLambdaResourceName))
	resources[subscriberResourceName] = customResourceSubscriber

	//////////////////////////////////////////////////////////////////////////////
	// And the custom resource unsubscriber
	customResourceUnsubscriber := ArbitraryJSONObject{
		"Type":    "AWS::CloudFormation::CustomResource",
		"Version": "1.0",
		"Properties": ArbitraryJSONObject{
			"ServiceToken": ArbitraryJSONObject{
				"Fn::GetAtt": []string{configuratorResName, "Arn"},
			},
			"Mode": "Unsubscribe",
			"SubscriptionArn": ArbitraryJSONObject{
				"Fn::GetAtt": []string{subscriberResourceName, "SubscriptionArn"},
			},
			"TopicArn": perm.BasePermission.SourceArn,
			// Use the LambdaTarget value in the JS custom resoruce
			// handler to create the ID used to manage S3 notifications
			"LambdaTarget": targetLambdaFuncRef,
		},
		"DependsOn": []string{subscriberResourceName},
	}
	// Save it
	unsubscriberResourceName := cloudFormationResourceName(fmt.Sprintf("UnsubscriberSNS%s", targetLambdaResourceName))
	resources[unsubscriberResourceName] = customResourceUnsubscriber

	return "", nil
}

func (perm SNSPermission) descriptionInfo() (string, string) {
	return perm.BasePermission.SourceArn, ""
}

////////////////////////////////////////////////////////////////////////////////
// START - IAM Role handlers
////////////////////////////////////////////////////////////////////////////////

type IAMRolePrivilege struct {
	Actions  []string
	Resource string
}

type IAMRoleDefinition struct {
	Privileges []IAMRolePrivilege
}

func (roleDefinition *IAMRoleDefinition) rolePolicy() ArbitraryJSONObject {
	statements := CommonIAMStatements
	for _, eachPrivilege := range roleDefinition.Privileges {
		statements = append(statements, ArbitraryJSONObject{
			"Effect":   "Allow",
			"Action":   eachPrivilege.Actions,
			"Resource": eachPrivilege.Resource,
		})
	}

	iamPolicy := ArbitraryJSONObject{"Type": "AWS::IAM::Role",
		"Properties": ArbitraryJSONObject{
			"AssumeRolePolicyDocument": AssumePolicyDocument,
			"Policies": []ArbitraryJSONObject{
				{
					"PolicyName": "lambdaRole",
					"PolicyDocument": ArbitraryJSONObject{
						"Version":   "2012-10-17",
						"Statement": statements,
					},
				},
			},
		},
	}
	return iamPolicy
}

func (roleDefinition *IAMRoleDefinition) logicalName() string {
	hash := sha1.New()
	hash.Write([]byte(fmt.Sprintf("%s", roleDefinition.Privileges)))
	return fmt.Sprintf("IAMRole%s", hex.EncodeToString(hash.Sum(nil)))
}

////////////////////////////////////////////////////////////////////////////////
// START - LambdaAWSInfo
////////////////////////////////////////////////////////////////////////////////
//
// Represents data to provision a golang-based NodeJS AWS Lambda function
type LambdaAWSInfo struct {
	// internal function name, determined by reflection
	lambdaFnName string
	// pointer to lambda function
	lambdaFn LambdaFunction
	// Role name (NOT ARN) to use during AWS Lambda Execution.  See
	// the FunctionConfiguration (http://docs.aws.amazon.com/lambda/latest/dg/API_FunctionConfiguration.html)
	// docs for more info.
	RoleName string
	// Role provider, either as string or a role definition
	RoleDefinition *IAMRoleDefinition
	// Additional exeuction options
	Options *LambdaFunctionOptions
	// Permissions to enable push-based Lambda execution.  See the
	// Permission Model docs (http://docs.aws.amazon.com/lambda/latest/dg/intro-permission-model.html)
	// for more information.
	Permissions []LambdaPermissionExporter
	// EventSource mappings to enable for pull-based Lambda execution.  See the
	// Event Source docs (http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html)
	// for more information
	EventSourceMappings []*lambda.CreateEventSourceMappingInput
	// TODO: Provide hook for TemplateDecorator
}

// Return a JavaScript compatible function name to proxy a golang reflection-derived
// name
func (info *LambdaAWSInfo) jsHandlerName() string {
	return reSanitize.ReplaceAllString(info.lambdaFnName, "_")
}

// Marshal this object into 1 or more CloudFormation resource definitions that are accumulated
// in the resources map
func (info *LambdaAWSInfo) export(S3Bucket string,
	S3Key string,
	roleNameMap map[string]interface{},
	resources ArbitraryJSONObject,
	logger *logrus.Logger) error {

	// If we have RoleName, then get the ARN, otherwise get the Ref
	dependsOn := make([]string, 0)

	iamRoleArnName := info.RoleName
	if iamRoleArnName == "" {
		iamRoleArnName = info.RoleDefinition.logicalName()
		dependsOn = append(dependsOn, iamRoleArnName)
	}

	// Create the primary resource
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
			"Role":        roleNameMap[iamRoleArnName],
			"Runtime":     "nodejs",
			"Timeout":     info.Options.Timeout,
		},
		"DependsOn": dependsOn,
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
		_, err := eachPermission.export(functionAttr, resources, logger)
		if nil != err {
			return err
		}
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

//
// END - LambdaAWSInfo
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// Private
////////////////////////////////////////////////////////////////////////////////
// Sanitize a function name
func sanitizedName(input string) string {
	return reSanitize.ReplaceAllString(input, "_")
}

func awsSession(logger *logrus.Logger) *session.Session {
	sess := session.New()
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

func init() {
	rand.Seed(time.Now().Unix())
}

func cloudFormationResourceName(prefix string) string {
	randValue := rand.Int63()
	hash := sha1.New()
	hash.Write([]byte(prefix))
	hash.Write([]byte(strconv.FormatInt(randValue, 10)))
	return fmt.Sprintf("%s%s", prefix, hex.EncodeToString(hash.Sum(nil)))
}

////////////////////////////////////////////////////////////////////////////////
// Public
////////////////////////////////////////////////////////////////////////////////

// Returns a new LambdaAWSInfo struct capable of being provisioned
func NewLambda(roleNameOrIAMRoleDefinition interface{}, fn LambdaFunction, lambdaOptions *LambdaFunctionOptions) *LambdaAWSInfo {
	if nil == lambdaOptions {
		lambdaOptions = &LambdaFunctionOptions{}
	}
	lambdaPtr := runtime.FuncForPC(reflect.ValueOf(fn).Pointer())
	lambda := &LambdaAWSInfo{
		lambdaFnName:        lambdaPtr.Name(),
		lambdaFn:            fn,
		Options:             lambdaOptions,
		Permissions:         make([]LambdaPermissionExporter, 0),
		EventSourceMappings: make([]*lambda.CreateEventSourceMappingInput, 0),
	}

	switch roleNameOrIAMRoleDefinition.(type) {
	case string:
		lambda.RoleName = roleNameOrIAMRoleDefinition.(string)
	case IAMRoleDefinition:
		definition := roleNameOrIAMRoleDefinition.(IAMRoleDefinition)
		lambda.RoleDefinition = &definition
	default:
		fmt.Println("unknown")
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

	// We need to be able to provision an IAM role that has capabilities to
	// manage the other sources.  That'll give us the role arn to use in the custom
	// resource execution.
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
		return Provision(serviceName, serviceDescription, lambdaAWSInfos, options.Provision.S3Bucket, logger)
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
