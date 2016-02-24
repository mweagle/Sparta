package sparta

import (
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

	gocf "github.com/crewjam/go-cloudformation"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/voxelbrain/goptions"
)

// SpartaVersion defines the current Sparta release
const SpartaVersion = "0.5.1"

func init() {
	rand.Seed(time.Now().Unix())
}

// ArbitraryJSONObject represents an untyped key-value object. CloudFormation resource representations
// are aggregated as []ArbitraryJSONObject before being marsharled to JSON
// for API operations.
type ArbitraryJSONObject map[string]interface{}

// AWS Principal ARNs from http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html
// See also
// http://docs.aws.amazon.com/general/latest/gr/rande.html
// for region specific principal names
const (
	// @enum AWSPrincipal
	APIGatewayPrincipal = "apigateway.amazonaws.com"
	// @enum CloudWatchEvents
	CloudWatchEventsPrincipal = "events.amazonaws.com"
	// @enum CloudWatchLogs
	// CloudWatchLogsPrincipal principal is region specific
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

var wildcardArn = gocf.String("*")

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

type iamPolicyStatement struct {
	Effect   string
	Action   []string
	Resource *gocf.StringExpr
}

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
var CommonIAMStatements = map[string][]iamPolicyStatement{
	"core": []iamPolicyStatement{
		iamPolicyStatement{
			Action: []string{"logs:CreateLogGroup",
				"logs:CreateLogStream",
				"logs:PutLogEvents"},
			Effect: "Allow",
			Resource: gocf.Join("",
				gocf.String("arn:aws:logs:"),
				gocf.Ref("AWS::Region"),
				gocf.String(":"),
				gocf.Ref("AWS::AccountId"),
				gocf.String("*")),
		},
		iamPolicyStatement{
			Action:   []string{"cloudwatch:PutMetricData"},
			Effect:   "Allow",
			Resource: wildcardArn,
		},
		iamPolicyStatement{
			Effect: "Allow",
			Action: []string{"cloudformation:DescribeStacks",
				"cloudformation:DescribeStackResource"},
			Resource: gocf.Join("", cloudFormationThisStackArn...),
		},
	},
	"dynamodb": []iamPolicyStatement{
		iamPolicyStatement{
			Effect: "Allow",
			Action: []string{"dynamodb:DescribeStream",
				"dynamodb:GetRecords",
				"dynamodb:GetShardIterator",
				"dynamodb:ListStreams",
			},
		}},
	"kinesis": []iamPolicyStatement{
		iamPolicyStatement{
			Effect: "Allow",
			Action: []string{"kinesis:GetRecords",
				"kinesis:GetShardIterator",
				"kinesis:DescribeStream",
				"kinesis:ListStreams",
			},
		},
	},
}

// RE for sanitizing golang/JS layer
var reSanitize = regexp.MustCompile("[\\.\\-\\s]+")

// RE to ensure CloudFormation compatible resource names
// Issue: https://github.com/mweagle/Sparta/issues/8
// Ref: http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/resources-section-structure.html
var reCloudFormationInvalidChars = regexp.MustCompile("[^A-Za-z0-9]+")

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
	lambdaResource gocf.LambdaFunction,
	template *gocf.Template,
	logger *logrus.Logger) error

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

func (rolePrivilege *IAMRolePrivilege) resourceExpr() *gocf.StringExpr {
	switch rolePrivilege.Resource.(type) {
	case string:
		return gocf.String(rolePrivilege.Resource.(string))
	default:
		return rolePrivilege.Resource.(*gocf.StringExpr)
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
	logger *logrus.Logger) gocf.IAMRole {

	statements := CommonIAMStatements["core"]
	for _, eachPrivilege := range roleDefinition.Privileges {
		statements = append(statements, iamPolicyStatement{
			Effect:   "Allow",
			Action:   eachPrivilege.Actions,
			Resource: eachPrivilege.resourceExpr(),
		})
	}

	// http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html
	for _, eachEventSourceMapping := range eventSourceMappings {
		arnParts := strings.Split(eachEventSourceMapping.EventSourceArn, ":")
		// 3rd slot is service scope
		if len(arnParts) >= 2 {
			awsService := arnParts[2]
			logger.Debug("Looking up common IAM privileges for EventSource: ", awsService)
			serviceStatements, exists := CommonIAMStatements[awsService]
			if exists {
				statements = append(statements, serviceStatements...)
				statements[len(statements)-1].Resource = gocf.String(eachEventSourceMapping.EventSourceArn)
			}
		}
	}

	return gocf.IAMRole{
		AssumeRolePolicyDocument: AssumePolicyDocument,
		Policies: &gocf.IAMPoliciesList{
			gocf.IAMPolicies{
				PolicyDocument: ArbitraryJSONObject{
					"Version":   "2012-10-17",
					"Statement": statements,
				},
				PolicyName: gocf.String(CloudFormationResourceName("LambdaPolicy")),
			},
		},
	}
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

// EventSourceMapping specifies data necessary for pull-based configuration. The fields
// directly correspond to the golang AWS SDK's CreateEventSourceMappingInput
// (http://docs.aws.amazon.com/sdk-for-go/api/service/lambda.html#type-CreateEventSourceMappingInput)
type EventSourceMapping struct {
	StartingPosition string
	EventSourceArn   string
	Disabled         bool
	BatchSize        int64
}

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
	EventSourceMappings []*EventSourceMapping
	// Template decorator. If defined, the decorator will be called to insert additional
	// resources on behalf of this lambda function
	Decorator TemplateDecorator
	// Optional array of infrastructure resource logical names, typically
	// defined by a TemplateDecorator, that this lambda depends on
	DependsOn []string
}

// URLPath returns the URL path that can be used as an argument
// to NewLambdaRequest or NewAPIGatewayRequest
func (info *LambdaAWSInfo) URLPath() string {
	return info.lambdaFnName
}

// Returns a JavaScript compatible function name for the golang function name.  This
// value will be used as the URL path component for the HTTP proxying layer.
func (info *LambdaAWSInfo) jsHandlerName() string {
	return sanitizedName(info.lambdaFnName)
}

// Returns the stable logical name for this LambdaAWSInfo value
func (info *LambdaAWSInfo) logicalName() string {
	// Per http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/resources-section-structure.html,
	// we can only use alphanumeric, so we'll take the sanitized name and
	// remove all underscores
	resourceName := strings.Replace(sanitizedName(info.lambdaFnName), "_", "", -1)
	prefix := fmt.Sprintf("%sLambda", resourceName)
	return CloudFormationResourceName(prefix, info.lambdaFnName)
}

// Marshal this object into 1 or more CloudFormation resource definitions that are accumulated
// in the resources map
func (info *LambdaAWSInfo) export(serviceName string,
	S3Bucket string,
	S3Key string,
	roleNameMap map[string]*gocf.StringExpr,
	template *gocf.Template,
	logger *logrus.Logger) error {

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
		iamRoleArnName = info.RoleDefinition.logicalName()
		dependsOn = append(dependsOn, info.RoleDefinition.logicalName())
	}
	lambdaDescription := info.Options.Description
	if "" == lambdaDescription {
		lambdaDescription = fmt.Sprintf("%s: %s", serviceName, info.lambdaFnName)
	}

	// Create the primary resource
	lambdaResource := gocf.LambdaFunction{
		Code: &gocf.LambdaFunctionCode{
			S3Bucket: gocf.String(S3Bucket),
			S3Key:    gocf.String(S3Key),
		},
		Description: gocf.String(lambdaDescription),
		Handler:     gocf.String(fmt.Sprintf("index.%s", info.jsHandlerName())),
		MemorySize:  gocf.Integer(info.Options.MemorySize),
		Role:        roleNameMap[iamRoleArnName],
		Runtime:     gocf.String("nodejs"),
		Timeout:     gocf.Integer(info.Options.Timeout),
	}
	cfResource := template.AddResource(info.logicalName(), lambdaResource)
	cfResource.DependsOn = append(cfResource.DependsOn, dependsOn...)
	safeMetadataInsert(cfResource, "golangFunc", info.lambdaFnName)

	// Create the lambda Ref in case we need a permission or event mapping
	functionAttr := gocf.GetAtt(info.logicalName(), "Arn")

	// Permissions
	for _, eachPermission := range info.Permissions {
		_, err := eachPermission.export(serviceName,
			info.lambdaFnName,
			info.logicalName(),
			template,
			S3Bucket,
			S3Key,
			logger)
		if nil != err {
			return err
		}
	}

	// Event Source Mappings
	hash := sha1.New()
	for _, eachEventSourceMapping := range info.EventSourceMappings {
		eventSourceMappingResource := gocf.LambdaEventSourceMapping{
			EventSourceArn:   gocf.String(eachEventSourceMapping.EventSourceArn),
			FunctionName:     functionAttr,
			StartingPosition: gocf.String(eachEventSourceMapping.StartingPosition),
			BatchSize:        gocf.Integer(eachEventSourceMapping.BatchSize),
			Enabled:          gocf.Bool(!eachEventSourceMapping.Disabled),
		}

		hash.Write([]byte(eachEventSourceMapping.EventSourceArn))
		binary.Write(hash, binary.LittleEndian, eachEventSourceMapping.BatchSize)
		hash.Write([]byte(eachEventSourceMapping.StartingPosition))
		resourceName := fmt.Sprintf("LambdaES%s", hex.EncodeToString(hash.Sum(nil)))
		template.AddResource(resourceName, eventSourceMappingResource)
	}

	// Decorator
	if nil != info.Decorator {
		logger.Debug("Decorator found for Lambda: ", info.lambdaFnName)
		// Create an empty template so that we can track whether things
		// are overwritten
		decoratorProxyTemplate := gocf.NewTemplate()
		err := info.Decorator(info.logicalName(),
			lambdaResource,
			decoratorProxyTemplate,
			logger)
		if nil != err {
			return err
		}
		// Append the custom resources
		err = safeMergeTemplates(decoratorProxyTemplate, template, logger)
		if nil != err {
			return fmt.Errorf("Lambda (%s) decorator created conflicting resources", info.lambdaFnName)
		}
	}
	return nil
}

//
// END - LambdaAWSInfo
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// Private
//

func validateSpartaPreconditions(lambdaAWSInfos []*LambdaAWSInfo, logger *logrus.Logger) error {
	var errorText []string
	collisionMemo := make(map[string]string, 0)

	// 1 - check for duplicate golang function references.
	for _, eachLambda := range lambdaAWSInfos {
		testName := eachLambda.lambdaFnName
		if _, exists := collisionMemo[testName]; !exists {
			collisionMemo[testName] = testName
			// We'll always find our own lambda
			duplicateCount := 0
			for _, eachCheckLambda := range lambdaAWSInfos {
				if testName == eachCheckLambda.lambdaFnName {
					duplicateCount++
				}
			}
			// We'll always find our own lambda
			if duplicateCount > 1 {
				logger.WithFields(logrus.Fields{
					"CollisionCount": duplicateCount,
					"Name":           testName,
				}).Error("Detected Sparta lambda function associated with multiple LambdaAWSInfo structs")
				errorText = append(errorText, fmt.Sprintf("Multiple definitions of lambda: %s", testName))
			}
		}
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
		awsConfig.LogLevel = aws.LogLevel(aws.LogDebugWithRequestErrors)
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

	logger.WithFields(logrus.Fields{
		"Name":    aws.SDKName,
		"Version": aws.SDKVersion,
	}).Debug("AWS SDK Info")

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
	resourceName := fmt.Sprintf("%s%s", prefix, hex.EncodeToString(hash.Sum(nil)))

	// Ensure that any non alphanumeric characters are replaced with ""
	return reCloudFormationInvalidChars.ReplaceAllString(resourceName, "x")
}

////////////////////////////////////////////////////////////////////////////////
// Public
////////////////////////////////////////////////////////////////////////////////

// NewLambda returns a LambdaAWSInfo value that can be provisioned via CloudFormation. The
// roleNameOrIAMRoleDefinition must either be a `string` or `IAMRoleDefinition`
// type
func NewLambda(roleNameOrIAMRoleDefinition interface{},
	fn LambdaFunction,
	lambdaOptions *LambdaFunctionOptions) *LambdaAWSInfo {
	if nil == lambdaOptions {
		lambdaOptions = &LambdaFunctionOptions{"", 128, 3}
	}
	lambdaPtr := runtime.FuncForPC(reflect.ValueOf(fn).Pointer())
	lambda := &LambdaAWSInfo{
		lambdaFnName:        lambdaPtr.Name(),
		lambdaFn:            fn,
		Options:             lambdaOptions,
		Permissions:         make([]LambdaPermissionExporter, 0),
		EventSourceMappings: make([]*EventSourceMapping, 0),
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
	if "execute" != options.Verb {
		logger.Info(strings.Repeat("-", 80))
	}
	err = validateSpartaPreconditions(lambdaAWSInfos, logger)
	if err != nil {
		return err
	}
	switch options.Verb {
	case "provision":
		err = Provision(options.Noop, serviceName, serviceDescription, lambdaAWSInfos, api, site, options.Provision.S3Bucket, nil, logger)
	case "execute":
		initializeDiscovery(serviceName, lambdaAWSInfos, logger)
		err = Execute(lambdaAWSInfos, options.Execute.Port, options.Execute.SignalParentPID, logger)
	case "delete":
		err = Delete(serviceName, logger)
	case "explore":
		err = Explore(lambdaAWSInfos, options.Explore.Port, logger)
	case "describe":
		fileWriter, err := os.Create(options.Describe.OutputFile)
		if err != nil {
			return fmt.Errorf("Failed to open %s output. Error: %s", options.Describe.OutputFile, err)
		}
		defer fileWriter.Close()
		err = Describe(serviceName, serviceDescription, lambdaAWSInfos, api, site, fileWriter, logger)
		if err != nil {
			err = fileWriter.Sync()
		}
	default:
		goptions.PrintHelp()
		err = fmt.Errorf("Unsupported subcommand: %s", string(options.Verb))
	}
	if nil != err {
		logger.Error(err)
	}
	return err
}
