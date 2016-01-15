package sparta

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	spartaPrivate "github.com/mweagle/Sparta/private"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/voxelbrain/goptions"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// SpartaVersion defines the current Sparta release
const SpartaVersion = "0.1.6"

// ArbitraryJSONObject represents an untyped key-value object. CloudFormation resource representations
// are aggregated as []ArbitraryJSONObject before being marsharled to JSON
// for API operations.
type ArbitraryJSONObject map[string]interface{}

// AWS Principal ARNs from http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html
const (
	// @enum APIGatewayPrincipal
	APIGatewayPrincipal = "apigateway.amazonaws.com"
	// @enum AWSPrincipal
	S3Principal = "s3.amazonaws.com"
	// @enum AWSPrincipal
	SESPrincipal = "ses.amazonaws.com"
	// @enum AWSPrincipal
	SNSPrincipal = "sns.amazonaws.com"
	// @enum AWSPrincipal
	EC2Principal = "ec2.amazonaws.com"
	// @enum AWSPrincipal
	LambdaPrincipal = "lambda.amazonaws.com"
)

const (
	// WildcardArn that applies to AWS resources which don't use ARNs to scope access (eg, SES)
	WildcardArn = "*"
)

// AssumePolicyDocument defines common a IAM::Role PolicyDocument
// used as part of IAM::Role resource definitions
var AssumePolicyDocument = ArbitraryJSONObject{
	"Version": "2012-10-17",
	"Statement": []ArbitraryJSONObject{
		{
			"Effect": "Allow",
			"Principal": ArbitraryJSONObject{
				"Service": []string{LambdaPrincipal},
			},
			"Action": []string{"sts:AssumeRole"},
		},
		{
			"Effect": "Allow",
			"Principal": ArbitraryJSONObject{
				"Service": []string{EC2Principal},
			},
			"Action": []string{"sts:AssumeRole"},
		},
		{
			"Effect": "Allow",
			"Principal": ArbitraryJSONObject{
				"Service": []string{APIGatewayPrincipal},
			},
			"Action": []string{"sts:AssumeRole"},
		},
	},
}

// Represents the CloudFormation Arn of this stack, referenced
// in CommonIAMStatements
var cfArn = []interface{}{"arn:aws:cloudformation:",
	ArbitraryJSONObject{
		"Ref": "AWS::Region",
	},
	":",
	ArbitraryJSONObject{
		"Ref": "AWS::AccountId",
	},
	":stack/",
	ArbitraryJSONObject{
		"Ref": "AWS::StackName",
	},
	"/*"}

// CommonIAMStatements defines common IAM::Role Policy Statement values for different AWS
// service types.  See http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html#genref-aws-service-namespaces
// for names.
// http://docs.aws.amazon.com/lambda/latest/dg/monitoring-functions.html
// for more information.
var CommonIAMStatements = map[string][]ArbitraryJSONObject{
	"core": []ArbitraryJSONObject{
		ArbitraryJSONObject{
			"Action": []string{"logs:CreateLogGroup",
				"logs:CreateLogStream",
				"logs:PutLogEvents"},
			"Effect":   "Allow",
			"Resource": "arn:aws:logs:*:*:*",
		},
		ArbitraryJSONObject{
			"Action":   []string{"cloudwatch:PutMetricData"},
			"Effect":   "Allow",
			"Resource": WildcardArn,
		},
		ArbitraryJSONObject{
			"Effect": "Allow",
			"Action": []string{"cloudformation:DescribeStacks"},
			"Resource": ArbitraryJSONObject{
				"Fn::Join": []interface{}{"", cfArn},
			},
		},
	},
	"dynamodb": []ArbitraryJSONObject{
		ArbitraryJSONObject{"Effect": "Allow",
			"Action": []string{"dynamodb:DescribeStream",
				"dynamodb:GetRecords",
				"dynamodb:GetShardIterator",
				"dynamodb:ListStreams",
			},
		}},
	"kinesis": []ArbitraryJSONObject{
		ArbitraryJSONObject{
			"Effect": "Allow",
			"Action": []string{"kinesis:GetRecords",
				"kinesis:GetShardIterator",
				"kinesis:DescribeStream",
				"kinesis:ListStreams",
			},
		},
	},
}

// RE for sanitizing golang/JS layer
var reSanitize = regexp.MustCompile("[\\.\\-\\s]+")

// LambdaContext defines the AWS Lambda Context object provided by the AWS Lambda runtime.
// See http://docs.aws.amazon.com/lambda/latest/dg/nodejs-prog-model-context.html
// for more information on field values.  Note that the golang version doesn't functions
// defined on the Context object.
type LambdaContext struct {
	AWSRequestID       string `json:"awsRequestId"`
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
	Event   json.RawMessage `json:"event"`
	Context LambdaContext   `json:"context"`
}

// LambdaFunction is the golang function signature required to support AWS Lambda execution.
// Standard HTTP response codes are used to signal AWS Lambda success/failure on the
// proxied context() object.  See http://docs.aws.amazon.com/lambda/latest/dg/nodejs-prog-model-context.html
// for more information.
//
// 	200 - 299       : Success
// 	<200 || >= 300  : Failure
//
// Content written to the ResponseWriter will be used as the
// response/Error value provided to AWS Lambda.
type LambdaFunction func(*json.RawMessage, *LambdaContext, http.ResponseWriter, *logrus.Logger)

// LambdaFunctionOptions defines additional AWS Lambda execution params.  See the
// AWS Lambda FunctionConfiguration (http://docs.aws.amazon.com/lambda/latest/dg/API_FunctionConfiguration.html)
// docs for more information. Note that the "Runtime" field will be automatically set
// to "nodejs" (at least until golang is officially supported)
type LambdaFunctionOptions struct {
	// Additional function description
	Description string
	// Memory limit
	MemorySize int64
	// Timeout (seconds)
	Timeout int64
}

// TemplateDecorator allows Lambda functions to annotate the CloudFormation
// template definition.  Both the resources and the outputs params
// are initialized to an empty ArbitraryJSONObject and should
// be populated with valid CloudFormation ArbitraryJSONObject values.  The
// CloudFormationResourceName() function can be used to generate
// logical CloudFormation-compatible resource names.
// See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html and
// http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/outputs-section-structure.html for
// more information.
type TemplateDecorator func(lambdaResourceName string,
	lambdaResourceDefinition ArbitraryJSONObject,
	resources ArbitraryJSONObject,
	outputs ArbitraryJSONObject,
	logger *logrus.Logger) error

////////////////////////////////////////////////////////////////////////////////
// Types to handle permissions & push source configuration

// LambdaPermissionExporter defines an interface for polymorphic collection of
// Permission entries that support specialization for additional resource generation.
type LambdaPermissionExporter interface {
	// Export the permission object to a set of CloudFormation resources
	// in the provided resources param.  The targetLambdaFuncRef
	// interface represents the Fn::GetAtt "Arn" JSON value
	// of the parent Lambda target
	export(serviceName string,
		targetLambdaFuncRef interface{},
		resources ArbitraryJSONObject,
		outputs ArbitraryJSONObject,
		S3Bucket string,
		S3Key string,
		logger *logrus.Logger) (string, error)
	// Return a `describe` compatible output for the given permission
	descriptionInfo() (string, string)
}

////////////////////////////////////////////////////////////////////////////////
// START - BasePermission
//

// BasePermission (http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-permission.html)
// type for common AWS Lambda permission data.
type BasePermission struct {
	// The AWS account ID (without hyphens) of the source owner
	SourceAccount string `json:"SourceAccount,omitempty"`
	// The ARN of a resource that is invoking your function.
	SourceArn interface{} `json:"SourceArn,omitempty"`
}

func (perm *BasePermission) sourceArnExpr(joinParts []interface{}) interface{} {
	switch perm.SourceArn.(type) {
	case string:
		return perm.SourceArn
	case ArbitraryJSONObject:
		if nil == joinParts {
			joinParts = make([]interface{}, 0)
		}
		joinParts = append(joinParts, perm.SourceArn)
		return ArbitraryJSONObject{
			"Fn::Join": []interface{}{"", joinParts},
		}
	default:
		return "Invalid SourceArn type"
	}
}

func (perm *BasePermission) sourceArnString() string {
	switch arn := perm.SourceArn.(type) {
	case string:
		return arn
	case ArbitraryJSONObject:
		return fmt.Sprintf("Ref:%s", arn["Ref"])
	default:
		return "Invalid SourceArn type"
	}
}

func (perm BasePermission) export(principal string,
	targetLambdaFuncRef interface{},
	sourceArnExpression interface{},
	resources ArbitraryJSONObject,
	S3Bucket string,
	S3Key string,
	logger *logrus.Logger) (string, error) {
	properties := ArbitraryJSONObject{
		"Action":       "lambda:InvokeFunction",
		"FunctionName": targetLambdaFuncRef,
		"Principal":    principal,
	}

	if arnValue, ok := perm.SourceArn.(string); ok {
		if WildcardArn != arnValue {
			properties["SourceArn"] = arnValue
		}
	}

	if "" != perm.SourceAccount {
		properties["SourceAccount"] = perm.SourceAccount
	}
	primaryPermission := ArbitraryJSONObject{
		"Type":       "AWS::Lambda::Permission",
		"Properties": properties,
	}
	hash := sha1.New()
	hash.Write([]byte(fmt.Sprintf("%v", properties)))
	resourceName := fmt.Sprintf("LambdaPerm%s", hex.EncodeToString(hash.Sum(nil)))
	resources[resourceName] = primaryPermission
	return resourceName, nil
}

func (perm BasePermission) descriptionInfo(b *bytes.Buffer, logger *logrus.Logger) error {
	return errors.New("Describe not implemented")
}

//
// END - BasePermission
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// START - LambdaPermission
//
var lambdaSourceArnParts = []interface{}{"arn:aws:lambda:",
	ArbitraryJSONObject{
		"Ref": "AWS::Region",
	},
	":function:"}

// LambdaPermission type that creates a Lambda::Permission entry
// in the generated template, but does NOT automatically register the lambda
// with the BasePermission.SourceArn.  Typically used to register lambdas with
// externally managed event producers
type LambdaPermission struct {
	BasePermission
	// The entity for which you are granting permission to invoke the Lambda function
	Principal string
}

func (perm LambdaPermission) export(serviceName string,
	targetLambdaFuncRef interface{},
	resources ArbitraryJSONObject,
	outputs ArbitraryJSONObject,
	S3Bucket string,
	S3Key string,
	logger *logrus.Logger) (string, error) {

	// arn:aws:lambda:us-west-2:027159405834:function:myLambdaTest

	return perm.BasePermission.export(perm.Principal,
		targetLambdaFuncRef,
		perm.BasePermission.sourceArnExpr(lambdaSourceArnParts),
		resources,
		S3Bucket,
		S3Key,
		logger)
}

func (perm LambdaPermission) descriptionInfo() (string, string) {
	return "Source", perm.sourceArnString()
}

//
// END - LambdaPermission
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// START - S3Permission
//

var s3SourceArnParts = []interface{}{"arn:aws:s3:::"}

// S3Permission struct that imples the S3 BasePermission.SourceArn should be
// updated (via PutBucketNotificationConfiguration) to automatically push
// events to the owning Lambda.
// See http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html#intro-core-components-event-sources
// for more information.
type S3Permission struct {
	BasePermission
	// S3 events to register for (eg: `[]string{"s3:ObjectCreated:*", "s3:ObjectRemoved:*"}`).
	Events []string `json:"Events,omitempty"`
	// S3.NotificationConfigurationFilter
	// to scope event forwarding.  See
	// 		http://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html
	// for more information.
	Filter s3.NotificationConfigurationFilter `json:"Filter,omitempty"`
}

func (perm S3Permission) export(serviceName string,
	targetLambdaFuncRef interface{},
	resources ArbitraryJSONObject,
	outputs ArbitraryJSONObject,
	S3Bucket string,
	S3Key string,
	logger *logrus.Logger) (string, error) {

	sourceArnExpression := perm.BasePermission.sourceArnExpr(s3SourceArnParts)
	targetLambdaResourceName, err := perm.BasePermission.export(S3Principal,
		targetLambdaFuncRef,
		sourceArnExpression,
		resources,
		S3Bucket,
		S3Key,
		logger)
	if nil != err {
		return "", err
	}

	// Make sure the custom lambda that manages s3 notifications is provisioned.
	configuratorResName, err := ensureConfiguratorLambdaResource(S3Principal,
		sourceArnExpression,
		[]string{},
		resources,
		S3Bucket,
		S3Key,
		logger)
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
			"BucketArn":    sourceArnExpression,
		},
		"DependsOn": []string{targetLambdaResourceName, configuratorResName},
	}
	// Need to use a stable name s.t that the CloudFormation update event
	// is sent rather than a series of create/delete events that refer to the
	// same Lambda Target
	resourceInvokerName := CloudFormationResourceName("ConfigS3",
		targetLambdaResourceName,
		perm.BasePermission.SourceAccount,
		fmt.Sprintf("%v", sourceArnExpression))
	resources[resourceInvokerName] = customResourceInvoker
	return "", nil
}

func (perm S3Permission) descriptionInfo() (string, string) {
	return perm.sourceArnString(), fmt.Sprintf("%s", perm.Events)
}

//
// END - S3Permission
///////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// SNSPermission - START
var snsSourceArnParts = make([]interface{}, 0)

// SNSPermission struct that imples the S3 BasePermission.SourceArn should be
// updated (via PutBucketNotificationConfiguration) to automatically push
// events to the parent Lambda.
// See http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html#intro-core-components-event-sources
// for more information.
type SNSPermission struct {
	BasePermission
}

func (perm SNSPermission) export(serviceName string,
	targetLambdaFuncRef interface{},
	resources ArbitraryJSONObject,
	outputs ArbitraryJSONObject,
	S3Bucket string,
	S3Key string,
	logger *logrus.Logger) (string, error) {
	sourceArnExpression := perm.BasePermission.sourceArnExpr(snsSourceArnParts)

	targetLambdaResourceName, err := perm.BasePermission.export(SNSPrincipal,
		targetLambdaFuncRef,
		sourceArnExpression,
		resources,
		S3Bucket,
		S3Key,
		logger)
	if nil != err {
		return "", err
	}

	// Make sure the custom lambda that manages SNS notifications is provisioned.
	configuratorResName, err := ensureConfiguratorLambdaResource(SNSPrincipal,
		sourceArnExpression,
		[]string{},
		resources,
		S3Bucket,
		S3Key,
		logger)
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
			"TopicArn": sourceArnExpression,
			// Use the LambdaTarget value in the JS custom resoruce
			// handler to create the ID used to manage S3 notifications
			"LambdaTarget": targetLambdaFuncRef,
		},
		"DependsOn": []string{targetLambdaResourceName, configuratorResName},
	}
	// Save it
	subscriberResourceName := CloudFormationResourceName("SubscriberSNS",
		targetLambdaResourceName,
		perm.BasePermission.SourceAccount,
		fmt.Sprintf("%v", perm.BasePermission.SourceArn))
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
			"TopicArn": sourceArnExpression,
			// Use the LambdaTarget value in the JS custom resoruce
			// handler to create the ID used to manage S3 notifications
			"LambdaTarget": targetLambdaFuncRef,
		},
		"DependsOn": []string{subscriberResourceName},
	}
	// Save it
	unsubscriberResourceName := CloudFormationResourceName(fmt.Sprintf("UnsubscriberSNS%s", targetLambdaResourceName))
	resources[unsubscriberResourceName] = customResourceUnsubscriber

	return "", nil
}

func (perm SNSPermission) descriptionInfo() (string, string) {
	return perm.BasePermission.sourceArnString(), ""
}

//
// END - SNSPermission
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// MessageBodyStorageOptions - START

// MessageBodyStorageOptions define additional options for storing SES
// message body content.  By default, all rules associated with the owning
// SESPermission object will store message bodies if the MessageBodyStorage
// field is non-nil.  Message bodies are by default prefixed with
// `ServiceName/RuleName/`, which can be overriden by specifying a non-empty
// ObjectKeyPrefix value.  A rule can opt-out of message body storage
// with the DisableStorage field.  See
// http://docs.aws.amazon.com/ses/latest/DeveloperGuide/receiving-email-action-s3.html
// for additional field documentation.
// The message body is saved as MIME (https://tools.ietf.org/html/rfc2045)
type MessageBodyStorageOptions struct {
	ObjectKeyPrefix string
	KmsKeyArn       string
	TopicArn        string
	DisableStorage  bool
}

//
// END - MessageBodyStorageOptions
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// MessageBodyStorage - START

// MessageBodyStorage represents either a new S3 bucket or an existing S3 bucket
// to which SES message bodies should be stored.
// NOTE: New MessageBodyStorage create S3 buckets which will be orphaned after your
// service is deleted.
type MessageBodyStorage struct {
	logicalBucketName                  string
	bucketNameRef                      interface{}
	cloudFormationS3BucketResourceName string
}

// BucketArn returns an Arn value that can be used as an
// lambdaFn.RoleDefinition.Privileges `Resource` value.
func (storage *MessageBodyStorage) BucketArn() interface{} {
	return spartaCF.S3ArnForBucket(storage.bucketNameRef)
}

func (storage *MessageBodyStorage) export(serviceName string,
	targetLambdaFuncRef interface{},
	resources ArbitraryJSONObject,
	outputs ArbitraryJSONObject,
	S3Bucket string,
	S3Key string,
	logger *logrus.Logger) (string, error) {

	tags := map[string]string{
		"sparta:logicalBucketName": storage.logicalBucketName,
	}
	if "" != storage.cloudFormationS3BucketResourceName {
		resources[storage.cloudFormationS3BucketResourceName] = ArbitraryJSONObject{
			"Type":           "AWS::S3::Bucket",
			"DeletionPolicy": "Retain",
			"Properties": ArbitraryJSONObject{
				"Tags": spartaCF.MapToResourceTags(tags),
			},
		}
		logger.WithFields(logrus.Fields{
			"LogicalResourceName": storage.cloudFormationS3BucketResourceName,
		}).Info("Service will orphan S3 Bucket on deletion")

		// Add the created bucket to the outputs?
		spartaPrivate.InsertSpartaOutput(storage.cloudFormationS3BucketResourceName,
			storage.BucketArn(),
			"SES Message Body Bucket",
			outputs)
	}
	// Add the S3 Access policy
	s3SiteBucketPolicy := ArbitraryJSONObject{
		"Type": "AWS::S3::BucketPolicy",
		"Properties": ArbitraryJSONObject{
			"Bucket": storage.bucketNameRef,
			"PolicyDocument": ArbitraryJSONObject{
				"Version": "2012-10-17",
				"Statement": []ArbitraryJSONObject{
					{
						"Sid":    "PermitSESServiceToSaveEmailBody",
						"Effect": "Allow",
						"Principal": ArbitraryJSONObject{
							"Service": "ses.amazonaws.com",
						},
						"Action":   []string{"s3:PutObjectAcl", "s3:PutObject"},
						"Resource": spartaCF.S3AllKeysArnForBucket(storage.bucketNameRef),
						"Condition": ArbitraryJSONObject{
							"StringEquals": ArbitraryJSONObject{
								"aws:Referer": ArbitraryJSONObject{"Ref": "AWS::AccountId"},
							},
						},
					},
				},
			},
		},
	}

	s3BucketPolicyResourceName := CloudFormationResourceName("SESMessageBodyBucketPolicy",
		fmt.Sprintf("%#v", storage.bucketNameRef))
	resources[s3BucketPolicyResourceName] = s3SiteBucketPolicy

	// Return the name of the bucket policy s.t. the configurator resource
	// is properly sequenced.  The configurator will fail iff the Bucket Policies aren't
	// applied b/c the SES Rule Actions check PutObject access to S3 buckets
	return s3BucketPolicyResourceName, nil
}

//
// END - MessageBodyStorage
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// ReceiptRule - START

// ReceiptRule represents an SES ReceiptRule
// (http://docs.aws.amazon.com/ses/latest/DeveloperGuide/receiving-email-receipt-rules.html)
// value.  To store message bodies, provide a non-nil MessageBodyStorage value
// to the owning SESPermission object
type ReceiptRule struct {
	Name               string
	Disabled           bool
	Recipients         []string
	ScanDisabled       bool
	TLSPolicy          string
	TopicArn           string
	BodyStorageOptions MessageBodyStorageOptions
}

func (rule *ReceiptRule) lambdaTargetReceiptRule(serviceName string,
	functionArnRef interface{},
	messageBodyStorage *MessageBodyStorage) ArbitraryJSONObject {

	var actions []ArbitraryJSONObject
	// If there is a MessageBodyStorage reference, push that S3Action
	// to the head of the Actions list
	if nil != messageBodyStorage && !rule.BodyStorageOptions.DisableStorage {
		s3Action := ArbitraryJSONObject{
			"BucketName": messageBodyStorage.bucketNameRef,
		}
		if "" != rule.BodyStorageOptions.ObjectKeyPrefix {
			s3Action["ObjectKeyPrefix"] = rule.BodyStorageOptions.ObjectKeyPrefix
		} else {
			s3Action["ObjectKeyPrefix"] = fmt.Sprintf("%s/%s/", serviceName, rule.Name)
		}
		if "" != rule.BodyStorageOptions.KmsKeyArn {
			s3Action["KmsKeyArn"] = rule.BodyStorageOptions.KmsKeyArn
		}
		if "" != rule.BodyStorageOptions.TopicArn {
			s3Action["TopicArn"] = rule.BodyStorageOptions.TopicArn
		}
		actions = append(actions, ArbitraryJSONObject{
			"S3Action": s3Action,
		})
	}
	// Then create the "LambdaAction", which is always present.
	ruleAction := ArbitraryJSONObject{
		"FunctionArn":    functionArnRef,
		"InvocationType": "Event",
	}
	if "" != rule.TopicArn {
		ruleAction["TopicArn"] = rule.TopicArn
	}
	if "" == rule.TLSPolicy {
		rule.TLSPolicy = "Optional"
	}
	actions = append(actions, ArbitraryJSONObject{
		"LambdaAction": ruleAction,
	})

	// Return it.
	return ArbitraryJSONObject{
		"Name":        fmt.Sprintf("%s.%s", serviceName, rule.Name),
		"Enabled":     !rule.Disabled,
		"Recipients":  rule.Recipients,
		"ScanEnabled": !rule.ScanDisabled,
		"TlsPolicy":   rule.TLSPolicy,
		"Actions":     actions,
	}
}

//
// END - ReceiptRule
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// SESPermission - START

// SES doesn't use ARNs to scope access
var sesSourcePartArn = WildcardArn

// SESPermission struct that imples the SES verified domain should be
// updated (via createReceiptRule) to automatically request or push events
// to the parent lambda
// See http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html#intro-core-components-event-sources
// for more information.  See http://docs.aws.amazon.com/ses/latest/DeveloperGuide/receiving-email-concepts.html
// for setting up email receiving.
type SESPermission struct {
	BasePermission
	InvocationType     string /* RequestResponse, Event */
	ReceiptRules       []ReceiptRule
	MessageBodyStorage *MessageBodyStorage
}

// NewMessageBodyStorageResource provisions a new S3 bucket to store message body
// content.
func (perm *SESPermission) NewMessageBodyStorageResource(bucketLogicalName string) (*MessageBodyStorage, error) {
	if len(bucketLogicalName) <= 0 {
		return nil, errors.New("NewMessageBodyStorageResource requires a unique, non-empty `bucketLogicalName` parameter ")
	}
	store := &MessageBodyStorage{
		logicalBucketName: bucketLogicalName,
	}
	store.cloudFormationS3BucketResourceName = CloudFormationResourceName("SESMessageStoreBucket", bucketLogicalName)
	store.bucketNameRef = ArbitraryJSONObject{"Ref": store.cloudFormationS3BucketResourceName}
	return store, nil
}

// NewMessageBodyStorageReference uses a pre-existing S3 bucket for MessageBody storage.
// Sparta assumes that prexistingBucketName exists and will add an S3::BucketPolicy
// to enable SES PutObject access.
func (perm *SESPermission) NewMessageBodyStorageReference(prexistingBucketName string) (*MessageBodyStorage, error) {
	store := &MessageBodyStorage{}
	store.bucketNameRef = prexistingBucketName
	return store, nil
}

func (perm SESPermission) export(serviceName string,
	targetLambdaFuncRef interface{},
	resources ArbitraryJSONObject,
	outputs ArbitraryJSONObject,
	S3Bucket string,
	S3Key string,
	logger *logrus.Logger) (string, error) {

	// SES doesn't use ARNs, so make sure that's what the user supplied
	if perm.BasePermission.SourceArn != WildcardArn {
		return "", errors.New("SESPermissions only support `SourceArn=*` values")
	}
	sourceArnExpression := perm.BasePermission.sourceArnExpr(snsSourceArnParts)

	targetLambdaResourceName, err := perm.BasePermission.export(SESPrincipal,
		targetLambdaFuncRef,
		sourceArnExpression,
		resources,
		S3Bucket,
		S3Key,
		logger)
	if nil != err {
		return "", err
	}

	// MessageBody storage?
	var dependsOn []string
	if nil != perm.MessageBodyStorage {
		s3Policy, err := perm.MessageBodyStorage.export(serviceName,
			targetLambdaFuncRef,
			resources,
			outputs,
			S3Bucket,
			S3Key,
			logger)
		if nil != err {
			return "", err
		}
		if "" != s3Policy {
			dependsOn = append(dependsOn, s3Policy)
		}
	}

	// Make sure the custom lambda that manages SNS notifications is provisioned.
	configuratorResName, err := ensureConfiguratorLambdaResource(SESPrincipal,
		sourceArnExpression,
		dependsOn,
		resources,
		S3Bucket,
		S3Key,
		logger)

	if nil != err {
		return "", err
	}

	// Add a custom resource invocation for this configuration
	invocationType := perm.InvocationType
	if "" == invocationType {
		invocationType = "Event"
	}
	// If there aren't any, just forward everything
	receiptRules := perm.ReceiptRules
	if nil == perm.ReceiptRules {
		receiptRules = []ReceiptRule{ReceiptRule{
			Name:         "Default",
			Disabled:     false,
			ScanDisabled: false,
			Recipients:   []string{},
			TLSPolicy:    "Optional",
		}}
	}

	var xformedRules []ArbitraryJSONObject
	for _, eachReceiptRule := range receiptRules {
		xformedRules = append(xformedRules,
			eachReceiptRule.lambdaTargetReceiptRule(
				serviceName,
				targetLambdaFuncRef,
				perm.MessageBodyStorage))
	}

	customResourceSubscriber := ArbitraryJSONObject{
		"Type":    "AWS::CloudFormation::CustomResource",
		"Version": "1.0",
		"Properties": ArbitraryJSONObject{
			"ServiceToken": ArbitraryJSONObject{
				"Fn::GetAtt": []string{configuratorResName, "Arn"},
			},
			"Rules": xformedRules,
		},
		"DependsOn": []string{targetLambdaResourceName, configuratorResName},
	}
	// Save it
	subscriberResourceName := CloudFormationResourceName("SubscriberSES",
		targetLambdaResourceName,
		perm.BasePermission.SourceAccount,
		fmt.Sprintf("%v", perm.BasePermission.SourceArn))
	resources[subscriberResourceName] = customResourceSubscriber

	return "", nil
}

func (perm SESPermission) descriptionInfo() (string, string) {
	return perm.BasePermission.sourceArnString(), ""
}

//
// END - SESPermission
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// START - IAMRolePrivilege
//

// IAMRolePrivilege struct stores data necessary to create an IAM Policy Document
// as part of the inline IAM::Role resource definition.  See
// http://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies.html
// for more information
type IAMRolePrivilege struct {
	// What actions you will allow.
	// Each AWS service has its own set of actions.
	// For example, you might allow a user to use the Amazon S3 ListBucket action,
	// which returns information about the items in a bucket.
	// Any actions that you don't explicitly allow are denied.
	Actions []string
	// Which resources you allow the action on. For example, what specific Amazon
	// S3 buckets will you allow the user to perform the ListBucket action on?
	// Users cannot access any resources that you have not explicitly granted
	// permissions to.
	Resource interface{}
}

// IAMRoleDefinition stores a slice of IAMRolePrivilege values
// to "Allow" for the given IAM::Role.
// Note that the CommonIAMStatements will be automatically included and do
// not need to be multiply specified.
type IAMRoleDefinition struct {
	// Slice of IAMRolePrivilege entries
	Privileges []IAMRolePrivilege
	// Cached logical resource name
	cachedLogicalName string
}

// Returns an IAM::Role policy entry for this definition
func (roleDefinition *IAMRoleDefinition) rolePolicy(eventSourceMappings []*lambda.CreateEventSourceMappingInput, logger *logrus.Logger) ArbitraryJSONObject {
	statements := CommonIAMStatements["core"]
	for _, eachPrivilege := range roleDefinition.Privileges {
		statements = append(statements, ArbitraryJSONObject{
			"Effect":   "Allow",
			"Action":   eachPrivilege.Actions,
			"Resource": eachPrivilege.Resource,
		})
	}

	// // http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html
	for _, eachEventSourceMapping := range eventSourceMappings {
		arnParts := strings.Split(*eachEventSourceMapping.EventSourceArn, ":")
		// 3rd slot is service scope
		if len(arnParts) >= 2 {
			awsService := arnParts[2]
			logger.Debug("Looking up common IAM privileges for EventSource: ", awsService)
			serviceStatements, exists := CommonIAMStatements[awsService]
			if exists {
				statements = append(statements, serviceStatements...)
				statements[len(statements)-1]["Resource"] = *eachEventSourceMapping.EventSourceArn
			}
		}
	}
	iamPolicy := ArbitraryJSONObject{"Type": "AWS::IAM::Role",
		"Properties": ArbitraryJSONObject{
			"AssumeRolePolicyDocument": AssumePolicyDocument,
			"Policies": []ArbitraryJSONObject{
				{
					"PolicyName": CloudFormationResourceName("LambdaPolicy"),
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

// Returns the stable logical name for this IAMRoleDefinition, which must be unique
// if the privileges are empty.
func (roleDefinition *IAMRoleDefinition) logicalName() string {
	if "" == roleDefinition.cachedLogicalName {
		// TODO: Name isn't stable across executions, which is a performance penalty across updates if the Permissions are unchanged.
		roleDefinition.cachedLogicalName = CloudFormationResourceName("IAMRole")
	}
	return roleDefinition.cachedLogicalName
}

//
// END - IAMRolePrivilege
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// Interface type to support multiple lambda runtimes

////////////////////////////////////////////////////////////////////////////////
// START - LambdaAWSInfo
//

// LambdaAWSInfo stores all data necessary to provision a golang-based AWS Lambda function.
type LambdaAWSInfo struct {
	// internal function name, determined by reflection
	lambdaFnName string
	// pointer to lambda function
	lambdaFn LambdaFunction
	// Role name (NOT ARN) to use during AWS Lambda Execution.  See
	// the FunctionConfiguration (http://docs.aws.amazon.com/lambda/latest/dg/API_FunctionConfiguration.html)
	// docs for more info.
	// Note that either `RoleName` or `RoleDefinition` must be supplied
	RoleName string
	// IAM Role Definition if the stack should implicitly create an IAM role for
	// lambda execution. Note that either `RoleName` or `RoleDefinition` must be supplied
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
	// Template decorator. If defined, the decorator will be called to insert additional
	// resources on behalf of this lambda function
	Decorator TemplateDecorator
}

// Returns a JavaScript compatible function name for the golang function name.  This
// value will be used as the URL path component for the HTTP proxying layer.
func (info *LambdaAWSInfo) jsHandlerName() string {
	return sanitizedName(info.lambdaFnName)
}

// Marshal this object into 1 or more CloudFormation resource definitions that are accumulated
// in the resources map
func (info *LambdaAWSInfo) export(serviceName string,
	S3Bucket string,
	S3Key string,
	roleNameMap map[string]interface{},
	resources ArbitraryJSONObject,
	outputs ArbitraryJSONObject,
	logger *logrus.Logger) error {

	// If we have RoleName, then get the ARN, otherwise get the Ref
	var dependsOn []string

	iamRoleArnName := info.RoleName
	// If there is no user supplied role, that means that the associated
	// IAMRoleDefinition name has been created and this resource needs to
	// depend on that existing.
	if iamRoleArnName == "" {
		iamRoleArnName = info.RoleDefinition.logicalName()
		dependsOn = append(dependsOn, iamRoleArnName)
	}
	lambdaDescription := info.Options.Description
	if "" == lambdaDescription {
		lambdaDescription = fmt.Sprintf("%s: %s", serviceName, info.lambdaFnName)
	}
	// Create the primary resource
	primaryResource := ArbitraryJSONObject{
		"Type": "AWS::Lambda::Function",
		"Properties": ArbitraryJSONObject{
			"Code": ArbitraryJSONObject{
				"S3Bucket": S3Bucket,
				"S3Key":    S3Key,
			},
			"Description": lambdaDescription,
			"Handler":     fmt.Sprintf("index.%s", info.jsHandlerName()),
			"MemorySize":  info.Options.MemorySize,
			"Role":        roleNameMap[iamRoleArnName],
			"Runtime":     "nodejs",
			"Timeout":     info.Options.Timeout,
		},
		"DependsOn": dependsOn,
		"Metadata": ArbitraryJSONObject{
			"golangFunc": info.lambdaFnName,
		},
	}

	// Get the resource name we're going to use s.t. we can tie it to the rest of the
	// lambda definition
	resourceName := info.logicalName()
	resources[resourceName] = primaryResource

	// Create the lambda Ref in case we need a permission or event mapping
	functionAttr := ArbitraryJSONObject{
		"Fn::GetAtt": []string{resourceName, "Arn"},
	}

	// Permissions
	for _, eachPermission := range info.Permissions {
		_, err := eachPermission.export(serviceName,
			functionAttr,
			resources,
			outputs,
			S3Bucket,
			S3Key,
			logger)
		if nil != err {
			return err
		}
	}

	// Event Source Mappings
	// TODO: verify that the event source ARN actually exists.
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
			"DependsOn":  dependsOn,
			"Properties": properties,
		}
		hash := sha1.New()
		hash.Write([]byte(*eachEventSourceMapping.EventSourceArn))
		binary.Write(hash, binary.LittleEndian, *eachEventSourceMapping.BatchSize)
		hash.Write([]byte(*eachEventSourceMapping.StartingPosition))
		resourceName := fmt.Sprintf("LambdaES%s", hex.EncodeToString(hash.Sum(nil)))
		resources[resourceName] = primaryEventSourceMapping
	}

	// Decorator
	if nil != info.Decorator {
		logger.Debug("Decorator found for Lambda: ", info.lambdaFnName)
		lambdaResources := make(ArbitraryJSONObject, 0)
		lambdaOutputs := make(ArbitraryJSONObject, 0)
		err := info.Decorator(resourceName, primaryResource, lambdaResources, lambdaOutputs, logger)
		if nil != err {
			return err
		}
		// Append the custom resources
		for eachKey, eachLambdaResource := range lambdaResources {
			_, exists := resources[eachKey]
			if exists {
				errorMsg := fmt.Sprintf("Duplicate CloudFormation resource name (%s) defined by Lambda: %s",
					eachKey,
					info.lambdaFnName)
				return errors.New(errorMsg)
			}
			resources[eachKey] = eachLambdaResource
		}
		// Append the custom outputs
		for eachKey, eachLambdaOutput := range lambdaOutputs {
			_, exists := outputs[eachKey]
			if exists {
				errorMsg := fmt.Sprintf("Duplicate CloudFormation output key name (%s) defined by Lambda: %s",
					eachKey,
					info.lambdaFnName)
				return errors.New(errorMsg)
			}
			outputs[eachKey] = eachLambdaOutput
		}
	}

	return nil
}

// Returns the stable logical name for this IAMRoleDefinition
func (info *LambdaAWSInfo) logicalName() string {
	// Per http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/resources-section-structure.html,
	// we can only use alphanumeric, so we'll take the sanitized name and
	// remove all underscores
	resourceName := strings.Replace(sanitizedName(info.lambdaFnName), "_", "", -1)
	prefix := fmt.Sprintf("%sLambda", resourceName)
	return CloudFormationResourceName(prefix, info.lambdaFnName)
}

//
// END - LambdaAWSInfo
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// Private
//
// Sanitize the provided input by replacing illegal characters with underscores
func sanitizedName(input string) string {
	return reSanitize.ReplaceAllString(input, "_")
}

// Returns an AWS Session (https://github.com/aws/aws-sdk-go/wiki/Getting-Started-Configuration)
// object that attaches a debug level handler to all AWS requests from services
// sharing the session value.
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

// CloudFormationResourceName returns a name suitable as a logical
// CloudFormation resource value.  See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/resources-section-structure.html
// for more information.  The `prefix` value should provide a hint as to the
// resource type (eg, `SNSConfigurator`, `ImageTranscoder`).  Note that the returned
// name is not content-addressable.
func CloudFormationResourceName(prefix string, parts ...string) string {
	hash := sha1.New()
	hash.Write([]byte(prefix))
	if len(parts) <= 0 {
		randValue := rand.Int63()
		hash.Write([]byte(strconv.FormatInt(randValue, 10)))
	} else {
		for _, eachPart := range parts {
			hash.Write([]byte(eachPart))
		}
	}
	return fmt.Sprintf("%s%s", prefix, hex.EncodeToString(hash.Sum(nil)))
}

////////////////////////////////////////////////////////////////////////////////
// Public
////////////////////////////////////////////////////////////////////////////////

// NewLambda returns a LambdaAWSInfo value that can be provisioned via CloudFormation. The
// roleNameOrIAMRoleDefinition must either be a `string` or `IAMRoleDefinition`
// type
func NewLambda(roleNameOrIAMRoleDefinition interface{}, fn LambdaFunction, lambdaOptions *LambdaFunctionOptions) *LambdaAWSInfo {
	if nil == lambdaOptions {
		lambdaOptions = &LambdaFunctionOptions{"", 128, 3}
	}
	lambdaPtr := runtime.FuncForPC(reflect.ValueOf(fn).Pointer())
	lambda := &LambdaAWSInfo{
		lambdaFnName:        lambdaPtr.Name(),
		lambdaFn:            fn,
		Options:             lambdaOptions,
		Permissions:         make([]LambdaPermissionExporter, 0),
		EventSourceMappings: make([]*lambda.CreateEventSourceMappingInput, 0),
	}

	switch v := roleNameOrIAMRoleDefinition.(type) {
	case string:
		lambda.RoleName = roleNameOrIAMRoleDefinition.(string)
	case IAMRoleDefinition:
		definition := roleNameOrIAMRoleDefinition.(IAMRoleDefinition)
		lambda.RoleDefinition = &definition
	default:
		panic(fmt.Sprintf("Unsupported IAM Role type: %s", v))
	}

	// Defaults
	if lambda.Options.MemorySize <= 0 {
		lambda.Options.MemorySize = 128
	}
	if lambda.Options.Timeout <= 0 {
		lambda.Options.Timeout = 3
	}
	return lambda
}

// NewLogger returns a new logrus.Logger instance. It is the caller's responsibility
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

// Main defines the primary handler for transforming an application into a Sparta package.  The
// serviceName is used to uniquely identify your service within a region and will
// be used for subsequent updates.  For provisioning, ensure that you've
// properly configured AWS credentials for the golang SDK.
// See http://docs.aws.amazon.com/sdk-for-go/api/aws/defaults.html#DefaultChainCredentials-constant
// for more information.
func Main(serviceName string, serviceDescription string, lambdaAWSInfos []*LambdaAWSInfo, api *API, site *S3Site) error {

	// We need to be able to provision an IAM role that has capabilities to
	// manage the other sources.  That'll give us the role arn to use in the custom
	// resource execution.
	options := struct {
		Noop     bool          `goptions:"-n, --noop, description='Dry-run behavior only (do not provision stack)'"`
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
			Port int `goptions:"-p,--port, description='Alternative port for HTTP binding (default=9999)'"`
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
	// Set the formatter before outputting the info s.t. it's properly
	// parsed by CloudWatch Logs
	if "execute" == options.Verb {
		logger.Formatter = new(logrus.JSONFormatter)
	} else {
		logger.Formatter = new(logrus.TextFormatter)
	}
	logger.WithFields(logrus.Fields{
		"Option":  options.Verb,
		"Version": SpartaVersion,
		"TS":      (time.Now().UTC().Format(time.RFC3339)),
	}).Info("Welcome to Sparta")

	switch options.Verb {
	case "provision":
		logger.Info(strings.Repeat("-", 80))
		err = Provision(options.Noop, serviceName, serviceDescription, lambdaAWSInfos, api, site, options.Provision.S3Bucket, nil, logger)
	case "execute":
		err = Execute(lambdaAWSInfos, options.Execute.Port, options.Execute.SignalParentPID, logger)
	case "delete":
		logger.Info(strings.Repeat("-", 80))
		err = Delete(serviceName, logger)
	case "explore":
		logger.Info(strings.Repeat("-", 80))
		err = Explore(lambdaAWSInfos, options.Explore.Port, logger)
	case "describe":
		logger.Info(strings.Repeat("-", 80))
		fileWriter, err := os.Create(options.Describe.OutputFile)
		if err != nil {
			return fmt.Errorf("Failed to open %s output. Error: %s", options.Describe.OutputFile, err)
		}
		defer fileWriter.Close()
		err = Describe(serviceName, serviceDescription, lambdaAWSInfos, api, site, fileWriter, logger)
	default:
		goptions.PrintHelp()
		err = fmt.Errorf("Unsupported subcommand: %s", string(options.Verb))
	}
	if nil != err {
		logger.Error(err)
	}
	return err
}
