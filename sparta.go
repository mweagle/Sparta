package sparta

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"

	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	spartaIAM "github.com/mweagle/Sparta/aws/iam"
	gocc "github.com/mweagle/go-cloudcondenser"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type cloudFormationLambdaCustomResource struct {
	gocf.CloudFormationCustomResource
	ServiceToken   *gocf.StringExpr
	UserProperties map[string]interface{} `json:",omitempty"`
}

func customResourceProvider(resourceType string) gocf.ResourceProperties {
	switch resourceType {
	case cloudFormationLambda:
		{
			return &cloudFormationLambdaCustomResource{}
		}
	default:
		return nil
	}
}

func init() {
	gocf.RegisterCustomResourceProvider(customResourceProvider)
	rand.Seed(time.Now().Unix())
}

func noopMessage(operationName string) string {
	return fmt.Sprintf("Skipping %s due to -n/-noop flag",
		operationName)
}

/******************************************************************************/
// Global options
type optionsGlobalStruct struct {
	ServiceName        string         `validate:"required"`
	ServiceDescription string         `validate:"-"`
	Noop               bool           `validate:"-"`
	LogLevel           string         `validate:"eq=panic|eq=fatal|eq=error|eq=warn|eq=info|eq=debug"`
	LogFormat          string         `validate:"eq=txt|eq=text|eq=json"`
	TimeStamps         bool           `validate:"-"`
	Logger             *logrus.Logger `validate:"-"`
	Command            string         `validate:"-"`
	BuildTags          string         `validate:"-"`
	LinkerFlags        string         `validate:"-"` // no requirements
	DisableColors      bool           `validate:"-"`
}

// OptionsGlobal stores the global command line options
var OptionsGlobal optionsGlobalStruct

////////////////////////////////////////////////////////////////////////////////
// Variables
////////////////////////////////////////////////////////////////////////////////

// Represents the CloudFormation Arn of this stack, referenced
// in CommonIAMStatements
var cloudFormationThisStackArn = []gocf.Stringable{gocf.String("arn:aws:cloudformation:"),
	gocf.Ref("AWS::Region").String(),
	gocf.String(":"),
	gocf.Ref("AWS::AccountId").String(),
	gocf.String(":stack/"),
	gocf.Ref("AWS::StackName").String(),
	gocf.String("/*")}

// CommonIAMStatements defines common IAM::Role Policy Statement values for different AWS
// service types.  See http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html#genref-aws-service-namespaces
// for names.
// http://docs.aws.amazon.com/lambda/latest/dg/monitoring-functions.html
// for more information.
var CommonIAMStatements = struct {
	Core     []spartaIAM.PolicyStatement
	VPC      []spartaIAM.PolicyStatement
	DynamoDB []spartaIAM.PolicyStatement
	Kinesis  []spartaIAM.PolicyStatement
	SQS      []spartaIAM.PolicyStatement
}{
	Core: []spartaIAM.PolicyStatement{
		{
			Action: []string{"logs:CreateLogGroup",
				"logs:CreateLogStream",
				"logs:PutLogEvents"},
			Effect: "Allow",
			Resource: gocf.Join("",
				gocf.String("arn:aws:logs:"),
				gocf.Ref("AWS::Region"),
				gocf.String(":"),
				gocf.Ref("AWS::AccountId"),
				gocf.String(":*")),
		},
		{
			Effect: "Allow",
			Action: []string{"cloudformation:DescribeStacks",
				"cloudformation:DescribeStackResource"},
			Resource: gocf.Join("", cloudFormationThisStackArn...),
		},
		// http://docs.aws.amazon.com/lambda/latest/dg/lambda-x-ray.html#enabling-x-ray
		{
			Effect: "Allow",
			Action: []string{"xray:PutTraceSegments",
				"xray:PutTelemetryRecords",
				"cloudwatch:PutMetricData"},
			Resource: gocf.String("*"),
		},
	},
	VPC: []spartaIAM.PolicyStatement{
		{
			Action: []string{"ec2:CreateNetworkInterface",
				"ec2:DescribeNetworkInterfaces",
				"ec2:DeleteNetworkInterface"},
			Effect:   "Allow",
			Resource: wildcardArn,
		},
	},
	DynamoDB: []spartaIAM.PolicyStatement{
		{
			Effect: "Allow",
			Action: []string{"dynamodb:DescribeStream",
				"dynamodb:GetRecords",
				"dynamodb:GetShardIterator",
				"dynamodb:ListStreams",
			},
		},
	},
	Kinesis: []spartaIAM.PolicyStatement{
		{
			Effect: "Allow",
			Action: []string{"kinesis:GetRecords",
				"kinesis:GetShardIterator",
				"kinesis:DescribeStream",
				"kinesis:ListStreams",
			},
		},
	},
	// https://docs.aws.amazon.com/lambda/latest/dg/with-sqs-create-execution-role.html
	SQS: []spartaIAM.PolicyStatement{
		{
			Effect: "Allow",
			Action: []string{"SQS:GetQueueAttributes",
				"SQS:ChangeMessageVisibility",
				"SQS:DeleteMessage",
				"SQS:ReceiveMessage",
			},
		},
	},
}

// RE for sanitizing names
var reSanitize = regexp.MustCompile(`\W+`)

// Wildcard ARN for any AWS resource
var wildcardArn = gocf.String("*")

// AssumePolicyDocument defines common a IAM::Role PolicyDocument
// used as part of IAM::Role resource definitions
var AssumePolicyDocument = ArbitraryJSONObject{
	"Version": "2012-10-17",
	"Statement": []ArbitraryJSONObject{
		{
			"Effect": "Allow",
			"Principal": ArbitraryJSONObject{
				"Service": []string{LambdaPrincipal,
					EC2Principal,
					APIGatewayPrincipal},
			},
			"Action": []string{"sts:AssumeRole"},
		},
	},
}

////////////////////////////////////////////////////////////////////////////////
// Types
////////////////////////////////////////////////////////////////////////////////

// ArbitraryJSONObject represents an untyped key-value object. CloudFormation resource representations
// are aggregated as []ArbitraryJSONObject before being marsharled to JSON
// for API operations.
type ArbitraryJSONObject map[string]interface{}

// LambdaContext defines the AWS Lambda Context object provided by the AWS Lambda runtime.
// See http://docs.aws.amazon.com/lambda/latest/dg/nodejs-prog-model-context.html
// for more information on field values.  Note that the golang version doesn't functions
// defined on the Context object.
type LambdaContext struct {
	FunctionName       string `json:"functionName"`
	FunctionVersion    string `json:"functionVersion"`
	InvokedFunctionARN string `json:"invokedFunctionArn"`
	MemoryLimitInMB    string `json:"memoryLimitInMB"`
	AWSRequestID       string `json:"awsRequestId"`
	LogGroupName       string `json:"logGroupName"`
	LogStreamName      string `json:"logStreamName"`
}

// LambdaFunctionOptions defines additional AWS Lambda execution params.  See the
// AWS Lambda FunctionConfiguration (http://docs.aws.amazon.com/lambda/latest/dg/API_FunctionConfiguration.html)
// docs for more information. Note that the "Runtime" field will be automatically set
// to "nodejs4.3" (at least until golang is officially supported). See
// http://docs.aws.amazon.com/lambda/latest/dg/programming-model.html
type LambdaFunctionOptions struct {
	// Additional function description
	Description string
	// Memory limit
	MemorySize int64
	// Timeout (seconds)
	Timeout int64
	// VPC Settings
	VpcConfig *gocf.LambdaFunctionVPCConfig
	// Environment Variables
	Environment map[string]*gocf.StringExpr
	// KMS Key Arn used to encrypt environment variables
	KmsKeyArn string
	// The maximum of concurrent executions you want reserved for the function
	ReservedConcurrentExecutions int64
	// DeadLetterConfigArn is how Lambda handles events that it can't process.If
	// you don't specify a Dead Letter Queue (DLQ) configuration, Lambda
	// discards events after the maximum number of retries. For more information,
	// see Dead Letter Queues in the AWS Lambda Developer Guide.
	DeadLetterConfigArn gocf.Stringable
	// Tags to associate with the Lambda function
	Tags map[string]string
	// Tracing options for XRay
	TracingConfig *gocf.LambdaFunctionTracingConfig
	// Additional params
	SpartaOptions *SpartaOptions
}

func defaultLambdaFunctionOptions() *LambdaFunctionOptions {
	return &LambdaFunctionOptions{Description: "",
		MemorySize:                   128,
		Timeout:                      3,
		VpcConfig:                    nil,
		Environment:                  make(map[string]*gocf.StringExpr),
		KmsKeyArn:                    "",
		ReservedConcurrentExecutions: 0,
		SpartaOptions:                nil,
	}
}

// SpartaOptions allow the passing in of additional options during the creation of a Lambda Function
type SpartaOptions struct {
	// User supplied function name to use for
	// http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-function.html#cfn-lambda-function-functionname
	// value. If this is not supplied, a reflection-based
	// name will be automatically used.
	Name string
}

// WorkflowHooks is a structure that allows callers to customize the Sparta provisioning
// pipeline to add contents the Lambda archive or perform other workflow operations.
// TODO: remove single-valued fields
type WorkflowHooks struct {
	// Initial hook context. May be empty
	Context map[string]interface{}
	// PreBuild is called before the current Sparta-binary is compiled
	PreBuild WorkflowHook
	// PreBuilds are called before the current Sparta-binary is compiled
	PreBuilds []WorkflowHookHandler
	// PostBuild is called after the current Sparta-binary is compiled
	PostBuild WorkflowHook
	// PostBuilds are called after the current Sparta-binary is compiled
	PostBuilds []WorkflowHookHandler
	// ArchiveHook is called after Sparta has populated the ZIP archive containing the
	// AWS Lambda code package and before the ZIP writer is closed.  Define this hook
	// to add additional resource files to your Lambda package
	Archive ArchiveHook
	// ArchiveHook is called after Sparta has populated the ZIP archive containing the
	// AWS Lambda code package and before the ZIP writer is closed.  Define this hook
	// to add additional resource files to your Lambda package
	Archives []ArchiveHookHandler
	// PreMarshall is called before Sparta marshalls the application contents to a CloudFormation template
	PreMarshall WorkflowHook
	// PreMarshalls are called before Sparta marshalls the application contents into a CloudFormation
	// template
	PreMarshalls []WorkflowHookHandler
	// ServiceDecorator is called before Sparta marshalls the CloudFormation template
	ServiceDecorator ServiceDecoratorHook
	// ServiceDecorators are called before Sparta marshalls the CloudFormation template
	ServiceDecorators []ServiceDecoratorHookHandler
	// PostMarshall is called after Sparta marshalls the application contents to a CloudFormation template
	PostMarshall WorkflowHook
	// PostMarshalls are called after Sparta marshalls the application contents to a CloudFormation
	// template
	PostMarshalls []WorkflowHookHandler

	// Validators are hooks that are called when all marshalling
	// is complete. Each hook receives a complete read-only
	// copy of the materialized template.
	Validators []ServiceValidationHookHandler

	// Rollback is called if there is an error performing the requested operation
	Rollback RollbackHook
	// Rollbacks are called if there is an error performing the requested operation
	Rollbacks []RollbackHookHandler
}

////////////////////////////////////////////////////////////////////////////////
// START - IAMRolePrivilege
//

// IAMRolePrivilege struct stores data necessary to create an IAM Policy Document
// as part of the inline IAM::Role resource definition.  See
// http://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies.html
// for more information
// Deprecated: Prefer github.com/aws/iam/PolicyStatement instead.
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
	Resource interface{} `json:",omitempty"`
	// Service that requires the action
	Principal interface{} `json:",omitempty"`
	// Optional condition for the privilege
	Condition interface{} `json:",omitempty"`
}

func (rolePrivilege *IAMRolePrivilege) resourceExpr() *gocf.StringExpr {
	switch typedPrivilege := rolePrivilege.Resource.(type) {
	case string:
		return gocf.String(typedPrivilege)
	case gocf.RefFunc:
		return typedPrivilege.String()
	default:
		return typedPrivilege.(*gocf.StringExpr)
	}
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

func (roleDefinition *IAMRoleDefinition) toResource(eventSourceMappings []*EventSourceMapping,
	options *LambdaFunctionOptions,
	logger *logrus.Logger) gocf.IAMRole {

	statements := CommonIAMStatements.Core
	for _, eachPrivilege := range roleDefinition.Privileges {
		policyStatement := spartaIAM.PolicyStatement{
			Effect:   "Allow",
			Action:   eachPrivilege.Actions,
			Resource: eachPrivilege.resourceExpr(),
		}
		statements = append(statements, policyStatement)
	}

	// Add VPC permissions iff needed
	if options != nil && options.VpcConfig != nil {
		statements = append(statements, CommonIAMStatements.VPC...)
	}
	// In the past Sparta used to attach EventSourceMapping policies here.
	// However, moving everything to dynamic references means that we can't
	// fully populate the PolicyDocument statement slice until all of
	// the dynamically provisioned resources are defined. So that logic has
	// been moved to annotateMaterializedTemplate and annotateEventSourceMappings
	// which is run as the final step right before the template is marshaled
	// for creation.

	// http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html
	iamPolicies := gocf.IAMRolePolicyList{}
	iamPolicies = append(iamPolicies, gocf.IAMRolePolicy{
		PolicyDocument: ArbitraryJSONObject{
			"Version":   "2012-10-17",
			"Statement": statements,
		},
		PolicyName: gocf.String("LambdaPolicy"),
	})
	return gocf.IAMRole{
		AssumeRolePolicyDocument: AssumePolicyDocument,
		Policies:                 &iamPolicies,
	}
}

// Returns the stable logical name for this IAMRoleDefinition, which depends on the serviceName
// and owning targetLambdaFnName.  This potentially creates semantically equivalent IAM::Role entries
// from the same struct pointer, so:
// TODO: Create a canonical IAMRoleDefinition serialization that can be used as the digest source
func (roleDefinition *IAMRoleDefinition) logicalName(serviceName string, targetLambdaFnName string) string {
	if roleDefinition.cachedLogicalName == "" {
		roleDefinition.cachedLogicalName = CloudFormationResourceName("IAMRole", serviceName, targetLambdaFnName)
	}
	return roleDefinition.cachedLogicalName
}

//
// END - IAMRolePrivilege
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// START - EventSourceMapping

// EventSourceMapping specifies data necessary for pull-based configuration. The fields
// directly correspond to the golang AWS SDK's CreateEventSourceMappingInput
// (http://docs.aws.amazon.com/sdk-for-go/api/service/lambda.html#type-CreateEventSourceMappingInput)
type EventSourceMapping struct {
	StartingPosition               string
	EventSourceArn                 interface{}
	Disabled                       bool
	BatchSize                      int64
	BisectBatchOnFunctionError     bool
	DestinationConfig              *gocf.LambdaEventSourceMappingDestinationConfig
	MaximumBatchingWindowInSeconds int64
	MaximumRecordAgeInSeconds      int64
	MaximumRetryAttempts           int64
	ParallelizationFactor          int64
}

func (mapping *EventSourceMapping) export(serviceName string,
	targetLambdaName string,
	targetLambdaArn *gocf.StringExpr,
	S3Bucket string,
	S3Key string,
	template *gocf.Template,
	logger *logrus.Logger) error {

	dynamicArn := spartaCF.DynamicValueToStringExpr(mapping.EventSourceArn)
	eventSourceMappingResource := gocf.LambdaEventSourceMapping{
		StartingPosition:               marshalString(mapping.StartingPosition),
		EventSourceArn:                 dynamicArn.String(),
		FunctionName:                   targetLambdaArn,
		BatchSize:                      gocf.Integer(mapping.BatchSize),
		Enabled:                        gocf.Bool(!mapping.Disabled),
		BisectBatchOnFunctionError:     gocf.Bool(mapping.BisectBatchOnFunctionError),
		DestinationConfig:              mapping.DestinationConfig,
		MaximumBatchingWindowInSeconds: marshalInt(mapping.MaximumBatchingWindowInSeconds),
		MaximumRecordAgeInSeconds:      marshalInt(mapping.MaximumRecordAgeInSeconds),
		MaximumRetryAttempts:           marshalInt(mapping.MaximumRetryAttempts),
		ParallelizationFactor:          marshalInt(mapping.ParallelizationFactor),
	}

	// Unique components for the hash for the EventSource mapping
	// resource name
	hashParts := []string{
		targetLambdaName,
		dynamicArn.String().Literal,
		targetLambdaArn.Literal,
		fmt.Sprintf("%d", mapping.BatchSize),
		mapping.StartingPosition,
	}
	hash := sha1.New()
	for _, eachHashPart := range hashParts {
		_, writeErr := hash.Write([]byte(eachHashPart))
		if writeErr != nil {
			return errors.Wrapf(writeErr,
				"Failed to update EventSourceMapping name: %s", eachHashPart)
		}
	}
	resourceName := fmt.Sprintf("LambdaES%s", hex.EncodeToString(hash.Sum(nil)))
	template.AddResource(resourceName, eventSourceMappingResource)
	return nil
}

//
// END - EventSourceMapping
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// START - customResourceInfo

// customResourceInfo wraps up information about any userDefined CloudFormation
// user-defined Resources
type customResourceInfo struct {
	roleDefinition   *IAMRoleDefinition
	roleName         string
	handlerSymbol    interface{}
	userFunctionName string
	options          *LambdaFunctionOptions
	properties       map[string]interface{}
}

// Returns the stable CloudFormation resource logical name for this resource.  For
// a CustomResource, this name corresponds to the AWS::CloudFormation::CustomResource
// invocation of the Lambda function, not the lambda function itself
func (resourceInfo *customResourceInfo) logicalName() string {
	hash := sha1.New()
	// The name has to be stable so that the ServiceToken value which is
	// part the CustomResource invocation doesn't change during stack updates. CF
	// will throw an error if the ServiceToken changes across updates.
	source := fmt.Sprintf("%#v", resourceInfo.userFunctionName)
	_, writeErr := hash.Write([]byte(source))
	if writeErr != nil {
		fmt.Printf("TODO: failed to update hash. Error: %s", writeErr)
	}
	return CloudFormationResourceName(resourceInfo.userFunctionName,
		hex.EncodeToString(hash.Sum(nil)))
}

func (resourceInfo *customResourceInfo) export(serviceName string,
	targetLambda *gocf.StringExpr,
	S3Bucket string,
	S3Key string,
	roleNameMap map[string]*gocf.StringExpr,
	template *gocf.Template,
	logger *logrus.Logger) error {

	// Is this valid
	invalidErr := ensureValidSignature(resourceInfo.userFunctionName,
		resourceInfo.handlerSymbol)
	if invalidErr != nil {
		return invalidErr
	}

	// Figure out the role name
	iamRoleArnName := resourceInfo.roleName

	// If there is no user supplied role, that means that the associated
	// IAMRoleDefinition name has been created and this resource needs to
	// depend on that being created.
	if iamRoleArnName == "" && resourceInfo.roleDefinition != nil {
		iamRoleArnName = resourceInfo.roleDefinition.logicalName(serviceName,
			resourceInfo.userFunctionName)
	}
	lambdaDescription := resourceInfo.options.Description
	if lambdaDescription == "" {
		lambdaDescription = fmt.Sprintf("%s CustomResource: %s",
			serviceName,
			resourceInfo.userFunctionName)
	}

	// Create the Lambda Function
	lambdaFunctionName := awsLambdaFunctionName(resourceInfo.userFunctionName)

	lambdaEnv, lambdaEnvErr := lambdaFunctionEnvironment(nil,
		resourceInfo.userFunctionName,
		nil,
		logger)
	if lambdaEnvErr != nil {
		return errors.Wrapf(lambdaEnvErr, "Failed to create environment resource for custom info")
	}

	lambdaResource := gocf.LambdaFunction{
		Code: &gocf.LambdaFunctionCode{
			S3Bucket: gocf.String(S3Bucket),
			S3Key:    gocf.String(S3Key),
		},
		FunctionName: lambdaFunctionName.String(),
		Description:  gocf.String(lambdaDescription),
		Handler:      gocf.String(SpartaBinaryName),
		MemorySize:   gocf.Integer(resourceInfo.options.MemorySize),
		Role:         roleNameMap[iamRoleArnName],
		Runtime:      gocf.String(GoLambdaVersion),
		Timeout:      gocf.Integer(resourceInfo.options.Timeout),
		VPCConfig:    resourceInfo.options.VpcConfig,
		// DISPATCH INFORMATION
		Environment: lambdaEnv,
	}

	lambdaFunctionCFName := CloudFormationResourceName("CustomResourceLambda",
		resourceInfo.userFunctionName,
		resourceInfo.logicalName())

	cfResource := template.AddResource(lambdaFunctionCFName, lambdaResource)
	safeMetadataInsert(cfResource, "golangFunc", resourceInfo.userFunctionName)

	// And create the CustomResource that actually invokes it...
	newResource, newResourceError := newCloudFormationResource(cloudFormationLambda, logger)
	if nil != newResourceError {
		return newResourceError
	}
	customResource := newResource.(*cloudFormationLambdaCustomResource)
	customResource.ServiceToken = gocf.GetAtt(lambdaFunctionCFName, "Arn")
	customResource.UserProperties = resourceInfo.properties
	template.AddResource(resourceInfo.logicalName(), customResource)
	return nil
}

// END - customResourceInfo
////////////////////////////////////////////////////////////////////////////////

// Interceptor is the type of an event interceptor that taps the event lifecycle
type Interceptor func(ctx context.Context, msg json.RawMessage) context.Context

// NamedInterceptor represents a named interceptor that's invoked in the event path
type NamedInterceptor struct {
	Name        string
	Interceptor Interceptor
}

// InterceptorList is a list of NamedInterceptors
type InterceptorList []*NamedInterceptor

////////////////////////////////////////////////////////////////////////////////
// START - LambdaEventInterceptors

// LambdaEventInterceptors is the struct that stores event handlers that tap into
// the normal event dispatching workflow
type LambdaEventInterceptors struct {
	Begin          InterceptorList
	BeforeSetup    InterceptorList
	AfterSetup     InterceptorList
	BeforeDispatch InterceptorList
	AfterDispatch  InterceptorList
	Complete       InterceptorList
}

// Register is a convenience function to register a struct that
// implements the LambdaInterceptorProvider interface
func (lei *LambdaEventInterceptors) Register(provider LambdaInterceptorProvider) *LambdaEventInterceptors {
	namedInterceptor := func(interceptor Interceptor) *NamedInterceptor {
		return &NamedInterceptor{
			Name:        fmt.Sprintf("%T", provider),
			Interceptor: interceptor,
		}
	}
	if lei.Begin == nil {
		lei.Begin = make(InterceptorList, 0)
	}
	lei.Begin = append(lei.Begin, namedInterceptor(provider.Begin))

	if lei.BeforeSetup == nil {
		lei.BeforeSetup = make(InterceptorList, 0)
	}
	lei.BeforeSetup = append(lei.BeforeSetup, namedInterceptor(provider.BeforeSetup))

	if lei.AfterSetup == nil {
		lei.AfterSetup = make(InterceptorList, 0)
	}
	lei.AfterSetup = append(lei.AfterSetup, namedInterceptor(provider.AfterSetup))

	if lei.BeforeDispatch == nil {
		lei.BeforeDispatch = make(InterceptorList, 0)
	}
	lei.BeforeDispatch = append(lei.BeforeDispatch, namedInterceptor(provider.BeforeDispatch))

	if lei.AfterDispatch == nil {
		lei.AfterDispatch = make(InterceptorList, 0)
	}
	lei.AfterDispatch = append(lei.AfterDispatch, namedInterceptor(provider.AfterDispatch))

	if lei.Complete == nil {
		lei.Complete = make(InterceptorList, 0)
	}
	lei.Complete = append(lei.Complete, namedInterceptor(provider.Complete))
	return lei
}

// LambdaInterceptorProvider is the interface that defines an event interceptor
// Interceptors are able to hook into the normal event processing pipeline
type LambdaInterceptorProvider interface {
	Begin(ctx context.Context, msg json.RawMessage) context.Context
	BeforeSetup(ctx context.Context, msg json.RawMessage) context.Context
	AfterSetup(ctx context.Context, msg json.RawMessage) context.Context
	BeforeDispatch(ctx context.Context, msg json.RawMessage) context.Context
	AfterDispatch(ctx context.Context, msg json.RawMessage) context.Context
	Complete(ctx context.Context, msg json.RawMessage) context.Context
}

////////////////////////////////////////////////////////////////////////////////
// START - LambdaAWSInfo

// LambdaAWSInfo stores all data necessary to provision a golang-based AWS Lambda function.
type LambdaAWSInfo struct {
	// AWS Go lambda compliant function
	handlerSymbol interface{}
	// pointer to lambda function
	//lambdaFn LambdaFunction
	// The user supplied internal name
	userSuppliedFunctionName string
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
	EventSourceMappings []*EventSourceMapping
	// Template decorators. If non empty, the decorators will be called,
	// in order, to annotate the template
	Decorators []TemplateDecoratorHandler
	// Template decorator. If defined, the decorator will be called to insert additional
	// resources on behalf of this lambda function
	Decorator TemplateDecorator
	// Optional array of infrastructure resource logical names, typically
	// defined by a TemplateDecorator, that this lambda depends on
	DependsOn []string

	// Lambda Layers
	// Ref: https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-function.html#cfn-lambda-function-layers
	Layers []gocf.Stringable

	// Slice of customResourceInfo pointers for any associated CloudFormation
	// CustomResources associated with this lambda
	customResources []*customResourceInfo
	// Cached lambda name s.t. we only compute it once
	cachedLambdaFunctionName string

	// deprecation notices
	deprecationNotices []string

	// interceptors
	Interceptors *LambdaEventInterceptors
}

// lambdaFunctionName returns the internal
// function name for lambda export binding
func (info *LambdaAWSInfo) lambdaFunctionName() string {
	if info.cachedLambdaFunctionName != "" {
		return info.cachedLambdaFunctionName
	}
	var lambdaFuncName string

	if info.Options != nil &&
		info.Options.SpartaOptions != nil &&
		info.Options.SpartaOptions.Name != "" {
		lambdaFuncName = info.Options.SpartaOptions.Name
	} else if info.userSuppliedFunctionName != "" {
		lambdaFuncName = info.userSuppliedFunctionName
	} else {
		// Using the default name, let's at least remove the
		// first prefix, since that's the SCM provider and
		// doesn't provide a lot of value...

		if info.handlerSymbol != nil {
			lambdaPtr := runtime.FuncForPC(reflect.ValueOf(info.handlerSymbol).Pointer())
			lambdaFuncName = lambdaPtr.Name()
		}

		// Split
		// cwd: /Users/mweagle/Documents/gopath/src/github.com/mweagle/SpartaHelloWorld
		// anonymous: github.com/mweagle/Sparta.(*StructHandler1).(github.com/mweagle/Sparta.handler)-fm
		//	RE==> var reSplit = regexp.MustCompile("[\\(\\)\\.\\*]+")
		// 	RESULT ==> Hello,[github com/mweagle/Sparta StructHandler1 github com/mweagle/Sparta handler -fm]
		// Same package: main.helloWorld
		// Other package, free function: github.com/mweagle/SpartaPython.HelloWorld

		// Grab the name of the function...
		structDefined := strings.Contains(lambdaFuncName, "(") || strings.Contains(lambdaFuncName, ")")
		otherPackage := strings.Contains(lambdaFuncName, "/")
		canonicalName := lambdaFuncName
		if structDefined {
			var reSplit = regexp.MustCompile(`[*\(\)\[\]]+`)
			// Function name:
			// github.com/mweagle/Sparta.(*StructHandler1).handler-fm
			parts := reSplit.Split(lambdaFuncName, -1)
			lastPart := parts[len(parts)-1]
			penultimatePart := lastPart
			if len(parts) > 1 {
				penultimatePart = parts[len(parts)-2]
			}
			intermediateName := fmt.Sprintf("%s-%s", penultimatePart, lastPart)
			reClean := regexp.MustCompile(`[\*\(\)]+`)
			canonicalName = reClean.ReplaceAllString(intermediateName, "")
		} else if otherPackage {
			parts := strings.Split(lambdaFuncName, "/")
			canonicalName = parts[len(parts)-1]
		}
		// Final sanitization
		// Issue: https://github.com/mweagle/Sparta/issues/63
		lambdaFuncName = sanitizedName(canonicalName)
	}
	// Cache it so we only do this once
	info.cachedLambdaFunctionName = lambdaFuncName
	return info.cachedLambdaFunctionName
}

// RequireCustomResource adds a Lambda-backed CustomResource entry to the CloudFormation
// template. This function will be made a dependency of the owning Lambda function.
// The returned string is the custom resource's CloudFormation logical resource
// name that can be used for `Fn:GetAtt` calls for metadata lookups
func (info *LambdaAWSInfo) RequireCustomResource(roleNameOrIAMRoleDefinition interface{},
	handlerSymbol interface{},
	lambdaOptions *LambdaFunctionOptions,
	resourceProps map[string]interface{}) (string, error) {
	if nil == handlerSymbol {
		return "", fmt.Errorf("RequireCustomResource userFunc must not be nil")
	}
	// Is it valid?
	// Get the function pointer for this...
	handlerType := reflect.TypeOf(handlerSymbol)
	if handlerType.Kind() != reflect.Func {
		return "", fmt.Errorf("CustomResourceHandler kind %s is not %s",
			handlerType.Kind(),
			reflect.Func)
	}

	if nil == lambdaOptions {
		lambdaOptions = defaultLambdaFunctionOptions()
	}
	funcPtr := runtime.FuncForPC(reflect.ValueOf(handlerSymbol).Pointer())
	resourceInfo := &customResourceInfo{
		handlerSymbol:    handlerSymbol,
		userFunctionName: funcPtr.Name(),
		options:          lambdaOptions,
		properties:       resourceProps,
	}
	switch v := roleNameOrIAMRoleDefinition.(type) {
	case string:
		resourceInfo.roleName = roleNameOrIAMRoleDefinition.(string)
	case IAMRoleDefinition:
		definition := roleNameOrIAMRoleDefinition.(IAMRoleDefinition)
		resourceInfo.roleDefinition = &definition
	default:
		panic(fmt.Sprintf("Unsupported IAM Role type: %s", v))
	}
	resourceInfo.options.Environment = make(map[string]*gocf.StringExpr)
	info.customResources = append(info.customResources, resourceInfo)
	info.DependsOn = append(info.DependsOn, resourceInfo.logicalName())
	return resourceInfo.logicalName(), nil
}

// LogicalResourceName returns the stable, content-addressable logical
// name for this LambdaAWSInfo value. This is the CloudFormation
// resource name
func (info *LambdaAWSInfo) LogicalResourceName() string {
	// Per http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/resources-section-structure.html,
	// we can only use alphanumeric, so we'll take the sanitized name and
	// remove all underscores
	// Prefer the user-supplied stable name to the internal one.
	baseName := info.lambdaFunctionName()
	resourceName := strings.Replace(sanitizedName(baseName), "_", "", -1)
	prefix := fmt.Sprintf("%sLambda", resourceName)
	return CloudFormationResourceName(prefix, info.lambdaFunctionName())
}

func (info *LambdaAWSInfo) applyDecorators(template *gocf.Template,
	lambdaResource gocf.LambdaFunction,
	cfResource *gocf.Resource,
	serviceName string,
	S3Bucket string,
	S3Key string,
	buildID string,
	context map[string]interface{},
	logger *logrus.Logger) error {

	decorators := info.Decorators
	if info.Decorator != nil {
		logger.Debug("Decorator found for Lambda: ", info.lambdaFunctionName())
		logger.Warn("DEPRECATED: Single `Decorator` field is superseded by `Decorators` slice")
		decorators = append(decorators, TemplateDecoratorHookFunc(info.Decorator))
	}

	for _, eachDecorator := range decorators {
		// Create an empty template so that we can track whether things
		// are overwritten
		metadataMap := make(map[string]interface{})
		decoratorProxyTemplate := gocf.NewTemplate()
		decoratorErr := eachDecorator.DecorateTemplate(serviceName,
			info.LogicalResourceName(),
			lambdaResource,
			metadataMap,
			S3Bucket,
			S3Key,
			buildID,
			decoratorProxyTemplate,
			context,
			logger)
		if decoratorErr != nil {
			// Can we get the name?
			decoratorName := fmt.Sprintf("%T", eachDecorator)
			errorValue := errors.Errorf("TemplateDecorator %s failed to apply. Error: %s",
				decoratorName,
				decoratorErr)
			return errorValue
		}
		// This data is marshalled into a DiscoveryInfo struct s.t. it can be
		// unmarshalled via sparta.Discover.  We're going to just stuff it into
		// it's own same named property
		if len(metadataMap) != 0 {
			safeMetadataInsert(cfResource, info.LogicalResourceName(), metadataMap)
		}
		// Append the custom resources
		safeMergeErrs := gocc.SafeMerge(decoratorProxyTemplate,
			template)
		if len(safeMergeErrs) != 0 {
			return errors.Errorf("Lambda (%s) decorator created conflicting resources: %v",
				info.lambdaFunctionName(),
				safeMergeErrs)
		}
	}
	return nil
}

// Marshal this object into 1 or more CloudFormation resource definitions that are accumulated
// in the resources map
func (info *LambdaAWSInfo) export(serviceName string,
	S3Bucket string,
	S3Key string,
	S3Version string,
	buildID string,
	roleNameMap map[string]*gocf.StringExpr,
	template *gocf.Template,
	context map[string]interface{},
	logger *logrus.Logger) error {

	// Let's make sure the handler has the proper signature...This is basically
	// copy-pasted from the SDK

	// If we have RoleName, then get the ARN, otherwise get the Ref
	var dependsOn []string
	if nil != info.DependsOn {
		dependsOn = append(dependsOn, info.DependsOn...)
	}

	iamRoleArnName := info.RoleName

	// If there is no user supplied role, that means that the associated
	// IAMRoleDefinition name has been created and this resource needs to
	// depend on that being created.
	if iamRoleArnName == "" && info.RoleDefinition != nil {
		iamRoleArnName = info.RoleDefinition.logicalName(serviceName, info.lambdaFunctionName())
		dependsOn = append(dependsOn, info.RoleDefinition.logicalName(serviceName, info.lambdaFunctionName()))
	}
	lambdaDescription := info.Options.Description
	if lambdaDescription == "" {
		lambdaDescription = fmt.Sprintf("%s: %s", serviceName, info.lambdaFunctionName())
	}

	// Create the primary resource
	lambdaResource := gocf.LambdaFunction{
		Code: &gocf.LambdaFunctionCode{
			S3Bucket: gocf.String(S3Bucket),
			S3Key:    gocf.String(S3Key),
		},
		Description: gocf.String(lambdaDescription),
		Handler:     gocf.String(SpartaBinaryName),
		MemorySize:  gocf.Integer(info.Options.MemorySize),
		Role:        roleNameMap[iamRoleArnName],
		Runtime:     gocf.String(GoLambdaVersion),
		Timeout:     gocf.Integer(info.Options.Timeout),
		VPCConfig:   info.Options.VpcConfig,
	}
	// Layers?
	if nil != info.Layers {
		lambdaResource.Layers = gocf.StringList(info.Layers...)
	}

	if S3Version != "" {
		lambdaResource.Code.S3ObjectVersion = gocf.String(S3Version)
	}
	if info.Options.ReservedConcurrentExecutions != 0 {
		lambdaResource.ReservedConcurrentExecutions = gocf.Integer(info.Options.ReservedConcurrentExecutions)
	}
	if info.Options.DeadLetterConfigArn != nil {
		lambdaResource.DeadLetterConfig = &gocf.LambdaFunctionDeadLetterConfig{
			TargetArn: info.Options.DeadLetterConfigArn.String(),
		}
	}
	if nil != info.Options.TracingConfig {
		lambdaResource.TracingConfig = info.Options.TracingConfig
	}
	if info.Options.KmsKeyArn != "" {
		lambdaResource.KmsKeyArn = gocf.String(info.Options.KmsKeyArn)
	}
	if nil != info.Options.Tags {
		tagList := gocf.TagList{}
		for eachKey, eachValue := range info.Options.Tags {
			tagList = append(tagList, gocf.Tag{
				Key:   gocf.String(eachKey),
				Value: gocf.String(eachValue),
			})
		}
		lambdaResource.Tags = &tagList
	}

	// DISPATCH INFORMATION
	// Make sure we set the environment variable that
	// tells us which function to actually execute in
	// execute_awsbinary.go
	if info.Options.Environment == nil {
		info.Options.Environment = make(map[string]*gocf.StringExpr)
	}
	info.Options.Environment[envVarLogLevel] =
		gocf.String(logger.Level.String())

	lambdaResource.Environment = &gocf.LambdaFunctionEnvironment{
		Variables: info.Options.Environment,
	}

	// This function name is set here to be the same
	// name that the dispatcher will look up in execute
	// using the same logic so that we can borrow the
	// `AWS_LAMBDA_FUNCTION_NAME` env var
	lambdaFunctionName := awsLambdaFunctionName(info.lambdaFunctionName())
	lambdaResource.FunctionName = lambdaFunctionName.String()

	cfResource := template.AddResource(info.LogicalResourceName(), lambdaResource)
	cfResource.DependsOn = append(cfResource.DependsOn, dependsOn...)
	safeMetadataInsert(cfResource, "golangFunc", info.lambdaFunctionName())

	// Create the lambda Ref in case we need a permission or event mapping
	functionAttr := gocf.GetAtt(info.LogicalResourceName(), "Arn")

	// Permissions
	for _, eachPermission := range info.Permissions {
		_, err := eachPermission.export(serviceName,
			info.lambdaFunctionName(),
			info.LogicalResourceName(),
			template,
			S3Bucket,
			S3Key,
			logger)
		if nil != err {
			return errors.Wrapf(err, "Failed to export lambda permission")
		}
	}

	// Event Source Mappings
	for _, eachEventSourceMapping := range info.EventSourceMappings {
		mappingErr := eachEventSourceMapping.export(serviceName,
			info.lambdaFunctionName(),
			functionAttr,
			S3Bucket,
			S3Key,
			template,
			logger)
		if nil != mappingErr {
			return mappingErr
		}
	}

	// CustomResource
	for _, eachCustomResource := range info.customResources {

		resourceErr := eachCustomResource.export(serviceName,
			functionAttr,
			S3Bucket,
			S3Key,
			roleNameMap,
			template,
			logger)
		if nil != resourceErr {
			return resourceErr
		}
	}

	decoratorErr := info.applyDecorators(template,
		lambdaResource,
		cfResource,
		serviceName,
		S3Bucket,
		S3Key,
		buildID,
		context,
		logger)

	if decoratorErr != nil {
		return decoratorErr
	}

	// Log any deprecation notices
	for _, eachDeprecation := range info.deprecationNotices {
		logger.Warn(eachDeprecation)
	}
	return nil
}

//
// END - LambdaAWSInfo
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
//
// BEGIN - Private
//

func validateSpartaPreconditions(lambdaAWSInfos []*LambdaAWSInfo,
	logger *logrus.Logger) error {

	var errorText []string
	collisionMemo := make(map[string]int)

	incrementCounter := func(keyName string) {
		_, exists := collisionMemo[keyName]
		if !exists {
			collisionMemo[keyName] = 1
		} else {
			collisionMemo[keyName] = collisionMemo[keyName] + 1
		}
	}
	// 0 - check for nil
	for eachIndex, eachLambda := range lambdaAWSInfos {
		if eachLambda == nil {
			errorText = append(errorText,
				fmt.Sprintf("Lambda at position %d is `nil`", eachIndex))
		}
	}
	// Semantic checks only iff lambdas are non-nil
	if len(errorText) == 0 {

		// 1 - check for invalid signatures
		for _, eachLambda := range lambdaAWSInfos {
			validationErr := ensureValidSignature(eachLambda.userSuppliedFunctionName,
				eachLambda.handlerSymbol)
			if validationErr != nil {
				errorText = append(errorText, validationErr.Error())
			}
		}

		// 2 - check for duplicate golang function references.
		for _, eachLambda := range lambdaAWSInfos {
			incrementCounter(eachLambda.lambdaFunctionName())
			for _, eachCustom := range eachLambda.customResources {
				incrementCounter(eachCustom.userFunctionName)
			}
		}
		// Duplicates?
		for eachLambdaName, eachCount := range collisionMemo {
			if eachCount > 1 {
				logger.WithFields(logrus.Fields{
					"CollisionCount": eachCount,
					"Name":           eachLambdaName,
				}).Error("NewAWSLambda")
				errorText = append(errorText,
					fmt.Sprintf("Multiple definitions of lambda: %s", eachLambdaName))
			}
		}
		logger.WithFields(logrus.Fields{
			"CollisionMap": collisionMemo,
		}).Debug("Lambda collision map")
	}
	if len(errorText) != 0 {
		return errors.New(strings.Join(errorText[:], "\n"))
	}

	return nil
}

// Sanitize the provided input by replacing illegal characters with underscores
func sanitizedName(input string) string {
	return reSanitize.ReplaceAllString(input, "_")
}

//
// END - Private
//
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// Public
////////////////////////////////////////////////////////////////////////////////

// CloudFormationResourceName returns a name suitable as a logical
// CloudFormation resource value.  See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/resources-section-structure.html
// for more information.  The `prefix` value should provide a hint as to the
// resource type (eg, `SNSConfigurator`, `ImageTranscoder`).  Note that the returned
// name is not content-addressable.
func CloudFormationResourceName(prefix string, parts ...string) string {
	return spartaCF.CloudFormationResourceName(prefix, parts...)
}

// LambdaName returns the Go-reflection discovered name for a given
// function
func LambdaName(handlerSymbol interface{}) string {
	funcPtr := runtime.FuncForPC(reflect.ValueOf(handlerSymbol).Pointer())
	return funcPtr.Name()
}

/*
Supported lambdaHandler signatures:

• func ()
• func () error
• func (TIn), error
• func () (TOut, error)
• func (context.Context) error
• func (context.Context, TIn) error
• func (context.Context) (TOut, error)
• func (context.Context, TIn) (TOut, error)
*/

// NewAWSLambda is the creation function that replaces HandleAWSLambda. It returns
// a *LambdaAWSInfo pointer to the struct representing the AWS lambda target. It's a
// go-friendly signature for creating a lambda function
func NewAWSLambda(functionName string,
	lambdaHandler interface{},
	roleNameOrIAMRoleDefinition interface{}) (*LambdaAWSInfo, error) {

	if functionName == "" {
		return nil, errors.Errorf("AWS Lambda function name must not be empty")
	}
	if lambdaHandler == nil {
		return nil, errors.Errorf("AWS Lambda function handler must not be nil")
	}

	lambda := &LambdaAWSInfo{
		userSuppliedFunctionName: functionName,
		handlerSymbol:            lambdaHandler,
		Options:                  defaultLambdaFunctionOptions(),
		Permissions:              make([]LambdaPermissionExporter, 0),
		EventSourceMappings:      make([]*EventSourceMapping, 0),
		deprecationNotices:       make([]string, 0),
	}

	switch v := roleNameOrIAMRoleDefinition.(type) {
	case string:
		lambda.RoleName = v
	case IAMRoleDefinition:
		definition := v
		lambda.RoleDefinition = &definition
	default:
		return nil, errors.Errorf("AWS Lambda function IAM role must not be empty")
	}
	return lambda, nil
}

// HandleAWSLambda is deprecated in favor of NewAWSLambda(...)
func HandleAWSLambda(functionName string,
	lambdaHandler interface{},
	roleNameOrIAMRoleDefinition interface{}) *LambdaAWSInfo {

	lambda, lambdaErr := NewAWSLambda(functionName, lambdaHandler, roleNameOrIAMRoleDefinition)
	if lambdaErr != nil {
		panic(lambdaErr)
	}
	lambda.deprecationNotices = append(lambda.deprecationNotices, "sparta.HandleAWSLambda is deprecated starting with v1.6.0. Prefer `sparta.NewAWSLambda(...) (*LambdaAWSInfo, error)`")
	return lambda
}

// IsExecutingInLambda is a utility function to return a boolean
// indicating whether the application is running in AWS Lambda.
// See the list of environment variables defined at:
// https://docs.aws.amazon.com/lambda/latest/dg/current-supported-versions.html
// for more information.
func IsExecutingInLambda() bool {
	return os.Getenv("LAMBDA_TASK_ROOT") != "" ||
		os.Getenv("AWS_EXECUTION_ENV") != ""
}
