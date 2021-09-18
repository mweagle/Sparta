package sparta

import (
	"encoding/json"
	"fmt"
	"strings"

	goftags "github.com/awslabs/goformation/v5/cloudformation/tags"

	"github.com/aws/aws-sdk-go/service/s3"
	gof "github.com/awslabs/goformation/v5/cloudformation"
	gofevents "github.com/awslabs/goformation/v5/cloudformation/events"
	goflambda "github.com/awslabs/goformation/v5/cloudformation/lambda"
	gofs3 "github.com/awslabs/goformation/v5/cloudformation/s3"
	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	cfCustomResources "github.com/mweagle/Sparta/aws/cloudformation/resources"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

////////////////////////////////////////////////////////////////////////////////
// Types to handle permissions & push source configuration

// describeInfoValue is a utility function that accepts
// some type of dynamic gocf value and transforms it into
// something that is `describe` output compatible
func describeInfoValue(dynamicValue interface{}) string {
	switch typedArn := dynamicValue.(type) {
	case string:
		data, dataErr := json.Marshal(typedArn)
		if dataErr != nil {
			data = []byte(fmt.Sprintf("%v", typedArn))
		}
		return string(data)
	default:
		panic(fmt.Sprintf("Unsupported dynamic value type for `describe`: %+v", typedArn))
	}
}

type descriptionNode struct {
	Name     string
	Relation string
	Color    string
}

// LambdaPermissionExporter defines an interface for polymorphic collection of
// Permission entries that support specialization for additional resource generation.
type LambdaPermissionExporter interface {
	// Export the permission object to a set of CloudFormation resources
	// in the provided resources param.  The targetLambdaFuncRef
	// interface represents the Fn::GetAtt "Arn" JSON value
	// of the parent Lambda target
	export(serviceName string,
		lambdaFunctionDisplayName string,
		lambdaLogicalCFResourceName string,
		template *gof.Template,
		lambdaFunctionCode *goflambda.Function_Code,
		logger *zerolog.Logger) (string, error)
	// Return a `describe` compatible output for the given permission.  Return
	// value is a list of tuples for node, edgeLabel
	descriptionInfo() ([]descriptionNode, error)
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

func (perm *BasePermission) sourceArnExpr(joinParts ...string) string {
	if perm.SourceArn == nil {
		return ""
	}
	stringARN, stringARNOk := perm.SourceArn.(string)
	if stringARNOk && strings.Contains(stringARN, "arn:aws:") {
		return stringARN
	}

	var parts []string
	if nil != joinParts {
		parts = append(parts, joinParts...)
	}
	parts = append(parts,
		spartaCF.DynamicValueToStringExpr(perm.SourceArn),
	)
	return gof.Join("", parts)
}

func (perm BasePermission) export(principal string,
	arnPrefixParts []string,
	lambdaFunctionDisplayName string,
	lambdaLogicalCFResourceName string,
	template *gof.Template,
	lambdaFunctionCode *goflambda.Function_Code,
	logger *zerolog.Logger) (string, error) {

	lambdaPermission := &goflambda.Permission{
		Action:       "lambda:InvokeFunction",
		FunctionName: gof.GetAtt(lambdaLogicalCFResourceName, "Arn"),
		Principal:    principal,
	}
	// If the Arn isn't the wildcard value, then include it.
	if nil != perm.SourceArn {
		switch typedARN := perm.SourceArn.(type) {
		case string:
			// Don't be smart if the Arn value is a user supplied literal
			if typedARN != "*" {
				lambdaPermission.SourceArn = typedARN
			}
		default:
			lambdaPermission.SourceArn = perm.sourceArnExpr(arnPrefixParts...)
		}
	}

	if perm.SourceAccount != "" {
		lambdaPermission.SourceAccount = perm.SourceAccount
	}

	arnLiteral, arnLiteralErr := json.Marshal(lambdaPermission.SourceArn)
	if nil != arnLiteralErr {
		return "", arnLiteralErr
	}
	resourceName := CloudFormationResourceName("LambdaPerm%s",
		principal,
		string(arnLiteral),
		lambdaLogicalCFResourceName)
	template.Resources[resourceName] = lambdaPermission
	return resourceName, nil
}

//
// END - BasePermission
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// START - S3Permission
//
var s3SourceArnParts = []string{
	"arn:aws:s3:::",
}

// S3Permission struct implies that the S3 BasePermission.SourceArn should be
// updated (via PutBucketNotificationConfiguration) to automatically push
// events to the owning Lambda.
// See http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html#intro-core-components-event-sources
// for more information.
type S3Permission struct {
	BasePermission
	// S3 events to register for (eg: `[]string{s3:GetObjectObjectCreated:*", "s3:ObjectRemoved:*"}`).
	Events []string `json:"Events,omitempty"`
	// S3.NotificationConfigurationFilter
	// to scope event forwarding.  See
	// 		http://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html
	// for more information.
	Filter s3.NotificationConfigurationFilter `json:"Filter,omitempty"`
}

func (perm S3Permission) export(serviceName string,
	lambdaFunctionDisplayName string,
	lambdaLogicalCFResourceName string,
	template *gof.Template,
	lambdaFunctionCode *goflambda.Function_Code,
	logger *zerolog.Logger) (string, error) {

	targetLambdaResourceName, err := perm.BasePermission.export("s3.amazonaws.com",
		s3SourceArnParts,
		lambdaFunctionDisplayName,
		lambdaLogicalCFResourceName,
		template,
		lambdaFunctionCode,
		logger)

	if nil != err {
		return "", errors.Wrap(err, "Failed to export S3 permission")
	}

	// Make sure the custom lambda that manages s3 notifications is provisioned.
	sourceArnExpression := perm.BasePermission.sourceArnExpr(s3SourceArnParts...)
	configuratorResName, err := EnsureCustomResourceHandler(serviceName,
		cfCustomResources.S3LambdaEventSource,
		sourceArnExpression,
		[]string{},
		template,
		lambdaFunctionCode,
		logger)

	if nil != err {
		return "", errors.Wrap(err, "Exporting S3 permission")
	}

	// Add a custom resource invocation for this configuration
	//////////////////////////////////////////////////////////////////////////////
	newResource, newResourceError := newCloudFormationResource(cfCustomResources.S3LambdaEventSource,
		logger)
	if nil != newResourceError {
		return "", newResourceError
	}

	// Setup the reqest for the S3 action
	s3Resource, s3ResourceOK := newResource.(*cfCustomResources.S3LambdaEventSourceResource)
	if !s3ResourceOK {
		return "", fmt.Errorf("failed to access typed S3CustomResource")
	}

	s3Resource.ServiceToken = gof.GetAtt(configuratorResName, "Arn")
	s3Resource.BucketArn = sourceArnExpression
	s3Resource.LambdaTargetArn = gof.GetAtt(lambdaLogicalCFResourceName, "Arn")
	s3Resource.Events = perm.Events
	if nil != perm.Filter.Key {
		s3Resource.Filter = &perm.Filter
	}

	// Name?
	resourceInvokerName := CloudFormationResourceName("ConfigS3",
		lambdaLogicalCFResourceName,
		perm.BasePermission.SourceAccount,
		fmt.Sprintf("%#v", s3Resource.Filter))

	// Add it
	s3Resource.AWSCloudFormationDependsOn = []string{
		targetLambdaResourceName,
		configuratorResName,
	}
	template.Resources[resourceInvokerName] = s3Resource
	return "", nil
}

func (perm S3Permission) descriptionInfo() ([]descriptionNode, error) {
	s3Events := ""
	for _, eachEvent := range perm.Events {
		s3Events = fmt.Sprintf("%s\n%s", eachEvent, s3Events)
	}
	nodes := make([]descriptionNode, 0)
	if perm.Filter.Key == nil || len(perm.Filter.Key.FilterRules) == 0 {
		nodes = append(nodes, descriptionNode{
			Name:     describeInfoValue(perm.SourceArn),
			Relation: s3Events,
		})
	} else {
		for _, eachFilter := range perm.Filter.Key.FilterRules {
			filterRel := fmt.Sprintf("%s (%s = %s)",
				s3Events,
				*eachFilter.Name,
				*eachFilter.Value)
			nodes = append(nodes, descriptionNode{
				Name:     describeInfoValue(perm.SourceArn),
				Relation: filterRel,
			})
		}
	}

	return nodes, nil
}

// END - S3Permission
///////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// SNSPermission - START
var snsSourceArnParts = []string{}

// SNSPermission struct implies that the BasePermisison.SourceArn should be
// configured for subscriptions as part of this stacks provisioning.
// See http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html#intro-core-components-event-sources
// for more information.
type SNSPermission struct {
	BasePermission
}

func (perm SNSPermission) export(serviceName string,
	lambdaFunctionDisplayName string,
	lambdaLogicalCFResourceName string,
	template *gof.Template,
	lambdaFunctionCode *goflambda.Function_Code,
	logger *zerolog.Logger) (string, error) {
	sourceArnExpression := perm.BasePermission.sourceArnExpr(snsSourceArnParts...)

	targetLambdaResourceName, err := perm.BasePermission.export(SNSPrincipal,
		snsSourceArnParts,
		lambdaFunctionDisplayName,
		lambdaLogicalCFResourceName,
		template,
		lambdaFunctionCode,
		logger)
	if nil != err {
		return "", errors.Wrap(err, "Failed to export SNS permission")
	}

	// Make sure the custom lambda that manages s3 notifications is provisioned.
	configuratorResName, err := EnsureCustomResourceHandler(serviceName,
		cfCustomResources.SNSLambdaEventSource,
		sourceArnExpression,
		[]string{},
		template,
		lambdaFunctionCode,
		logger)

	if nil != err {
		return "", errors.Wrap(err, "Exporing SNS permission handler")
	}

	// Add a custom resource invocation for this configuration
	//////////////////////////////////////////////////////////////////////////////
	newResource, newResourceError := newCloudFormationResource(cfCustomResources.SNSLambdaEventSource,
		logger)
	if nil != newResourceError {
		return "", newResourceError
	}

	customResource := newResource.(*cfCustomResources.SNSLambdaEventSourceResource)
	customResource.ServiceToken = gof.GetAtt(configuratorResName, "Arn")
	customResource.LambdaTargetArn = gof.GetAtt(lambdaLogicalCFResourceName, "Arn")
	customResource.SNSTopicArn = sourceArnExpression

	// Name?
	resourceInvokerName := CloudFormationResourceName("ConfigSNS",
		lambdaLogicalCFResourceName,
		perm.BasePermission.SourceAccount)

	// Add it
	customResource.AWSCloudFormationDependsOn = []string{
		targetLambdaResourceName,
		configuratorResName,
	}
	template.Resources[resourceInvokerName] = customResource
	return "", nil
}

func (perm SNSPermission) descriptionInfo() ([]descriptionNode, error) {
	nodes := []descriptionNode{
		{
			Name:     describeInfoValue(perm.SourceArn),
			Relation: "",
		},
	}
	return nodes, nil
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
// `ServiceName/RuleName/`, which can be overridden by specifying a non-empty
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
	bucketNameExpr                     string
	cloudFormationS3BucketResourceName string
}

// BucketArn returns an Arn value that can be used as an
// lambdaFn.RoleDefinition.Privileges `Resource` value.
func (storage *MessageBodyStorage) BucketArn() string {
	return gof.Join("", []string{
		"arn:aws:s3:::",
		storage.bucketNameExpr})
}

// BucketArnAllKeys returns an Arn value that can be used
// lambdaFn.RoleDefinition.Privileges `Resource` value.  It includes
// the trailing `/*` wildcard to support item acccess
func (storage *MessageBodyStorage) BucketArnAllKeys() string {
	return gof.Join("", []string{
		"arn:aws:s3:::",
		storage.bucketNameExpr,
		"/*"})
}

func (storage *MessageBodyStorage) export(serviceName string,
	lambdaFunctionDisplayName string,
	lambdaLogicalCFResourceName string,
	template *gof.Template,
	lambdaFunctionCode *goflambda.Function_Code,
	logger *zerolog.Logger) (string, error) {

	if storage.cloudFormationS3BucketResourceName != "" {
		s3Bucket := &gofs3.Bucket{
			Tags: []goftags.Tag{
				goftags.Tag{
					Key:   "sparta:logicalBucketName",
					Value: storage.logicalBucketName,
				},
			},
		}
		s3Bucket.AWSCloudFormationDeletionPolicy = "Retain"
		template.Resources[storage.cloudFormationS3BucketResourceName] = s3Bucket

		lambdaResource, lambdaResourceExists := template.Resources[lambdaLogicalCFResourceName]
		if !lambdaResourceExists {
			safeAppendDependency(lambdaResource, storage.cloudFormationS3BucketResourceName)
		}

		logger.Info().
			Str("LogicalResourceName", storage.cloudFormationS3BucketResourceName).
			Msg("Service will orphan S3 Bucket on deletion")

		// Save the output
		template.Outputs[storage.cloudFormationS3BucketResourceName] = gof.Output{
			Description: "SES Message Body Bucket",
			Value:       gof.Ref(storage.cloudFormationS3BucketResourceName),
		}
	}
	// Add the S3 Access policy
	s3BodyStoragePolicy := &gofs3.BucketPolicy{
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
					"Resource": gof.Join("", []string{
						"arn:aws:s3:::",
						storage.bucketNameExpr,
						"/*"}),
					"Condition": ArbitraryJSONObject{
						"StringEquals": ArbitraryJSONObject{
							"aws:Referer": gof.Ref("AWS::AccountId"),
						},
					},
				},
			},
		},
	}

	s3BucketPolicyResourceName := CloudFormationResourceName("SESMessageBodyBucketPolicy",
		fmt.Sprintf("%#v", storage.bucketNameExpr))
	template.Resources[s3BucketPolicyResourceName] = s3BodyStoragePolicy

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
	InvocationType     string
	BodyStorageOptions MessageBodyStorageOptions
}

func (rule *ReceiptRule) toResourceRule(serviceName string,
	functionArnRef interface{},
	messageBodyStorage *MessageBodyStorage) *cfCustomResources.SESLambdaEventSourceResourceRule {

	resourceRule := &cfCustomResources.SESLambdaEventSourceResourceRule{
		Name:        gocf.String(rule.Name),
		ScanEnabled: gocf.Bool(!rule.ScanDisabled),
		Enabled:     gocf.Bool(!rule.Disabled),
		Actions:     make([]*cfCustomResources.SESLambdaEventSourceResourceAction, 0),
		Recipients:  make([]*gocf.StringExpr, 0),
	}
	for _, eachRecipient := range rule.Recipients {
		resourceRule.Recipients = append(resourceRule.Recipients, gocf.String(eachRecipient))
	}
	if rule.TLSPolicy != "" {
		resourceRule.TLSPolicy = gocf.String(rule.TLSPolicy)
	}

	// If there is a MessageBodyStorage reference, push that S3Action
	// to the head of the Actions list
	if nil != messageBodyStorage && !rule.BodyStorageOptions.DisableStorage {
		s3Action := &cfCustomResources.SESLambdaEventSourceResourceAction{
			ActionType: gocf.String("S3Action"),
			ActionProperties: map[string]interface{}{
				"BucketName": messageBodyStorage.bucketNameExpr,
			},
		}
		if rule.BodyStorageOptions.ObjectKeyPrefix != "" {
			s3Action.ActionProperties["ObjectKeyPrefix"] = rule.BodyStorageOptions.ObjectKeyPrefix
		}
		if rule.BodyStorageOptions.KmsKeyArn != "" {
			s3Action.ActionProperties["KmsKeyArn"] = rule.BodyStorageOptions.KmsKeyArn
		}
		if rule.BodyStorageOptions.TopicArn != "" {
			s3Action.ActionProperties["TopicArn"] = rule.BodyStorageOptions.TopicArn
		}
		resourceRule.Actions = append(resourceRule.Actions, s3Action)
	}
	// There's always a lambda action
	lambdaAction := &cfCustomResources.SESLambdaEventSourceResourceAction{
		ActionType: gocf.String("LambdaAction"),
		ActionProperties: map[string]interface{}{
			"FunctionArn": functionArnRef,
		},
	}
	lambdaAction.ActionProperties["InvocationType"] = rule.InvocationType
	if rule.InvocationType == "" {
		lambdaAction.ActionProperties["InvocationType"] = "Event"
	}
	if rule.TopicArn != "" {
		lambdaAction.ActionProperties["TopicArn"] = rule.TopicArn
	}
	resourceRule.Actions = append(resourceRule.Actions, lambdaAction)
	return resourceRule
}

//
// END - ReceiptRule
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// SESPermission - START

// SES doesn't use ARNs to scope access
var sesSourcePartArn = []string{wildcardArn}

// SESPermission struct implies that the SES verified domain should be
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
	store.bucketNameExpr = gof.Ref(store.cloudFormationS3BucketResourceName)
	return store, nil
}

// NewMessageBodyStorageReference uses a pre-existing S3 bucket for MessageBody storage.
// Sparta assumes that prexistingBucketName exists and will add an S3::BucketPolicy
// to enable SES PutObject access.
func (perm *SESPermission) NewMessageBodyStorageReference(prexistingBucketName string) (*MessageBodyStorage, error) {
	store := &MessageBodyStorage{}
	store.bucketNameExpr = prexistingBucketName
	return store, nil
}

func (perm SESPermission) export(serviceName string,
	lambdaFunctionDisplayName string,
	lambdaLogicalCFResourceName string,
	template *gof.Template,
	lambdaFunctionCode *goflambda.Function_Code,
	logger *zerolog.Logger) (string, error) {

	sourceArnExpression := perm.BasePermission.sourceArnExpr(snsSourceArnParts...)

	targetLambdaResourceName, err := perm.BasePermission.export(SESPrincipal,
		sesSourcePartArn,
		lambdaFunctionDisplayName,
		lambdaLogicalCFResourceName,
		template,
		lambdaFunctionCode,
		logger)
	if nil != err {
		return "", errors.Wrap(err, "Failed to export SES permission")
	}

	// MessageBody storage?
	var dependsOn []string
	if nil != perm.MessageBodyStorage {
		s3Policy, s3PolicyErr := perm.MessageBodyStorage.export(serviceName,
			lambdaFunctionDisplayName,
			lambdaLogicalCFResourceName,
			template,
			lambdaFunctionCode,
			logger)
		if nil != s3PolicyErr {
			return "", s3PolicyErr
		}
		if s3Policy != "" {
			dependsOn = append(dependsOn, s3Policy)
		}
	}

	// Make sure the custom lambda that manages SNS notifications is provisioned.
	configuratorResName, err := EnsureCustomResourceHandler(serviceName,
		cfCustomResources.SESLambdaEventSource,
		sourceArnExpression,
		dependsOn,
		template,
		lambdaFunctionCode,
		logger)

	if nil != err {
		return "", errors.Wrap(err, "Ensuring custom resource handler for SES")
	}

	// Add a custom resource invocation for this configuration
	//////////////////////////////////////////////////////////////////////////////
	newResource, newResourceError := newCloudFormationResource(cfCustomResources.SESLambdaEventSource, logger)
	if nil != newResourceError {
		return "", newResourceError
	}
	customResource := newResource.(*cfCustomResources.SESLambdaEventSourceResource)
	customResource.ServiceToken = gof.GetAtt(configuratorResName, "Arn")
	// The shared ruleset name used by all Sparta applications
	customResource.RuleSetName = gocf.String("RuleSet")

	///////////////////
	// Build up the Rules
	// If there aren't any rules, make one that forwards everything...
	sesLength := 0
	if perm.ReceiptRules == nil {
		sesLength = 1
	} else {
		sesLength = len(perm.ReceiptRules)
	}
	sesRules := make([]*cfCustomResources.SESLambdaEventSourceResourceRule, sesLength)
	if nil == perm.ReceiptRules {
		sesRules[0] = &cfCustomResources.SESLambdaEventSourceResourceRule{
			Name:        gocf.String("Default"),
			Actions:     make([]*cfCustomResources.SESLambdaEventSourceResourceAction, 0),
			ScanEnabled: gocf.Bool(false),
			Enabled:     gocf.Bool(true),
			Recipients:  []*gocf.StringExpr{},
			TLSPolicy:   gocf.String("Optional"),
		}
	} else {
		// Append all the user defined ones
		for eachIndex, eachReceiptRule := range perm.ReceiptRules {
			sesRules[eachIndex] = eachReceiptRule.toResourceRule(
				serviceName,
				gof.GetAtt(lambdaLogicalCFResourceName, "Arn"),
				perm.MessageBodyStorage)
		}
	}

	customResource.Rules = sesRules
	// Name?
	resourceInvokerName := CloudFormationResourceName("ConfigSNS",
		lambdaLogicalCFResourceName,
		perm.BasePermission.SourceAccount)

	// Add it
	customResource.AWSCloudFormationDependsOn = []string{
		targetLambdaResourceName,
		configuratorResName,
	}
	template.Resources[resourceInvokerName] = customResource
	return "", nil
}

func (perm SESPermission) descriptionInfo() ([]descriptionNode, error) {
	nodes := []descriptionNode{
		{
			Name:     "SimpleEmailService",
			Relation: "All verified domain(s) email",
		},
	}
	return nodes, nil
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

	if rule.Description != "" {
		ruleJSON["Description"] = rule.Description
	}
	if nil != rule.EventPattern {
		eventPatternString, err := json.Marshal(rule.EventPattern)
		if nil != err {
			return nil, err
		}
		ruleJSON["EventPattern"] = string(eventPatternString)
	}
	if rule.ScheduleExpression != "" {
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
var cloudformationEventsSourceArnParts = []string{}

// CloudWatchEventsPermission struct implies that the CloudWatchEvent sources
// should be configured as part of provisioning.  The BasePermission.SourceArn
// isn't considered for this configuration. Each CloudWatchEventsRule struct
// in the Rules map is used to register for push based event notifications via
// `putRule` and `deleteRule`.
// See http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html#intro-core-components-event-sources
// for more information.
type CloudWatchEventsPermission struct {
	BasePermission
	// Map of rule names to events that trigger the lambda function
	Rules map[string]CloudWatchEventsRule
}

func (perm CloudWatchEventsPermission) export(serviceName string,
	lambdaFunctionDisplayName string,
	lambdaLogicalCFResourceName string,
	template *gof.Template,
	lambdaFunctionCode *goflambda.Function_Code,
	logger *zerolog.Logger) (string, error) {

	// There needs to be at least one rule to apply
	if len(perm.Rules) <= 0 {
		return "", fmt.Errorf("function %s CloudWatchEventsPermission does not specify any expressions", lambdaFunctionDisplayName)
	}

	// Tell the user we're ignoring any Arns provided, since it doesn't make sense for this.
	if nil != perm.BasePermission.SourceArn &&
		perm.BasePermission.sourceArnExpr(cloudformationEventsSourceArnParts...) != wildcardArn {
		logger.Warn().
			Interface("Arn", perm.BasePermission.sourceArnExpr(cloudformationEventsSourceArnParts...)).
			Msg("CloudWatchEvents do not support literal ARN values")
	}

	arnPermissionForRuleName := func(ruleName string) string {
		return gof.Join("", []string{
			"arn:aws:events:",
			gof.Ref("AWS::Region"),
			":",
			gof.Ref("AWS::AccountId"),
			":rule/",
			ruleName})
	}

	// Add the permission to invoke the lambda function
	uniqueRuleNameMap := make(map[string]int)
	for eachRuleName, eachRuleDefinition := range perm.Rules {

		// We need a stable unique name s.t. the permission is properly configured...
		uniqueRuleName := CloudFormationResourceName(eachRuleName, lambdaFunctionDisplayName, serviceName)
		uniqueRuleNameMap[uniqueRuleName]++

		// Add the permission
		basePerm := BasePermission{
			SourceArn: arnPermissionForRuleName(uniqueRuleName),
		}
		_, exportErr := basePerm.export(CloudWatchEventsPrincipal,
			cloudformationEventsSourceArnParts,
			lambdaFunctionDisplayName,
			lambdaLogicalCFResourceName,
			template,
			lambdaFunctionCode,
			logger)

		if nil != exportErr {
			return "", exportErr
		}

		cwEventsRuleTargetList := []gofevents.Rule_Target{
			gofevents.Rule_Target{
				Arn: gof.GetAtt(lambdaLogicalCFResourceName, "Arn"),
				Id:  uniqueRuleName,
			},
		}
		// Add the rule
		eventsRule := &gofevents.Rule{
			Name:        uniqueRuleName,
			Description: eachRuleDefinition.Description,
			Targets:     cwEventsRuleTargetList,
		}
		if nil != eachRuleDefinition.EventPattern && eachRuleDefinition.ScheduleExpression != "" {
			return "", fmt.Errorf("rule %s CloudWatchEvents specifies both EventPattern and ScheduleExpression", eachRuleName)
		}
		if nil != eachRuleDefinition.EventPattern {
			eventsRule.EventPattern = eachRuleDefinition.EventPattern
		} else if eachRuleDefinition.ScheduleExpression != "" {
			eventsRule.ScheduleExpression = eachRuleDefinition.ScheduleExpression
		}
		cloudWatchLogsEventResName := CloudFormationResourceName(fmt.Sprintf("%s-CloudWatchEventsRule", eachRuleName),
			lambdaLogicalCFResourceName,
			lambdaFunctionDisplayName)

		template.Resources[cloudWatchLogsEventResName] = eventsRule
	}
	// Validate it
	for _, eachCount := range uniqueRuleNameMap {
		if eachCount != 1 {
			return "", fmt.Errorf("integrity violation for CloudWatchEvent Rulenames: %#v", uniqueRuleNameMap)
		}
	}
	return "", nil
}

func (perm CloudWatchEventsPermission) descriptionInfo() ([]descriptionNode, error) {
	var ruleTriggers = " "
	for eachName, eachRule := range perm.Rules {
		filter := eachRule.ScheduleExpression
		if filter == "" && eachRule.EventPattern != nil {
			filter = fmt.Sprintf("%v", eachRule.EventPattern["source"])
		}
		ruleTriggers = fmt.Sprintf("%s-(%s)\n%s", eachName, filter, ruleTriggers)
	}
	nodes := []descriptionNode{
		{
			Name:     "CloudWatch Events",
			Relation: ruleTriggers,
		},
	}
	return nodes, nil
}

//
// END - CloudWatchEventsPermission
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// START - EventBridgeRule
//

// EventBridgeRule defines parameters for invoking a lambda function
// in response to specific EventBridge triggers
type EventBridgeRule struct {
	Description  string
	EventBusName string
	// ArbitraryJSONObject filter for events as documented at
	// https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-events-rule.html#cfn-events-rule-eventpattern
	// Rules matches should use the JSON representation (NOT the string form).  Sparta will serialize
	// the map[string]interface{} to a string form during CloudFormation Template
	// marshalling.
	EventPattern map[string]interface{} `json:"EventPattern,omitempty"`
	// Schedule pattern per
	// https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-events-rule.html#cfn-events-rule-scheduleexpression
	ScheduleExpression string
}

// MarshalJSON customizes the JSON representation used when serializing to the
// CloudFormation template representation.
func (rule EventBridgeRule) MarshalJSON() ([]byte, error) {
	ruleJSON := map[string]interface{}{}

	ruleJSON["Description"] = rule.Description
	ruleJSON["EventBusName"] = rule.EventBusName
	if rule.EventPattern != nil {
		ruleJSON["EventPattern"] = rule.EventPattern
	}
	if rule.ScheduleExpression != "" {
		ruleJSON["ScheduleExpression"] = rule.ScheduleExpression
	}
	return json.Marshal(ruleJSON)
}

//
// END - EventBridgeRule
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// START - EventBridgePermission
//

// EventBridgePermission struct implies that the EventBridge sources
// should be configured as part of provisioning.  The BasePermission.SourceArn
// isn't considered for this configuration. Each EventBridge Rule or Schedule struct
// in the Rules map is used to register for push based event notifications via
// `putRule` and `deleteRule`.
// See http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html#intro-core-components-event-sources
// for more information.
type EventBridgePermission struct {
	BasePermission
	// EventBridgeRule for this permission
	Rule *EventBridgeRule
}

func (perm EventBridgePermission) export(serviceName string,
	lambdaFunctionDisplayName string,
	lambdaLogicalCFResourceName string,
	template *gof.Template,
	lambdaFunctionCode *goflambda.Function_Code,
	logger *zerolog.Logger) (string, error) {

	// There needs to be at least one rule to apply
	if perm.Rule == nil {
		return "", fmt.Errorf("function %s EventBridgePermission does not specify any EventBridgeRule",
			lambdaFunctionDisplayName)
	}

	// Name for the rule...
	eventBridgeRuleResourceName := CloudFormationResourceName(fmt.Sprintf("EventBridge-%s", lambdaLogicalCFResourceName),
		lambdaFunctionDisplayName)

	// Tell the user we're ignoring any Arns provided, since it doesn't make sense for this.
	if nil != perm.BasePermission.SourceArn &&
		perm.BasePermission.sourceArnExpr(cloudformationEventsSourceArnParts...) != wildcardArn {
		logger.Warn().
			Interface("Arn", perm.BasePermission.sourceArnExpr(cloudformationEventsSourceArnParts...)).
			Msg("EventBridge Events do not support literal ARN values")
	}

	// Add the permission
	basePerm := BasePermission{
		SourceArn: gof.GetAtt(eventBridgeRuleResourceName, "Arn"),
	}
	_, exportErr := basePerm.export(EventBridgePrincipal,
		cloudformationEventsSourceArnParts,
		lambdaFunctionDisplayName,
		lambdaLogicalCFResourceName,
		template,
		lambdaFunctionCode,
		logger)

	if nil != exportErr {
		return "", exportErr
	}

	eventBridgeRuleTargetList := []gofevents.Rule_Target{
		gofevents.Rule_Target{
			Arn: gof.GetAtt(lambdaLogicalCFResourceName, "Arn"),
			Id:  serviceName,
		},
	}
	if nil != perm.Rule.EventPattern &&
		perm.Rule.ScheduleExpression != "" {
		return "", fmt.Errorf("rule %s EventBridge specifies both EventPattern and ScheduleExpression",
			perm.Rule)
	}

	// Add the rule
	eventsRule := &gofevents.Rule{
		Targets: eventBridgeRuleTargetList,
	}
	if perm.Rule.EventBusName != "" {
		eventsRule.EventBusName = perm.Rule.EventBusName
	}
	// Setup the description placeholder...we'll set it in a bit...
	ruleDescription := ""
	if perm.Rule.EventPattern != nil {
		eventsRule.EventPattern = perm.Rule.EventPattern
		ruleDescription = fmt.Sprintf("%s (Stack: %s) event pattern subscriber",
			lambdaFunctionDisplayName,
			serviceName)
	} else if perm.Rule.ScheduleExpression != "" {
		eventsRule.ScheduleExpression = perm.Rule.ScheduleExpression
		ruleDescription = fmt.Sprintf("%s (Stack: %s) scheduled subscriber",
			lambdaFunctionDisplayName,
			serviceName)
	}
	eventsRule.Description = ruleDescription
	template.Resources[eventBridgeRuleResourceName] = eventsRule
	return "", nil
}

func (perm EventBridgePermission) descriptionInfo() ([]descriptionNode, error) {
	var ruleTriggers = " "

	filter := perm.Rule.ScheduleExpression
	if filter == "" && perm.Rule.EventPattern != nil {
		filter = fmt.Sprintf("%v", perm.Rule.EventPattern)
	}
	ruleTriggers = fmt.Sprintf("EventBridge-(%s)\n%s", filter, ruleTriggers)

	nodes := []descriptionNode{
		{
			Name:     "EventBridge Event",
			Relation: ruleTriggers,
		},
	}
	return nodes, nil
}

//
// END - CloudWatchEventsPermission
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// START - CloudWatchLogsPermission
//

// CloudWatchLogsSubscriptionFilter represents the CloudWatch Log filter
// information
type CloudWatchLogsSubscriptionFilter struct {
	FilterPattern string
	LogGroupName  string
}

var cloudformationLogsSourceArnParts = []string{
	"arn:aws:logs:",
}

// CloudWatchLogsPermission struct implies that the corresponding
// CloudWatchLogsSubscriptionFilter definitions should be configured during
// stack provisioning.  The BasePermission.SourceArn isn't considered for
// this configuration operation.  Configuration of the remote push source
// is done via `putSubscriptionFilter` and `deleteSubscriptionFilter`.
// See http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html#intro-core-components-event-sources
// for more information.
type CloudWatchLogsPermission struct {
	BasePermission
	// Map of filter names to the CloudWatchLogsSubscriptionFilter settings
	Filters map[string]CloudWatchLogsSubscriptionFilter
}

func (perm CloudWatchLogsPermission) export(serviceName string,
	lambdaFunctionDisplayName string,
	lambdaLogicalCFResourceName string,
	template *gof.Template,
	lambdaFunctionCode *goflambda.Function_Code,
	logger *zerolog.Logger) (string, error) {

	// If there aren't any expressions to register with?
	if len(perm.Filters) <= 0 {
		return "", fmt.Errorf("function %s CloudWatchLogsPermission does not specify any filters", lambdaFunctionDisplayName)
	}

	// The principal is region specific, so build that up...
	regionalPrincipal := gof.Join(".", []string{
		"logs",
		gof.Ref("AWS::Region"),
		"amazonaws.com"})

	// Tell the user we're ignoring any Arns provided, since it doesn't make sense for
	// this.
	if nil != perm.BasePermission.SourceArn &&
		perm.BasePermission.sourceArnExpr(cloudformationLogsSourceArnParts...) != wildcardArn {
		logger.Warn().
			Interface("Arn", perm.BasePermission.sourceArnExpr(cloudformationEventsSourceArnParts...)).
			Msg("CloudWatchLogs do not support literal ARN values")
	}

	// Make sure we grant InvokeFunction privileges to CloudWatchLogs
	lambdaInvokePermission, err := perm.BasePermission.export(regionalPrincipal,
		cloudformationLogsSourceArnParts,
		lambdaFunctionDisplayName,
		lambdaLogicalCFResourceName,
		template,
		lambdaFunctionCode,
		logger)
	if nil != err {
		return "", errors.Wrap(err, "Exporting regional CloudWatch log permission")
	}

	// Then we need to uniqueify the rule names s.t. we prevent
	// collisions with other stacks.
	configurationResourceNames := make(map[string]int)
	// Store the last name.  We'll do a uniqueness check when exiting the loop,
	// and if that passes, the last name will also be the unique one.
	var configurationResourceName string
	// Create the CustomResource entries
	globallyUniqueFilters := make(map[string]CloudWatchLogsSubscriptionFilter, len(perm.Filters))
	for eachFilterName, eachFilter := range perm.Filters {
		filterPrefix := fmt.Sprintf("%s_%s", serviceName, eachFilterName)
		uniqueFilterName := CloudFormationResourceName(filterPrefix, lambdaLogicalCFResourceName)
		globallyUniqueFilters[uniqueFilterName] = eachFilter

		// The ARN we supply to IAM is built up using the user supplied groupname
		cloudWatchLogsArn := gof.Join("", []string{
			"arn:aws:logs:",
			gof.Ref("AWS::Region"),
			":",
			gof.Ref("AWS::AccountId"),
			":log-group:",
			eachFilter.LogGroupName,
			":log-stream:*"})

		lastConfigurationResourceName, ensureCustomHandlerError := EnsureCustomResourceHandler(serviceName,
			cfCustomResources.CloudWatchLogsLambdaEventSource,
			cloudWatchLogsArn,
			[]string{},
			template,
			lambdaFunctionCode,
			logger)
		if nil != ensureCustomHandlerError {
			return "", errors.Wrap(err, "Ensuring CloudWatch permissions handler")
		}
		configurationResourceNames[configurationResourceName] = 1
		configurationResourceName = lastConfigurationResourceName
	}
	if len(configurationResourceNames) > 1 {
		return "", fmt.Errorf("internal integrity check failed. Multiple configurators (%d) provisioned for CloudWatchLogs",
			len(configurationResourceNames))
	}

	// Get the single configurator name from the

	// Add the custom resource that uses this...
	//////////////////////////////////////////////////////////////////////////////

	newResource, newResourceError := newCloudFormationResource(cfCustomResources.CloudWatchLogsLambdaEventSource, logger)
	if nil != newResourceError {
		return "", newResourceError
	}

	customResource := newResource.(*cfCustomResources.CloudWatchLogsLambdaEventSourceResource)
	customResource.ServiceToken = gof.GetAtt(configurationResourceName, "Arn")
	customResource.LambdaTargetArn = gof.GetAtt(lambdaLogicalCFResourceName, "Arn")
	// Build up the filters...
	customResource.Filters = make([]*cfCustomResources.CloudWatchLogsLambdaEventSourceFilter, 0)
	for eachName, eachFilter := range globallyUniqueFilters {
		customResource.Filters = append(customResource.Filters,
			&cfCustomResources.CloudWatchLogsLambdaEventSourceFilter{
				Name:         eachName,
				Pattern:      eachFilter.FilterPattern,
				LogGroupName: eachFilter.LogGroupName,
			})

	}

	resourceInvokerName := CloudFormationResourceName("ConfigCloudWatchLogs",
		lambdaLogicalCFResourceName,
		perm.BasePermission.SourceAccount)
	// Add it
	customResource.AWSCloudFormationDependsOn = []string{
		lambdaInvokePermission,
		lambdaLogicalCFResourceName,
		configurationResourceName,
	}
	template.Resources[resourceInvokerName] = customResource
	return "", nil
}

func (perm CloudWatchLogsPermission) descriptionInfo() ([]descriptionNode, error) {
	nodes := make([]descriptionNode, len(perm.Filters))
	nodeIndex := 0
	for eachFilterName, eachFilterDef := range perm.Filters {
		nodes[nodeIndex] = descriptionNode{
			Name:     describeInfoValue(eachFilterDef.LogGroupName),
			Relation: fmt.Sprintf("%s (%s)", eachFilterName, eachFilterDef.FilterPattern),
		}
		nodeIndex++
	}
	return nodes, nil
}

//
// END - CloudWatchLogsPermission
///////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// START - CodeCommitPermission
//
// arn:aws:codecommit:us-west-2:123412341234:myRepo
var codeCommitSourceArnParts = []string{
	"arn:aws:codecommit:",
	gof.Ref("AWS::Region"),
	":",
	gof.Ref("AWS::AccountId"),
	":",
}

// CodeCommitPermission struct encapsulates the data necessary
// to trigger the owning LambdaFunction in response to
// CodeCommit events
type CodeCommitPermission struct {
	BasePermission
	// RepositoryName
	RepositoryName string
	// Branches to register for
	Branches []string `json:"branches,omitempty"`
	// Events to subscribe to. Defaults to "all" if empty.
	Events []string `json:"events,omitempty"`
}

func (perm CodeCommitPermission) export(serviceName string,
	lambdaFunctionDisplayName string,
	lambdaLogicalCFResourceName string,
	template *gof.Template,
	lambdaFunctionCode *goflambda.Function_Code,
	logger *zerolog.Logger) (string, error) {

	principal := gof.Join("", []string{
		"codecommit.",
		gof.Ref("AWS::Region"),
		".amazonaws.com"})

	sourceArnExpression := perm.BasePermission.sourceArnExpr(codeCommitSourceArnParts...)

	targetLambdaResourceName, err := perm.BasePermission.export(principal,
		codeCommitSourceArnParts,
		lambdaFunctionDisplayName,
		lambdaLogicalCFResourceName,
		template,
		lambdaFunctionCode,
		logger)

	if nil != err {
		return "", errors.Wrap(err, "Failed to export CodeCommit permission")
	}

	// Make sure that the handler that manages triggers is registered.
	configuratorResName, err := EnsureCustomResourceHandler(serviceName,
		cfCustomResources.CodeCommitLambdaEventSource,
		sourceArnExpression,
		[]string{},
		template,
		lambdaFunctionCode,
		logger)

	if nil != err {
		return "", errors.Wrap(err, "Exporing CodeCommit permission handler")
	}

	// Add a custom resource invocation for this configuration
	//////////////////////////////////////////////////////////////////////////////
	newResource, newResourceError := newCloudFormationResource(cfCustomResources.CodeCommitLambdaEventSource,
		logger)
	if nil != newResourceError {
		return "", newResourceError
	}
	repoEvents := perm.Events
	if len(repoEvents) <= 0 {
		repoEvents = []string{"all"}
	}

	customResource := newResource.(*cfCustomResources.CodeCommitLambdaEventSourceResource)
	customResource.ServiceToken = gof.GetAtt(configuratorResName, "Arn")
	customResource.LambdaTargetArn = gof.GetAtt(lambdaLogicalCFResourceName, "Arn")
	customResource.TriggerName = gof.Ref(lambdaLogicalCFResourceName)
	customResource.RepositoryName = perm.RepositoryName
	customResource.Events = repoEvents
	customResource.Branches = perm.Branches

	// Name?
	resourceInvokerName := CloudFormationResourceName("ConfigCodeCommit",
		lambdaLogicalCFResourceName,
		perm.BasePermission.SourceAccount)

	// Add it
	customResource.AWSCloudFormationDependsOn = []string{
		targetLambdaResourceName,
		configuratorResName,
	}
	template.Resources[resourceInvokerName] = customResource
	return "", nil
}

func (perm CodeCommitPermission) descriptionInfo() ([]descriptionNode, error) {
	nodes := make([]descriptionNode, 0)
	if len(perm.Branches) <= 0 {
		nodes = append(nodes, descriptionNode{
			Name:     describeInfoValue(perm.SourceArn),
			Relation: "all",
		})
	} else {
		for _, eachBranch := range perm.Branches {
			filterRel := fmt.Sprintf("%s (%#v)",
				eachBranch,
				perm.Events)
			nodes = append(nodes, descriptionNode{
				Name:     describeInfoValue(perm.SourceArn),
				Relation: filterRel,
			})
		}
	}
	return nodes, nil
}

// END - CodeCommitPermission
///////////////////////////////////////////////////////////////////////////////////
