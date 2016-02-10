package sparta

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	gocf "github.com/crewjam/go-cloudformation"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/service/s3"
)

/*
Notes to future self...

TODO - Simplify this as part of: https://trello.com/c/aOULlJcz/14-port-nodejs-customresources-to-go

Adding a new permission type?
  1. Add the principal name value to sparta.go constants
  2. Define the new struct and satisfy LambdaPermissionExporter
  3. Update provision_utils.go's `PushSourceConfigurationActions` map with the new principal's permissions
  4. Update `PROXIED_MODULES` in resources/index.js to include the first principal component name( eg, 'events')
  5. Update `customResourceScripts` in provision.go to ensure the embedded JS file is included in the deployed archive.
  6. Implement the custom type defined in 2
  7. Implement the service configuration logic referred to in 4.
*/

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
		lambdaLogicalCFResourceName string,
		template *gocf.Template,
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

func (perm *BasePermission) sourceArnExpr(joinParts ...gocf.Stringable) *gocf.StringExpr {
	var parts []gocf.Stringable
	if nil != joinParts {
		parts = append(parts, joinParts...)
	}
	switch perm.SourceArn.(type) {
	case string:
		// Don't be smart if the Arn value is a user supplied literal
		parts = []gocf.Stringable{gocf.String(perm.SourceArn.(string))}
	case *gocf.StringExpr:
		parts = append(parts, perm.SourceArn.(*gocf.StringExpr))
	case gocf.RefFunc:
		parts = append(parts, perm.SourceArn.(gocf.RefFunc).String())
	default:
		panic(fmt.Sprintf("Unsupported SourceArn value type: %+v", perm.SourceArn))
	}
	return gocf.Join("", parts...)
}

func (perm *BasePermission) describeInfoArn() string {
	switch perm.SourceArn.(type) {
	case string:
		return perm.SourceArn.(string)
	case *gocf.StringExpr,
		gocf.RefFunc:
		data, _ := json.Marshal(perm.SourceArn)
		return string(data)
	default:
		panic(fmt.Sprintf("Unsupported SourceArn value type: %+v", perm.SourceArn))
	}
}

func (perm BasePermission) export(principal string,
	arnPrefixParts []gocf.Stringable,
	lambdaLogicalCFResourceName string,
	template *gocf.Template,
	S3Bucket string,
	S3Key string,
	logger *logrus.Logger) (string, error) {

	lambdaPermission := gocf.LambdaPermission{
		Action:       gocf.String("lambda:InvokeFunction"),
		FunctionName: gocf.GetAtt(lambdaLogicalCFResourceName, "Arn"),
		Principal:    gocf.String(principal),
	}
	// If the Arn isn't the wildcard value, then include it.
	if nil != perm.SourceArn {
		switch perm.SourceArn.(type) {
		case string:
			// Don't be smart if the Arn value is a user supplied literal
			if "*" != perm.SourceArn.(string) {
				lambdaPermission.SourceArn = gocf.String(perm.SourceArn.(string))
			}
		default:
			lambdaPermission.SourceArn = perm.sourceArnExpr(arnPrefixParts...)
		}
	}

	if "" != perm.SourceAccount {
		lambdaPermission.SourceAccount = gocf.String(perm.SourceAccount)
	}

	hash := sha1.New()
	hash.Write([]byte(fmt.Sprintf("%v", lambdaPermission)))
	resourceName := fmt.Sprintf("LambdaPerm%s", hex.EncodeToString(hash.Sum(nil)))
	template.AddResource(resourceName, lambdaPermission)
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
var lambdaSourceArnParts = []gocf.Stringable{
	gocf.String("arn:aws:lambda:"),
	gocf.Ref("AWS::Region"),
	gocf.String(":function:"),
}

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
	lambdaLogicalCFResourceName string,
	template *gocf.Template,
	S3Bucket string,
	S3Key string,
	logger *logrus.Logger) (string, error) {

	return perm.BasePermission.export(perm.Principal,
		lambdaSourceArnParts,
		lambdaLogicalCFResourceName,
		template,
		S3Bucket,
		S3Key,
		logger)
}

func (perm LambdaPermission) descriptionInfo() (string, string) {
	return "Source", perm.describeInfoArn()
}

//
// END - LambdaPermission
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// START - S3Permission
//
var s3SourceArnParts = []gocf.Stringable{
	gocf.String("arn:aws:s3:::"),
}

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
	lambdaLogicalCFResourceName string,
	template *gocf.Template,
	S3Bucket string,
	S3Key string,
	logger *logrus.Logger) (string, error) {

	targetLambdaResourceName, err := perm.BasePermission.export(S3Principal,
		s3SourceArnParts,
		lambdaLogicalCFResourceName,
		template,
		S3Bucket,
		S3Key,
		logger)
	if nil != err {
		return "", err
	}

	// Make sure the custom lambda that manages s3 notifications is provisioned.
	sourceArnExpression := perm.BasePermission.sourceArnExpr(s3SourceArnParts...)
	configuratorResName, err := ensureConfiguratorLambdaResource(S3Principal,
		sourceArnExpression,
		[]string{},
		template,
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
	newResource, err := newCloudFormationResource("Custom::SpartaS3Permission", logger)
	if nil != err {
		return "", err
	}
	customResource := newResource.(*cloudFormationS3PermissionResource)
	customResource.ServiceToken = gocf.GetAtt(configuratorResName, "Arn")
	customResource.Permission = permissionData
	customResource.LambdaTarget = gocf.GetAtt(lambdaLogicalCFResourceName, "Arn")
	customResource.BucketArn = sourceArnExpression

	// Name?
	resourceInvokerName := CloudFormationResourceName("ConfigS3",
		targetLambdaResourceName,
		perm.BasePermission.SourceAccount,
		fmt.Sprintf("%v", sourceArnExpression))
	// Add it
	cfResource := template.AddResource(resourceInvokerName, customResource)
	cfResource.DependsOn = append(cfResource.DependsOn,
		targetLambdaResourceName,
		configuratorResName)
	return "", nil
}

func (perm S3Permission) descriptionInfo() (string, string) {
	s3Events := ""
	for _, eachEvent := range perm.Events {
		s3Events = fmt.Sprintf("%s\n%s", eachEvent, s3Events)
	}
	return perm.describeInfoArn(), s3Events
}

//
// END - S3Permission
///////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// SNSPermission - START
var snsSourceArnParts = []gocf.Stringable{}

// SNSPermission struct that imples the S3 BasePermission.SourceArn should be
// updated (via PutBucketNotificationConfiguration) to automatically push
// events to the parent Lambda.
// See http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html#intro-core-components-event-sources
// for more information.
type SNSPermission struct {
	BasePermission
}

func (perm SNSPermission) export(serviceName string,
	lambdaLogicalCFResourceName string,
	template *gocf.Template,
	S3Bucket string,
	S3Key string,
	logger *logrus.Logger) (string, error) {
	sourceArnExpression := perm.BasePermission.sourceArnExpr(snsSourceArnParts...)

	targetLambdaResourceName, err := perm.BasePermission.export(SNSPrincipal,
		snsSourceArnParts,
		lambdaLogicalCFResourceName,
		template,
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
		template,
		S3Bucket,
		S3Key,
		logger)
	if nil != err {
		return "", err
	}

	// Add a custom resource invocation for this configuration
	//////////////////////////////////////////////////////////////////////////////
	// And the custom resource forwarder

	newResource, err := newCloudFormationResource("Custom::SpartaSNSPermission", logger)
	if nil != err {
		return "", err
	}
	customResource := newResource.(*cloudFormationSNSPermissionResource)
	customResource.ServiceToken = gocf.GetAtt(configuratorResName, "Arn")
	customResource.Mode = "Subscribe"
	customResource.TopicArn = sourceArnExpression
	customResource.LambdaTarget = gocf.GetAtt(lambdaLogicalCFResourceName, "Arn")
	subscriberResourceName := CloudFormationResourceName("SubscriberSNS",
		targetLambdaResourceName,
		perm.BasePermission.SourceAccount,
		fmt.Sprintf("%v", perm.BasePermission.SourceArn))
	cfResource := template.AddResource(subscriberResourceName, customResource)
	cfResource.DependsOn = append(cfResource.DependsOn, targetLambdaResourceName, configuratorResName)

	//////////////////////////////////////////////////////////////////////////////
	// And the custom resource unsubscriber
	newResource, err = newCloudFormationResource("Custom::SpartaSNSPermission", logger)
	if nil != err {
		return "", err
	}
	customResource = newResource.(*cloudFormationSNSPermissionResource)
	customResource.ServiceToken = gocf.GetAtt(configuratorResName, "Arn")
	customResource.Mode = "Unsubscribe"
	customResource.TopicArn = sourceArnExpression
	customResource.LambdaTarget = gocf.GetAtt(lambdaLogicalCFResourceName, "Arn")
	unsubscriberResourceName := CloudFormationResourceName("UnsubscriberSNS",
		targetLambdaResourceName)
	template.AddResource(unsubscriberResourceName, customResource)

	return "", nil
}

func (perm SNSPermission) descriptionInfo() (string, string) {
	return perm.BasePermission.describeInfoArn(), ""
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
	bucketNameExpr                     *gocf.StringExpr
	cloudFormationS3BucketResourceName string
}

// BucketArn returns an Arn value that can be used as an
// lambdaFn.RoleDefinition.Privileges `Resource` value.
func (storage *MessageBodyStorage) BucketArn() *gocf.StringExpr {
	return gocf.Join("",
		gocf.String("arn:aws:s3:::"),
		storage.bucketNameExpr)
}

// BucketArnAllKeys returns an Arn value that can be used
// lambdaFn.RoleDefinition.Privileges `Resource` value.  It includes
// the trailing `/*` wildcard to support item acccess
func (storage *MessageBodyStorage) BucketArnAllKeys() *gocf.StringExpr {
	return gocf.Join("",
		gocf.String("arn:aws:s3:::"),
		storage.bucketNameExpr,
		gocf.String("/*"))
}

func (storage *MessageBodyStorage) export(serviceName string,
	lambdaLogicalCFResourceName string,
	template *gocf.Template,
	S3Bucket string,
	S3Key string,
	logger *logrus.Logger) (string, error) {

	if "" != storage.cloudFormationS3BucketResourceName {
		s3Bucket := &gocf.S3Bucket{
			Tags: []gocf.ResourceTag{
				gocf.ResourceTag{
					Key:   gocf.String("sparta:logicalBucketName"),
					Value: gocf.String(storage.logicalBucketName),
				},
			},
		}
		cfResource := template.AddResource(storage.cloudFormationS3BucketResourceName, s3Bucket)
		cfResource.DeletionPolicy = "Retain"

		lambdaResource, _ := template.Resources[lambdaLogicalCFResourceName]
		if nil != lambdaResource {
			safeAppendDependency(lambdaResource, storage.cloudFormationS3BucketResourceName)
		}

		logger.WithFields(logrus.Fields{
			"LogicalResourceName": storage.cloudFormationS3BucketResourceName,
		}).Info("Service will orphan S3 Bucket on deletion")

		// Save the output
		template.Outputs[storage.cloudFormationS3BucketResourceName] = &gocf.Output{
			Description: "SES Message Body Bucket",
			Value:       gocf.Ref(storage.cloudFormationS3BucketResourceName),
		}
	}
	// Add the S3 Access policy
	s3BodyStoragePolicy := &gocf.S3BucketPolicy{
		Bucket: storage.bucketNameExpr,
		PolicyDocument: ArbitraryJSONObject{
			"Version": "2012-10-17",
			"Statement": []ArbitraryJSONObject{
				{
					"Sid":    "PermitSESServiceToSaveEmailBody",
					"Effect": "Allow",
					"Principal": ArbitraryJSONObject{
						"Service": "ses.amazonaws.com",
					},
					"Action": []string{"s3:PutObjectAcl", "s3:PutObject"},
					"Resource": gocf.Join("",
						gocf.String("arn:aws:s3:::"),
						storage.bucketNameExpr,
						gocf.String("/*")),
					"Condition": ArbitraryJSONObject{
						"StringEquals": ArbitraryJSONObject{
							"aws:Referer": gocf.Ref("AWS::AccountId"),
						},
					},
				},
			},
		},
	}

	s3BucketPolicyResourceName := CloudFormationResourceName("SESMessageBodyBucketPolicy",
		fmt.Sprintf("%#v", storage.bucketNameExpr))
	template.AddResource(s3BucketPolicyResourceName, s3BodyStoragePolicy)

	// Return the name of the bucket policy s.t. the configurator resource
	// is properly sequenced.  The configurator will fail iff the Bucket Policies aren't
	// applied b/c the SES Rule Actions check PutObject access to S3 buckets
	return s3BucketPolicyResourceName, nil
}

// Return a function that

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
			"BucketName": messageBodyStorage.bucketNameExpr,
		}
		if "" != rule.BodyStorageOptions.ObjectKeyPrefix {
			s3Action["ObjectKeyPrefix"] = rule.BodyStorageOptions.ObjectKeyPrefix
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
var sesSourcePartArn = []gocf.Stringable{wildcardArn}

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
	store.bucketNameExpr = gocf.Ref(store.cloudFormationS3BucketResourceName).String()
	return store, nil
}

// NewMessageBodyStorageReference uses a pre-existing S3 bucket for MessageBody storage.
// Sparta assumes that prexistingBucketName exists and will add an S3::BucketPolicy
// to enable SES PutObject access.
func (perm *SESPermission) NewMessageBodyStorageReference(prexistingBucketName string) (*MessageBodyStorage, error) {
	store := &MessageBodyStorage{}
	store.bucketNameExpr = gocf.String(prexistingBucketName)
	return store, nil
}

func (perm SESPermission) export(serviceName string,
	lambdaLogicalCFResourceName string,
	template *gocf.Template,
	S3Bucket string,
	S3Key string,
	logger *logrus.Logger) (string, error) {

	sourceArnExpression := perm.BasePermission.sourceArnExpr(snsSourceArnParts...)

	targetLambdaResourceName, err := perm.BasePermission.export(SESPrincipal,
		sesSourcePartArn,
		lambdaLogicalCFResourceName,
		template,
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
			lambdaLogicalCFResourceName,
			template,
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
		template,
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
				gocf.GetAtt(lambdaLogicalCFResourceName, "Arn"),
				perm.MessageBodyStorage))
	}

	newResource, err := newCloudFormationResource("Custom::SpartaSESPermission", logger)
	if nil != err {
		return "", err
	}
	customResource := newResource.(*cloudFormationSESPermissionResource)
	customResource.ServiceToken = gocf.GetAtt(configuratorResName, "Arn")
	customResource.Rules = xformedRules

	subscriberResourceName := CloudFormationResourceName("SubscriberSES",
		targetLambdaResourceName,
		perm.BasePermission.SourceAccount,
		fmt.Sprintf("%v", perm.BasePermission.SourceArn))
	cfResource := template.AddResource(subscriberResourceName, customResource)
	cfResource.DependsOn = append(cfResource.DependsOn, targetLambdaResourceName, configuratorResName)
	return "", nil
}

func (perm SESPermission) descriptionInfo() (string, string) {
	// SES doesn't use ARNs, but "*" breaks mermaids parser, so
	// use entity code per: http://knsv.github.io/mermaid/#special-characters-that-break-syntax
	return "SimpleEmailService", "All verified domain(s) email"
}

//
// END - SESPermission
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// START - CloudWatchEventsRuleTarget
//

// CloudWatchEventsRuleTarget specifies additional input and JSON selection
// paths to apply prior to forwarding the event to a lambda function
type CloudWatchEventsRuleTarget struct {
	Input     string
	InputPath string
}

//
// END - CloudWatchEventsRuleTarget
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// START - CloudWatchEventsRule
//

// CloudWatchEventsRule defines parameters for invoking a lambda function
// in response to specific CloudWatchEvents or cron triggers
type CloudWatchEventsRule struct {
	Description string
	// ArbitraryJSONObject filter for events as documented at
	// http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/CloudWatchEventsandEventPatterns.html
	// Rules matches should use the JSON representation (NOT the string form).  Sparta will serialize
	// the map[string]interface{} to a string form during CloudFormation Template
	// marshalling.
	EventPattern map[string]interface{} `json:"EventPattern,omitempty"`
	// Schedule pattern per http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/ScheduledEvents.html
	ScheduleExpression string
	RuleTarget         *CloudWatchEventsRuleTarget `json:"RuleTarget,omitempty"`
}

// MarshalJSON customizes the JSON representation used when serializing to the
// CloudFormation template representation.
func (rule CloudWatchEventsRule) MarshalJSON() ([]byte, error) {
	ruleJSON := map[string]interface{}{}

	if "" != rule.Description {
		ruleJSON["Description"] = rule.Description
	}
	if nil != rule.EventPattern {
		eventPatternString, err := json.Marshal(rule.EventPattern)
		if nil != err {
			return nil, err
		}
		ruleJSON["EventPattern"] = string(eventPatternString)
	}
	if "" != rule.ScheduleExpression {
		ruleJSON["ScheduleExpression"] = rule.ScheduleExpression
	}
	if nil != rule.RuleTarget {
		ruleJSON["RuleTarget"] = rule.RuleTarget
	}
	return json.Marshal(ruleJSON)
}

//
// END - CloudWatchEventsRule
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// START - CloudWatchEventsPermission
//
var cloudformationEventsSourceArnParts = []gocf.Stringable{}

// CloudWatchEventsPermission struct that imples the S3 BasePermission.SourceArn should be
// updated (via PutBucketNotificationConfiguration) to automatically push
// events to the owning Lambda.
// See http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html#intro-core-components-event-sources
// for more information.
type CloudWatchEventsPermission struct {
	BasePermission
	// Map of rule names to events that trigger the lambda function
	Rules map[string]CloudWatchEventsRule
}

func (perm CloudWatchEventsPermission) export(serviceName string,
	lambdaLogicalCFResourceName string,
	template *gocf.Template,
	S3Bucket string,
	S3Key string,
	logger *logrus.Logger) (string, error) {

	// Tell the user we're ignoring any Arns provided, since it doesn't make sense for
	// this.
	if nil != perm.BasePermission.SourceArn &&
		perm.BasePermission.sourceArnExpr(cloudformationEventsSourceArnParts...).String() != wildcardArn.String() {
		logger.WithFields(logrus.Fields{
			"Arn": perm.BasePermission.sourceArnExpr(cloudformationEventsSourceArnParts...),
		}).Warn("CloudWatchEvents do not support literal ARN values")
	}

	arnPermissionForRuleName := func(ruleName string) *gocf.StringExpr {
		return gocf.Join("",
			gocf.String("arn:aws:events:"),
			gocf.Ref("AWS::Region"),
			gocf.String(":"),
			gocf.Ref("AWS::AccountId"),
			gocf.String(":rule/"),
			gocf.String(ruleName))
	}

	// First thing we need to do is uniqueify the rule names s.t. we prevent
	// collisions with other stacks.
	globallyUniqueRules := make(map[string]CloudWatchEventsRule, len(perm.Rules))
	for eachRuleName, eachDefinition := range perm.Rules {
		uniqueRuleName := CloudFormationResourceName(eachRuleName, lambdaLogicalCFResourceName, serviceName)
		// Trim it...
		if len(eachDefinition.Description) <= 0 {
			eachDefinition.Description = fmt.Sprintf("%s CloudWatch Events rule for service: %s", eachRuleName, serviceName)
		}
		globallyUniqueRules[uniqueRuleName] = eachDefinition
	}
	// Integrity test - there should only be 1 element since we're only ever configuring
	// the same AWS principal service.  If we end up with multiple configuration resource names
	// it means that the stable resource name logic is broken
	configurationResourceNames := make(map[string]int, 0)
	var dependsOn []string
	for eachRuleName := range globallyUniqueRules {
		basePerm := BasePermission{
			SourceArn: arnPermissionForRuleName(eachRuleName),
		}
		dependOn, err := basePerm.export(CloudWatchEventsPrincipal,
			cloudformationEventsSourceArnParts,
			lambdaLogicalCFResourceName,
			template,
			S3Bucket,
			S3Key,
			logger)
		if nil != err {
			return "", err
		}
		dependsOn = append(dependsOn, dependOn)

		// Ensure the configurator for this ARNs
		sourceArnExpression := basePerm.sourceArnExpr(cloudformationEventsSourceArnParts...)

		// Make sure the custom lambda that manages CloudWatch Events is provisioned.
		configuratorResName, err := ensureConfiguratorLambdaResource(CloudWatchEventsPrincipal,
			sourceArnExpression,
			[]string{},
			template,
			S3Bucket,
			S3Key,
			logger)
		if nil != err {
			return "", err
		}
		configurationResourceNames[configuratorResName] = 1
	}
	// Although we ensured multiple configuration resources, they were all for the
	// same AWS principal.  We're only supposed to get a single name back.
	if len(configurationResourceNames) > 1 {
		return "", fmt.Errorf("Multiple configuration resources detected: %#v", configurationResourceNames)
	} else if len(configurationResourceNames) == 0 {
		return "", fmt.Errorf("CloudWatchEvent configuration provider failed")
	}

	// Insert the invocation
	for eachConfigResource := range configurationResourceNames {
		//////////////////////////////////////////////////////////////////////////////
		// And finally the custom resource forwarder
		newResource, err := newCloudFormationResource("Custom::SpartaCloudWatchEventsPermission", logger)
		if nil != err {
			return "", err
		}
		customResource := newResource.(*cloudformationCloudWatchEventsPermissionResource)
		customResource.ServiceToken = gocf.GetAtt(eachConfigResource, "Arn")
		customResource.Rules = globallyUniqueRules
		customResource.LambdaTarget = gocf.GetAtt(lambdaLogicalCFResourceName, "Arn")

		// Name?
		resourceInvokerName := CloudFormationResourceName("ConfigCloudWatchEvents",
			lambdaLogicalCFResourceName,
			perm.BasePermission.SourceAccount)
		// Add it
		cfResource := template.AddResource(resourceInvokerName, customResource)
		cfResource.DependsOn = append(cfResource.DependsOn, dependsOn...)
	}
	return "", nil
}

func (perm CloudWatchEventsPermission) descriptionInfo() (string, string) {
	var ruleTriggers = " "
	for eachName, eachRule := range perm.Rules {
		filter := eachRule.ScheduleExpression
		if "" == filter && nil != eachRule.EventPattern {
			filter = fmt.Sprintf("%v", eachRule.EventPattern["source"])
		}
		ruleTriggers = fmt.Sprintf("%s-(%s)\n%s", eachName, filter, ruleTriggers)
	}
	return "CloudWatch Events", fmt.Sprintf("%s", ruleTriggers)
}

//
// END - CloudWatchEventsPermission
///////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// START - CloudWatchLogsPermission
//

// CloudWatchLogsSubscriptionFilter represents the CloudWatchLog filters
type CloudWatchLogsSubscriptionFilter struct {
	FilterPattern string
	LogGroupName  interface{}
}

var cloudformationLogsSourceArnParts = []gocf.Stringable{}

// CloudWatchLogsPermission struct that imples the S3 BasePermission.SourceArn should be
// updated (via PutBucketNotificationConfiguration) to automatically push
// events to the owning Lambda.
// See http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html#intro-core-components-event-sources
// for more information.
type CloudWatchLogsPermission struct {
	BasePermission
	// Map of filter names to the CloudWatchLogsSubscriptionFilter settings
	Filters map[string]CloudWatchLogsSubscriptionFilter
}

func (perm CloudWatchLogsPermission) export(serviceName string,
	lambdaLogicalCFResourceName string,
	template *gocf.Template,
	S3Bucket string,
	S3Key string,
	logger *logrus.Logger) (string, error) {

	// Tell the user we're ignoring any Arns provided, since it doesn't make sense for
	// this.
	if nil != perm.BasePermission.SourceArn &&
		perm.BasePermission.sourceArnExpr(cloudformationLogsSourceArnParts...).String() != wildcardArn.String() {
		logger.WithFields(logrus.Fields{
			"Arn": perm.BasePermission.sourceArnExpr(cloudformationEventsSourceArnParts...),
		}).Warn("CloudWatchLogs do not support literal ARN values")
	}
	// First thing we need to do is uniqueify the rule names s.t. we prevent
	// collisions with other stacks.
	configurationResourceNames := make(map[string]int, 0)

	globallyUniqueFilters := make(map[string]CloudWatchLogsSubscriptionFilter, len(perm.Filters))
	for eachFilterName, eachFilter := range perm.Filters {
		filterPrefix := fmt.Sprintf("%s_%s", serviceName, eachFilterName)
		uniqueFilterName := CloudFormationResourceName(filterPrefix, lambdaLogicalCFResourceName)
		globallyUniqueFilters[uniqueFilterName] = eachFilter

		// Ensure the configuration resource exists for this log source.  Cache the returned
		// logical resource name s.t. we can validate we're reusing the same resource
		configuratorResName, err := ensureConfiguratorLambdaResource(CloudWatchLogsPrincipal,
			gocf.String("arn:aws:logs:*:*:*"),
			[]string{},
			template,
			S3Bucket,
			S3Key,
			logger)
		if nil != err {
			return "", err
		}

		configurationResourceNames[configuratorResName] = 1
	}
	if len(configurationResourceNames) > 1 {
		return "", fmt.Errorf("Internal integrity check failed. Multiple configurators (%d) provisioned for CloudWatchLogs",
			len(configurationResourceNames))
	}
	logger.Info("WTF 6")
	// Insert the invocation
	for eachConfigResource := range configurationResourceNames {
		//////////////////////////////////////////////////////////////////////////////
		// And finally the custom resource forwarder
		newResource, err := newCloudFormationResource("Custom::SpartaCloudWatchLogsPermission", logger)
		if nil != err {
			return "", err
		}
		customResource := newResource.(*cloudformationCloudWatchLogsPermissionResource)
		customResource.ServiceToken = gocf.GetAtt(eachConfigResource, "Arn")
		customResource.Filters = globallyUniqueFilters
		customResource.LambdaTarget = gocf.GetAtt(lambdaLogicalCFResourceName, "Arn")

		// Name?
		resourceInvokerName := CloudFormationResourceName("ConfigCloudWatchLogs",
			lambdaLogicalCFResourceName,
			perm.BasePermission.SourceAccount)
		// Add it
		cfResource := template.AddResource(resourceInvokerName, customResource)

		cfResource.DependsOn = append(cfResource.DependsOn,
			lambdaLogicalCFResourceName,
			eachConfigResource)
	}
	return "", nil
}

func (perm CloudWatchLogsPermission) descriptionInfo() (string, string) {
	return "CloudWatch Logs", "TBD"
}

//
// END - CloudWatchLogsPermission
///////////////////////////////////////////////////////////////////////////////////
