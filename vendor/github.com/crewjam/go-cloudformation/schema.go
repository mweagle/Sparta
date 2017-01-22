package cloudformation

import "time"
import "encoding/json"

// CustomResourceProvider allows extend the NewResourceByType factory method
// with their own resource types.
type CustomResourceProvider func(customResourceType string) ResourceProperties

var customResourceProviders []CustomResourceProvider

// Register a CustomResourceProvider with go-cloudformation. Multiple
// providers may be registered. The first provider that returns a non-nil
// interface will be used and there is no check for a uniquely registered
// resource type.
func RegisterCustomResourceProvider(provider CustomResourceProvider) {
	customResourceProviders = append(customResourceProviders, provider)
}

// ApiGatewayAccount represents AWS::ApiGateway::Account
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-apigateway-account.html
type ApiGatewayAccount struct {
	// The Amazon Resource Name (ARN) of an IAM role that has write access to
	// CloudWatch Logs in your account.
	CloudWatchRoleArn *StringExpr `json:"CloudWatchRoleArn,omitempty"`
}

// CfnResourceType returns AWS::ApiGateway::Account to implement the ResourceProperties interface
func (s ApiGatewayAccount) CfnResourceType() string {
	return "AWS::ApiGateway::Account"
}

// ApiGatewayApiKey represents AWS::ApiGateway::ApiKey
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-apigateway-apikey.html
type ApiGatewayApiKey struct {
	// A description of the purpose of the API key.
	Description *StringExpr `json:"Description,omitempty"`

	// Indicates whether the API key can be used by clients.
	Enabled *BoolExpr `json:"Enabled,omitempty"`

	// A name for the API key. If you don't specify a name, AWS
	// CloudFormation generates a unique physical ID and uses that ID for the
	// API key name. For more information, see Name Type.
	Name *StringExpr `json:"Name,omitempty"`

	// A list of stages to associated with this API key.
	StageKeys *APIGatewayApiKeyStageKeyList `json:"StageKeys,omitempty"`
}

// CfnResourceType returns AWS::ApiGateway::ApiKey to implement the ResourceProperties interface
func (s ApiGatewayApiKey) CfnResourceType() string {
	return "AWS::ApiGateway::ApiKey"
}

// ApiGatewayAuthorizer represents AWS::ApiGateway::Authorizer
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-apigateway-authorizer.html
type ApiGatewayAuthorizer struct {
	// The credentials required for the authorizer. To specify an AWS
	// Identity and Access Management (IAM) role that API Gateway assumes,
	// specify the role's Amazon Resource Name (ARN). To use resource-based
	// permissions on the AWS Lambda (Lambda) function, specify null.
	AuthorizerCredentials *StringExpr `json:"AuthorizerCredentials,omitempty"`

	// The time-to-live (TTL) period, in seconds, that specifies how long API
	// Gateway caches authorizer results. If you specify a value greater than
	// 0, API Gateway caches the authorizer responses. By default, API
	// Gateway sets this property to 300. The maximum value is 3600, or 1
	// hour.
	AuthorizerResultTtlInSeconds *IntegerExpr `json:"AuthorizerResultTtlInSeconds,omitempty"`

	// The authorizer's Uniform Resource Identifier (URI). If you specify
	// TOKEN for the authorizer's Type property, specify a Lambda function
	// URI, which has the form arn:aws:apigateway:region:lambda:path/path.
	// The path usually has the form
	// /2015-03-31/functions/LambdaFunctionARN/invocations.
	AuthorizerUri *StringExpr `json:"AuthorizerUri,omitempty"`

	// The source of the identity in an incoming request. If you specify
	// TOKEN for the authorizer's Type property, specify a mapping
	// expression. The custom header mapping expression has the form
	// method.request.header.name, where name is the name of a custom
	// authorization header that clients submit as part of their requests.
	IdentitySource *StringExpr `json:"IdentitySource,omitempty"`

	// A validation expression for the incoming identity. If you specify
	// TOKEN for the authorizer's Type property, specify a regular
	// expression. API Gateway uses the expression to attempt to match the
	// incoming client token, and proceeds if the token matches. If the token
	// doesn't match, API Gateway responds with a 401 (unauthorized request)
	// error code.
	IdentityValidationExpression *StringExpr `json:"IdentityValidationExpression,omitempty"`

	// The name of the authorizer.
	Name *StringExpr `json:"Name,omitempty"`

	// A list of the Amazon Cognito user pool Amazon Resource Names (ARNs) to
	// associate with this authorizer. For more information, see Use Amazon
	// Cognito Your User Pool in the API Gateway Developer Guide.
	ProviderARNs *StringListExpr `json:"ProviderARNs,omitempty"`

	// The ID of the RestApi resource in which API Gateway creates the
	// authorizer.
	RestApiId *StringExpr `json:"RestApiId,omitempty"`

	// The type of authorizer:
	Type *StringExpr `json:"Type,omitempty"`
}

// CfnResourceType returns AWS::ApiGateway::Authorizer to implement the ResourceProperties interface
func (s ApiGatewayAuthorizer) CfnResourceType() string {
	return "AWS::ApiGateway::Authorizer"
}

// ApiGatewayBasePathMapping represents AWS::ApiGateway::BasePathMapping
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-apigateway-basepathmapping.html
type ApiGatewayBasePathMapping struct {
	// The base path name that callers of the API must provide in the URL
	// after the domain name.
	BasePath *StringExpr `json:"BasePath,omitempty"`

	// The name of a DomainName resource.
	DomainName *StringExpr `json:"DomainName,omitempty"`

	// The name of the API.
	RestApiId *StringExpr `json:"RestApiId,omitempty"`

	// The name of the API's stage.
	Stage *StringExpr `json:"Stage,omitempty"`
}

// CfnResourceType returns AWS::ApiGateway::BasePathMapping to implement the ResourceProperties interface
func (s ApiGatewayBasePathMapping) CfnResourceType() string {
	return "AWS::ApiGateway::BasePathMapping"
}

// ApiGatewayClientCertificate represents AWS::ApiGateway::ClientCertificate
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-apigateway-clientcertificate.html
type ApiGatewayClientCertificate struct {
	// A description of the client certificate.
	Description *StringExpr `json:"Description,omitempty"`
}

// CfnResourceType returns AWS::ApiGateway::ClientCertificate to implement the ResourceProperties interface
func (s ApiGatewayClientCertificate) CfnResourceType() string {
	return "AWS::ApiGateway::ClientCertificate"
}

// ApiGatewayDeployment represents AWS::ApiGateway::Deployment
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-apigateway-deployment.html
type ApiGatewayDeployment struct {
	// A description of the purpose of the API Gateway deployment.
	Description *StringExpr `json:"Description,omitempty"`

	// The ID of the RestApi resource to deploy.
	RestApiId *StringExpr `json:"RestApiId,omitempty"`

	// Configures the stage that API Gateway creates with this deployment.
	StageDescription *APIGatewayDeploymentStageDescription `json:"StageDescription,omitempty"`

	// A name for the stage that API Gateway creates with this deployment.
	// Use only alphanumeric characters.
	StageName *StringExpr `json:"StageName,omitempty"`
}

// CfnResourceType returns AWS::ApiGateway::Deployment to implement the ResourceProperties interface
func (s ApiGatewayDeployment) CfnResourceType() string {
	return "AWS::ApiGateway::Deployment"
}

// ApiGatewayMethod represents AWS::ApiGateway::Method
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-apigateway-method.html
type ApiGatewayMethod struct {
	// Indicates whether the method requires clients to submit a valid API
	// key.
	ApiKeyRequired *BoolExpr `json:"ApiKeyRequired,omitempty"`

	// The method's authorization type.
	AuthorizationType *StringExpr `json:"AuthorizationType,omitempty"`

	// The identifier of the authorizer to use on this method. If you specify
	// this property, specify CUSTOM for the AuthorizationType property.
	AuthorizerId *StringExpr `json:"AuthorizerId,omitempty"`

	// The HTTP method that clients will use to call this method.
	HttpMethod *StringExpr `json:"HttpMethod,omitempty"`

	// The back-end system that the method calls when it receives a request.
	Integration *APIGatewayMethodIntegration `json:"Integration,omitempty"`

	// The responses that can be sent to the client who calls the method.
	MethodResponses *APIGatewayMethodMethodResponseList `json:"MethodResponses,omitempty"`

	// The resources used for the response's content type. Specify response
	// models as key-value pairs (string-to-string map), with a content type
	// as the key and a Model resource name as the value.
	RequestModels interface{} `json:"RequestModels,omitempty"`

	// Request parameters that API Gateway accepts. Specify request
	// parameters as key-value pairs (string-to-Boolean map), with a source
	// as the key and a Boolean as the value. The Boolean specifies whether a
	// parameter is required. A source must match the following format
	// method.request.location.name, where the location is querystring, path,
	// or header, and name is a valid, unique parameter name.
	RequestParameters interface{} `json:"RequestParameters,omitempty"`

	// The ID of an API Gateway resource. For root resource methods, specify
	// the RestApi root resource ID, such as { "Fn::GetAtt": ["MyRestApi",
	// "RootResourceId"] }.
	ResourceId *StringExpr `json:"ResourceId,omitempty"`

	// The ID of the RestApi resource in which API Gateway creates the
	// method.
	RestApiId *StringExpr `json:"RestApiId,omitempty"`
}

// CfnResourceType returns AWS::ApiGateway::Method to implement the ResourceProperties interface
func (s ApiGatewayMethod) CfnResourceType() string {
	return "AWS::ApiGateway::Method"
}

// ApiGatewayModel represents AWS::ApiGateway::Model
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-apigateway-model.html
type ApiGatewayModel struct {
	// The content type for the model.
	ContentType *StringExpr `json:"ContentType,omitempty"`

	// A description that identifies this model.
	Description *StringExpr `json:"Description,omitempty"`

	// A name for the mode. If you don't specify a name, AWS CloudFormation
	// generates a unique physical ID and uses that ID for the model name.
	// For more information, see Name Type.
	Name *StringExpr `json:"Name,omitempty"`

	// The ID of a REST API with which to associate this model.
	RestApiId *StringExpr `json:"RestApiId,omitempty"`

	// The schema to use to transform data to one or more output formats.
	Schema interface{} `json:"Schema,omitempty"`
}

// CfnResourceType returns AWS::ApiGateway::Model to implement the ResourceProperties interface
func (s ApiGatewayModel) CfnResourceType() string {
	return "AWS::ApiGateway::Model"
}

// ApiGatewayResource represents AWS::ApiGateway::Resource
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-apigateway-resource.html
type ApiGatewayResource struct {
	// If you want to create a child resource, the ID of the parent resource.
	// For resources without a parent, specify the RestApi root resource ID,
	// such as { "Fn::GetAtt": ["MyRestApi", "RootResourceId"] }.
	ParentId *StringExpr `json:"ParentId,omitempty"`

	// A path name for the resource.
	PathPart *StringExpr `json:"PathPart,omitempty"`

	// The ID of the RestApi resource in which you want to create this
	// resource.
	RestApiId *StringExpr `json:"RestApiId,omitempty"`
}

// CfnResourceType returns AWS::ApiGateway::Resource to implement the ResourceProperties interface
func (s ApiGatewayResource) CfnResourceType() string {
	return "AWS::ApiGateway::Resource"
}

// ApiGatewayRestApi represents AWS::ApiGateway::RestApi
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-apigateway-restapi.html
type ApiGatewayRestApi struct {
	// A Swagger specification that defines a set of RESTful APIs in the JSON
	// format.
	Body interface{} `json:"Body,omitempty"`

	// The Amazon Simple Storage Service (Amazon S3) location that points to
	// a Swagger file, which defines a set of RESTful APIs in JSON or YAML
	// format.
	BodyS3Location *APIGatewayRestApiS3Location `json:"BodyS3Location,omitempty"`

	// The ID of the API Gateway RestApi resource that you want to clone.
	CloneFrom *StringExpr `json:"CloneFrom,omitempty"`

	// A description of the purpose of this API Gateway RestApi resource.
	Description *StringExpr `json:"Description,omitempty"`

	// If a warning occurs while API Gateway is creating the RestApi
	// resource, indicates whether to roll back the resource.
	FailOnWarnings *BoolExpr `json:"FailOnWarnings,omitempty"`

	// A name for the API Gateway RestApi resource.
	Name *StringExpr `json:"Name,omitempty"`

	// Custom header parameters for the request.
	Parameters *StringListExpr `json:"Parameters,omitempty"`
}

// CfnResourceType returns AWS::ApiGateway::RestApi to implement the ResourceProperties interface
func (s ApiGatewayRestApi) CfnResourceType() string {
	return "AWS::ApiGateway::RestApi"
}

// ApiGatewayStage represents AWS::ApiGateway::Stage
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-apigateway-stage.html
type ApiGatewayStage struct {
	// Indicates whether cache clustering is enabled for the stage.
	CacheClusterEnabled *BoolExpr `json:"CacheClusterEnabled,omitempty"`

	// The stage's cache cluster size.
	CacheClusterSize *StringExpr `json:"CacheClusterSize,omitempty"`

	// The identifier of the client certificate that API Gateway uses to call
	// your integration endpoints in the stage.
	ClientCertificateId *StringExpr `json:"ClientCertificateId,omitempty"`

	// The ID of the deployment that the stage points to.
	DeploymentId *StringExpr `json:"DeploymentId,omitempty"`

	// A description of the stage's purpose.
	Description *StringExpr `json:"Description,omitempty"`

	// Settings for all methods in the stage.
	MethodSettings *APIGatewayStageMethodSettingList `json:"MethodSettings,omitempty"`

	// The ID of the RestApi resource that you're deploying with this stage.
	RestApiId *StringExpr `json:"RestApiId,omitempty"`

	// The name of the stage, which API Gateway uses as the first path
	// segment in the invoke Uniform Resource Identifier (URI).
	StageName *StringExpr `json:"StageName,omitempty"`

	// A map (string to string map) that defines the stage variables, where
	// the variable name is the key and the variable value is the value.
	// Variable names are limited to alphanumeric characters. Values must
	// match the following regular expression: [A-Za-z0-9-._~:/?#&amp;=,]+.
	Variables interface{} `json:"Variables,omitempty"`
}

// CfnResourceType returns AWS::ApiGateway::Stage to implement the ResourceProperties interface
func (s ApiGatewayStage) CfnResourceType() string {
	return "AWS::ApiGateway::Stage"
}

// ApiGatewayUsagePlan represents AWS::ApiGateway::UsagePlan
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-apigateway-usageplan.html
type ApiGatewayUsagePlan struct {
	// The APIs and API stages to associate with this usage plan.
	ApiStages *APIGatewayUsagePlanApiStageList `json:"ApiStages,omitempty"`

	// The purpose of this usage plan.
	Description *StringExpr `json:"Description,omitempty"`

	// Configures the number of requests that users can make within a given
	// interval.
	Quota *APIGatewayUsagePlanQuotaSettings `json:"Quota,omitempty"`

	// Configures the overall request rate (average requests per second) and
	// burst capacity.
	Throttle *APIGatewayUsagePlanThrottleSettings `json:"Throttle,omitempty"`

	// A name for this usage plan.
	UsagePlanName *StringExpr `json:"UsagePlanName,omitempty"`
}

// CfnResourceType returns AWS::ApiGateway::UsagePlan to implement the ResourceProperties interface
func (s ApiGatewayUsagePlan) CfnResourceType() string {
	return "AWS::ApiGateway::UsagePlan"
}

// ApplicationAutoScalingScalableTarget represents AWS::ApplicationAutoScaling::ScalableTarget
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-applicationautoscaling-scalabletarget.html
type ApplicationAutoScalingScalableTarget struct {
	// The maximum value that Application Auto Scaling can use to scale a
	// target during a scaling activity.
	MaxCapacity *IntegerExpr `json:"MaxCapacity,omitempty"`

	// The minimum value that Application Auto Scaling can use to scale a
	// target during a scaling activity.
	MinCapacity *IntegerExpr `json:"MinCapacity,omitempty"`

	// The unique resource identifier to associate with this scalable target.
	// For more information, see the ResourceId parameter for the
	// RegisterScalableTarget action in the Application Auto Scaling API
	// Reference.
	ResourceId *StringExpr `json:"ResourceId,omitempty"`

	// The Amazon Resource Name (ARN) of an AWS Identity and Access
	// Management (IAM) role that allows Application Auto Scaling to modify
	// your scalable target.
	RoleARN *StringExpr `json:"RoleARN,omitempty"`

	// The scalable dimension associated with the scalable target. Specify
	// the service namespace, resource type, and scaling property, such as
	// ecs:service:DesiredCount for the desired task count of an Amazon EC2
	// Container Service service. For valid values, see the ScalableDimension
	// content for the ScalingPolicy data type in the Application Auto
	// Scaling API Reference.
	ScalableDimension *StringExpr `json:"ScalableDimension,omitempty"`

	// The AWS service namespace of the scalable target. For a list of
	// service namespaces, see AWS Service Namespaces in the AWS General
	// Reference.
	ServiceNamespace *StringExpr `json:"ServiceNamespace,omitempty"`
}

// CfnResourceType returns AWS::ApplicationAutoScaling::ScalableTarget to implement the ResourceProperties interface
func (s ApplicationAutoScalingScalableTarget) CfnResourceType() string {
	return "AWS::ApplicationAutoScaling::ScalableTarget"
}

// ApplicationAutoScalingScalingPolicy represents AWS::ApplicationAutoScaling::ScalingPolicy
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-applicationautoscaling-scalingpolicy.html
type ApplicationAutoScalingScalingPolicy struct {
	// A name for the scaling policy.
	PolicyName *StringExpr `json:"PolicyName,omitempty"`

	// An Application Auto Scaling policy type. For valid values, see the
	// PolicyType parameter for the PutScalingPolicy action in the
	// Application Auto Scaling API Reference.
	PolicyType *StringExpr `json:"PolicyType,omitempty"`

	// The unique resource identifier for the scalable target that this
	// scaling policy applies to. For more information, see the ResourceId
	// parameter for the PutScalingPolicy action in the Application Auto
	// Scaling API Reference.
	ResourceId *StringExpr `json:"ResourceId,omitempty"`

	// The scalable dimension of the scalable target that this scaling policy
	// applies to. The scalable dimension contains the service namespace,
	// resource type, and scaling property, such as ecs:service:DesiredCount
	// for the desired task count of an Amazon ECS service.
	ScalableDimension *StringExpr `json:"ScalableDimension,omitempty"`

	// The AWS service namespace of the scalable target that this scaling
	// policy applies to. For a list of service namespaces, see AWS Service
	// Namespaces in the AWS General Reference.
	ServiceNamespace *StringExpr `json:"ServiceNamespace,omitempty"`

	// The AWS CloudFormation-generated ID of an Application Auto Scaling
	// scalable target. For more information about the ID, see the Return
	// Value section of the AWS::ApplicationAutoScaling::ScalableTarget
	// resource.
	ScalingTargetId *StringExpr `json:"ScalingTargetId,omitempty"`

	// A step policy that configures when Application Auto Scaling scales
	// resources up or down, and by how much.
	StepScalingPolicyConfiguration *ApplicationAutoScalingScalingPolicyStepScalingPolicyConfiguration `json:"StepScalingPolicyConfiguration,omitempty"`
}

// CfnResourceType returns AWS::ApplicationAutoScaling::ScalingPolicy to implement the ResourceProperties interface
func (s ApplicationAutoScalingScalingPolicy) CfnResourceType() string {
	return "AWS::ApplicationAutoScaling::ScalingPolicy"
}

// AutoScalingAutoScalingGroup represents AWS::AutoScaling::AutoScalingGroup
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-as-group.html
type AutoScalingAutoScalingGroup struct {
	// Contains a list of availability zones for the group.
	AvailabilityZones *StringListExpr `json:"AvailabilityZones,omitempty"`

	// The number of seconds after a scaling activity is completed before any
	// further scaling activities can start.
	Cooldown *StringExpr `json:"Cooldown,omitempty"`

	// Specifies the desired capacity for the Auto Scaling group.
	DesiredCapacity *StringExpr `json:"DesiredCapacity,omitempty"`

	// The length of time in seconds after a new EC2 instance comes into
	// service that Auto Scaling starts checking its health.
	HealthCheckGracePeriod *IntegerExpr `json:"HealthCheckGracePeriod,omitempty"`

	// The service you want the health status from, Amazon EC2 or Elastic
	// Load Balancer. Valid values are EC2 or ELB.
	HealthCheckType *StringExpr `json:"HealthCheckType,omitempty"`

	// The ID of the Amazon EC2 instance you want to use to create the Auto
	// Scaling group. Use this property if you want to create an Auto Scaling
	// group that uses an existing Amazon EC2 instance instead of a launch
	// configuration.
	InstanceId *StringExpr `json:"InstanceId,omitempty"`

	// Specifies the name of the associated
	// AWS::AutoScaling::LaunchConfiguration.
	LaunchConfigurationName *StringExpr `json:"LaunchConfigurationName,omitempty"`

	// A list of Classic load balancers associated with this Auto Scaling
	// group. To specify Application load balancers, use TargetGroupARNs.
	LoadBalancerNames *StringListExpr `json:"LoadBalancerNames,omitempty"`

	// The maximum size of the Auto Scaling group.
	MaxSize *StringExpr `json:"MaxSize,omitempty"`

	// Enables the monitoring of group metrics of an Auto Scaling group.
	MetricsCollection *AutoScalingMetricsCollectionList `json:"MetricsCollection,omitempty"`

	// The minimum size of the Auto Scaling group.
	MinSize *StringExpr `json:"MinSize,omitempty"`

	// An embedded property that configures an Auto Scaling group to send
	// notifications when specified events take place.
	NotificationConfigurations *AutoScalingNotificationConfigurationsList `json:"NotificationConfigurations,omitempty"`

	// The name of an existing cluster placement group into which you want to
	// launch your instances. A placement group is a logical grouping of
	// instances within a single Availability Zone. You cannot specify
	// multiple Availability Zones and a placement group.
	PlacementGroup *StringExpr `json:"PlacementGroup,omitempty"`

	// The tags you want to attach to this resource.
	Tags *AutoScalingTagsList `json:"Tags,omitempty"`

	// A list of Amazon Resource Names (ARN) of target groups to associate
	// with the Auto Scaling group.
	TargetGroupARNs *StringListExpr `json:"TargetGroupARNs,omitempty"`

	// A policy or a list of policies that are used to select the instances
	// to terminate. The policies are executed in the order that you list
	// them.
	TerminationPolicies *StringListExpr `json:"TerminationPolicies,omitempty"`

	// A list of subnet identifiers of Amazon Virtual Private Cloud (Amazon
	// VPCs).
	VPCZoneIdentifier *StringListExpr `json:"VPCZoneIdentifier,omitempty"`
}

// CfnResourceType returns AWS::AutoScaling::AutoScalingGroup to implement the ResourceProperties interface
func (s AutoScalingAutoScalingGroup) CfnResourceType() string {
	return "AWS::AutoScaling::AutoScalingGroup"
}

// AutoScalingLaunchConfiguration represents AWS::AutoScaling::LaunchConfiguration
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-as-launchconfig.html
type AutoScalingLaunchConfiguration struct {
	// For Amazon EC2 instances in a VPC, indicates whether instances in the
	// Auto Scaling group receive public IP addresses. If you specify true,
	// each instance in the Auto Scaling receives a unique public IP address.
	AssociatePublicIpAddress *BoolExpr `json:"AssociatePublicIpAddress,omitempty"`

	// Specifies how block devices are exposed to the instance. You can
	// specify virtual devices and EBS volumes.
	BlockDeviceMappings *AutoScalingBlockDeviceMappingList `json:"BlockDeviceMappings,omitempty"`

	// The ID of a ClassicLink-enabled VPC to link your EC2-Classic instances
	// to. You can specify this property only for EC2-Classic instances. For
	// more information, see ClassicLink in the Amazon Elastic Compute Cloud
	// User Guide.
	ClassicLinkVPCId *StringExpr `json:"ClassicLinkVPCId,omitempty"`

	// The IDs of one or more security groups for the VPC that you specified
	// in the ClassicLinkVPCId property.
	ClassicLinkVPCSecurityGroups *StringListExpr `json:"ClassicLinkVPCSecurityGroups,omitempty"`

	// Specifies whether the launch configuration is optimized for EBS I/O.
	// This optimization provides dedicated throughput to Amazon EBS and an
	// optimized configuration stack to provide optimal EBS I/O performance.
	EbsOptimized *BoolExpr `json:"EbsOptimized,omitempty"`

	// Provides the name or the Amazon Resource Name (ARN) of the instance
	// profile associated with the IAM role for the instance. The instance
	// profile contains the IAM role.
	IamInstanceProfile *StringExpr `json:"IamInstanceProfile,omitempty"`

	// Provides the unique ID of the Amazon Machine Image (AMI) that was
	// assigned during registration.
	ImageId *StringExpr `json:"ImageId,omitempty"`

	// The ID of the Amazon EC2 instance you want to use to create the launch
	// configuration. Use this property if you want the launch configuration
	// to use settings from an existing Amazon EC2 instance.
	InstanceId *StringExpr `json:"InstanceId,omitempty"`

	// Indicates whether detailed instance monitoring is enabled for the Auto
	// Scaling group. By default, this property is set to true (enabled).
	InstanceMonitoring *BoolExpr `json:"InstanceMonitoring,omitempty"`

	// Specifies the instance type of the EC2 instance.
	InstanceType *StringExpr `json:"InstanceType,omitempty"`

	// Provides the ID of the kernel associated with the EC2 AMI.
	KernelId *StringExpr `json:"KernelId,omitempty"`

	// Provides the name of the EC2 key pair.
	KeyName *StringExpr `json:"KeyName,omitempty"`

	// The tenancy of the instance. An instance with a tenancy of dedicated
	// runs on single-tenant hardware and can only be launched in a VPC. You
	// must set the value of this parameter to dedicated if want to launch
	// dedicated instances in a shared tenancy VPC (a VPC with the instance
	// placement tenancy attribute set to default). For more information, see
	// CreateLaunchConfiguration in the Auto Scaling API Reference.
	PlacementTenancy *StringExpr `json:"PlacementTenancy,omitempty"`

	// The ID of the RAM disk to select. Some kernels require additional
	// drivers at launch. Check the kernel requirements for information about
	// whether you need to specify a RAM disk. To find kernel requirements,
	// refer to the AWS Resource Center and search for the kernel ID.
	RamDiskId *StringExpr `json:"RamDiskId,omitempty"`

	// A list that contains the EC2 security groups to assign to the Amazon
	// EC2 instances in the Auto Scaling group. The list can contain the name
	// of existing EC2 security groups or references to
	// AWS::EC2::SecurityGroup resources created in the template. If your
	// instances are launched within VPC, specify Amazon VPC security group
	// IDs.
	SecurityGroups interface{} `json:"SecurityGroups,omitempty"`

	// The spot price for this autoscaling group. If a spot price is set,
	// then the autoscaling group will launch when the current spot price is
	// less than the amount specified in the template.
	SpotPrice *StringExpr `json:"SpotPrice,omitempty"`

	// The user data available to the launched EC2 instances.
	UserData *StringExpr `json:"UserData,omitempty"`
}

// CfnResourceType returns AWS::AutoScaling::LaunchConfiguration to implement the ResourceProperties interface
func (s AutoScalingLaunchConfiguration) CfnResourceType() string {
	return "AWS::AutoScaling::LaunchConfiguration"
}

// AutoScalingLifecycleHook represents AWS::AutoScaling::LifecycleHook
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-as-lifecyclehook.html
type AutoScalingLifecycleHook struct {
	// The name of the Auto Scaling group for the lifecycle hook.
	AutoScalingGroupName *StringExpr `json:"AutoScalingGroupName,omitempty"`

	// The action the Auto Scaling group takes when the lifecycle hook
	// timeout elapses or if an unexpected failure occurs.
	DefaultResult *StringExpr `json:"DefaultResult,omitempty"`

	// The amount of time that can elapse before the lifecycle hook times
	// out. When the lifecycle hook times out, Auto Scaling performs the
	// action that you specified in the DefaultResult property.
	HeartbeatTimeout *IntegerExpr `json:"HeartbeatTimeout,omitempty"`

	// The state of the Amazon EC2 instance to which you want to attach the
	// lifecycle hook. For valid values, see the LifecycleTransition content
	// for the LifecycleHook data type in the Auto Scaling API Reference.
	LifecycleTransition *StringExpr `json:"LifecycleTransition,omitempty"`

	// Additional information that you want to include when Auto Scaling
	// sends a message to the notification target.
	NotificationMetadata *StringExpr `json:"NotificationMetadata,omitempty"`

	// The Amazon resource name (ARN) of the notification target that Auto
	// Scaling uses to notify you when an instance is in the transition state
	// for the lifecycle hook. You can specify an Amazon SQS queue or an
	// Amazon SNS topic. The notification message includes the following
	// information: lifecycle action token, user account ID, Auto Scaling
	// group name, lifecycle hook name, instance ID, lifecycle transition,
	// and notification metadata.
	NotificationTargetARN *StringExpr `json:"NotificationTargetARN,omitempty"`

	// The ARN of the IAM role that allows the Auto Scaling group to publish
	// to the specified notification target. The role requires permissions to
	// Amazon SNS and Amazon SQS.
	RoleARN *StringExpr `json:"RoleARN,omitempty"`
}

// CfnResourceType returns AWS::AutoScaling::LifecycleHook to implement the ResourceProperties interface
func (s AutoScalingLifecycleHook) CfnResourceType() string {
	return "AWS::AutoScaling::LifecycleHook"
}

// AutoScalingScalingPolicy represents AWS::AutoScaling::ScalingPolicy
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-as-policy.html
type AutoScalingScalingPolicy struct {
	// Specifies whether the ScalingAdjustment is an absolute number or a
	// percentage of the current capacity. Valid values are ChangeInCapacity,
	// ExactCapacity, and PercentChangeInCapacity.
	AdjustmentType *StringExpr `json:"AdjustmentType,omitempty"`

	// The name or Amazon Resource Name (ARN) of the Auto Scaling Group that
	// you want to attach the policy to.
	AutoScalingGroupName *StringExpr `json:"AutoScalingGroupName,omitempty"`

	// The amount of time, in seconds, after a scaling activity completes
	// before any further trigger-related scaling activities can start.
	Cooldown *StringExpr `json:"Cooldown,omitempty"`

	// The estimated time, in seconds, until a newly launched instance can
	// send metrics to CloudWatch. By default, Auto Scaling uses the cooldown
	// period, as specified in the Cooldown property.
	EstimatedInstanceWarmup *IntegerExpr `json:"EstimatedInstanceWarmup,omitempty"`

	// The aggregation type for the CloudWatch metrics. You can specify
	// Minimum, Maximum, or Average. By default, AWS CloudFormation specifies
	// Average.
	MetricAggregationType *StringExpr `json:"MetricAggregationType,omitempty"`

	// For the PercentChangeInCapacity adjustment type, the minimum number of
	// instances to scale. The scaling policy changes the desired capacity of
	// the Auto Scaling group by a minimum of this many instances. This
	// property replaces the MinAdjustmentStep property.
	MinAdjustmentMagnitude *IntegerExpr `json:"MinAdjustmentMagnitude,omitempty"`

	// An Auto Scaling policy type. You can specify SimpleScaling or
	// StepScaling. By default, AWS CloudFormation specifies SimpleScaling.
	// For more information, see Scaling Policy Types in the Auto Scaling
	// User Guide.
	PolicyType *StringExpr `json:"PolicyType,omitempty"`

	// The number of instances by which to scale. The AdjustmentType property
	// determines if AWS CloudFormation interprets this number as an absolute
	// number (when the ExactCapacity value is specified), increase or
	// decrease capacity by a specified number (when the ChangeInCapacity
	// value is specified), or increase or decrease capacity as a percentage
	// of the existing Auto Scaling group size (when the
	// PercentChangeInCapacity value is specified). A positive value adds to
	// the current capacity and a negative value subtracts from the current
	// capacity. For exact capacity, you must specify a positive value.
	ScalingAdjustment *IntegerExpr `json:"ScalingAdjustment,omitempty"`

	// A set of adjustments that enable you to scale based on the size of the
	// alarm breach.
	StepAdjustments *AutoScalingScalingPolicyStepAdjustmentsList `json:"StepAdjustments,omitempty"`
}

// CfnResourceType returns AWS::AutoScaling::ScalingPolicy to implement the ResourceProperties interface
func (s AutoScalingScalingPolicy) CfnResourceType() string {
	return "AWS::AutoScaling::ScalingPolicy"
}

// AutoScalingScheduledAction represents AWS::AutoScaling::ScheduledAction
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-as-scheduledaction.html
type AutoScalingScheduledAction struct {
	// The name or ARN of the Auto Scaling group.
	AutoScalingGroupName *StringExpr `json:"AutoScalingGroupName,omitempty"`

	// The number of Amazon EC2 instances that should be running in the Auto
	// Scaling group.
	DesiredCapacity *IntegerExpr `json:"DesiredCapacity,omitempty"`

	// The time in UTC for this schedule to end. For example,
	// 2010-06-01T00:00:00Z.
	EndTime time.Time `json:"EndTime,omitempty"`

	// The maximum number of Amazon EC2 instances in the Auto Scaling group.
	MaxSize *IntegerExpr `json:"MaxSize,omitempty"`

	// The minimum number of Amazon EC2 instances in the Auto Scaling group.
	MinSize *IntegerExpr `json:"MinSize,omitempty"`

	// The time in UTC when recurring future actions will start. You specify
	// the start time by following the Unix cron syntax format. For more
	// information about cron syntax, go to
	// http://en.wikipedia.org/wiki/Cron.
	Recurrence *StringExpr `json:"Recurrence,omitempty"`

	// The time in UTC for this schedule to start. For example,
	// 2010-06-01T00:00:00Z.
	StartTime time.Time `json:"StartTime,omitempty"`
}

// CfnResourceType returns AWS::AutoScaling::ScheduledAction to implement the ResourceProperties interface
func (s AutoScalingScheduledAction) CfnResourceType() string {
	return "AWS::AutoScaling::ScheduledAction"
}

// CertificateManagerCertificate represents AWS::CertificateManager::Certificate
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-certificatemanager-certificate.html
type CertificateManagerCertificate struct {
	// Fully qualified domain name (FQDN), such as www.example.com, of the
	// site that you want to secure with the ACM certificate. To protect
	// several sites in the same domain, use an asterisk (*) to specify a
	// wildcard. For example, *.example.com protects www.example.com,
	// site.example.com, and images.example.com.
	DomainName *StringExpr `json:"DomainName,omitempty"`

	// Domain information that domain name registrars use to verify your
	// identity. For more information and the default values, see Configure
	// Email for Your Domain and Validate Domain Ownership in the AWS
	// Certificate Manager User Guide.
	DomainValidationOptions *CertificateManagerCertificateDomainValidationOptionList `json:"DomainValidationOptions,omitempty"`

	// FQDNs to be included in the Subject Alternative Name extension of the
	// ACM certificate. For example, you can add www.example.net to a
	// certificate for the www.example.com domain name so that users can
	// reach your site by using either name.
	SubjectAlternativeNames *StringListExpr `json:"SubjectAlternativeNames,omitempty"`

	// An arbitrary set of tags (key–value pairs) for this ACM certificate.
	Tags []ResourceTag `json:"Tags,omitempty"`
}

// CfnResourceType returns AWS::CertificateManager::Certificate to implement the ResourceProperties interface
func (s CertificateManagerCertificate) CfnResourceType() string {
	return "AWS::CertificateManager::Certificate"
}

// CloudFormationAuthentication represents AWS::CloudFormation::Authentication
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-authentication.html
type CloudFormationAuthentication struct {
	// Specifies the access key ID for S3 authentication.
	AccessKeyId *StringExpr `json:"accessKeyId,omitempty"`

	// A comma-delimited list of Amazon S3 buckets to be associated with the
	// S3 authentication credentials.
	Buckets *StringListExpr `json:"buckets,omitempty"`

	// Specifies the password for basic authentication.
	Password *StringExpr `json:"password,omitempty"`

	// Specifies the secret key for S3 authentication.
	SecretKey *StringExpr `json:"secretKey,omitempty"`

	// Specifies whether the authentication scheme uses a user name and
	// password ("basic") or an access key ID and secret key ("S3").
	Type *StringExpr `json:"type,omitempty"`

	// A comma-delimited list of URIs to be associated with the basic
	// authentication credentials. The authorization applies to the specified
	// URIs and any more specific URI. For example, if you specify
	// http://www.example.com, the authorization will also apply to
	// http://www.example.com/test.
	Uris *StringListExpr `json:"uris,omitempty"`

	// Specifies the user name for basic authentication.
	Username *StringExpr `json:"username,omitempty"`

	// Describes the role for role-based authentication.
	RoleName *StringExpr `json:"roleName,omitempty"`
}

// CfnResourceType returns AWS::CloudFormation::Authentication to implement the ResourceProperties interface
func (s CloudFormationAuthentication) CfnResourceType() string {
	return "AWS::CloudFormation::Authentication"
}

// CloudFormationCustomResource represents AWS::CloudFormation::CustomResource
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-cfn-customresource.html
type CloudFormationCustomResource struct {
	// The service token that was given to the template developer by the
	// service provider to access the service, such as an Amazon SNS topic
	// ARN or Lambda function ARN. The service token must be from the same
	// region in which you are creating the stack.
	ServiceToken *StringExpr `json:"ServiceToken,omitempty"`
}

// CfnResourceType returns AWS::CloudFormation::CustomResource to implement the ResourceProperties interface
func (s CloudFormationCustomResource) CfnResourceType() string {
	return "AWS::CloudFormation::CustomResource"
}

// CloudFormationInit represents AWS::CloudFormation::Init
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-init.html
type CloudFormationInit struct {
}

// CfnResourceType returns AWS::CloudFormation::Init to implement the ResourceProperties interface
func (s CloudFormationInit) CfnResourceType() string {
	return "AWS::CloudFormation::Init"
}

// CloudFormationInterface represents AWS::CloudFormation::Interface
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-cloudformation-interface.html
type CloudFormationInterface struct {
	// A list of parameter group types, where you specify group names, the
	// parameters in each group, and the order in which the parameters are
	// shown.
	ParameterGroups *InterfaceParameterGroupList `json:"ParameterGroups,omitempty"`

	// A mapping of parameters and their friendly names that the AWS
	// CloudFormation console shows when a stack is created or updated.
	ParameterLabels *InterfaceParameterLabel `json:"ParameterLabels,omitempty"`
}

// CfnResourceType returns AWS::CloudFormation::Interface to implement the ResourceProperties interface
func (s CloudFormationInterface) CfnResourceType() string {
	return "AWS::CloudFormation::Interface"
}

// CloudFormationStack represents AWS::CloudFormation::Stack
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-stack.html
type CloudFormationStack struct {
	// A list of existing Amazon SNS topics where notifications about stack
	// events are sent.
	NotificationARNs *StringListExpr `json:"NotificationARNs,omitempty"`

	// The set of parameters passed to AWS CloudFormation when this nested
	// stack is created.
	Parameters *CloudFormationStackParameters `json:"Parameters,omitempty"`

	// An arbitrary set of tags (key–value pairs) to describe this stack.
	Tags []ResourceTag `json:"Tags,omitempty"`

	// The URL of a template that specifies the stack that you want to create
	// as a resource. The template must be stored on an Amazon S3 bucket, so
	// the URL must have the form:
	// https://s3.amazonaws.com/.../TemplateName.template
	TemplateURL *StringExpr `json:"TemplateURL,omitempty"`

	// The length of time, in minutes, that AWS CloudFormation waits for the
	// nested stack to reach the CREATE_COMPLETE state. The default is no
	// timeout. When AWS CloudFormation detects that the nested stack has
	// reached the CREATE_COMPLETE state, it marks the nested stack resource
	// as CREATE_COMPLETE in the parent stack and resumes creating the parent
	// stack. If the timeout period expires before the nested stack reaches
	// CREATE_COMPLETE, AWS CloudFormation marks the nested stack as failed
	// and rolls back both the nested stack and parent stack.
	TimeoutInMinutes *StringExpr `json:"TimeoutInMinutes,omitempty"`
}

// CfnResourceType returns AWS::CloudFormation::Stack to implement the ResourceProperties interface
func (s CloudFormationStack) CfnResourceType() string {
	return "AWS::CloudFormation::Stack"
}

// CloudFormationWaitCondition represents AWS::CloudFormation::WaitCondition
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-waitcondition.html
type CloudFormationWaitCondition struct {
	// The number of success signals that AWS CloudFormation must receive
	// before it continues the stack creation process. When the wait
	// condition receives the requisite number of success signals, AWS
	// CloudFormation resumes the creation of the stack. If the wait
	// condition does not receive the specified number of success signals
	// before the Timeout period expires, AWS CloudFormation assumes that the
	// wait condition has failed and rolls the stack back.
	Count *StringExpr `json:"Count,omitempty"`

	// A reference to the wait condition handle used to signal this wait
	// condition. Use the Ref intrinsic function to specify an
	// AWS::CloudFormation::WaitConditionHandle resource.
	Handle *StringExpr `json:"Handle,omitempty"`

	// The length of time (in seconds) to wait for the number of signals that
	// the Count property specifies. Timeout is a minimum-bound property,
	// meaning the timeout occurs no sooner than the time you specify, but
	// can occur shortly thereafter. The maximum time that can be specified
	// for this property is 12 hours (43200 seconds).
	Timeout *StringExpr `json:"Timeout,omitempty"`
}

// CfnResourceType returns AWS::CloudFormation::WaitCondition to implement the ResourceProperties interface
func (s CloudFormationWaitCondition) CfnResourceType() string {
	return "AWS::CloudFormation::WaitCondition"
}

// CloudFormationWaitConditionHandle represents AWS::CloudFormation::WaitConditionHandle
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-waitconditionhandle.html
type CloudFormationWaitConditionHandle struct {
}

// CfnResourceType returns AWS::CloudFormation::WaitConditionHandle to implement the ResourceProperties interface
func (s CloudFormationWaitConditionHandle) CfnResourceType() string {
	return "AWS::CloudFormation::WaitConditionHandle"
}

// CloudFrontDistribution represents AWS::CloudFront::Distribution
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cloudfront-distribution.html
type CloudFrontDistribution struct {
	// The distribution's configuration information.
	DistributionConfig *CloudFrontDistributionConfig `json:"DistributionConfig,omitempty"`
}

// CfnResourceType returns AWS::CloudFront::Distribution to implement the ResourceProperties interface
func (s CloudFrontDistribution) CfnResourceType() string {
	return "AWS::CloudFront::Distribution"
}

// CloudTrailTrail represents AWS::CloudTrail::Trail
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-cloudtrail-trail.html
type CloudTrailTrail struct {
	// The Amazon Resource Name (ARN) of a log group to which CloudTrail logs
	// will be delivered.
	CloudWatchLogsLogGroupArn *StringExpr `json:"CloudWatchLogsLogGroupArn,omitempty"`

	// The role ARN that Amazon CloudWatch Logs (CloudWatch Logs) assumes to
	// write logs to a log group. For more information, see Role Policy
	// Document for CloudTrail to Use CloudWatch Logs for Monitoring in the
	// AWS CloudTrail User Guide.
	CloudWatchLogsRoleArn *StringExpr `json:"CloudWatchLogsRoleArn,omitempty"`

	// Indicates whether CloudTrail validates the integrity of log files. By
	// default, AWS CloudFormation sets this value to false. When you disable
	// log file integrity validation, CloudTrail stops creating digest files.
	// For more information, see CreateTrail in the AWS CloudTrail API
	// Reference.
	EnableLogFileValidation *BoolExpr `json:"EnableLogFileValidation,omitempty"`

	// Indicates whether the trail is publishing events from global services,
	// such as IAM, to the log files. By default, AWS CloudFormation sets
	// this value to false.
	IncludeGlobalServiceEvents *BoolExpr `json:"IncludeGlobalServiceEvents,omitempty"`

	// Indicates whether the CloudTrail trail is currently logging AWS API
	// calls.
	IsLogging *BoolExpr `json:"IsLogging,omitempty"`

	// Indicates whether the CloudTrail trail is created in the region in
	// which you create the stack (false) or in all regions (true). By
	// default, AWS CloudFormation sets this value to false. For more
	// information, see How Does CloudTrail Behave Regionally and Globally?
	// in the AWS CloudTrail User Guide.
	IsMultiRegionTrail *BoolExpr `json:"IsMultiRegionTrail,omitempty"`

	// The AWS Key Management Service (AWS KMS) key ID that you want to use
	// to encrypt CloudTrail logs. You can specify an alias name (prefixed
	// with alias/), an alias ARN, a key ARN, or a globally unique
	// identifier.
	KMSKeyId *StringExpr `json:"KMSKeyId,omitempty"`

	// The name of the Amazon S3 bucket where CloudTrail publishes log files.
	S3BucketName *StringExpr `json:"S3BucketName,omitempty"`

	// An Amazon S3 object key prefix that precedes the name of all log
	// files.
	S3KeyPrefix *StringExpr `json:"S3KeyPrefix,omitempty"`

	// The name of an Amazon SNS topic that is notified when new log files
	// are published.
	SnsTopicName *StringExpr `json:"SnsTopicName,omitempty"`

	// An arbitrary set of tags (key–value pairs) for this trail.
	Tags []ResourceTag `json:"Tags,omitempty"`
}

// CfnResourceType returns AWS::CloudTrail::Trail to implement the ResourceProperties interface
func (s CloudTrailTrail) CfnResourceType() string {
	return "AWS::CloudTrail::Trail"
}

// CloudWatchAlarm represents AWS::CloudWatch::Alarm
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cw-alarm.html
type CloudWatchAlarm struct {
	// Indicates whether or not actions should be executed during any changes
	// to the alarm's state.
	ActionsEnabled *BoolExpr `json:"ActionsEnabled,omitempty"`

	// The list of actions to execute when this alarm transitions into an
	// ALARM state from any other state. Each action is specified as an
	// Amazon Resource Number (ARN). For more information about creating
	// alarms and the actions you can specify, see Creating Amazon CloudWatch
	// Alarms in the Amazon CloudWatch User Guide.
	AlarmActions *StringListExpr `json:"AlarmActions,omitempty"`

	// The description for the alarm.
	AlarmDescription *StringExpr `json:"AlarmDescription,omitempty"`

	// A name for the alarm. If you don't specify a name, AWS CloudFormation
	// generates a unique physical ID and uses that ID for the alarm name.
	// For more information, see Name Type.
	AlarmName *StringExpr `json:"AlarmName,omitempty"`

	// The arithmetic operation to use when comparing the specified Statistic
	// and Threshold. The specified Statistic value is used as the first
	// operand.
	ComparisonOperator *StringExpr `json:"ComparisonOperator,omitempty"`

	// The dimensions for the alarm's associated metric.
	Dimensions *CloudWatchMetricDimensionList `json:"Dimensions,omitempty"`

	// The number of periods over which data is compared to the specified
	// threshold.
	EvaluationPeriods *StringExpr `json:"EvaluationPeriods,omitempty"`

	// The list of actions to execute when this alarm transitions into an
	// INSUFFICIENT_DATA state from any other state. Each action is specified
	// as an Amazon Resource Number (ARN). Currently the only action
	// supported is publishing to an Amazon SNS topic or an Amazon Auto
	// Scaling policy.
	InsufficientDataActions *StringListExpr `json:"InsufficientDataActions,omitempty"`

	// The name for the alarm's associated metric. For more information about
	// the metrics that you can specify, see Amazon CloudWatch Namespaces,
	// Dimensions, and Metrics Reference in the Amazon CloudWatch User Guide.
	MetricName *StringExpr `json:"MetricName,omitempty"`

	// The namespace for the alarm's associated metric.
	Namespace *StringExpr `json:"Namespace,omitempty"`

	// The list of actions to execute when this alarm transitions into an OK
	// state from any other state. Each action is specified as an Amazon
	// Resource Number (ARN). Currently the only action supported is
	// publishing to an Amazon SNS topic or an Amazon Auto Scaling policy.
	OKActions *StringListExpr `json:"OKActions,omitempty"`

	// The time over which the specified statistic is applied. You must
	// specify a time in seconds that is also a multiple of 60.
	Period *StringExpr `json:"Period,omitempty"`

	// The statistic to apply to the alarm's associated metric.
	Statistic *StringExpr `json:"Statistic,omitempty"`

	// The value against which the specified statistic is compared.
	Threshold *StringExpr `json:"Threshold,omitempty"`

	// The unit for the alarm's associated metric.
	Unit *StringExpr `json:"Unit,omitempty"`
}

// CfnResourceType returns AWS::CloudWatch::Alarm to implement the ResourceProperties interface
func (s CloudWatchAlarm) CfnResourceType() string {
	return "AWS::CloudWatch::Alarm"
}

// CodeCommitRepository represents AWS::CodeCommit::Repository
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-codecommit-repository.html
type CodeCommitRepository struct {
	// A description about the AWS CodeCommit repository. For constraints,
	// see the CreateRepository action in the AWS CodeCommit API Reference.
	RepositoryDescription *StringExpr `json:"RepositoryDescription,omitempty"`

	// A name for the AWS CodeCommit repository.
	RepositoryName *StringExpr `json:"RepositoryName,omitempty"`

	// Defines the actions to take in response to events that occur in the
	// repository. For example, you can send email notifications when someone
	// pushes to the repository.
	Triggers *CodeCommitRepositoryTriggerList `json:"Triggers,omitempty"`
}

// CfnResourceType returns AWS::CodeCommit::Repository to implement the ResourceProperties interface
func (s CodeCommitRepository) CfnResourceType() string {
	return "AWS::CodeCommit::Repository"
}

// CodeDeployApplication represents AWS::CodeDeploy::Application
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-codedeploy-application.html
type CodeDeployApplication struct {
	// A name for the application. If you don't specify a name, AWS
	// CloudFormation generates a unique physical ID and uses that ID for the
	// application name. For more information, see Name Type.
	ApplicationName *StringExpr `json:"ApplicationName,omitempty"`
}

// CfnResourceType returns AWS::CodeDeploy::Application to implement the ResourceProperties interface
func (s CodeDeployApplication) CfnResourceType() string {
	return "AWS::CodeDeploy::Application"
}

// CodeDeployDeploymentConfig represents AWS::CodeDeploy::DeploymentConfig
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-codedeploy-deploymentconfig.html
type CodeDeployDeploymentConfig struct {
	// A name for the deployment configuration. If you don't specify a name,
	// AWS CloudFormation generates a unique physical ID and uses that ID for
	// the deployment configuration name. For more information, see Name
	// Type.
	DeploymentConfigName *StringExpr `json:"DeploymentConfigName,omitempty"`

	// The minimum number of healthy instances that must be available at any
	// time during an AWS CodeDeploy deployment. For example, for a fleet of
	// nine instances, if you specify a minimum of six healthy instances, AWS
	// CodeDeploy deploys your application up to three instances at a time so
	// that you always have six healthy instances. The deployment succeeds if
	// your application successfully deploys to six or more instances;
	// otherwise, the deployment fails.
	MinimumHealthyHosts *CodeDeployDeploymentConfigMinimumHealthyHosts `json:"MinimumHealthyHosts,omitempty"`
}

// CfnResourceType returns AWS::CodeDeploy::DeploymentConfig to implement the ResourceProperties interface
func (s CodeDeployDeploymentConfig) CfnResourceType() string {
	return "AWS::CodeDeploy::DeploymentConfig"
}

// CodeDeployDeploymentGroup represents AWS::CodeDeploy::DeploymentGroup
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-codedeploy-deploymentgroup.html
type CodeDeployDeploymentGroup struct {
	// The name of an AWS CodeDeploy application for this deployment group.
	ApplicationName *StringExpr `json:"ApplicationName,omitempty"`

	// A list of associated Auto Scaling groups that AWS CodeDeploy
	// automatically deploys revisions to when new instances are created.
	AutoScalingGroups *StringListExpr `json:"AutoScalingGroups,omitempty"`

	// The application revision that will be deployed to this deployment
	// group.
	Deployment *CodeDeployDeploymentGroupDeployment `json:"Deployment,omitempty"`

	// A deployment configuration name or a predefined configuration name.
	// With predefined configurations, you can deploy application revisions
	// to one instance at a time, half of the instances at a time, or all the
	// instances at once. For more information and valid values, see the
	// DeploymentConfigName parameter for the CreateDeploymentGroup action in
	// the AWS CodeDeploy API Reference.
	DeploymentConfigName *StringExpr `json:"DeploymentConfigName,omitempty"`

	// A name for the deployment group. If you don't specify a name, AWS
	// CloudFormation generates a unique physical ID and uses that ID for the
	// deployment group name. For more information, see Name Type.
	DeploymentGroupName *StringExpr `json:"DeploymentGroupName,omitempty"`

	// The Amazon EC2 tags to filter on. AWS CodeDeploy includes all
	// instances that match the tag filter with this deployment group.
	Ec2TagFilters *CodeDeployDeploymentGroupEc2TagFiltersList `json:"Ec2TagFilters,omitempty"`

	// The on-premises instance tags to filter on. AWS CodeDeploy includes
	// all on-premises instances that match the tag filter with this
	// deployment group. To register on-premises instances with AWS
	// CodeDeploy, see Configure Existing On-Premises Instances by Using AWS
	// CodeDeploy in the AWS CodeDeploy User Guide.
	OnPremisesInstanceTagFilters *CodeDeployDeploymentGroupOnPremisesInstanceTagFiltersList `json:"OnPremisesInstanceTagFilters,omitempty"`

	// A service role Amazon Resource Name (ARN) that grants AWS CodeDeploy
	// permission to make calls to AWS services on your behalf. For more
	// information, see Create a Service Role for AWS CodeDeploy in the AWS
	// CodeDeploy User Guide.
	ServiceRoleArn *StringExpr `json:"ServiceRoleArn,omitempty"`
}

// CfnResourceType returns AWS::CodeDeploy::DeploymentGroup to implement the ResourceProperties interface
func (s CodeDeployDeploymentGroup) CfnResourceType() string {
	return "AWS::CodeDeploy::DeploymentGroup"
}

// CodePipelineCustomAction represents AWS::CodePipeline::CustomActionType
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-codepipeline-customactiontype.html
type CodePipelineCustomAction struct {
	// The category of the custom action, such as a source action or a build
	// action. For valid values, see CreateCustomActionType in the AWS
	// CodePipeline API Reference.
	Category *StringExpr `json:"Category,omitempty"`

	// The configuration properties for the custom action.
	ConfigurationProperties *CodePipelineCustomActionTypeConfigurationPropertiesList `json:"ConfigurationProperties,omitempty"`

	// The input artifact details for this custom action.
	InputArtifactDetails *CodePipelineCustomActionTypeArtifactDetails `json:"InputArtifactDetails,omitempty"`

	// The output artifact details for this custom action.
	OutputArtifactDetails *CodePipelineCustomActionTypeArtifactDetails `json:"OutputArtifactDetails,omitempty"`

	// The name of the service provider that AWS CodePipeline uses for this
	// custom action.
	Provider *StringExpr `json:"Provider,omitempty"`

	// URLs that provide users information about this custom action.
	Settings *CodePipelineCustomActionTypeSettings `json:"Settings,omitempty"`

	// The version number of this custom action.
	Version *StringExpr `json:"Version,omitempty"`
}

// CfnResourceType returns AWS::CodePipeline::CustomActionType to implement the ResourceProperties interface
func (s CodePipelineCustomAction) CfnResourceType() string {
	return "AWS::CodePipeline::CustomActionType"
}

// CodePipelinePipeline represents AWS::CodePipeline::Pipeline
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-codepipeline-pipeline.html
type CodePipelinePipeline struct {
	// The Amazon Simple Storage Service (Amazon S3) location where AWS
	// CodePipeline stores pipeline artifacts. The S3 bucket must have
	// versioning enabled. For more information, see Create an Amazon S3
	// Bucket for Your Application in the AWS CodePipeline User Guide.
	ArtifactStore *CodePipelinePipelineArtifactStore `json:"ArtifactStore,omitempty"`

	// Prevents artifacts in a pipeline from transitioning to the stage that
	// you specified. This enables you to manually control transitions.
	DisableInboundStageTransitions *CodePipelinePipelineDisableInboundStageTransitionsList `json:"DisableInboundStageTransitions,omitempty"`

	// The name of your AWS CodePipeline pipeline.
	Name *StringExpr `json:"Name,omitempty"`

	// Indicates whether to rerun the AWS CodePipeline pipeline after you
	// update it.
	RestartExecutionOnUpdate *BoolExpr `json:"RestartExecutionOnUpdate,omitempty"`

	// A service role Amazon Resource Name (ARN) that grants AWS CodePipeline
	// permission to make calls to AWS services on your behalf. For more
	// information, see AWS CodePipeline Access Permissions Reference in the
	// AWS CodePipeline User Guide.
	RoleArn *StringExpr `json:"RoleArn,omitempty"`

	// Defines the AWS CodePipeline pipeline stages.
	Stages *CodePipelinePipelineStagesList `json:"Stages,omitempty"`
}

// CfnResourceType returns AWS::CodePipeline::Pipeline to implement the ResourceProperties interface
func (s CodePipelinePipeline) CfnResourceType() string {
	return "AWS::CodePipeline::Pipeline"
}

// ConfigConfigRule represents AWS::Config::ConfigRule
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-config-configrule.html
type ConfigConfigRule struct {
	// A name for the AWS Config rule. If you don't specify a name, AWS
	// CloudFormation generates a unique physical ID and uses that ID for the
	// rule name. For more information, see Name Type.
	ConfigRuleName *StringExpr `json:"ConfigRuleName,omitempty"`

	// A description about this AWS Config rule.
	Description *StringExpr `json:"Description,omitempty"`

	// Input parameter values that are passed to the AWS Config rule (Lambda
	// function).
	InputParameters interface{} `json:"InputParameters,omitempty"`

	// The maximum frequency at which the AWS Config rule runs evaluations.
	// For valid values, see the ConfigRule data type in the AWS Config API
	// Reference.
	MaximumExecutionFrequency *StringExpr `json:"MaximumExecutionFrequency,omitempty"`

	// Defines which AWS resources will trigger an evaluation when their
	// configurations change. The scope can include one or more resource
	// types, a combination of a tag key and value, or a combination of one
	// resource type and one resource ID. Specify a scope to constrain the
	// resources that are evaluated. If you don't specify a scope, the rule
	// evaluates all resources in the recording group.
	Scope *ConfigConfigRuleScope `json:"Scope,omitempty"`

	// Specifies the rule owner, the rule identifier, and the events that
	// cause the function to evaluate your AWS resources.
	Source *ConfigConfigRuleSource `json:"Source,omitempty"`
}

// CfnResourceType returns AWS::Config::ConfigRule to implement the ResourceProperties interface
func (s ConfigConfigRule) CfnResourceType() string {
	return "AWS::Config::ConfigRule"
}

// ConfigConfigurationRecorder represents AWS::Config::ConfigurationRecorder
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-config-configurationrecorder.html
type ConfigConfigurationRecorder struct {
	// A name for the configuration recorder. If you don't specify a name,
	// AWS CloudFormation generates a unique physical ID and uses that ID for
	// the configuration recorder name. For more information, see Name Type.
	Name *StringExpr `json:"Name,omitempty"`

	// Indicates whether to record configurations for all supported resources
	// or for a list of resource types. The resource types that you list must
	// be supported by AWS Config.
	RecordingGroup *ConfigConfigurationRecorderRecordingGroup `json:"RecordingGroup,omitempty"`

	// The Amazon Resource Name (ARN) of the AWS Identity and Access
	// Management (IAM) role that is used to make read or write requests to
	// the delivery channel that you specify and to get configuration details
	// for supported AWS resources. For more information, see Permissions for
	// the AWS Config IAM Role in the AWS Config Developer Guide.
	RoleARN *StringExpr `json:"RoleARN,omitempty"`
}

// CfnResourceType returns AWS::Config::ConfigurationRecorder to implement the ResourceProperties interface
func (s ConfigConfigurationRecorder) CfnResourceType() string {
	return "AWS::Config::ConfigurationRecorder"
}

// ConfigDeliveryChannel represents AWS::Config::DeliveryChannel
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-config-deliverychannel.html
type ConfigDeliveryChannel struct {
	// Provides options for how AWS Config delivers configuration snapshots
	// to the S3 bucket in your delivery channel.
	ConfigSnapshotDeliveryProperties *ConfigDeliveryChannelConfigSnapshotDeliveryProperties `json:"ConfigSnapshotDeliveryProperties,omitempty"`

	// A name for the delivery channel. If you don't specify a name, AWS
	// CloudFormation generates a unique physical ID and uses that ID for the
	// delivery channel name. For more information, see Name Type.
	Name *StringExpr `json:"Name,omitempty"`

	// The name of an S3 bucket where you want to store configuration history
	// for the delivery channel.
	S3BucketName *StringExpr `json:"S3BucketName,omitempty"`

	// A key prefix (folder) for the specified S3 bucket.
	S3KeyPrefix *StringExpr `json:"S3KeyPrefix,omitempty"`

	// The Amazon Resource Name (ARN) of the Amazon Simple Notification
	// Service (Amazon SNS) topic that AWS Config delivers notifications to.
	SnsTopicARN *StringExpr `json:"SnsTopicARN,omitempty"`
}

// CfnResourceType returns AWS::Config::DeliveryChannel to implement the ResourceProperties interface
func (s ConfigDeliveryChannel) CfnResourceType() string {
	return "AWS::Config::DeliveryChannel"
}

// DataPipelinePipeline represents AWS::DataPipeline::Pipeline
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-datapipeline-pipeline.html
type DataPipelinePipeline struct {
	// Indicates whether to validate and start the pipeline or stop an active
	// pipeline. By default, the value is set to true.
	Activate *BoolExpr `json:"Activate,omitempty"`

	// A description for the pipeline.
	Description *StringExpr `json:"Description,omitempty"`

	// A name for the pipeline. Because AWS CloudFormation assigns each new
	// pipeline a unique identifier, you can use the same name for multiple
	// pipelines that are associated with your AWS account.
	Name *StringExpr `json:"Name,omitempty"`

	// Defines the variables that are in the pipeline definition. For more
	// information, see Creating a Pipeline Using Parameterized Templates in
	// the AWS Data Pipeline Developer Guide.
	ParameterObjects *DataPipelinePipelineParameterObjectsList `json:"ParameterObjects,omitempty"`

	// Defines the values for the parameters that are defined in the
	// ParameterObjects property. For more information, see Creating a
	// Pipeline Using Parameterized Templates in the AWS Data Pipeline
	// Developer Guide.
	ParameterValues *DataPipelinePipelineParameterValuesList `json:"ParameterValues,omitempty"`

	// A list of pipeline objects that make up the pipeline. For more
	// information about pipeline objects and a description of each object,
	// see Pipeline Object Reference in the AWS Data Pipeline Developer
	// Guide.
	PipelineObjects *DataPipelinePipelineObjectsList `json:"PipelineObjects,omitempty"`

	// A list of arbitrary tags (key-value pairs) to associate with the
	// pipeline, which you can use to control permissions. For more
	// information, see Controlling Access to Pipelines and Resources in the
	// AWS Data Pipeline Developer Guide.
	PipelineTags *DataPipelinePipelinePipelineTagsList `json:"PipelineTags,omitempty"`
}

// CfnResourceType returns AWS::DataPipeline::Pipeline to implement the ResourceProperties interface
func (s DataPipelinePipeline) CfnResourceType() string {
	return "AWS::DataPipeline::Pipeline"
}

// DirectoryServiceMicrosoftAD represents AWS::DirectoryService::MicrosoftAD
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-directoryservice-microsoftad.html
type DirectoryServiceMicrosoftAD struct {
	// A unique alias to assign to the Microsoft Active Directory in AWS. AWS
	// Directory Service uses the alias to construct the access URL for the
	// directory, such as http://alias.awsapps.com. By default, AWS
	// CloudFormation does not create an alias.
	CreateAlias *BoolExpr `json:"CreateAlias,omitempty"`

	// Whether to enable single sign-on for a Microsoft Active Directory in
	// AWS. Single sign-on allows users in your directory to access certain
	// AWS services from a computer joined to the directory without having to
	// enter their credentials separately. If you don't specify a value, AWS
	// CloudFormation disables single sign-on by default.
	EnableSso *BoolExpr `json:"EnableSso,omitempty"`

	// The fully qualified name for the Microsoft Active Directory in AWS,
	// such as corp.example.com. The name doesn't need to be publicly
	// resolvable; it will resolve inside your VPC only.
	Name *StringExpr `json:"Name,omitempty"`

	// The password for the default administrative user, Admin.
	Password *StringExpr `json:"Password,omitempty"`

	// The NetBIOS name for your domain, such as CORP. If you don't specify a
	// value, AWS Directory Service uses the first part of your directory DNS
	// server name. For example, if your directory DNS server name is
	// corp.example.com, AWS Directory Service specifies CORP for the NetBIOS
	// name.
	ShortName *StringExpr `json:"ShortName,omitempty"`

	// Specifies the VPC settings of the Microsoft Active Directory server in
	// AWS.
	VpcSettings *DirectoryServiceMicrosoftADVpcSettings `json:"VpcSettings,omitempty"`
}

// CfnResourceType returns AWS::DirectoryService::MicrosoftAD to implement the ResourceProperties interface
func (s DirectoryServiceMicrosoftAD) CfnResourceType() string {
	return "AWS::DirectoryService::MicrosoftAD"
}

// DirectoryServiceSimpleAD represents AWS::DirectoryService::SimpleAD
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-directoryservice-simplead.html
type DirectoryServiceSimpleAD struct {
	// A unique alias to assign to the directory. AWS Directory Service uses
	// the alias to construct the access URL for the directory, such as
	// http://alias.awsapps.com. By default, AWS CloudFormation does not
	// create an alias.
	CreateAlias *BoolExpr `json:"CreateAlias,omitempty"`

	// A description of the directory.
	Description *StringExpr `json:"Description,omitempty"`

	// Whether to enable single sign-on for a directory. If you don't specify
	// a value, AWS CloudFormation disables single sign-on by default.
	EnableSso *BoolExpr `json:"EnableSso,omitempty"`

	// The fully qualified name for the directory, such as corp.example.com.
	Name *StringExpr `json:"Name,omitempty"`

	// The password for the directory administrator. AWS Directory Service
	// creates a directory administrator account with the user name
	// Administrator and this password.
	Password *StringExpr `json:"Password,omitempty"`

	// The NetBIOS name of the on-premises directory, such as CORP.
	ShortName *StringExpr `json:"ShortName,omitempty"`

	// The size of the directory. For valid values, see CreateDirectory in
	// the AWS Directory Service API Reference.
	Size *StringExpr `json:"Size,omitempty"`

	// Specifies the VPC settings of the directory server.
	VpcSettings *DirectoryServiceSimpleADVpcSettings `json:"VpcSettings,omitempty"`
}

// CfnResourceType returns AWS::DirectoryService::SimpleAD to implement the ResourceProperties interface
func (s DirectoryServiceSimpleAD) CfnResourceType() string {
	return "AWS::DirectoryService::SimpleAD"
}

// DynamoDBTable represents AWS::DynamoDB::Table
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-dynamodb-table.html
type DynamoDBTable struct {
	// A list of AttributeName and AttributeType objects that describe the
	// key schema for the table and indexes.
	AttributeDefinitions *DynamoDBAttributeDefinitionsList `json:"AttributeDefinitions,omitempty"`

	// Global secondary indexes to be created on the table. You can create up
	// to 5 global secondary indexes.
	GlobalSecondaryIndexes *DynamoDBGlobalSecondaryIndexesList `json:"GlobalSecondaryIndexes,omitempty"`

	// Specifies the attributes that make up the primary key for the table.
	// The attributes in the KeySchema property must also be defined in the
	// AttributeDefinitions property.
	KeySchema *DynamoDBKeySchemaList `json:"KeySchema,omitempty"`

	// Local secondary indexes to be created on the table. You can create up
	// to 5 local secondary indexes. Each index is scoped to a given hash key
	// value. The size of each hash key can be up to 10 gigabytes.
	LocalSecondaryIndexes *DynamoDBLocalSecondaryIndexesList `json:"LocalSecondaryIndexes,omitempty"`

	// Throughput for the specified table, consisting of values for
	// ReadCapacityUnits and WriteCapacityUnits. For more information about
	// the contents of a provisioned throughput structure, see DynamoDB
	// Provisioned Throughput.
	ProvisionedThroughput *DynamoDBProvisionedThroughput `json:"ProvisionedThroughput,omitempty"`

	// The settings for the DynamoDB table stream, which capture changes to
	// items stored in the table.
	StreamSpecification *DynamoDBTableStreamSpecification `json:"StreamSpecification,omitempty"`

	// A name for the table. If you don't specify a name, AWS CloudFormation
	// generates a unique physical ID and uses that ID for the table name.
	// For more information, see Name Type.
	TableName *StringExpr `json:"TableName,omitempty"`
}

// CfnResourceType returns AWS::DynamoDB::Table to implement the ResourceProperties interface
func (s DynamoDBTable) CfnResourceType() string {
	return "AWS::DynamoDB::Table"
}

// EC2CustomerGateway represents AWS::EC2::CustomerGateway
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-customer-gateway.html
type EC2CustomerGateway struct {
	// The customer gateway's Border Gateway Protocol (BGP) Autonomous System
	// Number (ASN).
	BgpAsn *IntegerExpr `json:"BgpAsn,omitempty"`

	// The internet-routable IP address for the customer gateway's outside
	// interface. The address must be static.
	IpAddress *StringExpr `json:"IpAddress,omitempty"`

	// The tags that you want to attach to the resource.
	Tags []ResourceTag `json:"Tags,omitempty"`

	// The type of VPN connection that this customer gateway supports.
	Type *StringExpr `json:"Type,omitempty"`
}

// CfnResourceType returns AWS::EC2::CustomerGateway to implement the ResourceProperties interface
func (s EC2CustomerGateway) CfnResourceType() string {
	return "AWS::EC2::CustomerGateway"
}

// EC2DHCPOptions represents AWS::EC2::DHCPOptions
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-dhcp-options.html
type EC2DHCPOptions struct {
	// A domain name of your choice.
	DomainName *StringExpr `json:"DomainName,omitempty"`

	// The IP (IPv4) address of a domain name server. You can specify up to
	// four addresses.
	DomainNameServers *StringListExpr `json:"DomainNameServers,omitempty"`

	// The IP address (IPv4) of a NetBIOS name server. You can specify up to
	// four addresses.
	NetbiosNameServers *StringListExpr `json:"NetbiosNameServers,omitempty"`

	// An integer value indicating the NetBIOS node type:
	NetbiosNodeType interface{} `json:"NetbiosNodeType,omitempty"`

	// The IP address (IPv4) of a Network Time Protocol (NTP) server. You can
	// specify up to four addresses.
	NtpServers *StringListExpr `json:"NtpServers,omitempty"`

	// An arbitrary set of tags (key–value pairs) for this resource.
	Tags []ResourceTag `json:"Tags,omitempty"`
}

// CfnResourceType returns AWS::EC2::DHCPOptions to implement the ResourceProperties interface
func (s EC2DHCPOptions) CfnResourceType() string {
	return "AWS::EC2::DHCPOptions"
}

// EC2EIP represents AWS::EC2::EIP
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-eip.html
type EC2EIP struct {
	// The Instance ID of the Amazon EC2 instance that you want to associate
	// with this Elastic IP address.
	InstanceId *StringExpr `json:"InstanceId,omitempty"`

	// Set to vpc to allocate the address to your Virtual Private Cloud
	// (VPC). No other values are supported.
	Domain *StringExpr `json:"Domain,omitempty"`
}

// CfnResourceType returns AWS::EC2::EIP to implement the ResourceProperties interface
func (s EC2EIP) CfnResourceType() string {
	return "AWS::EC2::EIP"
}

// EC2EIPAssociation represents AWS::EC2::EIPAssociation
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-eip-association.html
type EC2EIPAssociation struct {
	// Allocation ID for the VPC Elastic IP address you want to associate
	// with an Amazon EC2 instance in your VPC.
	AllocationId *StringExpr `json:"AllocationId,omitempty"`

	// Elastic IP address that you want to associate with the Amazon EC2
	// instance specified by the InstanceId property. You can specify an
	// existing Elastic IP address or a reference to an Elastic IP address
	// allocated with a AWS::EC2::EIP resource.
	EIP *StringExpr `json:"EIP,omitempty"`

	// Instance ID of the Amazon EC2 instance that you want to associate with
	// the Elastic IP address specified by the EIP property.
	InstanceId *StringExpr `json:"InstanceId,omitempty"`

	// The ID of the network interface to associate with the Elastic IP
	// address (VPC only).
	NetworkInterfaceId *StringExpr `json:"NetworkInterfaceId,omitempty"`

	// The private IP address that you want to associate with the Elastic IP
	// address. The private IP address is restricted to the primary and
	// secondary private IP addresses that are associated with the network
	// interface. By default, the private IP address that is associated with
	// the EIP is the primary private IP address of the network interface.
	PrivateIpAddress *StringExpr `json:"PrivateIpAddress,omitempty"`
}

// CfnResourceType returns AWS::EC2::EIPAssociation to implement the ResourceProperties interface
func (s EC2EIPAssociation) CfnResourceType() string {
	return "AWS::EC2::EIPAssociation"
}

// EC2FlowLog represents AWS::EC2::FlowLog
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-flowlog.html
type EC2FlowLog struct {
	// The Amazon Resource Name (ARN) of an AWS Identity and Access
	// Management (IAM) role that permits Amazon EC2 to publish flow logs to
	// a CloudWatch Logs log group in your account.
	DeliverLogsPermissionArn *StringExpr `json:"DeliverLogsPermissionArn,omitempty"`

	// The name of a new or existing CloudWatch Logs log group where Amazon
	// EC2 publishes your flow logs.
	LogGroupName *StringExpr `json:"LogGroupName,omitempty"`

	// The ID of the subnet, network interface, or VPC for which you want to
	// create a flow log.
	ResourceId *StringExpr `json:"ResourceId,omitempty"`

	// The type of resource that you specified in the ResourceId property.
	// For example, if you specified a VPC ID for the ResourceId property,
	// specify VPC for this property. For valid values, see the ResourceType
	// parameter for the CreateFlowLogs action in the Amazon EC2 API
	// Reference.
	ResourceType *StringExpr `json:"ResourceType,omitempty"`

	// The type of traffic to log. You can log traffic that the resource
	// accepts or rejects, or all traffic. For valid values, see the
	// TrafficType parameter for the CreateFlowLogs action in the Amazon EC2
	// API Reference.
	TrafficType *StringExpr `json:"TrafficType,omitempty"`
}

// CfnResourceType returns AWS::EC2::FlowLog to implement the ResourceProperties interface
func (s EC2FlowLog) CfnResourceType() string {
	return "AWS::EC2::FlowLog"
}

// EC2Host represents AWS::EC2::Host
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-host.html
type EC2Host struct {
	// Indicates if the host accepts EC2 instances with only matching
	// configurations or if instances must also specify the host ID.
	// Instances that don't specify a host ID can't launch onto a host with
	// AutoPlacement set to off. By default, AWS CloudFormation sets this
	// property to on. For more information, see Understanding Instance
	// Placement and Host Affinity in the Amazon EC2 User Guide for Linux
	// Instances.
	AutoPlacement *StringExpr `json:"AutoPlacement,omitempty"`

	// The Availability Zone (AZ) in which to launch the dedicated host.
	AvailabilityZone *StringExpr `json:"AvailabilityZone,omitempty"`

	// The instance type that the dedicated host accepts. Only instances of
	// this type can be launched onto the host. For more information, see
	// Supported Instance Types in the Amazon EC2 User Guide for Linux
	// Instances.
	InstanceType *StringExpr `json:"InstanceType,omitempty"`
}

// CfnResourceType returns AWS::EC2::Host to implement the ResourceProperties interface
func (s EC2Host) CfnResourceType() string {
	return "AWS::EC2::Host"
}

// EC2Instance represents AWS::EC2::Instance
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-instance.html
type EC2Instance struct {
	// Indicates whether Amazon Elastic Compute Cloud (Amazon EC2) always
	// associates the instance with a dedicated host. If you want Amazon EC2
	// to always restart the instance (if it was stopped) onto the same host
	// on which it was launched, specify host. If you want Amazon EC2 to
	// restart the instance on any available host, but to try to launch the
	// instance onto the last host it ran on (on a best-effort basis),
	// specify default.
	Affinity *StringExpr `json:"Affinity,omitempty"`

	// Specifies the name of the Availability Zone in which the instance is
	// located.
	AvailabilityZone *StringExpr `json:"AvailabilityZone,omitempty"`

	// Defines a set of Amazon Elastic Block Store block device mappings,
	// ephemeral instance store block device mappings, or both. For more
	// information, see Amazon Elastic Block Store or Amazon EC2 Instance
	// Store in the Amazon EC2 User Guide for Linux Instances.
	BlockDeviceMappings *EC2BlockDeviceMappingPropertyList `json:"BlockDeviceMappings,omitempty"`

	// Specifies whether the instance can be terminated through the API.
	DisableApiTermination *BoolExpr `json:"DisableApiTermination,omitempty"`

	// Specifies whether the instance is optimized for Amazon Elastic Block
	// Store I/O. This optimization provides dedicated throughput to Amazon
	// EBS and an optimized configuration stack to provide optimal EBS I/O
	// performance.
	EbsOptimized *BoolExpr `json:"EbsOptimized,omitempty"`

	// If you specify host for the Affinity property, the ID of a dedicated
	// host that the instance is associated with. If you don't specify an ID,
	// Amazon EC2 launches the instance onto any available, compatible
	// dedicated host in your account. This type of launch is called an
	// untargeted launch. Note that for untargeted launches, you must have a
	// compatible, dedicated host available to successfully launch instances.
	HostId *StringExpr `json:"HostId,omitempty"`

	// The physical ID (resource name) of an instance profile or a reference
	// to an AWS::IAM::InstanceProfile resource.
	IamInstanceProfile *StringExpr `json:"IamInstanceProfile,omitempty"`

	// Provides the unique ID of the Amazon Machine Image (AMI) that was
	// assigned during registration.
	ImageId *StringExpr `json:"ImageId,omitempty"`

	// Indicates whether an instance stops or terminates when you shut down
	// the instance from the instance's operating system shutdown command.
	// You can specify stop or terminate. For more information, see the
	// RunInstances command in the Amazon EC2 API Reference.
	InstanceInitiatedShutdownBehavior *StringExpr `json:"InstanceInitiatedShutdownBehavior,omitempty"`

	// The instance type, such as t2.micro. The default type is "m3.medium".
	// For a list of instance types, see Instance Families and Types.
	InstanceType *StringExpr `json:"InstanceType,omitempty"`

	// The kernel ID.
	KernelId *StringExpr `json:"KernelId,omitempty"`

	// Provides the name of the Amazon EC2 key pair.
	KeyName *StringExpr `json:"KeyName,omitempty"`

	// Specifies whether detailed monitoring is enabled for the instance.
	Monitoring *BoolExpr `json:"Monitoring,omitempty"`

	// A list of embedded objects that describes the network interfaces to
	// associate with this instance.
	NetworkInterfaces *EC2NetworkInterfaceEmbeddedList `json:"NetworkInterfaces,omitempty"`

	// The name of an existing placement group that you want to launch the
	// instance into (for cluster instances).
	PlacementGroupName *StringExpr `json:"PlacementGroupName,omitempty"`

	// The private IP address for this instance.
	PrivateIpAddress *StringExpr `json:"PrivateIpAddress,omitempty"`

	// The ID of the RAM disk to select. Some kernels require additional
	// drivers at launch. Check the kernel requirements for information about
	// whether you need to specify a RAM disk. To find kernel requirements,
	// go to the AWS Resource Center and search for the kernel ID.
	RamdiskId *StringExpr `json:"RamdiskId,omitempty"`

	// A list that contains the security group IDs for VPC security groups to
	// assign to the Amazon EC2 instance. If you specified the
	// NetworkInterfaces property, do not specify this property.
	SecurityGroupIds *StringListExpr `json:"SecurityGroupIds,omitempty"`

	// Valid only for Amazon EC2 security groups. A list that contains the
	// Amazon EC2 security groups to assign to the Amazon EC2 instance. The
	// list can contain both the name of existing Amazon EC2 security groups
	// or references to AWS::EC2::SecurityGroup resources created in the
	// template.
	SecurityGroups *StringListExpr `json:"SecurityGroups,omitempty"`

	// Controls whether source/destination checking is enabled on the
	// instance. Also determines if an instance in a VPC will perform network
	// address translation (NAT).
	SourceDestCheck *BoolExpr `json:"SourceDestCheck,omitempty"`

	// The Amazon EC2 Simple Systems Manager (SSM) document and parameter
	// values to associate with this instance. To use this property, you must
	// specify an IAM role for the instance. For more information, see
	// Prerequisites for Remotely Running Commands on EC2 Instances in the
	// Amazon EC2 User Guide for Windows Instances.
	SsmAssociations *EC2InstanceSsmAssociationsList `json:"SsmAssociations,omitempty"`

	// If you're using Amazon VPC, this property specifies the ID of the
	// subnet that you want to launch the instance into. If you specified the
	// NetworkInterfaces property, do not specify this property.
	SubnetId *StringExpr `json:"SubnetId,omitempty"`

	// An arbitrary set of tags (key–value pairs) for this instance.
	Tags []ResourceTag `json:"Tags,omitempty"`

	// The tenancy of the instance that you want to launch, such as default,
	// dedicated, or host. If you specify a tenancy value of dedicated or
	// host, you must launch the instance in a VPC. For more information, see
	// Dedicated Instances in the Amazon VPC User Guide.
	Tenancy *StringExpr `json:"Tenancy,omitempty"`

	// Base64-encoded MIME user data that is made available to the instances.
	UserData *StringExpr `json:"UserData,omitempty"`

	// The Amazon EBS volumes to attach to the instance.
	Volumes *EC2MountPointList `json:"Volumes,omitempty"`

	// Reserved.
	AdditionalInfo *StringExpr `json:"AdditionalInfo,omitempty"`
}

// CfnResourceType returns AWS::EC2::Instance to implement the ResourceProperties interface
func (s EC2Instance) CfnResourceType() string {
	return "AWS::EC2::Instance"
}

// EC2InternetGateway represents AWS::EC2::InternetGateway
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-internet-gateway.html
type EC2InternetGateway struct {
	// An arbitrary set of tags (key–value pairs) for this resource.
	Tags []ResourceTag `json:"Tags,omitempty"`
}

// CfnResourceType returns AWS::EC2::InternetGateway to implement the ResourceProperties interface
func (s EC2InternetGateway) CfnResourceType() string {
	return "AWS::EC2::InternetGateway"
}

// EC2NatGateway represents AWS::EC2::NatGateway
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-natgateway.html
type EC2NatGateway struct {
	// The allocation ID of an Elastic IP address to associate with the NAT
	// gateway. If the Elastic IP address is associated with another
	// resource, you must first disassociate it.
	AllocationId *StringExpr `json:"AllocationId,omitempty"`

	// The public subnet in which to create the NAT gateway.
	SubnetId *StringExpr `json:"SubnetId,omitempty"`
}

// CfnResourceType returns AWS::EC2::NatGateway to implement the ResourceProperties interface
func (s EC2NatGateway) CfnResourceType() string {
	return "AWS::EC2::NatGateway"
}

// EC2NetworkAcl represents AWS::EC2::NetworkAcl
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-network-acl.html
type EC2NetworkAcl struct {
	// An arbitrary set of tags (key–value pairs) for this ACL.
	Tags []ResourceTag `json:"Tags,omitempty"`

	// The ID of the VPC where the network ACL will be created.
	VpcId *StringExpr `json:"VpcId,omitempty"`
}

// CfnResourceType returns AWS::EC2::NetworkAcl to implement the ResourceProperties interface
func (s EC2NetworkAcl) CfnResourceType() string {
	return "AWS::EC2::NetworkAcl"
}

// EC2NetworkAclEntry represents AWS::EC2::NetworkAclEntry
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-network-acl-entry.html
type EC2NetworkAclEntry struct {
	// The CIDR range to allow or deny, in CIDR notation (e.g.,
	// 172.16.0.0/24).
	CidrBlock *StringExpr `json:"CidrBlock,omitempty"`

	// Whether this rule applies to egress traffic from the subnet (true) or
	// ingress traffic to the subnet (false). By default, AWS CloudFormation
	// specifies false.
	Egress *BoolExpr `json:"Egress,omitempty"`

	// The Internet Control Message Protocol (ICMP) code and type.
	Icmp *EC2NetworkAclEntryIcmp `json:"Icmp,omitempty"`

	// ID of the ACL where the entry will be created.
	NetworkAclId *StringExpr `json:"NetworkAclId,omitempty"`

	// The range of port numbers for the UDP/TCP protocol.
	PortRange *EC2NetworkAclEntryPortRange `json:"PortRange,omitempty"`

	// The IP protocol that the rule applies to. You must specify -1 or a
	// protocol number (go to Protocol Numbers at iana.org). You can specify
	// -1 for all protocols.
	Protocol *IntegerExpr `json:"Protocol,omitempty"`

	// Whether to allow or deny traffic that matches the rule; valid values
	// are "allow" or "deny".
	RuleAction *StringExpr `json:"RuleAction,omitempty"`

	// Rule number to assign to the entry (e.g., 100). This must be a
	// positive integer from 1 to 32766.
	RuleNumber *IntegerExpr `json:"RuleNumber,omitempty"`
}

// CfnResourceType returns AWS::EC2::NetworkAclEntry to implement the ResourceProperties interface
func (s EC2NetworkAclEntry) CfnResourceType() string {
	return "AWS::EC2::NetworkAclEntry"
}

// EC2NetworkInterface represents AWS::EC2::NetworkInterface
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-network-interface.html
type EC2NetworkInterface struct {
	// The description of this network interface.
	Description *StringExpr `json:"Description,omitempty"`

	// A list of security group IDs associated with this network interface.
	GroupSet *StringListExpr `json:"GroupSet,omitempty"`

	// Assigns a single private IP address to the network interface, which is
	// used as the primary private IP address. If you want to specify
	// multiple private IP address, use the PrivateIpAddresses property.
	PrivateIpAddress *StringExpr `json:"PrivateIpAddress,omitempty"`

	// Assigns a list of private IP addresses to the network interface. You
	// can specify a primary private IP address by setting the value of the
	// Primary property to true in the PrivateIpAddressSpecification
	// property. If you want Amazon EC2 to automatically assign private IP
	// addresses, use the SecondaryPrivateIpAddressCount property and do not
	// specify this property.
	PrivateIpAddresses *EC2NetworkInterfacePrivateIPSpecificationList `json:"PrivateIpAddresses,omitempty"`

	// The number of secondary private IP addresses that Amazon EC2
	// automatically assigns to the network interface. Amazon EC2 uses the
	// value of the PrivateIpAddress property as the primary private IP
	// address. If you don't specify that property, Amazon EC2 automatically
	// assigns both the primary and secondary private IP addresses.
	SecondaryPrivateIpAddressCount *IntegerExpr `json:"SecondaryPrivateIpAddressCount,omitempty"`

	// Flag indicating whether traffic to or from the instance is validated.
	SourceDestCheck *BoolExpr `json:"SourceDestCheck,omitempty"`

	// The ID of the subnet to associate with the network interface.
	SubnetId *StringExpr `json:"SubnetId,omitempty"`

	// An arbitrary set of tags (key–value pairs) for this network
	// interface.
	Tags []ResourceTag `json:"Tags,omitempty"`
}

// CfnResourceType returns AWS::EC2::NetworkInterface to implement the ResourceProperties interface
func (s EC2NetworkInterface) CfnResourceType() string {
	return "AWS::EC2::NetworkInterface"
}

// EC2NetworkInterfaceAttachment represents AWS::EC2::NetworkInterfaceAttachment
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-network-interface-attachment.html
type EC2NetworkInterfaceAttachment struct {
	// Whether to delete the network interface when the instance terminates.
	// By default, this value is set to True.
	DeleteOnTermination *BoolExpr `json:"DeleteOnTermination,omitempty"`

	// The network interface's position in the attachment order. For example,
	// the first attached network interface has a DeviceIndex of 0.
	DeviceIndex *StringExpr `json:"DeviceIndex,omitempty"`

	// The ID of the instance to which you will attach the ENI.
	InstanceId *StringExpr `json:"InstanceId,omitempty"`

	// The ID of the ENI that you want to attach.
	NetworkInterfaceId *StringExpr `json:"NetworkInterfaceId,omitempty"`
}

// CfnResourceType returns AWS::EC2::NetworkInterfaceAttachment to implement the ResourceProperties interface
func (s EC2NetworkInterfaceAttachment) CfnResourceType() string {
	return "AWS::EC2::NetworkInterfaceAttachment"
}

// EC2PlacementGroup represents AWS::EC2::PlacementGroup
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-placementgroup.html
type EC2PlacementGroup struct {
	// The placement strategy, which relates to the instance types that can
	// be added to the placement group. For example, for the cluster
	// strategy, you can cluster C4 instance types but not T2 instance types.
	// For valid values, see CreatePlacementGroup in the Amazon EC2 API
	// Reference. By default, AWS CloudFormation sets the value of this
	// property to cluster.
	Strategy *StringExpr `json:"Strategy,omitempty"`
}

// CfnResourceType returns AWS::EC2::PlacementGroup to implement the ResourceProperties interface
func (s EC2PlacementGroup) CfnResourceType() string {
	return "AWS::EC2::PlacementGroup"
}

// EC2Route represents AWS::EC2::Route
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-route.html
type EC2Route struct {
	// The CIDR address block used for the destination match. For example,
	// 0.0.0.0/0. Routing decisions are based on the most specific match.
	DestinationCidrBlock *StringExpr `json:"DestinationCidrBlock,omitempty"`

	// The ID of an Internet gateway or virtual private gateway that is
	// attached to your VPC. For example: igw-eaad4883.
	GatewayId *StringExpr `json:"GatewayId,omitempty"`

	// The ID of a NAT instance in your VPC. For example, i-1a2b3c4d.
	InstanceId *StringExpr `json:"InstanceId,omitempty"`

	// The ID of a NAT gateway. For example, nat-0a12bc456789de0fg.
	NatGatewayId *StringExpr `json:"NatGatewayId,omitempty"`

	// Allows the routing of network interface IDs.
	NetworkInterfaceId *StringExpr `json:"NetworkInterfaceId,omitempty"`

	// The ID of the route table where the route will be added.
	RouteTableId *StringExpr `json:"RouteTableId,omitempty"`

	// The ID of a VPC peering connection.
	VpcPeeringConnectionId *StringExpr `json:"VpcPeeringConnectionId,omitempty"`
}

// CfnResourceType returns AWS::EC2::Route to implement the ResourceProperties interface
func (s EC2Route) CfnResourceType() string {
	return "AWS::EC2::Route"
}

// EC2RouteTable represents AWS::EC2::RouteTable
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-route-table.html
type EC2RouteTable struct {
	// The ID of the VPC where the route table will be created.
	VpcId *StringExpr `json:"VpcId,omitempty"`

	// An arbitrary set of tags (key–value pairs) for this route table.
	Tags []ResourceTag `json:"Tags,omitempty"`
}

// CfnResourceType returns AWS::EC2::RouteTable to implement the ResourceProperties interface
func (s EC2RouteTable) CfnResourceType() string {
	return "AWS::EC2::RouteTable"
}

// EC2SecurityGroup represents AWS::EC2::SecurityGroup
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-security-group.html
type EC2SecurityGroup struct {
	// Description of the security group.
	GroupDescription *StringExpr `json:"GroupDescription,omitempty"`

	// A list of Amazon EC2 security group egress rules.
	SecurityGroupEgress *EC2SecurityGroupRuleList `json:"SecurityGroupEgress,omitempty"`

	// A list of Amazon EC2 security group ingress rules.
	SecurityGroupIngress *EC2SecurityGroupRuleList `json:"SecurityGroupIngress,omitempty"`

	// The tags that you want to attach to the resource.
	Tags []ResourceTag `json:"Tags,omitempty"`

	// The physical ID of the VPC. Can be obtained by using a reference to an
	// AWS::EC2::VPC, such as: { "Ref" : "myVPC" }.
	VpcId *StringExpr `json:"VpcId,omitempty"`
}

// CfnResourceType returns AWS::EC2::SecurityGroup to implement the ResourceProperties interface
func (s EC2SecurityGroup) CfnResourceType() string {
	return "AWS::EC2::SecurityGroup"
}

// EC2SecurityGroupEgress represents AWS::EC2::SecurityGroupEgress
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-security-group-egress.html
type EC2SecurityGroupEgress struct {
	// CIDR range.
	CidrIp *StringExpr `json:"CidrIp,omitempty"`

	// The AWS service prefix of an Amazon VPC endpoint. For more
	// information, see VPC Endpoints in the Amazon VPC User Guide.
	DestinationPrefixListId *StringExpr `json:"DestinationPrefixListId,omitempty"`

	// Specifies the group ID of the destination Amazon VPC security group.
	DestinationSecurityGroupId *StringExpr `json:"DestinationSecurityGroupId,omitempty"`

	// Start of port range for the TCP and UDP protocols, or an ICMP type
	// number. If you specify icmp for the IpProtocol property, you can
	// specify -1 as a wildcard (i.e., any ICMP type number).
	FromPort *IntegerExpr `json:"FromPort,omitempty"`

	// ID of the Amazon VPC security group to modify. This value can be a
	// reference to an AWS::EC2::SecurityGroup resource that has a valid
	// VpcId property or the ID of an existing Amazon VPC security group.
	GroupId *StringExpr `json:"GroupId,omitempty"`

	// IP protocol name or number. For valid values, see the IpProtocol
	// parameter in AuthorizeSecurityGroupIngress
	IpProtocol *StringExpr `json:"IpProtocol,omitempty"`

	// End of port range for the TCP and UDP protocols, or an ICMP code. If
	// you specify icmp for the IpProtocol property, you can specify -1 as a
	// wildcard (i.e., any ICMP code).
	ToPort *IntegerExpr `json:"ToPort,omitempty"`
}

// CfnResourceType returns AWS::EC2::SecurityGroupEgress to implement the ResourceProperties interface
func (s EC2SecurityGroupEgress) CfnResourceType() string {
	return "AWS::EC2::SecurityGroupEgress"
}

// EC2SecurityGroupIngress represents AWS::EC2::SecurityGroupIngress
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-security-group-ingress.html
type EC2SecurityGroupIngress struct {
	// Specifies a CIDR range.
	CidrIp *StringExpr `json:"CidrIp,omitempty"`

	// Start of port range for the TCP and UDP protocols, or an ICMP type
	// number. If you specify icmp for the IpProtocol property, you can
	// specify -1 as a wildcard (i.e., any ICMP type number).
	FromPort *IntegerExpr `json:"FromPort,omitempty"`

	// ID of the Amazon EC2 or VPC security group to modify. The group must
	// belong to your account.
	GroupId *StringExpr `json:"GroupId,omitempty"`

	// Name of the Amazon EC2 security group (non-VPC security group) to
	// modify. This value can be a reference to an AWS::EC2::SecurityGroup
	// resource or the name of an existing Amazon EC2 security group.
	GroupName *StringExpr `json:"GroupName,omitempty"`

	// IP protocol name or number. For valid values, see the IpProtocol
	// parameter in AuthorizeSecurityGroupIngress
	IpProtocol *StringExpr `json:"IpProtocol,omitempty"`

	// Specifies the ID of the source security group or uses the Ref
	// intrinsic function to refer to the logical ID of a security group
	// defined in the same template.
	SourceSecurityGroupId *StringExpr `json:"SourceSecurityGroupId,omitempty"`

	// Specifies the name of the Amazon EC2 security group (non-VPC security
	// group) to allow access or uses the Ref intrinsic function to refer to
	// the logical name of a security group defined in the same template. For
	// instances in a VPC, specify the SourceSecurityGroupId property.
	SourceSecurityGroupName *StringExpr `json:"SourceSecurityGroupName,omitempty"`

	// Specifies the AWS Account ID of the owner of the Amazon EC2 security
	// group specified in the SourceSecurityGroupName property.
	SourceSecurityGroupOwnerId *StringExpr `json:"SourceSecurityGroupOwnerId,omitempty"`

	// End of port range for the TCP and UDP protocols, or an ICMP code. If
	// you specify icmp for the IpProtocol property, you can specify -1 as a
	// wildcard (i.e., any ICMP code).
	ToPort *IntegerExpr `json:"ToPort,omitempty"`
}

// CfnResourceType returns AWS::EC2::SecurityGroupIngress to implement the ResourceProperties interface
func (s EC2SecurityGroupIngress) CfnResourceType() string {
	return "AWS::EC2::SecurityGroupIngress"
}

// EC2SpotFleet represents AWS::EC2::SpotFleet
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-spotfleet.html
type EC2SpotFleet struct {
	// The configuration for a Spot fleet request.
	SpotFleetRequestConfigData *EC2SpotFleetSpotFleetRequestConfigData `json:"SpotFleetRequestConfigData,omitempty"`
}

// CfnResourceType returns AWS::EC2::SpotFleet to implement the ResourceProperties interface
func (s EC2SpotFleet) CfnResourceType() string {
	return "AWS::EC2::SpotFleet"
}

// EC2Subnet represents AWS::EC2::Subnet
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-subnet.html
type EC2Subnet struct {
	// The availability zone in which you want the subnet. Default: AWS
	// selects a zone for you (recommended).
	AvailabilityZone *StringExpr `json:"AvailabilityZone,omitempty"`

	// The CIDR block that you want the subnet to cover (for example,
	// "10.0.0.0/24").
	CidrBlock *StringExpr `json:"CidrBlock,omitempty"`

	// Indicates whether instances that are launched in this subnet receive a
	// public IP address. By default, the value is false.
	MapPublicIpOnLaunch *BoolExpr `json:"MapPublicIpOnLaunch,omitempty"`

	// An arbitrary set of tags (key–value pairs) for this subnet.
	Tags []ResourceTag `json:"Tags,omitempty"`

	// A Ref structure that contains the ID of the VPC on which you want to
	// create the subnet. The VPC ID is provided as the value of the "Ref"
	// property, as: { "Ref": "VPCID" }.
	VpcId interface{} `json:"VpcId,omitempty"`
}

// CfnResourceType returns AWS::EC2::Subnet to implement the ResourceProperties interface
func (s EC2Subnet) CfnResourceType() string {
	return "AWS::EC2::Subnet"
}

// EC2SubnetNetworkAclAssociation represents AWS::EC2::SubnetNetworkAclAssociation
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-subnet-network-acl-assoc.html
type EC2SubnetNetworkAclAssociation struct {
	// The ID representing the current association between the original
	// network ACL and the subnet.
	SubnetId *StringExpr `json:"SubnetId,omitempty"`

	// The ID of the new ACL to associate with the subnet.
	NetworkAclId *StringExpr `json:"NetworkAclId,omitempty"`
}

// CfnResourceType returns AWS::EC2::SubnetNetworkAclAssociation to implement the ResourceProperties interface
func (s EC2SubnetNetworkAclAssociation) CfnResourceType() string {
	return "AWS::EC2::SubnetNetworkAclAssociation"
}

// EC2SubnetRouteTableAssociation represents AWS::EC2::SubnetRouteTableAssociation
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-subnet-route-table-assoc.html
type EC2SubnetRouteTableAssociation struct {
	// The ID of the route table. This is commonly written as a reference to
	// a route table declared elsewhere in the template. For example:
	RouteTableId *StringExpr `json:"RouteTableId,omitempty"`

	// The ID of the subnet. This is commonly written as a reference to a
	// subnet declared elsewhere in the template. For example:
	SubnetId *StringExpr `json:"SubnetId,omitempty"`
}

// CfnResourceType returns AWS::EC2::SubnetRouteTableAssociation to implement the ResourceProperties interface
func (s EC2SubnetRouteTableAssociation) CfnResourceType() string {
	return "AWS::EC2::SubnetRouteTableAssociation"
}

// EC2Volume represents AWS::EC2::Volume
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-ebs-volume.html
type EC2Volume struct {
	// Indicates whether the volume is auto-enabled for I/O operations. By
	// default, Amazon EBS disables I/O to the volume from attached EC2
	// instances when it determines that a volume's data is potentially
	// inconsistent. If the consistency of the volume is not a concern, and
	// you prefer that the volume be made available immediately if it's
	// impaired, you can configure the volume to automatically enable I/O.
	// For more information, see Working with the AutoEnableIO Volume
	// Attribute in the Amazon EC2 User Guide for Linux Instances.
	AutoEnableIO *BoolExpr `json:"AutoEnableIO,omitempty"`

	// The Availability Zone in which to create the new volume.
	AvailabilityZone *StringExpr `json:"AvailabilityZone,omitempty"`

	// Indicates whether the volume is encrypted. Encrypted Amazon EBS
	// volumes can only be attached to instance types that support Amazon EBS
	// encryption. Volumes that are created from encrypted snapshots are
	// automatically encrypted. You cannot create an encrypted volume from an
	// unencrypted snapshot or vice versa. If your AMI uses encrypted
	// volumes, you can only launch the AMI on supported instance types. For
	// more information, see Amazon EBS encryption in the Amazon EC2 User
	// Guide for Linux Instances.
	Encrypted *BoolExpr `json:"Encrypted,omitempty"`

	// The number of I/O operations per second (IOPS) that the volume
	// supports. For more information about the valid sizes for each volume
	// type, see the Iops parameter for the CreateVolume action in the Amazon
	// EC2 API Reference.
	Iops *IntegerExpr `json:"Iops,omitempty"`

	// The Amazon Resource Name (ARN) of the AWS Key Management Service
	// master key that is used to create the encrypted volume, such as
	// arn:aws:kms:us-east-1:012345678910:key/abcd1234-a123-456a-a12b-a123b4cd56ef.
	// If you create an encrypted volume and don't specify this property, the
	// default master key is used.
	KmsKeyId *StringExpr `json:"KmsKeyId,omitempty"`

	// The size of the volume, in gibibytes (GiBs). For more information
	// about the valid sizes for each volume type, see the Size parameter for
	// the CreateVolume action in the Amazon EC2 API Reference.
	Size *StringExpr `json:"Size,omitempty"`

	// The snapshot from which to create the new volume.
	SnapshotId *StringExpr `json:"SnapshotId,omitempty"`

	// An arbitrary set of tags (key–value pairs) for this volume.
	Tags []ResourceTag `json:"Tags,omitempty"`

	// The volume type. If you set the type to io1, you must also set the
	// Iops property. For valid values, see the VolumeType parameter for the
	// CreateVolume action in the Amazon EC2 API Reference.
	VolumeType *StringExpr `json:"VolumeType,omitempty"`
}

// CfnResourceType returns AWS::EC2::Volume to implement the ResourceProperties interface
func (s EC2Volume) CfnResourceType() string {
	return "AWS::EC2::Volume"
}

// EC2VolumeAttachment represents AWS::EC2::VolumeAttachment
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-ebs-volumeattachment.html
type EC2VolumeAttachment struct {
	// How the device is exposed to the instance (e.g., /dev/sdh, or xvdh).
	Device *StringExpr `json:"Device,omitempty"`

	// The ID of the instance to which the volume attaches. This value can be
	// a reference to an AWS::EC2::Instance resource, or it can be the
	// physical ID of an existing EC2 instance.
	InstanceId *StringExpr `json:"InstanceId,omitempty"`

	// The ID of the Amazon EBS volume. The volume and instance must be
	// within the same Availability Zone. This value can be a reference to an
	// AWS::EC2::Volume resource, or it can be the volume ID of an existing
	// Amazon EBS volume.
	VolumeId *StringExpr `json:"VolumeId,omitempty"`
}

// CfnResourceType returns AWS::EC2::VolumeAttachment to implement the ResourceProperties interface
func (s EC2VolumeAttachment) CfnResourceType() string {
	return "AWS::EC2::VolumeAttachment"
}

// EC2VPC represents AWS::EC2::VPC
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-vpc.html
type EC2VPC struct {
	// The CIDR block you want the VPC to cover. For example: "10.0.0.0/16".
	CidrBlock *StringExpr `json:"CidrBlock,omitempty"`

	// Specifies whether DNS resolution is supported for the VPC. If this
	// attribute is true, the Amazon DNS server resolves DNS hostnames for
	// your instances to their corresponding IP addresses; otherwise, it does
	// not. By default the value is set to true.
	EnableDnsSupport *BoolExpr `json:"EnableDnsSupport,omitempty"`

	// Specifies whether the instances launched in the VPC get DNS hostnames.
	// If this attribute is true, instances in the VPC get DNS hostnames;
	// otherwise, they do not. You can only set EnableDnsHostnames to true if
	// you also set the EnableDnsSupport attribute to true. By default, the
	// value is set to false.
	EnableDnsHostnames *BoolExpr `json:"EnableDnsHostnames,omitempty"`

	// The allowed tenancy of instances launched into the VPC.
	InstanceTenancy *StringExpr `json:"InstanceTenancy,omitempty"`

	// An arbitrary set of tags (key–value pairs) for this VPC. To name a
	// VPC resource, specify a value for the Name key.
	Tags []ResourceTag `json:"Tags,omitempty"`
}

// CfnResourceType returns AWS::EC2::VPC to implement the ResourceProperties interface
func (s EC2VPC) CfnResourceType() string {
	return "AWS::EC2::VPC"
}

// EC2VPCDHCPOptionsAssociation represents AWS::EC2::VPCDHCPOptionsAssociation
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-vpc-dhcp-options-assoc.html
type EC2VPCDHCPOptionsAssociation struct {
	// The ID of the DHCP options you want to associate with the VPC. Specify
	// default if you want the VPC to use no DHCP options.
	DhcpOptionsId *StringExpr `json:"DhcpOptionsId,omitempty"`

	// The ID of the VPC to associate with this DHCP options set.
	VpcId *StringExpr `json:"VpcId,omitempty"`
}

// CfnResourceType returns AWS::EC2::VPCDHCPOptionsAssociation to implement the ResourceProperties interface
func (s EC2VPCDHCPOptionsAssociation) CfnResourceType() string {
	return "AWS::EC2::VPCDHCPOptionsAssociation"
}

// EC2VPCEndpoint represents AWS::EC2::VPCEndpoint
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-vpcendpoint.html
type EC2VPCEndpoint struct {
	// A policy to attach to the endpoint that controls access to the
	// service. The policy must be valid JSON. The default policy allows full
	// access to the AWS service. For more information, see Controlling
	// Access to Services in the Amazon VPC User Guide.
	PolicyDocument interface{} `json:"PolicyDocument,omitempty"`

	// One or more route table IDs that are used by the VPC to reach the
	// endpoint.
	RouteTableIds *StringListExpr `json:"RouteTableIds,omitempty"`

	// The AWS service to which you want to establish a connection. Specify
	// the service name in the form of com.amazonaws.region.service.
	ServiceName *StringExpr `json:"ServiceName,omitempty"`

	// The ID of the VPC in which the endpoint is used.
	VpcId *StringExpr `json:"VpcId,omitempty"`
}

// CfnResourceType returns AWS::EC2::VPCEndpoint to implement the ResourceProperties interface
func (s EC2VPCEndpoint) CfnResourceType() string {
	return "AWS::EC2::VPCEndpoint"
}

// EC2VPCGatewayAttachment represents AWS::EC2::VPCGatewayAttachment
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-vpc-gateway-attachment.html
type EC2VPCGatewayAttachment struct {
	// The ID of the Internet gateway.
	InternetGatewayId *StringExpr `json:"InternetGatewayId,omitempty"`

	// The ID of the VPC to associate with this gateway.
	VpcId *StringExpr `json:"VpcId,omitempty"`

	// The ID of the virtual private network (VPN) gateway to attach to the
	// VPC.
	VpnGatewayId *StringExpr `json:"VpnGatewayId,omitempty"`
}

// CfnResourceType returns AWS::EC2::VPCGatewayAttachment to implement the ResourceProperties interface
func (s EC2VPCGatewayAttachment) CfnResourceType() string {
	return "AWS::EC2::VPCGatewayAttachment"
}

// EC2VPCPeeringConnection represents AWS::EC2::VPCPeeringConnection
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-vpcpeeringconnection.html
type EC2VPCPeeringConnection struct {
	// The ID of the VPC with which you are creating the peering connection.
	PeerVpcId *StringExpr `json:"PeerVpcId,omitempty"`

	// An arbitrary set of tags (key–value pairs) for this resource.
	Tags []ResourceTag `json:"Tags,omitempty"`

	// The ID of the VPC that is requesting a peering connection.
	VpcId *StringExpr `json:"VpcId,omitempty"`
}

// CfnResourceType returns AWS::EC2::VPCPeeringConnection to implement the ResourceProperties interface
func (s EC2VPCPeeringConnection) CfnResourceType() string {
	return "AWS::EC2::VPCPeeringConnection"
}

// EC2VPNConnection represents AWS::EC2::VPNConnection
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-vpn-connection.html
type EC2VPNConnection struct {
	// The type of VPN connection this virtual private gateway supports.
	Type *StringExpr `json:"Type,omitempty"`

	// The ID of the customer gateway. This can either be an embedded JSON
	// object or a reference to a Gateway ID.
	CustomerGatewayId *StringExpr `json:"CustomerGatewayId,omitempty"`

	// Indicates whether the VPN connection requires static routes.
	StaticRoutesOnly *BoolExpr `json:"StaticRoutesOnly,omitempty"`

	// The tags that you want to attach to the resource.
	Tags []ResourceTag `json:"Tags,omitempty"`

	// The ID of the virtual private gateway. This can either be an embedded
	// JSON object or a reference to a Gateway ID.
	VpnGatewayId *StringExpr `json:"VpnGatewayId,omitempty"`
}

// CfnResourceType returns AWS::EC2::VPNConnection to implement the ResourceProperties interface
func (s EC2VPNConnection) CfnResourceType() string {
	return "AWS::EC2::VPNConnection"
}

// EC2VPNConnectionRoute represents AWS::EC2::VPNConnectionRoute
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-vpn-connection-route.html
type EC2VPNConnectionRoute struct {
	// The CIDR block that is associated with the local subnet of the
	// customer network.
	DestinationCidrBlock *StringExpr `json:"DestinationCidrBlock,omitempty"`

	// The ID of the VPN connection.
	VpnConnectionId *StringExpr `json:"VpnConnectionId,omitempty"`
}

// CfnResourceType returns AWS::EC2::VPNConnectionRoute to implement the ResourceProperties interface
func (s EC2VPNConnectionRoute) CfnResourceType() string {
	return "AWS::EC2::VPNConnectionRoute"
}

// EC2VPNGateway represents AWS::EC2::VPNGateway
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-vpn-gateway.html
type EC2VPNGateway struct {
	// The type of VPN connection this virtual private gateway supports. The
	// only valid value is "ipsec.1".
	Type *StringExpr `json:"Type,omitempty"`

	// An arbitrary set of tags (key–value pairs) for this resource.
	Tags []ResourceTag `json:"Tags,omitempty"`
}

// CfnResourceType returns AWS::EC2::VPNGateway to implement the ResourceProperties interface
func (s EC2VPNGateway) CfnResourceType() string {
	return "AWS::EC2::VPNGateway"
}

// EC2VPNGatewayRoutePropagation represents AWS::EC2::VPNGatewayRoutePropagation
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-vpn-gatewayrouteprop.html
type EC2VPNGatewayRoutePropagation struct {
	// A list of routing table IDs that are associated with a VPC. The
	// routing tables must be associated with the same VPC that the virtual
	// private gateway is attached to.
	RouteTableIds interface{} `json:"RouteTableIds,omitempty"`

	// The ID of the virtual private gateway that is attached to a VPC. The
	// virtual private gateway must be attached to the same VPC that the
	// routing tables are associated with.
	VpnGatewayId *StringExpr `json:"VpnGatewayId,omitempty"`
}

// CfnResourceType returns AWS::EC2::VPNGatewayRoutePropagation to implement the ResourceProperties interface
func (s EC2VPNGatewayRoutePropagation) CfnResourceType() string {
	return "AWS::EC2::VPNGatewayRoutePropagation"
}

// ECRRepository represents AWS::ECR::Repository
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ecr-repository.html
type ECRRepository struct {
	// A name for the image repository. If you don't specify a name, AWS
	// CloudFormation generates a unique physical ID and uses that ID for the
	// repository name. For more information, see Name Type.
	RepositoryName *StringExpr `json:"RepositoryName,omitempty"`

	// A policy that controls who has access to the repository and which
	// actions they can perform on it. For more information, see Amazon ECR
	// Repository Policies in the Amazon EC2 Container Registry User Guide.
	RepositoryPolicyText interface{} `json:"RepositoryPolicyText,omitempty"`
}

// CfnResourceType returns AWS::ECR::Repository to implement the ResourceProperties interface
func (s ECRRepository) CfnResourceType() string {
	return "AWS::ECR::Repository"
}

// ECSCluster represents AWS::ECS::Cluster
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ecs-cluster.html
type ECSCluster struct {
	// A name for the cluster. If you don't specify a name, AWS
	// CloudFormation generates a unique physical ID for the name. For more
	// information, see Name Type.
	ClusterName *StringExpr `json:"ClusterName,omitempty"`
}

// CfnResourceType returns AWS::ECS::Cluster to implement the ResourceProperties interface
func (s ECSCluster) CfnResourceType() string {
	return "AWS::ECS::Cluster"
}

// ECSService represents AWS::ECS::Service
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ecs-service.html
type ECSService struct {
	// The name or Amazon Resource Name (ARN) of the cluster that you want to
	// run your service on. If you do not specify a cluster, Amazon ECS uses
	// the default cluster.
	Cluster *StringExpr `json:"Cluster,omitempty"`

	// Configures how many tasks run during a deployment.
	DeploymentConfiguration *EC2ContainerServiceServiceDeploymentConfiguration `json:"DeploymentConfiguration,omitempty"`

	// The number of simultaneous tasks, which you specify by using the
	// TaskDefinition property, that you want to run on the cluster.
	DesiredCount *IntegerExpr `json:"DesiredCount,omitempty"`

	// A list of load balancer objects to associate with the cluster. For
	// information about the number of load balancers you can specify per
	// service, see Service Load Balancing in the Amazon EC2 Container
	// Service Developer Guide.
	LoadBalancers *EC2ContainerServiceServiceLoadBalancersList `json:"LoadBalancers,omitempty"`

	// The name or ARN of an AWS Identity and Access Management (IAM) role
	// that allows your Amazon ECS container agent to make calls to your load
	// balancer.
	Role *StringExpr `json:"Role,omitempty"`

	// The ARN of the task definition (including the revision number) that
	// you want to run on the cluster, such as
	// arn:aws:ecs:us-east-1:123456789012:task-definition/mytask:3. You can't
	// use :latest to specify a revision because it's ambiguous. For example,
	// if AWS CloudFormation needed to rollback an update, it wouldn't know
	// which revision to rollback to.
	TaskDefinition *StringExpr `json:"TaskDefinition,omitempty"`
}

// CfnResourceType returns AWS::ECS::Service to implement the ResourceProperties interface
func (s ECSService) CfnResourceType() string {
	return "AWS::ECS::Service"
}

// ECSTaskDefinition represents AWS::ECS::TaskDefinition
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ecs-taskdefinition.html
type ECSTaskDefinition struct {
	// A list of container definitions in JSON format that describe the
	// containers that make up your task.
	ContainerDefinitions *EC2ContainerServiceTaskDefinitionContainerDefinitionsList `json:"ContainerDefinitions,omitempty"`

	// The name of a family that this task definition is registered to. A
	// family groups multiple versions of a task definition. Amazon ECS gives
	// the first task definition that you registered to a family a revision
	// number of 1. Amazon ECS gives sequential revision numbers to each task
	// definition that you add.
	Family *StringExpr `json:"Family,omitempty"`

	// The Amazon Resource Name (ARN) of an AWS Identity and Access
	// Management (IAM) role that grants containers in the task permission to
	// call AWS APIs on your behalf. For more information, see IAM Roles for
	// Tasks in the Amazon EC2 Container Service Developer Guide.
	TaskRoleArn *StringExpr `json:"TaskRoleArn,omitempty"`

	// A list of volume definitions in JSON format for volumes that you can
	// use in your container definitions.
	Volumes *EC2ContainerServiceTaskDefinitionVolumesList `json:"Volumes,omitempty"`
}

// CfnResourceType returns AWS::ECS::TaskDefinition to implement the ResourceProperties interface
func (s ECSTaskDefinition) CfnResourceType() string {
	return "AWS::ECS::TaskDefinition"
}

// EFSFileSystem represents AWS::EFS::FileSystem
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-efs-filesystem.html
type EFSFileSystem struct {
	// Tags to associate with the file system.
	FileSystemTags *ElasticFileSystemFileSystemFileSystemTagsList `json:"FileSystemTags,omitempty"`

	// The performance mode of the file system. For valid values, see the
	// PerformanceMode parameter for the CreateFileSystem action in the
	// Amazon Elastic File System User Guide.
	PerformanceMode *StringExpr `json:"PerformanceMode,omitempty"`
}

// CfnResourceType returns AWS::EFS::FileSystem to implement the ResourceProperties interface
func (s EFSFileSystem) CfnResourceType() string {
	return "AWS::EFS::FileSystem"
}

// EFSMountTarget represents AWS::EFS::MountTarget
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-efs-mounttarget.html
type EFSMountTarget struct {
	// The ID of the file system for which you want to create the mount
	// target.
	FileSystemId *StringExpr `json:"FileSystemId,omitempty"`

	// An IPv4 address that is within the address range of the subnet that is
	// specified in the SubnetId property. If you don't specify an IP
	// address, Amazon EFS automatically assigns an address that is within
	// the range of the subnet.
	IpAddress *StringExpr `json:"IpAddress,omitempty"`

	// A maximum of five VPC security group IDs that are in the same VPC as
	// the subnet that is specified in the SubnetId property. For more
	// information about security groups and mount targets, see Security in
	// the Amazon Elastic File System User Guide.
	SecurityGroups *StringListExpr `json:"SecurityGroups,omitempty"`

	// The ID of the subnet in which you want to add the mount target.
	SubnetId *StringExpr `json:"SubnetId,omitempty"`
}

// CfnResourceType returns AWS::EFS::MountTarget to implement the ResourceProperties interface
func (s EFSMountTarget) CfnResourceType() string {
	return "AWS::EFS::MountTarget"
}

// ElastiCacheCacheCluster represents AWS::ElastiCache::CacheCluster
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-elasticache-cache-cluster.html
type ElastiCacheCacheCluster struct {
	// Indicates that minor engine upgrades will be applied automatically to
	// the cache cluster during the maintenance window.
	AutoMinorVersionUpgrade *BoolExpr `json:"AutoMinorVersionUpgrade,omitempty"`

	// For Memcached cache clusters, indicates whether the nodes are created
	// in a single Availability Zone or across multiple Availability Zones in
	// the cluster's region. For valid values, see CreateCacheCluster in the
	// Amazon ElastiCache API Reference.
	AZMode *StringExpr `json:"AZMode,omitempty"`

	// The compute and memory capacity of nodes in a cache cluster.
	CacheNodeType *StringExpr `json:"CacheNodeType,omitempty"`

	// The name of the cache parameter group that is associated with this
	// cache cluster.
	CacheParameterGroupName *StringExpr `json:"CacheParameterGroupName,omitempty"`

	// A list of cache security group names that are associated with this
	// cache cluster. If your cache cluster is in a VPC, specify the
	// VpcSecurityGroupIds property instead.
	CacheSecurityGroupNames *StringListExpr `json:"CacheSecurityGroupNames,omitempty"`

	// The cache subnet group that you associate with a cache cluster.
	CacheSubnetGroupName *StringExpr `json:"CacheSubnetGroupName,omitempty"`

	// A name for the cache cluster. If you don't specify a name, AWS
	// CloudFormation generates a unique physical ID and uses that ID for the
	// cache cluster. For more information, see Name Type.
	ClusterName *StringExpr `json:"ClusterName,omitempty"`

	// The name of the cache engine to be used for this cache cluster, such
	// as memcached or redis.
	Engine *StringExpr `json:"Engine,omitempty"`

	// The version of the cache engine to be used for this cluster.
	EngineVersion *StringExpr `json:"EngineVersion,omitempty"`

	// The Amazon Resource Name (ARN) of the Amazon Simple Notification
	// Service (SNS) topic to which notifications will be sent.
	NotificationTopicArn *StringExpr `json:"NotificationTopicArn,omitempty"`

	// The number of cache nodes that the cache cluster should have.
	NumCacheNodes *StringExpr `json:"NumCacheNodes,omitempty"`

	// The port number on which each of the cache nodes will accept
	// connections.
	Port *IntegerExpr `json:"Port,omitempty"`

	// The Amazon EC2 Availability Zone in which the cache cluster is
	// created.
	PreferredAvailabilityZone *StringExpr `json:"PreferredAvailabilityZone,omitempty"`

	// For Memcached cache clusters, the list of Availability Zones in which
	// cache nodes are created. The number of Availability Zones listed must
	// equal the number of cache nodes. For example, if you want to create
	// three nodes in two different Availability Zones, you can specify
	// ["us-east-1a", "us-east-1a", "us-east-1b"], which would create two
	// nodes in us-east-1a and one node in us-east-1b.
	PreferredAvailabilityZones *StringListExpr `json:"PreferredAvailabilityZones,omitempty"`

	// The weekly time range (in UTC) during which system maintenance can
	// occur.
	PreferredMaintenanceWindow *StringExpr `json:"PreferredMaintenanceWindow,omitempty"`

	// The ARN of the snapshot file that you want to use to seed a new Redis
	// cache cluster. If you manage a Redis instance outside of Amazon
	// ElastiCache, you can create a new cache cluster in ElastiCache by
	// using a snapshot file that is stored in an Amazon S3 bucket.
	SnapshotArns *StringListExpr `json:"SnapshotArns,omitempty"`

	// The name of a snapshot from which to restore data into a new Redis
	// cache cluster.
	SnapshotName *StringExpr `json:"SnapshotName,omitempty"`

	// For Redis cache clusters, the number of days for which ElastiCache
	// retains automatic snapshots before deleting them. For example, if you
	// set the value to 5, a snapshot that was taken today will be retained
	// for 5 days before being deleted.
	SnapshotRetentionLimit *IntegerExpr `json:"SnapshotRetentionLimit,omitempty"`

	// For Redis cache clusters, the daily time range (in UTC) during which
	// ElastiCache will begin taking a daily snapshot of your node group. For
	// example, you can specify 05:00-09:00.
	SnapshotWindow *StringExpr `json:"SnapshotWindow,omitempty"`

	// An arbitrary set of tags (key–value pairs) for this cache cluster.
	Tags []ResourceTag `json:"Tags,omitempty"`

	// A list of VPC security group IDs. If your cache cluster isn't in a
	// VPC, specify the CacheSecurityGroupNames property instead.
	VpcSecurityGroupIds *StringListExpr `json:"VpcSecurityGroupIds,omitempty"`
}

// CfnResourceType returns AWS::ElastiCache::CacheCluster to implement the ResourceProperties interface
func (s ElastiCacheCacheCluster) CfnResourceType() string {
	return "AWS::ElastiCache::CacheCluster"
}

// ElastiCacheParameterGroup represents AWS::ElastiCache::ParameterGroup
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-elasticache-parameter-group.html
type ElastiCacheParameterGroup struct {
	// The name of the cache parameter group family that the cache parameter
	// group can be used with.
	CacheParameterGroupFamily *StringExpr `json:"CacheParameterGroupFamily,omitempty"`

	// The description for the Cache Parameter Group.
	Description *StringExpr `json:"Description,omitempty"`

	// A comma-delimited list of parameter name/value pairs. For more
	// information, go to ModifyCacheParameterGroup in the Amazon ElastiCache
	// API Reference Guide.
	Properties interface{} `json:"Properties,omitempty"`
}

// CfnResourceType returns AWS::ElastiCache::ParameterGroup to implement the ResourceProperties interface
func (s ElastiCacheParameterGroup) CfnResourceType() string {
	return "AWS::ElastiCache::ParameterGroup"
}

// ElastiCacheReplicationGroup represents AWS::ElastiCache::ReplicationGroup
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-elasticache-replicationgroup.html
type ElastiCacheReplicationGroup struct {
	// Indicates whether Multi-AZ is enabled. When Multi-AZ is enabled, a
	// read-only replica is automatically promoted to a read-write primary
	// cluster if the existing primary cluster fails. If you specify true,
	// you must specify a value greater than 1 for the NumCacheNodes
	// property. By default, AWS CloudFormation sets the value to true.
	AutomaticFailoverEnabled *BoolExpr `json:"AutomaticFailoverEnabled,omitempty"`

	// Currently, this property isn't used by ElastiCache.
	AutoMinorVersionUpgrade *BoolExpr `json:"AutoMinorVersionUpgrade,omitempty"`

	// The compute and memory capacity of nodes in the node group. To see
	// valid values, see CreateReplicationGroup in the Amazon ElastiCache API
	// Reference Guide.
	CacheNodeType *StringExpr `json:"CacheNodeType,omitempty"`

	// The name of the parameter group to associate with this replication
	// group. For valid and default values, see CreateReplicationGroup in the
	// Amazon ElastiCache API Reference Guide.
	CacheParameterGroupName *StringExpr `json:"CacheParameterGroupName,omitempty"`

	// A list of cache security group names to associate with this
	// replication group.
	CacheSecurityGroupNames *StringListExpr `json:"CacheSecurityGroupNames,omitempty"`

	// The name of a cache subnet group to use for this replication group.
	CacheSubnetGroupName *StringExpr `json:"CacheSubnetGroupName,omitempty"`

	// The name of the cache engine to use for the cache clusters in this
	// replication group. Currently, you can specify only redis.
	Engine *StringExpr `json:"Engine,omitempty"`

	// The version number of the cache engine to use for the cache clusters
	// in this replication group.
	EngineVersion *StringExpr `json:"EngineVersion,omitempty"`

	// Configuration options for the node group (shard).
	NodeGroupConfiguration *ElastiCacheReplicationGroupNodeGroupConfigurationList `json:"NodeGroupConfiguration,omitempty"`

	// The Amazon Resource Name (ARN) of the Amazon Simple Notification
	// Service topic to which notifications are sent.
	NotificationTopicArn *StringExpr `json:"NotificationTopicArn,omitempty"`

	// The number of cache clusters for this replication group. If automatic
	// failover is enabled, you must specify a value greater than 1. For
	// valid values, see CreateReplicationGroup in the Amazon ElastiCache API
	// Reference Guide.
	NumCacheClusters *IntegerExpr `json:"NumCacheClusters,omitempty"`

	// The number of node groups (shards) for this Redis (clustered mode
	// enabled) replication group. For Redis (clustered mode disabled), omit
	// this property.
	NumNodeGroups *IntegerExpr `json:"NumNodeGroups,omitempty"`

	// The port number on which each member of the replication group accepts
	// connections.
	Port *IntegerExpr `json:"Port,omitempty"`

	// A list of Availability Zones (AZs) in which the cache clusters in this
	// replication group are created.
	PreferredCacheClusterAZs *StringListExpr `json:"PreferredCacheClusterAZs,omitempty"`

	// The weekly time range during which system maintenance can occur. Use
	// the following format to specify a time range: ddd:hh24:mi-ddd:hh24:mi
	// (24H Clock UTC). For example, you can specify sun:22:00-sun:23:30 for
	// Sunday from 10 PM to 11:30 PM.
	PreferredMaintenanceWindow *StringExpr `json:"PreferredMaintenanceWindow,omitempty"`

	// The cache cluster that ElastiCache uses as the primary cluster for the
	// replication group. The cache cluster must have a status of available.
	PrimaryClusterId *StringExpr `json:"PrimaryClusterId,omitempty"`

	// The number of replica nodes in each node group (shard). For valid
	// values, see CreateReplicationGroup in the Amazon ElastiCache API
	// Reference Guide.
	ReplicasPerNodeGroup *IntegerExpr `json:"ReplicasPerNodeGroup,omitempty"`

	// The description of the replication group.
	ReplicationGroupDescription *StringExpr `json:"ReplicationGroupDescription,omitempty"`

	// An ID for the replication group. If you don't specify an ID, AWS
	// CloudFormation generates a unique physical ID. For more information,
	// see Name Type.
	ReplicationGroupId *StringExpr `json:"ReplicationGroupId,omitempty"`

	// A list of Amazon Virtual Private Cloud (Amazon VPC) security groups to
	// associate with this replication group.
	SecurityGroupIds *StringListExpr `json:"SecurityGroupIds,omitempty"`

	// A single-element string list that specifies an ARN of a Redis .rdb
	// snapshot file that is stored in Amazon Simple Storage Service (Amazon
	// S3). The snapshot file populates the node group. The Amazon S3 object
	// name in the ARN cannot contain commas. For example, you can specify
	// arn:aws:s3:::my_bucket/snapshot1.rdb.
	SnapshotArns *StringListExpr `json:"SnapshotArns,omitempty"`

	// The name of a snapshot from which to restore data into the replication
	// group.
	SnapshotName *StringExpr `json:"SnapshotName,omitempty"`

	// The number of days that ElastiCache retains automatic snapshots before
	// deleting them.
	SnapshotRetentionLimit *IntegerExpr `json:"SnapshotRetentionLimit,omitempty"`

	// The ID of the cache cluster that ElastiCache uses as the daily
	// snapshot source for the replication group.
	SnapshottingClusterId *StringExpr `json:"SnapshottingClusterId,omitempty"`

	// The time range (in UTC) when ElastiCache takes a daily snapshot of
	// your node group that you specified in the SnapshottingClusterId
	// property. For example, you can specify 05:00-09:00.
	SnapshotWindow *StringExpr `json:"SnapshotWindow,omitempty"`

	// An arbitrary set of tags (key–value pairs) for this replication
	// group.
	Tags []ResourceTag `json:"Tags,omitempty"`
}

// CfnResourceType returns AWS::ElastiCache::ReplicationGroup to implement the ResourceProperties interface
func (s ElastiCacheReplicationGroup) CfnResourceType() string {
	return "AWS::ElastiCache::ReplicationGroup"
}

// ElastiCacheSecurityGroup represents AWS::ElastiCache::SecurityGroup
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-elasticache-security-group.html
type ElastiCacheSecurityGroup struct {
	// A description for the cache security group.
	Description *StringExpr `json:"Description,omitempty"`
}

// CfnResourceType returns AWS::ElastiCache::SecurityGroup to implement the ResourceProperties interface
func (s ElastiCacheSecurityGroup) CfnResourceType() string {
	return "AWS::ElastiCache::SecurityGroup"
}

// ElastiCacheSecurityGroupIngress represents AWS::ElastiCache::SecurityGroupIngress
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-elasticache-security-group-ingress.html
type ElastiCacheSecurityGroupIngress struct {
	// The name of the Cache Security Group to authorize.
	CacheSecurityGroupName *StringExpr `json:"CacheSecurityGroupName,omitempty"`

	// Name of the EC2 Security Group to include in the authorization.
	EC2SecurityGroupName *StringExpr `json:"EC2SecurityGroupName,omitempty"`

	// Specifies the AWS Account ID of the owner of the EC2 security group
	// specified in the EC2SecurityGroupName property. The AWS access key ID
	// is not an acceptable value.
	EC2SecurityGroupOwnerId *StringExpr `json:"EC2SecurityGroupOwnerId,omitempty"`
}

// CfnResourceType returns AWS::ElastiCache::SecurityGroupIngress to implement the ResourceProperties interface
func (s ElastiCacheSecurityGroupIngress) CfnResourceType() string {
	return "AWS::ElastiCache::SecurityGroupIngress"
}

// ElastiCacheSubnetGroup represents AWS::ElastiCache::SubnetGroup
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-elasticache-subnetgroup.html
type ElastiCacheSubnetGroup struct {
}

// CfnResourceType returns AWS::ElastiCache::SubnetGroup  to implement the ResourceProperties interface
func (s ElastiCacheSubnetGroup) CfnResourceType() string {
	return "AWS::ElastiCache::SubnetGroup "
}

// ElasticBeanstalkApplication represents AWS::ElasticBeanstalk::Application
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-beanstalk.html
type ElasticBeanstalkApplication struct {
	// A name for the Elastic Beanstalk application. If you don't specify a
	// name, AWS CloudFormation generates a unique physical ID and uses that
	// ID for the application name. For more information, see Name Type.
	ApplicationName *StringExpr `json:"ApplicationName,omitempty"`

	// An optional description of this application.
	Description *StringExpr `json:"Description,omitempty"`
}

// CfnResourceType returns AWS::ElasticBeanstalk::Application to implement the ResourceProperties interface
func (s ElasticBeanstalkApplication) CfnResourceType() string {
	return "AWS::ElasticBeanstalk::Application"
}

// ElasticBeanstalkApplicationVersion represents AWS::ElasticBeanstalk::ApplicationVersion
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-beanstalk-version.html
type ElasticBeanstalkApplicationVersion struct {
	// Name of the Elastic Beanstalk application that is associated with this
	// application version.
	ApplicationName *StringExpr `json:"ApplicationName,omitempty"`

	// A description of this application version.
	Description *StringExpr `json:"Description,omitempty"`

	// The location of the source bundle for this version.
	SourceBundle *ElasticBeanstalkSourceBundle `json:"SourceBundle,omitempty"`
}

// CfnResourceType returns AWS::ElasticBeanstalk::ApplicationVersion to implement the ResourceProperties interface
func (s ElasticBeanstalkApplicationVersion) CfnResourceType() string {
	return "AWS::ElasticBeanstalk::ApplicationVersion"
}

// ElasticBeanstalkConfigurationTemplate represents AWS::ElasticBeanstalk::ConfigurationTemplate
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-beanstalk-configurationtemplate.html
type ElasticBeanstalkConfigurationTemplate struct {
	// Name of the Elastic Beanstalk application that is associated with this
	// configuration template.
	ApplicationName *StringExpr `json:"ApplicationName,omitempty"`

	// An optional description for this configuration.
	Description *StringExpr `json:"Description,omitempty"`

	// An environment whose settings you want to use to create the
	// configuration template. You must specify this property if you don't
	// specify the SolutionStackName or SourceConfiguration properties.
	EnvironmentId *StringExpr `json:"EnvironmentId,omitempty"`

	// A list of OptionSettings for this Elastic Beanstalk configuration. For
	// a complete list of Elastic Beanstalk configuration options, see Option
	// Values, in the AWS Elastic Beanstalk Developer Guide.
	OptionSettings *ElasticBeanstalkOptionSettingsList `json:"OptionSettings,omitempty"`

	// The name of an Elastic Beanstalk solution stack that this
	// configuration will use. A solution stack specifies the operating
	// system, architecture, and application server for a configuration
	// template, such as 64bit Amazon Linux 2013.09 running Tomcat 7 Java 7.
	// For more information, see Supported Platforms in the AWS Elastic
	// Beanstalk Developer Guide.
	SolutionStackName *StringExpr `json:"SolutionStackName,omitempty"`

	// A configuration template that is associated with another Elastic
	// Beanstalk application. If you specify the SolutionStackName property
	// and the SourceConfiguration property, the solution stack in the source
	// configuration template must match the value that you specified for the
	// SolutionStackName property.
	SourceConfiguration *ElasticBeanstalkSourceConfiguration `json:"SourceConfiguration,omitempty"`
}

// CfnResourceType returns AWS::ElasticBeanstalk::ConfigurationTemplate to implement the ResourceProperties interface
func (s ElasticBeanstalkConfigurationTemplate) CfnResourceType() string {
	return "AWS::ElasticBeanstalk::ConfigurationTemplate"
}

// ElasticBeanstalkEnvironment represents AWS::ElasticBeanstalk::Environment
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-beanstalk-environment.html
type ElasticBeanstalkEnvironment struct {
	// The name of the application that is associated with this environment.
	ApplicationName *StringExpr `json:"ApplicationName,omitempty"`

	// A prefix for your Elastic Beanstalk environment URL.
	CNAMEPrefix *StringExpr `json:"CNAMEPrefix,omitempty"`

	// A description that helps you identify this environment.
	Description *StringExpr `json:"Description,omitempty"`

	// A name for the Elastic Beanstalk environment. If you don't specify a
	// name, AWS CloudFormation generates a unique physical ID and uses that
	// ID for the environment name. For more information, see Name Type.
	EnvironmentName *StringExpr `json:"EnvironmentName,omitempty"`

	// Key-value pairs defining configuration options for this environment.
	// These options override the values that are defined in the solution
	// stack or the configuration template. If you remove any options during
	// a stack update, the removed options revert to default values.
	OptionSettings *ElasticBeanstalkOptionSettingsList `json:"OptionSettings,omitempty"`

	// The name of an Elastic Beanstalk solution stack that this
	// configuration will use. For more information, see Supported Platforms
	// in the AWS Elastic Beanstalk Developer Guide. You must specify either
	// this parameter or an Elastic Beanstalk configuration template name.
	SolutionStackName *StringExpr `json:"SolutionStackName,omitempty"`

	// An arbitrary set of tags (key–value pairs) for this environment.
	Tags []ResourceTag `json:"Tags,omitempty"`

	// The name of the Elastic Beanstalk configuration template to use with
	// the environment. You must specify either this parameter or a solution
	// stack name.
	TemplateName *StringExpr `json:"TemplateName,omitempty"`

	// Specifies the tier to use in creating this environment. The
	// environment tier that you choose determines whether Elastic Beanstalk
	// provisions resources to support a web application that handles HTTP(S)
	// requests or a web application that handles background-processing
	// tasks.
	Tier *ElasticBeanstalkEnvironmentTier `json:"Tier,omitempty"`

	// The version to associate with the environment.
	VersionLabel *StringExpr `json:"VersionLabel,omitempty"`
}

// CfnResourceType returns AWS::ElasticBeanstalk::Environment to implement the ResourceProperties interface
func (s ElasticBeanstalkEnvironment) CfnResourceType() string {
	return "AWS::ElasticBeanstalk::Environment"
}

// ElasticLoadBalancingLoadBalancer represents AWS::ElasticLoadBalancing::LoadBalancer
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-elb.html
type ElasticLoadBalancingLoadBalancer struct {
	// Captures detailed information for all requests made to your load
	// balancer, such as the time a request was received, client’s IP
	// address, latencies, request path, and server responses.
	AccessLoggingPolicy *ElasticLoadBalancingAccessLoggingPolicy `json:"AccessLoggingPolicy,omitempty"`

	// Generates one or more stickiness policies with sticky session
	// lifetimes that follow that of an application-generated cookie. These
	// policies can be associated only with HTTP/HTTPS listeners.
	AppCookieStickinessPolicy *ElasticLoadBalancingAppCookieStickinessPolicyList `json:"AppCookieStickinessPolicy,omitempty"`

	// The Availability Zones in which to create the load balancer. You can
	// specify the AvailabilityZones or Subnets property, but not both.
	AvailabilityZones *StringListExpr `json:"AvailabilityZones,omitempty"`

	// Whether deregistered or unhealthy instances can complete all in-flight
	// requests.
	ConnectionDrainingPolicy *ElasticLoadBalancingConnectionDrainingPolicy `json:"ConnectionDrainingPolicy,omitempty"`

	// Specifies how long front-end and back-end connections of your load
	// balancer can remain idle.
	ConnectionSettings *ElasticLoadBalancingConnectionSettings `json:"ConnectionSettings,omitempty"`

	// Whether cross-zone load balancing is enabled for the load balancer.
	// With cross-zone load balancing, your load balancer nodes route traffic
	// to the back-end instances across all Availability Zones. By default
	// the CrossZone property is false.
	CrossZone *BoolExpr `json:"CrossZone,omitempty"`

	// Application health check for the instances.
	HealthCheck *ElasticLoadBalancingHealthCheck `json:"HealthCheck,omitempty"`

	// A list of EC2 instance IDs for the load balancer.
	Instances *StringListExpr `json:"Instances,omitempty"`

	// Generates a stickiness policy with sticky session lifetimes controlled
	// by the lifetime of the browser (user-agent), or by a specified
	// expiration period. This policy can be associated only with HTTP/HTTPS
	// listeners.
	LBCookieStickinessPolicy *ElasticLoadBalancingLBCookieStickinessPolicyList `json:"LBCookieStickinessPolicy,omitempty"`

	// A name for the load balancer. For valid values, see the
	// LoadBalancerName parameter for the CreateLoadBalancer action in the
	// Elastic Load Balancing API Reference version 2012-06-01.
	LoadBalancerName *StringExpr `json:"LoadBalancerName,omitempty"`

	// One or more listeners for this load balancer. Each listener must be
	// registered for a specific port, and you cannot have more than one
	// listener for a given port.
	Listeners *ElasticLoadBalancingListenerList `json:"Listeners,omitempty"`

	// A list of elastic load balancing policies to apply to this elastic
	// load balancer. Specify only back-end server policies. For more
	// information, see DescribeLoadBalancerPolicyTypes in the Elastic Load
	// Balancing API Reference version 2012-06-01.
	Policies *ElasticLoadBalancingPolicyList `json:"Policies,omitempty"`

	// For load balancers attached to an Amazon VPC, this parameter can be
	// used to specify the type of load balancer to use. Specify internal to
	// create an internal load balancer with a DNS name that resolves to
	// private IP addresses or internet-facing to create a load balancer with
	// a publicly resolvable DNS name, which resolves to public IP addresses.
	Scheme *StringExpr `json:"Scheme,omitempty"`

	// Required: No
	SecurityGroups interface{} `json:"SecurityGroups,omitempty"`

	// A list of subnet IDs in your virtual private cloud (VPC) to attach to
	// your load balancer. Do not specify multiple subnets that are in the
	// same Availability Zone. You can specify the AvailabilityZones or
	// Subnets property, but not both.
	Subnets *StringListExpr `json:"Subnets,omitempty"`

	// An arbitrary set of tags (key-value pairs) for this load balancer.
	Tags []ResourceTag `json:"Tags,omitempty"`
}

// CfnResourceType returns AWS::ElasticLoadBalancing::LoadBalancer to implement the ResourceProperties interface
func (s ElasticLoadBalancingLoadBalancer) CfnResourceType() string {
	return "AWS::ElasticLoadBalancing::LoadBalancer"
}

// ElasticLoadBalancingV2Listener represents AWS::ElasticLoadBalancingV2::Listener
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-elasticloadbalancingv2-listener.html
type ElasticLoadBalancingV2Listener struct {
	// The SSL server certificate for the listener. With a certificate, you
	// can encrypt traffic between the load balancer and the clients that
	// initiate HTTPS sessions, and traffic between the load balancer and
	// your targets.
	Certificates *ElasticLoadBalancingListenerCertificatesList `json:"Certificates,omitempty"`

	// The default actions that the listener takes when handling incoming
	// requests.
	DefaultActions *ElasticLoadBalancingListenerDefaultActionsList `json:"DefaultActions,omitempty"`

	// The Amazon Resource Name (ARN) of the load balancer to associate with
	// the listener.
	LoadBalancerArn *StringExpr `json:"LoadBalancerArn,omitempty"`

	// The port on which the listener listens for requests.
	Port *IntegerExpr `json:"Port,omitempty"`

	// The protocol that clients must use to send requests to the listener.
	Protocol *StringExpr `json:"Protocol,omitempty"`

	// The security policy that defines the ciphers and protocols that the
	// load balancer supports.
	SslPolicy *StringExpr `json:"SslPolicy,omitempty"`
}

// CfnResourceType returns AWS::ElasticLoadBalancingV2::Listener to implement the ResourceProperties interface
func (s ElasticLoadBalancingV2Listener) CfnResourceType() string {
	return "AWS::ElasticLoadBalancingV2::Listener"
}

// ElasticLoadBalancingV2ListenerRule represents AWS::ElasticLoadBalancingV2::ListenerRule
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-elasticloadbalancingv2-listenerrule.html
type ElasticLoadBalancingV2ListenerRule struct {
	// The action that the listener takes when a request meets the specified
	// condition.
	Actions *ElasticLoadBalancingListenerRuleActionsList `json:"Actions,omitempty"`

	// The conditions under which a rule takes effect.
	Conditions *ElasticLoadBalancingListenerRuleConditionsList `json:"Conditions,omitempty"`

	// The Amazon Resource Name (ARN) of the listener that the rule applies
	// to.
	ListenerArn *StringExpr `json:"ListenerArn,omitempty"`

	// The priority for the rule. Elastic Load Balancing evaluates rules in
	// priority order, from the lowest value to the highest value. If a
	// request satisfies a rule, Elastic Load Balancing ignores all
	// subsequent rules.
	Priority *IntegerExpr `json:"Priority,omitempty"`
}

// CfnResourceType returns AWS::ElasticLoadBalancingV2::ListenerRule to implement the ResourceProperties interface
func (s ElasticLoadBalancingV2ListenerRule) CfnResourceType() string {
	return "AWS::ElasticLoadBalancingV2::ListenerRule"
}

// ElasticLoadBalancingV2LoadBalancer represents AWS::ElasticLoadBalancingV2::LoadBalancer
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-elasticloadbalancingv2-loadbalancer.html
type ElasticLoadBalancingV2LoadBalancer struct {
	// Specifies the load balancer configuration.
	LoadBalancerAttributes *ElasticLoadBalancingLoadBalancerLoadBalancerAttributesList `json:"LoadBalancerAttributes,omitempty"`

	// Specifies a name for the load balancer. This name must be unique
	// within your AWS account and can have a maximum of 32 alphanumeric
	// characters and hyphens. A name can't begin or end with a hyphen.
	Name *StringExpr `json:"Name,omitempty"`

	// Specifies whether the load balancer is internal or Internet-facing. An
	// internal load balancer routes requests to targets using private IP
	// addresses. An Internet-facing load balancer routes requests from
	// clients over the Internet to targets in your public subnets.
	Scheme *StringExpr `json:"Scheme,omitempty"`

	// Specifies a list of the IDs of the security groups to assign to the
	// load balancer.
	SecurityGroups *StringListExpr `json:"SecurityGroups,omitempty"`

	// Specifies a list of at least two IDs of the subnets to associate with
	// the load balancer. The subnets must be in different Availability
	// Zones.
	Subnets *StringListExpr `json:"Subnets,omitempty"`

	// Specifies an arbitrary set of tags (key–value pairs) to associate
	// with this load balancer. Use tags to manage your resources.
	Tags []ResourceTag `json:"Tags,omitempty"`
}

// CfnResourceType returns AWS::ElasticLoadBalancingV2::LoadBalancer to implement the ResourceProperties interface
func (s ElasticLoadBalancingV2LoadBalancer) CfnResourceType() string {
	return "AWS::ElasticLoadBalancingV2::LoadBalancer"
}

// ElasticLoadBalancingV2TargetGroup represents AWS::ElasticLoadBalancingV2::TargetGroup
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-elasticloadbalancingv2-targetgroup.html
type ElasticLoadBalancingV2TargetGroup struct {
	// The approximate number of seconds between health checks for an
	// individual target.
	HealthCheckIntervalSeconds *IntegerExpr `json:"HealthCheckIntervalSeconds,omitempty"`

	// The ping path destination where Elastic Load Balancing sends health
	// check requests.
	HealthCheckPath *StringExpr `json:"HealthCheckPath,omitempty"`

	// The port that the load balancer uses when performing health checks on
	// the targets.
	HealthCheckPort *StringExpr `json:"HealthCheckPort,omitempty"`

	// The protocol that the load balancer uses when performing health checks
	// on the targets, such as HTTP or HTTPS.
	HealthCheckProtocol *StringExpr `json:"HealthCheckProtocol,omitempty"`

	// The number of seconds to wait for a response before considering that a
	// health check has failed.
	HealthCheckTimeoutSeconds *IntegerExpr `json:"HealthCheckTimeoutSeconds,omitempty"`

	// The number of consecutive successful health checks that are required
	// before an unhealthy target is considered healthy.
	HealthyThresholdCount *IntegerExpr `json:"HealthyThresholdCount,omitempty"`

	// The HTTP codes that a healthy target uses when responding to a health
	// check.
	Matcher *ElasticLoadBalancingTargetGroupMatcher `json:"Matcher,omitempty"`

	// A name for the target group.
	Name *StringExpr `json:"Name,omitempty"`

	// The port on which the targets receive traffic.
	Port *IntegerExpr `json:"Port,omitempty"`

	// The protocol to use for routing traffic to the targets.
	Protocol *StringExpr `json:"Protocol,omitempty"`

	// An arbitrary set of tags (key–value pairs) for the target group. Use
	// tags to help manage resources.
	Tags []ResourceTag `json:"Tags,omitempty"`

	// Target group configurations.
	TargetGroupAttributes *ElasticLoadBalancingTargetGroupTargetGroupAttributesList `json:"TargetGroupAttributes,omitempty"`

	// The targets to add to this target group.
	Targets *ElasticLoadBalancingTargetGroupTargetDescriptionList `json:"Targets,omitempty"`

	// The number of consecutive failed health checks that are required
	// before a target is considered unhealthy.
	UnhealthyThresholdCount *IntegerExpr `json:"UnhealthyThresholdCount,omitempty"`

	// The ID of the VPC in which your targets are located.
	VpcId *StringExpr `json:"VpcId,omitempty"`
}

// CfnResourceType returns AWS::ElasticLoadBalancingV2::TargetGroup to implement the ResourceProperties interface
func (s ElasticLoadBalancingV2TargetGroup) CfnResourceType() string {
	return "AWS::ElasticLoadBalancingV2::TargetGroup"
}

// ElasticsearchDomain represents AWS::Elasticsearch::Domain
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-elasticsearch-domain.html
type ElasticsearchDomain struct {
	// An AWS Identity and Access Management (IAM) policy document that
	// specifies who can access the Amazon ES domain and their permissions.
	// For more information, see Configuring Access Policies in the Amazon
	// Elasticsearch Service Developer Guide.
	AccessPolicies interface{} `json:"AccessPolicies,omitempty"`

	// Additional options to specify for the Amazon ES domain. For more
	// information, see Configuring Advanced Options in the Amazon
	// Elasticsearch Service Developer Guide.
	AdvancedOptions interface{} `json:"AdvancedOptions,omitempty"`

	// A name for the Amazon ES domain. For valid values, see the DomainName
	// data type in the Amazon Elasticsearch Service Developer Guide.
	DomainName *StringExpr `json:"DomainName,omitempty"`

	// The configurations of Amazon Elastic Block Store (Amazon EBS) volumes
	// that are attached to data nodes in the Amazon ES domain. For more
	// information, see Configuring EBS-based Storage in the Amazon
	// Elasticsearch Service Developer Guide.
	EBSOptions *ElasticsearchServiceDomainEBSOptions `json:"EBSOptions,omitempty"`

	// The cluster configuration for the Amazon ES domain. You can specify
	// options such as the instance type and the number of instances. For
	// more information, see Configuring Amazon ES Domains in the Amazon
	// Elasticsearch Service Developer Guide.
	ElasticsearchClusterConfig *ElasticsearchServiceDomainElasticsearchClusterConfig `json:"ElasticsearchClusterConfig,omitempty"`

	// The version of Elasticsearch to use, such as 2.3. For information
	// about the versions that Amazon ES supports, see the
	// Elasticsearch-Version parameter for the CreateElasticsearchDomain
	// action in the Amazon Elasticsearch Service Developer Guide.
	ElasticsearchVersion *StringExpr `json:"ElasticsearchVersion,omitempty"`

	// The automated snapshot configuration for the Amazon ES domain indices.
	SnapshotOptions *ElasticsearchServiceDomainSnapshotOptions `json:"SnapshotOptions,omitempty"`

	// An arbitrary set of tags (key–value pairs) to associate with the
	// Amazon ES domain.
	Tags []ResourceTag `json:"Tags,omitempty"`
}

// CfnResourceType returns AWS::Elasticsearch::Domain to implement the ResourceProperties interface
func (s ElasticsearchDomain) CfnResourceType() string {
	return "AWS::Elasticsearch::Domain"
}

// EMRCluster represents AWS::EMR::Cluster
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-emr-cluster.html
type EMRCluster struct {
	// Additional features that you want to select.
	AdditionalInfo interface{} `json:"AdditionalInfo,omitempty"`

	// The software applications to deploy on the cluster, and the arguments
	// that Amazon EMR passes to those applications.
	Applications *EMRClusterApplicationList `json:"Applications,omitempty"`

	// A list of bootstrap actions that Amazon EMR runs before starting
	// applications on the cluster.
	BootstrapActions *EMRClusterBootstrapActionConfigList `json:"BootstrapActions,omitempty"`

	// The software configuration of the Amazon EMR cluster.
	Configurations *EMRClusterConfigurationList `json:"Configurations,omitempty"`

	// Configures the EC2 instances that will run jobs in the Amazon EMR
	// cluster.
	Instances *EMRClusterJobFlowInstancesConfig `json:"Instances,omitempty"`

	// Also called instance profile and EC2 role. Accepts an instance profile
	// associated with the role that you want to use. All EC2 instances in
	// the cluster assume this role.
	JobFlowRole *StringExpr `json:"JobFlowRole,omitempty"`

	// An S3 bucket location to which Amazon EMR writes logs files from a job
	// flow. If you don't specify a value, Amazon EMR doesn't write any log
	// files.
	LogUri *StringExpr `json:"LogUri,omitempty"`

	// A name for the Amazon EMR cluster.
	Name *StringExpr `json:"Name,omitempty"`

	// The Amazon EMR software release label. A release is a set of software
	// applications and components that you can install and configure on an
	// Amazon EMR cluster. For more information, see About Amazon EMR
	// Releases in the Amazon EMR Release Guide.
	ReleaseLabel *StringExpr `json:"ReleaseLabel,omitempty"`

	// The IAM role that Amazon EMR assumes to access AWS resources on your
	// behalf. For more information, see Configure IAM Roles for Amazon EMR
	// in the Amazon EMR Management Guide.
	ServiceRole *StringExpr `json:"ServiceRole,omitempty"`

	// An arbitrary set of tags (key–value pairs) to help you identify the
	// Amazon EMR cluster.
	Tags []ResourceTag `json:"Tags,omitempty"`

	// Indicates whether the instances in the cluster are visible to all IAM
	// users in the AWS account. If you specify true, all IAM users can view
	// and (if they have permissions) manage the instances. If you specify
	// false, only the IAM user that created the cluster can view and manage
	// it. By default, AWS CloudFormation sets this property to false.
	VisibleToAllUsers *BoolExpr `json:"VisibleToAllUsers,omitempty"`
}

// CfnResourceType returns AWS::EMR::Cluster to implement the ResourceProperties interface
func (s EMRCluster) CfnResourceType() string {
	return "AWS::EMR::Cluster"
}

// EMRInstanceGroupConfig represents AWS::EMR::InstanceGroupConfig
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-emr-instancegroupconfig.html
type EMRInstanceGroupConfig struct {
	// The bid price in USD for each EC2 instance in the instance group when
	// launching instances (nodes) as Spot Instances.
	BidPrice *StringExpr `json:"BidPrice,omitempty"`

	// A list of configurations to apply to this instance group. For more
	// information see, Configuring Applications in the Amazon EMR Release
	// Guide.
	Configurations *EMRClusterConfigurationList `json:"Configurations,omitempty"`

	// Configures Amazon Elastic Block Store (Amazon EBS) storage volumes to
	// attach to your instances.
	EbsConfiguration *EMREbsConfiguration `json:"EbsConfiguration,omitempty"`

	// The number of instances to launch in the instance group.
	InstanceCount *IntegerExpr `json:"InstanceCount,omitempty"`

	// The role of the servers in the Amazon EMR cluster, such as TASK. For
	// more information, see Instance Groups in the Amazon EMR Management
	// Guide.
	InstanceRole *StringExpr `json:"InstanceRole,omitempty"`

	// The EC2 instance type for all instances in the instance group. For
	// more information, see Instance Configurations in the Amazon EMR
	// Management Guide.
	InstanceType *StringExpr `json:"InstanceType,omitempty"`

	// The ID of an Amazon EMR cluster that you want to associate this
	// instance group with.
	JobFlowId *StringExpr `json:"JobFlowId,omitempty"`

	// The type of marketplace from which your instances are provisioned into
	// this group, either ON_DEMAND or SPOT. For more information, see Amazon
	// EC2 Purchasing Options.
	Market *StringExpr `json:"Market,omitempty"`

	// A name for the instance group.
	Name *StringExpr `json:"Name,omitempty"`
}

// CfnResourceType returns AWS::EMR::InstanceGroupConfig to implement the ResourceProperties interface
func (s EMRInstanceGroupConfig) CfnResourceType() string {
	return "AWS::EMR::InstanceGroupConfig"
}

// EMRStep represents AWS::EMR::Step
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-emr-step.html
type EMRStep struct {
	// The action to take if the job flow step fails. Currently, AWS
	// CloudFormation supports CONTINUE and CANCEL_AND_WAIT. For more
	// information, see Managing Cluster Termination in the Amazon EMR
	// Management Guide.
	ActionOnFailure *StringExpr `json:"ActionOnFailure,omitempty"`

	// The JAR file that includes the main function that Amazon EMR executes.
	HadoopJarStep *EMRStepHadoopJarStepConfig `json:"HadoopJarStep,omitempty"`

	// The ID of a cluster in which you want to run this job flow step.
	JobFlowId *StringExpr `json:"JobFlowId,omitempty"`

	// A name for the job flow step.
	Name *StringExpr `json:"Name,omitempty"`
}

// CfnResourceType returns AWS::EMR::Step to implement the ResourceProperties interface
func (s EMRStep) CfnResourceType() string {
	return "AWS::EMR::Step"
}

// EventsRule represents AWS::Events::Rule
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-events-rule.html
type EventsRule struct {
	// A description of the rule's purpose.
	Description *StringExpr `json:"Description,omitempty"`

	// Describes which events CloudWatch Events routes to the specified
	// target. These routed events are matched events. For more information,
	// see Events and Event Patterns in the Amazon CloudWatch User Guide.
	EventPattern interface{} `json:"EventPattern,omitempty"`

	// A name for the rule. If you don't specify a name, AWS CloudFormation
	// generates a unique physical ID and uses that ID for the rule name. For
	// more information, see Name Type.
	Name *StringExpr `json:"Name,omitempty"`

	// The Amazon Resource Name (ARN) of the AWS Identity and Access
	// Management (IAM) role that grants CloudWatch Events permission to make
	// calls to target services, such as AWS Lambda (Lambda) or Amazon
	// Kinesis streams.
	RoleArn *StringExpr `json:"RoleArn,omitempty"`

	// The schedule or rate (frequency) that determines when CloudWatch
	// Events runs the rule. For more information, see Schedule Expression
	// Syntax for Rules in the Amazon CloudWatch User Guide.
	ScheduleExpression *StringExpr `json:"ScheduleExpression,omitempty"`

	// Indicates whether the rule is enabled. For valid values, see the State
	// parameter for the PutRule action in the Amazon CloudWatch Events API
	// Reference.
	State *StringExpr `json:"State,omitempty"`

	// The resources, such as Lambda functions or Amazon Kinesis streams,
	// that CloudWatch Events routes events to and invokes when the rule is
	// triggered. For information about valid targets, see the PutTargets
	// action in the Amazon CloudWatch Events API Reference.
	Targets *CloudWatchEventsRuleTargetList `json:"Targets,omitempty"`
}

// CfnResourceType returns AWS::Events::Rule to implement the ResourceProperties interface
func (s EventsRule) CfnResourceType() string {
	return "AWS::Events::Rule"
}

// GameLiftAlias represents AWS::GameLift::Alias
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-gamelift-alias.html
type GameLiftAlias struct {
	// Information that helps you identify the purpose of this alias.
	Description *StringExpr `json:"Description,omitempty"`

	// An identifier to associate with this alias. Alias names don't need to
	// be unique.
	Name *StringExpr `json:"Name,omitempty"`

	// A routing configuration that specifies where traffic is directed for
	// this alias, such as to a fleet or to a message.
	RoutingStrategy *GameLiftAliasRoutingStrategy `json:"RoutingStrategy,omitempty"`
}

// CfnResourceType returns AWS::GameLift::Alias to implement the ResourceProperties interface
func (s GameLiftAlias) CfnResourceType() string {
	return "AWS::GameLift::Alias"
}

// GameLiftBuild represents AWS::GameLift::Build
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-gamelift-build.html
type GameLiftBuild struct {
	// An identifier to associate with this build. Build names don't need to
	// be unique.
	Name *StringExpr `json:"Name,omitempty"`

	// The Amazon Simple Storage Service (Amazon S3) location where your
	// build package files are located.
	StorageLocation *GameLiftBuildStorageLocation `json:"StorageLocation,omitempty"`

	// A version to associate with this build. Version is useful if you want
	// to track updates to your build package files. Versions don't need to
	// be unique.
	Version *StringExpr `json:"Version,omitempty"`
}

// CfnResourceType returns AWS::GameLift::Build to implement the ResourceProperties interface
func (s GameLiftBuild) CfnResourceType() string {
	return "AWS::GameLift::Build"
}

// GameLiftFleet represents AWS::GameLift::Fleet
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-gamelift-fleet.html
type GameLiftFleet struct {
	// The unique identifier for the build that you want to use with this
	// fleet.
	BuildId *StringExpr `json:"BuildId,omitempty"`

	// Information that helps you identify the purpose of this fleet.
	Description *StringExpr `json:"Description,omitempty"`

	// The number of EC2 instances that you want in this fleet.
	DesiredEC2Instances *IntegerExpr `json:"DesiredEC2Instances,omitempty"`

	// The incoming traffic, expressed as IP ranges and port numbers, that is
	// permitted to access the game server. If you don't specify values, no
	// traffic is permitted to your game servers.
	EC2InboundPermissions *GameLiftFleetEC2InboundPermissionList `json:"EC2InboundPermissions,omitempty"`

	// The type of EC2 instances that the fleet uses. EC2 instance types
	// define the CPU, memory, storage, and networking capacity of the
	// fleet's hosts. For more information about the instance types that are
	// supported by GameLift, see the EC2InstanceType parameter in the Amazon
	// GameLift API Reference.
	EC2InstanceType *StringExpr `json:"EC2InstanceType,omitempty"`

	// The path to game-session log files that are generated by your game
	// server, with the slashes (\) escaped. After a game session has been
	// terminated, GameLift captures and stores the logs in an S3 bucket.
	LogPaths *StringListExpr `json:"LogPaths,omitempty"`

	// The maximum number of EC2 instances that you want to allow in this
	// fleet. By default, AWS CloudFormation, sets this property to 1.
	MaxSize *IntegerExpr `json:"MaxSize,omitempty"`

	// The minimum number of EC2 instances that you want to allow in this
	// fleet. By default, AWS CloudFormation, sets this property to 0.
	MinSize *IntegerExpr `json:"MinSize,omitempty"`

	// An identifier to associate with this fleet. Fleet names don't need to
	// be unique.
	Name *StringExpr `json:"Name,omitempty"`

	// The parameters that are required to launch your game server. Specify
	// these parameters as a string of command-line parameters, such as
	// +sv_port 33435 +start_lobby.
	ServerLaunchParameters *StringExpr `json:"ServerLaunchParameters,omitempty"`

	// The location of your game server that GameLift launches. You must
	// escape the slashes (\) and use the following pattern:
	// C:\\game\\launchpath. For example, if your game server files are in
	// the MyGame folder, the path should be C:\\game\\MyGame\\server.exe.
	ServerLaunchPath *StringExpr `json:"ServerLaunchPath,omitempty"`
}

// CfnResourceType returns AWS::GameLift::Fleet to implement the ResourceProperties interface
func (s GameLiftFleet) CfnResourceType() string {
	return "AWS::GameLift::Fleet"
}

// IAMAccessKey represents AWS::IAM::AccessKey
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-iam-accesskey.html
type IAMAccessKey struct {
	// This value is specific to AWS CloudFormation and can only be
	// incremented. Incrementing this value notifies AWS CloudFormation that
	// you want to rotate your access key. When you update your stack, AWS
	// CloudFormation will replace the existing access key with a new key.
	Serial *IntegerExpr `json:"Serial,omitempty"`

	// The status of the access key. By default, AWS CloudFormation sets this
	// property value to Active.
	Status *StringExpr `json:"Status,omitempty"`

	// The name of the user that the new key will belong to.
	UserName *StringExpr `json:"UserName,omitempty"`
}

// CfnResourceType returns AWS::IAM::AccessKey to implement the ResourceProperties interface
func (s IAMAccessKey) CfnResourceType() string {
	return "AWS::IAM::AccessKey"
}

// IAMGroup represents AWS::IAM::Group
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-iam-group.html
type IAMGroup struct {
	// A name for the IAM group. For valid values, see the GroupName
	// parameter for the CreateGroup action in the IAM API Reference. If you
	// don't specify a name, AWS CloudFormation generates a unique physical
	// ID and uses that ID for the group name.
	GroupName *StringExpr `json:"GroupName,omitempty"`

	// One or more managed policy ARNs to attach to this group.
	ManagedPolicyArns *StringListExpr `json:"ManagedPolicyArns,omitempty"`

	// The path to the group. For more information about paths, see IAM
	// Identifiers in the IAM User Guide.
	Path *StringExpr `json:"Path,omitempty"`

	// The policies to associate with this group. For information about
	// policies, see Overview of IAM Policies in the IAM User Guide.
	Policies *IAMPoliciesList `json:"Policies,omitempty"`
}

// CfnResourceType returns AWS::IAM::Group to implement the ResourceProperties interface
func (s IAMGroup) CfnResourceType() string {
	return "AWS::IAM::Group"
}

// IAMInstanceProfile represents AWS::IAM::InstanceProfile
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-iam-instanceprofile.html
type IAMInstanceProfile struct {
	// The path associated with this IAM instance profile. For information
	// about IAM paths, see Friendly Names and Paths in the AWS Identity and
	// Access Management User Guide.
	Path *StringExpr `json:"Path,omitempty"`

	// The roles associated with this IAM instance profile.
	Roles interface{} `json:"Roles,omitempty"`
}

// CfnResourceType returns AWS::IAM::InstanceProfile to implement the ResourceProperties interface
func (s IAMInstanceProfile) CfnResourceType() string {
	return "AWS::IAM::InstanceProfile"
}

// IAMManagedPolicy represents AWS::IAM::ManagedPolicy
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-iam-managedpolicy.html
type IAMManagedPolicy struct {
	// A description of the policy. For example, you can describe the
	// permissions that are defined in the policy.
	Description *StringExpr `json:"Description,omitempty"`

	// The names of groups to attach to this policy.
	Groups *StringListExpr `json:"Groups,omitempty"`

	// The path for the policy. By default, the path is /. For more
	// information, see IAM Identifiers in the IAM User Guide guide.
	Path *StringExpr `json:"Path,omitempty"`

	// Policies that define the permissions for this managed policy. For more
	// information about policy syntax, see IAM Policy Elements Reference in
	// IAM User Guide.
	PolicyDocument interface{} `json:"PolicyDocument,omitempty"`

	// The names of roles to attach to this policy.
	Roles *StringListExpr `json:"Roles,omitempty"`

	// The names of users to attach to this policy.
	Users *StringListExpr `json:"Users,omitempty"`
}

// CfnResourceType returns AWS::IAM::ManagedPolicy to implement the ResourceProperties interface
func (s IAMManagedPolicy) CfnResourceType() string {
	return "AWS::IAM::ManagedPolicy"
}

// IAMPolicy represents AWS::IAM::Policy
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-iam-policy.html
type IAMPolicy struct {
	// The names of groups to which you want to add the policy.
	Groups *StringListExpr `json:"Groups,omitempty"`

	// A policy document that contains permissions to add to the specified
	// users or groups.
	PolicyDocument interface{} `json:"PolicyDocument,omitempty"`

	// The name of the policy. If you specify multiple policies for an
	// entity, specify unique names. For example, if you specify a list of
	// policies for an IAM role, each policy must have a unique name.
	PolicyName *StringExpr `json:"PolicyName,omitempty"`

	// The names of AWS::IAM::Roles to attach to this policy.
	Roles *StringListExpr `json:"Roles,omitempty"`

	// The names of users for whom you want to add the policy.
	Users *StringListExpr `json:"Users,omitempty"`
}

// CfnResourceType returns AWS::IAM::Policy to implement the ResourceProperties interface
func (s IAMPolicy) CfnResourceType() string {
	return "AWS::IAM::Policy"
}

// IAMRole represents AWS::IAM::Role
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-iam-role.html
type IAMRole struct {
	// The trust policy that is associated with this role.
	AssumeRolePolicyDocument interface{} `json:"AssumeRolePolicyDocument,omitempty"`

	// One or more managed policy ARNs to attach to this role.
	ManagedPolicyArns *StringListExpr `json:"ManagedPolicyArns,omitempty"`

	// The path associated with this role. For information about IAM paths,
	// see Friendly Names and Paths in IAM User Guide.
	Path *StringExpr `json:"Path,omitempty"`

	// The policies to associate with this role. For sample templates, see
	// Template Examples.
	Policies *IAMPoliciesList `json:"Policies,omitempty"`

	// A name for the IAM role. For valid values, see the RoleName parameter
	// for the CreateRole action in the IAM API Reference. If you don't
	// specify a name, AWS CloudFormation generates a unique physical ID and
	// uses that ID for the group name.
	RoleName *StringExpr `json:"RoleName,omitempty"`
}

// CfnResourceType returns AWS::IAM::Role to implement the ResourceProperties interface
func (s IAMRole) CfnResourceType() string {
	return "AWS::IAM::Role"
}

// IAMUser represents AWS::IAM::User
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-iam-user.html
type IAMUser struct {
	// A name of a group to which you want to add the user.
	Groups *StringListExpr `json:"Groups,omitempty"`

	// Creates a login profile so that the user can access the AWS Management
	// Console.
	LoginProfile *IAMUserLoginProfile `json:"LoginProfile,omitempty"`

	// One or more managed policy ARNs to attach to this user.
	ManagedPolicyArns *StringListExpr `json:"ManagedPolicyArns,omitempty"`

	// The path for the user name. For more information about paths, see IAM
	// Identifiers in the IAM User Guide.
	Path *StringExpr `json:"Path,omitempty"`

	// The policies to associate with this user. For information about
	// policies, see Overview of IAM Policies in the IAM User Guide.
	Policies *IAMPoliciesList `json:"Policies,omitempty"`

	// A name for the IAM user. For valid values, see the UserName parameter
	// for the CreateUser action in the IAM API Reference. If you don't
	// specify a name, AWS CloudFormation generates a unique physical ID and
	// uses that ID for the user name.
	UserName *StringExpr `json:"UserName,omitempty"`
}

// CfnResourceType returns AWS::IAM::User to implement the ResourceProperties interface
func (s IAMUser) CfnResourceType() string {
	return "AWS::IAM::User"
}

// IAMUserToGroupAddition represents AWS::IAM::UserToGroupAddition
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-iam-addusertogroup.html
type IAMUserToGroupAddition struct {
	// The name of group to add users to.
	GroupName *StringExpr `json:"GroupName,omitempty"`

	// Required: Yes
	Users interface{} `json:"Users,omitempty"`
}

// CfnResourceType returns AWS::IAM::UserToGroupAddition to implement the ResourceProperties interface
func (s IAMUserToGroupAddition) CfnResourceType() string {
	return "AWS::IAM::UserToGroupAddition"
}

// IoTCertificate represents AWS::IoT::Certificate
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-iot-certificate.html
type IoTCertificate struct {
	// The certificate signing request (CSR).
	CertificateSigningRequest *StringExpr `json:"CertificateSigningRequest,omitempty"`

	// The status of the certificate.
	Status *StringExpr `json:"Status,omitempty"`
}

// CfnResourceType returns AWS::IoT::Certificate to implement the ResourceProperties interface
func (s IoTCertificate) CfnResourceType() string {
	return "AWS::IoT::Certificate"
}

// IoTPolicy represents AWS::IoT::Policy
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-iot-policy.html
type IoTPolicy struct {
	// The JSON document that describes the policy.
	PolicyDocument interface{} `json:"PolicyDocument,omitempty"`

	// The name (the physical ID) of the AWS IoT policy.
	PolicyName *StringExpr `json:"PolicyName,omitempty"`
}

// CfnResourceType returns AWS::IoT::Policy to implement the ResourceProperties interface
func (s IoTPolicy) CfnResourceType() string {
	return "AWS::IoT::Policy"
}

// IoTPolicyPrincipalAttachment represents AWS::IoT::PolicyPrincipalAttachment
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-iot-policyprincipalattachment.html
type IoTPolicyPrincipalAttachment struct {
	// The name of the policy.
	PolicyName *StringExpr `json:"PolicyName,omitempty"`

	// The principal, which can be a certificate ARN (as returned from the
	// CreateCertificate operation) or an Amazon Cognito ID.
	Principal *StringExpr `json:"Principal,omitempty"`
}

// CfnResourceType returns AWS::IoT::PolicyPrincipalAttachment to implement the ResourceProperties interface
func (s IoTPolicyPrincipalAttachment) CfnResourceType() string {
	return "AWS::IoT::PolicyPrincipalAttachment"
}

// IoTThing represents AWS::IoT::Thing
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-iot-thing.html
type IoTThing struct {
	// A JSON string that contains up to three key-value pairs, for example:
	// { "attributes": { "string1":"string2" } }.
	AttributePayload interface{} `json:"AttributePayload,omitempty"`

	// The name (the physical ID) of the AWS IoT thing.
	ThingName *StringExpr `json:"ThingName,omitempty"`
}

// CfnResourceType returns AWS::IoT::Thing to implement the ResourceProperties interface
func (s IoTThing) CfnResourceType() string {
	return "AWS::IoT::Thing"
}

// IoTThingPrincipalAttachment represents AWS::IoT::ThingPrincipalAttachment
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-iot-thingprincipalattachment.html
type IoTThingPrincipalAttachment struct {
	// The principal, which can be a certificate ARN (as returned from the
	// CreateCertificate operation) or an Amazon Cognito ID.
	Principal *StringExpr `json:"Principal,omitempty"`

	// The name of the AWS IoT thing.
	ThingName *StringExpr `json:"ThingName,omitempty"`
}

// CfnResourceType returns AWS::IoT::ThingPrincipalAttachment to implement the ResourceProperties interface
func (s IoTThingPrincipalAttachment) CfnResourceType() string {
	return "AWS::IoT::ThingPrincipalAttachment"
}

// IoTTopicRule represents AWS::IoT::TopicRule
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-iot-topicrule.html
type IoTTopicRule struct {
	// The name (the physical ID) of the AWS IoT rule.
	RuleName *StringExpr `json:"RuleName,omitempty"`

	// The actions associated with the AWS IoT rule.
	TopicRulePayload *IoTTopicRulePayload `json:"TopicRulePayload,omitempty"`
}

// CfnResourceType returns AWS::IoT::TopicRule to implement the ResourceProperties interface
func (s IoTTopicRule) CfnResourceType() string {
	return "AWS::IoT::TopicRule"
}

// KinesisStream represents AWS::Kinesis::Stream
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-kinesis-stream.html
type KinesisStream struct {
	// The name of the Amazon Kinesis stream. If you don't specify a name,
	// AWS CloudFormation generates a unique physical ID and uses that ID for
	// the stream name. For more information, see Name Type.
	Name *StringExpr `json:"Name,omitempty"`

	// The number of shards that the stream uses. For greater provisioned
	// throughput, increase the number of shards.
	ShardCount *IntegerExpr `json:"ShardCount,omitempty"`

	// An arbitrary set of tags (key–value pairs) to associate with the
	// Amazon Kinesis stream.
	Tags []ResourceTag `json:"Tags,omitempty"`
}

// CfnResourceType returns AWS::Kinesis::Stream to implement the ResourceProperties interface
func (s KinesisStream) CfnResourceType() string {
	return "AWS::Kinesis::Stream"
}

// KinesisFirehoseDeliveryStream represents AWS::KinesisFirehose::DeliveryStream
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-kinesisfirehose-deliverystream.html
type KinesisFirehoseDeliveryStream struct {
	// A name for the delivery stream.
	DeliveryStreamName *StringExpr `json:"DeliveryStreamName,omitempty"`

	// An Amazon ES destination for the delivery stream.
	ElasticsearchDestinationConfiguration *KinesisFirehoseDeliveryStreamElasticsearchDestinationConfiguration `json:"ElasticsearchDestinationConfiguration,omitempty"`

	// An Amazon Redshift destination for the delivery stream.
	RedshiftDestinationConfiguration *KinesisFirehoseDeliveryStreamRedshiftDestinationConfiguration `json:"RedshiftDestinationConfiguration,omitempty"`

	// An Amazon S3 destination for the delivery stream.
	S3DestinationConfiguration *KinesisFirehoseDeliveryStreamS3DestinationConfiguration `json:"S3DestinationConfiguration,omitempty"`
}

// CfnResourceType returns AWS::KinesisFirehose::DeliveryStream to implement the ResourceProperties interface
func (s KinesisFirehoseDeliveryStream) CfnResourceType() string {
	return "AWS::KinesisFirehose::DeliveryStream"
}

// KMSAlias represents AWS::KMS::Alias
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-kms-alias.html
type KMSAlias struct {
	// The name of the alias. The name must start with alias followed by a
	// forward slash, such as alias/. You can't specify aliases that begin
	// with alias/AWS. These aliases are reserved.
	AliasName *StringExpr `json:"AliasName,omitempty"`

	// The ID of the key for which you are creating the alias. Specify the
	// key's globally unique identifier or Amazon Resource Name (ARN). You
	// can't specify another alias.
	TargetKeyId *StringExpr `json:"TargetKeyId,omitempty"`
}

// CfnResourceType returns AWS::KMS::Alias to implement the ResourceProperties interface
func (s KMSAlias) CfnResourceType() string {
	return "AWS::KMS::Alias"
}

// KMSKey represents AWS::KMS::Key
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-kms-key.html
type KMSKey struct {
	// A description of the key. Use a description that helps your users
	// decide whether the key is appropriate for a particular task.
	Description *StringExpr `json:"Description,omitempty"`

	// Indicates whether the key is available for use. AWS CloudFormation
	// sets this value to true by default.
	Enabled *BoolExpr `json:"Enabled,omitempty"`

	// Indicates whether AWS KMS rotates the key. AWS CloudFormation sets
	// this value to false by default.
	EnableKeyRotation *BoolExpr `json:"EnableKeyRotation,omitempty"`

	// An AWS KMS key policy to attach to the key. Use a policy to specify
	// who has permission to use the key and which actions they can perform.
	// For more information, see Key Policies in the AWS Key Management
	// Service Developer Guide.
	KeyPolicy interface{} `json:"KeyPolicy,omitempty"`
}

// CfnResourceType returns AWS::KMS::Key to implement the ResourceProperties interface
func (s KMSKey) CfnResourceType() string {
	return "AWS::KMS::Key"
}

// LambdaEventSourceMapping represents AWS::Lambda::EventSourceMapping
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-eventsourcemapping.html
type LambdaEventSourceMapping struct {
	// The largest number of records that Lambda retrieves from your event
	// source when invoking your function. Your function receives an event
	// with all the retrieved records. For the default and valid values, see
	// CreateEventSourceMapping in the AWS Lambda Developer Guide.
	BatchSize *IntegerExpr `json:"BatchSize,omitempty"`

	// Indicates whether Lambda begins polling the event source.
	Enabled *BoolExpr `json:"Enabled,omitempty"`

	// The Amazon Resource Name (ARN) of the Amazon Kinesis or DynamoDB
	// stream that is the source of events. Any record added to this stream
	// can invoke the Lambda function. For more information, see
	// CreateEventSourceMapping in the AWS Lambda Developer Guide.
	EventSourceArn *StringExpr `json:"EventSourceArn,omitempty"`

	// The name or ARN of a Lambda function to invoke when Lambda detects an
	// event on the stream.
	FunctionName *StringExpr `json:"FunctionName,omitempty"`

	// The position in the stream where Lambda starts reading. For valid
	// values, see CreateEventSourceMapping in the AWS Lambda Developer
	// Guide.
	StartingPosition *StringExpr `json:"StartingPosition,omitempty"`
}

// CfnResourceType returns AWS::Lambda::EventSourceMapping to implement the ResourceProperties interface
func (s LambdaEventSourceMapping) CfnResourceType() string {
	return "AWS::Lambda::EventSourceMapping"
}

// LambdaAlias represents AWS::Lambda::Alias
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-alias.html
type LambdaAlias struct {
	// Information about the alias, such as its purpose or the Lambda
	// function that is associated with it.
	Description *StringExpr `json:"Description,omitempty"`

	// The Lambda function that you want to associate with this alias. You
	// can specify the function's name or its Amazon Resource Name (ARN).
	FunctionName *StringExpr `json:"FunctionName,omitempty"`

	// The version of the Lambda function that you want to associate with
	// this alias.
	FunctionVersion *StringExpr `json:"FunctionVersion,omitempty"`

	// A name for the alias.
	Name *StringExpr `json:"Name,omitempty"`
}

// CfnResourceType returns AWS::Lambda::Alias to implement the ResourceProperties interface
func (s LambdaAlias) CfnResourceType() string {
	return "AWS::Lambda::Alias"
}

// LambdaFunction represents AWS::Lambda::Function
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-function.html
type LambdaFunction struct {
	// The source code of your Lambda function. You can point to a file in an
	// Amazon Simple Storage Service (Amazon S3) bucket or specify your
	// source code as inline text.
	Code *LambdaFunctionCode `json:"Code,omitempty"`

	// A description of the function.
	Description *StringExpr `json:"Description,omitempty"`

	// Key-value pairs that Lambda caches and makes available for your Lambda
	// functions. Use environment variables to apply configuration changes,
	// such as test and production environment configurations, without
	// changing your Lambda function source code.
	Environment *LambdaFunctionEnvironment `json:"Environment,omitempty"`

	// A name for the function. If you don't specify a name, AWS
	// CloudFormation generates a unique physical ID and uses that ID for the
	// function's name. For more information, see Name Type.
	FunctionName *StringExpr `json:"FunctionName,omitempty"`

	// The name of the function (within your source code) that Lambda calls
	// to start running your code. For more information, see the Handler
	// property in the AWS Lambda Developer Guide.
	Handler *StringExpr `json:"Handler,omitempty"`

	// The Amazon Resource Name (ARN) of an AWS Key Management Service (AWS
	// KMS) key that Lambda uses to encrypt and decrypt environment variable
	// values.
	KmsKeyArn *StringExpr `json:"KmsKeyArn,omitempty"`

	// The amount of memory, in MB, that is allocated to your Lambda
	// function. Lambda uses this value to proportionally allocate the amount
	// of CPU power. For more information, see Resource Model in the AWS
	// Lambda Developer Guide.
	MemorySize *IntegerExpr `json:"MemorySize,omitempty"`

	// The Amazon Resource Name (ARN) of the AWS Identity and Access
	// Management (IAM) execution role that Lambda assumes when it runs your
	// code to access AWS services.
	Role *StringExpr `json:"Role,omitempty"`

	// The runtime environment for the Lambda function that you are
	// uploading. For valid values, see the Runtime property in the AWS
	// Lambda Developer Guide.
	Runtime *StringExpr `json:"Runtime,omitempty"`

	// The function execution time (in seconds) after which Lambda terminates
	// the function. Because the execution time affects cost, set this value
	// based on the function's expected execution time. By default, Timeout
	// is set to 3 seconds.
	Timeout *IntegerExpr `json:"Timeout,omitempty"`

	// If the Lambda function requires access to resources in a VPC, specify
	// a VPC configuration that Lambda uses to set up an elastic network
	// interface (ENI). The ENI enables your function to connect to other
	// resources in your VPC, but it doesn't provide public Internet access.
	// If your function requires Internet access (for example, to access AWS
	// services that don't have VPC endpoints), configure a Network Address
	// Translation (NAT) instance inside your VPC or use an Amazon Virtual
	// Private Cloud (Amazon VPC) NAT gateway. For more information, see NAT
	// Gateways in the Amazon VPC User Guide.
	VpcConfig *LambdaFunctionVPCConfig `json:"VpcConfig,omitempty"`
}

// CfnResourceType returns AWS::Lambda::Function to implement the ResourceProperties interface
func (s LambdaFunction) CfnResourceType() string {
	return "AWS::Lambda::Function"
}

// LambdaPermission represents AWS::Lambda::Permission
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-permission.html
type LambdaPermission struct {
	// The Lambda actions that you want to allow in this statement. For
	// example, you can specify lambda:CreateFunction to specify a certain
	// action, or use a wildcard (lambda:*) to grant permission to all Lambda
	// actions. For a list of actions, see Actions and Condition Context Keys
	// for AWS Lambda in the IAM User Guide.
	Action *StringExpr `json:"Action,omitempty"`

	// The name (physical ID), Amazon Resource Name (ARN), or alias ARN of
	// the Lambda function that you want to associate with this statement.
	// Lambda adds this statement to the function's access policy.
	FunctionName *StringExpr `json:"FunctionName,omitempty"`

	// The entity for which you are granting permission to invoke the Lambda
	// function. This entity can be any valid AWS service principal, such as
	// s3.amazonaws.com or sns.amazonaws.com, or, if you are granting
	// cross-account permission, an AWS account ID. For example, you might
	// want to allow a custom application in another AWS account to push
	// events to Lambda by invoking your function.
	Principal *StringExpr `json:"Principal,omitempty"`

	// The AWS account ID (without hyphens) of the source owner. For example,
	// if you specify an S3 bucket in the SourceArn property, this value is
	// the bucket owner's account ID. You can use this property to ensure
	// that all source principals are owned by a specific account.
	SourceAccount *StringExpr `json:"SourceAccount,omitempty"`

	// The ARN of a resource that is invoking your function. When granting
	// Amazon Simple Storage Service (Amazon S3) permission to invoke your
	// function, specify this property with the bucket ARN as its value. This
	// ensures that events generated only from the specified bucket, not just
	// any bucket from any AWS account that creates a mapping to your
	// function, can invoke the function.
	SourceArn *StringExpr `json:"SourceArn,omitempty"`
}

// CfnResourceType returns AWS::Lambda::Permission to implement the ResourceProperties interface
func (s LambdaPermission) CfnResourceType() string {
	return "AWS::Lambda::Permission"
}

// LambdaVersion represents AWS::Lambda::Version
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-version.html
type LambdaVersion struct {
	// The SHA-256 hash of the deployment package that you want to publish.
	// This value must match the SHA-256 hash of the $LATEST version of the
	// function. Specify this property to validate that you are publishing
	// the correct package.
	CodeSha256 *StringExpr `json:"CodeSha256,omitempty"`

	// A description of the version you are publishing. If you don't specify
	// a value, Lambda copies the description from the $LATEST version of the
	// function.
	Description *StringExpr `json:"Description,omitempty"`

	// The Lambda function for which you want to publish a version. You can
	// specify the function's name or its Amazon Resource Name (ARN).
	FunctionName *StringExpr `json:"FunctionName,omitempty"`
}

// CfnResourceType returns AWS::Lambda::Version to implement the ResourceProperties interface
func (s LambdaVersion) CfnResourceType() string {
	return "AWS::Lambda::Version"
}

// LogsDestination represents AWS::Logs::Destination
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-logs-destination.html
type LogsDestination struct {
	// The name of the CloudWatch Logs destination.
	DestinationName *StringExpr `json:"DestinationName,omitempty"`

	// An AWS Identity and Access Management (IAM) policy that specifies who
	// can write to your destination.
	DestinationPolicy *StringExpr `json:"DestinationPolicy,omitempty"`

	// The Amazon Resource Name (ARN) of an IAM role that permits CloudWatch
	// Logs to send data to the specified AWS resource (TargetArn).
	RoleArn *StringExpr `json:"RoleArn,omitempty"`

	// The ARN of the AWS resource that receives log events. Currently, you
	// can specify only an Amazon Kinesis stream.
	TargetArn *StringExpr `json:"TargetArn,omitempty"`
}

// CfnResourceType returns AWS::Logs::Destination to implement the ResourceProperties interface
func (s LogsDestination) CfnResourceType() string {
	return "AWS::Logs::Destination"
}

// LogsLogGroup represents AWS::Logs::LogGroup
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-logs-loggroup.html
type LogsLogGroup struct {
	// A name for the log group. If you don't specify a name, AWS
	// CloudFormation generates a unique physical ID and uses that ID for the
	// table name. For more information, see Name Type.
	LogGroupName *StringExpr `json:"LogGroupName,omitempty"`

	// The number of days log events are kept in CloudWatch Logs. When a log
	// event expires, CloudWatch Logs automatically deletes it. For valid
	// values, see PutRetentionPolicy in the Amazon CloudWatch Logs API
	// Reference.
	RetentionInDays *IntegerExpr `json:"RetentionInDays,omitempty"`
}

// CfnResourceType returns AWS::Logs::LogGroup to implement the ResourceProperties interface
func (s LogsLogGroup) CfnResourceType() string {
	return "AWS::Logs::LogGroup"
}

// LogsLogStream represents AWS::Logs::LogStream
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-logs-logstream.html
type LogsLogStream struct {
	// The name of the log group where the log stream is created.
	LogGroupName *StringExpr `json:"LogGroupName,omitempty"`

	// The name of the log stream to create. The name must be unique within
	// the log group.
	LogStreamName *StringExpr `json:"LogStreamName,omitempty"`
}

// CfnResourceType returns AWS::Logs::LogStream to implement the ResourceProperties interface
func (s LogsLogStream) CfnResourceType() string {
	return "AWS::Logs::LogStream"
}

// LogsMetricFilter represents AWS::Logs::MetricFilter
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-logs-metricfilter.html
type LogsMetricFilter struct {
	// Describes the pattern that CloudWatch Logs follows to interpret each
	// entry in a log. For example, a log entry might contain fields such as
	// timestamps, IP addresses, error codes, bytes transferred, and so on.
	// You use the pattern to specify those fields and to specify what to
	// look for in the log file. For example, if you're interested in error
	// codes that begin with 1234, your filter pattern might be [timestamps,
	// ip_addresses, error_codes = 1234*, size, ...].
	FilterPattern *StringExpr `json:"FilterPattern,omitempty"`

	// The name of an existing log group that you want to associate with this
	// metric filter.
	LogGroupName *StringExpr `json:"LogGroupName,omitempty"`

	// Describes how to transform data from a log into a CloudWatch metric.
	MetricTransformations *CloudWatchLogsMetricFilterMetricTransformationPropertyList `json:"MetricTransformations,omitempty"`
}

// CfnResourceType returns AWS::Logs::MetricFilter to implement the ResourceProperties interface
func (s LogsMetricFilter) CfnResourceType() string {
	return "AWS::Logs::MetricFilter"
}

// LogsSubscriptionFilter represents AWS::Logs::SubscriptionFilter
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-logs-subscriptionfilter.html
type LogsSubscriptionFilter struct {
	// The Amazon Resource Name (ARN) of the Amazon Kinesis stream or Lambda
	// function that you want to use as the subscription feed destination.
	DestinationArn *StringExpr `json:"DestinationArn,omitempty"`

	// The filtering expressions that restrict what gets delivered to the
	// destination AWS resource. For more information about the filter
	// pattern syntax, see Filter and Pattern Syntax in the Amazon CloudWatch
	// User Guide.
	FilterPattern *StringExpr `json:"FilterPattern,omitempty"`

	// The log group to associate with the subscription filter. All log
	// events that are uploaded to this log group are filtered and delivered
	// to the specified AWS resource if the filter pattern matches the log
	// events.
	LogGroupName *StringExpr `json:"LogGroupName,omitempty"`

	// An IAM role that grants CloudWatch Logs permission to put data into
	// the specified Amazon Kinesis stream. For Lambda and CloudWatch Logs
	// destinations, don't specify this property because CloudWatch Logs gets
	// the necessary permissions from the destination resource.
	RoleArn *StringExpr `json:"RoleArn,omitempty"`
}

// CfnResourceType returns AWS::Logs::SubscriptionFilter to implement the ResourceProperties interface
func (s LogsSubscriptionFilter) CfnResourceType() string {
	return "AWS::Logs::SubscriptionFilter"
}

// OpsWorksApp represents AWS::OpsWorks::App
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-opsworks-app.html
type OpsWorksApp struct {
	// The information required to retrieve an app from a repository.
	AppSource *OpsWorksSource `json:"AppSource,omitempty"`

	// One or more user-defined key-value pairs to be added to the stack
	// attributes bag.
	Attributes interface{} `json:"Attributes,omitempty"`

	// A description of the app.
	Description *StringExpr `json:"Description,omitempty"`

	// A list of databases to associate with the AWS OpsWorks app.
	DataSources *DataSourceList `json:"DataSources,omitempty"`

	// The app virtual host settings, with multiple domains separated by
	// commas. For example, 'www.example.com, example.com'.
	Domains *StringListExpr `json:"Domains,omitempty"`

	// Whether to enable SSL for this app.
	EnableSsl *BoolExpr `json:"EnableSsl,omitempty"`

	// The environment variables to associate with the AWS OpsWorks app.
	Environment *OpsWorksAppEnvironmentList `json:"Environment,omitempty"`

	// The name of the AWS OpsWorks app.
	Name *StringExpr `json:"Name,omitempty"`

	// The app short name, which is used internally by AWS OpsWorks and by
	// Chef recipes.
	Shortname *StringExpr `json:"Shortname,omitempty"`

	// The SSL configuration
	SslConfiguration *OpsWorksSslConfiguration `json:"SslConfiguration,omitempty"`

	// The ID of the AWS OpsWorks stack to associate this app with.
	StackId *StringExpr `json:"StackId,omitempty"`

	// The app type. Each supported type is associated with a particular
	// layer. For more information, see CreateApp in the AWS OpsWorks API
	// Reference.
	Type *StringExpr `json:"Type,omitempty"`
}

// CfnResourceType returns AWS::OpsWorks::App to implement the ResourceProperties interface
func (s OpsWorksApp) CfnResourceType() string {
	return "AWS::OpsWorks::App"
}

// OpsWorksElasticLoadBalancerAttachment represents AWS::OpsWorks::ElasticLoadBalancerAttachment
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-opsworks-elbattachment.html
type OpsWorksElasticLoadBalancerAttachment struct {
	// Elastic Load Balancing load balancer name.
	ElasticLoadBalancerName *StringExpr `json:"ElasticLoadBalancerName,omitempty"`

	// The AWS OpsWorks layer ID that the Elastic Load Balancing load
	// balancer will be attached to.
	LayerId *StringExpr `json:"LayerId,omitempty"`
}

// CfnResourceType returns AWS::OpsWorks::ElasticLoadBalancerAttachment to implement the ResourceProperties interface
func (s OpsWorksElasticLoadBalancerAttachment) CfnResourceType() string {
	return "AWS::OpsWorks::ElasticLoadBalancerAttachment"
}

// OpsWorksInstance represents AWS::OpsWorks::Instance
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-opsworks-instance.html
type OpsWorksInstance struct {
	// The version of the AWS OpsWorks agent that AWS OpsWorks installs on
	// each instance. AWS OpsWorks sends commands to the agent to performs
	// tasks on your instances, such as starting Chef runs. For valid values,
	// see the AgentVersion parameter for the CreateInstance action in the
	// AWS OpsWorks API Reference.
	AgentVersion *StringExpr `json:"AgentVersion,omitempty"`

	// The ID of the custom Amazon Machine Image (AMI) to be used to create
	// the instance. For more information about custom AMIs, see Using Custom
	// AMIs in the AWS OpsWorks User Guide.
	AmiId *StringExpr `json:"AmiId,omitempty"`

	// The instance architecture.
	Architecture *StringExpr `json:"Architecture,omitempty"`

	// For scaling instances, the type of scaling. If you specify load-based
	// scaling, do not specify a time-based scaling configuration. For valid
	// values, see CreateInstance in the AWS OpsWorks API Reference.
	AutoScalingType *StringExpr `json:"AutoScalingType,omitempty"`

	// The instance Availability Zone.
	AvailabilityZone *StringExpr `json:"AvailabilityZone,omitempty"`

	// A list of block devices that are mapped to the AWS OpsWorks instance.
	// For more information, see the BlockDeviceMappings parameter for the
	// CreateInstance action in the AWS OpsWorks API Reference.
	BlockDeviceMappings *OpsWorksInstanceBlockDeviceMappingList `json:"BlockDeviceMappings,omitempty"`

	// Whether the instance is optimized for Amazon Elastic Block Store
	// (Amazon EBS) I/O. If you specify an Amazon EBS-optimized instance
	// type, AWS OpsWorks enables EBS optimization by default. For more
	// information, see Amazon EBS–Optimized Instances in the Amazon EC2
	// User Guide for Linux Instances.
	EbsOptimized *BoolExpr `json:"EbsOptimized,omitempty"`

	// A list of Elastic IP addresses to associate with the instance.
	ElasticIps *StringListExpr `json:"ElasticIps,omitempty"`

	// The name of the instance host.
	Hostname *StringExpr `json:"Hostname,omitempty"`

	// Whether to install operating system and package updates when the
	// instance boots.
	InstallUpdatesOnBoot *BoolExpr `json:"InstallUpdatesOnBoot,omitempty"`

	// The instance type, which must be supported by AWS OpsWorks. For more
	// information, see CreateInstance in the AWS OpsWorks API Reference.
	InstanceType *StringExpr `json:"InstanceType,omitempty"`

	// The IDs of the AWS OpsWorks layers to associate with this instance.
	LayerIds *StringListExpr `json:"LayerIds,omitempty"`

	// The instance operating system. For more information, see
	// CreateInstance in the AWS OpsWorks API Reference.
	Os *StringExpr `json:"Os,omitempty"`

	// The root device type of the instance.
	RootDeviceType *StringExpr `json:"RootDeviceType,omitempty"`

	// The SSH key name of the instance.
	SshKeyName *StringExpr `json:"SshKeyName,omitempty"`

	// The ID of the AWS OpsWorks stack that this instance will be associated
	// with.
	StackId *StringExpr `json:"StackId,omitempty"`

	// The ID of the instance's subnet. If the stack is running in a VPC, you
	// can use this parameter to override the stack's default subnet ID value
	// and direct AWS OpsWorks to launch the instance in a different subnet.
	SubnetId *StringExpr `json:"SubnetId,omitempty"`

	// The tenancy of the instance. For more information, see the Tenancy
	// parameter for the CreateInstance action in the AWS OpsWorks API
	// Reference.
	Tenancy *StringExpr `json:"Tenancy,omitempty"`

	// The time-based scaling configuration for the instance.
	TimeBasedAutoScaling *OpsWorksTimeBasedAutoScaling `json:"TimeBasedAutoScaling,omitempty"`

	// The instance's virtualization type, paravirtual or hvm.
	VirtualizationType *StringExpr `json:"VirtualizationType,omitempty"`

	// A list of Amazon EBS volume IDs to associate with the instance.
	Volumes *StringListExpr `json:"Volumes,omitempty"`
}

// CfnResourceType returns AWS::OpsWorks::Instance to implement the ResourceProperties interface
func (s OpsWorksInstance) CfnResourceType() string {
	return "AWS::OpsWorks::Instance"
}

// OpsWorksLayer represents AWS::OpsWorks::Layer
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-opsworks-layer.html
type OpsWorksLayer struct {
	// One or more user-defined key-value pairs to be added to the stack
	// attributes bag.
	Attributes interface{} `json:"Attributes,omitempty"`

	// Whether to automatically assign an Elastic IP address to Amazon EC2
	// instances in this layer.
	AutoAssignElasticIps *BoolExpr `json:"AutoAssignElasticIps,omitempty"`

	// For AWS OpsWorks stacks that are running in a VPC, whether to
	// automatically assign a public IP address to Amazon EC2 instances in
	// this layer.
	AutoAssignPublicIps *BoolExpr `json:"AutoAssignPublicIps,omitempty"`

	// The Amazon Resource Name (ARN) of an IAM instance profile that is to
	// be used for the Amazon EC2 instances in this layer.
	CustomInstanceProfileArn *StringExpr `json:"CustomInstanceProfileArn,omitempty"`

	// A custom stack configuration and deployment attributes that AWS
	// OpsWorks installs on the layer's instances. For more information, see
	// the CustomJson parameter for the CreateLayer action in the AWS
	// OpsWorks API Reference.
	CustomJson interface{} `json:"CustomJson,omitempty"`

	// Custom event recipes for this layer.
	CustomRecipes *OpsWorksRecipes `json:"CustomRecipes,omitempty"`

	// Custom security group IDs for this layer.
	CustomSecurityGroupIds *StringListExpr `json:"CustomSecurityGroupIds,omitempty"`

	// Whether to automatically heal Amazon EC2 instances that have become
	// disconnected or timed out.
	EnableAutoHealing *BoolExpr `json:"EnableAutoHealing,omitempty"`

	// Whether to install operating system and package updates when the
	// instance boots.
	InstallUpdatesOnBoot *BoolExpr `json:"InstallUpdatesOnBoot,omitempty"`

	// The lifecycle events for the AWS OpsWorks layer.
	LifecycleEventConfiguration *OpsWorksLayerLifeCycleConfiguration `json:"LifecycleEventConfiguration,omitempty"`

	// The load-based scaling configuration for the AWS OpsWorks layer.
	LoadBasedAutoScaling *OpsWorksLoadBasedAutoScaling `json:"LoadBasedAutoScaling,omitempty"`

	// The AWS OpsWorks layer name.
	Name *StringExpr `json:"Name,omitempty"`

	// The packages for this layer.
	Packages *StringListExpr `json:"Packages,omitempty"`

	// The layer short name, which is used internally by AWS OpsWorks and by
	// Chef recipes. The short name is also used as the name for the
	// directory where your app files are installed.
	Shortname *StringExpr `json:"Shortname,omitempty"`

	// The ID of the AWS OpsWorks stack that this layer will be associated
	// with.
	StackId *StringExpr `json:"StackId,omitempty"`

	// The layer type. A stack cannot have more than one layer of the same
	// type, except for the custom type. You can have any number of custom
	// types. For more information, see CreateLayer in the AWS OpsWorks API
	// Reference.
	Type *StringExpr `json:"Type,omitempty"`

	// Describes the Amazon EBS volumes for this layer.
	VolumeConfigurations *OpsWorksVolumeConfigurationList `json:"VolumeConfigurations,omitempty"`
}

// CfnResourceType returns AWS::OpsWorks::Layer to implement the ResourceProperties interface
func (s OpsWorksLayer) CfnResourceType() string {
	return "AWS::OpsWorks::Layer"
}

// OpsWorksStack represents AWS::OpsWorks::Stack
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-opsworks-stack.html
type OpsWorksStack struct {
	// The AWS OpsWorks agent version that you want to use. The agent
	// communicates with the service and handles tasks such as initiating
	// Chef runs in response to lifecycle events. For valid values, see the
	// AgentVersion parameter for the CreateStack action in the AWS OpsWorks
	// API Reference.
	AgentVersion *StringExpr `json:"AgentVersion,omitempty"`

	// One or more user-defined key-value pairs to be added to the stack
	// attributes bag.
	Attributes interface{} `json:"Attributes,omitempty"`

	// Describes the Chef configuration. For more information, see the
	// CreateStack ChefConfiguration parameter in the AWS OpsWorks API
	// Reference.
	ChefConfiguration *OpsWorksChefConfiguration `json:"ChefConfiguration,omitempty"`

	// If you're cloning an AWS OpsWorks stack, a list of AWS OpsWorks
	// application stack IDs from the source stack to include in the cloned
	// stack.
	CloneAppIds *StringListExpr `json:"CloneAppIds,omitempty"`

	// If you're cloning an AWS OpsWorks stack, indicates whether to clone
	// the source stack's permissions.
	ClonePermissions *BoolExpr `json:"ClonePermissions,omitempty"`

	// Describes the configuration manager. When you create a stack, you use
	// the configuration manager to specify the Chef version. For supported
	// Chef versions, see the CreateStack ConfigurationManager parameter in
	// the AWS OpsWorks API Reference.
	ConfigurationManager *OpsWorksStackConfigurationManager `json:"ConfigurationManager,omitempty"`

	// Contains the information required to retrieve a cookbook from a
	// repository.
	CustomCookbooksSource *OpsWorksSource `json:"CustomCookbooksSource,omitempty"`

	// A user-defined custom JSON object. The custom JSON is used to override
	// the corresponding default stack configuration JSON values. For more
	// information, see CreateStack in the AWS OpsWorks API Reference.
	CustomJson interface{} `json:"CustomJson,omitempty"`

	// The stack's default Availability Zone, which must be in the specified
	// region.
	DefaultAvailabilityZone *StringExpr `json:"DefaultAvailabilityZone,omitempty"`

	// The Amazon Resource Name (ARN) of an IAM instance profile that is the
	// default profile for all of the stack's Amazon EC2 instances.
	DefaultInstanceProfileArn *StringExpr `json:"DefaultInstanceProfileArn,omitempty"`

	// The stack's default operating system. For more information, see
	// CreateStack in the AWS OpsWorks API Reference.
	DefaultOs *StringExpr `json:"DefaultOs,omitempty"`

	// The default root device type. This value is used by default for all
	// instances in the stack, but you can override it when you create an
	// instance. For more information, see CreateStack in the AWS OpsWorks
	// API Reference.
	DefaultRootDeviceType *StringExpr `json:"DefaultRootDeviceType,omitempty"`

	// A default SSH key for the stack instances. You can override this value
	// when you create or update an instance.
	DefaultSshKeyName *StringExpr `json:"DefaultSshKeyName,omitempty"`

	// The stack's default subnet ID. All instances are launched into this
	// subnet unless you specify another subnet ID when you create the
	// instance.
	DefaultSubnetId *StringExpr `json:"DefaultSubnetId,omitempty"`

	// The Amazon Resource Name (ARN) of the Amazon EC2 Container Service
	// (Amazon ECS) cluster to register with the AWS OpsWorks stack.
	EcsClusterArn *StringExpr `json:"EcsClusterArn,omitempty"`

	// A list of Elastic IP addresses to register with the AWS OpsWorks
	// stack.
	ElasticIps *OpsWorksStackElasticIpList `json:"ElasticIps,omitempty"`

	// The stack's host name theme, with spaces replaced by underscores. The
	// theme is used to generate host names for the stack's instances. For
	// more information, see CreateStack in the AWS OpsWorks API Reference.
	HostnameTheme *StringExpr `json:"HostnameTheme,omitempty"`

	// The name of the AWS OpsWorks stack.
	Name *StringExpr `json:"Name,omitempty"`

	// The Amazon Relational Database Service (Amazon RDS) DB instance to
	// register with the AWS OpsWorks stack.
	RdsDbInstances *OpsWorksStackRdsDbInstanceList `json:"RdsDbInstances,omitempty"`

	// The AWS Identity and Access Management (IAM) role that AWS OpsWorks
	// uses to work with AWS resources on your behalf. You must specify an
	// Amazon Resource Name (ARN) for an existing IAM role.
	ServiceRoleArn *StringExpr `json:"ServiceRoleArn,omitempty"`

	// If you're cloning an AWS OpsWorks stack, the stack ID of the source
	// AWS OpsWorks stack to clone.
	SourceStackId *StringExpr `json:"SourceStackId,omitempty"`

	// Whether the stack uses custom cookbooks.
	UseCustomCookbooks *BoolExpr `json:"UseCustomCookbooks,omitempty"`

	// Whether to associate the AWS OpsWorks built-in security groups with
	// the stack's layers.
	UseOpsworksSecurityGroups *BoolExpr `json:"UseOpsworksSecurityGroups,omitempty"`

	// The ID of the VPC that the stack is to be launched into, which must be
	// in the specified region. All instances are launched into this VPC. If
	// you specify this property, you must specify the DefaultSubnetId
	// property.
	VpcId *StringExpr `json:"VpcId,omitempty"`
}

// CfnResourceType returns AWS::OpsWorks::Stack to implement the ResourceProperties interface
func (s OpsWorksStack) CfnResourceType() string {
	return "AWS::OpsWorks::Stack"
}

// OpsWorksUserProfile represents AWS::OpsWorks::UserProfile
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-opsworks-userprofile.html
type OpsWorksUserProfile struct {
	// Indicates whether users can use the AWS OpsWorks My Settings page to
	// specify their own SSH public key. For more information, see Setting an
	// IAM User's Public SSH Key in the AWS OpsWorks User Guide.
	AllowSelfManagement *BoolExpr `json:"AllowSelfManagement,omitempty"`

	// The Amazon Resource Name (ARN) of the AWS Identity and Access
	// Management (IAM) user to associate with this configuration.
	IamUserArn *StringExpr `json:"IamUserArn,omitempty"`

	// The public SSH key that is associated with the IAM user. The IAM user
	// must have or be given the corresponding private key to access
	// instances.
	SshPublicKey *StringExpr `json:"SshPublicKey,omitempty"`
}

// CfnResourceType returns AWS::OpsWorks::UserProfile to implement the ResourceProperties interface
func (s OpsWorksUserProfile) CfnResourceType() string {
	return "AWS::OpsWorks::UserProfile"
}

// OpsWorksVolume represents AWS::OpsWorks::Volume
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-opsworks-volume.html
type OpsWorksVolume struct {
	// The ID of the Amazon EBS volume to register with the AWS OpsWorks
	// stack.
	Ec2VolumeId *StringExpr `json:"Ec2VolumeId,omitempty"`

	// The mount point for the Amazon EBS volume, such as /mnt/disk1.
	MountPoint *StringExpr `json:"MountPoint,omitempty"`

	// A name for the Amazon EBS volume.
	Name *StringExpr `json:"Name,omitempty"`

	// The ID of the AWS OpsWorks stack that AWS OpsWorks registers the
	// volume to.
	StackId *StringExpr `json:"StackId,omitempty"`
}

// CfnResourceType returns AWS::OpsWorks::Volume to implement the ResourceProperties interface
func (s OpsWorksVolume) CfnResourceType() string {
	return "AWS::OpsWorks::Volume"
}

// RDSDBCluster represents AWS::RDS::DBCluster
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-rds-dbcluster.html
type RDSDBCluster struct {
	// A list of Availability Zones (AZs) in which DB instances in the
	// cluster can be created.
	AvailabilityZones *StringExpr `json:"AvailabilityZones,omitempty"`

	// The number of days for which automatic backups are retained. For more
	// information, see CreateDBCluster in the Amazon Relational Database
	// Service API Reference.
	BackupRetentionPeriod *IntegerExpr `json:"BackupRetentionPeriod,omitempty"`

	// The name of your database. You can specify a name of up to eight
	// alpha-numeric characters. If you do not provide a name, Amazon
	// Relational Database Service (Amazon RDS) won't create a database in
	// this DB cluster.
	DatabaseName *StringExpr `json:"DatabaseName,omitempty"`

	// The name of the DB cluster parameter group to associate with this DB
	// cluster. For the default value, see the DBClusterParameterGroupName
	// parameter of the CreateDBCluster action in the Amazon Relational
	// Database Service API Reference.
	DBClusterParameterGroupName *StringExpr `json:"DBClusterParameterGroupName,omitempty"`

	// A DB subnet group that you want to associate with this DB cluster.
	DBSubnetGroupName *StringExpr `json:"DBSubnetGroupName,omitempty"`

	// The name of the database engine that you want to use for this DB
	// cluster.
	Engine *StringExpr `json:"Engine,omitempty"`

	// The version number of the database engine that you want to use.
	EngineVersion *StringExpr `json:"EngineVersion,omitempty"`

	// The Amazon Resource Name (ARN) of the AWS Key Management Service
	// master key that is used to encrypt the database instances in the DB
	// cluster, such as
	// arn:aws:kms:us-east-1:012345678910:key/abcd1234-a123-456a-a12b-a123b4cd56ef.
	// If you enable the StorageEncrypted property but don't specify this
	// property, the default master key is used. If you specify this
	// property, you must set the StorageEncrypted property to true.
	KmsKeyId *StringExpr `json:"KmsKeyId,omitempty"`

	// The master user name for the DB instance.
	MasterUsername *StringExpr `json:"MasterUsername,omitempty"`

	// The password for the master database user.
	MasterUserPassword *StringExpr `json:"MasterUserPassword,omitempty"`

	// The port number on which the DB instances in the cluster can accept
	// connections.
	Port *IntegerExpr `json:"Port,omitempty"`

	// if automated backups are enabled (see the BackupRetentionPeriod
	// property), the daily time range in UTC during which you want to create
	// automated backups.
	PreferredBackupWindow *StringExpr `json:"PreferredBackupWindow,omitempty"`

	// The weekly time range (in UTC) during which system maintenance can
	// occur.
	PreferredMaintenanceWindow *StringExpr `json:"PreferredMaintenanceWindow,omitempty"`

	// The identifier for the DB cluster snapshot from which you want to
	// restore.
	SnapshotIdentifier *StringExpr `json:"SnapshotIdentifier,omitempty"`

	// Indicates whether the DB instances in the cluster are encrypted.
	StorageEncrypted *BoolExpr `json:"StorageEncrypted,omitempty"`

	// The tags that you want to attach to this DB cluster.
	Tags *ResourceTagList `json:"Tags,omitempty"`

	// A list of VPC security groups to associate with this DB cluster.
	VpcSecurityGroupIds *StringListExpr `json:"VpcSecurityGroupIds,omitempty"`
}

// CfnResourceType returns AWS::RDS::DBCluster to implement the ResourceProperties interface
func (s RDSDBCluster) CfnResourceType() string {
	return "AWS::RDS::DBCluster"
}

// RDSDBClusterParameterGroup represents AWS::RDS::DBClusterParameterGroup
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-rds-dbclusterparametergroup.html
type RDSDBClusterParameterGroup struct {
	// A friendly description for this DB cluster parameter group.
	Description interface{} `json:"Description,omitempty"`

	// The database family of this DB cluster parameter group, such as
	// aurora5.6.
	Family interface{} `json:"Family,omitempty"`

	// The parameters to set for this DB cluster parameter group. For a list
	// of parameter keys, see Appendix: DB Cluster and DB Instance Parameters
	// in the Amazon Relational Database Service User Guide.
	Parameters interface{} `json:"Parameters,omitempty"`

	// The tags that you want to attach to this parameter group.
	Tags *ResourceTagList `json:"Tags,omitempty"`
}

// CfnResourceType returns AWS::RDS::DBClusterParameterGroup to implement the ResourceProperties interface
func (s RDSDBClusterParameterGroup) CfnResourceType() string {
	return "AWS::RDS::DBClusterParameterGroup"
}

// RDSDBInstance represents AWS::RDS::DBInstance
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-rds-database-instance.html
type RDSDBInstance struct {
	// The allocated storage size, specified in gigabytes (GB).
	AllocatedStorage *StringExpr `json:"AllocatedStorage,omitempty"`

	// Indicates whether major version upgrades are allowed. Setting this
	// parameter does not result in an outage, and the change is applied
	// asynchronously as soon as possible.
	AllowMajorVersionUpgrade *BoolExpr `json:"AllowMajorVersionUpgrade,omitempty"`

	// Indicates that minor engine upgrades are applied automatically to the
	// DB instance during the maintenance window. The default value is true.
	AutoMinorVersionUpgrade *BoolExpr `json:"AutoMinorVersionUpgrade,omitempty"`

	// The name of the Availability Zone where the DB instance is located.
	// You cannot set the AvailabilityZone parameter if the MultiAZ parameter
	// is set to true.
	AvailabilityZone *StringExpr `json:"AvailabilityZone,omitempty"`

	// The number of days during which automatic DB snapshots are retained.
	BackupRetentionPeriod *StringExpr `json:"BackupRetentionPeriod,omitempty"`

	// For supported engines, specifies the character set to associate with
	// the DB instance. For more information, see Appendix: Oracle Character
	// Sets Supported in Amazon RDS in the Amazon Relational Database Service
	// User Guide.
	CharacterSetName *StringExpr `json:"CharacterSetName,omitempty"`

	// Indicates whether to copy all of the user-defined tags from the DB
	// instance to snapshots of the DB instance. By default, Amazon RDS
	// doesn't copy tags to snapshots. Amazon RDS doesn't copy tags with the
	// aws:: prefix unless it's the DB instance's final snapshot (the
	// snapshot when you delete the DB instance).
	CopyTagsToSnapshot *BoolExpr `json:"CopyTagsToSnapshot,omitempty"`

	// The name of an existing DB cluster that this instance will be
	// associated with. If you specify this property, specify aurora for the
	// Engine property and do not specify any of the following properties:
	// AllocatedStorage, BackupRetentionPeriod, CharacterSetName,
	// DBSecurityGroups, PreferredBackupWindow, PreferredMaintenanceWindow,
	// Port, SourceDBInstanceIdentifier, or StorageType.
	DBClusterIdentifier *StringExpr `json:"DBClusterIdentifier,omitempty"`

	// The name of the compute and memory capacity classes of the DB
	// instance.
	DBInstanceClass *StringExpr `json:"DBInstanceClass,omitempty"`

	// A name for the DB instance. If you specify a name, AWS CloudFormation
	// converts it to lower case. If you don't specify a name, AWS
	// CloudFormation generates a unique physical ID and uses that ID for the
	// DB instance. For more information, see Name Type.
	DBInstanceIdentifier *StringExpr `json:"DBInstanceIdentifier,omitempty"`

	// The name of the DB instance that was provided at the time of creation,
	// if one was specified. This same name is returned for the life of the
	// DB instance.
	DBName *StringExpr `json:"DBName,omitempty"`

	// The name of an existing DB parameter group or a reference to an
	// AWS::RDS::DBParameterGroup resource created in the template.
	DBParameterGroupName *StringExpr `json:"DBParameterGroupName,omitempty"`

	// A list of the DB security groups to assign to the DB instance. The
	// list can include both the name of existing DB security groups or
	// references to AWS::RDS::DBSecurityGroup resources created in the
	// template.
	DBSecurityGroups *StringListExpr `json:"DBSecurityGroups,omitempty"`

	// The name or ARN of the DB snapshot used to restore the DB instance. If
	// you are restoring from a shared manual DB snapshot, you must specify
	// the Amazon Resource Name (ARN) of the snapshot.
	DBSnapshotIdentifier *StringExpr `json:"DBSnapshotIdentifier,omitempty"`

	// A DB subnet group to associate with the DB instance.
	DBSubnetGroupName *StringExpr `json:"DBSubnetGroupName,omitempty"`

	// For an Amazon RDS DB instance that is running Microsoft SQL Server,
	// the Active Directory directory ID to create the instance in. Amazon
	// RDS uses Windows Authentication to authenticate users that connect to
	// the DB instance. For more information, see Using Windows
	// Authentication with an Amazon RDS DB Instance Running Microsoft SQL
	// Server in the Amazon Relational Database Service User Guide.
	Domain *StringExpr `json:"Domain,omitempty"`

	// The name of an IAM role that Amazon RDS uses when calling the
	// Directory Service APIs.
	DomainIAMRoleName *StringExpr `json:"DomainIAMRoleName,omitempty"`

	// The database engine that the DB instance uses. This property is
	// optional when you specify the DBSnapshotIdentifier property to create
	// DB instances.
	Engine *StringExpr `json:"Engine,omitempty"`

	// The version number of the database engine that the DB instance uses.
	EngineVersion *StringExpr `json:"EngineVersion,omitempty"`

	// The number of I/O operations per second (IOPS) that the database
	// provisions. The value must be equal to or greater than 1000.
	Iops *IntegerExpr `json:"Iops,omitempty"`

	// The ARN of the AWS Key Management Service (AWS KMS) master key that is
	// used to encrypt the DB instance, such as
	// arn:aws:kms:us-east-1:012345678910:key/abcd1234-a123-456a-a12b-a123b4cd56ef.
	// If you enable the StorageEncrypted property but don't specify this
	// property, AWS CloudFormation uses the default master key. If you
	// specify this property, you must set the StorageEncrypted property to
	// true.
	KmsKeyId *StringExpr `json:"KmsKeyId,omitempty"`

	// The license model of the DB instance.
	LicenseModel *StringExpr `json:"LicenseModel,omitempty"`

	// The master user name for the DB instance.
	MasterUsername *StringExpr `json:"MasterUsername,omitempty"`

	// The master password for the DB instance.
	MasterUserPassword *StringExpr `json:"MasterUserPassword,omitempty"`

	// The interval, in seconds, between points when Amazon RDS collects
	// enhanced monitoring metrics for the DB instance. To disable metrics
	// collection, specify 0.
	MonitoringInterval *IntegerExpr `json:"MonitoringInterval,omitempty"`

	// The ARN of the AWS Identity and Access Management (IAM) role that
	// permits Amazon RDS to send enhanced monitoring metrics to Amazon
	// CloudWatch, for example, arn:aws:iam:123456789012:role/emaccess. For
	// information on creating a monitoring role, see To create an IAM role
	// for Amazon RDS Enhanced Monitoring in the Amazon Relational Database
	// Service User Guide.
	MonitoringRoleArn *StringExpr `json:"MonitoringRoleArn,omitempty"`

	// Specifies if the database instance is a multiple Availability Zone
	// deployment. You cannot set the AvailabilityZone parameter if the
	// MultiAZ parameter is set to true.
	MultiAZ *BoolExpr `json:"MultiAZ,omitempty"`

	// The option group that this DB instance is associated with.
	OptionGroupName *StringExpr `json:"OptionGroupName,omitempty"`

	// The port for the instance.
	Port *StringExpr `json:"Port,omitempty"`

	// The daily time range during which automated backups are performed if
	// automated backups are enabled, as determined by the
	// BackupRetentionPeriod property. For valid values, see the
	// PreferredBackupWindow parameter for the CreateDBInstance action in the
	// Amazon Relational Database Service API Reference.
	PreferredBackupWindow *StringExpr `json:"PreferredBackupWindow,omitempty"`

	// The weekly time range (in UTC) during which system maintenance can
	// occur. For valid values, see the PreferredMaintenanceWindow parameter
	// for the CreateDBInstance action in the Amazon Relational Database
	// Service API Reference.
	PreferredMaintenanceWindow *StringExpr `json:"PreferredMaintenanceWindow,omitempty"`

	// Indicates whether the DB instance is an Internet-facing instance. If
	// you specify true, AWS CloudFormation creates an instance with a
	// publicly resolvable DNS name, which resolves to a public IP address.
	// If you specify false, AWS CloudFormation creates an internal instance
	// with a DNS name that resolves to a private IP address.
	PubliclyAccessible *BoolExpr `json:"PubliclyAccessible,omitempty"`

	// If you want to create a read replica DB instance, specify the ID of
	// the source DB instance. Each DB instance can have a limited number of
	// read replicas. For more information, see Working with Read Replicas in
	// the Amazon Relational Database Service Developer Guide.
	SourceDBInstanceIdentifier *StringExpr `json:"SourceDBInstanceIdentifier,omitempty"`

	// Indicates whether the DB instance is encrypted.
	StorageEncrypted *BoolExpr `json:"StorageEncrypted,omitempty"`

	// The storage type associated with this DB instance.
	StorageType *StringExpr `json:"StorageType,omitempty"`

	// An arbitrary set of tags (key–value pairs) for this DB instance.
	Tags []ResourceTag `json:"Tags,omitempty"`

	// A list of the VPC security group IDs to assign to the DB instance. The
	// list can include both the physical IDs of existing VPC security groups
	// and references to AWS::EC2::SecurityGroup resources created in the
	// template.
	VPCSecurityGroups *StringListExpr `json:"VPCSecurityGroups,omitempty"`
}

// CfnResourceType returns AWS::RDS::DBInstance to implement the ResourceProperties interface
func (s RDSDBInstance) CfnResourceType() string {
	return "AWS::RDS::DBInstance"
}

// RDSDBParameterGroup represents AWS::RDS::DBParameterGroup
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-rds-dbparametergroup.html
type RDSDBParameterGroup struct {
	// A friendly description of the RDS parameter group. For example, "My
	// Parameter Group".
	Description interface{} `json:"Description,omitempty"`

	// The database family of this RDS parameter group. For example,
	// "MySQL5.1".
	Family interface{} `json:"Family,omitempty"`

	// The parameters to set for this RDS parameter group.
	Parameters interface{} `json:"Parameters,omitempty"`

	// The tags that you want to attach to the RDS parameter group.
	Tags *ResourceTagList `json:"Tags,omitempty"`
}

// CfnResourceType returns AWS::RDS::DBParameterGroup to implement the ResourceProperties interface
func (s RDSDBParameterGroup) CfnResourceType() string {
	return "AWS::RDS::DBParameterGroup"
}

// RDSDBSecurityGroup represents AWS::RDS::DBSecurityGroup
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-rds-security-group.html
type RDSDBSecurityGroup struct {
	// The Id of VPC. Indicates which VPC this DB Security Group should
	// belong to.
	EC2VpcId *StringExpr `json:"EC2VpcId,omitempty"`

	// Network ingress authorization for an Amazon EC2 security group or an
	// IP address range.
	DBSecurityGroupIngress *RDSSecurityGroupRuleList `json:"DBSecurityGroupIngress,omitempty"`

	// Description of the security group.
	GroupDescription *StringExpr `json:"GroupDescription,omitempty"`

	// The tags that you want to attach to the Amazon RDS DB security group.
	Tags *ResourceTagList `json:"Tags,omitempty"`
}

// CfnResourceType returns AWS::RDS::DBSecurityGroup to implement the ResourceProperties interface
func (s RDSDBSecurityGroup) CfnResourceType() string {
	return "AWS::RDS::DBSecurityGroup"
}

// RDSDBSecurityGroupIngress represents AWS::RDS::DBSecurityGroupIngress
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-rds-security-group-ingress.html
type RDSDBSecurityGroupIngress struct {
	// The IP range to authorize.
	CIDRIP *StringExpr `json:"CIDRIP,omitempty"`

	// The name (ARN) of the AWS::RDS::DBSecurityGroup to which this ingress
	// will be added.
	DBSecurityGroupName *StringExpr `json:"DBSecurityGroupName,omitempty"`

	// The ID of the VPC or EC2 security group to authorize.
	EC2SecurityGroupId *StringExpr `json:"EC2SecurityGroupId,omitempty"`

	// The name of the EC2 security group to authorize.
	EC2SecurityGroupName *StringExpr `json:"EC2SecurityGroupName,omitempty"`

	// The AWS Account Number of the owner of the EC2 security group
	// specified in the EC2SecurityGroupName parameter. The AWS Access Key ID
	// is not an acceptable value.
	EC2SecurityGroupOwnerId *StringExpr `json:"EC2SecurityGroupOwnerId,omitempty"`
}

// CfnResourceType returns AWS::RDS::DBSecurityGroupIngress to implement the ResourceProperties interface
func (s RDSDBSecurityGroupIngress) CfnResourceType() string {
	return "AWS::RDS::DBSecurityGroupIngress"
}

// RDSDBSubnetGroup represents AWS::RDS::DBSubnetGroup
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-rds-dbsubnet-group.html
type RDSDBSubnetGroup struct {
	// The description for the DB Subnet Group.
	DBSubnetGroupDescription *StringExpr `json:"DBSubnetGroupDescription,omitempty"`

	// The EC2 Subnet IDs for the DB Subnet Group.
	SubnetIds *StringListExpr `json:"SubnetIds,omitempty"`

	// The tags that you want to attach to the RDS database subnet group.
	Tags *ResourceTagList `json:"Tags,omitempty"`
}

// CfnResourceType returns AWS::RDS::DBSubnetGroup to implement the ResourceProperties interface
func (s RDSDBSubnetGroup) CfnResourceType() string {
	return "AWS::RDS::DBSubnetGroup"
}

// RDSEventSubscription represents AWS::RDS::EventSubscription
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-rds-eventsubscription.html
type RDSEventSubscription struct {
	// Indicates whether to activate the subscription. If you don't specify
	// this property, AWS CloudFormation activates the subscription.
	Enabled *BoolExpr `json:"Enabled,omitempty"`

	// A list of event categories that you want to subscribe to for a given
	// source type. If you don't specify this property, you are notified
	// about all event categories. For more information, see Using Amazon RDS
	// Event Notification in the Amazon Relational Database Service User
	// Guide.
	EventCategories *StringListExpr `json:"EventCategories,omitempty"`

	// The Amazon Resource Name (ARN) of an Amazon SNS topic that you want to
	// send event notifications to.
	SnsTopicArn *StringExpr `json:"SnsTopicArn,omitempty"`

	// A list of identifiers for which Amazon RDS provides notification
	// events.
	SourceIds *StringListExpr `json:"SourceIds,omitempty"`

	// The type of source for which Amazon RDS provides notification events.
	// For example, if you want to be notified of events generated by a
	// database instance, set this parameter to db-instance. If you don't
	// specify a value, notifications are provided for all source types. For
	// valid values, see the SourceType parameter for the
	// CreateEventSubscription action in the Amazon Relational Database
	// Service API Reference.
	SourceType *StringExpr `json:"SourceType,omitempty"`
}

// CfnResourceType returns AWS::RDS::EventSubscription to implement the ResourceProperties interface
func (s RDSEventSubscription) CfnResourceType() string {
	return "AWS::RDS::EventSubscription"
}

// RDSOptionGroup represents AWS::RDS::OptionGroup
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-rds-optiongroup.html
type RDSOptionGroup struct {
	// The name of the database engine that this option group is associated
	// with.
	EngineName *StringExpr `json:"EngineName,omitempty"`

	// The major version number of the database engine that this option group
	// is associated with.
	MajorEngineVersion *StringExpr `json:"MajorEngineVersion,omitempty"`

	// A description of the option group.
	OptionGroupDescription *StringExpr `json:"OptionGroupDescription,omitempty"`

	// The configurations for this option group.
	OptionConfigurations *RDSOptionGroupOptionConfigurationsList `json:"OptionConfigurations,omitempty"`

	// An arbitrary set of tags (key–value pairs) for this option group.
	Tags []ResourceTag `json:"Tags,omitempty"`
}

// CfnResourceType returns AWS::RDS::OptionGroup to implement the ResourceProperties interface
func (s RDSOptionGroup) CfnResourceType() string {
	return "AWS::RDS::OptionGroup"
}

// RedshiftCluster represents AWS::Redshift::Cluster
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-redshift-cluster.html
type RedshiftCluster struct {
	// When a new version of the Amazon Redshift is released, indicates
	// whether upgrades can be applied to the engine that is running on the
	// cluster. The upgrades are applied during the maintenance window.
	AllowVersionUpgrade *BoolExpr `json:"AllowVersionUpgrade,omitempty"`

	// The number of days that automated snapshots are retained. If you set
	// the value to 0, automated snapshots are disabled.
	AutomatedSnapshotRetentionPeriod *IntegerExpr `json:"AutomatedSnapshotRetentionPeriod,omitempty"`

	// The Amazon EC2 Availability Zone in which you want to provision your
	// Amazon Redshift cluster. For example, if you have several Amazon EC2
	// instances running in a specific Availability Zone, you might want the
	// cluster to be provisioned in the same zone in order to decrease
	// network latency.
	AvailabilityZone *StringExpr `json:"AvailabilityZone,omitempty"`

	// The name of the parameter group that you want to associate with this
	// cluster.
	ClusterParameterGroupName *StringExpr `json:"ClusterParameterGroupName,omitempty"`

	// A list of security groups that you want to associate with this
	// cluster.
	ClusterSecurityGroups *StringListExpr `json:"ClusterSecurityGroups,omitempty"`

	// The name of a cluster subnet group that you want to associate with
	// this cluster.
	ClusterSubnetGroupName *StringExpr `json:"ClusterSubnetGroupName,omitempty"`

	// The type of cluster. You can specify single-node or multi-node.
	ClusterType *StringExpr `json:"ClusterType,omitempty"`

	// The Amazon Redshift engine version that you want to deploy on the
	// cluster.
	ClusterVersion *StringExpr `json:"ClusterVersion,omitempty"`

	// The name of the first database that is created when the cluster is
	// created.
	DBName *StringExpr `json:"DBName,omitempty"`

	// The Elastic IP (EIP) address for the cluster.
	ElasticIp *StringExpr `json:"ElasticIp,omitempty"`

	// Indicates whether the data in the cluster is encrypted at rest.
	Encrypted *BoolExpr `json:"Encrypted,omitempty"`

	// Specifies the name of the HSM client certificate that the Amazon
	// Redshift cluster uses to retrieve the data encryption keys stored in
	// an HSM.
	HsmClientCertificateIdentifier *StringExpr `json:"HsmClientCertificateIdentifier,omitempty"`

	// Specifies the name of the HSM configuration that contains the
	// information that the Amazon Redshift cluster can use to retrieve and
	// store keys in an HSM.
	HsmConfigurationIdentifier *StringExpr `json:"HsmConfigurationIdentifier,omitempty"`

	// The AWS Key Management Service (AWS KMS) key ID that you want to use
	// to encrypt data in the cluster.
	KmsKeyId *StringExpr `json:"KmsKeyId,omitempty"`

	// The user name that is associated with the master user account for this
	// cluster.
	MasterUsername *StringExpr `json:"MasterUsername,omitempty"`

	// The password associated with the master user account for this cluster.
	MasterUserPassword *StringExpr `json:"MasterUserPassword,omitempty"`

	// The node type that is provisioned for this cluster.
	NodeType *StringExpr `json:"NodeType,omitempty"`

	// The number of compute nodes in the cluster. If you specify multi-node
	// for the ClusterType parameter, you must specify a number greater than
	// 1.
	NumberOfNodes *IntegerExpr `json:"NumberOfNodes,omitempty"`

	// When you restore from a snapshot from another AWS account, the
	// 12-digit AWS account ID that contains that snapshot.
	OwnerAccount *StringExpr `json:"OwnerAccount,omitempty"`

	// The port number on which the cluster accepts incoming connections.
	Port *IntegerExpr `json:"Port,omitempty"`

	// The weekly time range (in UTC) during which automated cluster
	// maintenance can occur. The format of the time range is
	// ddd:hh24:mi-ddd:hh24:mi.
	PreferredMaintenanceWindow *StringExpr `json:"PreferredMaintenanceWindow,omitempty"`

	// Indicates whether the cluster can be accessed from a public network.
	PubliclyAccessible *BoolExpr `json:"PubliclyAccessible,omitempty"`

	// The name of the cluster the source snapshot was created from. For more
	// information about restoring from a snapshot, see the
	// RestoreFromClusterSnapshot action in the Amazon Redshift API
	// Reference.
	SnapshotClusterIdentifier interface{} `json:"SnapshotClusterIdentifier,omitempty"`

	// The name of the snapshot from which to create a new cluster.
	SnapshotIdentifier *StringExpr `json:"SnapshotIdentifier,omitempty"`

	// A list of VPC security groups that are associated with this cluster.
	VpcSecurityGroupIds *StringListExpr `json:"VpcSecurityGroupIds,omitempty"`
}

// CfnResourceType returns AWS::Redshift::Cluster to implement the ResourceProperties interface
func (s RedshiftCluster) CfnResourceType() string {
	return "AWS::Redshift::Cluster"
}

// RedshiftClusterParameterGroup represents AWS::Redshift::ClusterParameterGroup
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-redshift-clusterparametergroup.html
type RedshiftClusterParameterGroup struct {
	// A description of the parameter group.
	Description *StringExpr `json:"Description,omitempty"`

	// The Amazon Redshift engine version that applies to this cluster
	// parameter group. The cluster engine version determines the set of
	// parameters that you can specify in the Parameters property.
	ParameterGroupFamily *StringExpr `json:"ParameterGroupFamily,omitempty"`

	// A list of parameter names and values that are allowed by the Amazon
	// Redshift engine version that you specified in the ParameterGroupFamily
	// property. For more information, see Amazon Redshift Parameter Groups
	// in the Amazon Redshift Cluster Management Guide.
	Parameters *RedshiftParameterList `json:"Parameters,omitempty"`
}

// CfnResourceType returns AWS::Redshift::ClusterParameterGroup to implement the ResourceProperties interface
func (s RedshiftClusterParameterGroup) CfnResourceType() string {
	return "AWS::Redshift::ClusterParameterGroup"
}

// RedshiftClusterSecurityGroup represents AWS::Redshift::ClusterSecurityGroup
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-redshift-clustersecuritygroup.html
type RedshiftClusterSecurityGroup struct {
	// A description of the security group.
	Description *StringExpr `json:"Description,omitempty"`
}

// CfnResourceType returns AWS::Redshift::ClusterSecurityGroup to implement the ResourceProperties interface
func (s RedshiftClusterSecurityGroup) CfnResourceType() string {
	return "AWS::Redshift::ClusterSecurityGroup"
}

// RedshiftClusterSecurityGroupIngress represents AWS::Redshift::ClusterSecurityGroupIngress
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-redshift-clustersecuritygroupingress.html
type RedshiftClusterSecurityGroupIngress struct {
	// The name of the Amazon Redshift security group that will be associated
	// with the ingress rule.
	ClusterSecurityGroupName *StringExpr `json:"ClusterSecurityGroupName,omitempty"`

	// The IP address range that has inbound access to the Amazon Redshift
	// security group.
	CIDRIP *StringExpr `json:"CIDRIP,omitempty"`

	// The Amazon EC2 security group that will be added the Amazon Redshift
	// security group.
	EC2SecurityGroupName *StringExpr `json:"EC2SecurityGroupName,omitempty"`

	// The 12-digit AWS account number of the owner of the Amazon EC2
	// security group that is specified by the EC2SecurityGroupName
	// parameter.
	EC2SecurityGroupOwnerId *StringExpr `json:"EC2SecurityGroupOwnerId,omitempty"`
}

// CfnResourceType returns AWS::Redshift::ClusterSecurityGroupIngress to implement the ResourceProperties interface
func (s RedshiftClusterSecurityGroupIngress) CfnResourceType() string {
	return "AWS::Redshift::ClusterSecurityGroupIngress"
}

// RedshiftClusterSubnetGroup represents AWS::Redshift::ClusterSubnetGroup
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-redshift-clustersubnetgroup.html
type RedshiftClusterSubnetGroup struct {
	// A description of the subnet group.
	Description *StringExpr `json:"Description,omitempty"`

	// A list of VPC subnet IDs. You can modify a maximum of 20 subnets.
	SubnetIds *StringListExpr `json:"SubnetIds,omitempty"`
}

// CfnResourceType returns AWS::Redshift::ClusterSubnetGroup to implement the ResourceProperties interface
func (s RedshiftClusterSubnetGroup) CfnResourceType() string {
	return "AWS::Redshift::ClusterSubnetGroup"
}

// Route53HealthCheck represents AWS::Route53::HealthCheck
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-route53-healthcheck.html
type Route53HealthCheck struct {
	// An Amazon Route 53 health check.
	HealthCheckConfig *Route53HealthCheckConfig `json:"HealthCheckConfig,omitempty"`

	// An arbitrary set of tags (key–value pairs) for this health check.
	HealthCheckTags *Route53HealthCheckTagsList `json:"HealthCheckTags,omitempty"`
}

// CfnResourceType returns AWS::Route53::HealthCheck to implement the ResourceProperties interface
func (s Route53HealthCheck) CfnResourceType() string {
	return "AWS::Route53::HealthCheck"
}

// Route53HostedZone represents AWS::Route53::HostedZone
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-route53-hostedzone.html
type Route53HostedZone struct {
	// A complex type that contains an optional comment about your hosted
	// zone.
	HostedZoneConfig *Route53HostedZoneConfigProperty `json:"HostedZoneConfig,omitempty"`

	// An arbitrary set of tags (key–value pairs) for this hosted zone.
	HostedZoneTags *Route53HostedZoneTagsList `json:"HostedZoneTags,omitempty"`

	// The name of the domain. For resource record types that include a
	// domain name, specify a fully qualified domain name.
	Name *StringExpr `json:"Name,omitempty"`

	// One or more VPCs that you want to associate with this hosted zone.
	// When you specify this property, AWS CloudFormation creates a private
	// hosted zone.
	VPCs *Route53HostedZoneVPCsList `json:"VPCs,omitempty"`
}

// CfnResourceType returns AWS::Route53::HostedZone to implement the ResourceProperties interface
func (s Route53HostedZone) CfnResourceType() string {
	return "AWS::Route53::HostedZone"
}

// Route53RecordSet represents AWS::Route53::RecordSet
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-route53-recordset.html
type Route53RecordSet struct {
	// Alias resource record sets only: Information about the domain to which
	// you are redirecting traffic.
	AliasTarget *Route53AliasTargetProperty `json:"AliasTarget,omitempty"`

	// Any comments that you want to include about the hosted zone.
	Comment *StringExpr `json:"Comment,omitempty"`

	// Designates the record set as a PRIMARY or SECONDARY failover record
	// set. When you have more than one resource performing the same
	// function, you can configure Amazon Route 53 to check the health of
	// your resources and use only health resources to respond to DNS
	// queries. You cannot create nonfailover resource record sets that have
	// the same Name and Type property values as failover resource record
	// sets. For more information, see the Failover element in the Amazon
	// Route 53 API Reference.
	Failover *StringExpr `json:"Failover,omitempty"`

	// Describes how Amazon Route 53 responds to DNS queries based on the
	// geographic origin of the query.
	GeoLocation *Route53RecordSetGeoLocationProperty `json:"GeoLocation,omitempty"`

	// The health check ID that you want to apply to this record set. Amazon
	// Route 53 returns this resource record set in response to a DNS query
	// only while record set is healthy.
	HealthCheckId *StringExpr `json:"HealthCheckId,omitempty"`

	// The ID of the hosted zone.
	HostedZoneId *StringExpr `json:"HostedZoneId,omitempty"`

	// The name of the domain for the hosted zone where you want to add the
	// record set.
	HostedZoneName *StringExpr `json:"HostedZoneName,omitempty"`

	// The name of the domain. You must specify a fully qualified domain name
	// that ends with a period as the last label indication. If you omit the
	// final period, AWS CloudFormation adds it.
	Name *StringExpr `json:"Name,omitempty"`

	// Latency resource record sets only: The Amazon EC2 region where the
	// resource that is specified in this resource record set resides. The
	// resource typically is an AWS resource, for example, Amazon EC2
	// instance or an Elastic Load Balancing load balancer, and is referred
	// to by an IP address or a DNS domain name, depending on the record
	// type.
	Region interface{} `json:"Region,omitempty"`

	// List of resource records to add. Each record should be in the format
	// appropriate for the record type specified by the Type property. For
	// information about different record types and their record formats, see
	// Appendix: Domain Name Format in the Amazon Route 53 Developer Guide.
	ResourceRecords *StringListExpr `json:"ResourceRecords,omitempty"`

	// A unique identifier that differentiates among multiple resource record
	// sets that have the same combination of DNS name and type.
	SetIdentifier *StringExpr `json:"SetIdentifier,omitempty"`

	// The resource record cache time to live (TTL), in seconds. If you
	// specify this property, do not specify the AliasTarget property. For
	// alias target records, the alias uses a TTL value from the target.
	TTL *StringExpr `json:"TTL,omitempty"`

	// The type of records to add.
	Type *StringExpr `json:"Type,omitempty"`

	// Weighted resource record sets only: Among resource record sets that
	// have the same combination of DNS name and type, a value that
	// determines what portion of traffic for the current resource record set
	// is routed to the associated location.
	Weight *IntegerExpr `json:"Weight,omitempty"`
}

// CfnResourceType returns AWS::Route53::RecordSet to implement the ResourceProperties interface
func (s Route53RecordSet) CfnResourceType() string {
	return "AWS::Route53::RecordSet"
}

// Route53RecordSetList represents a list of Route53RecordSet
type Route53RecordSetList []Route53RecordSet

// UnmarshalJSON sets the object from the provided JSON representation
func (l *Route53RecordSetList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := Route53RecordSet{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = Route53RecordSetList{item}
		return nil
	}
	list := []Route53RecordSet{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = Route53RecordSetList(list)
		return nil
	}
	return err
}

// Route53RecordSetGroup represents AWS::Route53::RecordSetGroup
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-route53-recordsetgroup.html
type Route53RecordSetGroup struct {
	// Any comments you want to include about the hosted zone.
	Comment *StringExpr `json:"Comment,omitempty"`

	// The ID of the hosted zone.
	HostedZoneId *StringExpr `json:"HostedZoneId,omitempty"`

	// The name of the domain for the hosted zone where you want to add the
	// record set.
	HostedZoneName *StringExpr `json:"HostedZoneName,omitempty"`

	// List of resource record sets to add.
	RecordSets *Route53RecordSetList `json:"RecordSets,omitempty"`
}

// CfnResourceType returns AWS::Route53::RecordSetGroup to implement the ResourceProperties interface
func (s Route53RecordSetGroup) CfnResourceType() string {
	return "AWS::Route53::RecordSetGroup"
}

// S3Bucket represents AWS::S3::Bucket
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket.html
type S3Bucket struct {
	// A canned access control list (ACL) that grants predefined permissions
	// to the bucket. For more information about canned ACLs, see Canned ACLs
	// in the Amazon S3 documentation.
	AccessControl *StringExpr `json:"AccessControl,omitempty"`

	// A name for the bucket. If you don't specify a name, AWS CloudFormation
	// generates a unique physical ID and uses that ID for the bucket name.
	// For more information, see Name Type. The bucket name must contain only
	// lowercase letters, numbers, periods (.), and dashes (-).
	BucketName *StringExpr `json:"BucketName,omitempty"`

	// Rules that define cross-origin resource sharing of objects in this
	// bucket. For more information, see Enabling Cross-Origin Resource
	// Sharing in the Amazon Simple Storage Service Developer Guide.
	CorsConfiguration *S3CorsConfiguration `json:"CorsConfiguration,omitempty"`

	// Rules that define how Amazon S3 manages objects during their lifetime.
	// For more information, see Object Lifecycle Management in the Amazon
	// Simple Storage Service Developer Guide.
	LifecycleConfiguration *S3LifecycleConfiguration `json:"LifecycleConfiguration,omitempty"`

	// Settings that defines where logs are stored.
	LoggingConfiguration *S3LoggingConfiguration `json:"LoggingConfiguration,omitempty"`

	// Configuration that defines how Amazon S3 handles bucket notifications.
	NotificationConfiguration *S3NotificationConfiguration `json:"NotificationConfiguration,omitempty"`

	// Configuration for replicating objects in an S3 bucket. To enable
	// replication, you must also enable versioning by using the
	// VersioningConfiguration property.
	ReplicationConfiguration *S3ReplicationConfiguration `json:"ReplicationConfiguration,omitempty"`

	// An arbitrary set of tags (key-value pairs) for this Amazon S3 bucket.
	Tags []ResourceTag `json:"Tags,omitempty"`

	// Enables multiple variants of all objects in this bucket. You might
	// enable versioning to prevent objects from being deleted or overwritten
	// by mistake or to archive objects so that you can retrieve previous
	// versions of them.
	VersioningConfiguration *S3VersioningConfiguration `json:"VersioningConfiguration,omitempty"`

	// Information used to configure the bucket as a static website. For more
	// information, see Hosting Websites on Amazon S3.
	WebsiteConfiguration *S3WebsiteConfigurationProperty `json:"WebsiteConfiguration,omitempty"`
}

// CfnResourceType returns AWS::S3::Bucket to implement the ResourceProperties interface
func (s S3Bucket) CfnResourceType() string {
	return "AWS::S3::Bucket"
}

// S3BucketPolicy represents AWS::S3::BucketPolicy
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-policy.html
type S3BucketPolicy struct {
	// The Amazon S3 bucket that the policy applies to.
	Bucket *StringExpr `json:"Bucket,omitempty"`

	// A policy document containing permissions to add to the specified
	// bucket. For more information, see Access Policy Language Overview in
	// the Amazon Simple Storage Service Developer Guide.
	PolicyDocument interface{} `json:"PolicyDocument,omitempty"`
}

// CfnResourceType returns AWS::S3::BucketPolicy to implement the ResourceProperties interface
func (s S3BucketPolicy) CfnResourceType() string {
	return "AWS::S3::BucketPolicy"
}

// SDBDomain represents AWS::SDB::Domain
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-simpledb.html
type SDBDomain struct {
	// Information about the Amazon SimpleDB domain.
	Description *StringExpr `json:"Description,omitempty"`
}

// CfnResourceType returns AWS::SDB::Domain to implement the ResourceProperties interface
func (s SDBDomain) CfnResourceType() string {
	return "AWS::SDB::Domain"
}

// SNSSubscription represents AWS::SNS::Subscription
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-sns-subscription.html
type SNSSubscription struct {
	// The endpoint that receives notifications from the Amazon SNS topic.
	// The endpoint value depends on the protocol that you specify. For more
	// information, see the Subscribe Endpoint parameter in the Amazon Simple
	// Notification Service API Reference.
	Endpoint *StringExpr `json:"Endpoint,omitempty"`

	// The subscription's protocol. For more information, see the Subscribe
	// Protocol parameter in the Amazon Simple Notification Service API
	// Reference.
	Protocol *StringExpr `json:"Protocol,omitempty"`

	// The Amazon Resource Name (ARN) of the topic to subscribe to.
	TopicArn *StringExpr `json:"TopicArn,omitempty"`
}

// CfnResourceType returns AWS::SNS::Subscription to implement the ResourceProperties interface
func (s SNSSubscription) CfnResourceType() string {
	return "AWS::SNS::Subscription"
}

// SNSTopic represents AWS::SNS::Topic
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-sns-topic.html
type SNSTopic struct {
	// A developer-defined string that can be used to identify this SNS
	// topic.
	DisplayName *StringExpr `json:"DisplayName,omitempty"`

	// The SNS subscriptions (endpoints) for this topic.
	Subscription *SNSSubscriptionPropertyList `json:"Subscription,omitempty"`

	// A name for the topic. If you don't specify a name, AWS CloudFormation
	// generates a unique physical ID and uses that ID for the topic name.
	// For more information, see Name Type.
	TopicName *StringExpr `json:"TopicName,omitempty"`
}

// CfnResourceType returns AWS::SNS::Topic to implement the ResourceProperties interface
func (s SNSTopic) CfnResourceType() string {
	return "AWS::SNS::Topic"
}

// SNSTopicPolicy represents AWS::SNS::TopicPolicy
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-sns-policy.html
type SNSTopicPolicy struct {
	// A policy document that contains permissions to add to the specified
	// SNS topics.
	PolicyDocument interface{} `json:"PolicyDocument,omitempty"`

	// The Amazon Resource Names (ARN) of the topics to which you want to add
	// the policy. You can use the Ref function to specify an AWS::SNS::Topic
	// resource.
	Topics interface{} `json:"Topics,omitempty"`
}

// CfnResourceType returns AWS::SNS::TopicPolicy to implement the ResourceProperties interface
func (s SNSTopicPolicy) CfnResourceType() string {
	return "AWS::SNS::TopicPolicy"
}

// SQSQueue represents AWS::SQS::Queue
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-sqs-queues.html
type SQSQueue struct {
	// The time in seconds that the delivery of all messages in the queue
	// will be delayed. You can specify an integer value of 0 to 900 (15
	// minutes). The default value is 0.
	DelaySeconds *IntegerExpr `json:"DelaySeconds,omitempty"`

	// The limit of how many bytes a message can contain before Amazon SQS
	// rejects it. You can specify an integer value from 1024 bytes (1 KiB)
	// to 262144 bytes (256 KiB). The default value is 262144 (256 KiB).
	MaximumMessageSize *IntegerExpr `json:"MaximumMessageSize,omitempty"`

	// The number of seconds Amazon SQS retains a message. You can specify an
	// integer value from 60 seconds (1 minute) to 1209600 seconds (14 days).
	// The default value is 345600 seconds (4 days).
	MessageRetentionPeriod *IntegerExpr `json:"MessageRetentionPeriod,omitempty"`

	// A name for the queue. If you don't specify a name, AWS CloudFormation
	// generates a unique physical ID and uses that ID for the queue name.
	// For more information, see Name Type.
	QueueName *StringExpr `json:"QueueName,omitempty"`

	// Specifies the duration, in seconds, that the ReceiveMessage action
	// call waits until a message is in the queue in order to include it in
	// the response, as opposed to returning an empty response if a message
	// is not yet available. You can specify an integer from 1 to 20. The
	// short polling is used as the default or when you specify 0 for this
	// property. For more information, see Amazon SQS Long Poll.
	ReceiveMessageWaitTimeSeconds *IntegerExpr `json:"ReceiveMessageWaitTimeSeconds,omitempty"`

	// Specifies an existing dead letter queue to receive messages after the
	// source queue (this queue) fails to process a message a specified
	// number of times.
	RedrivePolicy *SQSRedrivePolicy `json:"RedrivePolicy,omitempty"`

	// The length of time during which a message will be unavailable once a
	// message is delivered from the queue. This blocks other components from
	// receiving the same message and gives the initial component time to
	// process and delete the message from the queue.
	VisibilityTimeout *IntegerExpr `json:"VisibilityTimeout,omitempty"`
}

// CfnResourceType returns AWS::SQS::Queue to implement the ResourceProperties interface
func (s SQSQueue) CfnResourceType() string {
	return "AWS::SQS::Queue"
}

// SQSQueuePolicy represents AWS::SQS::QueuePolicy
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-sqs-policy.html
type SQSQueuePolicy struct {
	// A policy document containing permissions to add to the specified SQS
	// queues.
	PolicyDocument interface{} `json:"PolicyDocument,omitempty"`

	// The URLs of the queues to which you want to add the policy. You can
	// use the Ref function to specify an AWS::SQS::Queue resource.
	Queues *StringListExpr `json:"Queues,omitempty"`
}

// CfnResourceType returns AWS::SQS::QueuePolicy to implement the ResourceProperties interface
func (s SQSQueuePolicy) CfnResourceType() string {
	return "AWS::SQS::QueuePolicy"
}

// SSMDocument represents AWS::SSM::Document
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ssm-document.html
type SSMDocument struct {
	// A JSON object that describes an instance configuration. For more
	// information, see SSM Documents in the Amazon EC2 Simple Systems
	// Manager API Reference.
	Content interface{} `json:"Content,omitempty"`
}

// CfnResourceType returns AWS::SSM::Document to implement the ResourceProperties interface
func (s SSMDocument) CfnResourceType() string {
	return "AWS::SSM::Document"
}

// WAFByteMatchSet represents AWS::WAF::ByteMatchSet
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-waf-bytematchset.html
type WAFByteMatchSet struct {
	// Settings for the ByteMatchSet, such as the bytes (typically a string
	// that corresponds with ASCII characters) that you want AWS WAF to
	// search for in web requests.
	ByteMatchTuples *WAFByteMatchSetByteMatchTuplesList `json:"ByteMatchTuples,omitempty"`

	// A friendly name or description of the ByteMatchSet.
	Name *StringExpr `json:"Name,omitempty"`
}

// CfnResourceType returns AWS::WAF::ByteMatchSet to implement the ResourceProperties interface
func (s WAFByteMatchSet) CfnResourceType() string {
	return "AWS::WAF::ByteMatchSet"
}

// WAFIPSet represents AWS::WAF::IPSet
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-waf-ipset.html
type WAFIPSet struct {
	// The IP address type and IP address range (in CIDR notation) from which
	// web requests originate. If you associate the IPSet with a web ACL that
	// is associated with a Amazon CloudFront (CloudFront) distribution, this
	// descriptor is the value of one of the following fields in the
	// CloudFront access logs:
	IPSetDescriptors *WAFIPSetIPSetDescriptorsList `json:"IPSetDescriptors,omitempty"`

	// If the viewer did not use an HTTP proxy or a load balancer to send the
	// request
	CXIp interface{} `json:"c-ip,omitempty"`

	// If the viewer did use an HTTP proxy or a load balancer to send the
	// request
	XXForwardedXFor interface{} `json:"x-forwarded-for,omitempty"`

	// A friendly name or description of the IPSet.
	Name *StringExpr `json:"Name,omitempty"`
}

// CfnResourceType returns AWS::WAF::IPSet to implement the ResourceProperties interface
func (s WAFIPSet) CfnResourceType() string {
	return "AWS::WAF::IPSet"
}

// WAFRule represents AWS::WAF::Rule
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-waf-rule.html
type WAFRule struct {
	// A friendly name or description for the metrics of the rule. For valid
	// values, see the MetricName parameter for the CreateRule action in the
	// AWS WAF API Reference.
	MetricName *StringExpr `json:"MetricName,omitempty"`

	// A friendly name or description of the rule.
	Name *StringExpr `json:"Name,omitempty"`

	// The ByteMatchSet, IPSet, SizeConstraintSet, SqlInjectionMatchSet, or
	// XssMatchSet objects to include in a rule. If you add more than one
	// predicate to a rule, a request must match all conditions in order to
	// be allowed or blocked.
	Predicates *WAFRulePredicatesList `json:"Predicates,omitempty"`
}

// CfnResourceType returns AWS::WAF::Rule to implement the ResourceProperties interface
func (s WAFRule) CfnResourceType() string {
	return "AWS::WAF::Rule"
}

// WAFSizeConstraintSet represents AWS::WAF::SizeConstraintSet
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-waf-sizeconstraintset.html
type WAFSizeConstraintSet struct {
	// A friendly name or description for the SizeConstraintSet.
	Name *StringExpr `json:"Name,omitempty"`

	// The size constraint and the part of the web request to check.
	SizeConstraints *WAFSizeConstraintSetSizeConstraintList `json:"SizeConstraints,omitempty"`
}

// CfnResourceType returns AWS::WAF::SizeConstraintSet to implement the ResourceProperties interface
func (s WAFSizeConstraintSet) CfnResourceType() string {
	return "AWS::WAF::SizeConstraintSet"
}

// WAFSqlInjectionMatchSet represents AWS::WAF::SqlInjectionMatchSet
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-waf-sqlinjectionmatchset.html
type WAFSqlInjectionMatchSet struct {
	// A friendly name or description of the SqlInjectionMatchSet.
	Name *StringExpr `json:"Name,omitempty"`

	// The parts of web requests that you want AWS WAF to inspect for
	// malicious SQL code and, if you want AWS WAF to inspect a header, the
	// name of the header.
	SqlInjectionMatchTuples *WAFSqlInjectionMatchSetSqlInjectionMatchTuplesList `json:"SqlInjectionMatchTuples,omitempty"`
}

// CfnResourceType returns AWS::WAF::SqlInjectionMatchSet to implement the ResourceProperties interface
func (s WAFSqlInjectionMatchSet) CfnResourceType() string {
	return "AWS::WAF::SqlInjectionMatchSet"
}

// WAFWebACL represents AWS::WAF::WebACL
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-waf-webacl.html
type WAFWebACL struct {
	// The action that you want AWS WAF to take when a request doesn't match
	// the criteria in any of the rules that are associated with the web ACL.
	DefaultAction *WAFWebACLAction `json:"DefaultAction,omitempty"`

	// A friendly name or description for the Amazon CloudWatch metric of
	// this web ACL. For valid values, see the MetricName parameter of the
	// CreateWebACL action in the AWS WAF API Reference.
	MetricName *StringExpr `json:"MetricName,omitempty"`

	// A friendly name or description of the web ACL.
	Name *StringExpr `json:"Name,omitempty"`

	// The rules to associate with the web ACL and the settings for each
	// rule.
	Rules *WAFWebACLRulesList `json:"Rules,omitempty"`
}

// CfnResourceType returns AWS::WAF::WebACL to implement the ResourceProperties interface
func (s WAFWebACL) CfnResourceType() string {
	return "AWS::WAF::WebACL"
}

// WAFXssMatchSet represents AWS::WAF::XssMatchSet
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-waf-xssmatchset.html
type WAFXssMatchSet struct {
	// A friendly name or description for the XssMatchSet.
	Name *StringExpr `json:"Name,omitempty"`

	// The parts of web requests that you want to inspect for cross-site
	// scripting attacks.
	XssMatchTuples *WAFXssMatchSetXssMatchTupleList `json:"XssMatchTuples,omitempty"`
}

// CfnResourceType returns AWS::WAF::XssMatchSet to implement the ResourceProperties interface
func (s WAFXssMatchSet) CfnResourceType() string {
	return "AWS::WAF::XssMatchSet"
}

// WorkSpacesWorkspace represents AWS::WorkSpaces::Workspace
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-workspaces-workspace.html
type WorkSpacesWorkspace struct {
	// The identifier of the bundle from which you want to create the
	// workspace. A bundle specifies the details of the workspace, such as
	// the installed applications and the size of CPU, memory, and storage.
	// Use the DescribeWorkspaceBundles action to list the bundles that AWS
	// offers.
	BundleId *StringExpr `json:"BundleId,omitempty"`

	// The identifier of the AWS Directory Service directory in which you
	// want to create the workspace. The directory must already be registered
	// with Amazon WorkSpaces. Use the DescribeWorkspaceDirectories action to
	// list the directories that are available.
	DirectoryId *StringExpr `json:"DirectoryId,omitempty"`

	// The name of the user to which the workspace is assigned. This user
	// name must exist in the specified AWS Directory Service directory.
	UserName *StringExpr `json:"UserName,omitempty"`

	// Indicates whether Amazon WorkSpaces encrypts data stored on the root
	// volume (C: drive).
	RootVolumeEncryptionEnabled *BoolExpr `json:"RootVolumeEncryptionEnabled,omitempty"`

	// Indicates whether Amazon WorkSpaces encrypts data stored on the user
	// volume (D: drive).
	UserVolumeEncryptionEnabled *BoolExpr `json:"UserVolumeEncryptionEnabled,omitempty"`

	// The AWS Key Management Service (AWS KMS) key ID that Amazon WorkSpaces
	// uses to encrypt data stored on your workspace.
	VolumeEncryptionKey *StringExpr `json:"VolumeEncryptionKey,omitempty"`
}

// CfnResourceType returns AWS::WorkSpaces::Workspace to implement the ResourceProperties interface
func (s WorkSpacesWorkspace) CfnResourceType() string {
	return "AWS::WorkSpaces::Workspace"
}

// APIGatewayApiKeyStageKey represents Amazon API Gateway ApiKey StageKey
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-apitgateway-apikey-stagekey.html
type APIGatewayApiKeyStageKey struct {
	// The ID of a RestApi resource that includes the stage with which you
	// want to associate the API key.
	RestApiId *StringExpr `json:"RestApiId,omitempty"`

	// The name of the stage with which to associate the API key. The stage
	// must be included in the RestApi resource that you specified in the
	// RestApiId property.
	StageName *StringExpr `json:"StageName,omitempty"`
}

// APIGatewayApiKeyStageKeyList represents a list of APIGatewayApiKeyStageKey
type APIGatewayApiKeyStageKeyList []APIGatewayApiKeyStageKey

// UnmarshalJSON sets the object from the provided JSON representation
func (l *APIGatewayApiKeyStageKeyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := APIGatewayApiKeyStageKey{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = APIGatewayApiKeyStageKeyList{item}
		return nil
	}
	list := []APIGatewayApiKeyStageKey{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = APIGatewayApiKeyStageKeyList(list)
		return nil
	}
	return err
}

// APIGatewayDeploymentStageDescription represents Amazon API Gateway Deployment StageDescription
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-apitgateway-deployment-stagedescription.html
type APIGatewayDeploymentStageDescription struct {
	// Indicates whether cache clustering is enabled for the stage.
	CacheClusterEnabled *BoolExpr `json:"CacheClusterEnabled,omitempty"`

	// The size of the stage's cache cluster.
	CacheClusterSize *StringExpr `json:"CacheClusterSize,omitempty"`

	// Indicates whether the cached responses are encrypted.
	CacheDataEncrypted *BoolExpr `json:"CacheDataEncrypted,omitempty"`

	// The time-to-live (TTL) period, in seconds, that specifies how long API
	// Gateway caches responses.
	CacheTtlInSeconds *IntegerExpr `json:"CacheTtlInSeconds,omitempty"`

	// Indicates whether responses are cached and returned for requests. You
	// must enable a cache cluster on the stage to cache responses. For more
	// information, see Enable API Gateway Caching in a Stage to Enhance API
	// Performance in the API Gateway Developer Guide.
	CachingEnabled *BoolExpr `json:"CachingEnabled,omitempty"`

	// The identifier of the client certificate that API Gateway uses to call
	// your integration endpoints in the stage.
	ClientCertificateId *StringExpr `json:"ClientCertificateId,omitempty"`

	// Indicates whether data trace logging is enabled for methods in the
	// stage. API Gateway pushes these logs to Amazon CloudWatch Logs.
	DataTraceEnabled *BoolExpr `json:"DataTraceEnabled,omitempty"`

	// A description of the purpose of the stage.
	Description *StringExpr `json:"Description,omitempty"`

	// The logging level for this method. For valid values, see the
	// loggingLevel property of the Stage resource in the Amazon API Gateway
	// API Reference.
	LoggingLevel *StringExpr `json:"LoggingLevel,omitempty"`

	// Configures settings for all of the stage's methods.
	MethodSettings *APIGatewayDeploymentStageDescriptionMethodSettingList `json:"MethodSettings,omitempty"`

	// Indicates whether Amazon CloudWatch metrics are enabled for methods in
	// the stage.
	MetricsEnabled *BoolExpr `json:"MetricsEnabled,omitempty"`

	// The name of the stage, which API Gateway uses as the first path
	// segment in the invoke Uniform Resource Identifier (URI).
	StageName *StringExpr `json:"StageName,omitempty"`

	// The number of burst requests per second that API Gateway permits
	// across all APIs, stages, and methods in your AWS account. For more
	// information, see Manage API Request Throttling in the API Gateway
	// Developer Guide.
	ThrottlingBurstLimit *IntegerExpr `json:"ThrottlingBurstLimit,omitempty"`

	// The number of steady-state requests per second that API Gateway
	// permits across all APIs, stages, and methods in your AWS account. For
	// more information, see Manage API Request Throttling in the API Gateway
	// Developer Guide.
	ThrottlingRateLimit *IntegerExpr `json:"ThrottlingRateLimit,omitempty"`

	// A map that defines the stage variables. Variable names must consist of
	// alphanumeric characters, and the values must match the following
	// regular expression: [A-Za-z0-9-._~:/?#&amp;=,]+.
	Variables interface{} `json:"Variables,omitempty"`
}

// APIGatewayDeploymentStageDescriptionList represents a list of APIGatewayDeploymentStageDescription
type APIGatewayDeploymentStageDescriptionList []APIGatewayDeploymentStageDescription

// UnmarshalJSON sets the object from the provided JSON representation
func (l *APIGatewayDeploymentStageDescriptionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := APIGatewayDeploymentStageDescription{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = APIGatewayDeploymentStageDescriptionList{item}
		return nil
	}
	list := []APIGatewayDeploymentStageDescription{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = APIGatewayDeploymentStageDescriptionList(list)
		return nil
	}
	return err
}

// APIGatewayDeploymentStageDescriptionMethodSetting represents Amazon API Gateway Deployment StageDescription MethodSetting
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-apitgateway-deployment-stagedescription-methodsetting.html
type APIGatewayDeploymentStageDescriptionMethodSetting struct {
	// Indicates whether the cached responses are encrypted.
	CacheDataEncrypted *BoolExpr `json:"CacheDataEncrypted,omitempty"`

	// The time-to-live (TTL) period, in seconds, that specifies how long API
	// Gateway caches responses.
	CacheTtlInSeconds *IntegerExpr `json:"CacheTtlInSeconds,omitempty"`

	// Indicates whether responses are cached and returned for requests. You
	// must enable a cache cluster on the stage to cache responses. For more
	// information, see Enable API Gateway Caching in a Stage to Enhance API
	// Performance in the API Gateway Developer Guide.
	CachingEnabled *BoolExpr `json:"CachingEnabled,omitempty"`

	// Indicates whether data trace logging is enabled for methods in the
	// stage. API Gateway pushes these logs to Amazon CloudWatch Logs.
	DataTraceEnabled *BoolExpr `json:"DataTraceEnabled,omitempty"`

	// The HTTP method.
	HttpMethod *StringExpr `json:"HttpMethod,omitempty"`

	// The logging level for this method. For valid values, see the
	// loggingLevel property of the Stage resource in the Amazon API Gateway
	// API Reference.
	LoggingLevel *StringExpr `json:"LoggingLevel,omitempty"`

	// Indicates whether Amazon CloudWatch metrics are enabled for methods in
	// the stage.
	MetricsEnabled *BoolExpr `json:"MetricsEnabled,omitempty"`

	// The resource path for this method. Forward slashes (/) are encoded as
	// ~1 and the initial slash must include a forward slash. For example,
	// the path value /resource/subresource must be encoded as
	// /~1resource~1subresource. To specify the root path, use only a slash
	// (/).
	ResourcePath *StringExpr `json:"ResourcePath,omitempty"`

	// The number of burst requests per second that API Gateway permits
	// across all APIs, stages, and methods in your AWS account. For more
	// information, see Manage API Request Throttling in the API Gateway
	// Developer Guide.
	ThrottlingBurstLimit *IntegerExpr `json:"ThrottlingBurstLimit,omitempty"`

	// The number of steady-state requests per second that API Gateway
	// permits across all APIs, stages, and methods in your AWS account. For
	// more information, see Manage API Request Throttling in the API Gateway
	// Developer Guide.
	ThrottlingRateLimit *IntegerExpr `json:"ThrottlingRateLimit,omitempty"`
}

// APIGatewayDeploymentStageDescriptionMethodSettingList represents a list of APIGatewayDeploymentStageDescriptionMethodSetting
type APIGatewayDeploymentStageDescriptionMethodSettingList []APIGatewayDeploymentStageDescriptionMethodSetting

// UnmarshalJSON sets the object from the provided JSON representation
func (l *APIGatewayDeploymentStageDescriptionMethodSettingList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := APIGatewayDeploymentStageDescriptionMethodSetting{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = APIGatewayDeploymentStageDescriptionMethodSettingList{item}
		return nil
	}
	list := []APIGatewayDeploymentStageDescriptionMethodSetting{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = APIGatewayDeploymentStageDescriptionMethodSettingList(list)
		return nil
	}
	return err
}

// APIGatewayMethodIntegration represents Amazon API Gateway Method Integration
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-apitgateway-method-integration.html
type APIGatewayMethodIntegration struct {
	// A list of request parameters whose values API Gateway will cache.
	CacheKeyParameters *StringListExpr `json:"CacheKeyParameters,omitempty"`

	// An API-specific tag group of related cached parameters.
	CacheNamespace *StringExpr `json:"CacheNamespace,omitempty"`

	// The credentials required for the integration. To specify an AWS
	// Identity and Access Management (IAM) role that API Gateway assumes,
	// specify the role's Amazon Resource Name (ARN). To require that the
	// caller's identity be passed through from the request, specify
	// arn:aws:iam::*:user/*.
	Credentials *StringExpr `json:"Credentials,omitempty"`

	// The integration's HTTP method type.
	IntegrationHttpMethod *StringExpr `json:"IntegrationHttpMethod,omitempty"`

	// The response that API Gateway provides after a method's back end
	// completes processing a request. API Gateway intercepts the back end's
	// response so that you can control how API Gateway surfaces back-end
	// responses. For example, you can map the back-end status codes to codes
	// that you define.
	IntegrationResponses *APIGatewayMethodIntegrationIntegrationResponseList `json:"IntegrationResponses,omitempty"`

	// Indicates when API Gateway passes requests to the targeted back end.
	// This behavior depends on the request's Content-Type header and whether
	// you defined a mapping template for it.
	PassthroughBehavior *StringExpr `json:"PassthroughBehavior,omitempty"`

	// The request parameters that API Gateway sends with the back-end
	// request. Specify request parameters as key-value pairs
	// (string-to-string maps), with a destination as the key and a source as
	// the value.
	RequestParameters interface{} `json:"RequestParameters,omitempty"`

	// A map of Apache Velocity templates that are applied on the request
	// payload. The template that API Gateway uses is based on the value of
	// the Content-Type header sent by the client. The content type value is
	// the key, and the template is the value (specified as a string), such
	// as the following snippet:
	RequestTemplates interface{} `json:"RequestTemplates,omitempty"`

	// The type of back end your method is running, such as HTTP, AWS, or
	// MOCK. For valid values, see the type property in the Amazon API
	// Gateway REST API Reference.
	Type *StringExpr `json:"Type,omitempty"`

	// The integration's Uniform Resource Identifier (URI).
	Uri *StringExpr `json:"Uri,omitempty"`
}

// APIGatewayMethodIntegrationList represents a list of APIGatewayMethodIntegration
type APIGatewayMethodIntegrationList []APIGatewayMethodIntegration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *APIGatewayMethodIntegrationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := APIGatewayMethodIntegration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = APIGatewayMethodIntegrationList{item}
		return nil
	}
	list := []APIGatewayMethodIntegration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = APIGatewayMethodIntegrationList(list)
		return nil
	}
	return err
}

// APIGatewayMethodIntegrationIntegrationResponse represents Amazon API Gateway Method Integration IntegrationResponse
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-apitgateway-method-integration-integrationresponse.html
type APIGatewayMethodIntegrationIntegrationResponse struct {
	// The response parameters from the back-end response that API Gateway
	// sends to the method response. Specify response parameters as key-value
	// pairs (string-to-string mappings).
	ResponseParameters interface{} `json:"ResponseParameters,omitempty"`

	// The templates used to transform the integration response body. Specify
	// templates as key-value pairs (string-to-string maps), with a content
	// type as the key and a template as the value. For more information, see
	// API Gateway API Request and Response Payload-Mapping Template
	// Reference in the API Gateway Developer Guide.
	ResponseTemplates interface{} `json:"ResponseTemplates,omitempty"`

	// A regular expression that specifies which error strings or status
	// codes from the back end map to the integration response.
	SelectionPattern *StringExpr `json:"SelectionPattern,omitempty"`

	// The status code that API Gateway uses to map the integration response
	// to a MethodResponse status code.
	StatusCode *StringExpr `json:"StatusCode,omitempty"`
}

// APIGatewayMethodIntegrationIntegrationResponseList represents a list of APIGatewayMethodIntegrationIntegrationResponse
type APIGatewayMethodIntegrationIntegrationResponseList []APIGatewayMethodIntegrationIntegrationResponse

// UnmarshalJSON sets the object from the provided JSON representation
func (l *APIGatewayMethodIntegrationIntegrationResponseList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := APIGatewayMethodIntegrationIntegrationResponse{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = APIGatewayMethodIntegrationIntegrationResponseList{item}
		return nil
	}
	list := []APIGatewayMethodIntegrationIntegrationResponse{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = APIGatewayMethodIntegrationIntegrationResponseList(list)
		return nil
	}
	return err
}

// APIGatewayMethodMethodResponse represents Amazon API Gateway Method MethodResponse
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-apitgateway-method-methodresponse.html
type APIGatewayMethodMethodResponse struct {
	// The resources used for the response's content type. Specify response
	// models as key-value pairs (string-to-string maps), with a content type
	// as the key and a Model resource name as the value.
	ResponseModels interface{} `json:"ResponseModels,omitempty"`

	// Response parameters that API Gateway sends to the client that called a
	// method. Specify response parameters as key-value pairs
	// (string-to-Boolean maps), with a destination as the key and a Boolean
	// as the value. Specify the destination using the following pattern:
	// method.response.header.name, where the name is a valid, unique header
	// name. The Boolean specifies whether a parameter is required.
	ResponseParameters interface{} `json:"ResponseParameters,omitempty"`

	// The method response's status code, which you map to an
	// IntegrationResponse.
	StatusCode *StringExpr `json:"StatusCode,omitempty"`
}

// APIGatewayMethodMethodResponseList represents a list of APIGatewayMethodMethodResponse
type APIGatewayMethodMethodResponseList []APIGatewayMethodMethodResponse

// UnmarshalJSON sets the object from the provided JSON representation
func (l *APIGatewayMethodMethodResponseList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := APIGatewayMethodMethodResponse{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = APIGatewayMethodMethodResponseList{item}
		return nil
	}
	list := []APIGatewayMethodMethodResponse{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = APIGatewayMethodMethodResponseList(list)
		return nil
	}
	return err
}

// APIGatewayRestApiS3Location represents Amazon API Gateway RestApi S3Location
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-apitgateway-restapi-bodys3location.html
type APIGatewayRestApiS3Location struct {
	// The name of the S3 bucket where the Swagger file is stored.
	Bucket *StringExpr `json:"Bucket,omitempty"`

	// The Amazon S3 ETag (a file checksum) of the Swagger file. If you don't
	// specify a value, API Gateway skips ETag validation of your Swagger
	// file.
	ETag *StringExpr `json:"ETag,omitempty"`

	// The file name of the Swagger file (Amazon S3 object name).
	Key *StringExpr `json:"Key,omitempty"`

	// For versioning-enabled buckets, a specific version of the Swagger
	// file.
	Version *StringExpr `json:"Version,omitempty"`
}

// APIGatewayRestApiS3LocationList represents a list of APIGatewayRestApiS3Location
type APIGatewayRestApiS3LocationList []APIGatewayRestApiS3Location

// UnmarshalJSON sets the object from the provided JSON representation
func (l *APIGatewayRestApiS3LocationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := APIGatewayRestApiS3Location{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = APIGatewayRestApiS3LocationList{item}
		return nil
	}
	list := []APIGatewayRestApiS3Location{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = APIGatewayRestApiS3LocationList(list)
		return nil
	}
	return err
}

// APIGatewayStageMethodSetting represents Amazon API Gateway Stage MethodSetting
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-apitgateway-stage-methodsetting.html
type APIGatewayStageMethodSetting struct {
	// Indicates whether the cached responses are encrypted.
	CacheDataEncrypted *BoolExpr `json:"CacheDataEncrypted,omitempty"`

	// The time-to-live (TTL) period, in seconds, that specifies how long API
	// Gateway caches responses.
	CacheTtlInSeconds *IntegerExpr `json:"CacheTtlInSeconds,omitempty"`

	// Indicates whether responses are cached and returned for requests. You
	// must enable a cache cluster on the stage to cache responses.
	CachingEnabled *BoolExpr `json:"CachingEnabled,omitempty"`

	// Indicates whether data trace logging is enabled for methods in the
	// stage. API Gateway pushes these logs to Amazon CloudWatch Logs.
	DataTraceEnabled *BoolExpr `json:"DataTraceEnabled,omitempty"`

	// The HTTP method.
	HttpMethod *StringExpr `json:"HttpMethod,omitempty"`

	// The logging level for this method. For valid values, see the
	// loggingLevel property of the Stage resource in the Amazon API Gateway
	// API Reference.
	LoggingLevel *StringExpr `json:"LoggingLevel,omitempty"`

	// Indicates whether Amazon CloudWatch metrics are enabled for methods in
	// the stage.
	MetricsEnabled *BoolExpr `json:"MetricsEnabled,omitempty"`

	// The resource path for this method. Forward slashes (/) are encoded as
	// ~1 and the initial slash must include a forward slash. For example,
	// the path value /resource/subresource must be encoded as
	// /~1resource~1subresource. To specify the root path, use only a slash
	// (/).
	ResourcePath *StringExpr `json:"ResourcePath,omitempty"`

	// The number of burst requests per second that API Gateway permits
	// across all APIs, stages, and methods in your AWS account. For more
	// information, see Manage API Request Throttling in the API Gateway
	// Developer Guide.
	ThrottlingBurstLimit *IntegerExpr `json:"ThrottlingBurstLimit,omitempty"`

	// The number of steady-state requests per second that API Gateway
	// permits across all APIs, stages, and methods in your AWS account. For
	// more information, see Manage API Request Throttling in the API Gateway
	// Developer Guide.
	ThrottlingRateLimit *IntegerExpr `json:"ThrottlingRateLimit,omitempty"`
}

// APIGatewayStageMethodSettingList represents a list of APIGatewayStageMethodSetting
type APIGatewayStageMethodSettingList []APIGatewayStageMethodSetting

// UnmarshalJSON sets the object from the provided JSON representation
func (l *APIGatewayStageMethodSettingList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := APIGatewayStageMethodSetting{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = APIGatewayStageMethodSettingList{item}
		return nil
	}
	list := []APIGatewayStageMethodSetting{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = APIGatewayStageMethodSettingList(list)
		return nil
	}
	return err
}

// APIGatewayUsagePlanApiStage represents Amazon API Gateway UsagePlan ApiStage
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-apigateway-usageplan-apistage.html
type APIGatewayUsagePlanApiStage struct {
	// The ID of an API that is in the specified Stage property that you want
	// to associate with the usage plan.
	ApiId *StringExpr `json:"ApiId,omitempty"`

	// The name of an API Gateway stage to associate with the usage plan.
	Stage *StringExpr `json:"Stage,omitempty"`
}

// APIGatewayUsagePlanApiStageList represents a list of APIGatewayUsagePlanApiStage
type APIGatewayUsagePlanApiStageList []APIGatewayUsagePlanApiStage

// UnmarshalJSON sets the object from the provided JSON representation
func (l *APIGatewayUsagePlanApiStageList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := APIGatewayUsagePlanApiStage{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = APIGatewayUsagePlanApiStageList{item}
		return nil
	}
	list := []APIGatewayUsagePlanApiStage{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = APIGatewayUsagePlanApiStageList(list)
		return nil
	}
	return err
}

// APIGatewayUsagePlanQuotaSettings represents Amazon API Gateway UsagePlan QuotaSettings
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-apigateway-usageplan-quotasettings.html
type APIGatewayUsagePlanQuotaSettings struct {
	// The maximum number of requests that users can make within the
	// specified time period.
	Limit *IntegerExpr `json:"Limit,omitempty"`

	// For the initial time period, the number of requests to subtract from
	// the specified limit. When you first implement a usage plan, the plan
	// might start in the middle of the week or month. With this property,
	// you can decrease the limit for this initial time period.
	Offset *IntegerExpr `json:"Offset,omitempty"`

	// The time period for which the maximum limit of requests applies, such
	// as DAY or WEEK. For valid values, see the period property for the
	// UsagePlan resource in the Amazon API Gateway REST API Reference.
	Period *StringExpr `json:"Period,omitempty"`
}

// APIGatewayUsagePlanQuotaSettingsList represents a list of APIGatewayUsagePlanQuotaSettings
type APIGatewayUsagePlanQuotaSettingsList []APIGatewayUsagePlanQuotaSettings

// UnmarshalJSON sets the object from the provided JSON representation
func (l *APIGatewayUsagePlanQuotaSettingsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := APIGatewayUsagePlanQuotaSettings{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = APIGatewayUsagePlanQuotaSettingsList{item}
		return nil
	}
	list := []APIGatewayUsagePlanQuotaSettings{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = APIGatewayUsagePlanQuotaSettingsList(list)
		return nil
	}
	return err
}

// APIGatewayUsagePlanThrottleSettings represents Amazon API Gateway UsagePlan ThrottleSettings
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-apigateway-usageplan-throttlesettings.html
type APIGatewayUsagePlanThrottleSettings struct {
	// The maximum API request rate limit over a time ranging from one to a
	// few seconds. The maximum API request rate limit depends on whether the
	// underlying token bucket is at its full capacity. For more information
	// about request throttling, see Manage API Request Throttling in the API
	// Gateway Developer Guide.
	BurstLimit *IntegerExpr `json:"BurstLimit,omitempty"`

	// The API request steady-state rate limit (average requests per second
	// over an extended period of time). For more information about request
	// throttling, see Manage API Request Throttling in the API Gateway
	// Developer Guide.
	RateLimit *IntegerExpr `json:"RateLimit,omitempty"`
}

// APIGatewayUsagePlanThrottleSettingsList represents a list of APIGatewayUsagePlanThrottleSettings
type APIGatewayUsagePlanThrottleSettingsList []APIGatewayUsagePlanThrottleSettings

// UnmarshalJSON sets the object from the provided JSON representation
func (l *APIGatewayUsagePlanThrottleSettingsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := APIGatewayUsagePlanThrottleSettings{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = APIGatewayUsagePlanThrottleSettingsList{item}
		return nil
	}
	list := []APIGatewayUsagePlanThrottleSettings{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = APIGatewayUsagePlanThrottleSettingsList(list)
		return nil
	}
	return err
}

// ApplicationAutoScalingScalingPolicyStepScalingPolicyConfiguration represents Application Auto Scaling ScalingPolicy StepScalingPolicyConfiguration
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-applicationautoscaling-scalingpolicy-stepscalingpolicyconfiguration.html
type ApplicationAutoScalingScalingPolicyStepScalingPolicyConfiguration struct {
	// Specifies whether the ScalingAdjustment value in the StepAdjustment
	// property is an absolute number or a percentage of the current
	// capacity. For valid values, see the AdjustmentType content for the
	// StepScalingPolicyConfiguration data type in the Application Auto
	// Scaling API Reference.
	AdjustmentType *StringExpr `json:"AdjustmentType,omitempty"`

	// The amount of time, in seconds, after a scaling activity completes
	// before any further trigger-related scaling activities can start. For
	// more information, see the Cooldown content for the
	// StepScalingPolicyConfiguration data type in the Application Auto
	// Scaling API Reference.
	Cooldown *IntegerExpr `json:"Cooldown,omitempty"`

	// The aggregation type for the CloudWatch metrics. You can specify
	// Minimum, Maximum, or Average. By default, AWS CloudFormation specifies
	// Average. For more information, see Aggregation in the Amazon
	// CloudWatch User Guide.
	MetricAggregationType *StringExpr `json:"MetricAggregationType,omitempty"`

	// The minimum number of resources to adjust when a scaling activity is
	// triggered. If you specify PercentChangeInCapacity for the adjustment
	// type, the scaling policy scales the target by this amount.
	MinAdjustmentMagnitude *IntegerExpr `json:"MinAdjustmentMagnitude,omitempty"`

	// A set of adjustments that enable you to scale based on the size of the
	// alarm breach.
	StepAdjustments *ApplicationAutoScalingScalingPolicyStepScalingPolicyConfigurationStepAdjustmentList `json:"StepAdjustments,omitempty"`
}

// ApplicationAutoScalingScalingPolicyStepScalingPolicyConfigurationList represents a list of ApplicationAutoScalingScalingPolicyStepScalingPolicyConfiguration
type ApplicationAutoScalingScalingPolicyStepScalingPolicyConfigurationList []ApplicationAutoScalingScalingPolicyStepScalingPolicyConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ApplicationAutoScalingScalingPolicyStepScalingPolicyConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ApplicationAutoScalingScalingPolicyStepScalingPolicyConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ApplicationAutoScalingScalingPolicyStepScalingPolicyConfigurationList{item}
		return nil
	}
	list := []ApplicationAutoScalingScalingPolicyStepScalingPolicyConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ApplicationAutoScalingScalingPolicyStepScalingPolicyConfigurationList(list)
		return nil
	}
	return err
}

// ApplicationAutoScalingScalingPolicyStepScalingPolicyConfigurationStepAdjustment represents Application Auto Scaling ScalingPolicy StepScalingPolicyConfiguration StepAdjustment
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-applicationautoscaling-scalingpolicy-stepscalingpolicyconfiguration-stepadjustment.html
type ApplicationAutoScalingScalingPolicyStepScalingPolicyConfigurationStepAdjustment struct {
	// The lower bound of the breach size. The lower bound is the difference
	// between the breach threshold and the aggregated CloudWatch metric
	// value. If the metric value is within the lower and upper bounds,
	// Application Auto Scaling triggers this step adjustment.
	MetricIntervalLowerBound *IntegerExpr `json:"MetricIntervalLowerBound,omitempty"`

	// The upper bound of the breach size. The upper bound is the difference
	// between the breach threshold and the CloudWatch metric value. If the
	// metric value is within the lower and upper bounds, Application Auto
	// Scaling triggers this step adjustment.
	MetricIntervalUpperBound *IntegerExpr `json:"MetricIntervalUpperBound,omitempty"`

	// The amount by which to scale. The adjustment is based on the value
	// that you specified in the AdjustmentType property (either an absolute
	// number or a percentage). A positive value adds to the current capacity
	// and a negative number subtracts from the current capacity.
	ScalingAdjustment *IntegerExpr `json:"ScalingAdjustment,omitempty"`
}

// ApplicationAutoScalingScalingPolicyStepScalingPolicyConfigurationStepAdjustmentList represents a list of ApplicationAutoScalingScalingPolicyStepScalingPolicyConfigurationStepAdjustment
type ApplicationAutoScalingScalingPolicyStepScalingPolicyConfigurationStepAdjustmentList []ApplicationAutoScalingScalingPolicyStepScalingPolicyConfigurationStepAdjustment

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ApplicationAutoScalingScalingPolicyStepScalingPolicyConfigurationStepAdjustmentList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ApplicationAutoScalingScalingPolicyStepScalingPolicyConfigurationStepAdjustment{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ApplicationAutoScalingScalingPolicyStepScalingPolicyConfigurationStepAdjustmentList{item}
		return nil
	}
	list := []ApplicationAutoScalingScalingPolicyStepScalingPolicyConfigurationStepAdjustment{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ApplicationAutoScalingScalingPolicyStepScalingPolicyConfigurationStepAdjustmentList(list)
		return nil
	}
	return err
}

// AutoScalingBlockDeviceMapping represents AWS CloudFormation AutoScaling Block Device Mapping Property Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-as-launchconfig-blockdev-mapping.html
type AutoScalingBlockDeviceMapping struct {
	// The name of the device within Amazon EC2.
	DeviceName *StringExpr `json:"DeviceName,omitempty"`

	// The Amazon Elastic Block Store volume information.
	Ebs *AutoScalingEBSBlockDevice `json:"Ebs,omitempty"`

	// Suppresses the device mapping. If NoDevice is set to true for the root
	// device, the instance might fail the Amazon EC2 health check. Auto
	// Scaling launches a replacement instance if the instance fails the
	// health check.
	NoDevice *BoolExpr `json:"NoDevice,omitempty"`

	// The name of the virtual device. The name must be in the form
	// ephemeralX where X is a number starting from zero (0), for example,
	// ephemeral0.
	VirtualName *StringExpr `json:"VirtualName,omitempty"`
}

// AutoScalingBlockDeviceMappingList represents a list of AutoScalingBlockDeviceMapping
type AutoScalingBlockDeviceMappingList []AutoScalingBlockDeviceMapping

// UnmarshalJSON sets the object from the provided JSON representation
func (l *AutoScalingBlockDeviceMappingList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := AutoScalingBlockDeviceMapping{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = AutoScalingBlockDeviceMappingList{item}
		return nil
	}
	list := []AutoScalingBlockDeviceMapping{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = AutoScalingBlockDeviceMappingList(list)
		return nil
	}
	return err
}

// AutoScalingEBSBlockDevice represents AWS CloudFormation AutoScaling EBS Block Device Property Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-as-launchconfig-blockdev-template.html
type AutoScalingEBSBlockDevice struct {
	// Indicates whether to delete the volume when the instance is
	// terminated. By default, Auto Scaling uses true.
	DeleteOnTermination *BoolExpr `json:"DeleteOnTermination,omitempty"`

	// Indicates whether the volume is encrypted. Encrypted EBS volumes must
	// be attached to instances that support Amazon EBS encryption. Volumes
	// that you create from encrypted snapshots are automatically encrypted.
	// You cannot create an encrypted volume from an unencrypted snapshot or
	// an unencrypted volume from an encrypted snapshot.
	Encrypted *BoolExpr `json:"Encrypted,omitempty"`

	// The number of I/O operations per second (IOPS) that the volume
	// supports. The maximum ratio of IOPS to volume size is 30.
	Iops *IntegerExpr `json:"Iops,omitempty"`

	// The snapshot ID of the volume to use.
	SnapshotId *StringExpr `json:"SnapshotId,omitempty"`

	// The volume size, in Gibibytes (GiB). This can be a number from 1 –
	// 1024. If the volume type is EBS optimized, the minimum value is 10.
	// For more information about specifying the volume type, see
	// EbsOptimized in AWS::AutoScaling::LaunchConfiguration.
	VolumeSize *IntegerExpr `json:"VolumeSize,omitempty"`

	// The volume type. By default, Auto Scaling uses the standard volume
	// type. For more information, see Ebs in the Auto Scaling API Reference.
	VolumeType *StringExpr `json:"VolumeType,omitempty"`
}

// AutoScalingEBSBlockDeviceList represents a list of AutoScalingEBSBlockDevice
type AutoScalingEBSBlockDeviceList []AutoScalingEBSBlockDevice

// UnmarshalJSON sets the object from the provided JSON representation
func (l *AutoScalingEBSBlockDeviceList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := AutoScalingEBSBlockDevice{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = AutoScalingEBSBlockDeviceList{item}
		return nil
	}
	list := []AutoScalingEBSBlockDevice{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = AutoScalingEBSBlockDeviceList(list)
		return nil
	}
	return err
}

// AutoScalingMetricsCollection represents Auto Scaling MetricsCollection
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-as-metricscollection.html
type AutoScalingMetricsCollection struct {
	// The frequency at which Auto Scaling sends aggregated data to
	// CloudWatch. For example, you can specify 1Minute to send aggregated
	// data to CloudWatch every minute.
	Granularity *StringExpr `json:"Granularity,omitempty"`

	// The list of metrics to collect. If you don't specify any metrics, all
	// metrics are enabled.
	Metrics *StringListExpr `json:"Metrics,omitempty"`
}

// AutoScalingMetricsCollectionList represents a list of AutoScalingMetricsCollection
type AutoScalingMetricsCollectionList []AutoScalingMetricsCollection

// UnmarshalJSON sets the object from the provided JSON representation
func (l *AutoScalingMetricsCollectionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := AutoScalingMetricsCollection{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = AutoScalingMetricsCollectionList{item}
		return nil
	}
	list := []AutoScalingMetricsCollection{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = AutoScalingMetricsCollectionList(list)
		return nil
	}
	return err
}

// AutoScalingNotificationConfigurations represents Auto Scaling NotificationConfigurations
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-as-notificationconfigurations.html
type AutoScalingNotificationConfigurations struct {
	// A list of event types that trigger a notification. Event types can
	// include any of the following types: autoscaling:EC2_INSTANCE_LAUNCH,
	// autoscaling:EC2_INSTANCE_LAUNCH_ERROR,
	// autoscaling:EC2_INSTANCE_TERMINATE,
	// autoscaling:EC2_INSTANCE_TERMINATE_ERROR, and
	// autoscaling:TEST_NOTIFICATION. For more information about event types,
	// see DescribeAutoScalingNotificationTypes in the Auto Scaling API
	// Reference.
	NotificationTypes *StringListExpr `json:"NotificationTypes,omitempty"`

	// The Amazon Resource Name (ARN) of the Amazon Simple Notification
	// Service (SNS) topic.
	TopicARN *StringExpr `json:"TopicARN,omitempty"`
}

// AutoScalingNotificationConfigurationsList represents a list of AutoScalingNotificationConfigurations
type AutoScalingNotificationConfigurationsList []AutoScalingNotificationConfigurations

// UnmarshalJSON sets the object from the provided JSON representation
func (l *AutoScalingNotificationConfigurationsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := AutoScalingNotificationConfigurations{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = AutoScalingNotificationConfigurationsList{item}
		return nil
	}
	list := []AutoScalingNotificationConfigurations{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = AutoScalingNotificationConfigurationsList(list)
		return nil
	}
	return err
}

// AutoScalingScalingPolicyStepAdjustments represents Auto Scaling ScalingPolicy StepAdjustments
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-autoscaling-scalingpolicy-stepadjustments.html
type AutoScalingScalingPolicyStepAdjustments struct {
	// The lower bound of the breach size. The lower bound is the difference
	// between the breach threshold and the aggregated CloudWatch metric
	// value. If the metric value is within the lower and upper bounds, Auto
	// Scaling triggers this step adjustment.
	MetricIntervalLowerBound *IntegerExpr `json:"MetricIntervalLowerBound,omitempty"`

	// The upper bound of the breach size. The upper bound is the difference
	// between the breach threshold and the CloudWatch metric value. If the
	// metric value is within the lower and upper bounds, Auto Scaling
	// triggers this step adjustment.
	MetricIntervalUpperBound *IntegerExpr `json:"MetricIntervalUpperBound,omitempty"`

	// The amount by which to scale. The adjustment is based on the value
	// that you specified in the AdjustmentType property (either an absolute
	// number or a percentage). A positive value adds to the current capacity
	// and a negative number subtracts from the current capacity.
	ScalingAdjustment *IntegerExpr `json:"ScalingAdjustment,omitempty"`
}

// AutoScalingScalingPolicyStepAdjustmentsList represents a list of AutoScalingScalingPolicyStepAdjustments
type AutoScalingScalingPolicyStepAdjustmentsList []AutoScalingScalingPolicyStepAdjustments

// UnmarshalJSON sets the object from the provided JSON representation
func (l *AutoScalingScalingPolicyStepAdjustmentsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := AutoScalingScalingPolicyStepAdjustments{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = AutoScalingScalingPolicyStepAdjustmentsList{item}
		return nil
	}
	list := []AutoScalingScalingPolicyStepAdjustments{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = AutoScalingScalingPolicyStepAdjustmentsList(list)
		return nil
	}
	return err
}

// AutoScalingTags represents Auto Scaling Tags Property Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-as-tags.html
type AutoScalingTags struct {
	// The key name of the tag.
	Key *StringExpr `json:"Key,omitempty"`

	// The value for the tag.
	Value *StringExpr `json:"Value,omitempty"`

	// Set to true if you want AWS CloudFormation to copy the tag to EC2
	// instances that are launched as part of the auto scaling group. Set to
	// false if you want the tag attached only to the auto scaling group and
	// not copied to any instances launched as part of the auto scaling
	// group.
	PropagateAtLaunch *BoolExpr `json:"PropagateAtLaunch,omitempty"`
}

// AutoScalingTagsList represents a list of AutoScalingTags
type AutoScalingTagsList []AutoScalingTags

// UnmarshalJSON sets the object from the provided JSON representation
func (l *AutoScalingTagsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := AutoScalingTags{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = AutoScalingTagsList{item}
		return nil
	}
	list := []AutoScalingTags{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = AutoScalingTagsList(list)
		return nil
	}
	return err
}

// CertificateManagerCertificateDomainValidationOption represents AWS Certificate Manager Certificate DomainValidationOption
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-certificatemanager-certificate-domainvalidationoption.html
type CertificateManagerCertificateDomainValidationOption struct {
	// Fully Qualified Domain Name (FQDN) of the Certificate that you are
	// requesting.
	DomainName *StringExpr `json:"DomainName,omitempty"`

	// The domain that domain name registrars use to send validation emails.
	// Registrars use this value as the email address suffix when sending
	// emails to verify your identity. This value must be the same as the
	// domain name or a superdomain of the domain name. For more information,
	// see the ValidationDomain content for the DomainValidationOption data
	// type in the AWS Certificate Manager API Reference.
	ValidationDomain *StringExpr `json:"ValidationDomain,omitempty"`
}

// CertificateManagerCertificateDomainValidationOptionList represents a list of CertificateManagerCertificateDomainValidationOption
type CertificateManagerCertificateDomainValidationOptionList []CertificateManagerCertificateDomainValidationOption

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CertificateManagerCertificateDomainValidationOptionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CertificateManagerCertificateDomainValidationOption{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CertificateManagerCertificateDomainValidationOptionList{item}
		return nil
	}
	list := []CertificateManagerCertificateDomainValidationOption{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CertificateManagerCertificateDomainValidationOptionList(list)
		return nil
	}
	return err
}

// CloudFormationStackParameters represents CloudFormation Stack Parameters Property Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-stack-parameters.html
type CloudFormationStackParameters struct {
}

// CloudFormationStackParametersList represents a list of CloudFormationStackParameters
type CloudFormationStackParametersList []CloudFormationStackParameters

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CloudFormationStackParametersList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CloudFormationStackParameters{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CloudFormationStackParametersList{item}
		return nil
	}
	list := []CloudFormationStackParameters{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CloudFormationStackParametersList(list)
		return nil
	}
	return err
}

// InterfaceLabel represents AWS CloudFormation Interface Label
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cloudformation-interface-label.html
type InterfaceLabel struct {
	// The default label that the AWS CloudFormation console uses to name a
	// parameter group or parameter.
	Default *StringExpr `json:"default,omitempty"`
}

// InterfaceLabelList represents a list of InterfaceLabel
type InterfaceLabelList []InterfaceLabel

// UnmarshalJSON sets the object from the provided JSON representation
func (l *InterfaceLabelList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := InterfaceLabel{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = InterfaceLabelList{item}
		return nil
	}
	list := []InterfaceLabel{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = InterfaceLabelList(list)
		return nil
	}
	return err
}

// InterfaceParameterGroup represents AWS CloudFormation Interface ParameterGroup
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cloudformation-interface-parametergroup.html
type InterfaceParameterGroup struct {
	// A name for the parameter group.
	Label *InterfaceLabel `json:"Label,omitempty"`

	// A list of case-sensitive parameter logical IDs to include in the
	// group. Parameters must already be defined in the Parameters section of
	// the template. A parameter can be included in only one parameter group.
	Parameters *StringListExpr `json:"Parameters,omitempty"`
}

// InterfaceParameterGroupList represents a list of InterfaceParameterGroup
type InterfaceParameterGroupList []InterfaceParameterGroup

// UnmarshalJSON sets the object from the provided JSON representation
func (l *InterfaceParameterGroupList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := InterfaceParameterGroup{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = InterfaceParameterGroupList{item}
		return nil
	}
	list := []InterfaceParameterGroup{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = InterfaceParameterGroupList(list)
		return nil
	}
	return err
}

// InterfaceParameterLabel represents AWS CloudFormation Interface ParameterLabel
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cloudformation-interface-parameterlabel.html
type InterfaceParameterLabel struct {
	// A label for a parameter. The label defines a friendly name or
	// description that the AWS CloudFormation console shows on the Specify
	// Parameters page when a stack is created or updated. The
	// ParameterLogicalID key must be the case-sensitive logical ID of a
	// valid parameter that has been declared in the Parameters section of
	// the template.
	ParameterLogicalID *InterfaceLabel `json:"ParameterLogicalID,omitempty"`
}

// InterfaceParameterLabelList represents a list of InterfaceParameterLabel
type InterfaceParameterLabelList []InterfaceParameterLabel

// UnmarshalJSON sets the object from the provided JSON representation
func (l *InterfaceParameterLabelList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := InterfaceParameterLabel{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = InterfaceParameterLabelList{item}
		return nil
	}
	list := []InterfaceParameterLabel{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = InterfaceParameterLabelList(list)
		return nil
	}
	return err
}

// CloudFrontDistributionConfig represents CloudFront DistributionConfig
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cloudfront-distributionconfig.html
type CloudFrontDistributionConfig struct {
	// CNAMEs (alternate domain names), if any, for the distribution.
	Aliases *StringListExpr `json:"Aliases,omitempty"`

	// A list of CacheBehavior types for the distribution.
	CacheBehaviors *CloudFrontDistributionConfigCacheBehaviorList `json:"CacheBehaviors,omitempty"`

	// Any comments that you want to include about the distribution.
	Comment *StringExpr `json:"Comment,omitempty"`

	// Whether CloudFront replaces HTTP status codes in the 4xx and 5xx range
	// with custom error messages before returning the response to the
	// viewer.
	CustomErrorResponses *CloudFrontDistributionConfigCustomErrorResponseList `json:"CustomErrorResponses,omitempty"`

	// The default cache behavior that is triggered if you do not specify the
	// CacheBehavior property or if files don't match any of the values of
	// PathPattern in the CacheBehavior property.
	DefaultCacheBehavior *CloudFrontDefaultCacheBehavior `json:"DefaultCacheBehavior,omitempty"`

	// The object (such as index.html) that you want CloudFront to request
	// from your origin when the root URL for your distribution (such as
	// http://example.com/) is requested.
	DefaultRootObject *StringExpr `json:"DefaultRootObject,omitempty"`

	// Controls whether the distribution is enabled to accept end user
	// requests for content.
	Enabled *BoolExpr `json:"Enabled,omitempty"`

	// The latest HTTP version that viewers can use to communicate with
	// CloudFront. Viewers that don't support the latest version
	// automatically use an earlier HTTP version. By default, AWS
	// CloudFormation specifies http1.1.
	HttpVersion *StringExpr `json:"HttpVersion,omitempty"`

	// Controls whether access logs are written for the distribution. To turn
	// on access logs, specify this property.
	Logging *CloudFrontLogging `json:"Logging,omitempty"`

	// A list of origins for this CloudFront distribution. For each origin,
	// you can specify whether it is an Amazon S3 or custom origin.
	Origins *CloudFrontDistributionConfigOriginList `json:"Origins,omitempty"`

	// The price class that corresponds with the maximum price that you want
	// to pay for the CloudFront service. For more information, see Choosing
	// the Price Class in the Amazon CloudFront Developer Guide.
	PriceClass *StringExpr `json:"PriceClass,omitempty"`

	// Specifies restrictions on who or how viewers can access your content.
	Restrictions *CloudFrontDistributionConfigurationRestrictions `json:"Restrictions,omitempty"`

	// The certificate to use when viewers use HTTPS to request objects.
	ViewerCertificate *CloudFrontDistributionConfigurationViewerCertificate `json:"ViewerCertificate,omitempty"`

	// The AWS WAF web ACL to associate with this distribution. AWS WAF is a
	// web application firewall that enables you to monitor the HTTP and
	// HTTPS requests that are forwarded to CloudFront and to control who can
	// access your content. CloudFront permits or forbids requests based on
	// conditions that you specify, such as the IP addresses from which
	// requests originate or the values of query strings.
	WebACLId *StringExpr `json:"WebACLId,omitempty"`
}

// CloudFrontDistributionConfigList represents a list of CloudFrontDistributionConfig
type CloudFrontDistributionConfigList []CloudFrontDistributionConfig

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CloudFrontDistributionConfigList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CloudFrontDistributionConfig{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CloudFrontDistributionConfigList{item}
		return nil
	}
	list := []CloudFrontDistributionConfig{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CloudFrontDistributionConfigList(list)
		return nil
	}
	return err
}

// CloudFrontDistributionConfigCacheBehavior represents CloudFront DistributionConfig CacheBehavior
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cloudfront-cachebehavior.html
type CloudFrontDistributionConfigCacheBehavior struct {
	// HTTP methods that CloudFront processes and forwards to your Amazon S3
	// bucket or your custom origin. You can specify ["HEAD", "GET"], ["GET",
	// "HEAD", "OPTIONS"], or ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH",
	// "POST", "PUT"]. If you don't specify a value, AWS CloudFormation
	// specifies ["HEAD", "GET"].
	AllowedMethods *StringListExpr `json:"AllowedMethods,omitempty"`

	// HTTP methods for which CloudFront caches responses. You can specify
	// ["HEAD", "GET"] or ["GET", "HEAD", "OPTIONS"]. If you don't specify a
	// value, AWS CloudFormation specifies ["HEAD", "GET"].
	CachedMethods *StringListExpr `json:"CachedMethods,omitempty"`

	// Indicates whether CloudFront automatically compresses certain files
	// for this cache behavior. For more information, see Serving Compressed
	// Files in the Amazon CloudFront Developer Guide.
	Compress *BoolExpr `json:"Compress,omitempty"`

	// The default time in seconds that objects stay in CloudFront caches
	// before CloudFront forwards another request to your custom origin to
	// determine whether the object has been updated. This value applies only
	// when your custom origin does not add HTTP headers, such as
	// Cache-Control max-age, Cache-Control s-maxage, and Expires to objects.
	DefaultTTL *IntegerExpr `json:"DefaultTTL,omitempty"`

	// Specifies how CloudFront handles query strings or cookies.
	ForwardedValues *CloudFrontForwardedValues `json:"ForwardedValues,omitempty"`

	// The maximum time in seconds that objects stay in CloudFront caches
	// before CloudFront forwards another request to your custom origin to
	// determine whether the object has been updated. This value applies only
	// when your custom origin does not add HTTP headers, such as
	// Cache-Control max-age, Cache-Control s-maxage, and Expires to objects.
	MaxTTL *IntegerExpr `json:"MaxTTL,omitempty"`

	// The minimum amount of time that you want objects to stay in the cache
	// before CloudFront queries your origin to see whether the object has
	// been updated.
	MinTTL *IntegerExpr `json:"MinTTL,omitempty"`

	// The pattern to which this cache behavior applies. For example, you can
	// specify images/*.jpg.
	PathPattern *StringExpr `json:"PathPattern,omitempty"`

	// Indicates whether to use the origin that is associated with this cache
	// behavior to distribute media files in the Microsoft Smooth Streaming
	// format. If you specify true, you can still use this cache behavior to
	// distribute other content if the content matches the PathPattern value.
	SmoothStreaming *BoolExpr `json:"SmoothStreaming,omitempty"`

	// The ID value of the origin to which you want CloudFront to route
	// requests when a request matches the value of the PathPattern property.
	TargetOriginId *StringExpr `json:"TargetOriginId,omitempty"`

	// A list of AWS accounts that can create signed URLs in order to access
	// private content.
	TrustedSigners *StringListExpr `json:"TrustedSigners,omitempty"`

	// The protocol that users can use to access the files in the origin that
	// you specified in the TargetOriginId property when a request matches
	// the value of the PathPattern property. For more information about the
	// valid values, see the ViewerProtocolPolicy content for the
	// CacheBehavior data type in the Amazon CloudFront API Reference.
	ViewerProtocolPolicy *StringExpr `json:"ViewerProtocolPolicy,omitempty"`
}

// CloudFrontDistributionConfigCacheBehaviorList represents a list of CloudFrontDistributionConfigCacheBehavior
type CloudFrontDistributionConfigCacheBehaviorList []CloudFrontDistributionConfigCacheBehavior

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CloudFrontDistributionConfigCacheBehaviorList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CloudFrontDistributionConfigCacheBehavior{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CloudFrontDistributionConfigCacheBehaviorList{item}
		return nil
	}
	list := []CloudFrontDistributionConfigCacheBehavior{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CloudFrontDistributionConfigCacheBehaviorList(list)
		return nil
	}
	return err
}

// CloudFrontDistributionConfigCustomErrorResponse represents CloudFront DistributionConfig CustomErrorResponse
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cloudfront-distributionconfig-customerrorresponse.html
type CloudFrontDistributionConfigCustomErrorResponse struct {
	// The minimum amount of time, in seconds, that Amazon CloudFront caches
	// the HTTP status code that you specified in the ErrorCode property. The
	// default value is 300.
	ErrorCachingMinTTL *IntegerExpr `json:"ErrorCachingMinTTL,omitempty"`

	// An HTTP status code for which you want to specify a custom error page.
	// You can specify 400, 403, 404, 405, 414, 500, 501, 502, 503, or 504.
	ErrorCode *IntegerExpr `json:"ErrorCode,omitempty"`

	// The HTTP status code that CloudFront returns to viewer along with the
	// custom error page. You can specify 200, 400, 403, 404, 405, 414, 500,
	// 501, 502, 503, or 504.
	ResponseCode *IntegerExpr `json:"ResponseCode,omitempty"`

	// The path to the custom error page that CloudFront returns to a viewer
	// when your origin returns the HTTP status code that you specified in
	// the ErrorCode property. For example, you can specify
	// /404-errors/403-forbidden.html.
	ResponsePagePath *StringExpr `json:"ResponsePagePath,omitempty"`
}

// CloudFrontDistributionConfigCustomErrorResponseList represents a list of CloudFrontDistributionConfigCustomErrorResponse
type CloudFrontDistributionConfigCustomErrorResponseList []CloudFrontDistributionConfigCustomErrorResponse

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CloudFrontDistributionConfigCustomErrorResponseList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CloudFrontDistributionConfigCustomErrorResponse{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CloudFrontDistributionConfigCustomErrorResponseList{item}
		return nil
	}
	list := []CloudFrontDistributionConfigCustomErrorResponse{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CloudFrontDistributionConfigCustomErrorResponseList(list)
		return nil
	}
	return err
}

// CloudFrontDefaultCacheBehavior represents CloudFront DefaultCacheBehavior
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cloudfront-defaultcachebehavior.html
type CloudFrontDefaultCacheBehavior struct {
	// HTTP methods that CloudFront processes and forwards to your Amazon S3
	// bucket or your custom origin. In AWS CloudFormation templates, you can
	// specify ["HEAD", "GET"], ["GET", "HEAD", "OPTIONS"], or ["DELETE",
	// "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]. If you don't
	// specify a value, AWS CloudFormation specifies ["HEAD", "GET"].
	AllowedMethods *StringListExpr `json:"AllowedMethods,omitempty"`

	// HTTP methods for which CloudFront caches responses. In AWS
	// CloudFormation templates, you can specify ["HEAD", "GET"] or ["GET",
	// "HEAD", "OPTIONS"]. If you don't specify a value, AWS CloudFormation
	// specifies ["HEAD", "GET"].
	CachedMethods *StringListExpr `json:"CachedMethods,omitempty"`

	// Indicates whether CloudFront automatically compresses certain files
	// for this cache behavior. For more information, see Serving Compressed
	// Files in the Amazon CloudFront Developer Guide.
	Compress *BoolExpr `json:"Compress,omitempty"`

	// The default time in seconds that objects stay in CloudFront caches
	// before CloudFront forwards another request to your custom origin to
	// determine whether the object has been updated. This value applies only
	// when your custom origin does not add HTTP headers, such as
	// Cache-Control max-age, Cache-Control s-maxage, and Expires to objects.
	DefaultTTL *IntegerExpr `json:"DefaultTTL,omitempty"`

	// Specifies how CloudFront handles query strings or cookies.
	ForwardedValues *CloudFrontForwardedValues `json:"ForwardedValues,omitempty"`

	// The maximum time in seconds that objects stay in CloudFront caches
	// before CloudFront forwards another request to your custom origin to
	// determine whether the object has been updated. This value applies only
	// when your custom origin does not add HTTP headers, such as
	// Cache-Control max-age, Cache-Control s-maxage, and Expires to objects.
	MaxTTL *IntegerExpr `json:"MaxTTL,omitempty"`

	// The minimum amount of time that you want objects to stay in the cache
	// before CloudFront queries your origin to see whether the object has
	// been updated.
	MinTTL *StringExpr `json:"MinTTL,omitempty"`

	// Indicates whether to use the origin that is associated with this cache
	// behavior to distribute media files in the Microsoft Smooth Streaming
	// format.
	SmoothStreaming *BoolExpr `json:"SmoothStreaming,omitempty"`

	// The value of ID for the origin that CloudFront routes requests to when
	// the default cache behavior is applied to a request.
	TargetOriginId *StringExpr `json:"TargetOriginId,omitempty"`

	// A list of AWS accounts that can create signed URLs in order to access
	// private content.
	TrustedSigners *StringListExpr `json:"TrustedSigners,omitempty"`

	// The protocol that users can use to access the files in the origin that
	// you specified in the TargetOriginId property when the default cache
	// behavior is applied to a request. For more information about the valid
	// values, see the ViewerProtocolPolicy content for the
	// DefaultCacheBehavior data type in the Amazon CloudFront API Reference.
	ViewerProtocolPolicy *StringExpr `json:"ViewerProtocolPolicy,omitempty"`
}

// CloudFrontDefaultCacheBehaviorList represents a list of CloudFrontDefaultCacheBehavior
type CloudFrontDefaultCacheBehaviorList []CloudFrontDefaultCacheBehavior

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CloudFrontDefaultCacheBehaviorList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CloudFrontDefaultCacheBehavior{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CloudFrontDefaultCacheBehaviorList{item}
		return nil
	}
	list := []CloudFrontDefaultCacheBehavior{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CloudFrontDefaultCacheBehaviorList(list)
		return nil
	}
	return err
}

// CloudFrontLogging represents CloudFront Logging
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cloudfront-logging.html
type CloudFrontLogging struct {
	// The Amazon S3 bucket address where access logs are stored, for
	// example, mybucket.s3.amazonaws.com.
	Bucket *StringExpr `json:"Bucket,omitempty"`

	// Indicates whether CloudFront includes cookies in access logs.
	IncludeCookies *BoolExpr `json:"IncludeCookies,omitempty"`

	// A prefix for the access log file names for this distribution.
	Prefix *StringExpr `json:"Prefix,omitempty"`
}

// CloudFrontLoggingList represents a list of CloudFrontLogging
type CloudFrontLoggingList []CloudFrontLogging

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CloudFrontLoggingList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CloudFrontLogging{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CloudFrontLoggingList{item}
		return nil
	}
	list := []CloudFrontLogging{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CloudFrontLoggingList(list)
		return nil
	}
	return err
}

// CloudFrontDistributionConfigOrigin represents CloudFront DistributionConfig Origin
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cloudfront-origin.html
type CloudFrontDistributionConfigOrigin struct {
	// Origin information to specify a custom origin.
	CustomOriginConfig *CloudFrontDistributionConfigOriginCustomOrigin `json:"CustomOriginConfig,omitempty"`

	// The DNS name of the Amazon Simple Storage Service (S3) bucket or the
	// HTTP server from which you want CloudFront to get objects for this
	// origin.
	DomainName *StringExpr `json:"DomainName,omitempty"`

	// An identifier for the origin. The value of Id must be unique within
	// the distribution.
	Id *StringExpr `json:"Id,omitempty"`

	// Custom headers that CloudFront includes when it forwards a request to
	// your origin.
	OriginCustomHeaders *CloudFrontDistributionConfigOriginOriginCustomHeaderList `json:"OriginCustomHeaders,omitempty"`

	// The path that CloudFront uses to request content from an S3 bucket or
	// custom origin. The combination of the DomainName and OriginPath
	// properties must resolve to a valid path. The value must start with a
	// slash mark (/) and cannot end with a slash mark.
	OriginPath *StringExpr `json:"OriginPath,omitempty"`

	// Origin information to specify an S3 origin.
	S3OriginConfig *CloudFrontDistributionConfigOriginS3Origin `json:"S3OriginConfig,omitempty"`
}

// CloudFrontDistributionConfigOriginList represents a list of CloudFrontDistributionConfigOrigin
type CloudFrontDistributionConfigOriginList []CloudFrontDistributionConfigOrigin

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CloudFrontDistributionConfigOriginList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CloudFrontDistributionConfigOrigin{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CloudFrontDistributionConfigOriginList{item}
		return nil
	}
	list := []CloudFrontDistributionConfigOrigin{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CloudFrontDistributionConfigOriginList(list)
		return nil
	}
	return err
}

// CloudFrontDistributionConfigOriginCustomOrigin represents CloudFront DistributionConfig Origin CustomOrigin
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cloudfront-customorigin.html
type CloudFrontDistributionConfigOriginCustomOrigin struct {
	// The HTTP port the custom origin listens on.
	HTTPPort *StringExpr `json:"HTTPPort,omitempty"`

	// The HTTPS port the custom origin listens on.
	HTTPSPort *StringExpr `json:"HTTPSPort,omitempty"`

	// The origin protocol policy to apply to your origin.
	OriginProtocolPolicy *StringExpr `json:"OriginProtocolPolicy,omitempty"`

	// The SSL protocols that CloudFront can use when establishing an HTTPS
	// connection with your origin. By default, AWS CloudFormation specifies
	// the TLSv1 and SSLv3 protocols.
	OriginSSLProtocols *StringListExpr `json:"OriginSSLProtocols,omitempty"`
}

// CloudFrontDistributionConfigOriginCustomOriginList represents a list of CloudFrontDistributionConfigOriginCustomOrigin
type CloudFrontDistributionConfigOriginCustomOriginList []CloudFrontDistributionConfigOriginCustomOrigin

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CloudFrontDistributionConfigOriginCustomOriginList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CloudFrontDistributionConfigOriginCustomOrigin{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CloudFrontDistributionConfigOriginCustomOriginList{item}
		return nil
	}
	list := []CloudFrontDistributionConfigOriginCustomOrigin{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CloudFrontDistributionConfigOriginCustomOriginList(list)
		return nil
	}
	return err
}

// CloudFrontDistributionConfigOriginOriginCustomHeader represents CloudFront DistributionConfig Origin OriginCustomHeader
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cloudfront-origin-origincustomheader.html
type CloudFrontDistributionConfigOriginOriginCustomHeader struct {
	// The name of a header that CloudFront forwards to your origin. For more
	// information, see Forwarding Custom Headers to Your Origin (Web
	// Distributions Only) in the Amazon CloudFront Developer Guide.
	HeaderName *StringExpr `json:"HeaderName,omitempty"`

	// The value for the header that you specified in the HeaderName
	// property.
	HeaderValue *StringExpr `json:"HeaderValue,omitempty"`
}

// CloudFrontDistributionConfigOriginOriginCustomHeaderList represents a list of CloudFrontDistributionConfigOriginOriginCustomHeader
type CloudFrontDistributionConfigOriginOriginCustomHeaderList []CloudFrontDistributionConfigOriginOriginCustomHeader

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CloudFrontDistributionConfigOriginOriginCustomHeaderList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CloudFrontDistributionConfigOriginOriginCustomHeader{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CloudFrontDistributionConfigOriginOriginCustomHeaderList{item}
		return nil
	}
	list := []CloudFrontDistributionConfigOriginOriginCustomHeader{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CloudFrontDistributionConfigOriginOriginCustomHeaderList(list)
		return nil
	}
	return err
}

// CloudFrontDistributionConfigOriginS3Origin represents CloudFront DistributionConfig Origin S3Origin
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cloudfront-s3origin.html
type CloudFrontDistributionConfigOriginS3Origin struct {
	// The CloudFront origin access identity to associate with the origin.
	// This is used to configure the origin so that end users can access
	// objects in an Amazon S3 bucket through CloudFront only.
	OriginAccessIdentity *StringExpr `json:"OriginAccessIdentity,omitempty"`
}

// CloudFrontDistributionConfigOriginS3OriginList represents a list of CloudFrontDistributionConfigOriginS3Origin
type CloudFrontDistributionConfigOriginS3OriginList []CloudFrontDistributionConfigOriginS3Origin

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CloudFrontDistributionConfigOriginS3OriginList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CloudFrontDistributionConfigOriginS3Origin{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CloudFrontDistributionConfigOriginS3OriginList{item}
		return nil
	}
	list := []CloudFrontDistributionConfigOriginS3Origin{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CloudFrontDistributionConfigOriginS3OriginList(list)
		return nil
	}
	return err
}

// CloudFrontDistributionConfigurationRestrictions represents CloudFront DistributionConfiguration Restrictions
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cloudfront-distributionconfig-restrictions.html
type CloudFrontDistributionConfigurationRestrictions struct {
	// The countries in which viewers are able to access your content.
	GeoRestriction *CloudFrontDistributionConfigRestrictionsGeoRestriction `json:"GeoRestriction,omitempty"`
}

// CloudFrontDistributionConfigurationRestrictionsList represents a list of CloudFrontDistributionConfigurationRestrictions
type CloudFrontDistributionConfigurationRestrictionsList []CloudFrontDistributionConfigurationRestrictions

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CloudFrontDistributionConfigurationRestrictionsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CloudFrontDistributionConfigurationRestrictions{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CloudFrontDistributionConfigurationRestrictionsList{item}
		return nil
	}
	list := []CloudFrontDistributionConfigurationRestrictions{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CloudFrontDistributionConfigurationRestrictionsList(list)
		return nil
	}
	return err
}

// CloudFrontDistributionConfigRestrictionsGeoRestriction represents CloudFront DistributionConfig Restrictions GeoRestriction
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cloudfront-distributionconfig-restrictions-georestriction.html
type CloudFrontDistributionConfigRestrictionsGeoRestriction struct {
	// The two-letter, uppercase country code for a country that you want to
	// include in your blacklist or whitelist.
	Locations *StringListExpr `json:"Locations,omitempty"`

	// The method to restrict distribution of your content:
	RestrictionType *StringExpr `json:"RestrictionType,omitempty"`

	// Prevents viewers in the countries that you specified from accessing
	// your content.
	Blacklist interface{} `json:"blacklist,omitempty"`

	// Allows viewers in the countries that you specified to access your
	// content.
	Whitelist interface{} `json:"whitelist,omitempty"`

	// No distribution restrictions by country.
	None interface{} `json:"none,omitempty"`
}

// CloudFrontDistributionConfigRestrictionsGeoRestrictionList represents a list of CloudFrontDistributionConfigRestrictionsGeoRestriction
type CloudFrontDistributionConfigRestrictionsGeoRestrictionList []CloudFrontDistributionConfigRestrictionsGeoRestriction

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CloudFrontDistributionConfigRestrictionsGeoRestrictionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CloudFrontDistributionConfigRestrictionsGeoRestriction{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CloudFrontDistributionConfigRestrictionsGeoRestrictionList{item}
		return nil
	}
	list := []CloudFrontDistributionConfigRestrictionsGeoRestriction{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CloudFrontDistributionConfigRestrictionsGeoRestrictionList(list)
		return nil
	}
	return err
}

// CloudFrontDistributionConfigurationViewerCertificate represents CloudFront DistributionConfiguration ViewerCertificate
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cloudfront-distributionconfig-viewercertificate.html
type CloudFrontDistributionConfigurationViewerCertificate struct {
	// If you're using an alternate domain name, the Amazon Resource Name
	// (ARN) of an AWS Certificate Manager (ACM) certificate. Use the ACM
	// service to provision and manage your certificates. For more
	// information, see the AWS Certificate Manager User Guide.
	AcmCertificateArn *StringExpr `json:"AcmCertificateArn,omitempty"`

	// Indicates whether to use the default certificate for your CloudFront
	// domain name when viewers use HTTPS to request your content.
	CloudFrontDefaultCertificate *BoolExpr `json:"CloudFrontDefaultCertificate,omitempty"`

	// If you're using an alternate domain name, the ID of a server
	// certificate that was purchased from a certificate authority. This ID
	// is the ServerCertificateId value, which AWS Identity and Access
	// Management (IAM) returns when the certificate is added to the IAM
	// certificate store, such as ASCACKCEVSQ6CEXAMPLE1.
	IamCertificateId *StringExpr `json:"IamCertificateId,omitempty"`

	// The minimum version of the SSL protocol that you want CloudFront to
	// use for HTTPS connections. CloudFront serves your objects only to
	// browsers or devices that support at least the SSL version that you
	// specify. For valid values, see the MinimumProtocolVersion content for
	// the ViewerCertificate data type in the Amazon CloudFront API
	// Reference.
	MinimumProtocolVersion *StringExpr `json:"MinimumProtocolVersion,omitempty"`

	// Specifies how CloudFront serves HTTPS requests. For valid values, see
	// the SslSupportMethod content for the ViewerCertificate data type in
	// the Amazon CloudFront API Reference.
	SslSupportMethod *StringExpr `json:"SslSupportMethod,omitempty"`
}

// CloudFrontDistributionConfigurationViewerCertificateList represents a list of CloudFrontDistributionConfigurationViewerCertificate
type CloudFrontDistributionConfigurationViewerCertificateList []CloudFrontDistributionConfigurationViewerCertificate

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CloudFrontDistributionConfigurationViewerCertificateList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CloudFrontDistributionConfigurationViewerCertificate{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CloudFrontDistributionConfigurationViewerCertificateList{item}
		return nil
	}
	list := []CloudFrontDistributionConfigurationViewerCertificate{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CloudFrontDistributionConfigurationViewerCertificateList(list)
		return nil
	}
	return err
}

// CloudFrontForwardedValues represents CloudFront ForwardedValues
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cloudfront-forwardedvalues.html
type CloudFrontForwardedValues struct {
	// Forwards specified cookies to the origin of the cache behavior. For
	// more information, see Configuring CloudFront to Cache Based on Cookies
	// in the Amazon CloudFront Developer Guide.
	Cookies *CloudFrontForwardedValuesCookies `json:"Cookies,omitempty"`

	// Specifies the headers that you want Amazon CloudFront to forward to
	// the origin for this cache behavior (whitelisted headers). For the
	// headers that you specify, Amazon CloudFront also caches separate
	// versions of a specified object that is based on the header values in
	// viewer requests.
	Headers *StringListExpr `json:"Headers,omitempty"`

	// Indicates whether you want CloudFront to forward query strings to the
	// origin that is associated with this cache behavior. If so, specify
	// true; if not, specify false. For more information, see Configuring
	// CloudFront to Cache Based on Query String Parameters in the Amazon
	// CloudFront Developer Guide.
	QueryString *BoolExpr `json:"QueryString,omitempty"`

	// If you forward query strings to the origin, specifies the query string
	// parameters that CloudFront uses to determine which content to cache.
	// For more information, see Configuring CloudFront to Cache Based on
	// Query String Parameters in the Amazon CloudFront Developer Guide.
	QueryStringCacheKeys *StringListExpr `json:"QueryStringCacheKeys,omitempty"`
}

// CloudFrontForwardedValuesList represents a list of CloudFrontForwardedValues
type CloudFrontForwardedValuesList []CloudFrontForwardedValues

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CloudFrontForwardedValuesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CloudFrontForwardedValues{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CloudFrontForwardedValuesList{item}
		return nil
	}
	list := []CloudFrontForwardedValues{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CloudFrontForwardedValuesList(list)
		return nil
	}
	return err
}

// CloudFrontForwardedValuesCookies represents CloudFront ForwardedValues Cookies
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cloudfront-forwardedvalues-cookies.html
type CloudFrontForwardedValuesCookies struct {
	// The cookies to forward to the origin of the cache behavior. You can
	// specify none, all, or whitelist.
	Forward *StringExpr `json:"Forward,omitempty"`

	// The names of cookies to forward to the origin for the cache behavior.
	WhitelistedNames *StringListExpr `json:"WhitelistedNames,omitempty"`
}

// CloudFrontForwardedValuesCookiesList represents a list of CloudFrontForwardedValuesCookies
type CloudFrontForwardedValuesCookiesList []CloudFrontForwardedValuesCookies

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CloudFrontForwardedValuesCookiesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CloudFrontForwardedValuesCookies{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CloudFrontForwardedValuesCookiesList{item}
		return nil
	}
	list := []CloudFrontForwardedValuesCookies{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CloudFrontForwardedValuesCookiesList(list)
		return nil
	}
	return err
}

// CloudWatchMetricDimension represents CloudWatch Metric Dimension Property Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cw-dimension.html
type CloudWatchMetricDimension struct {
	// The name of the dimension, from 1–255 characters in length.
	Name *StringExpr `json:"Name,omitempty"`

	// The value representing the dimension measurement, from 1–255
	// characters in length.
	Value *StringExpr `json:"Value,omitempty"`
}

// CloudWatchMetricDimensionList represents a list of CloudWatchMetricDimension
type CloudWatchMetricDimensionList []CloudWatchMetricDimension

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CloudWatchMetricDimensionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CloudWatchMetricDimension{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CloudWatchMetricDimensionList{item}
		return nil
	}
	list := []CloudWatchMetricDimension{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CloudWatchMetricDimensionList(list)
		return nil
	}
	return err
}

// CloudWatchEventsRuleTarget represents Amazon CloudWatch Events Rule Target
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-events-rule-target.html
type CloudWatchEventsRuleTarget struct {
	// The Amazon Resource Name (ARN) of the target.
	Arn *StringExpr `json:"Arn,omitempty"`

	// A unique, user-defined identifier for the target. Acceptable values
	// include alphanumeric characters, periods (.), hypens (-), and
	// underscores (_).
	Id *StringExpr `json:"Id,omitempty"`

	// A JSON-formatted text string that is passed to the target. This value
	// overrides the matched event.
	Input *StringExpr `json:"Input,omitempty"`

	// When you don't want to pass the entire matched event, the JSONPath
	// that describes which part of the event to pass to the target.
	InputPath *StringExpr `json:"InputPath,omitempty"`
}

// CloudWatchEventsRuleTargetList represents a list of CloudWatchEventsRuleTarget
type CloudWatchEventsRuleTargetList []CloudWatchEventsRuleTarget

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CloudWatchEventsRuleTargetList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CloudWatchEventsRuleTarget{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CloudWatchEventsRuleTargetList{item}
		return nil
	}
	list := []CloudWatchEventsRuleTarget{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CloudWatchEventsRuleTargetList(list)
		return nil
	}
	return err
}

// CloudWatchLogsMetricFilterMetricTransformationProperty represents CloudWatch Logs MetricFilter MetricTransformation Property
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-logs-metricfilter-metrictransformation.html
type CloudWatchLogsMetricFilterMetricTransformationProperty struct {
	// The name of the CloudWatch metric to which the log information will be
	// published.
	MetricName *StringExpr `json:"MetricName,omitempty"`

	// The destination namespace of the CloudWatch metric. Namespaces are
	// containers for metrics. For example, you can add related metrics in
	// the same namespace.
	MetricNamespace *StringExpr `json:"MetricNamespace,omitempty"`

	// The value that is published to the CloudWatch metric. For example, if
	// you're counting the occurrences of a particular term like Error,
	// specify 1 for the metric value. If you're counting the number of bytes
	// transferred, reference the value that is in the log event by using $
	// followed by the name of the field that you specified in the filter
	// pattern, such as $size.
	MetricValue *StringExpr `json:"MetricValue,omitempty"`
}

// CloudWatchLogsMetricFilterMetricTransformationPropertyList represents a list of CloudWatchLogsMetricFilterMetricTransformationProperty
type CloudWatchLogsMetricFilterMetricTransformationPropertyList []CloudWatchLogsMetricFilterMetricTransformationProperty

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CloudWatchLogsMetricFilterMetricTransformationPropertyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CloudWatchLogsMetricFilterMetricTransformationProperty{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CloudWatchLogsMetricFilterMetricTransformationPropertyList{item}
		return nil
	}
	list := []CloudWatchLogsMetricFilterMetricTransformationProperty{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CloudWatchLogsMetricFilterMetricTransformationPropertyList(list)
		return nil
	}
	return err
}

// CodeCommitRepositoryTrigger represents AWS CodeCommit Repository Trigger
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-codecommit-repository-triggers.html
type CodeCommitRepositoryTrigger struct {
	// The names of the branches in the AWS CodeCommit repository that
	// contain events that you want to include in the trigger. If you don't
	// specify at least one branch, the trigger applies to all branches.
	Branches *StringListExpr `json:"Branches,omitempty"`

	// When an event is triggered, additional information that AWS CodeCommit
	// includes when it sends information to the target.
	CustomData *StringExpr `json:"CustomData,omitempty"`

	// The Amazon Resource Name (ARN) of the resource that is the target for
	// this trigger. For valid targets, see Manage Triggers for an AWS
	// CodeCommit Repository in the AWS CodeCommit User Guide.
	DestinationArn *StringExpr `json:"DestinationArn,omitempty"`

	// The repository events for which AWS CodeCommit sends information to
	// the target, which you specified in the DestinationArn property. If you
	// don't specify events, the trigger runs for all repository events. For
	// valid values, see the RepositoryTrigger data type in the AWS
	// CodeCommit API Reference.
	Events *StringListExpr `json:"Events,omitempty"`

	// A name for the trigger.
	Name *StringExpr `json:"Name,omitempty"`
}

// CodeCommitRepositoryTriggerList represents a list of CodeCommitRepositoryTrigger
type CodeCommitRepositoryTriggerList []CodeCommitRepositoryTrigger

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CodeCommitRepositoryTriggerList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CodeCommitRepositoryTrigger{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CodeCommitRepositoryTriggerList{item}
		return nil
	}
	list := []CodeCommitRepositoryTrigger{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CodeCommitRepositoryTriggerList(list)
		return nil
	}
	return err
}

// CodeDeployDeploymentConfigMinimumHealthyHosts represents AWS CodeDeploy DeploymentConfig MinimumHealthyHosts
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-codedeploy-deploymentconfig-minimumhealthyhosts.html
type CodeDeployDeploymentConfigMinimumHealthyHosts struct {
	// The type of count to use, such as an absolute value or a percentage of
	// the total number of instances in the deployment. For valid values, see
	// MinimumHealthyHosts in the AWS CodeDeploy API Reference.
	Type *StringExpr `json:"Type,omitempty"`

	// The minimum number of healthy instances.
	Value *IntegerExpr `json:"Value,omitempty"`
}

// CodeDeployDeploymentConfigMinimumHealthyHostsList represents a list of CodeDeployDeploymentConfigMinimumHealthyHosts
type CodeDeployDeploymentConfigMinimumHealthyHostsList []CodeDeployDeploymentConfigMinimumHealthyHosts

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CodeDeployDeploymentConfigMinimumHealthyHostsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CodeDeployDeploymentConfigMinimumHealthyHosts{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CodeDeployDeploymentConfigMinimumHealthyHostsList{item}
		return nil
	}
	list := []CodeDeployDeploymentConfigMinimumHealthyHosts{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CodeDeployDeploymentConfigMinimumHealthyHostsList(list)
		return nil
	}
	return err
}

// CodeDeployDeploymentGroupDeployment represents AWS CodeDeploy DeploymentGroup Deployment
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-codedeploy-deploymentgroup-deployment.html
type CodeDeployDeploymentGroupDeployment struct {
	// A description about this deployment.
	Description *StringExpr `json:"Description,omitempty"`

	// Whether to continue the deployment if the ApplicationStop deployment
	// lifecycle event fails. If you want AWS CodeDeploy to continue the
	// deployment lifecycle even if the ApplicationStop event fails on an
	// instance, specify true. The deployment continues to the BeforeInstall
	// deployment lifecycle event. If you want AWS CodeDeploy to stop
	// deployment on the instance if the ApplicationStop event fails, specify
	// false or do not specify a value.
	IgnoreApplicationStopFailures *BoolExpr `json:"IgnoreApplicationStopFailures,omitempty"`

	// The location of the application revision to deploy.
	Revision *CodeDeployDeploymentGroupDeploymentRevision `json:"Revision,omitempty"`
}

// CodeDeployDeploymentGroupDeploymentList represents a list of CodeDeployDeploymentGroupDeployment
type CodeDeployDeploymentGroupDeploymentList []CodeDeployDeploymentGroupDeployment

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CodeDeployDeploymentGroupDeploymentList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CodeDeployDeploymentGroupDeployment{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CodeDeployDeploymentGroupDeploymentList{item}
		return nil
	}
	list := []CodeDeployDeploymentGroupDeployment{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CodeDeployDeploymentGroupDeploymentList(list)
		return nil
	}
	return err
}

// CodeDeployDeploymentGroupDeploymentRevision represents AWS CodeDeploy DeploymentGroup Deployment Revision
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-codedeploy-deploymentgroup-deployment-revision.html
type CodeDeployDeploymentGroupDeploymentRevision struct {
	// If your application revision is stored in GitHub, information about
	// the location where it is stored.
	GitHubLocation *CodeDeployDeploymentGroupDeploymentRevisionGitHubLocation `json:"GitHubLocation,omitempty"`

	// The application revision's location, such as in an S3 bucket or GitHub
	// repository. For valid values, see RevisionLocation in the AWS
	// CodeDeploy API Reference.
	RevisionType *StringExpr `json:"RevisionType,omitempty"`

	// If the application revision is stored in an S3 bucket, information
	// about the location.
	S3Location *CodeDeployDeploymentGroupDeploymentRevisionS3Location `json:"S3Location,omitempty"`
}

// CodeDeployDeploymentGroupDeploymentRevisionList represents a list of CodeDeployDeploymentGroupDeploymentRevision
type CodeDeployDeploymentGroupDeploymentRevisionList []CodeDeployDeploymentGroupDeploymentRevision

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CodeDeployDeploymentGroupDeploymentRevisionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CodeDeployDeploymentGroupDeploymentRevision{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CodeDeployDeploymentGroupDeploymentRevisionList{item}
		return nil
	}
	list := []CodeDeployDeploymentGroupDeploymentRevision{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CodeDeployDeploymentGroupDeploymentRevisionList(list)
		return nil
	}
	return err
}

// CodeDeployDeploymentGroupDeploymentRevisionGitHubLocation represents AWS CodeDeploy DeploymentGroup Deployment Revision GitHubLocation
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-codedeploy-deploymentgroup-deployment-revision-githublocation.html
type CodeDeployDeploymentGroupDeploymentRevisionGitHubLocation struct {
	// The SHA1 commit ID of the GitHub commit to use as your application
	// revision.
	CommitId *StringExpr `json:"CommitId,omitempty"`

	// The GitHub account and repository name that includes the application
	// revision. Specify the value as account/repository_name.
	Repository *StringExpr `json:"Repository,omitempty"`
}

// CodeDeployDeploymentGroupDeploymentRevisionGitHubLocationList represents a list of CodeDeployDeploymentGroupDeploymentRevisionGitHubLocation
type CodeDeployDeploymentGroupDeploymentRevisionGitHubLocationList []CodeDeployDeploymentGroupDeploymentRevisionGitHubLocation

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CodeDeployDeploymentGroupDeploymentRevisionGitHubLocationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CodeDeployDeploymentGroupDeploymentRevisionGitHubLocation{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CodeDeployDeploymentGroupDeploymentRevisionGitHubLocationList{item}
		return nil
	}
	list := []CodeDeployDeploymentGroupDeploymentRevisionGitHubLocation{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CodeDeployDeploymentGroupDeploymentRevisionGitHubLocationList(list)
		return nil
	}
	return err
}

// CodeDeployDeploymentGroupDeploymentRevisionS3Location represents AWS CodeDeploy DeploymentGroup Deployment Revision S3Location
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-codedeploy-deploymentgroup-deployment-revision-s3location.html
type CodeDeployDeploymentGroupDeploymentRevisionS3Location struct {
	// The name of the S3 bucket where the application revision is stored.
	Bucket *StringExpr `json:"Bucket,omitempty"`

	// The file type of the application revision, such as tar, tgz, or zip.
	// For valid values, see S3Location in the AWS CodeDeploy API Reference.
	BundleType *StringExpr `json:"BundleType,omitempty"`

	// The Amazon S3 ETag (a file checksum) of the application revision. If
	// you don't specify a value, AWS CodeDeploy skips the ETag validation of
	// your application revision.
	ETag *StringExpr `json:"ETag,omitempty"`

	// The file name of the application revision (Amazon S3 object name).
	Key *StringExpr `json:"Key,omitempty"`

	// For versioning-enabled buckets, a specific version of the application
	// revision.
	Version *StringExpr `json:"Version,omitempty"`
}

// CodeDeployDeploymentGroupDeploymentRevisionS3LocationList represents a list of CodeDeployDeploymentGroupDeploymentRevisionS3Location
type CodeDeployDeploymentGroupDeploymentRevisionS3LocationList []CodeDeployDeploymentGroupDeploymentRevisionS3Location

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CodeDeployDeploymentGroupDeploymentRevisionS3LocationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CodeDeployDeploymentGroupDeploymentRevisionS3Location{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CodeDeployDeploymentGroupDeploymentRevisionS3LocationList{item}
		return nil
	}
	list := []CodeDeployDeploymentGroupDeploymentRevisionS3Location{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CodeDeployDeploymentGroupDeploymentRevisionS3LocationList(list)
		return nil
	}
	return err
}

// CodeDeployDeploymentGroupEc2TagFilters represents AWS CodeDeploy DeploymentGroup Ec2TagFilters
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-codedeploy-deploymentgroup-ec2tagfilters.html
type CodeDeployDeploymentGroupEc2TagFilters struct {
	// Filter instances with this key.
	Key *StringExpr `json:"Key,omitempty"`

	// The filter type. For example, you can filter instances by the key, tag
	// value, or both. For valid values, see EC2TagFilter in the AWS
	// CodeDeploy API Reference.
	Type *StringExpr `json:"Type,omitempty"`

	// Filter instances with this tag value.
	Value *StringExpr `json:"Value,omitempty"`
}

// CodeDeployDeploymentGroupEc2TagFiltersList represents a list of CodeDeployDeploymentGroupEc2TagFilters
type CodeDeployDeploymentGroupEc2TagFiltersList []CodeDeployDeploymentGroupEc2TagFilters

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CodeDeployDeploymentGroupEc2TagFiltersList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CodeDeployDeploymentGroupEc2TagFilters{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CodeDeployDeploymentGroupEc2TagFiltersList{item}
		return nil
	}
	list := []CodeDeployDeploymentGroupEc2TagFilters{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CodeDeployDeploymentGroupEc2TagFiltersList(list)
		return nil
	}
	return err
}

// CodeDeployDeploymentGroupOnPremisesInstanceTagFilters represents AWS CodeDeploy DeploymentGroup OnPremisesInstanceTagFilters
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-codedeploy-deploymentgroup-onpremisesinstancetagfilters.html
type CodeDeployDeploymentGroupOnPremisesInstanceTagFilters struct {
	// Filter on-premises instances with this key.
	Key *StringExpr `json:"Key,omitempty"`

	// The filter type. For example, you can filter on-premises instances by
	// the key, tag value, or both. For valid values, see EC2TagFilter in the
	// AWS CodeDeploy API Reference.
	Type *StringExpr `json:"Type,omitempty"`

	// Filter on-premises instances with this tag value.
	Value *StringExpr `json:"Value,omitempty"`
}

// CodeDeployDeploymentGroupOnPremisesInstanceTagFiltersList represents a list of CodeDeployDeploymentGroupOnPremisesInstanceTagFilters
type CodeDeployDeploymentGroupOnPremisesInstanceTagFiltersList []CodeDeployDeploymentGroupOnPremisesInstanceTagFilters

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CodeDeployDeploymentGroupOnPremisesInstanceTagFiltersList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CodeDeployDeploymentGroupOnPremisesInstanceTagFilters{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CodeDeployDeploymentGroupOnPremisesInstanceTagFiltersList{item}
		return nil
	}
	list := []CodeDeployDeploymentGroupOnPremisesInstanceTagFilters{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CodeDeployDeploymentGroupOnPremisesInstanceTagFiltersList(list)
		return nil
	}
	return err
}

// CodePipelineCustomActionTypeArtifactDetails represents AWS CodePipeline CustomActionType ArtifactDetails
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-codepipeline-customactiontype-artifactdetails.html
type CodePipelineCustomActionTypeArtifactDetails struct {
	// The maximum number of artifacts allowed for the action type.
	MaximumCount *IntegerExpr `json:"MaximumCount,omitempty"`

	// The minimum number of artifacts allowed for the action type.
	MinimumCount *IntegerExpr `json:"MinimumCount,omitempty"`
}

// CodePipelineCustomActionTypeArtifactDetailsList represents a list of CodePipelineCustomActionTypeArtifactDetails
type CodePipelineCustomActionTypeArtifactDetailsList []CodePipelineCustomActionTypeArtifactDetails

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CodePipelineCustomActionTypeArtifactDetailsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CodePipelineCustomActionTypeArtifactDetails{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CodePipelineCustomActionTypeArtifactDetailsList{item}
		return nil
	}
	list := []CodePipelineCustomActionTypeArtifactDetails{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CodePipelineCustomActionTypeArtifactDetailsList(list)
		return nil
	}
	return err
}

// CodePipelineCustomActionTypeConfigurationProperties represents AWS CodePipeline CustomActionType ConfigurationProperties
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-codepipeline-customactiontype-configurationproperties.html
type CodePipelineCustomActionTypeConfigurationProperties struct {
	// A description of this configuration property that will be displayed to
	// users.
	Description *StringExpr `json:"Description,omitempty"`

	// Indicates whether the configuration property is a key.
	Key *BoolExpr `json:"Key,omitempty"`

	// A name for this configuration property.
	Name *StringExpr `json:"Name,omitempty"`

	// Indicates whether the configuration property will be used with the
	// PollForJobs call. A custom action can have one queryable property. The
	// queryable property must be required (see the Required property) and
	// must not be secret (see the Secret property). For more information,
	// see the queryable contents for the ActionConfigurationProperty data
	// type in the AWS CodePipeline API Reference.
	Queryable *BoolExpr `json:"Queryable,omitempty"`

	// Indicates whether the configuration property is a required value.
	Required *BoolExpr `json:"Required,omitempty"`

	// Indicates whether the configuration property is secret. Secret
	// configuration properties are hidden from all AWS CodePipeline calls
	// except for GetJobDetails, GetThirdPartyJobDetails, PollForJobs, and
	// PollForThirdPartyJobs.
	Secret *BoolExpr `json:"Secret,omitempty"`

	// The type of the configuration property, such as String, Number, or
	// Boolean.
	Type *StringExpr `json:"Type,omitempty"`
}

// CodePipelineCustomActionTypeConfigurationPropertiesList represents a list of CodePipelineCustomActionTypeConfigurationProperties
type CodePipelineCustomActionTypeConfigurationPropertiesList []CodePipelineCustomActionTypeConfigurationProperties

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CodePipelineCustomActionTypeConfigurationPropertiesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CodePipelineCustomActionTypeConfigurationProperties{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CodePipelineCustomActionTypeConfigurationPropertiesList{item}
		return nil
	}
	list := []CodePipelineCustomActionTypeConfigurationProperties{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CodePipelineCustomActionTypeConfigurationPropertiesList(list)
		return nil
	}
	return err
}

// CodePipelineCustomActionTypeSettings represents AWS CodePipeline CustomActionType Settings
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-codepipeline-customactiontype-settings.html
type CodePipelineCustomActionTypeSettings struct {
	// The URL that is returned to the AWS CodePipeline console that links to
	// the resources of the external system, such as the configuration page
	// for an AWS CodeDeploy deployment group.
	EntityUrlTemplate *StringExpr `json:"EntityUrlTemplate,omitempty"`

	// The URL that is returned to the AWS CodePipeline console that links to
	// the top-level landing page for the external system, such as the
	// console page for AWS CodeDeploy.
	ExecutionUrlTemplate *StringExpr `json:"ExecutionUrlTemplate,omitempty"`

	// The URL that is returned to the AWS CodePipeline console that links to
	// the page where customers can update or change the configuration of the
	// external action.
	RevisionUrlTemplate *StringExpr `json:"RevisionUrlTemplate,omitempty"`

	// The URL of a sign-up page where users can sign up for an external
	// service and specify the initial configurations for the service's
	// action.
	ThirdPartyConfigurationUrl *StringExpr `json:"ThirdPartyConfigurationUrl,omitempty"`
}

// CodePipelineCustomActionTypeSettingsList represents a list of CodePipelineCustomActionTypeSettings
type CodePipelineCustomActionTypeSettingsList []CodePipelineCustomActionTypeSettings

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CodePipelineCustomActionTypeSettingsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CodePipelineCustomActionTypeSettings{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CodePipelineCustomActionTypeSettingsList{item}
		return nil
	}
	list := []CodePipelineCustomActionTypeSettings{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CodePipelineCustomActionTypeSettingsList(list)
		return nil
	}
	return err
}

// CodePipelinePipelineArtifactStore represents AWS CodePipeline Pipeline ArtifactStore
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-codepipeline-pipeline-artifactstore.html
type CodePipelinePipelineArtifactStore struct {
	// The encryption key AWS CodePipeline uses to encrypt the data in the
	// artifact store, such as an AWS Key Management Service (AWS KMS) key.
	// If you don't specify a key, AWS CodePipeline uses the default key for
	// Amazon Simple Storage Service (Amazon S3).
	EncryptionKey *CodePipelinePipelineArtifactStoreEncryptionKey `json:"EncryptionKey,omitempty"`

	// The location where AWS CodePipeline stores artifacts for a pipeline,
	// such as an S3 bucket.
	Location *StringExpr `json:"Location,omitempty"`

	// The type of the artifact store, such as Amazon S3. For valid values,
	// see ArtifactStore in the AWS CodePipeline API Reference.
	Type *StringExpr `json:"Type,omitempty"`
}

// CodePipelinePipelineArtifactStoreList represents a list of CodePipelinePipelineArtifactStore
type CodePipelinePipelineArtifactStoreList []CodePipelinePipelineArtifactStore

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CodePipelinePipelineArtifactStoreList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CodePipelinePipelineArtifactStore{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CodePipelinePipelineArtifactStoreList{item}
		return nil
	}
	list := []CodePipelinePipelineArtifactStore{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CodePipelinePipelineArtifactStoreList(list)
		return nil
	}
	return err
}

// CodePipelinePipelineArtifactStoreEncryptionKey represents AWS CodePipeline Pipeline ArtifactStore EncryptionKey
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-codepipeline-pipeline-artifactstore-encryptionkey.html
type CodePipelinePipelineArtifactStoreEncryptionKey struct {
	// The ID of the key. For an AWS KMS key, specify the key ID or key
	// Amazon Resource Number (ARN).
	Id *StringExpr `json:"Id,omitempty"`

	// The type of encryption key, such as KMS. For valid values, see
	// EncryptionKey in the AWS CodePipeline API Reference.
	Type *StringExpr `json:"Type,omitempty"`
}

// CodePipelinePipelineArtifactStoreEncryptionKeyList represents a list of CodePipelinePipelineArtifactStoreEncryptionKey
type CodePipelinePipelineArtifactStoreEncryptionKeyList []CodePipelinePipelineArtifactStoreEncryptionKey

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CodePipelinePipelineArtifactStoreEncryptionKeyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CodePipelinePipelineArtifactStoreEncryptionKey{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CodePipelinePipelineArtifactStoreEncryptionKeyList{item}
		return nil
	}
	list := []CodePipelinePipelineArtifactStoreEncryptionKey{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CodePipelinePipelineArtifactStoreEncryptionKeyList(list)
		return nil
	}
	return err
}

// CodePipelinePipelineDisableInboundStageTransitions represents AWS CodePipeline Pipeline DisableInboundStageTransitions
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-codepipeline-pipeline-disableinboundstagetransitions.html
type CodePipelinePipelineDisableInboundStageTransitions struct {
	// An explanation of why the transition between two stages of a pipeline
	// was disabled.
	Reason *StringExpr `json:"Reason,omitempty"`

	// The name of the stage to which transitions are disabled.
	StageName *StringExpr `json:"StageName,omitempty"`
}

// CodePipelinePipelineDisableInboundStageTransitionsList represents a list of CodePipelinePipelineDisableInboundStageTransitions
type CodePipelinePipelineDisableInboundStageTransitionsList []CodePipelinePipelineDisableInboundStageTransitions

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CodePipelinePipelineDisableInboundStageTransitionsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CodePipelinePipelineDisableInboundStageTransitions{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CodePipelinePipelineDisableInboundStageTransitionsList{item}
		return nil
	}
	list := []CodePipelinePipelineDisableInboundStageTransitions{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CodePipelinePipelineDisableInboundStageTransitionsList(list)
		return nil
	}
	return err
}

// CodePipelinePipelineStages represents AWS CodePipeline Pipeline Stages
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-codepipeline-pipeline-stages.html
type CodePipelinePipelineStages struct {
	// The actions to include in this stage.
	Actions *CodePipelinePipelineStagesActionsList `json:"Actions,omitempty"`

	// The gates included in a stage.
	Blockers *CodePipelinePipelineStagesBlockersList `json:"Blockers,omitempty"`

	// A name for this stage.
	Name *StringExpr `json:"Name,omitempty"`
}

// CodePipelinePipelineStagesList represents a list of CodePipelinePipelineStages
type CodePipelinePipelineStagesList []CodePipelinePipelineStages

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CodePipelinePipelineStagesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CodePipelinePipelineStages{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CodePipelinePipelineStagesList{item}
		return nil
	}
	list := []CodePipelinePipelineStages{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CodePipelinePipelineStagesList(list)
		return nil
	}
	return err
}

// CodePipelinePipelineStagesActions represents AWS CodePipeline Pipeline Stages Actions
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-codepipeline-pipeline-stages-actions.html
type CodePipelinePipelineStagesActions struct {
	// Specifies the action type and the provider of the action.
	ActionTypeId *CodePipelinePipelineStagesActionsActionTypeId `json:"ActionTypeId,omitempty"`

	// The action's configuration. These are key-value pairs that specify
	// input values for an action.
	Configuration interface{} `json:"Configuration,omitempty"`

	// The name or ID of the artifact that the action consumes, such as a
	// test or build artifact.
	InputArtifacts *CodePipelinePipelineStagesActionsInputArtifactsList `json:"InputArtifacts,omitempty"`

	// The action name.
	Name *StringExpr `json:"Name,omitempty"`

	// The artifact name or ID that is a result of the action, such as a test
	// or build artifact.
	OutputArtifacts *CodePipelinePipelineStagesActionsOutputArtifactsList `json:"OutputArtifacts,omitempty"`

	// The Amazon Resource Name (ARN) of a service role that the action uses.
	// The pipeline's role assumes this role.
	RoleArn *StringExpr `json:"RoleArn,omitempty"`

	// The order in which AWS CodePipeline runs this action.
	RunOrder *IntegerExpr `json:"RunOrder,omitempty"`
}

// CodePipelinePipelineStagesActionsList represents a list of CodePipelinePipelineStagesActions
type CodePipelinePipelineStagesActionsList []CodePipelinePipelineStagesActions

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CodePipelinePipelineStagesActionsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CodePipelinePipelineStagesActions{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CodePipelinePipelineStagesActionsList{item}
		return nil
	}
	list := []CodePipelinePipelineStagesActions{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CodePipelinePipelineStagesActionsList(list)
		return nil
	}
	return err
}

// CodePipelinePipelineStagesActionsActionTypeId represents AWS CodePipeline Pipeline Stages Actions ActionTypeId
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-codepipeline-pipeline-stages-actions-actiontypeid.html
type CodePipelinePipelineStagesActionsActionTypeId struct {
	// A category that defines which action type the owner (the entitiy that
	// performs the action) performs. The category that you select determine
	// the providers that you can specify for the Provider property. For
	// valid values, see ActionTypeId in the AWS CodePipeline API Reference.
	Category *StringExpr `json:"Category,omitempty"`

	// The entity that performs the action. For valid values, see
	// ActionTypeId in the AWS CodePipeline API Reference.
	Owner *StringExpr `json:"Owner,omitempty"`

	// The service provider that the action calls. The providers that you can
	// specify are determined by the category that you select. For example, a
	// valid provider for the Deploy category is AWS CodeDeploy, which you
	// would specify as CodeDeploy.
	Provider *StringExpr `json:"Provider,omitempty"`

	// A version identifier for this action.
	Version *StringExpr `json:"Version,omitempty"`
}

// CodePipelinePipelineStagesActionsActionTypeIdList represents a list of CodePipelinePipelineStagesActionsActionTypeId
type CodePipelinePipelineStagesActionsActionTypeIdList []CodePipelinePipelineStagesActionsActionTypeId

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CodePipelinePipelineStagesActionsActionTypeIdList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CodePipelinePipelineStagesActionsActionTypeId{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CodePipelinePipelineStagesActionsActionTypeIdList{item}
		return nil
	}
	list := []CodePipelinePipelineStagesActionsActionTypeId{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CodePipelinePipelineStagesActionsActionTypeIdList(list)
		return nil
	}
	return err
}

// CodePipelinePipelineStagesActionsInputArtifacts represents AWS CodePipeline Pipeline Stages Actions InputArtifacts
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-codepipeline-pipeline-stages-actions-inputartifacts.html
type CodePipelinePipelineStagesActionsInputArtifacts struct {
	// The name of the artifact that the AWS CodePipeline action works on,
	// such as My App.The input artifact of an action must match the output
	// artifact from any preceding action.
	Name *StringExpr `json:"Name,omitempty"`
}

// CodePipelinePipelineStagesActionsInputArtifactsList represents a list of CodePipelinePipelineStagesActionsInputArtifacts
type CodePipelinePipelineStagesActionsInputArtifactsList []CodePipelinePipelineStagesActionsInputArtifacts

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CodePipelinePipelineStagesActionsInputArtifactsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CodePipelinePipelineStagesActionsInputArtifacts{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CodePipelinePipelineStagesActionsInputArtifactsList{item}
		return nil
	}
	list := []CodePipelinePipelineStagesActionsInputArtifacts{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CodePipelinePipelineStagesActionsInputArtifactsList(list)
		return nil
	}
	return err
}

// CodePipelinePipelineStagesActionsOutputArtifacts represents AWS CodePipeline Pipeline Stages Actions OutputArtifacts
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-codepipeline-pipeline-stages-actions-outputartifacts.html
type CodePipelinePipelineStagesActionsOutputArtifacts struct {
	// The name of the artifact that is the result of an AWS CodePipeline
	// action, such as My App. Output artifact names must be unique within a
	// pipeline.
	Name *StringExpr `json:"Name,omitempty"`
}

// CodePipelinePipelineStagesActionsOutputArtifactsList represents a list of CodePipelinePipelineStagesActionsOutputArtifacts
type CodePipelinePipelineStagesActionsOutputArtifactsList []CodePipelinePipelineStagesActionsOutputArtifacts

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CodePipelinePipelineStagesActionsOutputArtifactsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CodePipelinePipelineStagesActionsOutputArtifacts{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CodePipelinePipelineStagesActionsOutputArtifactsList{item}
		return nil
	}
	list := []CodePipelinePipelineStagesActionsOutputArtifacts{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CodePipelinePipelineStagesActionsOutputArtifactsList(list)
		return nil
	}
	return err
}

// CodePipelinePipelineStagesBlockers represents AWS CodePipeline Pipeline Stages Blockers
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-codepipeline-pipeline-stages-blockers.html
type CodePipelinePipelineStagesBlockers struct {
	// The name of the gate declaration.
	Name *StringExpr `json:"Name,omitempty"`

	// The type of gate declaration. For valid values, see BlockerDeclaration
	// in the AWS CodePipeline API Reference.
	Type *StringExpr `json:"Type,omitempty"`
}

// CodePipelinePipelineStagesBlockersList represents a list of CodePipelinePipelineStagesBlockers
type CodePipelinePipelineStagesBlockersList []CodePipelinePipelineStagesBlockers

// UnmarshalJSON sets the object from the provided JSON representation
func (l *CodePipelinePipelineStagesBlockersList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := CodePipelinePipelineStagesBlockers{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = CodePipelinePipelineStagesBlockersList{item}
		return nil
	}
	list := []CodePipelinePipelineStagesBlockers{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = CodePipelinePipelineStagesBlockersList(list)
		return nil
	}
	return err
}

// ConfigConfigRuleScope represents AWS Config ConfigRule Scope
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-config-configrule-scope.html
type ConfigConfigRuleScope struct {
	// The ID of an AWS resource that you want AWS Config to evaluate against
	// a rule. If you specify an ID, you must also specify a resource type
	// for the ComplianceResourceTypes property.
	ComplianceResourceId *StringExpr `json:"ComplianceResourceId,omitempty"`

	// The types of AWS resources that you want AWS Config to evaluate
	// against the rule. If you specify the ComplianceResourceId property,
	// specify only one resource type.
	ComplianceResourceTypes *StringListExpr `json:"ComplianceResourceTypes,omitempty"`

	// The tag key that is applied to the AWS resources that you want AWS
	// Config to evaluate against the rule.
	TagKey *StringExpr `json:"TagKey,omitempty"`

	// The tag value that is applied to the AWS resources that you want AWS
	// Config to evaluate against the rule.
	TagValue *StringExpr `json:"TagValue,omitempty"`
}

// ConfigConfigRuleScopeList represents a list of ConfigConfigRuleScope
type ConfigConfigRuleScopeList []ConfigConfigRuleScope

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ConfigConfigRuleScopeList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ConfigConfigRuleScope{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ConfigConfigRuleScopeList{item}
		return nil
	}
	list := []ConfigConfigRuleScope{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ConfigConfigRuleScopeList(list)
		return nil
	}
	return err
}

// ConfigConfigRuleSource represents AWS Config ConfigRule Source
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-config-configrule-source.html
type ConfigConfigRuleSource struct {
	// Indicates who owns and manages the AWS Config rule. For valid values,
	// see the Source data type in the AWS Config API Reference.
	Owner *StringExpr `json:"Owner,omitempty"`

	// Provides the source and type of event that triggers AWS Config to
	// evaluate your AWS resources.
	SourceDetails *ConfigConfigRuleSourceSourceDetailsList `json:"SourceDetails,omitempty"`

	// For AWS managed rules, the identifier of the rule. For a list of
	// identifiers, see AWS Managed Rules in the AWS Config Developer Guide.
	SourceIdentifier *StringExpr `json:"SourceIdentifier,omitempty"`
}

// ConfigConfigRuleSourceList represents a list of ConfigConfigRuleSource
type ConfigConfigRuleSourceList []ConfigConfigRuleSource

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ConfigConfigRuleSourceList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ConfigConfigRuleSource{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ConfigConfigRuleSourceList{item}
		return nil
	}
	list := []ConfigConfigRuleSource{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ConfigConfigRuleSourceList(list)
		return nil
	}
	return err
}

// ConfigConfigRuleSourceSourceDetails represents AWS Config ConfigRule Source SourceDetails
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-config-configrule-source-sourcedetails.html
type ConfigConfigRuleSourceSourceDetails struct {
	// The source, such as an AWS service, that generate events, triggering
	// AWS Config to evaluate your AWS resources. For valid values, see the
	// SourceDetail data type in the AWS Config API Reference.
	EventSource *StringExpr `json:"EventSource,omitempty"`

	// The type of Amazon Simple Notification Service (Amazon SNS) message
	// that triggers AWS Config to run an evaluation.
	MessageType *StringExpr `json:"MessageType,omitempty"`
}

// ConfigConfigRuleSourceSourceDetailsList represents a list of ConfigConfigRuleSourceSourceDetails
type ConfigConfigRuleSourceSourceDetailsList []ConfigConfigRuleSourceSourceDetails

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ConfigConfigRuleSourceSourceDetailsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ConfigConfigRuleSourceSourceDetails{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ConfigConfigRuleSourceSourceDetailsList{item}
		return nil
	}
	list := []ConfigConfigRuleSourceSourceDetails{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ConfigConfigRuleSourceSourceDetailsList(list)
		return nil
	}
	return err
}

// ConfigConfigurationRecorderRecordingGroup represents AWS Config ConfigurationRecorder RecordingGroup
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-config-configurationrecorder-recordinggroup.html
type ConfigConfigurationRecorderRecordingGroup struct {
	// Indicates whether to record all supported resource types. If you
	// specify this property, do not specify the ResourceTypes property.
	AllSupported *BoolExpr `json:"AllSupported,omitempty"`

	// Indicates whether AWS Config records all supported global resource
	// types. When AWS Config supports new global resource types, AWS Config
	// will automatically start recording them if you enable this property.
	IncludeGlobalResourceTypes *BoolExpr `json:"IncludeGlobalResourceTypes,omitempty"`

	// A list of valid AWS resource types to include in this recording group,
	// such as AWS::EC2::Instance or AWS::CloudTrail::Trail. If you specify
	// this property, do not specify the AllSupported property. For a list of
	// supported resource types, see Supported resource types in the AWS
	// Config Developer Guide.
	ResourceTypes *StringListExpr `json:"ResourceTypes,omitempty"`
}

// ConfigConfigurationRecorderRecordingGroupList represents a list of ConfigConfigurationRecorderRecordingGroup
type ConfigConfigurationRecorderRecordingGroupList []ConfigConfigurationRecorderRecordingGroup

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ConfigConfigurationRecorderRecordingGroupList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ConfigConfigurationRecorderRecordingGroup{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ConfigConfigurationRecorderRecordingGroupList{item}
		return nil
	}
	list := []ConfigConfigurationRecorderRecordingGroup{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ConfigConfigurationRecorderRecordingGroupList(list)
		return nil
	}
	return err
}

// ConfigDeliveryChannelConfigSnapshotDeliveryProperties represents AWS Config DeliveryChannel ConfigSnapshotDeliveryProperties
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-config-deliverychannel-configsnapshotdeliveryproperties.html
type ConfigDeliveryChannelConfigSnapshotDeliveryProperties struct {
	// The frequency with which AWS Config delivers configuration snapshots.
	// For valid values, see ConfigSnapshotDeliveryProperties in the AWS
	// Config API Reference.
	DeliveryFrequency *StringExpr `json:"DeliveryFrequency,omitempty"`
}

// ConfigDeliveryChannelConfigSnapshotDeliveryPropertiesList represents a list of ConfigDeliveryChannelConfigSnapshotDeliveryProperties
type ConfigDeliveryChannelConfigSnapshotDeliveryPropertiesList []ConfigDeliveryChannelConfigSnapshotDeliveryProperties

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ConfigDeliveryChannelConfigSnapshotDeliveryPropertiesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ConfigDeliveryChannelConfigSnapshotDeliveryProperties{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ConfigDeliveryChannelConfigSnapshotDeliveryPropertiesList{item}
		return nil
	}
	list := []ConfigDeliveryChannelConfigSnapshotDeliveryProperties{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ConfigDeliveryChannelConfigSnapshotDeliveryPropertiesList(list)
		return nil
	}
	return err
}

// DataPipelinePipelineParameterObjects represents AWS Data Pipeline Pipeline ParameterObjects
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-datapipeline-pipeline-parameterobjects.html
type DataPipelinePipelineParameterObjects struct {
	// Key-value pairs that define the attributes of the parameter object.
	Attributes *DataPipelineParameterObjectsAttributesList `json:"Attributes,omitempty"`

	// The identifier of the parameter object.
	Id *StringExpr `json:"Id,omitempty"`
}

// DataPipelinePipelineParameterObjectsList represents a list of DataPipelinePipelineParameterObjects
type DataPipelinePipelineParameterObjectsList []DataPipelinePipelineParameterObjects

// UnmarshalJSON sets the object from the provided JSON representation
func (l *DataPipelinePipelineParameterObjectsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := DataPipelinePipelineParameterObjects{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = DataPipelinePipelineParameterObjectsList{item}
		return nil
	}
	list := []DataPipelinePipelineParameterObjects{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = DataPipelinePipelineParameterObjectsList(list)
		return nil
	}
	return err
}

// DataPipelineParameterObjectsAttributes represents AWS Data Pipeline Parameter Objects Attributes
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-datapipeline-pipeline-parameterobjects-attributes.html
type DataPipelineParameterObjectsAttributes struct {
	// Specifies the name of a parameter attribute. To view parameter
	// attributes, see Creating a Pipeline Using Parameterized Templates in
	// the AWS Data Pipeline Developer Guide.
	Key *StringExpr `json:"Key,omitempty"`

	// A parameter attribute value.
	StringValue *StringExpr `json:"StringValue,omitempty"`
}

// DataPipelineParameterObjectsAttributesList represents a list of DataPipelineParameterObjectsAttributes
type DataPipelineParameterObjectsAttributesList []DataPipelineParameterObjectsAttributes

// UnmarshalJSON sets the object from the provided JSON representation
func (l *DataPipelineParameterObjectsAttributesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := DataPipelineParameterObjectsAttributes{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = DataPipelineParameterObjectsAttributesList{item}
		return nil
	}
	list := []DataPipelineParameterObjectsAttributes{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = DataPipelineParameterObjectsAttributesList(list)
		return nil
	}
	return err
}

// DataPipelinePipelineParameterValues represents AWS Data Pipeline Pipeline ParameterValues
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-datapipeline-pipeline-parametervalues.html
type DataPipelinePipelineParameterValues struct {
	// The ID of a parameter object.
	Id *StringExpr `json:"Id,omitempty"`

	// A value to associate with the parameter object.
	StringValue *StringExpr `json:"StringValue,omitempty"`
}

// DataPipelinePipelineParameterValuesList represents a list of DataPipelinePipelineParameterValues
type DataPipelinePipelineParameterValuesList []DataPipelinePipelineParameterValues

// UnmarshalJSON sets the object from the provided JSON representation
func (l *DataPipelinePipelineParameterValuesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := DataPipelinePipelineParameterValues{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = DataPipelinePipelineParameterValuesList{item}
		return nil
	}
	list := []DataPipelinePipelineParameterValues{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = DataPipelinePipelineParameterValuesList(list)
		return nil
	}
	return err
}

// DataPipelinePipelineObjects represents AWS Data Pipeline PipelineObjects
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-datapipeline-pipeline-pipelineobjects.html
type DataPipelinePipelineObjects struct {
	// Key-value pairs that define the properties of the object.
	Fields *DataPipelineDataPipelineObjectFieldsList `json:"Fields,omitempty"`

	// Identifier of the object.
	Id *StringExpr `json:"Id,omitempty"`

	// Name of the object.
	Name *StringExpr `json:"Name,omitempty"`
}

// DataPipelinePipelineObjectsList represents a list of DataPipelinePipelineObjects
type DataPipelinePipelineObjectsList []DataPipelinePipelineObjects

// UnmarshalJSON sets the object from the provided JSON representation
func (l *DataPipelinePipelineObjectsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := DataPipelinePipelineObjects{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = DataPipelinePipelineObjectsList{item}
		return nil
	}
	list := []DataPipelinePipelineObjects{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = DataPipelinePipelineObjectsList(list)
		return nil
	}
	return err
}

// DataPipelineDataPipelineObjectFields represents AWS Data Pipeline Data Pipeline Object Fields
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-datapipeline-pipeline-pipelineobjects-fields.html
type DataPipelineDataPipelineObjectFields struct {
	// Specifies the name of a field for a particular object. To view fields
	// for a data pipeline object, see Pipeline Object Reference in the AWS
	// Data Pipeline Developer Guide.
	Key *StringExpr `json:"Key,omitempty"`

	// A field value that you specify as an identifier of another object in
	// the same pipeline definition.
	RefValue *StringExpr `json:"RefValue,omitempty"`

	// A field value that you specify as a string. To view valid values for a
	// particular field, see Pipeline Object Reference in the AWS Data
	// Pipeline Developer Guide.
	StringValue *StringExpr `json:"StringValue,omitempty"`
}

// DataPipelineDataPipelineObjectFieldsList represents a list of DataPipelineDataPipelineObjectFields
type DataPipelineDataPipelineObjectFieldsList []DataPipelineDataPipelineObjectFields

// UnmarshalJSON sets the object from the provided JSON representation
func (l *DataPipelineDataPipelineObjectFieldsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := DataPipelineDataPipelineObjectFields{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = DataPipelineDataPipelineObjectFieldsList{item}
		return nil
	}
	list := []DataPipelineDataPipelineObjectFields{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = DataPipelineDataPipelineObjectFieldsList(list)
		return nil
	}
	return err
}

// DataPipelinePipelinePipelineTags represents AWS Data Pipeline Pipeline PipelineTags
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-datapipeline-pipeline-pipelinetags.html
type DataPipelinePipelinePipelineTags struct {
	// The key name of a tag.
	Key *StringExpr `json:"Key,omitempty"`

	// The value to associate with the key name.
	Value *StringExpr `json:"Value,omitempty"`
}

// DataPipelinePipelinePipelineTagsList represents a list of DataPipelinePipelinePipelineTags
type DataPipelinePipelinePipelineTagsList []DataPipelinePipelinePipelineTags

// UnmarshalJSON sets the object from the provided JSON representation
func (l *DataPipelinePipelinePipelineTagsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := DataPipelinePipelinePipelineTags{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = DataPipelinePipelinePipelineTagsList{item}
		return nil
	}
	list := []DataPipelinePipelinePipelineTags{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = DataPipelinePipelinePipelineTagsList(list)
		return nil
	}
	return err
}

// DirectoryServiceMicrosoftADVpcSettings represents AWS Directory Service MicrosoftAD VpcSettings
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-directoryservice-microsoftad-vpcsettings.html
type DirectoryServiceMicrosoftADVpcSettings struct {
	// A list of two subnet IDs for the directory servers. Each subnet must
	// be in different Availability Zones (AZs). AWS Directory Service
	// creates a directory server and a DNS server in each subnet.
	SubnetIds *StringListExpr `json:"SubnetIds,omitempty"`

	// The VPC ID in which to create the Microsoft Active Directory server.
	VpcId *StringExpr `json:"VpcId,omitempty"`
}

// DirectoryServiceMicrosoftADVpcSettingsList represents a list of DirectoryServiceMicrosoftADVpcSettings
type DirectoryServiceMicrosoftADVpcSettingsList []DirectoryServiceMicrosoftADVpcSettings

// UnmarshalJSON sets the object from the provided JSON representation
func (l *DirectoryServiceMicrosoftADVpcSettingsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := DirectoryServiceMicrosoftADVpcSettings{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = DirectoryServiceMicrosoftADVpcSettingsList{item}
		return nil
	}
	list := []DirectoryServiceMicrosoftADVpcSettings{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = DirectoryServiceMicrosoftADVpcSettingsList(list)
		return nil
	}
	return err
}

// DirectoryServiceSimpleADVpcSettings represents AWS Directory Service SimpleAD VpcSettings
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-directoryservice-simplead-vpcsettings.html
type DirectoryServiceSimpleADVpcSettings struct {
	// A list of two subnet IDs for the directory servers. Each subnet must
	// be in different Availability Zones (AZ). AWS Directory Service creates
	// a directory server and a DNS server in each subnet.
	SubnetIds *StringListExpr `json:"SubnetIds,omitempty"`

	// The VPC ID in which to create the Simple AD directory.
	VpcId *StringExpr `json:"VpcId,omitempty"`
}

// DirectoryServiceSimpleADVpcSettingsList represents a list of DirectoryServiceSimpleADVpcSettings
type DirectoryServiceSimpleADVpcSettingsList []DirectoryServiceSimpleADVpcSettings

// UnmarshalJSON sets the object from the provided JSON representation
func (l *DirectoryServiceSimpleADVpcSettingsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := DirectoryServiceSimpleADVpcSettings{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = DirectoryServiceSimpleADVpcSettingsList{item}
		return nil
	}
	list := []DirectoryServiceSimpleADVpcSettings{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = DirectoryServiceSimpleADVpcSettingsList(list)
		return nil
	}
	return err
}

// DynamoDBAttributeDefinitions represents DynamoDB Attribute Definitions
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-dynamodb-attributedef.html
type DynamoDBAttributeDefinitions struct {
	// The name of an attribute. Attribute names can be 1 – 255 characters
	// long and have no character restrictions.
	AttributeName *StringExpr `json:"AttributeName,omitempty"`

	// The data type for the attribute. You can specify S for string data, N
	// for numeric data, or B for binary data.
	AttributeType *StringExpr `json:"AttributeType,omitempty"`
}

// DynamoDBAttributeDefinitionsList represents a list of DynamoDBAttributeDefinitions
type DynamoDBAttributeDefinitionsList []DynamoDBAttributeDefinitions

// UnmarshalJSON sets the object from the provided JSON representation
func (l *DynamoDBAttributeDefinitionsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := DynamoDBAttributeDefinitions{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = DynamoDBAttributeDefinitionsList{item}
		return nil
	}
	list := []DynamoDBAttributeDefinitions{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = DynamoDBAttributeDefinitionsList(list)
		return nil
	}
	return err
}

// DynamoDBGlobalSecondaryIndexes represents DynamoDB Global Secondary Indexes
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-dynamodb-gsi.html
type DynamoDBGlobalSecondaryIndexes struct {
	// The name of the global secondary index. The index name can be 3 –
	// 255 characters long and have no character restrictions.
	IndexName *StringExpr `json:"IndexName,omitempty"`

	// The complete index key schema for the global secondary index, which
	// consists of one or more pairs of attribute names and key types.
	KeySchema *DynamoDBKeySchemaList `json:"KeySchema,omitempty"`

	// Attributes that are copied (projected) from the source table into the
	// index. These attributes are in addition to the primary key attributes
	// and index key attributes, which are automatically projected.
	Projection *DynamoDBProjectionObject `json:"Projection,omitempty"`

	// The provisioned throughput settings for the index.
	ProvisionedThroughput *DynamoDBProvisionedThroughput `json:"ProvisionedThroughput,omitempty"`
}

// DynamoDBGlobalSecondaryIndexesList represents a list of DynamoDBGlobalSecondaryIndexes
type DynamoDBGlobalSecondaryIndexesList []DynamoDBGlobalSecondaryIndexes

// UnmarshalJSON sets the object from the provided JSON representation
func (l *DynamoDBGlobalSecondaryIndexesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := DynamoDBGlobalSecondaryIndexes{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = DynamoDBGlobalSecondaryIndexesList{item}
		return nil
	}
	list := []DynamoDBGlobalSecondaryIndexes{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = DynamoDBGlobalSecondaryIndexesList(list)
		return nil
	}
	return err
}

// DynamoDBKeySchema represents DynamoDB Key Schema
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-dynamodb-keyschema.html
type DynamoDBKeySchema struct {
	// The attribute name that is used as the primary key for this table.
	// Primary key element names can be 1 – 255 characters long and have no
	// character restrictions.
	AttributeName *StringExpr `json:"AttributeName,omitempty"`

	// Represents the attribute data, consisting of the data type and the
	// attribute value itself. You can specify HASH or RANGE.
	KeyType *StringExpr `json:"KeyType,omitempty"`
}

// DynamoDBKeySchemaList represents a list of DynamoDBKeySchema
type DynamoDBKeySchemaList []DynamoDBKeySchema

// UnmarshalJSON sets the object from the provided JSON representation
func (l *DynamoDBKeySchemaList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := DynamoDBKeySchema{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = DynamoDBKeySchemaList{item}
		return nil
	}
	list := []DynamoDBKeySchema{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = DynamoDBKeySchemaList(list)
		return nil
	}
	return err
}

// DynamoDBLocalSecondaryIndexes represents DynamoDB Local Secondary Indexes
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-dynamodb-lsi.html
type DynamoDBLocalSecondaryIndexes struct {
	// The name of the local secondary index. The index name can be 3 – 255
	// characters long and have no character restrictions.
	IndexName *StringExpr `json:"IndexName,omitempty"`

	// The complete index key schema for the local secondary index, which
	// consists of one or more pairs of attribute names and key types. For
	// local secondary indexes, the hash key must be the same as that of the
	// source table.
	KeySchema *DynamoDBKeySchemaList `json:"KeySchema,omitempty"`

	// Attributes that are copied (projected) from the source table into the
	// index. These attributes are additions to the primary key attributes
	// and index key attributes, which are automatically projected.
	Projection *DynamoDBProjectionObject `json:"Projection,omitempty"`
}

// DynamoDBLocalSecondaryIndexesList represents a list of DynamoDBLocalSecondaryIndexes
type DynamoDBLocalSecondaryIndexesList []DynamoDBLocalSecondaryIndexes

// UnmarshalJSON sets the object from the provided JSON representation
func (l *DynamoDBLocalSecondaryIndexesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := DynamoDBLocalSecondaryIndexes{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = DynamoDBLocalSecondaryIndexesList{item}
		return nil
	}
	list := []DynamoDBLocalSecondaryIndexes{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = DynamoDBLocalSecondaryIndexesList(list)
		return nil
	}
	return err
}

// DynamoDBProjectionObject represents DynamoDB Projection Object
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-dynamodb-projectionobject.html
type DynamoDBProjectionObject struct {
	// The non-key attribute names that are projected into the index.
	NonKeyAttributes *StringListExpr `json:"NonKeyAttributes,omitempty"`

	// The set of attributes that are projected into the index:
	ProjectionType *StringExpr `json:"ProjectionType,omitempty"`

	// Only the index and primary keys are projected into the index.
	KEYSXONLY interface{} `json:"KEYS_ONLY,omitempty"`

	// Only the specified table attributes are projected into the index. The
	// list of projected attributes are in NonKeyAttributes.
	INCLUDE interface{} `json:"INCLUDE,omitempty"`

	// All of the table attributes are projected into the index.
	ALL interface{} `json:"ALL,omitempty"`
}

// DynamoDBProjectionObjectList represents a list of DynamoDBProjectionObject
type DynamoDBProjectionObjectList []DynamoDBProjectionObject

// UnmarshalJSON sets the object from the provided JSON representation
func (l *DynamoDBProjectionObjectList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := DynamoDBProjectionObject{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = DynamoDBProjectionObjectList{item}
		return nil
	}
	list := []DynamoDBProjectionObject{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = DynamoDBProjectionObjectList(list)
		return nil
	}
	return err
}

// DynamoDBProvisionedThroughput represents DynamoDB Provisioned Throughput
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-dynamodb-provisionedthroughput.html
type DynamoDBProvisionedThroughput struct {
	// Sets the desired minimum number of consistent reads of items (up to
	// 1KB in size) per second for the specified table before Amazon DynamoDB
	// balances the load.
	ReadCapacityUnits *IntegerExpr `json:"ReadCapacityUnits,omitempty"`

	// Sets the desired minimum number of consistent writes of items (up to
	// 1KB in size) per second for the specified table before Amazon DynamoDB
	// balances the load.
	WriteCapacityUnits *IntegerExpr `json:"WriteCapacityUnits,omitempty"`
}

// DynamoDBProvisionedThroughputList represents a list of DynamoDBProvisionedThroughput
type DynamoDBProvisionedThroughputList []DynamoDBProvisionedThroughput

// UnmarshalJSON sets the object from the provided JSON representation
func (l *DynamoDBProvisionedThroughputList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := DynamoDBProvisionedThroughput{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = DynamoDBProvisionedThroughputList{item}
		return nil
	}
	list := []DynamoDBProvisionedThroughput{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = DynamoDBProvisionedThroughputList(list)
		return nil
	}
	return err
}

// DynamoDBTableStreamSpecification represents DynamoDB Table StreamSpecification
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-dynamodb-streamspecification.html
type DynamoDBTableStreamSpecification struct {
	// Determines the information that the stream captures when an item in
	// the table is modified. For valid values, see StreamSpecification in
	// the Amazon DynamoDB API Reference.
	StreamViewType *StringExpr `json:"StreamViewType,omitempty"`
}

// DynamoDBTableStreamSpecificationList represents a list of DynamoDBTableStreamSpecification
type DynamoDBTableStreamSpecificationList []DynamoDBTableStreamSpecification

// UnmarshalJSON sets the object from the provided JSON representation
func (l *DynamoDBTableStreamSpecificationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := DynamoDBTableStreamSpecification{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = DynamoDBTableStreamSpecificationList{item}
		return nil
	}
	list := []DynamoDBTableStreamSpecification{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = DynamoDBTableStreamSpecificationList(list)
		return nil
	}
	return err
}

// EC2BlockDeviceMappingProperty represents Amazon EC2 Block Device Mapping Property
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-blockdev-mapping.html
type EC2BlockDeviceMappingProperty struct {
	// The name of the device within Amazon EC2.
	DeviceName *StringExpr `json:"DeviceName,omitempty"`

	// Required: Conditional You can specify either VirtualName or Ebs, but
	// not both.
	Ebs *ElasticBlockStoreBlockDeviceProperty `json:"Ebs,omitempty"`

	// This property can be used to unmap a defined device.
	NoDevice interface{} `json:"NoDevice,omitempty"`

	// The name of the virtual device. The name must be in the form
	// ephemeralX where X is a number starting from zero (0); for example,
	// ephemeral0.
	VirtualName *StringExpr `json:"VirtualName,omitempty"`
}

// EC2BlockDeviceMappingPropertyList represents a list of EC2BlockDeviceMappingProperty
type EC2BlockDeviceMappingPropertyList []EC2BlockDeviceMappingProperty

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2BlockDeviceMappingPropertyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2BlockDeviceMappingProperty{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2BlockDeviceMappingPropertyList{item}
		return nil
	}
	list := []EC2BlockDeviceMappingProperty{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2BlockDeviceMappingPropertyList(list)
		return nil
	}
	return err
}

// ElasticBlockStoreBlockDeviceProperty represents Amazon Elastic Block Store Block Device Property
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-blockdev-template.html
type ElasticBlockStoreBlockDeviceProperty struct {
	// Determines whether to delete the volume on instance termination. The
	// default value is true.
	DeleteOnTermination *BoolExpr `json:"DeleteOnTermination,omitempty"`

	// Indicates whether the volume is encrypted. Encrypted Amazon EBS
	// volumes can only be attached to instance types that support Amazon EBS
	// encryption. Volumes that are created from encrypted snapshots are
	// automatically encrypted. You cannot create an encrypted volume from an
	// unencrypted snapshot or vice versa. If your AMI uses encrypted
	// volumes, you can only launch the AMI on supported instance types. For
	// more information, see Amazon EBS encryption in the Amazon EC2 User
	// Guide for Linux Instances.
	Encrypted *BoolExpr `json:"Encrypted,omitempty"`

	// The number of I/O operations per second (IOPS) that the volume
	// supports. This can be an integer from 100 – 2000.
	Iops *IntegerExpr `json:"Iops,omitempty"`

	// The snapshot ID of the volume to use to create a block device.
	SnapshotId *StringExpr `json:"SnapshotId,omitempty"`

	// The volume size, in gibibytes (GiB). For valid values, see the Size
	// parameter for the CreateVolume action in the Amazon EC2 API Reference.
	VolumeSize *StringExpr `json:"VolumeSize,omitempty"`

	// The volume type. If you set the type to io1, you must also set the
	// Iops property. For valid values, see the VolumeType parameter for the
	// CreateVolume action in the Amazon EC2 API Reference.
	VolumeType *StringExpr `json:"VolumeType,omitempty"`
}

// ElasticBlockStoreBlockDevicePropertyList represents a list of ElasticBlockStoreBlockDeviceProperty
type ElasticBlockStoreBlockDevicePropertyList []ElasticBlockStoreBlockDeviceProperty

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticBlockStoreBlockDevicePropertyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticBlockStoreBlockDeviceProperty{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticBlockStoreBlockDevicePropertyList{item}
		return nil
	}
	list := []ElasticBlockStoreBlockDeviceProperty{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticBlockStoreBlockDevicePropertyList(list)
		return nil
	}
	return err
}

// EC2InstanceSsmAssociations represents Amazon EC2 Instance SsmAssociations
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-instance-ssmassociations.html
type EC2InstanceSsmAssociations struct {
	// The input parameter values to use with the associated SSM document.
	AssociationParameters *EC2InstanceSsmAssociationsAssociationParametersList `json:"AssociationParameters,omitempty"`

	// The name of an SSM document to associate with the instance.
	DocumentName *StringExpr `json:"DocumentName,omitempty"`
}

// EC2InstanceSsmAssociationsList represents a list of EC2InstanceSsmAssociations
type EC2InstanceSsmAssociationsList []EC2InstanceSsmAssociations

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2InstanceSsmAssociationsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2InstanceSsmAssociations{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2InstanceSsmAssociationsList{item}
		return nil
	}
	list := []EC2InstanceSsmAssociations{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2InstanceSsmAssociationsList(list)
		return nil
	}
	return err
}

// EC2InstanceSsmAssociationsAssociationParameters represents Amazon EC2 Instance SsmAssociations AssociationParameters
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-instance-ssmassociations-associationparameters.html
type EC2InstanceSsmAssociationsAssociationParameters struct {
	// The name of an input parameter that is in the associated SSM document.
	Key *StringExpr `json:"Key,omitempty"`

	// The value of an input parameter.
	Value *StringListExpr `json:"Value,omitempty"`
}

// EC2InstanceSsmAssociationsAssociationParametersList represents a list of EC2InstanceSsmAssociationsAssociationParameters
type EC2InstanceSsmAssociationsAssociationParametersList []EC2InstanceSsmAssociationsAssociationParameters

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2InstanceSsmAssociationsAssociationParametersList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2InstanceSsmAssociationsAssociationParameters{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2InstanceSsmAssociationsAssociationParametersList{item}
		return nil
	}
	list := []EC2InstanceSsmAssociationsAssociationParameters{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2InstanceSsmAssociationsAssociationParametersList(list)
		return nil
	}
	return err
}

// EC2MountPoint represents EC2 MountPoint Property Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-mount-point.html
type EC2MountPoint struct {
	// How the device is exposed to the instance (such as /dev/sdh, or xvdh).
	Device *StringExpr `json:"Device,omitempty"`

	// The ID of the Amazon EBS volume. The volume and instance must be
	// within the same Availability Zone and the instance must be running.
	VolumeId *StringExpr `json:"VolumeId,omitempty"`
}

// EC2MountPointList represents a list of EC2MountPoint
type EC2MountPointList []EC2MountPoint

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2MountPointList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2MountPoint{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2MountPointList{item}
		return nil
	}
	list := []EC2MountPoint{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2MountPointList(list)
		return nil
	}
	return err
}

// EC2NetworkInterfaceEmbedded represents EC2 NetworkInterface Embedded Property Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-network-iface-embedded.html
type EC2NetworkInterfaceEmbedded struct {
	// Indicates whether the network interface receives a public IP address.
	// You can associate a public IP address with a network interface only if
	// it has a device index of eth0 and if it is a new network interface
	// (not an existing one). In other words, if you specify true, don't
	// specify a network interface ID. For more information, see Amazon EC2
	// Instance IP Addressing.
	AssociatePublicIpAddress *BoolExpr `json:"AssociatePublicIpAddress,omitempty"`

	// Whether to delete the network interface when the instance terminates.
	DeleteOnTermination *BoolExpr `json:"DeleteOnTermination,omitempty"`

	// The description of this network interface.
	Description *StringExpr `json:"Description,omitempty"`

	// The network interface's position in the attachment order.
	DeviceIndex *StringExpr `json:"DeviceIndex,omitempty"`

	// A list of security group IDs associated with this network interface.
	GroupSet *StringListExpr `json:"GroupSet,omitempty"`

	// An existing network interface ID.
	NetworkInterfaceId *StringExpr `json:"NetworkInterfaceId,omitempty"`

	// Assigns a single private IP address to the network interface, which is
	// used as the primary private IP address. If you want to specify
	// multiple private IP address, use the PrivateIpAddresses property.
	PrivateIpAddress *StringExpr `json:"PrivateIpAddress,omitempty"`

	// Assigns a list of private IP addresses to the network interface. You
	// can specify a primary private IP address by setting the value of the
	// Primary property to true in the PrivateIpAddressSpecification
	// property. If you want Amazon EC2 to automatically assign private IP
	// addresses, use the SecondaryPrivateIpCount property and do not specify
	// this property.
	PrivateIpAddresses *EC2NetworkInterfacePrivateIPSpecificationList `json:"PrivateIpAddresses,omitempty"`

	// The number of secondary private IP addresses that Amazon EC2 auto
	// assigns to the network interface. Amazon EC2 uses the value of the
	// PrivateIpAddress property as the primary private IP address. If you
	// don't specify that property, Amazon EC2 auto assigns both the primary
	// and secondary private IP addresses.
	SecondaryPrivateIpAddressCount *IntegerExpr `json:"SecondaryPrivateIpAddressCount,omitempty"`

	// The ID of the subnet to associate with the network interface.
	SubnetId *StringExpr `json:"SubnetId,omitempty"`
}

// EC2NetworkInterfaceEmbeddedList represents a list of EC2NetworkInterfaceEmbedded
type EC2NetworkInterfaceEmbeddedList []EC2NetworkInterfaceEmbedded

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2NetworkInterfaceEmbeddedList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2NetworkInterfaceEmbedded{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2NetworkInterfaceEmbeddedList{item}
		return nil
	}
	list := []EC2NetworkInterfaceEmbedded{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2NetworkInterfaceEmbeddedList(list)
		return nil
	}
	return err
}

// EC2NetworkAclEntryIcmp represents EC2 NetworkAclEntry Icmp
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-networkaclentry-icmp.html
type EC2NetworkAclEntryIcmp struct {
	// The Internet Control Message Protocol (ICMP) code. You can use -1 to
	// specify all ICMP codes for the given ICMP type.
	Code *IntegerExpr `json:"Code,omitempty"`

	// The Internet Control Message Protocol (ICMP) type. You can use -1 to
	// specify all ICMP types.
	Type *IntegerExpr `json:"Type,omitempty"`
}

// EC2NetworkAclEntryIcmpList represents a list of EC2NetworkAclEntryIcmp
type EC2NetworkAclEntryIcmpList []EC2NetworkAclEntryIcmp

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2NetworkAclEntryIcmpList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2NetworkAclEntryIcmp{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2NetworkAclEntryIcmpList{item}
		return nil
	}
	list := []EC2NetworkAclEntryIcmp{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2NetworkAclEntryIcmpList(list)
		return nil
	}
	return err
}

// EC2NetworkAclEntryPortRange represents EC2 NetworkAclEntry PortRange
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-networkaclentry-portrange.html
type EC2NetworkAclEntryPortRange struct {
	// The first port in the range.
	From *IntegerExpr `json:"From,omitempty"`

	// The last port in the range.
	To *IntegerExpr `json:"To,omitempty"`
}

// EC2NetworkAclEntryPortRangeList represents a list of EC2NetworkAclEntryPortRange
type EC2NetworkAclEntryPortRangeList []EC2NetworkAclEntryPortRange

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2NetworkAclEntryPortRangeList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2NetworkAclEntryPortRange{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2NetworkAclEntryPortRangeList{item}
		return nil
	}
	list := []EC2NetworkAclEntryPortRange{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2NetworkAclEntryPortRangeList(list)
		return nil
	}
	return err
}

// EC2NetworkInterfacePrivateIPSpecification represents EC2 Network Interface Private IP Specification
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-network-interface-privateipspec.html
type EC2NetworkInterfacePrivateIPSpecification struct {
	// The private IP address of the network interface.
	PrivateIpAddress *StringExpr `json:"PrivateIpAddress,omitempty"`

	// Sets the private IP address as the primary private address. You can
	// set only one primary private IP address. If you don't specify a
	// primary private IP address, Amazon EC2 automatically assigns a primary
	// private IP address.
	Primary *BoolExpr `json:"Primary,omitempty"`
}

// EC2NetworkInterfacePrivateIPSpecificationList represents a list of EC2NetworkInterfacePrivateIPSpecification
type EC2NetworkInterfacePrivateIPSpecificationList []EC2NetworkInterfacePrivateIPSpecification

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2NetworkInterfacePrivateIPSpecificationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2NetworkInterfacePrivateIPSpecification{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2NetworkInterfacePrivateIPSpecificationList{item}
		return nil
	}
	list := []EC2NetworkInterfacePrivateIPSpecification{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2NetworkInterfacePrivateIPSpecificationList(list)
		return nil
	}
	return err
}

// EC2SecurityGroupRule represents EC2 Security Group Rule Property Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-security-group-rule.html
type EC2SecurityGroupRule struct {
	// Specifies a CIDR range.
	CidrIp *StringExpr `json:"CidrIp,omitempty"`

	// The AWS service prefix of an Amazon VPC endpoint. For more
	// information, see VPC Endpoints in the Amazon VPC User Guide.
	DestinationPrefixListIdXXSecurityGroupEgressXOnlyX *StringExpr `json:"DestinationPrefixListId (SecurityGroupEgress only),omitempty"`

	// Specifies the GroupId of the destination Amazon VPC security group.
	DestinationSecurityGroupIdXXSecurityGroupEgressXOnlyX *StringExpr `json:"DestinationSecurityGroupId (SecurityGroupEgress only),omitempty"`

	// The start of port range for the TCP and UDP protocols, or an ICMP type
	// number. An ICMP type number of -1 indicates a wildcard (i.e., any ICMP
	// type number).
	FromPort *IntegerExpr `json:"FromPort,omitempty"`

	// An IP protocol name or number. For valid values, go to the IpProtocol
	// parameter in AuthorizeSecurityGroupIngress
	IpProtocol *StringExpr `json:"IpProtocol,omitempty"`

	// For VPC security groups only. Specifies the ID of the Amazon EC2
	// Security Group to allow access. You can use the Ref intrinsic function
	// to refer to the logical ID of a security group defined in the same
	// template.
	SourceSecurityGroupIdXXSecurityGroupIngressXOnlyX *StringExpr `json:"SourceSecurityGroupId (SecurityGroupIngress only),omitempty"`

	// For non-VPC security groups only. Specifies the name of the Amazon EC2
	// Security Group to use for access. You can use the Ref intrinsic
	// function to refer to the logical name of a security group that is
	// defined in the same template.
	SourceSecurityGroupNameXXSecurityGroupIngressXOnlyX *StringExpr `json:"SourceSecurityGroupName (SecurityGroupIngress only),omitempty"`

	// Specifies the AWS Account ID of the owner of the Amazon EC2 Security
	// Group that is specified in the SourceSecurityGroupName property.
	SourceSecurityGroupOwnerIdXXSecurityGroupIngressXOnlyX *StringExpr `json:"SourceSecurityGroupOwnerId (SecurityGroupIngress only),omitempty"`

	// The end of port range for the TCP and UDP protocols, or an ICMP code.
	// An ICMP code of -1 indicates a wildcard (i.e., any ICMP code).
	ToPort *IntegerExpr `json:"ToPort,omitempty"`
}

// EC2SecurityGroupRuleList represents a list of EC2SecurityGroupRule
type EC2SecurityGroupRuleList []EC2SecurityGroupRule

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2SecurityGroupRuleList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2SecurityGroupRule{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2SecurityGroupRuleList{item}
		return nil
	}
	list := []EC2SecurityGroupRule{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2SecurityGroupRuleList(list)
		return nil
	}
	return err
}

// EC2SpotFleetSpotFleetRequestConfigData represents Amazon EC2 SpotFleet SpotFleetRequestConfigData
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-spotfleet-spotfleetrequestconfigdata.html
type EC2SpotFleetSpotFleetRequestConfigData struct {
	// Indicates how to allocate the target capacity across the Spot pools
	// that you specified in the Spot fleet request. For valid values, see
	// SpotFleetRequestConfigData in the Amazon EC2 API Reference.
	AllocationStrategy *StringExpr `json:"AllocationStrategy,omitempty"`

	// Indicates whether running Spot instances are terminated if you
	// decrease the target capacity of the Spot fleet request below the
	// current size of the Spot fleet. For valid values, see
	// SpotFleetRequestConfigData in the Amazon EC2 API Reference.
	ExcessCapacityTerminationPolicy *StringExpr `json:"ExcessCapacityTerminationPolicy,omitempty"`

	// The Amazon Resource Name (ARN) of an AWS Identity and Access
	// Management (IAM) role that grants the Spot fleet the ability to bid
	// on, launch, and terminate instances on your behalf. For more
	// information, see Spot Fleet Prerequisites in the Amazon EC2 User Guide
	// for Linux Instances.
	IamFleetRole *StringExpr `json:"IamFleetRole,omitempty"`

	// The launch specifications for the Spot fleet request.
	LaunchSpecifications *ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsList `json:"LaunchSpecifications,omitempty"`

	// The bid price per unit hour. For more information, see How Spot Fleet
	// Works in the Amazon EC2 User Guide for Linux Instances.
	SpotPrice *StringExpr `json:"SpotPrice,omitempty"`

	// The number of units to request for the spot fleet. You can choose to
	// set the target capacity as the number of instances or as a performance
	// characteristic that is important to your application workload, such as
	// vCPUs, memory, or I/O. For more information, see How Spot Fleet Works
	// in the Amazon EC2 User Guide for Linux Instances.
	TargetCapacity *IntegerExpr `json:"TargetCapacity,omitempty"`

	// Indicates whether running Spot instances are terminated when the Spot
	// fleet request expires.
	TerminateInstancesWithExpiration *BoolExpr `json:"TerminateInstancesWithExpiration,omitempty"`

	// The start date and time of the request, in UTC format
	// (YYYY-MM-DDTHH:MM:SSZ). By default, Amazon Elastic Compute Cloud
	// (Amazon EC2 ) starts fulfilling the request immediately.
	ValidFrom *StringExpr `json:"ValidFrom,omitempty"`

	// The end date and time of the request, in UTC format
	// (YYYY-MM-DDTHH:MM:SSZ). After the end date and time, Amazon EC2
	// doesn't request new Spot instances or enable them to fulfill the
	// request.
	ValidUntil *StringExpr `json:"ValidUntil,omitempty"`
}

// EC2SpotFleetSpotFleetRequestConfigDataList represents a list of EC2SpotFleetSpotFleetRequestConfigData
type EC2SpotFleetSpotFleetRequestConfigDataList []EC2SpotFleetSpotFleetRequestConfigData

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2SpotFleetSpotFleetRequestConfigDataList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2SpotFleetSpotFleetRequestConfigData{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2SpotFleetSpotFleetRequestConfigDataList{item}
		return nil
	}
	list := []EC2SpotFleetSpotFleetRequestConfigData{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2SpotFleetSpotFleetRequestConfigDataList(list)
		return nil
	}
	return err
}

// ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecifications represents Amazon Elastic Compute Cloud SpotFleet SpotFleetRequestConfigData LaunchSpecifications
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-spotfleet-spotfleetrequestconfigdata-launchspecifications.html
type ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecifications struct {
	// Defines the block devices that are mapped to the Spot instances.
	BlockDeviceMappings *ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappingsList `json:"BlockDeviceMappings,omitempty"`

	// Indicates whether the instances are optimized for Amazon Elastic Block
	// Store (Amazon EBS) I/O. This optimization provides dedicated
	// throughput to Amazon EBS and an optimized configuration stack to
	// provide optimal EBS I/O performance. This optimization isn't available
	// with all instance types. Additional usage charges apply when you use
	// an Amazon EBS-optimized instance.
	EbsOptimized *BoolExpr `json:"EbsOptimized,omitempty"`

	// Defines the AWS Identity and Access Management (IAM) instance profile
	// to associate with the instances.
	IamInstanceProfile *ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsIamInstanceProfile `json:"IamInstanceProfile,omitempty"`

	// The unique ID of the Amazon Machine Image (AMI) to launch on the
	// instances.
	ImageId *StringExpr `json:"ImageId,omitempty"`

	// Specifies the instance type of the EC2 instances.
	InstanceType *StringExpr `json:"InstanceType,omitempty"`

	// The ID of the kernel that is associated with the Amazon Elastic
	// Compute Cloud (Amazon EC2) AMI.
	KernelId *StringExpr `json:"KernelId,omitempty"`

	// An Amazon EC2 key pair to associate with the instances.
	KeyName *StringExpr `json:"KeyName,omitempty"`

	// Enable or disable monitoring for the instances.
	Monitoring *EC2SpotFleetSpotFleetRequestConfigDataLaunchSpecificationsMonitoring `json:"Monitoring,omitempty"`

	// The network interfaces to associate with the instances.
	NetworkInterfaces *ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfacesList `json:"NetworkInterfaces,omitempty"`

	// Defines a placement group, which is a logical grouping of instances
	// within a single Availability Zone (AZ).
	Placement *ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsPlacement `json:"Placement,omitempty"`

	// The ID of the RAM disk to select. Some kernels require additional
	// drivers at launch. Check the kernel requirements for information about
	// whether you need to specify a RAM disk. To find kernel requirements,
	// refer to the AWS Resource Center and search for the kernel ID.
	RamdiskId *StringExpr `json:"RamdiskId,omitempty"`

	// One or more security group IDs to associate with the instances.
	SecurityGroups *ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsSecurityGroupsList `json:"SecurityGroups,omitempty"`

	// The bid price per unit hour for the specified instance type. If you
	// don't specify a value, Amazon EC2 uses the Spot bid price for the
	// fleet. For more information, see How Spot Fleet Works in the Amazon
	// EC2 User Guide for Linux Instances.
	SpotPrice *StringExpr `json:"SpotPrice,omitempty"`

	// The ID of the subnet in which to launch the instances.
	SubnetId *StringExpr `json:"SubnetId,omitempty"`

	// Base64-encoded MIME user data that instances use when starting up.
	UserData *StringExpr `json:"UserData,omitempty"`

	// The number of units provided by the specified instance type. These
	// units are the same units that you chose to set the target capacity in
	// terms of instances or a performance characteristic, such as vCPUs,
	// memory, or I/O. For more information, see How Spot Fleet Works in the
	// Amazon EC2 User Guide for Linux Instances.
	WeightedCapacity *IntegerExpr `json:"WeightedCapacity,omitempty"`
}

// ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsList represents a list of ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecifications
type ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsList []ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecifications

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecifications{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsList{item}
		return nil
	}
	list := []ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecifications{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsList(list)
		return nil
	}
	return err
}

// ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappings represents Amazon Elastic Compute Cloud SpotFleet SpotFleetRequestConfigData LaunchSpecifications BlockDeviceMappings
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-spotfleet-spotfleetrequestconfigdata-launchspecifications-blockdevicemappings.html
type ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappings struct {
	// The name of the device within the EC2 instance, such as /dev/dsh or
	// xvdh.
	DeviceName *StringExpr `json:"DeviceName,omitempty"`

	// The Amazon Elastic Block Store (Amazon EBS) volume information.
	Ebs *ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappingsEbs `json:"Ebs,omitempty"`

	// Suppresses the specified device that is included in the block device
	// mapping of the Amazon Machine Image (AMI).
	NoDevice *BoolExpr `json:"NoDevice,omitempty"`

	// The name of the virtual device. The name must be in the form
	// ephemeralX where X is a number equal to or greater than zero (0), for
	// example, ephemeral0.
	VirtualName *StringExpr `json:"VirtualName,omitempty"`
}

// ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappingsList represents a list of ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappings
type ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappingsList []ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappings

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappingsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappings{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappingsList{item}
		return nil
	}
	list := []ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappings{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappingsList(list)
		return nil
	}
	return err
}

// ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappingsEbs represents Amazon Elastic Compute Cloud SpotFleet SpotFleetRequestConfigData LaunchSpecifications BlockDeviceMappings Ebs
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-spotfleet-spotfleetrequestconfigdata-launchspecifications-blockdevicemappings-ebs.html
type ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappingsEbs struct {
	// Indicates whether to delete the volume when the instance is
	// terminated.
	DeleteOnTermination *BoolExpr `json:"DeleteOnTermination,omitempty"`

	// Indicates whether the EBS volume is encrypted. Encrypted Amazon EBS
	// volumes can be attached only to instances that support Amazon EBS
	// encryption.
	Encrypted *BoolExpr `json:"Encrypted,omitempty"`

	// The number of I/O operations per second (IOPS) that the volume
	// supports. For more information, see Iops for the EbsBlockDevice action
	// in the Amazon EC2 API Reference.
	Iops *IntegerExpr `json:"Iops,omitempty"`

	// The snapshot ID of the volume that you want to use. If you specify
	// both the SnapshotId and VolumeSize properties, VolumeSize must be
	// equal to or greater than the size of the snapshot.
	SnapshotId *StringExpr `json:"SnapshotId,omitempty"`

	// The volume size, in Gibibytes (GiB). If you specify both the
	// SnapshotId and VolumeSize properties, VolumeSize must be equal to or
	// greater than the size of the snapshot. For more information about
	// specifying the volume size, see VolumeSize for the EbsBlockDevice
	// action in the Amazon EC2 API Reference.
	VolumeSize *IntegerExpr `json:"VolumeSize,omitempty"`

	// The volume type. For more information about specifying the volume
	// type, see VolumeType for the EbsBlockDevice action in the Amazon EC2
	// API Reference.
	VolumeType *StringExpr `json:"VolumeType,omitempty"`
}

// ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappingsEbsList represents a list of ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappingsEbs
type ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappingsEbsList []ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappingsEbs

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappingsEbsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappingsEbs{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappingsEbsList{item}
		return nil
	}
	list := []ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappingsEbs{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsBlockDeviceMappingsEbsList(list)
		return nil
	}
	return err
}

// ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsIamInstanceProfile represents Amazon Elastic Compute Cloud SpotFleet SpotFleetRequestConfigData LaunchSpecifications IamInstanceProfile
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-spotfleet-spotfleetrequestconfigdata-launchspecifications-iaminstanceprofile.html
type ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsIamInstanceProfile struct {
	// The Amazon Resource Name (ARN) of the instance profile to associate
	// with the instances. The instance profile contains the IAM role that is
	// associated with the instances.
	Arn *StringExpr `json:"Arn,omitempty"`
}

// ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsIamInstanceProfileList represents a list of ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsIamInstanceProfile
type ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsIamInstanceProfileList []ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsIamInstanceProfile

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsIamInstanceProfileList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsIamInstanceProfile{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsIamInstanceProfileList{item}
		return nil
	}
	list := []ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsIamInstanceProfile{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsIamInstanceProfileList(list)
		return nil
	}
	return err
}

// EC2SpotFleetSpotFleetRequestConfigDataLaunchSpecificationsMonitoring represents Amazon EC2 SpotFleet SpotFleetRequestConfigData LaunchSpecifications Monitoring
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-spotfleet-spotfleetrequestconfigdata-launchspecifications-monitoring.html
type EC2SpotFleetSpotFleetRequestConfigDataLaunchSpecificationsMonitoring struct {
	// Indicates whether monitoring is enabled for the instances.
	Enabled *BoolExpr `json:"Enabled,omitempty"`
}

// EC2SpotFleetSpotFleetRequestConfigDataLaunchSpecificationsMonitoringList represents a list of EC2SpotFleetSpotFleetRequestConfigDataLaunchSpecificationsMonitoring
type EC2SpotFleetSpotFleetRequestConfigDataLaunchSpecificationsMonitoringList []EC2SpotFleetSpotFleetRequestConfigDataLaunchSpecificationsMonitoring

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2SpotFleetSpotFleetRequestConfigDataLaunchSpecificationsMonitoringList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2SpotFleetSpotFleetRequestConfigDataLaunchSpecificationsMonitoring{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2SpotFleetSpotFleetRequestConfigDataLaunchSpecificationsMonitoringList{item}
		return nil
	}
	list := []EC2SpotFleetSpotFleetRequestConfigDataLaunchSpecificationsMonitoring{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2SpotFleetSpotFleetRequestConfigDataLaunchSpecificationsMonitoringList(list)
		return nil
	}
	return err
}

// ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfaces represents Amazon Elastic Compute Cloud SpotFleet SpotFleetRequestConfigData LaunchSpecifications NetworkInterfaces
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-spotfleet-spotfleetrequestconfigdata-launchspecifications-networkinterfaces.html
type ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfaces struct {
	// Indicates whether to assign a public IP address to an instance that
	// you launch in a VPC. The public IP address can only be assigned to a
	// network interface for eth0, and can only be assigned to a new network
	// interface, not an existing one.
	AssociatePublicIpAddress *BoolExpr `json:"AssociatePublicIpAddress,omitempty"`

	// Indicates whether to delete the network interface when the instance
	// terminates.
	DeleteOnTermination *BoolExpr `json:"DeleteOnTermination,omitempty"`

	// The description of this network interface.
	Description *StringExpr `json:"Description,omitempty"`

	// The network interface's position in the attachment order.
	DeviceIndex *IntegerExpr `json:"DeviceIndex,omitempty"`

	// A list of security group IDs to associate with this network interface.
	Groups *StringListExpr `json:"Groups,omitempty"`

	// A network interface ID.
	NetworkInterfaceId *StringExpr `json:"NetworkInterfaceId,omitempty"`

	// One or more private IP addresses to assign to the network interface.
	// You can designate only one private IP address as primary.
	PrivateIpAddresses *ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfacesPrivateIpAddressesList `json:"PrivateIpAddresses,omitempty"`

	// The number of secondary private IP addresses that Amazon Elastic
	// Compute Cloud (Amazon EC2) automatically assigns to the network
	// interface.
	SecondaryPrivateIpAddressCount *IntegerExpr `json:"SecondaryPrivateIpAddressCount,omitempty"`

	// The ID of the subnet to associate with the network interface.
	SubnetId *StringExpr `json:"SubnetId,omitempty"`
}

// ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfacesList represents a list of ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfaces
type ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfacesList []ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfaces

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfacesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfaces{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfacesList{item}
		return nil
	}
	list := []ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfaces{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfacesList(list)
		return nil
	}
	return err
}

// ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfacesPrivateIpAddresses represents Amazon Elastic Compute Cloud SpotFleet SpotFleetRequestConfigData LaunchSpecifications NetworkInterfaces PrivateIpAddresses
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-spotfleet-spotfleetrequestconfigdata-launchspecifications-networkinterfaces-privateipaddresses.html
type ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfacesPrivateIpAddresses struct {
	// Indicates whether the private IP address is the primary private IP
	// address. You can designate only one IP address as primary.
	Primary *BoolExpr `json:"Primary,omitempty"`

	// The private IP address.
	PrivateIpAddress *StringExpr `json:"PrivateIpAddress,omitempty"`
}

// ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfacesPrivateIpAddressesList represents a list of ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfacesPrivateIpAddresses
type ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfacesPrivateIpAddressesList []ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfacesPrivateIpAddresses

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfacesPrivateIpAddressesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfacesPrivateIpAddresses{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfacesPrivateIpAddressesList{item}
		return nil
	}
	list := []ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfacesPrivateIpAddresses{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsNetworkInterfacesPrivateIpAddressesList(list)
		return nil
	}
	return err
}

// ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsPlacement represents Amazon Elastic Compute Cloud SpotFleet SpotFleetRequestConfigData LaunchSpecifications Placement
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-spotfleet-spotfleetrequestconfigdata-launchspecifications-placement.html
type ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsPlacement struct {
	// The Availability Zone (AZ) of the placement group.
	AvailabilityZone *StringExpr `json:"AvailabilityZone,omitempty"`

	// The name of the placement group (for cluster instances).
	GroupName *StringExpr `json:"GroupName,omitempty"`
}

// ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsPlacementList represents a list of ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsPlacement
type ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsPlacementList []ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsPlacement

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsPlacementList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsPlacement{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsPlacementList{item}
		return nil
	}
	list := []ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsPlacement{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsPlacementList(list)
		return nil
	}
	return err
}

// ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsSecurityGroups represents Amazon Elastic Compute Cloud SpotFleet SpotFleetRequestConfigData LaunchSpecifications SecurityGroups
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-spotfleet-spotfleetrequestconfigdata-launchspecifications-securitygroups.html
type ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsSecurityGroups struct {
	// The ID of a security group.
	GroupId *StringExpr `json:"GroupId,omitempty"`
}

// ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsSecurityGroupsList represents a list of ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsSecurityGroups
type ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsSecurityGroupsList []ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsSecurityGroups

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsSecurityGroupsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsSecurityGroups{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsSecurityGroupsList{item}
		return nil
	}
	list := []ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsSecurityGroups{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticComputeCloudSpotFleetSpotFleetRequestConfigDataLaunchSpecificationsSecurityGroupsList(list)
		return nil
	}
	return err
}

// EC2ContainerServiceServiceDeploymentConfiguration represents Amazon EC2 Container Service Service DeploymentConfiguration
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-service-deploymentconfiguration.html
type EC2ContainerServiceServiceDeploymentConfiguration struct {
	// The maximum number of tasks, specified as a percentage of the Amazon
	// ECS service's DesiredCount value, that can run in a service during a
	// deployment. To calculate the maximum number of tasks, Amazon ECS uses
	// this formula: the value of DesiredCount * (the value of the
	// MaximumPercent/100), rounded down to the nearest integer value.
	MaximumPercent *IntegerExpr `json:"MaximumPercent,omitempty"`

	// The minimum number of tasks, specified as a percentage of the Amazon
	// ECS service's DesiredCount value, that must continue to run and remain
	// healthy during a deployment. To calculate the minimum number of tasks,
	// Amazon ECS uses this formula: the value of DesiredCount * (the value
	// of the MinimumHealthyPercent/100), rounded up to the nearest integer
	// value.
	MinimumHealthyPercent *IntegerExpr `json:"MinimumHealthyPercent,omitempty"`
}

// EC2ContainerServiceServiceDeploymentConfigurationList represents a list of EC2ContainerServiceServiceDeploymentConfiguration
type EC2ContainerServiceServiceDeploymentConfigurationList []EC2ContainerServiceServiceDeploymentConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2ContainerServiceServiceDeploymentConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2ContainerServiceServiceDeploymentConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2ContainerServiceServiceDeploymentConfigurationList{item}
		return nil
	}
	list := []EC2ContainerServiceServiceDeploymentConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2ContainerServiceServiceDeploymentConfigurationList(list)
		return nil
	}
	return err
}

// EC2ContainerServiceServiceLoadBalancers represents Amazon EC2 Container Service Service LoadBalancers
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-service-loadbalancers.html
type EC2ContainerServiceServiceLoadBalancers struct {
	// The name of a container to use with the load balancer.
	ContainerName *StringExpr `json:"ContainerName,omitempty"`

	// The port number on the container to direct load balancer traffic to.
	// Your container instances must allow ingress traffic on this port.
	ContainerPort *IntegerExpr `json:"ContainerPort,omitempty"`

	// The name of a Classic Load Balancer to associate with the Amazon ECS
	// service.
	LoadBalancerName *StringExpr `json:"LoadBalancerName,omitempty"`

	// An Application load balancer target group Amazon Resource Name (ARN)
	// to associate with the Amazon ECS service.
	TargetGroupArn *StringExpr `json:"TargetGroupArn,omitempty"`
}

// EC2ContainerServiceServiceLoadBalancersList represents a list of EC2ContainerServiceServiceLoadBalancers
type EC2ContainerServiceServiceLoadBalancersList []EC2ContainerServiceServiceLoadBalancers

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2ContainerServiceServiceLoadBalancersList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2ContainerServiceServiceLoadBalancers{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2ContainerServiceServiceLoadBalancersList{item}
		return nil
	}
	list := []EC2ContainerServiceServiceLoadBalancers{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2ContainerServiceServiceLoadBalancersList(list)
		return nil
	}
	return err
}

// EC2ContainerServiceTaskDefinitionContainerDefinitions represents Amazon EC2 Container Service TaskDefinition ContainerDefinitions
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions.html
type EC2ContainerServiceTaskDefinitionContainerDefinitions struct {
	// The CMD value to pass to the container. For more information about the
	// Docker CMD parameter, see
	// https://docs.docker.com/reference/builder/#cmd.
	Command *StringListExpr `json:"Command,omitempty"`

	// The minimum number of CPU units to reserve for the container.
	// Containers share unallocated CPU units with other containers on the
	// instance by using the same ratio as their allocated CPU units. For
	// more information, see the cpu content for the ContainerDefinition data
	// type in the Amazon EC2 Container Service API Reference.
	Cpu *IntegerExpr `json:"Cpu,omitempty"`

	// Indicates whether networking is disabled within the container.
	DisableNetworking *BoolExpr `json:"DisableNetworking,omitempty"`

	// A list of DNS search domains that are provided to the container. The
	// domain names that the DNS logic looks up when a process attempts to
	// access a bare unqualified hostname.
	DnsSearchDomains *StringListExpr `json:"DnsSearchDomains,omitempty"`

	// A list of DNS servers that Amazon ECS provides to the container.
	DnsServers *StringListExpr `json:"DnsServers,omitempty"`

	// A key-value map of labels for the container.
	DockerLabels interface{} `json:"DockerLabels,omitempty"`

	// A list of custom labels for SELinux and AppArmor multi-level security
	// systems. For more information, see the dockerSecurityOptions content
	// for the ContainerDefinition data type in the Amazon EC2 Container
	// Service API Reference.
	DockerSecurityOptions *StringListExpr `json:"DockerSecurityOptions,omitempty"`

	// The ENTRYPOINT value to pass to the container. For more information
	// about the Docker ENTRYPOINT parameter, see
	// https://docs.docker.com/reference/builder/#entrypoint.
	EntryPoint *StringListExpr `json:"EntryPoint,omitempty"`

	// The environment variables to pass to the container.
	Environment *EC2ContainerServiceTaskDefinitionContainerDefinitionsEnvironmentList `json:"Environment,omitempty"`

	// Indicates whether the task stops if this container fails. If you
	// specify true and the container fails, all other containers in the task
	// stop. If you specify false and the container fails, none of the other
	// containers in the task is affected. This value is true by default.
	Essential *BoolExpr `json:"Essential,omitempty"`

	// A list of hostnames and IP address mappings to append to the
	// /etc/hosts file on the container.
	ExtraHosts *EC2ContainerServiceTaskDefinitionContainerDefinitionsHostEntryList `json:"ExtraHosts,omitempty"`

	// The name that Docker will use for the container's hostname.
	Hostname *StringExpr `json:"Hostname,omitempty"`

	// The image to use for a container, which is passed directly to the
	// Docker daemon. You can use images in the Docker Hub registry or
	// specify other repositories (repository-url/image:tag).
	Image *StringExpr `json:"Image,omitempty"`

	// The name of another container to connect to. With links, containers
	// can communicate with each other without using port mappings.
	Links *StringListExpr `json:"Links,omitempty"`

	// Configures a custom log driver for the container. For more
	// information, see the logConfiguration content for the
	// ContainerDefinition data type in the Amazon EC2 Container Service API
	// Reference.
	LogConfiguration *EC2ContainerServiceTaskDefinitionContainerDefinitionsLogConfiguration `json:"LogConfiguration,omitempty"`

	// The number of MiB of memory to reserve for the container. If your
	// container attempts to exceed the allocated memory, the container is
	// terminated.
	Memory *IntegerExpr `json:"Memory,omitempty"`

	// The mount points for data volumes in the container.
	MountPoints *EC2ContainerServiceTaskDefinitionContainerDefinitionsMountPointsList `json:"MountPoints,omitempty"`

	// A name for the container.
	Name *StringExpr `json:"Name,omitempty"`

	// A mapping of the container port to a host port. Port mappings enable
	// containers to access ports on the host container instance to send or
	// receive traffic.
	PortMappings *EC2ContainerServiceTaskDefinitionContainerDefinitionsPortMappingsList `json:"PortMappings,omitempty"`

	// Indicates whether the container is given full access to the host
	// container instance.
	Privileged *BoolExpr `json:"Privileged,omitempty"`

	// Indicates whether the container's root file system is mounted as read
	// only.
	ReadonlyRootFilesystem *BoolExpr `json:"ReadonlyRootFilesystem,omitempty"`

	// A list of ulimits to set in the container. The ulimits set constraints
	// on how much resources a container can consume so that it doesn't
	// deplete all available resources on the host.
	Ulimits *EC2ContainerServiceTaskDefinitionContainerDefinitionsUlimitList `json:"Ulimits,omitempty"`

	// The user name to use inside the container.
	User *StringExpr `json:"User,omitempty"`

	// The data volumes to mount from another container.
	VolumesFrom *EC2ContainerServiceTaskDefinitionContainerDefinitionsVolumesFromList `json:"VolumesFrom,omitempty"`

	// The working directory in the container in which to run commands.
	WorkingDirectory *StringExpr `json:"WorkingDirectory,omitempty"`
}

// EC2ContainerServiceTaskDefinitionContainerDefinitionsList represents a list of EC2ContainerServiceTaskDefinitionContainerDefinitions
type EC2ContainerServiceTaskDefinitionContainerDefinitionsList []EC2ContainerServiceTaskDefinitionContainerDefinitions

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2ContainerServiceTaskDefinitionContainerDefinitionsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2ContainerServiceTaskDefinitionContainerDefinitions{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2ContainerServiceTaskDefinitionContainerDefinitionsList{item}
		return nil
	}
	list := []EC2ContainerServiceTaskDefinitionContainerDefinitions{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2ContainerServiceTaskDefinitionContainerDefinitionsList(list)
		return nil
	}
	return err
}

// EC2ContainerServiceTaskDefinitionContainerDefinitionsEnvironment represents Amazon EC2 Container Service TaskDefinition ContainerDefinitions Environment
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions-environment.html
type EC2ContainerServiceTaskDefinitionContainerDefinitionsEnvironment struct {
	// The name of the environment variable.
	Name *StringExpr `json:"Name,omitempty"`

	// The value of the environment variable.
	Value *StringExpr `json:"Value,omitempty"`
}

// EC2ContainerServiceTaskDefinitionContainerDefinitionsEnvironmentList represents a list of EC2ContainerServiceTaskDefinitionContainerDefinitionsEnvironment
type EC2ContainerServiceTaskDefinitionContainerDefinitionsEnvironmentList []EC2ContainerServiceTaskDefinitionContainerDefinitionsEnvironment

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2ContainerServiceTaskDefinitionContainerDefinitionsEnvironmentList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2ContainerServiceTaskDefinitionContainerDefinitionsEnvironment{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2ContainerServiceTaskDefinitionContainerDefinitionsEnvironmentList{item}
		return nil
	}
	list := []EC2ContainerServiceTaskDefinitionContainerDefinitionsEnvironment{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2ContainerServiceTaskDefinitionContainerDefinitionsEnvironmentList(list)
		return nil
	}
	return err
}

// EC2ContainerServiceTaskDefinitionContainerDefinitionsHostEntry represents Amazon EC2 Container Service TaskDefinition ContainerDefinitions HostEntry
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions-hostentry.html
type EC2ContainerServiceTaskDefinitionContainerDefinitionsHostEntry struct {
	// The hostname to use in the /etc/hosts file.
	Hostname *StringExpr `json:"Hostname,omitempty"`

	// The IP address to use in the /etc/hosts file.
	IpAddress *StringExpr `json:"IpAddress,omitempty"`
}

// EC2ContainerServiceTaskDefinitionContainerDefinitionsHostEntryList represents a list of EC2ContainerServiceTaskDefinitionContainerDefinitionsHostEntry
type EC2ContainerServiceTaskDefinitionContainerDefinitionsHostEntryList []EC2ContainerServiceTaskDefinitionContainerDefinitionsHostEntry

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2ContainerServiceTaskDefinitionContainerDefinitionsHostEntryList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2ContainerServiceTaskDefinitionContainerDefinitionsHostEntry{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2ContainerServiceTaskDefinitionContainerDefinitionsHostEntryList{item}
		return nil
	}
	list := []EC2ContainerServiceTaskDefinitionContainerDefinitionsHostEntry{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2ContainerServiceTaskDefinitionContainerDefinitionsHostEntryList(list)
		return nil
	}
	return err
}

// EC2ContainerServiceTaskDefinitionContainerDefinitionsLogConfiguration represents Amazon EC2 Container Service TaskDefinition ContainerDefinitions LogConfiguration
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions-logconfiguration.html
type EC2ContainerServiceTaskDefinitionContainerDefinitionsLogConfiguration struct {
	// The log driver to use for the container. This parameter requires that
	// your container instance uses Docker Remote API Version 1.18 or
	// greater. For more information, see the logDriver content for the
	// LogConfiguration data type in the Amazon EC2 Container Service API
	// Reference.
	LogDriver *StringExpr `json:"LogDriver,omitempty"`

	// The configuration options to send to the log driver. This parameter
	// requires that your container instance uses Docker Remote API Version
	// 1.18 or greater.
	Options interface{} `json:"Options,omitempty"`
}

// EC2ContainerServiceTaskDefinitionContainerDefinitionsLogConfigurationList represents a list of EC2ContainerServiceTaskDefinitionContainerDefinitionsLogConfiguration
type EC2ContainerServiceTaskDefinitionContainerDefinitionsLogConfigurationList []EC2ContainerServiceTaskDefinitionContainerDefinitionsLogConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2ContainerServiceTaskDefinitionContainerDefinitionsLogConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2ContainerServiceTaskDefinitionContainerDefinitionsLogConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2ContainerServiceTaskDefinitionContainerDefinitionsLogConfigurationList{item}
		return nil
	}
	list := []EC2ContainerServiceTaskDefinitionContainerDefinitionsLogConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2ContainerServiceTaskDefinitionContainerDefinitionsLogConfigurationList(list)
		return nil
	}
	return err
}

// EC2ContainerServiceTaskDefinitionContainerDefinitionsMountPoints represents Amazon EC2 Container Service TaskDefinition ContainerDefinitions MountPoints
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions-mountpoints.html
type EC2ContainerServiceTaskDefinitionContainerDefinitionsMountPoints struct {
	// The path on the container that indicates where you want to mount the
	// volume.
	ContainerPath *StringExpr `json:"ContainerPath,omitempty"`

	// The name of the volume to mount.
	SourceVolume *StringExpr `json:"SourceVolume,omitempty"`

	// Indicates whether the container can write to the volume. If you
	// specify true, the container has read-only access to the volume. If you
	// specify false, the container can write to the volume. By default, the
	// value is false.
	ReadOnly *BoolExpr `json:"ReadOnly,omitempty"`
}

// EC2ContainerServiceTaskDefinitionContainerDefinitionsMountPointsList represents a list of EC2ContainerServiceTaskDefinitionContainerDefinitionsMountPoints
type EC2ContainerServiceTaskDefinitionContainerDefinitionsMountPointsList []EC2ContainerServiceTaskDefinitionContainerDefinitionsMountPoints

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2ContainerServiceTaskDefinitionContainerDefinitionsMountPointsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2ContainerServiceTaskDefinitionContainerDefinitionsMountPoints{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2ContainerServiceTaskDefinitionContainerDefinitionsMountPointsList{item}
		return nil
	}
	list := []EC2ContainerServiceTaskDefinitionContainerDefinitionsMountPoints{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2ContainerServiceTaskDefinitionContainerDefinitionsMountPointsList(list)
		return nil
	}
	return err
}

// EC2ContainerServiceTaskDefinitionContainerDefinitionsPortMappings represents Amazon EC2 Container Service TaskDefinition ContainerDefinitions PortMappings
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions-portmappings.html
type EC2ContainerServiceTaskDefinitionContainerDefinitionsPortMappings struct {
	// The port number on the container bound to the host port.
	ContainerPort *IntegerExpr `json:"ContainerPort,omitempty"`

	// The host port number on the container instance that you want to
	// reserve for your container. You can specify a non-reserved host port
	// for your container port mapping, omit the host port, or set the host
	// port to 0. If you specify a container port but no host port, your
	// container host port is assigned automatically .
	HostPort *IntegerExpr `json:"HostPort,omitempty"`

	// The protocol used for the port mapping. For valid values, see the
	// protocol parameter in the Amazon EC2 Container Service Developer
	// Guide. By default, AWS CloudFormation specifies tcp.
	Protocol *StringExpr `json:"Protocol,omitempty"`
}

// EC2ContainerServiceTaskDefinitionContainerDefinitionsPortMappingsList represents a list of EC2ContainerServiceTaskDefinitionContainerDefinitionsPortMappings
type EC2ContainerServiceTaskDefinitionContainerDefinitionsPortMappingsList []EC2ContainerServiceTaskDefinitionContainerDefinitionsPortMappings

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2ContainerServiceTaskDefinitionContainerDefinitionsPortMappingsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2ContainerServiceTaskDefinitionContainerDefinitionsPortMappings{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2ContainerServiceTaskDefinitionContainerDefinitionsPortMappingsList{item}
		return nil
	}
	list := []EC2ContainerServiceTaskDefinitionContainerDefinitionsPortMappings{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2ContainerServiceTaskDefinitionContainerDefinitionsPortMappingsList(list)
		return nil
	}
	return err
}

// EC2ContainerServiceTaskDefinitionContainerDefinitionsUlimit represents Amazon EC2 Container Service TaskDefinition ContainerDefinitions Ulimit
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions-ulimit.html
type EC2ContainerServiceTaskDefinitionContainerDefinitionsUlimit struct {
	// The hard limit for the ulimit type.
	HardLimit *IntegerExpr `json:"HardLimit,omitempty"`

	// The type of ulimit. For valid values, see the name content for the
	// Ulimit data type in the Amazon EC2 Container Service API Reference.
	Name *StringExpr `json:"Name,omitempty"`

	// The soft limit for the ulimit type.
	SoftLimit *IntegerExpr `json:"SoftLimit,omitempty"`
}

// EC2ContainerServiceTaskDefinitionContainerDefinitionsUlimitList represents a list of EC2ContainerServiceTaskDefinitionContainerDefinitionsUlimit
type EC2ContainerServiceTaskDefinitionContainerDefinitionsUlimitList []EC2ContainerServiceTaskDefinitionContainerDefinitionsUlimit

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2ContainerServiceTaskDefinitionContainerDefinitionsUlimitList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2ContainerServiceTaskDefinitionContainerDefinitionsUlimit{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2ContainerServiceTaskDefinitionContainerDefinitionsUlimitList{item}
		return nil
	}
	list := []EC2ContainerServiceTaskDefinitionContainerDefinitionsUlimit{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2ContainerServiceTaskDefinitionContainerDefinitionsUlimitList(list)
		return nil
	}
	return err
}

// EC2ContainerServiceTaskDefinitionContainerDefinitionsVolumesFrom represents Amazon EC2 Container Service TaskDefinition ContainerDefinitions VolumesFrom
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions-volumesfrom.html
type EC2ContainerServiceTaskDefinitionContainerDefinitionsVolumesFrom struct {
	// The name of the container that has the volumes to mount.
	SourceContainer *StringExpr `json:"SourceContainer,omitempty"`

	// Indicates whether the container can write to the volume. If you
	// specify true, the container has read-only access to the volume. If you
	// specify false, the container can write to the volume. By default, the
	// value is false.
	ReadOnly *BoolExpr `json:"ReadOnly,omitempty"`
}

// EC2ContainerServiceTaskDefinitionContainerDefinitionsVolumesFromList represents a list of EC2ContainerServiceTaskDefinitionContainerDefinitionsVolumesFrom
type EC2ContainerServiceTaskDefinitionContainerDefinitionsVolumesFromList []EC2ContainerServiceTaskDefinitionContainerDefinitionsVolumesFrom

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2ContainerServiceTaskDefinitionContainerDefinitionsVolumesFromList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2ContainerServiceTaskDefinitionContainerDefinitionsVolumesFrom{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2ContainerServiceTaskDefinitionContainerDefinitionsVolumesFromList{item}
		return nil
	}
	list := []EC2ContainerServiceTaskDefinitionContainerDefinitionsVolumesFrom{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2ContainerServiceTaskDefinitionContainerDefinitionsVolumesFromList(list)
		return nil
	}
	return err
}

// EC2ContainerServiceTaskDefinitionVolumes represents Amazon EC2 Container Service TaskDefinition Volumes
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-volumes.html
type EC2ContainerServiceTaskDefinitionVolumes struct {
	// The name of the volume. To specify mount points in your container
	// definitions, use the value of this property.
	Name *StringExpr `json:"Name,omitempty"`

	// Determines whether your data volume persists on the host container
	// instance and at the location where it is stored.
	Host *EC2ContainerServiceTaskDefinitionVolumesHost `json:"Host,omitempty"`
}

// EC2ContainerServiceTaskDefinitionVolumesList represents a list of EC2ContainerServiceTaskDefinitionVolumes
type EC2ContainerServiceTaskDefinitionVolumesList []EC2ContainerServiceTaskDefinitionVolumes

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2ContainerServiceTaskDefinitionVolumesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2ContainerServiceTaskDefinitionVolumes{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2ContainerServiceTaskDefinitionVolumesList{item}
		return nil
	}
	list := []EC2ContainerServiceTaskDefinitionVolumes{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2ContainerServiceTaskDefinitionVolumesList(list)
		return nil
	}
	return err
}

// EC2ContainerServiceTaskDefinitionVolumesHost represents Amazon EC2 Container Service TaskDefinition Volumes Host
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-volumes-host.html
type EC2ContainerServiceTaskDefinitionVolumesHost struct {
	// The data volume path on the host container instance.
	SourcePath *StringExpr `json:"SourcePath,omitempty"`
}

// EC2ContainerServiceTaskDefinitionVolumesHostList represents a list of EC2ContainerServiceTaskDefinitionVolumesHost
type EC2ContainerServiceTaskDefinitionVolumesHostList []EC2ContainerServiceTaskDefinitionVolumesHost

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EC2ContainerServiceTaskDefinitionVolumesHostList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EC2ContainerServiceTaskDefinitionVolumesHost{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EC2ContainerServiceTaskDefinitionVolumesHostList{item}
		return nil
	}
	list := []EC2ContainerServiceTaskDefinitionVolumesHost{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EC2ContainerServiceTaskDefinitionVolumesHostList(list)
		return nil
	}
	return err
}

// ElasticFileSystemFileSystemFileSystemTags represents Amazon Elastic File System FileSystem FileSystemTags
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-efs-filesystem-filesystemtags.html
type ElasticFileSystemFileSystemFileSystemTags struct {
	// The key name of the tag. You can specify a value that is from 1 to 128
	// Unicode characters in length, but you cannot use the prefix aws:.
	Key *StringExpr `json:"Key,omitempty"`

	// The value of the tag key. You can specify a value that is from 0 to
	// 128 Unicode characters in length.
	Value *StringExpr `json:"Value,omitempty"`
}

// ElasticFileSystemFileSystemFileSystemTagsList represents a list of ElasticFileSystemFileSystemFileSystemTags
type ElasticFileSystemFileSystemFileSystemTagsList []ElasticFileSystemFileSystemFileSystemTags

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticFileSystemFileSystemFileSystemTagsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticFileSystemFileSystemFileSystemTags{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticFileSystemFileSystemFileSystemTagsList{item}
		return nil
	}
	list := []ElasticFileSystemFileSystemFileSystemTags{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticFileSystemFileSystemFileSystemTagsList(list)
		return nil
	}
	return err
}

// ElasticBeanstalkEnvironmentTier represents Elastic Beanstalk Environment Tier Property Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-beanstalk-environment-tier.html
type ElasticBeanstalkEnvironmentTier struct {
	// The name of the environment tier. You can specify WebServer or Worker.
	Name *StringExpr `json:"Name,omitempty"`

	// The type of this environment tier. You can specify Standard for the
	// WebServer tier or SQS/HTTP for the Worker tier.
	Type *StringExpr `json:"Type,omitempty"`

	// The version of this environment tier.
	Version *StringExpr `json:"Version,omitempty"`
}

// ElasticBeanstalkEnvironmentTierList represents a list of ElasticBeanstalkEnvironmentTier
type ElasticBeanstalkEnvironmentTierList []ElasticBeanstalkEnvironmentTier

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticBeanstalkEnvironmentTierList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticBeanstalkEnvironmentTier{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticBeanstalkEnvironmentTierList{item}
		return nil
	}
	list := []ElasticBeanstalkEnvironmentTier{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticBeanstalkEnvironmentTierList(list)
		return nil
	}
	return err
}

// ElasticBeanstalkOptionSettings represents Elastic Beanstalk OptionSettings Property Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-beanstalk-option-settings.html
type ElasticBeanstalkOptionSettings struct {
	// A unique namespace identifying the option's associated AWS resource.
	// For a list of namespaces that you can use, see Configuration Options
	// in the AWS Elastic Beanstalk Developer Guide.
	Namespace *StringExpr `json:"Namespace,omitempty"`

	// The name of the configuration option. For a list of options that you
	// can use, see Configuration Options in the AWS Elastic Beanstalk
	// Developer Guide.
	OptionName *StringExpr `json:"OptionName,omitempty"`

	// The value of the setting.
	Value *StringExpr `json:"Value,omitempty"`
}

// ElasticBeanstalkOptionSettingsList represents a list of ElasticBeanstalkOptionSettings
type ElasticBeanstalkOptionSettingsList []ElasticBeanstalkOptionSettings

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticBeanstalkOptionSettingsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticBeanstalkOptionSettings{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticBeanstalkOptionSettingsList{item}
		return nil
	}
	list := []ElasticBeanstalkOptionSettings{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticBeanstalkOptionSettingsList(list)
		return nil
	}
	return err
}

// ElasticBeanstalkSourceBundle represents Elastic Beanstalk SourceBundle Property Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-beanstalk-sourcebundle.html
type ElasticBeanstalkSourceBundle struct {
	// The Amazon S3 bucket where the data is located.
	S3Bucket *StringExpr `json:"S3Bucket,omitempty"`

	// The Amazon S3 key where the data is located.
	S3Key *StringExpr `json:"S3Key,omitempty"`
}

// ElasticBeanstalkSourceBundleList represents a list of ElasticBeanstalkSourceBundle
type ElasticBeanstalkSourceBundleList []ElasticBeanstalkSourceBundle

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticBeanstalkSourceBundleList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticBeanstalkSourceBundle{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticBeanstalkSourceBundleList{item}
		return nil
	}
	list := []ElasticBeanstalkSourceBundle{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticBeanstalkSourceBundleList(list)
		return nil
	}
	return err
}

// ElasticBeanstalkSourceConfiguration represents Elastic Beanstalk SourceConfiguration Property Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-beanstalk-configurationtemplate-sourceconfiguration.html
type ElasticBeanstalkSourceConfiguration struct {
	// The name of the Elastic Beanstalk application that contains the
	// configuration template that you want to use.
	ApplicationName *StringExpr `json:"ApplicationName,omitempty"`

	// The name of the configuration template.
	TemplateName *StringExpr `json:"TemplateName,omitempty"`
}

// ElasticBeanstalkSourceConfigurationList represents a list of ElasticBeanstalkSourceConfiguration
type ElasticBeanstalkSourceConfigurationList []ElasticBeanstalkSourceConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticBeanstalkSourceConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticBeanstalkSourceConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticBeanstalkSourceConfigurationList{item}
		return nil
	}
	list := []ElasticBeanstalkSourceConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticBeanstalkSourceConfigurationList(list)
		return nil
	}
	return err
}

// ElastiCacheReplicationGroupNodeGroupConfiguration represents Amazon ElastiCache ReplicationGroup NodeGroupConfiguration
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-elasticache-replicationgroup-nodegroupconfiguration.html
type ElastiCacheReplicationGroupNodeGroupConfiguration struct {
	// The Availability Zone where ElastiCache launches the node group's
	// primary node.
	PrimaryAvailabilityZone *StringExpr `json:"PrimaryAvailabilityZone,omitempty"`

	// A list of Availability Zones where ElastiCache launches the read
	// replicas. The number of Availability Zones must match the value of the
	// ReplicaCount property or, if you don't specify the ReplicaCount
	// property, the replication group's ReplicasPerNodeGroup property.
	ReplicaAvailabilityZones *StringListExpr `json:"ReplicaAvailabilityZones,omitempty"`

	// The number of read replica nodes in the node group.
	ReplicaCount *IntegerExpr `json:"ReplicaCount,omitempty"`

	// A string of comma-separated values where the first set of values are
	// the slot numbers (zero based), and the second set of values are the
	// keyspaces for each slot. The following example specifies three slots
	// (numbered 0, 1, and 2): 0,1,2,0-4999,5000-9999,10000-16,383.
	Slots *StringExpr `json:"Slots,omitempty"`
}

// ElastiCacheReplicationGroupNodeGroupConfigurationList represents a list of ElastiCacheReplicationGroupNodeGroupConfiguration
type ElastiCacheReplicationGroupNodeGroupConfigurationList []ElastiCacheReplicationGroupNodeGroupConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElastiCacheReplicationGroupNodeGroupConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElastiCacheReplicationGroupNodeGroupConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElastiCacheReplicationGroupNodeGroupConfigurationList{item}
		return nil
	}
	list := []ElastiCacheReplicationGroupNodeGroupConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElastiCacheReplicationGroupNodeGroupConfigurationList(list)
		return nil
	}
	return err
}

// ElasticLoadBalancingAccessLoggingPolicy represents Elastic Load Balancing AccessLoggingPolicy
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-elb-accessloggingpolicy.html
type ElasticLoadBalancingAccessLoggingPolicy struct {
	// The interval for publishing access logs in minutes. You can specify an
	// interval of either 5 minutes or 60 minutes.
	EmitInterval *IntegerExpr `json:"EmitInterval,omitempty"`

	// Whether logging is enabled for the load balancer.
	Enabled *BoolExpr `json:"Enabled,omitempty"`

	// The name of an Amazon S3 bucket where access log files are stored.
	S3BucketName *StringExpr `json:"S3BucketName,omitempty"`

	// A prefix for the all log object keys, such as
	// my-load-balancer-logs/prod. If you store log files from multiple
	// sources in a single bucket, you can use a prefix to distinguish each
	// log file and its source.
	S3BucketPrefix *StringExpr `json:"S3BucketPrefix,omitempty"`
}

// ElasticLoadBalancingAccessLoggingPolicyList represents a list of ElasticLoadBalancingAccessLoggingPolicy
type ElasticLoadBalancingAccessLoggingPolicyList []ElasticLoadBalancingAccessLoggingPolicy

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticLoadBalancingAccessLoggingPolicyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticLoadBalancingAccessLoggingPolicy{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticLoadBalancingAccessLoggingPolicyList{item}
		return nil
	}
	list := []ElasticLoadBalancingAccessLoggingPolicy{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticLoadBalancingAccessLoggingPolicyList(list)
		return nil
	}
	return err
}

// ElasticLoadBalancingAppCookieStickinessPolicy represents ElasticLoadBalancing AppCookieStickinessPolicy Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-elb-AppCookieStickinessPolicy.html
type ElasticLoadBalancingAppCookieStickinessPolicy struct {
	// Name of the application cookie used for stickiness.
	CookieName *StringExpr `json:"CookieName,omitempty"`

	// The name of the policy being created. The name must be unique within
	// the set of policies for this Load Balancer.
	PolicyName *StringExpr `json:"PolicyName,omitempty"`
}

// ElasticLoadBalancingAppCookieStickinessPolicyList represents a list of ElasticLoadBalancingAppCookieStickinessPolicy
type ElasticLoadBalancingAppCookieStickinessPolicyList []ElasticLoadBalancingAppCookieStickinessPolicy

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticLoadBalancingAppCookieStickinessPolicyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticLoadBalancingAppCookieStickinessPolicy{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticLoadBalancingAppCookieStickinessPolicyList{item}
		return nil
	}
	list := []ElasticLoadBalancingAppCookieStickinessPolicy{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticLoadBalancingAppCookieStickinessPolicyList(list)
		return nil
	}
	return err
}

// ElasticLoadBalancingConnectionDrainingPolicy represents Elastic Load Balancing ConnectionDrainingPolicy
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-elb-connectiondrainingpolicy.html
type ElasticLoadBalancingConnectionDrainingPolicy struct {
	// Whether or not connection draining is enabled for the load balancer.
	Enabled *BoolExpr `json:"Enabled,omitempty"`

	// The time in seconds after the load balancer closes all connections to
	// a deregistered or unhealthy instance.
	Timeout *IntegerExpr `json:"Timeout,omitempty"`
}

// ElasticLoadBalancingConnectionDrainingPolicyList represents a list of ElasticLoadBalancingConnectionDrainingPolicy
type ElasticLoadBalancingConnectionDrainingPolicyList []ElasticLoadBalancingConnectionDrainingPolicy

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticLoadBalancingConnectionDrainingPolicyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticLoadBalancingConnectionDrainingPolicy{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticLoadBalancingConnectionDrainingPolicyList{item}
		return nil
	}
	list := []ElasticLoadBalancingConnectionDrainingPolicy{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticLoadBalancingConnectionDrainingPolicyList(list)
		return nil
	}
	return err
}

// ElasticLoadBalancingConnectionSettings represents Elastic Load Balancing ConnectionSettings
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-elb-connectionsettings.html
type ElasticLoadBalancingConnectionSettings struct {
	// The time (in seconds) that a connection to the load balancer can
	// remain idle, which means no data is sent over the connection. After
	// the specified time, the load balancer closes the connection.
	IdleTimeout *IntegerExpr `json:"IdleTimeout,omitempty"`
}

// ElasticLoadBalancingConnectionSettingsList represents a list of ElasticLoadBalancingConnectionSettings
type ElasticLoadBalancingConnectionSettingsList []ElasticLoadBalancingConnectionSettings

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticLoadBalancingConnectionSettingsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticLoadBalancingConnectionSettings{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticLoadBalancingConnectionSettingsList{item}
		return nil
	}
	list := []ElasticLoadBalancingConnectionSettings{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticLoadBalancingConnectionSettingsList(list)
		return nil
	}
	return err
}

// ElasticLoadBalancingHealthCheck represents ElasticLoadBalancing HealthCheck Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-elb-health-check.html
type ElasticLoadBalancingHealthCheck struct {
	// Specifies the number of consecutive health probe successes required
	// before moving the instance to the Healthy state.
	HealthyThreshold *StringExpr `json:"HealthyThreshold,omitempty"`

	// Specifies the approximate interval, in seconds, between health checks
	// of an individual instance.
	Interval *StringExpr `json:"Interval,omitempty"`

	// Specifies the instance's protocol and port to check. The protocol can
	// be TCP, HTTP, HTTPS, or SSL. The range of valid ports is 1 through
	// 65535.
	Target *StringExpr `json:"Target,omitempty"`

	// Specifies the amount of time, in seconds, during which no response
	// means a failed health probe. This value must be less than the value
	// for Interval.
	Timeout *StringExpr `json:"Timeout,omitempty"`

	// Specifies the number of consecutive health probe failures required
	// before moving the instance to the Unhealthy state.
	UnhealthyThreshold *StringExpr `json:"UnhealthyThreshold,omitempty"`
}

// ElasticLoadBalancingHealthCheckList represents a list of ElasticLoadBalancingHealthCheck
type ElasticLoadBalancingHealthCheckList []ElasticLoadBalancingHealthCheck

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticLoadBalancingHealthCheckList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticLoadBalancingHealthCheck{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticLoadBalancingHealthCheckList{item}
		return nil
	}
	list := []ElasticLoadBalancingHealthCheck{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticLoadBalancingHealthCheckList(list)
		return nil
	}
	return err
}

// ElasticLoadBalancingLBCookieStickinessPolicy represents ElasticLoadBalancing LBCookieStickinessPolicy Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-elb-LBCookieStickinessPolicy.html
type ElasticLoadBalancingLBCookieStickinessPolicy struct {
	// The time period, in seconds, after which the cookie should be
	// considered stale. If this parameter isn't specified, the sticky
	// session will last for the duration of the browser session.
	CookieExpirationPeriod *StringExpr `json:"CookieExpirationPeriod,omitempty"`

	// The name of the policy being created. The name must be unique within
	// the set of policies for this load balancer.
	PolicyName interface{} `json:"PolicyName,omitempty"`
}

// ElasticLoadBalancingLBCookieStickinessPolicyList represents a list of ElasticLoadBalancingLBCookieStickinessPolicy
type ElasticLoadBalancingLBCookieStickinessPolicyList []ElasticLoadBalancingLBCookieStickinessPolicy

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticLoadBalancingLBCookieStickinessPolicyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticLoadBalancingLBCookieStickinessPolicy{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticLoadBalancingLBCookieStickinessPolicyList{item}
		return nil
	}
	list := []ElasticLoadBalancingLBCookieStickinessPolicy{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticLoadBalancingLBCookieStickinessPolicyList(list)
		return nil
	}
	return err
}

// ElasticLoadBalancingListener represents ElasticLoadBalancing Listener Property Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-elb-listener.html
type ElasticLoadBalancingListener struct {
	// Specifies the TCP port on which the instance server listens. You can't
	// modify this property during the life of the load balancer.
	InstancePort *StringExpr `json:"InstancePort,omitempty"`

	// Specifies the protocol to use for routing traffic to back-end
	// instances: HTTP, HTTPS, TCP, or SSL. You can't modify this property
	// during the life of the load balancer.
	InstanceProtocol *StringExpr `json:"InstanceProtocol,omitempty"`

	// Specifies the external load balancer port number. You can't modify
	// this property during the life of the load balancer.
	LoadBalancerPort *StringExpr `json:"LoadBalancerPort,omitempty"`

	// A list of ElasticLoadBalancing policy names to associate with the
	// Listener. Specify only policies that are compatible with a Listener.
	// For more information, see DescribeLoadBalancerPolicyTypes in the
	// Elastic Load Balancing API Reference version 2012-06-01.
	PolicyNames *StringListExpr `json:"PolicyNames,omitempty"`

	// Specifies the load balancer transport protocol to use for routing:
	// HTTP, HTTPS, TCP or SSL. You can't modify this property during the
	// life of the load balancer.
	Protocol *StringExpr `json:"Protocol,omitempty"`

	// The ARN of the SSL certificate to use. For more information about SSL
	// certificates, see Managing Server Certificates in the AWS Identity and
	// Access Management User Guide.
	SSLCertificateId *StringExpr `json:"SSLCertificateId,omitempty"`
}

// ElasticLoadBalancingListenerList represents a list of ElasticLoadBalancingListener
type ElasticLoadBalancingListenerList []ElasticLoadBalancingListener

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticLoadBalancingListenerList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticLoadBalancingListener{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticLoadBalancingListenerList{item}
		return nil
	}
	list := []ElasticLoadBalancingListener{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticLoadBalancingListenerList(list)
		return nil
	}
	return err
}

// ElasticLoadBalancingPolicy represents ElasticLoadBalancing Policy Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-elb-policy.html
type ElasticLoadBalancingPolicy struct {
	// A list of arbitrary attributes for this policy. If you don't need to
	// specify any policy attributes, specify an empty list ([]).
	Attributes interface{} `json:"Attributes,omitempty"`

	// A list of instance ports for the policy. These are the ports
	// associated with the back-end server.
	InstancePorts interface{} `json:"InstancePorts,omitempty"`

	// A list of external load balancer ports for the policy.
	LoadBalancerPorts interface{} `json:"LoadBalancerPorts,omitempty"`

	// A name for this policy that is unique to the load balancer.
	PolicyName *StringExpr `json:"PolicyName,omitempty"`

	// The name of the policy type for this policy. This must be one of the
	// types reported by the Elastic Load Balancing
	// DescribeLoadBalancerPolicyTypes action.
	PolicyType *StringExpr `json:"PolicyType,omitempty"`
}

// ElasticLoadBalancingPolicyList represents a list of ElasticLoadBalancingPolicy
type ElasticLoadBalancingPolicyList []ElasticLoadBalancingPolicy

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticLoadBalancingPolicyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticLoadBalancingPolicy{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticLoadBalancingPolicyList{item}
		return nil
	}
	list := []ElasticLoadBalancingPolicy{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticLoadBalancingPolicyList(list)
		return nil
	}
	return err
}

// ElasticLoadBalancingListenerCertificates represents Elastic Load Balancing Listener Certificates
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-elasticloadbalancingv2-listener-certificates.html
type ElasticLoadBalancingListenerCertificates struct {
	// The Amazon Resource Name (ARN) of the certificate to associate with
	// the listener.
	CertificateArn *StringExpr `json:"CertificateArn,omitempty"`
}

// ElasticLoadBalancingListenerCertificatesList represents a list of ElasticLoadBalancingListenerCertificates
type ElasticLoadBalancingListenerCertificatesList []ElasticLoadBalancingListenerCertificates

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticLoadBalancingListenerCertificatesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticLoadBalancingListenerCertificates{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticLoadBalancingListenerCertificatesList{item}
		return nil
	}
	list := []ElasticLoadBalancingListenerCertificates{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticLoadBalancingListenerCertificatesList(list)
		return nil
	}
	return err
}

// ElasticLoadBalancingListenerDefaultActions represents Elastic Load Balancing Listener DefaultActions
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-elasticloadbalancingv2-listener-defaultactions.html
type ElasticLoadBalancingListenerDefaultActions struct {
	// The Amazon Resource Name (ARN) of the target group to which Elastic
	// Load Balancing routes the traffic.
	TargetGroupArn *StringExpr `json:"TargetGroupArn,omitempty"`

	// The type of action. For valid values, see the Type contents for the
	// Action data type in the Elastic Load Balancing API Reference version
	// 2015-12-01.
	Type *StringExpr `json:"Type,omitempty"`
}

// ElasticLoadBalancingListenerDefaultActionsList represents a list of ElasticLoadBalancingListenerDefaultActions
type ElasticLoadBalancingListenerDefaultActionsList []ElasticLoadBalancingListenerDefaultActions

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticLoadBalancingListenerDefaultActionsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticLoadBalancingListenerDefaultActions{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticLoadBalancingListenerDefaultActionsList{item}
		return nil
	}
	list := []ElasticLoadBalancingListenerDefaultActions{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticLoadBalancingListenerDefaultActionsList(list)
		return nil
	}
	return err
}

// ElasticLoadBalancingListenerRuleActions represents Elastic Load Balancing ListenerRule Actions
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-elasticloadbalancingv2-listenerrule-actions.html
type ElasticLoadBalancingListenerRuleActions struct {
	// The Amazon Resource Name (ARN) of the target group to which Elastic
	// Load Balancing routes the traffic.
	TargetGroupArn *StringExpr `json:"TargetGroupArn,omitempty"`

	// The type of action. For valid values, see the Type contents for the
	// Action data type in the Elastic Load Balancing API Reference version
	// 2015-12-01.
	Type *StringExpr `json:"Type,omitempty"`
}

// ElasticLoadBalancingListenerRuleActionsList represents a list of ElasticLoadBalancingListenerRuleActions
type ElasticLoadBalancingListenerRuleActionsList []ElasticLoadBalancingListenerRuleActions

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticLoadBalancingListenerRuleActionsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticLoadBalancingListenerRuleActions{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticLoadBalancingListenerRuleActionsList{item}
		return nil
	}
	list := []ElasticLoadBalancingListenerRuleActions{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticLoadBalancingListenerRuleActionsList(list)
		return nil
	}
	return err
}

// ElasticLoadBalancingListenerRuleConditions represents Elastic Load Balancing ListenerRule Conditions
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-elasticloadbalancingv2-listenerrule-conditions.html
type ElasticLoadBalancingListenerRuleConditions struct {
	// The name of the condition that you want to define, such as
	// path-pattern (which forwards requests based on the URL of the
	// request).
	Field *StringExpr `json:"Field,omitempty"`

	// The value for the field that you specified in the Field property.
	Values *StringListExpr `json:"Values,omitempty"`
}

// ElasticLoadBalancingListenerRuleConditionsList represents a list of ElasticLoadBalancingListenerRuleConditions
type ElasticLoadBalancingListenerRuleConditionsList []ElasticLoadBalancingListenerRuleConditions

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticLoadBalancingListenerRuleConditionsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticLoadBalancingListenerRuleConditions{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticLoadBalancingListenerRuleConditionsList{item}
		return nil
	}
	list := []ElasticLoadBalancingListenerRuleConditions{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticLoadBalancingListenerRuleConditionsList(list)
		return nil
	}
	return err
}

// ElasticLoadBalancingLoadBalancerLoadBalancerAttributes represents Elastic Load Balancing LoadBalancer LoadBalancerAttributes
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-elasticloadbalancingv2-loadbalancer-loadbalancerattributes.html
type ElasticLoadBalancingLoadBalancerLoadBalancerAttributes struct {
	// The name of an attribute that you want to configure. For the list of
	// attributes that you can configure, see the Key contents for the
	// LoadBalancerAttribute data type in the Elastic Load Balancing API
	// Reference version 2015-12-01.
	Key *StringExpr `json:"Key,omitempty"`

	// A value for the attribute.
	Value *StringExpr `json:"Value,omitempty"`
}

// ElasticLoadBalancingLoadBalancerLoadBalancerAttributesList represents a list of ElasticLoadBalancingLoadBalancerLoadBalancerAttributes
type ElasticLoadBalancingLoadBalancerLoadBalancerAttributesList []ElasticLoadBalancingLoadBalancerLoadBalancerAttributes

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticLoadBalancingLoadBalancerLoadBalancerAttributesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticLoadBalancingLoadBalancerLoadBalancerAttributes{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticLoadBalancingLoadBalancerLoadBalancerAttributesList{item}
		return nil
	}
	list := []ElasticLoadBalancingLoadBalancerLoadBalancerAttributes{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticLoadBalancingLoadBalancerLoadBalancerAttributesList(list)
		return nil
	}
	return err
}

// ElasticLoadBalancingTargetGroupMatcher represents Elastic Load Balancing TargetGroup Matcher
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-elasticloadbalancingv2-targetgroup-matcher.html
type ElasticLoadBalancingTargetGroupMatcher struct {
	// The HTTP codes that a healthy target must use when responding to a
	// health check, such as 200,202 or 200-399.
	HttpCode *StringExpr `json:"HttpCode,omitempty"`
}

// ElasticLoadBalancingTargetGroupMatcherList represents a list of ElasticLoadBalancingTargetGroupMatcher
type ElasticLoadBalancingTargetGroupMatcherList []ElasticLoadBalancingTargetGroupMatcher

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticLoadBalancingTargetGroupMatcherList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticLoadBalancingTargetGroupMatcher{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticLoadBalancingTargetGroupMatcherList{item}
		return nil
	}
	list := []ElasticLoadBalancingTargetGroupMatcher{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticLoadBalancingTargetGroupMatcherList(list)
		return nil
	}
	return err
}

// ElasticLoadBalancingTargetGroupTargetDescription represents Elastic Load Balancing TargetGroup TargetDescription
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-elasticloadbalancingv2-targetgroup-targetdescription.html
type ElasticLoadBalancingTargetGroupTargetDescription struct {
	// The ID of the target, such as an EC2 instance ID.
	Id *StringExpr `json:"Id,omitempty"`

	// The port number on which the target is listening for traffic.
	Port *IntegerExpr `json:"Port,omitempty"`
}

// ElasticLoadBalancingTargetGroupTargetDescriptionList represents a list of ElasticLoadBalancingTargetGroupTargetDescription
type ElasticLoadBalancingTargetGroupTargetDescriptionList []ElasticLoadBalancingTargetGroupTargetDescription

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticLoadBalancingTargetGroupTargetDescriptionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticLoadBalancingTargetGroupTargetDescription{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticLoadBalancingTargetGroupTargetDescriptionList{item}
		return nil
	}
	list := []ElasticLoadBalancingTargetGroupTargetDescription{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticLoadBalancingTargetGroupTargetDescriptionList(list)
		return nil
	}
	return err
}

// ElasticLoadBalancingTargetGroupTargetGroupAttributes represents Elastic Load Balancing TargetGroup TargetGroupAttributes
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-elasticloadbalancingv2-targetgroup-targetgroupattributes.html
type ElasticLoadBalancingTargetGroupTargetGroupAttributes struct {
	// The name of the attribute that you want to configure. For the list of
	// attributes that you can configure, see the Key contents for the
	// TargetGroupAttribute data type in the Elastic Load Balancing API
	// Reference version 2015-12-01.
	Key *StringExpr `json:"Key,omitempty"`

	// A value for the attribute.
	Value *StringExpr `json:"Value,omitempty"`
}

// ElasticLoadBalancingTargetGroupTargetGroupAttributesList represents a list of ElasticLoadBalancingTargetGroupTargetGroupAttributes
type ElasticLoadBalancingTargetGroupTargetGroupAttributesList []ElasticLoadBalancingTargetGroupTargetGroupAttributes

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticLoadBalancingTargetGroupTargetGroupAttributesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticLoadBalancingTargetGroupTargetGroupAttributes{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticLoadBalancingTargetGroupTargetGroupAttributesList{item}
		return nil
	}
	list := []ElasticLoadBalancingTargetGroupTargetGroupAttributes{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticLoadBalancingTargetGroupTargetGroupAttributesList(list)
		return nil
	}
	return err
}

// ElasticsearchServiceDomainEBSOptions represents Amazon Elasticsearch Service Domain EBSOptions
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-elasticsearch-domain-ebsoptions.html
type ElasticsearchServiceDomainEBSOptions struct {
	// Specifies whether Amazon EBS volumes are attached to data nodes in the
	// Amazon ES domain.
	EBSEnabled *BoolExpr `json:"EBSEnabled,omitempty"`

	// The number of I/O operations per second (IOPS) that the volume
	// supports. This property applies only to the Provisioned IOPS (SSD) EBS
	// volume type.
	Iops *IntegerExpr `json:"Iops,omitempty"`

	// The size of the EBS volume for each data node. The minimum and maximum
	// size of an EBS volume depends on the EBS volume type and the instance
	// type to which it is attached. For more information, see Configuring
	// EBS-based Storage in the Amazon Elasticsearch Service Developer Guide.
	VolumeSize *IntegerExpr `json:"VolumeSize,omitempty"`

	// The EBS volume type to use with the Amazon ES domain, such as
	// standard, gp2, or io1. For more information about each type, see
	// Amazon EBS Volume Types in the Amazon EC2 User Guide for Linux
	// Instances.
	VolumeType *StringExpr `json:"VolumeType,omitempty"`
}

// ElasticsearchServiceDomainEBSOptionsList represents a list of ElasticsearchServiceDomainEBSOptions
type ElasticsearchServiceDomainEBSOptionsList []ElasticsearchServiceDomainEBSOptions

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticsearchServiceDomainEBSOptionsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticsearchServiceDomainEBSOptions{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticsearchServiceDomainEBSOptionsList{item}
		return nil
	}
	list := []ElasticsearchServiceDomainEBSOptions{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticsearchServiceDomainEBSOptionsList(list)
		return nil
	}
	return err
}

// ElasticsearchServiceDomainElasticsearchClusterConfig represents Amazon Elasticsearch Service Domain ElasticsearchClusterConfig
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-elasticsearch-domain-elasticsearchclusterconfig.html
type ElasticsearchServiceDomainElasticsearchClusterConfig struct {
	// The number of instances to use for the master node.
	DedicatedMasterCount *IntegerExpr `json:"DedicatedMasterCount,omitempty"`

	// Indicates whether to use a dedicated master node for the Amazon ES
	// domain. A dedicated master node is a cluster node that performs
	// cluster management tasks, but doesn't hold data or respond to data
	// upload requests. Dedicated master nodes offload cluster management
	// tasks to increase the stability of your search clusters.
	DedicatedMasterEnabled *BoolExpr `json:"DedicatedMasterEnabled,omitempty"`

	// The hardware configuration of the computer that hosts the dedicated
	// master node, such as m3.medium.elasticsearch. For valid values, see
	// Configuring Amazon ES Domains in the Amazon Elasticsearch Service
	// Developer Guide.
	DedicatedMasterType *StringExpr `json:"DedicatedMasterType,omitempty"`

	// The number of data nodes (instances) to use in the Amazon ES domain.
	InstanceCount *IntegerExpr `json:"InstanceCount,omitempty"`

	// The instance type for your data nodes, such as
	// m3.medium.elasticsearch. For valid values, see Configuring Amazon ES
	// Domains in the Amazon Elasticsearch Service Developer Guide.
	InstanceType *StringExpr `json:"InstanceType,omitempty"`

	// Indicates whether to enable zone awareness for the Amazon ES domain.
	// When you enable zone awareness, Amazon ES allocates the nodes and
	// replica index shards that belong to a cluster across two Availability
	// Zones (AZs) in the same region to prevent data loss and minimize
	// downtime in the event of node or data center failure. Don't enable
	// zone awareness if your cluster has no replica index shards or is a
	// single-node cluster. For more information, see Enabling Zone Awareness
	// in the Amazon Elasticsearch Service Developer Guide.
	ZoneAwarenessEnabled *BoolExpr `json:"ZoneAwarenessEnabled,omitempty"`
}

// ElasticsearchServiceDomainElasticsearchClusterConfigList represents a list of ElasticsearchServiceDomainElasticsearchClusterConfig
type ElasticsearchServiceDomainElasticsearchClusterConfigList []ElasticsearchServiceDomainElasticsearchClusterConfig

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticsearchServiceDomainElasticsearchClusterConfigList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticsearchServiceDomainElasticsearchClusterConfig{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticsearchServiceDomainElasticsearchClusterConfigList{item}
		return nil
	}
	list := []ElasticsearchServiceDomainElasticsearchClusterConfig{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticsearchServiceDomainElasticsearchClusterConfigList(list)
		return nil
	}
	return err
}

// ElasticsearchServiceDomainSnapshotOptions represents Amazon Elasticsearch Service Domain SnapshotOptions
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-elasticsearch-domain-snapshotoptions.html
type ElasticsearchServiceDomainSnapshotOptions struct {
	// The hour in UTC during which the service takes an automated daily
	// snapshot of the indices in the Amazon ES domain. For example, if you
	// specify 0, Amazon ES takes an automated snapshot everyday between
	// midnight and 1 am. You can specify a value between 0 and 23.
	AutomatedSnapshotStartHour *IntegerExpr `json:"AutomatedSnapshotStartHour,omitempty"`
}

// ElasticsearchServiceDomainSnapshotOptionsList represents a list of ElasticsearchServiceDomainSnapshotOptions
type ElasticsearchServiceDomainSnapshotOptionsList []ElasticsearchServiceDomainSnapshotOptions

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ElasticsearchServiceDomainSnapshotOptionsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ElasticsearchServiceDomainSnapshotOptions{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ElasticsearchServiceDomainSnapshotOptionsList{item}
		return nil
	}
	list := []ElasticsearchServiceDomainSnapshotOptions{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ElasticsearchServiceDomainSnapshotOptionsList(list)
		return nil
	}
	return err
}

// EMRClusterApplication represents Amazon EMR Cluster Application
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-emr-cluster-application.html
type EMRClusterApplication struct {
	// Metadata about third-party applications that third-party vendors use
	// for testing purposes.
	AdditionalInfo *StringExpr `json:"AdditionalInfo,omitempty"`

	// Arguments that Amazon EMR passes to the application.
	Args *StringListExpr `json:"Args,omitempty"`

	// The name of the application to add to your cluster, such as Hadoop or
	// Hive. For valid values, see the Applications parameter in the Amazon
	// EMR API Reference.
	Name *StringExpr `json:"Name,omitempty"`

	// The version of the application.
	Version *StringExpr `json:"Version,omitempty"`
}

// EMRClusterApplicationList represents a list of EMRClusterApplication
type EMRClusterApplicationList []EMRClusterApplication

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EMRClusterApplicationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EMRClusterApplication{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EMRClusterApplicationList{item}
		return nil
	}
	list := []EMRClusterApplication{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EMRClusterApplicationList(list)
		return nil
	}
	return err
}

// EMRClusterBootstrapActionConfig represents Amazon EMR Cluster BootstrapActionConfig
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-emr-cluster-bootstrapactionconfig.html
type EMRClusterBootstrapActionConfig struct {
	// The name of the bootstrap action to add to your cluster.
	Name *StringExpr `json:"Name,omitempty"`

	// The script that the bootstrap action runs.
	ScriptBootstrapAction *EMRClusterBootstrapActionConfigScriptBootstrapActionConfig `json:"ScriptBootstrapAction,omitempty"`
}

// EMRClusterBootstrapActionConfigList represents a list of EMRClusterBootstrapActionConfig
type EMRClusterBootstrapActionConfigList []EMRClusterBootstrapActionConfig

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EMRClusterBootstrapActionConfigList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EMRClusterBootstrapActionConfig{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EMRClusterBootstrapActionConfigList{item}
		return nil
	}
	list := []EMRClusterBootstrapActionConfig{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EMRClusterBootstrapActionConfigList(list)
		return nil
	}
	return err
}

// EMRClusterBootstrapActionConfigScriptBootstrapActionConfig represents Amazon EMR Cluster BootstrapActionConfig ScriptBootstrapActionConfig
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-emr-cluster-bootstrapactionconfig-scriptbootstrapactionconfig.html
type EMRClusterBootstrapActionConfigScriptBootstrapActionConfig struct {
	// A list of command line arguments to pass to the bootstrap action
	// script.
	Args *StringListExpr `json:"Args,omitempty"`

	// The location of the script that Amazon EMR runs during a bootstrap
	// action. Specify a location in an S3 bucket or your local file system.
	Path *StringExpr `json:"Path,omitempty"`
}

// EMRClusterBootstrapActionConfigScriptBootstrapActionConfigList represents a list of EMRClusterBootstrapActionConfigScriptBootstrapActionConfig
type EMRClusterBootstrapActionConfigScriptBootstrapActionConfigList []EMRClusterBootstrapActionConfigScriptBootstrapActionConfig

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EMRClusterBootstrapActionConfigScriptBootstrapActionConfigList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EMRClusterBootstrapActionConfigScriptBootstrapActionConfig{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EMRClusterBootstrapActionConfigScriptBootstrapActionConfigList{item}
		return nil
	}
	list := []EMRClusterBootstrapActionConfigScriptBootstrapActionConfig{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EMRClusterBootstrapActionConfigScriptBootstrapActionConfigList(list)
		return nil
	}
	return err
}

// EMRClusterConfiguration represents Amazon EMR Cluster Configuration
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-emr-cluster-configuration.html
type EMRClusterConfiguration struct {
	// The name of an application-specific configuration file. For more
	// information see, Configuring Applications in the Amazon EMR Release
	// Guide.
	Classification *StringExpr `json:"Classification,omitempty"`

	// The settings that you want to change in the application-specific
	// configuration file. For more information see, Configuring Applications
	// in the Amazon EMR Release Guide.
	ConfigurationProperties *StringExpr `json:"ConfigurationProperties,omitempty"`

	// A list of configurations to apply to this configuration. You can nest
	// configurations so that a single configuration can have its own
	// configurations. In other words, you can configure a configuration. For
	// more information see, Configuring Applications in the Amazon EMR
	// Release Guide.
	Configurations *EMRClusterConfigurationList `json:"Configurations,omitempty"`
}

// EMRClusterConfigurationList represents a list of EMRClusterConfiguration
type EMRClusterConfigurationList []EMRClusterConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EMRClusterConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EMRClusterConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EMRClusterConfigurationList{item}
		return nil
	}
	list := []EMRClusterConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EMRClusterConfigurationList(list)
		return nil
	}
	return err
}

// EMRClusterJobFlowInstancesConfig represents Amazon EMR Cluster JobFlowInstancesConfig
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-emr-cluster-jobflowinstancesconfig.html
type EMRClusterJobFlowInstancesConfig struct {
	// A list of additional EC2 security group IDs to assign to the master
	// instance (master node) in your Amazon EMR cluster. Use this property
	// to supplement the rules specified by the Amazon EMR managed master
	// security group.
	AdditionalMasterSecurityGroups *StringListExpr `json:"AdditionalMasterSecurityGroups,omitempty"`

	// A list of additional EC2 security group IDs to assign to the slave
	// instances (slave nodes) in your Amazon EMR cluster. Use this property
	// to supplement the rules specified by the Amazon EMR managed slave
	// security group.
	AdditionalSlaveSecurityGroups *StringListExpr `json:"AdditionalSlaveSecurityGroups,omitempty"`

	// The settings for the core instances in your Amazon EMR cluster.
	CoreInstanceGroup *EMRClusterJobFlowInstancesConfigInstanceGroupConfig `json:"CoreInstanceGroup,omitempty"`

	// The name of an Amazon Elastic Compute Cloud (Amazon EC2) key pair,
	// which you can use to access the instances in your Amazon EMR cluster.
	Ec2KeyName *StringExpr `json:"Ec2KeyName,omitempty"`

	// The ID of a subnet where you want to launch your instances.
	Ec2SubnetId *StringExpr `json:"Ec2SubnetId,omitempty"`

	// The ID of an EC2 security group (managed by Amazon EMR) that is
	// assigned to the master instance (master node) in your Amazon EMR
	// cluster.
	EmrManagedMasterSecurityGroup *StringExpr `json:"EmrManagedMasterSecurityGroup,omitempty"`

	// The ID of an EC2 security group (managed by Amazon EMR) that is
	// assigned to the slave instances (slave nodes) in your Amazon EMR
	// cluster.
	EmrManagedSlaveSecurityGroup *StringExpr `json:"EmrManagedSlaveSecurityGroup,omitempty"`

	// The Hadoop version for the job flow. For valid values, see the
	// HadoopVersion parameter in the Amazon EMR API Reference.
	HadoopVersion *StringExpr `json:"HadoopVersion,omitempty"`

	// The settings for the master instance (master node).
	MasterInstanceGroup *EMRClusterJobFlowInstancesConfigInstanceGroupConfig `json:"MasterInstanceGroup,omitempty"`

	// The Availability Zone (AZ) in which the job flow runs.
	Placement *EMRClusterJobFlowInstancesConfigPlacement `json:"Placement,omitempty"`

	// The ID of an EC2 security group (managed by Amazon EMR) that services
	// use to access clusters in private subnets.
	ServiceAccessSecurityGroup *StringExpr `json:"ServiceAccessSecurityGroup,omitempty"`

	// Indicates whether to prevent the EC2 instances from being terminated
	// by an API call or user intervention. If you want to delete a stack
	// with protected instances, update this value to false before you delete
	// the stack. By default, AWS CloudFormation sets this property to false.
	TerminationProtected *BoolExpr `json:"TerminationProtected,omitempty"`
}

// EMRClusterJobFlowInstancesConfigList represents a list of EMRClusterJobFlowInstancesConfig
type EMRClusterJobFlowInstancesConfigList []EMRClusterJobFlowInstancesConfig

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EMRClusterJobFlowInstancesConfigList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EMRClusterJobFlowInstancesConfig{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EMRClusterJobFlowInstancesConfigList{item}
		return nil
	}
	list := []EMRClusterJobFlowInstancesConfig{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EMRClusterJobFlowInstancesConfigList(list)
		return nil
	}
	return err
}

// EMRClusterJobFlowInstancesConfigInstanceGroupConfig represents Amazon EMR Cluster JobFlowInstancesConfig InstanceGroupConfig
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-emr-cluster-jobflowinstancesconfig-instancegroupconfig.html
type EMRClusterJobFlowInstancesConfigInstanceGroupConfig struct {
	// When launching instances as Spot Instances, the bid price in USD for
	// each EC2 instance in the instance group.
	BidPrice *StringExpr `json:"BidPrice,omitempty"`

	// A list of configurations to apply to this instance group. For more
	// information see, Configuring Applications in the Amazon EMR Release
	// Guide.
	Configurations *EMRClusterConfigurationList `json:"Configurations,omitempty"`

	// Configures Amazon Elastic Block Store (Amazon EBS) storage volumes to
	// attach to your instances.
	EbsConfiguration *EMREbsConfiguration `json:"EbsConfiguration,omitempty"`

	// The number of instances to launch in the instance group.
	InstanceCount *IntegerExpr `json:"InstanceCount,omitempty"`

	// The EC2 instance type for all instances in the instance group. For
	// more information, see Instance Configurations in the Amazon EMR
	// Management Guide.
	InstanceType *StringExpr `json:"InstanceType,omitempty"`

	// The type of marketplace from which your instances are provisioned into
	// this group, either ON_DEMAND or SPOT. For more information, see Amazon
	// EC2 Purchasing Options.
	Market *StringExpr `json:"Market,omitempty"`

	// A name for the instance group.
	Name *StringExpr `json:"Name,omitempty"`
}

// EMRClusterJobFlowInstancesConfigInstanceGroupConfigList represents a list of EMRClusterJobFlowInstancesConfigInstanceGroupConfig
type EMRClusterJobFlowInstancesConfigInstanceGroupConfigList []EMRClusterJobFlowInstancesConfigInstanceGroupConfig

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EMRClusterJobFlowInstancesConfigInstanceGroupConfigList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EMRClusterJobFlowInstancesConfigInstanceGroupConfig{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EMRClusterJobFlowInstancesConfigInstanceGroupConfigList{item}
		return nil
	}
	list := []EMRClusterJobFlowInstancesConfigInstanceGroupConfig{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EMRClusterJobFlowInstancesConfigInstanceGroupConfigList(list)
		return nil
	}
	return err
}

// EMRClusterJobFlowInstancesConfigPlacement represents Amazon EMR Cluster JobFlowInstancesConfig PlacementType
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-emr-cluster-jobflowinstancesconfig-placementtype.html
type EMRClusterJobFlowInstancesConfigPlacement struct {
	// The Amazon Elastic Compute Cloud (Amazon EC2) AZ for the job flow. For
	// more information, see
	// http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html
	// in the Amazon EC2 User Guide for Linux Instances.
	AvailabilityZone *StringExpr `json:"AvailabilityZone,omitempty"`
}

// EMRClusterJobFlowInstancesConfigPlacementList represents a list of EMRClusterJobFlowInstancesConfigPlacement
type EMRClusterJobFlowInstancesConfigPlacementList []EMRClusterJobFlowInstancesConfigPlacement

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EMRClusterJobFlowInstancesConfigPlacementList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EMRClusterJobFlowInstancesConfigPlacement{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EMRClusterJobFlowInstancesConfigPlacementList{item}
		return nil
	}
	list := []EMRClusterJobFlowInstancesConfigPlacement{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EMRClusterJobFlowInstancesConfigPlacementList(list)
		return nil
	}
	return err
}

// EMREbsConfiguration represents Amazon EMR EbsConfiguration
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-emr-ebsconfiguration.html
type EMREbsConfiguration struct {
	// Configures the block storage devices that are associated with your EMR
	// instances.
	EbsBlockDeviceConfigs *EMREbsConfigurationEbsBlockDeviceConfigsList `json:"EbsBlockDeviceConfigs,omitempty"`

	// Indicates whether the instances are optimized for Amazon EBS I/O. This
	// optimization provides dedicated throughput to Amazon EBS and an
	// optimized configuration stack to provide optimal EBS I/O performance.
	// For more information about fees and supported instance types, see
	// EBS-Optimized Instances in the Amazon EC2 User Guide for Linux
	// Instances.
	EbsOptimized *BoolExpr `json:"EbsOptimized,omitempty"`
}

// EMREbsConfigurationList represents a list of EMREbsConfiguration
type EMREbsConfigurationList []EMREbsConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EMREbsConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EMREbsConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EMREbsConfigurationList{item}
		return nil
	}
	list := []EMREbsConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EMREbsConfigurationList(list)
		return nil
	}
	return err
}

// EMREbsConfigurationEbsBlockDeviceConfigs represents Amazon EMR EbsConfiguration EbsBlockDeviceConfigs
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-emr-ebsconfiguration-ebsblockdeviceconfig.html
type EMREbsConfigurationEbsBlockDeviceConfigs struct {
	// The settings for the Amazon EBS volumes.
	VolumeSpecification *EMREbsConfigurationEbsBlockDeviceConfigVolumeSpecification `json:"VolumeSpecification,omitempty"`

	// The number of Amazon EBS volumes that you want to create for each
	// instance in the EMR cluster or instance group.
	VolumesPerInstance *IntegerExpr `json:"VolumesPerInstance,omitempty"`
}

// EMREbsConfigurationEbsBlockDeviceConfigsList represents a list of EMREbsConfigurationEbsBlockDeviceConfigs
type EMREbsConfigurationEbsBlockDeviceConfigsList []EMREbsConfigurationEbsBlockDeviceConfigs

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EMREbsConfigurationEbsBlockDeviceConfigsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EMREbsConfigurationEbsBlockDeviceConfigs{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EMREbsConfigurationEbsBlockDeviceConfigsList{item}
		return nil
	}
	list := []EMREbsConfigurationEbsBlockDeviceConfigs{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EMREbsConfigurationEbsBlockDeviceConfigsList(list)
		return nil
	}
	return err
}

// EMREbsConfigurationEbsBlockDeviceConfigVolumeSpecification represents Amazon EMR EbsConfiguration EbsBlockDeviceConfig VolumeSpecification
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-emr-ebsconfiguration-ebsblockdeviceconfig-volumespecification.html
type EMREbsConfigurationEbsBlockDeviceConfigVolumeSpecification struct {
	// The number of I/O operations per second (IOPS) that the volume
	// supports. For more information, see Iops for the EbsBlockDevice action
	// in the Amazon EC2 API Reference.
	Iops *IntegerExpr `json:"Iops,omitempty"`

	// The volume size, in Gibibytes (GiB). For more information about
	// specifying the volume size, see VolumeSize for the EbsBlockDevice
	// action in the Amazon EC2 API Reference.
	SizeInGB *IntegerExpr `json:"SizeInGB,omitempty"`

	// The volume type, such as standard or io1. For more information about
	// specifying the volume type, see VolumeType for the EbsBlockDevice
	// action in the Amazon EC2 API Reference.
	VolumeType *StringExpr `json:"VolumeType,omitempty"`
}

// EMREbsConfigurationEbsBlockDeviceConfigVolumeSpecificationList represents a list of EMREbsConfigurationEbsBlockDeviceConfigVolumeSpecification
type EMREbsConfigurationEbsBlockDeviceConfigVolumeSpecificationList []EMREbsConfigurationEbsBlockDeviceConfigVolumeSpecification

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EMREbsConfigurationEbsBlockDeviceConfigVolumeSpecificationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EMREbsConfigurationEbsBlockDeviceConfigVolumeSpecification{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EMREbsConfigurationEbsBlockDeviceConfigVolumeSpecificationList{item}
		return nil
	}
	list := []EMREbsConfigurationEbsBlockDeviceConfigVolumeSpecification{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EMREbsConfigurationEbsBlockDeviceConfigVolumeSpecificationList(list)
		return nil
	}
	return err
}

// EMRStepHadoopJarStepConfig represents Amazon EMR Step HadoopJarStepConfig
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-emr-step-hadoopjarstepconfig.html
type EMRStepHadoopJarStepConfig struct {
	// A list of command line arguments passed to the JAR file's main
	// function when the function is executed.
	Args *StringListExpr `json:"Args,omitempty"`

	// A path to the JAR file that Amazon EMR runs for the job flow step.
	Jar *StringExpr `json:"Jar,omitempty"`

	// The name of the main class in the specified JAR file. If you don't
	// specify a value, you must specify a main class in the JAR file's
	// manifest file.
	MainClass *StringExpr `json:"MainClass,omitempty"`

	// A list of Java properties that are set when the job flow step runs.
	// You can use these properties to pass key-value pairs to your main
	// function in the JAR file.
	StepProperties *EMRStepHadoopJarStepConfigKeyValueList `json:"StepProperties,omitempty"`
}

// EMRStepHadoopJarStepConfigList represents a list of EMRStepHadoopJarStepConfig
type EMRStepHadoopJarStepConfigList []EMRStepHadoopJarStepConfig

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EMRStepHadoopJarStepConfigList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EMRStepHadoopJarStepConfig{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EMRStepHadoopJarStepConfigList{item}
		return nil
	}
	list := []EMRStepHadoopJarStepConfig{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EMRStepHadoopJarStepConfigList(list)
		return nil
	}
	return err
}

// EMRStepHadoopJarStepConfigKeyValue represents Amazon EMR Step HadoopJarStepConfig KeyValue
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-emr-step-hadoopjarstepconfig-keyvalue.html
type EMRStepHadoopJarStepConfigKeyValue struct {
	// The unique identifier of a key-value pair.
	Key *StringExpr `json:"Key,omitempty"`

	// The value part of the identified key.
	Value *StringExpr `json:"Value,omitempty"`
}

// EMRStepHadoopJarStepConfigKeyValueList represents a list of EMRStepHadoopJarStepConfigKeyValue
type EMRStepHadoopJarStepConfigKeyValueList []EMRStepHadoopJarStepConfigKeyValue

// UnmarshalJSON sets the object from the provided JSON representation
func (l *EMRStepHadoopJarStepConfigKeyValueList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := EMRStepHadoopJarStepConfigKeyValue{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = EMRStepHadoopJarStepConfigKeyValueList{item}
		return nil
	}
	list := []EMRStepHadoopJarStepConfigKeyValue{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = EMRStepHadoopJarStepConfigKeyValueList(list)
		return nil
	}
	return err
}

// GameLiftAliasRoutingStrategy represents Amazon GameLift Alias RoutingStrategy
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-gamelift-alias-routingstrategy.html
type GameLiftAliasRoutingStrategy struct {
	// A unique identifier of a GameLift fleet to associate with the alias.
	FleetId *StringExpr `json:"FleetId,omitempty"`

	// A text message that GameLift displays for the Terminal routing type.
	Message *StringExpr `json:"Message,omitempty"`

	// The type of routing strategy. For the SIMPLE type, traffic is routed
	// to an active GameLift fleet. For the Terminal type, GameLift returns
	// an exception with the message that you specified in the Message
	// property.
	Type *StringExpr `json:"Type,omitempty"`
}

// GameLiftAliasRoutingStrategyList represents a list of GameLiftAliasRoutingStrategy
type GameLiftAliasRoutingStrategyList []GameLiftAliasRoutingStrategy

// UnmarshalJSON sets the object from the provided JSON representation
func (l *GameLiftAliasRoutingStrategyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := GameLiftAliasRoutingStrategy{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = GameLiftAliasRoutingStrategyList{item}
		return nil
	}
	list := []GameLiftAliasRoutingStrategy{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = GameLiftAliasRoutingStrategyList(list)
		return nil
	}
	return err
}

// GameLiftBuildStorageLocation represents Amazon GameLift Build StorageLocation
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-gamelift-build-storagelocation.html
type GameLiftBuildStorageLocation struct {
	// The S3 bucket where the GameLift build package files are stored.
	Bucket *StringExpr `json:"Bucket,omitempty"`

	// The prefix (folder name) where the GameLift build package files are
	// located.
	Key *StringExpr `json:"Key,omitempty"`

	// An AWS Identity and Access Management (IAM) role Amazon Resource Name
	// (ARN) that GameLift can assume to retrieve the build package files
	// from Amazon Simple Storage Service (Amazon S3).
	RoleArn *StringExpr `json:"RoleArn,omitempty"`
}

// GameLiftBuildStorageLocationList represents a list of GameLiftBuildStorageLocation
type GameLiftBuildStorageLocationList []GameLiftBuildStorageLocation

// UnmarshalJSON sets the object from the provided JSON representation
func (l *GameLiftBuildStorageLocationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := GameLiftBuildStorageLocation{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = GameLiftBuildStorageLocationList{item}
		return nil
	}
	list := []GameLiftBuildStorageLocation{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = GameLiftBuildStorageLocationList(list)
		return nil
	}
	return err
}

// GameLiftFleetEC2InboundPermission represents Amazon GameLift Fleet EC2InboundPermission
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-gamelift-fleet-ec2inboundpermission.html
type GameLiftFleetEC2InboundPermission struct {
	// The starting value for a range of allowed port numbers. This value
	// must be lower than the ToPort value.
	FromPort *IntegerExpr `json:"FromPort,omitempty"`

	// The range of allowed IP addresses in CIDR notation.
	IpRange *StringExpr `json:"IpRange,omitempty"`

	// The network communication protocol that is used by the fleet. For
	// valid values, see the IpPermission data type in the Amazon GameLift
	// API Reference.
	Protocol *StringExpr `json:"Protocol,omitempty"`

	// The ending value for a range of allowed port numbers. This value must
	// be higher than the FromPort value.
	ToPort *IntegerExpr `json:"ToPort,omitempty"`
}

// GameLiftFleetEC2InboundPermissionList represents a list of GameLiftFleetEC2InboundPermission
type GameLiftFleetEC2InboundPermissionList []GameLiftFleetEC2InboundPermission

// UnmarshalJSON sets the object from the provided JSON representation
func (l *GameLiftFleetEC2InboundPermissionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := GameLiftFleetEC2InboundPermission{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = GameLiftFleetEC2InboundPermissionList{item}
		return nil
	}
	list := []GameLiftFleetEC2InboundPermission{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = GameLiftFleetEC2InboundPermissionList(list)
		return nil
	}
	return err
}

// IAMPolicies represents IAM Policies
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-iam-policy.html
type IAMPolicies struct {
	// A policy document that describes what actions are allowed on which
	// resources.
	PolicyDocument interface{} `json:"PolicyDocument,omitempty"`

	// The name of the policy.
	PolicyName *StringExpr `json:"PolicyName,omitempty"`
}

// IAMPoliciesList represents a list of IAMPolicies
type IAMPoliciesList []IAMPolicies

// UnmarshalJSON sets the object from the provided JSON representation
func (l *IAMPoliciesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := IAMPolicies{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = IAMPoliciesList{item}
		return nil
	}
	list := []IAMPolicies{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = IAMPoliciesList(list)
		return nil
	}
	return err
}

// IAMUserLoginProfile represents IAM User LoginProfile
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-iam-user-loginprofile.html
type IAMUserLoginProfile struct {
	// The password for the user.
	Password *StringExpr `json:"Password,omitempty"`

	// Specifies whether the user is required to set a new password the next
	// time the user logs in to the AWS Management Console.
	PasswordResetRequired *BoolExpr `json:"PasswordResetRequired,omitempty"`
}

// IAMUserLoginProfileList represents a list of IAMUserLoginProfile
type IAMUserLoginProfileList []IAMUserLoginProfile

// UnmarshalJSON sets the object from the provided JSON representation
func (l *IAMUserLoginProfileList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := IAMUserLoginProfile{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = IAMUserLoginProfileList{item}
		return nil
	}
	list := []IAMUserLoginProfile{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = IAMUserLoginProfileList(list)
		return nil
	}
	return err
}

// IoTActions represents AWS IoT Actions
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-iot-actions.html
type IoTActions struct {
	// Changes the state of a CloudWatch alarm.
	CloudwatchAlarm *IoTCloudwatchAlarmAction `json:"CloudwatchAlarm,omitempty"`

	// Captures a CloudWatch metric.
	CloudwatchMetric *IoTCloudwatchMetricAction `json:"CloudwatchMetric,omitempty"`

	// Writes data to a DynamoDB table.
	DynamoDB *IoTDynamoDBAction `json:"DynamoDB,omitempty"`

	// Writes data to an Elasticsearch domain.
	Elasticsearch *IoTElasticsearchAction `json:"Elasticsearch,omitempty"`

	// Writes data to a Firehose stream.
	Firehose *IoTFirehoseAction `json:"Firehose,omitempty"`

	// Writes data to an Amazon Kinesis stream.
	Kinesis *IoTKinesisAction `json:"Kinesis,omitempty"`

	// Invokes a Lambda function.
	Lambda *IoTLambdaAction `json:"Lambda,omitempty"`

	// Publishes data to an MQ Telemetry Transport (MQTT) topic different
	// from the one currently specified.
	Republish *IoTRepublishAction `json:"Republish,omitempty"`

	// Writes data to an S3 bucket.
	S3 *IoTS3Action `json:"S3,omitempty"`

	// Publishes data to an SNS topic.
	Sns *IoTSnsAction `json:"Sns,omitempty"`

	// Publishes data to an SQS queue.
	Sqs *IoTSqsAction `json:"Sqs,omitempty"`
}

// IoTActionsList represents a list of IoTActions
type IoTActionsList []IoTActions

// UnmarshalJSON sets the object from the provided JSON representation
func (l *IoTActionsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := IoTActions{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = IoTActionsList{item}
		return nil
	}
	list := []IoTActions{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = IoTActionsList(list)
		return nil
	}
	return err
}

// IoTCloudwatchAlarmAction represents AWS IoT CloudwatchAlarm Action
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-iot-cloudwatchalarm.html
type IoTCloudwatchAlarmAction struct {
	// The CloudWatch alarm name.
	AlarmName *StringExpr `json:"AlarmName,omitempty"`

	// The IAM role that allows access to the CloudWatch alarm.
	RoleArn *StringExpr `json:"RoleArn,omitempty"`

	// The reason for the change of the alarm state.
	StateReason *StringExpr `json:"StateReason,omitempty"`

	// The value of the alarm state.
	StateValue *StringExpr `json:"StateValue,omitempty"`
}

// IoTCloudwatchAlarmActionList represents a list of IoTCloudwatchAlarmAction
type IoTCloudwatchAlarmActionList []IoTCloudwatchAlarmAction

// UnmarshalJSON sets the object from the provided JSON representation
func (l *IoTCloudwatchAlarmActionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := IoTCloudwatchAlarmAction{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = IoTCloudwatchAlarmActionList{item}
		return nil
	}
	list := []IoTCloudwatchAlarmAction{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = IoTCloudwatchAlarmActionList(list)
		return nil
	}
	return err
}

// IoTCloudwatchMetricAction represents AWS IoT CloudwatchMetric Action
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-iot-cloudwatchmetric.html
type IoTCloudwatchMetricAction struct {
	// The name of the CloudWatch metric.
	MetricName *StringExpr `json:"MetricName,omitempty"`

	// The name of the CloudWatch metric namespace.
	MetricNamespace *StringExpr `json:"MetricNamespace,omitempty"`

	// An optional Unix timestamp.
	MetricTimestamp *StringExpr `json:"MetricTimestamp,omitempty"`

	// The metric unit supported by Amazon CloudWatch.
	MetricUnit *StringExpr `json:"MetricUnit,omitempty"`

	// The value to publish to the metric. For example, if you count the
	// occurrences of a particular term such as Error, the value will be 1
	// for each occurrence.
	MetricValue *StringExpr `json:"MetricValue,omitempty"`

	// The ARN of the IAM role that grants access to the CloudWatch metric.
	RoleArn *StringExpr `json:"RoleArn,omitempty"`
}

// IoTCloudwatchMetricActionList represents a list of IoTCloudwatchMetricAction
type IoTCloudwatchMetricActionList []IoTCloudwatchMetricAction

// UnmarshalJSON sets the object from the provided JSON representation
func (l *IoTCloudwatchMetricActionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := IoTCloudwatchMetricAction{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = IoTCloudwatchMetricActionList{item}
		return nil
	}
	list := []IoTCloudwatchMetricAction{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = IoTCloudwatchMetricActionList(list)
		return nil
	}
	return err
}

// IoTDynamoDBAction represents AWS IoT DynamoDB Action
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-iot-dynamodb.html
type IoTDynamoDBAction struct {
	// The name of the hash key.
	HashKeyField *StringExpr `json:"HashKeyField,omitempty"`

	// The value of the hash key.
	HashKeyValue *StringExpr `json:"HashKeyValue,omitempty"`

	// The name of the column in the DynamoDB table that contains the result
	// of the query. You can customize this name.
	PayloadField *StringExpr `json:"PayloadField,omitempty"`

	// The name of the range key.
	RangeKeyField *StringExpr `json:"RangeKeyField,omitempty"`

	// The value of the range key.
	RangeKeyValue *StringExpr `json:"RangeKeyValue,omitempty"`

	// The ARN of the IAM role that grants access to the DynamoDB table.
	RoleArn *StringExpr `json:"RoleArn,omitempty"`

	// The name of the DynamoDB table.
	TableName *StringExpr `json:"TableName,omitempty"`
}

// IoTDynamoDBActionList represents a list of IoTDynamoDBAction
type IoTDynamoDBActionList []IoTDynamoDBAction

// UnmarshalJSON sets the object from the provided JSON representation
func (l *IoTDynamoDBActionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := IoTDynamoDBAction{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = IoTDynamoDBActionList{item}
		return nil
	}
	list := []IoTDynamoDBAction{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = IoTDynamoDBActionList(list)
		return nil
	}
	return err
}

// IoTElasticsearchAction represents AWS IoT Elasticsearch Action
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-iot-elasticsearch.html
type IoTElasticsearchAction struct {
	// The endpoint of your Elasticsearch domain.
	Endpoint *StringExpr `json:"Endpoint,omitempty"`

	// A unique identifier for the stored data.
	Id *StringExpr `json:"Id,omitempty"`

	// The Elasticsearch index where the data is stored.
	Index *StringExpr `json:"Index,omitempty"`

	// The ARN of the IAM role that grants access to Elasticsearch.
	RoleArn *StringExpr `json:"RoleArn,omitempty"`

	// The type of stored data.
	Type *StringExpr `json:"Type,omitempty"`
}

// IoTElasticsearchActionList represents a list of IoTElasticsearchAction
type IoTElasticsearchActionList []IoTElasticsearchAction

// UnmarshalJSON sets the object from the provided JSON representation
func (l *IoTElasticsearchActionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := IoTElasticsearchAction{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = IoTElasticsearchActionList{item}
		return nil
	}
	list := []IoTElasticsearchAction{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = IoTElasticsearchActionList(list)
		return nil
	}
	return err
}

// IoTFirehoseAction represents AWS IoT Firehose Action
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-iot-firehose.html
type IoTFirehoseAction struct {
	// The delivery stream name.
	DeliveryStreamName *StringExpr `json:"DeliveryStreamName,omitempty"`

	// The ARN of the IAM role that grants access to the Firehose stream.
	RoleArn *StringExpr `json:"RoleArn,omitempty"`
}

// IoTFirehoseActionList represents a list of IoTFirehoseAction
type IoTFirehoseActionList []IoTFirehoseAction

// UnmarshalJSON sets the object from the provided JSON representation
func (l *IoTFirehoseActionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := IoTFirehoseAction{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = IoTFirehoseActionList{item}
		return nil
	}
	list := []IoTFirehoseAction{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = IoTFirehoseActionList(list)
		return nil
	}
	return err
}

// IoTKinesisAction represents AWS IoT Kinesis Action
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-iot-kinesis.html
type IoTKinesisAction struct {
	// The partition key (the grouping of data by shard within an an Amazon
	// Kinesis stream).
	PartitionKey *StringExpr `json:"PartitionKey,omitempty"`

	// The ARN of the IAM role that grants access to an Amazon Kinesis
	// stream.
	RoleArn *StringExpr `json:"RoleArn,omitempty"`

	// The name of the Amazon Kinesis stream.
	StreamName *StringExpr `json:"StreamName,omitempty"`
}

// IoTKinesisActionList represents a list of IoTKinesisAction
type IoTKinesisActionList []IoTKinesisAction

// UnmarshalJSON sets the object from the provided JSON representation
func (l *IoTKinesisActionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := IoTKinesisAction{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = IoTKinesisActionList{item}
		return nil
	}
	list := []IoTKinesisAction{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = IoTKinesisActionList(list)
		return nil
	}
	return err
}

// IoTLambdaAction represents AWS IoT Lambda Action
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-iot-lambda.html
type IoTLambdaAction struct {
	// The ARN of the Lambda function.
	FunctionArn *StringExpr `json:"FunctionArn,omitempty"`
}

// IoTLambdaActionList represents a list of IoTLambdaAction
type IoTLambdaActionList []IoTLambdaAction

// UnmarshalJSON sets the object from the provided JSON representation
func (l *IoTLambdaActionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := IoTLambdaAction{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = IoTLambdaActionList{item}
		return nil
	}
	list := []IoTLambdaAction{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = IoTLambdaActionList(list)
		return nil
	}
	return err
}

// IoTRepublishAction represents AWS IoT Republish Action
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-iot-republish.html
type IoTRepublishAction struct {
	// The ARN of the IAM role that grants publishing access.
	RoleArn *StringExpr `json:"RoleArn,omitempty"`

	// The name of the MQTT topic topic different from the one currently
	// specified.
	Topic *StringExpr `json:"Topic,omitempty"`
}

// IoTRepublishActionList represents a list of IoTRepublishAction
type IoTRepublishActionList []IoTRepublishAction

// UnmarshalJSON sets the object from the provided JSON representation
func (l *IoTRepublishActionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := IoTRepublishAction{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = IoTRepublishActionList{item}
		return nil
	}
	list := []IoTRepublishAction{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = IoTRepublishActionList(list)
		return nil
	}
	return err
}

// IoTS3Action represents AWS IoT S3 Action
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-iot-s3.html
type IoTS3Action struct {
	// The name of the S3 bucket.
	BucketName *StringExpr `json:"BucketName,omitempty"`

	// The object key (the name of an object in the S3 bucket).
	Key *StringExpr `json:"Key,omitempty"`

	// The ARN of the IAM role that grants access to Amazon S3.
	RoleArn *StringExpr `json:"RoleArn,omitempty"`
}

// IoTS3ActionList represents a list of IoTS3Action
type IoTS3ActionList []IoTS3Action

// UnmarshalJSON sets the object from the provided JSON representation
func (l *IoTS3ActionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := IoTS3Action{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = IoTS3ActionList{item}
		return nil
	}
	list := []IoTS3Action{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = IoTS3ActionList(list)
		return nil
	}
	return err
}

// IoTSnsAction represents AWS IoT Sns Action
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-iot-sns.html
type IoTSnsAction struct {
	// The format of the published message. Amazon SNS uses this setting to
	// determine whether it should parse the payload and extract the
	// platform-specific bits from the payload.
	MessageFormat *StringExpr `json:"MessageFormat,omitempty"`

	// The ARN of the IAM role that grants access to Amazon SNS.
	RoleArn *StringExpr `json:"RoleArn,omitempty"`

	// The ARN of the Amazon SNS topic.
	TargetArn *StringExpr `json:"TargetArn,omitempty"`
}

// IoTSnsActionList represents a list of IoTSnsAction
type IoTSnsActionList []IoTSnsAction

// UnmarshalJSON sets the object from the provided JSON representation
func (l *IoTSnsActionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := IoTSnsAction{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = IoTSnsActionList{item}
		return nil
	}
	list := []IoTSnsAction{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = IoTSnsActionList(list)
		return nil
	}
	return err
}

// IoTSqsAction represents AWS IoT Sqs Action
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-iot-sqs.html
type IoTSqsAction struct {
	// The URL of the Amazon Simple Queue Service (Amazon SQS) queue.
	QueueUrl *StringExpr `json:"QueueUrl,omitempty"`

	// The ARN of the IAM role that grants access to Amazon SQS.
	RoleArn *StringExpr `json:"RoleArn,omitempty"`

	// Specifies whether Base64 encoding should be used.
	UseBase64 *StringExpr `json:"UseBase64,omitempty"`
}

// IoTSqsActionList represents a list of IoTSqsAction
type IoTSqsActionList []IoTSqsAction

// UnmarshalJSON sets the object from the provided JSON representation
func (l *IoTSqsActionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := IoTSqsAction{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = IoTSqsActionList{item}
		return nil
	}
	list := []IoTSqsAction{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = IoTSqsActionList(list)
		return nil
	}
	return err
}

// IoTTopicRulePayload represents AWS IoT TopicRulePayload
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-iot-topicrulepayload.html
type IoTTopicRulePayload struct {
	// The actions associated with the rule.
	Actions *IoTActionsList `json:"Actions,omitempty"`

	// The version of the SQL rules engine to use when evaluating the rule.
	AwsIotSqlVersion *StringExpr `json:"AwsIotSqlVersion,omitempty"`

	// The description of the rule.
	Description *StringExpr `json:"Description,omitempty"`

	// Specifies whether the rule is disabled.
	RuleDisabled *BoolExpr `json:"RuleDisabled,omitempty"`

	// The SQL statement that queries the topic. For more information, see
	// Rules for AWS IoT in the AWS IoT Developer Guide.
	Sql *StringExpr `json:"Sql,omitempty"`
}

// IoTTopicRulePayloadList represents a list of IoTTopicRulePayload
type IoTTopicRulePayloadList []IoTTopicRulePayload

// UnmarshalJSON sets the object from the provided JSON representation
func (l *IoTTopicRulePayloadList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := IoTTopicRulePayload{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = IoTTopicRulePayloadList{item}
		return nil
	}
	list := []IoTTopicRulePayload{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = IoTTopicRulePayloadList(list)
		return nil
	}
	return err
}

// KinesisFirehoseDeliveryStreamDestinationCloudWatchLoggingOptions represents Amazon Kinesis Firehose DeliveryStream Destination CloudWatchLoggingOptions
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-kinesisfirehose-kinesisdeliverystream-destination-cloudwatchloggingoptions.html
type KinesisFirehoseDeliveryStreamDestinationCloudWatchLoggingOptions struct {
	// Indicates whether CloudWatch Logs logging is enabled.
	Enabled *BoolExpr `json:"Enabled,omitempty"`

	// The name of the CloudWatch Logs log group that contains the log stream
	// that Firehose will use.
	LogGroupName *StringExpr `json:"LogGroupName,omitempty"`

	// The name of the CloudWatch Logs log stream that Firehose uses to send
	// logs about data delivery.
	LogStreamName *StringExpr `json:"LogStreamName,omitempty"`
}

// KinesisFirehoseDeliveryStreamDestinationCloudWatchLoggingOptionsList represents a list of KinesisFirehoseDeliveryStreamDestinationCloudWatchLoggingOptions
type KinesisFirehoseDeliveryStreamDestinationCloudWatchLoggingOptionsList []KinesisFirehoseDeliveryStreamDestinationCloudWatchLoggingOptions

// UnmarshalJSON sets the object from the provided JSON representation
func (l *KinesisFirehoseDeliveryStreamDestinationCloudWatchLoggingOptionsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := KinesisFirehoseDeliveryStreamDestinationCloudWatchLoggingOptions{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = KinesisFirehoseDeliveryStreamDestinationCloudWatchLoggingOptionsList{item}
		return nil
	}
	list := []KinesisFirehoseDeliveryStreamDestinationCloudWatchLoggingOptions{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = KinesisFirehoseDeliveryStreamDestinationCloudWatchLoggingOptionsList(list)
		return nil
	}
	return err
}

// KinesisFirehoseDeliveryStreamElasticsearchDestinationConfiguration represents Amazon Kinesis Firehose DeliveryStream ElasticsearchDestinationConfiguration
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-kinesisfirehose-kinesisdeliverystream-elasticsearchdestinationconfiguration.html
type KinesisFirehoseDeliveryStreamElasticsearchDestinationConfiguration struct {
	// Configures how Firehose buffers incoming data while delivering it to
	// the Amazon ES domain.
	BufferingHints *KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationBufferingHints `json:"BufferingHints,omitempty"`

	// The Amazon CloudWatch Logs logging options for the delivery stream.
	CloudWatchLoggingOptions *KinesisFirehoseDeliveryStreamDestinationCloudWatchLoggingOptions `json:"CloudWatchLoggingOptions,omitempty"`

	// The Amazon Resource Name (ARN) of the Amazon ES domain that Firehose
	// delivers data to.
	DomainARN *StringExpr `json:"DomainARN,omitempty"`

	// The name of the Elasticsearch index to which Firehose adds data for
	// indexing.
	IndexName *StringExpr `json:"IndexName,omitempty"`

	// The frequency of Elasticsearch index rotation. If you enable index
	// rotation, Firehose appends a portion of the UTC arrival timestamp to
	// the specified index name, and rotates the appended timestamp
	// accordingly. For more information, see Index Rotation for the Amazon
	// ES Destination in the Amazon Kinesis Firehose Developer Guide.
	IndexRotationPeriod *StringExpr `json:"IndexRotationPeriod,omitempty"`

	// The retry behavior when Firehose is unable to deliver data to Amazon
	// ES.
	RetryOptions *KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationRetryOptions `json:"RetryOptions,omitempty"`

	// The ARN of the AWS Identity and Access Management (IAM) role that
	// grants Firehose access to your S3 bucket, AWS KMS (if you enable data
	// encryption), and Amazon CloudWatch Logs (if you enable logging).
	RoleARN *StringExpr `json:"RoleARN,omitempty"`

	// The condition under which Firehose delivers data to Amazon Simple
	// Storage Service (Amazon S3). You can send Amazon S3 all documents (all
	// data) or only the documents that Firehose could not deliver to the
	// Amazon ES destination. For more information and valid values, see the
	// S3BackupMode content for the ElasticsearchDestinationConfiguration
	// data type in the Amazon Kinesis Firehose API Reference.
	S3BackupMode *StringExpr `json:"S3BackupMode,omitempty"`

	// The S3 bucket where Firehose backs up incoming data.
	S3Configuration *KinesisFirehoseDeliveryStreamS3DestinationConfiguration `json:"S3Configuration,omitempty"`

	// The Elasticsearch type name that Amazon ES adds to documents when
	// indexing data.
	TypeName *StringExpr `json:"TypeName,omitempty"`
}

// KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationList represents a list of KinesisFirehoseDeliveryStreamElasticsearchDestinationConfiguration
type KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationList []KinesisFirehoseDeliveryStreamElasticsearchDestinationConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := KinesisFirehoseDeliveryStreamElasticsearchDestinationConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationList{item}
		return nil
	}
	list := []KinesisFirehoseDeliveryStreamElasticsearchDestinationConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationList(list)
		return nil
	}
	return err
}

// KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationBufferingHints represents Amazon Kinesis Firehose DeliveryStream ElasticsearchDestinationConfiguration BufferingHints
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-kinesisfirehose-kinesisdeliverystream-elasticsearchdestinationconfiguration-bufferinghints.html
type KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationBufferingHints struct {
	// The length of time, in seconds, that Firehose buffers incoming data
	// before delivering it to the destination. For valid values, see the
	// IntervalInSeconds content for the BufferingHints data type in the
	// Amazon Kinesis Firehose API Reference.
	IntervalInSeconds *IntegerExpr `json:"IntervalInSeconds,omitempty"`

	// The size of the buffer, in MBs, that Firehose uses for incoming data
	// before delivering it to the destination. For valid values, see the
	// SizeInMBs content for the BufferingHints data type in the Amazon
	// Kinesis Firehose API Reference.
	SizeInMBs *IntegerExpr `json:"SizeInMBs,omitempty"`
}

// KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationBufferingHintsList represents a list of KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationBufferingHints
type KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationBufferingHintsList []KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationBufferingHints

// UnmarshalJSON sets the object from the provided JSON representation
func (l *KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationBufferingHintsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationBufferingHints{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationBufferingHintsList{item}
		return nil
	}
	list := []KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationBufferingHints{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationBufferingHintsList(list)
		return nil
	}
	return err
}

// KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationRetryOptions represents Amazon Kinesis Firehose DeliveryStream ElasticsearchDestinationConfiguration RetryOptions
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-kinesisfirehose-kinesisdeliverystream-elasticsearchdestinationconfiguration-retryoptions.html
type KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationRetryOptions struct {
	// After an initial failure to deliver to Amazon ES, the total amount of
	// time during which Firehose re-attempts delivery (including the first
	// attempt). If Firehose can't deliver the data within the specified
	// time, it writes the data to the backup S3 bucket. For valid values,
	// see the DurationInSeconds content for the ElasticsearchRetryOptions
	// data type in the Amazon Kinesis Firehose API Reference.
	DurationInSeconds *IntegerExpr `json:"DurationInSeconds,omitempty"`
}

// KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationRetryOptionsList represents a list of KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationRetryOptions
type KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationRetryOptionsList []KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationRetryOptions

// UnmarshalJSON sets the object from the provided JSON representation
func (l *KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationRetryOptionsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationRetryOptions{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationRetryOptionsList{item}
		return nil
	}
	list := []KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationRetryOptions{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = KinesisFirehoseDeliveryStreamElasticsearchDestinationConfigurationRetryOptionsList(list)
		return nil
	}
	return err
}

// KinesisFirehoseDeliveryStreamRedshiftDestinationConfiguration represents Amazon Kinesis Firehose DeliveryStream RedshiftDestinationConfiguration
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-kinesisfirehose-kinesisdeliverystream-redshiftdestinationconfiguration.html
type KinesisFirehoseDeliveryStreamRedshiftDestinationConfiguration struct {
	// The Amazon CloudWatch Logs logging options for the delivery stream.
	CloudWatchLoggingOptions *KinesisFirehoseDeliveryStreamDestinationCloudWatchLoggingOptions `json:"CloudWatchLoggingOptions,omitempty"`

	// The connection string that Firehose uses to connect to the Amazon
	// Redshift cluster.
	ClusterJDBCURL *StringExpr `json:"ClusterJDBCURL,omitempty"`

	// Configures the Amazon Redshift COPY command that Firehose uses to load
	// data into the cluster from the S3 bucket.
	CopyCommand *KinesisFirehoseDeliveryStreamRedshiftDestinationConfigurationCopyCommand `json:"CopyCommand,omitempty"`

	// The password for the Amazon Redshift user that you specified in the
	// Username property.
	Password *StringExpr `json:"Password,omitempty"`

	// The ARN of the AWS Identity and Access Management (IAM) role that
	// grants Firehose access to your S3 bucket and AWS KMS (if you enable
	// data encryption).
	RoleARN *StringExpr `json:"RoleARN,omitempty"`

	// The S3 bucket where Firehose first delivers data. After the data is in
	// the bucket, Firehose uses the COPY command to load the data into the
	// Amazon Redshift cluster. For the S3 bucket's compression format, don't
	// specify SNAPPY or ZIP because the Amazon Redshift COPY command doesn't
	// support them.
	S3Configuration *KinesisFirehoseDeliveryStreamS3DestinationConfiguration `json:"S3Configuration,omitempty"`

	// The Amazon Redshift user that has permission to access the Amazon
	// Redshift cluster. This user must have INSERT privileges for copying
	// data from the S3 bucket to the cluster.
	Username *StringExpr `json:"Username,omitempty"`
}

// KinesisFirehoseDeliveryStreamRedshiftDestinationConfigurationList represents a list of KinesisFirehoseDeliveryStreamRedshiftDestinationConfiguration
type KinesisFirehoseDeliveryStreamRedshiftDestinationConfigurationList []KinesisFirehoseDeliveryStreamRedshiftDestinationConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *KinesisFirehoseDeliveryStreamRedshiftDestinationConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := KinesisFirehoseDeliveryStreamRedshiftDestinationConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = KinesisFirehoseDeliveryStreamRedshiftDestinationConfigurationList{item}
		return nil
	}
	list := []KinesisFirehoseDeliveryStreamRedshiftDestinationConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = KinesisFirehoseDeliveryStreamRedshiftDestinationConfigurationList(list)
		return nil
	}
	return err
}

// KinesisFirehoseDeliveryStreamRedshiftDestinationConfigurationCopyCommand represents Amazon Kinesis Firehose DeliveryStream RedshiftDestinationConfiguration CopyCommand
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-kinesisfirehose-kinesisdeliverystream-redshiftdestinationconfiguration-copycommand.html
type KinesisFirehoseDeliveryStreamRedshiftDestinationConfigurationCopyCommand struct {
	// Parameters to use with the Amazon Redshift COPY command. For examples,
	// see the CopyOptions content for the CopyCommand data type in the
	// Amazon Kinesis Firehose API Reference.
	CopyOptions *StringExpr `json:"CopyOptions,omitempty"`

	// A comma-separated list of the column names in the table that Firehose
	// copies data to.
	DataTableColumns *StringExpr `json:"DataTableColumns,omitempty"`

	// The name of the table where Firehose adds the copied data.
	DataTableName *StringExpr `json:"DataTableName,omitempty"`
}

// KinesisFirehoseDeliveryStreamRedshiftDestinationConfigurationCopyCommandList represents a list of KinesisFirehoseDeliveryStreamRedshiftDestinationConfigurationCopyCommand
type KinesisFirehoseDeliveryStreamRedshiftDestinationConfigurationCopyCommandList []KinesisFirehoseDeliveryStreamRedshiftDestinationConfigurationCopyCommand

// UnmarshalJSON sets the object from the provided JSON representation
func (l *KinesisFirehoseDeliveryStreamRedshiftDestinationConfigurationCopyCommandList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := KinesisFirehoseDeliveryStreamRedshiftDestinationConfigurationCopyCommand{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = KinesisFirehoseDeliveryStreamRedshiftDestinationConfigurationCopyCommandList{item}
		return nil
	}
	list := []KinesisFirehoseDeliveryStreamRedshiftDestinationConfigurationCopyCommand{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = KinesisFirehoseDeliveryStreamRedshiftDestinationConfigurationCopyCommandList(list)
		return nil
	}
	return err
}

// KinesisFirehoseDeliveryStreamS3DestinationConfiguration represents Amazon Kinesis Firehose DeliveryStream S3DestinationConfiguration
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-kinesisfirehose-kinesisdeliverystream-s3destinationconfiguration.html
type KinesisFirehoseDeliveryStreamS3DestinationConfiguration struct {
	// The Amazon Resource Name (ARN) of the S3 bucket to send data to.
	BucketARN *StringExpr `json:"BucketARN,omitempty"`

	// Configures how Firehose buffers incoming data while delivering it to
	// the S3 bucket.
	BufferingHints *KinesisFirehoseDeliveryStreamS3DestinationConfigurationBufferingHints `json:"BufferingHints,omitempty"`

	// The Amazon CloudWatch Logs logging options for the delivery stream.
	CloudWatchLoggingOptions *KinesisFirehoseDeliveryStreamDestinationCloudWatchLoggingOptions `json:"CloudWatchLoggingOptions,omitempty"`

	// The type of compression that Firehose uses to compress the data that
	// it delivers to the S3 bucket. For valid values, see the
	// CompressionFormat content for the S3DestinationConfiguration data type
	// in the Amazon Kinesis Firehose API Reference.
	CompressionFormat *StringExpr `json:"CompressionFormat,omitempty"`

	// Configures Amazon Simple Storage Service (Amazon S3) server-side
	// encryption. Firehose uses AWS Key Management Service (AWS KMS) to
	// encrypt the data that it delivers to your S3 bucket.
	EncryptionConfiguration *KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfiguration `json:"EncryptionConfiguration,omitempty"`

	// A prefix that Firehose adds to the files that it delivers to the S3
	// bucket. The prefix helps you identify the files that Firehose
	// delivered.
	Prefix *StringExpr `json:"Prefix,omitempty"`

	// The ARN of an AWS Identity and Access Management (IAM) role that
	// grants Firehose access to your S3 bucket and AWS KMS (if you enable
	// data encryption).
	RoleARN *StringExpr `json:"RoleARN,omitempty"`
}

// KinesisFirehoseDeliveryStreamS3DestinationConfigurationList represents a list of KinesisFirehoseDeliveryStreamS3DestinationConfiguration
type KinesisFirehoseDeliveryStreamS3DestinationConfigurationList []KinesisFirehoseDeliveryStreamS3DestinationConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *KinesisFirehoseDeliveryStreamS3DestinationConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := KinesisFirehoseDeliveryStreamS3DestinationConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = KinesisFirehoseDeliveryStreamS3DestinationConfigurationList{item}
		return nil
	}
	list := []KinesisFirehoseDeliveryStreamS3DestinationConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = KinesisFirehoseDeliveryStreamS3DestinationConfigurationList(list)
		return nil
	}
	return err
}

// KinesisFirehoseDeliveryStreamS3DestinationConfigurationBufferingHints represents Amazon Kinesis Firehose DeliveryStream S3DestinationConfiguration BufferingHints
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-kinesisfirehose-kinesisdeliverystream-s3destinationconfiguration-bufferinghints.html
type KinesisFirehoseDeliveryStreamS3DestinationConfigurationBufferingHints struct {
	// The length of time, in seconds, that Firehose buffers incoming data
	// before delivering it to the destination. For valid values, see the
	// IntervalInSeconds content for the BufferingHints data type in the
	// Amazon Kinesis Firehose API Reference.
	IntervalInSeconds *IntegerExpr `json:"IntervalInSeconds,omitempty"`

	// The size of the buffer, in MBs, that Firehose uses for incoming data
	// before delivering it to the destination. For valid values, see the
	// SizeInMBs content for the BufferingHints data type in the Amazon
	// Kinesis Firehose API Reference.
	SizeInMBs *IntegerExpr `json:"SizeInMBs,omitempty"`
}

// KinesisFirehoseDeliveryStreamS3DestinationConfigurationBufferingHintsList represents a list of KinesisFirehoseDeliveryStreamS3DestinationConfigurationBufferingHints
type KinesisFirehoseDeliveryStreamS3DestinationConfigurationBufferingHintsList []KinesisFirehoseDeliveryStreamS3DestinationConfigurationBufferingHints

// UnmarshalJSON sets the object from the provided JSON representation
func (l *KinesisFirehoseDeliveryStreamS3DestinationConfigurationBufferingHintsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := KinesisFirehoseDeliveryStreamS3DestinationConfigurationBufferingHints{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = KinesisFirehoseDeliveryStreamS3DestinationConfigurationBufferingHintsList{item}
		return nil
	}
	list := []KinesisFirehoseDeliveryStreamS3DestinationConfigurationBufferingHints{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = KinesisFirehoseDeliveryStreamS3DestinationConfigurationBufferingHintsList(list)
		return nil
	}
	return err
}

// KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfigurationKMSEncryptionConfig represents Amazon Kinesis Firehose DeliveryStream S3DestinationConfiguration EncryptionConfiguration KMSEncryptionConfig
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-kinesisfirehose-kinesisdeliverystream-s3destinationconfiguration-encryptionconfiguration-kmsencryptionconfig.html
type KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfigurationKMSEncryptionConfig struct {
	// The Amazon Resource Name (ARN) of the AWS KMS encryption key that
	// Amazon S3 uses to encrypt data delivered by the Firehose stream. The
	// key must belong to the same region as the destination S3 bucket.
	AWSKMSKeyARN *StringExpr `json:"AWSKMSKeyARN,omitempty"`
}

// KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfigurationKMSEncryptionConfigList represents a list of KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfigurationKMSEncryptionConfig
type KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfigurationKMSEncryptionConfigList []KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfigurationKMSEncryptionConfig

// UnmarshalJSON sets the object from the provided JSON representation
func (l *KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfigurationKMSEncryptionConfigList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfigurationKMSEncryptionConfig{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfigurationKMSEncryptionConfigList{item}
		return nil
	}
	list := []KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfigurationKMSEncryptionConfig{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfigurationKMSEncryptionConfigList(list)
		return nil
	}
	return err
}

// KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfiguration represents Amazon Kinesis Firehose DeliveryStream S3DestinationConfiguration EncryptionConfiguration
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-kinesisfirehose-kinesisdeliverystream-s3destinationconfiguration-encryptionconfiguration.html
type KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfiguration struct {
	// The AWS Key Management Service (AWS KMS) encryption key that Amazon S3
	// uses to encrypt your data.
	KMSEncryptionConfig *KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfigurationKMSEncryptionConfig `json:"KMSEncryptionConfig,omitempty"`

	// Disables encryption. For valid values, see the NoEncryptionConfig
	// content for the EncryptionConfiguration data type in the Amazon
	// Kinesis Firehose API Reference.
	NoEncryptionConfig *StringExpr `json:"NoEncryptionConfig,omitempty"`
}

// KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfigurationList represents a list of KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfiguration
type KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfigurationList []KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfigurationList{item}
		return nil
	}
	list := []KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = KinesisFirehoseDeliveryStreamS3DestinationConfigurationEncryptionConfigurationList(list)
		return nil
	}
	return err
}

// LambdaFunctionEnvironment represents AWS Lambda Function Environment
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-lambda-function-environment.html
type LambdaFunctionEnvironment struct {
	// A map of key-value pairs that the Lambda function can access.
	Variables interface{} `json:"Variables,omitempty"`
}

// LambdaFunctionEnvironmentList represents a list of LambdaFunctionEnvironment
type LambdaFunctionEnvironmentList []LambdaFunctionEnvironment

// UnmarshalJSON sets the object from the provided JSON representation
func (l *LambdaFunctionEnvironmentList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := LambdaFunctionEnvironment{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = LambdaFunctionEnvironmentList{item}
		return nil
	}
	list := []LambdaFunctionEnvironment{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = LambdaFunctionEnvironmentList(list)
		return nil
	}
	return err
}

// LambdaFunctionCode represents AWS Lambda Function Code
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-lambda-function-code.html
type LambdaFunctionCode struct {
	// The name of the S3 bucket that contains the source code of your Lambda
	// function. The S3 bucket must be in the same region as the stack.
	S3Bucket *StringExpr `json:"S3Bucket,omitempty"`

	// The location and name of the .zip file that contains your source code.
	// If you specify this property, you must also specify the S3Bucket
	// property.
	S3Key *StringExpr `json:"S3Key,omitempty"`

	// If you have S3 versioning enabled, the version ID of the.zip file that
	// contains your source code. You can specify this property only if you
	// specify the S3Bucket and S3Key properties.
	S3ObjectVersion *StringExpr `json:"S3ObjectVersion,omitempty"`

	// For nodejs4.3 and python2.7 runtime environments, the source code of
	// your Lambda function. You can't use this property with other runtime
	// environments.
	ZipFile *StringExpr `json:"ZipFile,omitempty"`
}

// LambdaFunctionCodeList represents a list of LambdaFunctionCode
type LambdaFunctionCodeList []LambdaFunctionCode

// UnmarshalJSON sets the object from the provided JSON representation
func (l *LambdaFunctionCodeList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := LambdaFunctionCode{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = LambdaFunctionCodeList{item}
		return nil
	}
	list := []LambdaFunctionCode{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = LambdaFunctionCodeList(list)
		return nil
	}
	return err
}

// LambdaFunctionVPCConfig represents AWS Lambda Function VPCConfig
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-lambda-function-vpcconfig.html
type LambdaFunctionVPCConfig struct {
	// A list of one or more security groups IDs in the VPC that includes the
	// resources to which your Lambda function requires access.
	SecurityGroupIds *StringListExpr `json:"SecurityGroupIds,omitempty"`

	// A list of one or more subnet IDs in the VPC that includes the
	// resources to which your Lambda function requires access.
	SubnetIds *StringListExpr `json:"SubnetIds,omitempty"`
}

// LambdaFunctionVPCConfigList represents a list of LambdaFunctionVPCConfig
type LambdaFunctionVPCConfigList []LambdaFunctionVPCConfig

// UnmarshalJSON sets the object from the provided JSON representation
func (l *LambdaFunctionVPCConfigList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := LambdaFunctionVPCConfig{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = LambdaFunctionVPCConfigList{item}
		return nil
	}
	list := []LambdaFunctionVPCConfig{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = LambdaFunctionVPCConfigList(list)
		return nil
	}
	return err
}

// Name represents Name Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-name.html
type Name struct {
}

// NameList represents a list of Name
type NameList []Name

// UnmarshalJSON sets the object from the provided JSON representation
func (l *NameList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := Name{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = NameList{item}
		return nil
	}
	list := []Name{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = NameList(list)
		return nil
	}
	return err
}

// DataSource represents DataSource
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-opsworks-app-datasource.html
type DataSource struct {
	// The ARN of the data source.
	Arn *StringExpr `json:"Arn,omitempty"`

	// The name of the database.
	DatabaseName *StringExpr `json:"DatabaseName,omitempty"`

	// The type of the data source, such as AutoSelectOpsworksMysqlInstance,
	// OpsworksMysqlInstance, or RdsDbInstance. For valid values, see the
	// DataSource type in the AWS OpsWorks API Reference.
	Type *StringExpr `json:"Type,omitempty"`
}

// DataSourceList represents a list of DataSource
type DataSourceList []DataSource

// UnmarshalJSON sets the object from the provided JSON representation
func (l *DataSourceList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := DataSource{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = DataSourceList{item}
		return nil
	}
	list := []DataSource{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = DataSourceList(list)
		return nil
	}
	return err
}

// OpsWorksAppEnvironment represents AWS OpsWorks App Environment
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-opsworks-app-environment.html
type OpsWorksAppEnvironment struct {
	// The name of the environment variable, which can consist of up to 64
	// characters. You can use upper and lowercase letters, numbers, and
	// underscores (_), but the name must start with a letter or underscore.
	Key *StringExpr `json:"Key,omitempty"`

	// Indicates whether the value of the environment variable is concealed,
	// such as with a DescribeApps response. To conceal an environment
	// variable's value, set the value to true.
	Secure *BoolExpr `json:"Secure,omitempty"`

	// The value of the environment variable, which can be empty. You can
	// specify a value of up to 256 characters.
	Value *StringExpr `json:"Value,omitempty"`
}

// OpsWorksAppEnvironmentList represents a list of OpsWorksAppEnvironment
type OpsWorksAppEnvironmentList []OpsWorksAppEnvironment

// UnmarshalJSON sets the object from the provided JSON representation
func (l *OpsWorksAppEnvironmentList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := OpsWorksAppEnvironment{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = OpsWorksAppEnvironmentList{item}
		return nil
	}
	list := []OpsWorksAppEnvironment{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = OpsWorksAppEnvironmentList(list)
		return nil
	}
	return err
}

// OpsWorksAutoScalingThresholds represents AWS OpsWorks AutoScalingThresholds Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-opsworks-layer-loadbasedautoscaling-autoscalingthresholds.html
type OpsWorksAutoScalingThresholds struct {
	// The percentage of CPU utilization that triggers the starting or
	// stopping of instances (scaling).
	CpuThreshold *IntegerExpr `json:"CpuThreshold,omitempty"`

	// The amount of time (in minutes) after a scaling event occurs that AWS
	// OpsWorks should ignore metrics and not start any additional scaling
	// events.
	IgnoreMetricsTime *IntegerExpr `json:"IgnoreMetricsTime,omitempty"`

	// The number of instances to add or remove when the load exceeds a
	// threshold.
	InstanceCount *IntegerExpr `json:"InstanceCount,omitempty"`

	// The degree of system load that triggers the starting or stopping of
	// instances (scaling). For more information about how load is computed,
	// see Load (computing).
	LoadThreshold *IntegerExpr `json:"LoadThreshold,omitempty"`

	// The percentage of memory consumption that triggers the starting or
	// stopping of instances (scaling).
	MemoryThreshold *IntegerExpr `json:"MemoryThreshold,omitempty"`

	// The amount of time, in minutes, that the load must exceed a threshold
	// before instances are added or removed.
	ThresholdsWaitTime *IntegerExpr `json:"ThresholdsWaitTime,omitempty"`
}

// OpsWorksAutoScalingThresholdsList represents a list of OpsWorksAutoScalingThresholds
type OpsWorksAutoScalingThresholdsList []OpsWorksAutoScalingThresholds

// UnmarshalJSON sets the object from the provided JSON representation
func (l *OpsWorksAutoScalingThresholdsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := OpsWorksAutoScalingThresholds{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = OpsWorksAutoScalingThresholdsList{item}
		return nil
	}
	list := []OpsWorksAutoScalingThresholds{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = OpsWorksAutoScalingThresholdsList(list)
		return nil
	}
	return err
}

// OpsWorksChefConfiguration represents AWS OpsWorks ChefConfiguration Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-opsworks-stack-chefconfiguration.html
type OpsWorksChefConfiguration struct {
	// The Berkshelf version.
	BerkshelfVersion *StringExpr `json:"BerkshelfVersion,omitempty"`

	// Whether to enable Berkshelf.
	ManageBerkshelf *BoolExpr `json:"ManageBerkshelf,omitempty"`
}

// OpsWorksChefConfigurationList represents a list of OpsWorksChefConfiguration
type OpsWorksChefConfigurationList []OpsWorksChefConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *OpsWorksChefConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := OpsWorksChefConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = OpsWorksChefConfigurationList{item}
		return nil
	}
	list := []OpsWorksChefConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = OpsWorksChefConfigurationList(list)
		return nil
	}
	return err
}

// OpsWorksLayerLifeCycleConfiguration represents AWS OpsWorks Layer LifeCycleConfiguration
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-opsworks-layer-lifecycleeventconfiguration.html
type OpsWorksLayerLifeCycleConfiguration struct {
	// Specifies the shutdown event configuration for a layer.
	ShutdownEventConfiguration *OpsWorksLayerLifeCycleConfigurationShutdownEventConfiguration `json:"ShutdownEventConfiguration,omitempty"`
}

// OpsWorksLayerLifeCycleConfigurationList represents a list of OpsWorksLayerLifeCycleConfiguration
type OpsWorksLayerLifeCycleConfigurationList []OpsWorksLayerLifeCycleConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *OpsWorksLayerLifeCycleConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := OpsWorksLayerLifeCycleConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = OpsWorksLayerLifeCycleConfigurationList{item}
		return nil
	}
	list := []OpsWorksLayerLifeCycleConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = OpsWorksLayerLifeCycleConfigurationList(list)
		return nil
	}
	return err
}

// OpsWorksLayerLifeCycleConfigurationShutdownEventConfiguration represents AWS OpsWorks Layer LifeCycleConfiguration ShutdownEventConfiguration
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-opsworks-layer-lifecycleeventconfiguration-shutdowneventconfiguration.html
type OpsWorksLayerLifeCycleConfigurationShutdownEventConfiguration struct {
	// Indicates whether to wait for connections to drain from the Elastic
	// Load Balancing load balancers.
	DelayUntilElbConnectionsDrained *BoolExpr `json:"DelayUntilElbConnectionsDrained,omitempty"`

	// The time, in seconds, that AWS OpsWorks waits after a shutdown event
	// has been triggered before shutting down an instance.
	ExecutionTimeout *IntegerExpr `json:"ExecutionTimeout,omitempty"`
}

// OpsWorksLayerLifeCycleConfigurationShutdownEventConfigurationList represents a list of OpsWorksLayerLifeCycleConfigurationShutdownEventConfiguration
type OpsWorksLayerLifeCycleConfigurationShutdownEventConfigurationList []OpsWorksLayerLifeCycleConfigurationShutdownEventConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *OpsWorksLayerLifeCycleConfigurationShutdownEventConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := OpsWorksLayerLifeCycleConfigurationShutdownEventConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = OpsWorksLayerLifeCycleConfigurationShutdownEventConfigurationList{item}
		return nil
	}
	list := []OpsWorksLayerLifeCycleConfigurationShutdownEventConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = OpsWorksLayerLifeCycleConfigurationShutdownEventConfigurationList(list)
		return nil
	}
	return err
}

// OpsWorksLoadBasedAutoScaling represents AWS OpsWorks LoadBasedAutoScaling Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-opsworks-layer-loadbasedautoscaling.html
type OpsWorksLoadBasedAutoScaling struct {
	// The threshold below which the instances are scaled down (stopped). If
	// the load falls below this threshold for a specified amount of time,
	// AWS OpsWorks stops a specified number of instances.
	DownScaling *OpsWorksAutoScalingThresholds `json:"DownScaling,omitempty"`

	// Whether to enable automatic load-based scaling for the layer.
	Enable *BoolExpr `json:"Enable,omitempty"`

	// The threshold above which the instances are scaled up (added). If the
	// load exceeds this thresholds for a specified amount of time, AWS
	// OpsWorks starts a specified number of instances.
	UpScaling *OpsWorksAutoScalingThresholds `json:"UpScaling,omitempty"`
}

// OpsWorksLoadBasedAutoScalingList represents a list of OpsWorksLoadBasedAutoScaling
type OpsWorksLoadBasedAutoScalingList []OpsWorksLoadBasedAutoScaling

// UnmarshalJSON sets the object from the provided JSON representation
func (l *OpsWorksLoadBasedAutoScalingList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := OpsWorksLoadBasedAutoScaling{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = OpsWorksLoadBasedAutoScalingList{item}
		return nil
	}
	list := []OpsWorksLoadBasedAutoScaling{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = OpsWorksLoadBasedAutoScalingList(list)
		return nil
	}
	return err
}

// OpsWorksInstanceBlockDeviceMapping represents AWS OpsWorks Instance BlockDeviceMapping
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-opsworks-instance-blockdevicemapping.html
type OpsWorksInstanceBlockDeviceMapping struct {
	// The name of the device that is exposed to the instance, such as
	// /dev/dsh or xvdh. For the root device, you can use the explicit device
	// name or you can set this parameter to ROOT_DEVICE. If you set the
	// parameter to ROOT_DEVICE, AWS OpsWorks provides the correct device
	// name.
	DeviceName *StringExpr `json:"DeviceName,omitempty"`

	// Configuration information about the Amazon Elastic Block Store (Amazon
	// EBS) volume.
	Ebs *OpsWorksInstanceBlockDeviceMappingEbsBlockDevice `json:"Ebs,omitempty"`

	// Suppresses the device that is specified in the block device mapping of
	// the AWS OpsWorks instance Amazon Machine Image (AMI).
	NoDevice *StringExpr `json:"NoDevice,omitempty"`

	// The name of the virtual device. The name must be in the form
	// ephemeralX, where X is a number equal to or greater than zero (0), for
	// example, ephemeral0.
	VirtualName *StringExpr `json:"VirtualName,omitempty"`
}

// OpsWorksInstanceBlockDeviceMappingList represents a list of OpsWorksInstanceBlockDeviceMapping
type OpsWorksInstanceBlockDeviceMappingList []OpsWorksInstanceBlockDeviceMapping

// UnmarshalJSON sets the object from the provided JSON representation
func (l *OpsWorksInstanceBlockDeviceMappingList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := OpsWorksInstanceBlockDeviceMapping{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = OpsWorksInstanceBlockDeviceMappingList{item}
		return nil
	}
	list := []OpsWorksInstanceBlockDeviceMapping{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = OpsWorksInstanceBlockDeviceMappingList(list)
		return nil
	}
	return err
}

// OpsWorksInstanceBlockDeviceMappingEbsBlockDevice represents AWS OpsWorks Instance BlockDeviceMapping EbsBlockDevice
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-opsworks-instance-blockdevicemapping-ebsblockdevice.html
type OpsWorksInstanceBlockDeviceMappingEbsBlockDevice struct {
	// Indicates whether to delete the volume when the instance is
	// terminated.
	DeleteOnTermination *BoolExpr `json:"DeleteOnTermination,omitempty"`

	// The number of I/O operations per second (IOPS) that the volume
	// supports. For more information, see Iops for the EbsBlockDevice action
	// in the Amazon EC2 API Reference.
	Iops *IntegerExpr `json:"Iops,omitempty"`

	// The snapshot ID of the volume that you want to use. If you specify
	// both the SnapshotId and VolumeSize properties, VolumeSize must be
	// equal to or greater than the size of the snapshot.
	SnapshotId *StringExpr `json:"SnapshotId,omitempty"`

	// The volume size, in Gibibytes (GiB). If you specify both the
	// SnapshotId and VolumeSize properties, VolumeSize must be equal to or
	// greater than the size of the snapshot. For more information about
	// specifying volume size, see VolumeSize for the EbsBlockDevice action
	// in the Amazon EC2 API Reference.
	VolumeSize *IntegerExpr `json:"VolumeSize,omitempty"`

	// The volume type. For more information about specifying the volume
	// type, see VolumeType for the EbsBlockDevice action in the Amazon EC2
	// API Reference.
	VolumeType *StringExpr `json:"VolumeType,omitempty"`
}

// OpsWorksInstanceBlockDeviceMappingEbsBlockDeviceList represents a list of OpsWorksInstanceBlockDeviceMappingEbsBlockDevice
type OpsWorksInstanceBlockDeviceMappingEbsBlockDeviceList []OpsWorksInstanceBlockDeviceMappingEbsBlockDevice

// UnmarshalJSON sets the object from the provided JSON representation
func (l *OpsWorksInstanceBlockDeviceMappingEbsBlockDeviceList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := OpsWorksInstanceBlockDeviceMappingEbsBlockDevice{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = OpsWorksInstanceBlockDeviceMappingEbsBlockDeviceList{item}
		return nil
	}
	list := []OpsWorksInstanceBlockDeviceMappingEbsBlockDevice{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = OpsWorksInstanceBlockDeviceMappingEbsBlockDeviceList(list)
		return nil
	}
	return err
}

// OpsWorksRecipes represents AWS OpsWorks Recipes Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-opsworks-layer-recipes.html
type OpsWorksRecipes struct {
	// Custom recipe names to be run following a Configure event. The event
	// occurs on all of the stack's instances when an instance enters or
	// leaves the online state.
	Configure *StringListExpr `json:"Configure,omitempty"`

	// Custom recipe names to be run following a Deploy event. The event
	// occurs when you run a deploy command, typically to deploy an
	// application to a set of application server instances.
	Deploy *StringListExpr `json:"Deploy,omitempty"`

	// Custom recipe names to be run following a Setup event. This event
	// occurs on a new instance after it successfully boots.
	Setup *StringListExpr `json:"Setup,omitempty"`

	// Custom recipe names to be run following a Shutdown event. This event
	// occurs after you direct AWS OpsWorks to shut an instance down before
	// the associated Amazon EC2 instance is actually terminated.
	Shutdown *StringListExpr `json:"Shutdown,omitempty"`

	// Custom recipe names to be run following a Undeploy event. This event
	// occurs when you delete an app or run an undeploy command to remove an
	// app from a set of application server instances.
	Undeploy *StringListExpr `json:"Undeploy,omitempty"`
}

// OpsWorksRecipesList represents a list of OpsWorksRecipes
type OpsWorksRecipesList []OpsWorksRecipes

// UnmarshalJSON sets the object from the provided JSON representation
func (l *OpsWorksRecipesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := OpsWorksRecipes{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = OpsWorksRecipesList{item}
		return nil
	}
	list := []OpsWorksRecipes{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = OpsWorksRecipesList(list)
		return nil
	}
	return err
}

// OpsWorksSource represents AWS OpsWorks Source Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-opsworks-stack-source.html
type OpsWorksSource struct {
	// This parameter depends on the repository type. For Amazon S3 bundles,
	// set Password to the appropriate IAM secret access key. For HTTP
	// bundles, Git repositories, and Subversion repositories, set Password
	// to the appropriate password.
	Password *StringExpr `json:"Password,omitempty"`

	// The application's version. With AWS OpsWorks, you can deploy new
	// versions of an application. One of the simplest approaches is to have
	// branches or revisions in your repository that represent different
	// versions that can potentially be deployed.
	Revision *StringExpr `json:"Revision,omitempty"`

	// The repository's SSH key. For more information, see Using Git
	// Repository SSH Keys in the AWS OpsWorks User Guide.
	SshKey *StringExpr `json:"SshKey,omitempty"`

	// The repository type.
	Type *StringExpr `json:"Type,omitempty"`

	// The source URL.
	Url *StringExpr `json:"Url,omitempty"`

	// This parameter depends on the repository type. For Amazon S3 bundles,
	// set Username to the appropriate IAM access key ID. For HTTP bundles,
	// Git repositories, and Subversion repositories, set Username to the
	// appropriate user name.
	Username *StringExpr `json:"Username,omitempty"`
}

// OpsWorksSourceList represents a list of OpsWorksSource
type OpsWorksSourceList []OpsWorksSource

// UnmarshalJSON sets the object from the provided JSON representation
func (l *OpsWorksSourceList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := OpsWorksSource{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = OpsWorksSourceList{item}
		return nil
	}
	list := []OpsWorksSource{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = OpsWorksSourceList(list)
		return nil
	}
	return err
}

// OpsWorksSslConfiguration represents AWS OpsWorks SslConfiguration Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-opsworks-app-sslconfiguration.html
type OpsWorksSslConfiguration struct {
	// The contents of the certificate's domain.crt file.
	Certificate *StringExpr `json:"Certificate,omitempty"`

	// An intermediate certificate authority key or client authentication.
	Chain *StringExpr `json:"Chain,omitempty"`

	// The private key; the contents of the certificate's domain.kex file.
	PrivateKey *StringExpr `json:"PrivateKey,omitempty"`
}

// OpsWorksSslConfigurationList represents a list of OpsWorksSslConfiguration
type OpsWorksSslConfigurationList []OpsWorksSslConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *OpsWorksSslConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := OpsWorksSslConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = OpsWorksSslConfigurationList{item}
		return nil
	}
	list := []OpsWorksSslConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = OpsWorksSslConfigurationList(list)
		return nil
	}
	return err
}

// OpsWorksStackElasticIp represents AWS OpsWorks Stack ElasticIp
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-opsworks-stack-elasticip.html
type OpsWorksStackElasticIp struct {
	// The Elastic IP address.
	Ip *StringExpr `json:"Ip,omitempty"`

	// A name for the Elastic IP address.
	Name *StringExpr `json:"Name,omitempty"`
}

// OpsWorksStackElasticIpList represents a list of OpsWorksStackElasticIp
type OpsWorksStackElasticIpList []OpsWorksStackElasticIp

// UnmarshalJSON sets the object from the provided JSON representation
func (l *OpsWorksStackElasticIpList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := OpsWorksStackElasticIp{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = OpsWorksStackElasticIpList{item}
		return nil
	}
	list := []OpsWorksStackElasticIp{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = OpsWorksStackElasticIpList(list)
		return nil
	}
	return err
}

// OpsWorksStackRdsDbInstance represents AWS OpsWorks Stack RdsDbInstance
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-opsworks-stack-rdsdbinstance.html
type OpsWorksStackRdsDbInstance struct {
	// The password of the registered database.
	DbPassword *StringExpr `json:"DbPassword,omitempty"`

	// The master user name of the registered database.
	DbUser *StringExpr `json:"DbUser,omitempty"`

	// The Amazon Resource Name (ARN) of the Amazon RDS DB instance to
	// register with the AWS OpsWorks stack.
	RdsDbInstanceArn *StringExpr `json:"RdsDbInstanceArn,omitempty"`
}

// OpsWorksStackRdsDbInstanceList represents a list of OpsWorksStackRdsDbInstance
type OpsWorksStackRdsDbInstanceList []OpsWorksStackRdsDbInstance

// UnmarshalJSON sets the object from the provided JSON representation
func (l *OpsWorksStackRdsDbInstanceList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := OpsWorksStackRdsDbInstance{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = OpsWorksStackRdsDbInstanceList{item}
		return nil
	}
	list := []OpsWorksStackRdsDbInstance{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = OpsWorksStackRdsDbInstanceList(list)
		return nil
	}
	return err
}

// OpsWorksStackConfigurationManager represents AWS OpsWorks StackConfigurationManager Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-opsworks-stack-stackconfigmanager.html
type OpsWorksStackConfigurationManager struct {
	// The name of the configuration manager.
	Name *StringExpr `json:"Name,omitempty"`

	// The Chef version.
	Version *StringExpr `json:"Version,omitempty"`
}

// OpsWorksStackConfigurationManagerList represents a list of OpsWorksStackConfigurationManager
type OpsWorksStackConfigurationManagerList []OpsWorksStackConfigurationManager

// UnmarshalJSON sets the object from the provided JSON representation
func (l *OpsWorksStackConfigurationManagerList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := OpsWorksStackConfigurationManager{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = OpsWorksStackConfigurationManagerList{item}
		return nil
	}
	list := []OpsWorksStackConfigurationManager{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = OpsWorksStackConfigurationManagerList(list)
		return nil
	}
	return err
}

// OpsWorksTimeBasedAutoScaling represents AWS OpsWorks TimeBasedAutoScaling Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-opsworks-instance-timebasedautoscaling.html
type OpsWorksTimeBasedAutoScaling struct {
	// The schedule for Friday.
	Friday *StringExpr `json:"Friday,omitempty"`

	// The schedule for Monday.
	Monday *StringExpr `json:"Monday,omitempty"`

	// The schedule for Saturday.
	Saturday *StringExpr `json:"Saturday,omitempty"`

	// The schedule for Sunday.
	Sunday *StringExpr `json:"Sunday,omitempty"`

	// The schedule for Thursday.
	Thursday *StringExpr `json:"Thursday,omitempty"`

	// The schedule for Tuesday.
	Tuesday *StringExpr `json:"Tuesday,omitempty"`

	// The schedule for Wednesday.
	Wednesday *StringExpr `json:"Wednesday,omitempty"`
}

// OpsWorksTimeBasedAutoScalingList represents a list of OpsWorksTimeBasedAutoScaling
type OpsWorksTimeBasedAutoScalingList []OpsWorksTimeBasedAutoScaling

// UnmarshalJSON sets the object from the provided JSON representation
func (l *OpsWorksTimeBasedAutoScalingList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := OpsWorksTimeBasedAutoScaling{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = OpsWorksTimeBasedAutoScalingList{item}
		return nil
	}
	list := []OpsWorksTimeBasedAutoScaling{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = OpsWorksTimeBasedAutoScalingList(list)
		return nil
	}
	return err
}

// OpsWorksVolumeConfiguration represents AWS OpsWorks VolumeConfiguration Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-opsworks-layer-volumeconfig.html
type OpsWorksVolumeConfiguration struct {
	// The number of I/O operations per second (IOPS) to provision for the
	// volume.
	Iops *IntegerExpr `json:"Iops,omitempty"`

	// The volume mount point, such as /dev/sdh.
	MountPoint *StringExpr `json:"MountPoint,omitempty"`

	// The number of disks in the volume.
	NumberOfDisks *IntegerExpr `json:"NumberOfDisks,omitempty"`

	// The volume RAID level.
	RaidLevel *IntegerExpr `json:"RaidLevel,omitempty"`

	// The volume size.
	Size *IntegerExpr `json:"Size,omitempty"`

	// The type of volume, such as magnetic or SSD. For valid values, see
	// VolumeConfiguration in the AWS OpsWorks API Reference.
	VolumeType *StringExpr `json:"VolumeType,omitempty"`
}

// OpsWorksVolumeConfigurationList represents a list of OpsWorksVolumeConfiguration
type OpsWorksVolumeConfigurationList []OpsWorksVolumeConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *OpsWorksVolumeConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := OpsWorksVolumeConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = OpsWorksVolumeConfigurationList{item}
		return nil
	}
	list := []OpsWorksVolumeConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = OpsWorksVolumeConfigurationList(list)
		return nil
	}
	return err
}

// RedshiftParameter represents Amazon Redshift Parameter Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-property-redshift-clusterparametergroup-parameter.html
type RedshiftParameter struct {
	// The name of the parameter.
	ParameterName *StringExpr `json:"ParameterName,omitempty"`

	// The value of the parameter.
	ParameterValue *StringExpr `json:"ParameterValue,omitempty"`
}

// RedshiftParameterList represents a list of RedshiftParameter
type RedshiftParameterList []RedshiftParameter

// UnmarshalJSON sets the object from the provided JSON representation
func (l *RedshiftParameterList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := RedshiftParameter{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = RedshiftParameterList{item}
		return nil
	}
	list := []RedshiftParameter{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = RedshiftParameterList(list)
		return nil
	}
	return err
}

// ResourceTag represents AWS CloudFormation Resource Tags Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-resource-tags.html
type ResourceTag struct {
	// The key name of the tag. You can specify a value that is 1 to 127
	// Unicode characters in length and cannot be prefixed with aws:. You can
	// use any of the following characters: the set of Unicode letters,
	// digits, whitespace, _, ., /, =, +, and -.
	Key *StringExpr `json:"Key,omitempty"`

	// The value for the tag. You can specify a value that is 1 to 255
	// Unicode characters in length and cannot be prefixed with aws:. You can
	// use any of the following characters: the set of Unicode letters,
	// digits, whitespace, _, ., /, =, +, and -.
	Value *StringExpr `json:"Value,omitempty"`
}

// ResourceTagList represents a list of ResourceTag
type ResourceTagList []ResourceTag

// UnmarshalJSON sets the object from the provided JSON representation
func (l *ResourceTagList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := ResourceTag{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = ResourceTagList{item}
		return nil
	}
	list := []ResourceTag{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = ResourceTagList(list)
		return nil
	}
	return err
}

// RDSOptionGroupOptionConfigurations represents Amazon RDS OptionGroup OptionConfigurations
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-rds-optiongroup-optionconfigurations.html
type RDSOptionGroupOptionConfigurations struct {
	// A list of database security group names for this option. If the option
	// requires access to a port, the security groups must allow access to
	// that port. If you specify this property, don't specify the
	// VPCSecurityGroupMemberships property.
	DBSecurityGroupMemberships *StringListExpr `json:"DBSecurityGroupMemberships,omitempty"`

	// The name of the option. For more information about options, see
	// Working with Option Groups in the Amazon Relational Database Service
	// User Guide.
	OptionName *StringExpr `json:"OptionName,omitempty"`

	// The settings for this option.
	OptionSettings *RDSOptionGroupOptionConfigurationsOptionSettingsList `json:"OptionSettings,omitempty"`

	// The port number that this option uses.
	Port *IntegerExpr `json:"Port,omitempty"`

	// A list of VPC security group IDs for this option. If the option
	// requires access to a port, the security groups must allow access to
	// that port. If you specify this property, don't specify the
	// DBSecurityGroupMemberships property.
	VpcSecurityGroupMemberships *StringListExpr `json:"VpcSecurityGroupMemberships,omitempty"`
}

// RDSOptionGroupOptionConfigurationsList represents a list of RDSOptionGroupOptionConfigurations
type RDSOptionGroupOptionConfigurationsList []RDSOptionGroupOptionConfigurations

// UnmarshalJSON sets the object from the provided JSON representation
func (l *RDSOptionGroupOptionConfigurationsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := RDSOptionGroupOptionConfigurations{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = RDSOptionGroupOptionConfigurationsList{item}
		return nil
	}
	list := []RDSOptionGroupOptionConfigurations{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = RDSOptionGroupOptionConfigurationsList(list)
		return nil
	}
	return err
}

// RDSOptionGroupOptionConfigurationsOptionSettings represents Amazon RDS OptionGroup OptionConfigurations OptionSettings
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-rds-optiongroup-optionconfigurations-optionsettings.html
type RDSOptionGroupOptionConfigurationsOptionSettings struct {
	// The name of the option setting that you want to specify.
	Name *StringExpr `json:"Name,omitempty"`

	// The value of the option setting.
	Value *StringExpr `json:"Value,omitempty"`
}

// RDSOptionGroupOptionConfigurationsOptionSettingsList represents a list of RDSOptionGroupOptionConfigurationsOptionSettings
type RDSOptionGroupOptionConfigurationsOptionSettingsList []RDSOptionGroupOptionConfigurationsOptionSettings

// UnmarshalJSON sets the object from the provided JSON representation
func (l *RDSOptionGroupOptionConfigurationsOptionSettingsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := RDSOptionGroupOptionConfigurationsOptionSettings{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = RDSOptionGroupOptionConfigurationsOptionSettingsList{item}
		return nil
	}
	list := []RDSOptionGroupOptionConfigurationsOptionSettings{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = RDSOptionGroupOptionConfigurationsOptionSettingsList(list)
		return nil
	}
	return err
}

// RDSSecurityGroupRule represents Amazon RDS Security Group Rule
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-rds-security-group-rule.html
type RDSSecurityGroupRule struct {
	// The IP range to authorize.
	CIDRIP *StringExpr `json:"CIDRIP,omitempty"`

	// Id of the VPC or EC2 Security Group to authorize.
	EC2SecurityGroupId *StringExpr `json:"EC2SecurityGroupId,omitempty"`

	// Name of the EC2 Security Group to authorize.
	EC2SecurityGroupName *StringExpr `json:"EC2SecurityGroupName,omitempty"`

	// AWS Account Number of the owner of the EC2 Security Group specified in
	// the EC2SecurityGroupName parameter. The AWS Access Key ID is not an
	// acceptable value.
	EC2SecurityGroupOwnerId *StringExpr `json:"EC2SecurityGroupOwnerId,omitempty"`
}

// RDSSecurityGroupRuleList represents a list of RDSSecurityGroupRule
type RDSSecurityGroupRuleList []RDSSecurityGroupRule

// UnmarshalJSON sets the object from the provided JSON representation
func (l *RDSSecurityGroupRuleList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := RDSSecurityGroupRule{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = RDSSecurityGroupRuleList{item}
		return nil
	}
	list := []RDSSecurityGroupRule{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = RDSSecurityGroupRuleList(list)
		return nil
	}
	return err
}

// Route53AliasTargetProperty represents Route 53 AliasTarget Property
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-route53-aliastarget.html
type Route53AliasTargetProperty struct {
	// The DNS name of the load balancer, the domain name of the CloudFront
	// distribution, the website endpoint of the Amazon S3 bucket, or another
	// record set in the same hosted zone that is the target of the alias.
	DNSName *StringExpr `json:"DNSName,omitempty"`

	// Whether Amazon Route 53 checks the health of the resource record sets
	// in the alias target when responding to DNS queries. For more
	// information about using this property, see EvaluateTargetHealth in the
	// Amazon Route 53 API Reference.
	EvaluateTargetHealth *BoolExpr `json:"EvaluateTargetHealth,omitempty"`

	// The hosted zone ID. For load balancers, use the canonical hosted zone
	// ID of the load balancer. For Amazon S3, use the hosted zone ID for
	// your bucket's website endpoint. For CloudFront, use Z2FDTNDATAQYW2.
	// For examples, see Example: Creating Alias Resource Record Sets in the
	// Amazon Route 53 API Reference.
	HostedZoneId *StringExpr `json:"HostedZoneId,omitempty"`
}

// Route53AliasTargetPropertyList represents a list of Route53AliasTargetProperty
type Route53AliasTargetPropertyList []Route53AliasTargetProperty

// UnmarshalJSON sets the object from the provided JSON representation
func (l *Route53AliasTargetPropertyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := Route53AliasTargetProperty{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = Route53AliasTargetPropertyList{item}
		return nil
	}
	list := []Route53AliasTargetProperty{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = Route53AliasTargetPropertyList(list)
		return nil
	}
	return err
}

// Route53RecordSetGeoLocationProperty represents Amazon Route 53 Record Set GeoLocation Property
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-route53-recordset-geolocation.html
type Route53RecordSetGeoLocationProperty struct {
	// All DNS queries from the continent that you specified are routed to
	// this resource record set. If you specify this property, omit the
	// CountryCode and SubdivisionCode properties.
	ContinentCode *StringExpr `json:"ContinentCode,omitempty"`

	// All DNS queries from the country that you specified are routed to this
	// resource record set. If you specify this property, omit the
	// ContinentCode property.
	CountryCode *StringExpr `json:"CountryCode,omitempty"`

	// If you specified US for the country code, you can specify a state in
	// the United States. All DNS queries from the state that you specified
	// are routed to this resource record set. If you specify this property,
	// you must specify US for the CountryCode and omit the ContinentCode
	// property.
	SubdivisionCode *StringExpr `json:"SubdivisionCode,omitempty"`
}

// Route53RecordSetGeoLocationPropertyList represents a list of Route53RecordSetGeoLocationProperty
type Route53RecordSetGeoLocationPropertyList []Route53RecordSetGeoLocationProperty

// UnmarshalJSON sets the object from the provided JSON representation
func (l *Route53RecordSetGeoLocationPropertyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := Route53RecordSetGeoLocationProperty{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = Route53RecordSetGeoLocationPropertyList{item}
		return nil
	}
	list := []Route53RecordSetGeoLocationProperty{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = Route53RecordSetGeoLocationPropertyList(list)
		return nil
	}
	return err
}

// Route53HealthCheckConfig represents Amazon Route 53 HealthCheckConfig
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-route53-healthcheck-healthcheckconfig.html
type Route53HealthCheckConfig struct {
	// The number of consecutive health checks that an endpoint must pass or
	// fail for Amazon Route 53 to change the current status of the endpoint
	// from unhealthy to healthy or healthy to unhealthy. For more
	// information, see How Amazon Route 53 Determines Whether an Endpoint
	// Is Healthy in the Amazon Route 53 Developer Guide.
	FailureThreshold *IntegerExpr `json:"FailureThreshold,omitempty"`

	// If you specified the IPAddress property, the value that you want
	// Amazon Route 53 to pass in the host header in all health checks
	// except for TCP health checks. If you don't specify an IP address, the
	// domain that Amazon Route 53 sends a DNS request to. Amazon Route 53
	// uses the IP address that the DNS returns to check the health of the
	// endpoint.
	FullyQualifiedDomainName *StringExpr `json:"FullyQualifiedDomainName,omitempty"`

	// The IPv4 IP address of the endpoint on which you want Amazon Route 53
	// to perform health checks. If you don't specify an IP address, Amazon
	// Route 53 sends a DNS request to resolve the domain name that you
	// specify in the FullyQualifiedDomainName property.
	IPAddress *StringExpr `json:"IPAddress,omitempty"`

	// The port on the endpoint on which you want Amazon Route 53 to perform
	// health checks.
	Port *IntegerExpr `json:"Port,omitempty"`

	// The number of seconds between the time that Amazon Route 53 gets a
	// response from your endpoint and the time that it sends the next
	// health-check request. Each Amazon Route 53 health checker makes
	// requests at this interval. For valid values, see the RequestInterval
	// element in the Amazon Route 53 API Reference.
	RequestInterval *IntegerExpr `json:"RequestInterval,omitempty"`

	// The path that you want Amazon Route 53 to request when performing
	// health checks. The path can be any value for which your endpoint
	// returns an HTTP status code of 2xx or 3xx when the endpoint is
	// healthy, such as /docs/route53-health-check.html.
	ResourcePath *StringExpr `json:"ResourcePath,omitempty"`

	// If the value of the Type property is HTTP_STR_MATCH or
	// HTTPS_STR_MATCH, the string that you want Amazon Route 53 to search
	// for in the response body from the specified resource. If the string
	// appears in the response body, Amazon Route 53 considers the resource
	// healthy.
	SearchString *StringExpr `json:"SearchString,omitempty"`

	// The type of health check that you want to create, which indicates how
	// Amazon Route 53 determines whether an endpoint is healthy. You can
	// specify HTTP, HTTPS, HTTP_STR_MATCH, HTTPS_STR_MATCH, or TCP. For
	// information about the different types, see the Type element in the
	// Amazon Route 53 API Reference.
	Type *StringExpr `json:"Type,omitempty"`
}

// Route53HealthCheckConfigList represents a list of Route53HealthCheckConfig
type Route53HealthCheckConfigList []Route53HealthCheckConfig

// UnmarshalJSON sets the object from the provided JSON representation
func (l *Route53HealthCheckConfigList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := Route53HealthCheckConfig{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = Route53HealthCheckConfigList{item}
		return nil
	}
	list := []Route53HealthCheckConfig{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = Route53HealthCheckConfigList(list)
		return nil
	}
	return err
}

// Route53HealthCheckTags represents Amazon Route 53 HealthCheckTags
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-route53-healthcheck-healthchecktags.html
type Route53HealthCheckTags struct {
	// The key name of the tag.
	Key *StringExpr `json:"Key,omitempty"`

	// The value for the tag.
	Value *StringExpr `json:"Value,omitempty"`
}

// Route53HealthCheckTagsList represents a list of Route53HealthCheckTags
type Route53HealthCheckTagsList []Route53HealthCheckTags

// UnmarshalJSON sets the object from the provided JSON representation
func (l *Route53HealthCheckTagsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := Route53HealthCheckTags{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = Route53HealthCheckTagsList{item}
		return nil
	}
	list := []Route53HealthCheckTags{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = Route53HealthCheckTagsList(list)
		return nil
	}
	return err
}

// Route53HostedZoneConfigProperty represents Amazon Route 53 HostedZoneConfig Property
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-route53-hostedzone-hostedzoneconfig.html
type Route53HostedZoneConfigProperty struct {
	// Any comments that you want to include about the hosted zone.
	Comment *StringExpr `json:"Comment,omitempty"`
}

// Route53HostedZoneConfigPropertyList represents a list of Route53HostedZoneConfigProperty
type Route53HostedZoneConfigPropertyList []Route53HostedZoneConfigProperty

// UnmarshalJSON sets the object from the provided JSON representation
func (l *Route53HostedZoneConfigPropertyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := Route53HostedZoneConfigProperty{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = Route53HostedZoneConfigPropertyList{item}
		return nil
	}
	list := []Route53HostedZoneConfigProperty{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = Route53HostedZoneConfigPropertyList(list)
		return nil
	}
	return err
}

// Route53HostedZoneTags represents Amazon Route 53 HostedZoneTags
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-route53-hostedzone-hostedzonetags.html
type Route53HostedZoneTags struct {
	// The key name of the tag.
	Key *StringExpr `json:"Key,omitempty"`

	// The value for the tag.
	Value *StringExpr `json:"Value,omitempty"`
}

// Route53HostedZoneTagsList represents a list of Route53HostedZoneTags
type Route53HostedZoneTagsList []Route53HostedZoneTags

// UnmarshalJSON sets the object from the provided JSON representation
func (l *Route53HostedZoneTagsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := Route53HostedZoneTags{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = Route53HostedZoneTagsList{item}
		return nil
	}
	list := []Route53HostedZoneTags{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = Route53HostedZoneTagsList(list)
		return nil
	}
	return err
}

// Route53HostedZoneVPCs represents Amazon Route 53 HostedZoneVPCs
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-route53-hostedzone-hostedzonevpcs.html
type Route53HostedZoneVPCs struct {
	// The ID of the Amazon VPC that you want to associate with the hosted
	// zone.
	VPCId *StringExpr `json:"VPCId,omitempty"`

	// The region in which the Amazon VPC was created as specified in the
	// VPCId property.
	VPCRegion *StringExpr `json:"VPCRegion,omitempty"`
}

// Route53HostedZoneVPCsList represents a list of Route53HostedZoneVPCs
type Route53HostedZoneVPCsList []Route53HostedZoneVPCs

// UnmarshalJSON sets the object from the provided JSON representation
func (l *Route53HostedZoneVPCsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := Route53HostedZoneVPCs{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = Route53HostedZoneVPCsList{item}
		return nil
	}
	list := []Route53HostedZoneVPCs{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = Route53HostedZoneVPCsList(list)
		return nil
	}
	return err
}

// S3CorsConfiguration represents Amazon S3 Cors Configuration
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-cors.html
type S3CorsConfiguration struct {
	// A set of origins and methods that you allow.
	CorsRules *S3CorsConfigurationRuleList `json:"CorsRules,omitempty"`
}

// S3CorsConfigurationList represents a list of S3CorsConfiguration
type S3CorsConfigurationList []S3CorsConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *S3CorsConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := S3CorsConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = S3CorsConfigurationList{item}
		return nil
	}
	list := []S3CorsConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = S3CorsConfigurationList(list)
		return nil
	}
	return err
}

// S3CorsConfigurationRule represents Amazon S3 Cors Configuration Rule
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-cors-corsrule.html
type S3CorsConfigurationRule struct {
	// Headers that are specified in the Access-Control-Request-Headers
	// header. These headers are allowed in a preflight OPTIONS request. In
	// response to any preflight OPTIONS request, Amazon S3 returns any
	// requested headers that are allowed.
	AllowedHeaders *StringListExpr `json:"AllowedHeaders,omitempty"`

	// An HTTP method that you allow the origin to execute. The valid values
	// are GET, PUT, HEAD, POST, and DELETE.
	AllowedMethods *StringListExpr `json:"AllowedMethods,omitempty"`

	// An origin that you allow to send cross-domain requests.
	AllowedOrigins *StringListExpr `json:"AllowedOrigins,omitempty"`

	// One or more headers in the response that are accessible to client
	// applications (for example, from a JavaScript XMLHttpRequest object).
	ExposedHeaders *StringListExpr `json:"ExposedHeaders,omitempty"`

	// A unique identifier for this rule. The value cannot be more than 255
	// characters.
	Id *StringExpr `json:"Id,omitempty"`

	// The time in seconds that your browser is to cache the preflight
	// response for the specified resource.
	MaxAge *IntegerExpr `json:"MaxAge,omitempty"`
}

// S3CorsConfigurationRuleList represents a list of S3CorsConfigurationRule
type S3CorsConfigurationRuleList []S3CorsConfigurationRule

// UnmarshalJSON sets the object from the provided JSON representation
func (l *S3CorsConfigurationRuleList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := S3CorsConfigurationRule{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = S3CorsConfigurationRuleList{item}
		return nil
	}
	list := []S3CorsConfigurationRule{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = S3CorsConfigurationRuleList(list)
		return nil
	}
	return err
}

// S3LifecycleConfiguration represents Amazon S3 Lifecycle Configuration
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-lifecycleconfig.html
type S3LifecycleConfiguration struct {
	// A lifecycle rule for individual objects in an S3 bucket.
	Rules *S3LifecycleRuleList `json:"Rules,omitempty"`
}

// S3LifecycleConfigurationList represents a list of S3LifecycleConfiguration
type S3LifecycleConfigurationList []S3LifecycleConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *S3LifecycleConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := S3LifecycleConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = S3LifecycleConfigurationList{item}
		return nil
	}
	list := []S3LifecycleConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = S3LifecycleConfigurationList(list)
		return nil
	}
	return err
}

// S3LifecycleRule represents Amazon S3 Lifecycle Rule
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-lifecycleconfig-rule.html
type S3LifecycleRule struct {
	// Indicates when objects are deleted from Amazon S3 and Amazon Glacier.
	// The date value must be in ISO 8601 format. The time is always midnight
	// UTC. If you specify an expiration and transition time, you must use
	// the same time unit for both properties (either in days or by date).
	// The expiration time must also be later than the transition time.
	ExpirationDate *StringExpr `json:"ExpirationDate,omitempty"`

	// Indicates the number of days after creation when objects are deleted
	// from Amazon S3 and Amazon Glacier. If you specify an expiration and
	// transition time, you must use the same time unit for both properties
	// (either in days or by date). The expiration time must also be later
	// than the transition time.
	ExpirationInDays *IntegerExpr `json:"ExpirationInDays,omitempty"`

	// A unique identifier for this rule. The value cannot be more than 255
	// characters.
	Id *StringExpr `json:"Id,omitempty"`

	// For buckets with versioning enabled (or suspended), specifies the
	// time, in days, between when a new version of the object is uploaded to
	// the bucket and when old versions of the object expire. When object
	// versions expire, Amazon S3 permanently deletes them. If you specify a
	// transition and expiration time, the expiration time must be later than
	// the transition time.
	NoncurrentVersionExpirationInDays *IntegerExpr `json:"NoncurrentVersionExpirationInDays,omitempty"`

	// For buckets with versioning enabled (or suspended), specifies when
	// non-current objects transition to a specified storage class. If you
	// specify a transition and expiration time, the expiration time must be
	// later than the transition time. If you specify this property, don't
	// specify the NoncurrentVersionTransitions property.
	NoncurrentVersionTransitionXXDeprecatedX *S3LifecycleRuleNoncurrentVersionTransition `json:"NoncurrentVersionTransition (deprecated),omitempty"`

	// For buckets with versioning enabled (or suspended), one or more
	// transition rules that specify when non-current objects transition to a
	// specified storage class. If you specify a transition and expiration
	// time, the expiration time must be later than the transition time. If
	// you specify this property, don't specify the
	// NoncurrentVersionTransition property.
	NoncurrentVersionTransitions *S3LifecycleRuleNoncurrentVersionTransitionList `json:"NoncurrentVersionTransitions,omitempty"`

	// Object key prefix that identifies one or more objects to which this
	// rule applies.
	Prefix *StringExpr `json:"Prefix,omitempty"`

	// Specify either Enabled or Disabled. If you specify Enabled, Amazon S3
	// executes this rule as scheduled. If you specify Disabled, Amazon S3
	// ignores this rule.
	Status *StringExpr `json:"Status,omitempty"`

	// Specifies when an object transitions to a specified storage class. If
	// you specify an expiration and transition time, you must use the same
	// time unit for both properties (either in days or by date). The
	// expiration time must also be later than the transition time. If you
	// specify this property, don't specify the Transitions property.
	TransitionXXDeprecatedX *S3LifecycleRuleTransition `json:"Transition (deprecated),omitempty"`

	// One or more transition rules that specify when an object transitions
	// to a specified storage class. If you specify an expiration and
	// transition time, you must use the same time unit for both properties
	// (either in days or by date). The expiration time must also be later
	// than the transition time. If you specify this property, don't specify
	// the Transition property.
	Transitions *S3LifecycleRuleTransitionList `json:"Transitions,omitempty"`
}

// S3LifecycleRuleList represents a list of S3LifecycleRule
type S3LifecycleRuleList []S3LifecycleRule

// UnmarshalJSON sets the object from the provided JSON representation
func (l *S3LifecycleRuleList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := S3LifecycleRule{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = S3LifecycleRuleList{item}
		return nil
	}
	list := []S3LifecycleRule{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = S3LifecycleRuleList(list)
		return nil
	}
	return err
}

// S3LifecycleRuleNoncurrentVersionTransition represents Amazon S3 Lifecycle Rule NoncurrentVersionTransition
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-lifecycleconfig-rule-noncurrentversiontransition.html
type S3LifecycleRuleNoncurrentVersionTransition struct {
	// The storage class to which you want the object to transition, such as
	// GLACIER. For valid values, see the StorageClass request element of the
	// PUT Bucket lifecycle action in the Amazon Simple Storage Service API
	// Reference.
	StorageClass *StringExpr `json:"StorageClass,omitempty"`

	// The number of days between the time that a new version of the object
	// is uploaded to the bucket and when old versions of the object are
	// transitioned to the specified storage class.
	TransitionInDays *IntegerExpr `json:"TransitionInDays,omitempty"`
}

// S3LifecycleRuleNoncurrentVersionTransitionList represents a list of S3LifecycleRuleNoncurrentVersionTransition
type S3LifecycleRuleNoncurrentVersionTransitionList []S3LifecycleRuleNoncurrentVersionTransition

// UnmarshalJSON sets the object from the provided JSON representation
func (l *S3LifecycleRuleNoncurrentVersionTransitionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := S3LifecycleRuleNoncurrentVersionTransition{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = S3LifecycleRuleNoncurrentVersionTransitionList{item}
		return nil
	}
	list := []S3LifecycleRuleNoncurrentVersionTransition{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = S3LifecycleRuleNoncurrentVersionTransitionList(list)
		return nil
	}
	return err
}

// S3LifecycleRuleTransition represents Amazon S3 Lifecycle Rule Transition
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-lifecycleconfig-rule-transition.html
type S3LifecycleRuleTransition struct {
	// The storage class to which you want the object to transition, such as
	// GLACIER. For valid values, see the StorageClass request element of the
	// PUT Bucket lifecycle action in the Amazon Simple Storage Service API
	// Reference.
	StorageClass *StringExpr `json:"StorageClass,omitempty"`

	// Indicates when objects are transitioned to the specified storage
	// class. The date value must be in ISO 8601 format. The time is always
	// midnight UTC.
	TransitionDate *StringExpr `json:"TransitionDate,omitempty"`

	// Indicates the number of days after creation when objects are
	// transitioned to the specified storage class.
	TransitionInDays *IntegerExpr `json:"TransitionInDays,omitempty"`
}

// S3LifecycleRuleTransitionList represents a list of S3LifecycleRuleTransition
type S3LifecycleRuleTransitionList []S3LifecycleRuleTransition

// UnmarshalJSON sets the object from the provided JSON representation
func (l *S3LifecycleRuleTransitionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := S3LifecycleRuleTransition{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = S3LifecycleRuleTransitionList{item}
		return nil
	}
	list := []S3LifecycleRuleTransition{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = S3LifecycleRuleTransitionList(list)
		return nil
	}
	return err
}

// S3LoggingConfiguration represents Amazon S3 Logging Configuration
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-loggingconfig.html
type S3LoggingConfiguration struct {
	// The name of an Amazon S3 bucket where Amazon S3 store server access
	// log files. You can store log files in any bucket that you own. By
	// default, logs are stored in the bucket where the LoggingConfiguration
	// property is defined.
	DestinationBucketName *StringExpr `json:"DestinationBucketName,omitempty"`

	// A prefix for the all log object keys. If you store log files from
	// multiple Amazon S3 buckets in a single bucket, you can use a prefix to
	// distinguish which log files came from which bucket.
	LogFilePrefix *StringExpr `json:"LogFilePrefix,omitempty"`
}

// S3LoggingConfigurationList represents a list of S3LoggingConfiguration
type S3LoggingConfigurationList []S3LoggingConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *S3LoggingConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := S3LoggingConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = S3LoggingConfigurationList{item}
		return nil
	}
	list := []S3LoggingConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = S3LoggingConfigurationList(list)
		return nil
	}
	return err
}

// S3NotificationConfiguration represents Amazon S3 NotificationConfiguration
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-notificationconfig.html
type S3NotificationConfiguration struct {
	// The AWS Lambda functions to invoke and the events for which to invoke
	// the functions.
	LambdaConfigurations *SimpleStorageServiceNotificationConfigurationLambdaConfigurationsList `json:"LambdaConfigurations,omitempty"`

	// The Amazon Simple Queue Service queues to publish messages to and the
	// events for which to publish messages.
	QueueConfigurations *SimpleStorageServiceNotificationConfigurationQueueConfigurationsList `json:"QueueConfigurations,omitempty"`

	// The topic to which notifications are sent and the events for which
	// notification are generated.
	TopicConfigurations *S3NotificationConfigurationTopicConfigurationsList `json:"TopicConfigurations,omitempty"`
}

// S3NotificationConfigurationList represents a list of S3NotificationConfiguration
type S3NotificationConfigurationList []S3NotificationConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *S3NotificationConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := S3NotificationConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = S3NotificationConfigurationList{item}
		return nil
	}
	list := []S3NotificationConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = S3NotificationConfigurationList(list)
		return nil
	}
	return err
}

// S3NotificationConfigurationConfigFilter represents Amazon S3 NotificationConfiguration Config Filter
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-notificationconfiguration-config-filter.html
type S3NotificationConfigurationConfigFilter struct {
	// Amazon S3 filtering rules that describe for which object key names to
	// send notifications.
	S3Key *S3NotificationConfigurationConfigFilterS3Key `json:"S3Key,omitempty"`
}

// S3NotificationConfigurationConfigFilterList represents a list of S3NotificationConfigurationConfigFilter
type S3NotificationConfigurationConfigFilterList []S3NotificationConfigurationConfigFilter

// UnmarshalJSON sets the object from the provided JSON representation
func (l *S3NotificationConfigurationConfigFilterList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := S3NotificationConfigurationConfigFilter{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = S3NotificationConfigurationConfigFilterList{item}
		return nil
	}
	list := []S3NotificationConfigurationConfigFilter{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = S3NotificationConfigurationConfigFilterList(list)
		return nil
	}
	return err
}

// S3NotificationConfigurationConfigFilterS3Key represents Amazon S3 NotificationConfiguration Config Filter S3Key
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-notificationconfiguration-config-filter-s3key.html
type S3NotificationConfigurationConfigFilterS3Key struct {
	// The object key name to filter on and whether to filter on the suffix
	// or prefix of the key name.
	Rules *S3NotificationConfigurationConfigFilterS3KeyRulesList `json:"Rules,omitempty"`
}

// S3NotificationConfigurationConfigFilterS3KeyList represents a list of S3NotificationConfigurationConfigFilterS3Key
type S3NotificationConfigurationConfigFilterS3KeyList []S3NotificationConfigurationConfigFilterS3Key

// UnmarshalJSON sets the object from the provided JSON representation
func (l *S3NotificationConfigurationConfigFilterS3KeyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := S3NotificationConfigurationConfigFilterS3Key{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = S3NotificationConfigurationConfigFilterS3KeyList{item}
		return nil
	}
	list := []S3NotificationConfigurationConfigFilterS3Key{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = S3NotificationConfigurationConfigFilterS3KeyList(list)
		return nil
	}
	return err
}

// S3NotificationConfigurationConfigFilterS3KeyRules represents Amazon S3 NotificationConfiguration Config Filter S3Key Rules
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-notificationconfiguration-config-filter-s3key-rules.html
type S3NotificationConfigurationConfigFilterS3KeyRules struct {
	// Whether the filter matches the prefix or suffix of object key names.
	// For valid values, see the Name request element of the PUT Bucket
	// notification action in the Amazon Simple Storage Service API
	// Reference.
	Name *StringExpr `json:"Name,omitempty"`

	// The value that the filter searches for in object key names.
	Value *StringExpr `json:"Value,omitempty"`
}

// S3NotificationConfigurationConfigFilterS3KeyRulesList represents a list of S3NotificationConfigurationConfigFilterS3KeyRules
type S3NotificationConfigurationConfigFilterS3KeyRulesList []S3NotificationConfigurationConfigFilterS3KeyRules

// UnmarshalJSON sets the object from the provided JSON representation
func (l *S3NotificationConfigurationConfigFilterS3KeyRulesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := S3NotificationConfigurationConfigFilterS3KeyRules{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = S3NotificationConfigurationConfigFilterS3KeyRulesList{item}
		return nil
	}
	list := []S3NotificationConfigurationConfigFilterS3KeyRules{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = S3NotificationConfigurationConfigFilterS3KeyRulesList(list)
		return nil
	}
	return err
}

// SimpleStorageServiceNotificationConfigurationLambdaConfigurations represents Amazon Simple Storage Service NotificationConfiguration LambdaConfigurations
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-notificationconfig-lambdaconfig.html
type SimpleStorageServiceNotificationConfigurationLambdaConfigurations struct {
	// The S3 bucket event for which to invoke the Lambda function. For more
	// information, see Supported Event Types in the Amazon Simple Storage
	// Service Developer Guide.
	Event *StringExpr `json:"Event,omitempty"`

	// The filtering rules that determine which objects invoke the Lambda
	// function. For example, you can create a filter so that only image
	// files with a .jpg extension invoke the function when they are added to
	// the S3 bucket.
	Filter *S3NotificationConfigurationConfigFilter `json:"Filter,omitempty"`

	// The Amazon Resource Name (ARN) of the Lambda function that Amazon S3
	// invokes when the specified event type occurs.
	Function *StringExpr `json:"Function,omitempty"`
}

// SimpleStorageServiceNotificationConfigurationLambdaConfigurationsList represents a list of SimpleStorageServiceNotificationConfigurationLambdaConfigurations
type SimpleStorageServiceNotificationConfigurationLambdaConfigurationsList []SimpleStorageServiceNotificationConfigurationLambdaConfigurations

// UnmarshalJSON sets the object from the provided JSON representation
func (l *SimpleStorageServiceNotificationConfigurationLambdaConfigurationsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := SimpleStorageServiceNotificationConfigurationLambdaConfigurations{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = SimpleStorageServiceNotificationConfigurationLambdaConfigurationsList{item}
		return nil
	}
	list := []SimpleStorageServiceNotificationConfigurationLambdaConfigurations{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = SimpleStorageServiceNotificationConfigurationLambdaConfigurationsList(list)
		return nil
	}
	return err
}

// SimpleStorageServiceNotificationConfigurationQueueConfigurations represents Amazon Simple Storage Service NotificationConfiguration QueueConfigurations
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-notificationconfig-queueconfig.html
type SimpleStorageServiceNotificationConfigurationQueueConfigurations struct {
	// The S3 bucket event about which you want to publish messages to Amazon
	// Simple Queue Service ( Amazon SQS). For more information, see
	// Supported Event Types in the Amazon Simple Storage Service Developer
	// Guide.
	Event *StringExpr `json:"Event,omitempty"`

	// The filtering rules that determine for which objects to send
	// notifications. For example, you can create a filter so that Amazon
	// Simple Storage Service (Amazon S3) sends notifications only when image
	// files with a .jpg extension are added to the bucket.
	Filter *S3NotificationConfigurationConfigFilter `json:"Filter,omitempty"`

	// The Amazon Resource Name (ARN) of the Amazon SQS queue that Amazon S3
	// publishes messages to when the specified event type occurs.
	Queue *StringExpr `json:"Queue,omitempty"`
}

// SimpleStorageServiceNotificationConfigurationQueueConfigurationsList represents a list of SimpleStorageServiceNotificationConfigurationQueueConfigurations
type SimpleStorageServiceNotificationConfigurationQueueConfigurationsList []SimpleStorageServiceNotificationConfigurationQueueConfigurations

// UnmarshalJSON sets the object from the provided JSON representation
func (l *SimpleStorageServiceNotificationConfigurationQueueConfigurationsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := SimpleStorageServiceNotificationConfigurationQueueConfigurations{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = SimpleStorageServiceNotificationConfigurationQueueConfigurationsList{item}
		return nil
	}
	list := []SimpleStorageServiceNotificationConfigurationQueueConfigurations{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = SimpleStorageServiceNotificationConfigurationQueueConfigurationsList(list)
		return nil
	}
	return err
}

// S3NotificationConfigurationTopicConfigurations represents Amazon S3 NotificationConfiguration TopicConfigurations
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-notificationconfig-topicconfig.html
type S3NotificationConfigurationTopicConfigurations struct {
	// The Amazon Simple Storage Service (Amazon S3) bucket event about which
	// to send notifications. For more information, see Supported Event Types
	// in the Amazon Simple Storage Service Developer Guide.
	Event *StringExpr `json:"Event,omitempty"`

	// The filtering rules that determine for which objects to send
	// notifications. For example, you can create a filter so that Amazon
	// Simple Storage Service (Amazon S3) sends notifications only when image
	// files with a .jpg extension are added to the bucket.
	Filter *S3NotificationConfigurationConfigFilter `json:"Filter,omitempty"`

	// The Amazon SNS topic Amazon Resource Name (ARN) to which Amazon S3
	// reports the specified events.
	Topic *StringExpr `json:"Topic,omitempty"`
}

// S3NotificationConfigurationTopicConfigurationsList represents a list of S3NotificationConfigurationTopicConfigurations
type S3NotificationConfigurationTopicConfigurationsList []S3NotificationConfigurationTopicConfigurations

// UnmarshalJSON sets the object from the provided JSON representation
func (l *S3NotificationConfigurationTopicConfigurationsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := S3NotificationConfigurationTopicConfigurations{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = S3NotificationConfigurationTopicConfigurationsList{item}
		return nil
	}
	list := []S3NotificationConfigurationTopicConfigurations{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = S3NotificationConfigurationTopicConfigurationsList(list)
		return nil
	}
	return err
}

// S3ReplicationConfiguration represents Amazon S3 ReplicationConfiguration
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-replicationconfiguration.html
type S3ReplicationConfiguration struct {
	// The Amazon Resource Name (ARN) of an AWS Identity and Access
	// Management (IAM) role that Amazon S3 assumes when replicating objects.
	// For more information, see How to Set Up Cross-Region Replication in
	// the Amazon Simple Storage Service Developer Guide.
	Role *StringExpr `json:"Role,omitempty"`

	// A replication rule that specifies which objects to replicate and where
	// they are stored.
	Rules *S3ReplicationConfigurationRulesList `json:"Rules,omitempty"`
}

// S3ReplicationConfigurationList represents a list of S3ReplicationConfiguration
type S3ReplicationConfigurationList []S3ReplicationConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *S3ReplicationConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := S3ReplicationConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = S3ReplicationConfigurationList{item}
		return nil
	}
	list := []S3ReplicationConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = S3ReplicationConfigurationList(list)
		return nil
	}
	return err
}

// S3ReplicationConfigurationRules represents Amazon S3 ReplicationConfiguration Rules
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-replicationconfiguration-rules.html
type S3ReplicationConfigurationRules struct {
	// Defines the destination where Amazon S3 stores replicated objects.
	Destination *S3ReplicationConfigurationRulesDestination `json:"Destination,omitempty"`

	// A unique identifier for the rule. If you don't specify a value, AWS
	// CloudFormation generates a random ID.
	Id *StringExpr `json:"Id,omitempty"`

	// An object prefix. This rule applies to all Amazon S3 objects with this
	// prefix. To specify all objects in an S3 bucket, specify an empty
	// string.
	Prefix *StringExpr `json:"Prefix,omitempty"`

	// Whether the rule is enabled. For valid values, see the Status element
	// of the PUT Bucket replication action in the Amazon Simple Storage
	// Service API Reference.
	Status *StringExpr `json:"Status,omitempty"`
}

// S3ReplicationConfigurationRulesList represents a list of S3ReplicationConfigurationRules
type S3ReplicationConfigurationRulesList []S3ReplicationConfigurationRules

// UnmarshalJSON sets the object from the provided JSON representation
func (l *S3ReplicationConfigurationRulesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := S3ReplicationConfigurationRules{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = S3ReplicationConfigurationRulesList{item}
		return nil
	}
	list := []S3ReplicationConfigurationRules{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = S3ReplicationConfigurationRulesList(list)
		return nil
	}
	return err
}

// S3ReplicationConfigurationRulesDestination represents Amazon S3 ReplicationConfiguration Rules Destination
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-replicationconfiguration-rules-destination.html
type S3ReplicationConfigurationRulesDestination struct {
	// The Amazon resource name (ARN) of an S3 bucket where Amazon S3 stores
	// replicated objects. This destination bucket must be in a different
	// region than your source bucket.
	Bucket *StringExpr `json:"Bucket,omitempty"`

	// The storage class to use when replicating objects, such as standard or
	// reduced redundancy. By default, Amazon S3 uses the storage class of
	// the source object to create object replica. For valid values, see the
	// StorageClass element of the PUT Bucket replication action in the
	// Amazon Simple Storage Service API Reference.
	StorageClass *StringExpr `json:"StorageClass,omitempty"`
}

// S3ReplicationConfigurationRulesDestinationList represents a list of S3ReplicationConfigurationRulesDestination
type S3ReplicationConfigurationRulesDestinationList []S3ReplicationConfigurationRulesDestination

// UnmarshalJSON sets the object from the provided JSON representation
func (l *S3ReplicationConfigurationRulesDestinationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := S3ReplicationConfigurationRulesDestination{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = S3ReplicationConfigurationRulesDestinationList{item}
		return nil
	}
	list := []S3ReplicationConfigurationRulesDestination{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = S3ReplicationConfigurationRulesDestinationList(list)
		return nil
	}
	return err
}

// S3VersioningConfiguration represents Amazon S3 Versioning Configuration
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-versioningconfig.html
type S3VersioningConfiguration struct {
	// The versioning state of an Amazon S3 bucket. If you enable versioning,
	// you must suspend versioning to disable it.
	Status *StringExpr `json:"Status,omitempty"`
}

// S3VersioningConfigurationList represents a list of S3VersioningConfiguration
type S3VersioningConfigurationList []S3VersioningConfiguration

// UnmarshalJSON sets the object from the provided JSON representation
func (l *S3VersioningConfigurationList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := S3VersioningConfiguration{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = S3VersioningConfigurationList{item}
		return nil
	}
	list := []S3VersioningConfiguration{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = S3VersioningConfigurationList(list)
		return nil
	}
	return err
}

// S3WebsiteConfigurationProperty represents Amazon S3 Website Configuration Property
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-websiteconfiguration.html
type S3WebsiteConfigurationProperty struct {
	// The name of the error document for the website.
	ErrorDocument *StringExpr `json:"ErrorDocument,omitempty"`

	// The name of the index document for the website.
	IndexDocument *StringExpr `json:"IndexDocument,omitempty"`

	// The redirect behavior for every request to this bucket's website
	// endpoint.
	RedirectAllRequestsTo *S3WebsiteConfigurationRedirectAllRequestsToProperty `json:"RedirectAllRequestsTo,omitempty"`

	// Rules that define when a redirect is applied and the redirect
	// behavior.
	RoutingRules *S3WebsiteConfigurationRoutingRulesPropertyList `json:"RoutingRules,omitempty"`
}

// S3WebsiteConfigurationPropertyList represents a list of S3WebsiteConfigurationProperty
type S3WebsiteConfigurationPropertyList []S3WebsiteConfigurationProperty

// UnmarshalJSON sets the object from the provided JSON representation
func (l *S3WebsiteConfigurationPropertyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := S3WebsiteConfigurationProperty{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = S3WebsiteConfigurationPropertyList{item}
		return nil
	}
	list := []S3WebsiteConfigurationProperty{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = S3WebsiteConfigurationPropertyList(list)
		return nil
	}
	return err
}

// S3WebsiteConfigurationRedirectAllRequestsToProperty represents Amazon S3 Website Configuration Redirect All Requests To Property
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-websiteconfiguration-redirectallrequeststo.html
type S3WebsiteConfigurationRedirectAllRequestsToProperty struct {
	// Name of the host where requests are redirected.
	HostName *StringExpr `json:"HostName,omitempty"`

	// Protocol to use (http or https) when redirecting requests. The default
	// is the protocol that is used in the original request.
	Protocol *StringExpr `json:"Protocol,omitempty"`
}

// S3WebsiteConfigurationRedirectAllRequestsToPropertyList represents a list of S3WebsiteConfigurationRedirectAllRequestsToProperty
type S3WebsiteConfigurationRedirectAllRequestsToPropertyList []S3WebsiteConfigurationRedirectAllRequestsToProperty

// UnmarshalJSON sets the object from the provided JSON representation
func (l *S3WebsiteConfigurationRedirectAllRequestsToPropertyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := S3WebsiteConfigurationRedirectAllRequestsToProperty{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = S3WebsiteConfigurationRedirectAllRequestsToPropertyList{item}
		return nil
	}
	list := []S3WebsiteConfigurationRedirectAllRequestsToProperty{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = S3WebsiteConfigurationRedirectAllRequestsToPropertyList(list)
		return nil
	}
	return err
}

// S3WebsiteConfigurationRoutingRulesProperty represents Amazon S3 Website Configuration Routing Rules Property
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-websiteconfiguration-routingrules.html
type S3WebsiteConfigurationRoutingRulesProperty struct {
	// Redirect requests to another host, to another page, or with another
	// protocol.
	RedirectRule *S3WebsiteConfigurationRoutingRulesRedirectRuleProperty `json:"RedirectRule,omitempty"`

	// Rules that define when a redirect is applied.
	RoutingRuleCondition *S3WebsiteConfigurationRoutingRulesRoutingRuleConditionProperty `json:"RoutingRuleCondition,omitempty"`
}

// S3WebsiteConfigurationRoutingRulesPropertyList represents a list of S3WebsiteConfigurationRoutingRulesProperty
type S3WebsiteConfigurationRoutingRulesPropertyList []S3WebsiteConfigurationRoutingRulesProperty

// UnmarshalJSON sets the object from the provided JSON representation
func (l *S3WebsiteConfigurationRoutingRulesPropertyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := S3WebsiteConfigurationRoutingRulesProperty{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = S3WebsiteConfigurationRoutingRulesPropertyList{item}
		return nil
	}
	list := []S3WebsiteConfigurationRoutingRulesProperty{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = S3WebsiteConfigurationRoutingRulesPropertyList(list)
		return nil
	}
	return err
}

// S3WebsiteConfigurationRoutingRulesRedirectRuleProperty represents Amazon S3 Website Configuration Routing Rules Redirect Rule Property
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-websiteconfiguration-routingrules-redirectrule.html
type S3WebsiteConfigurationRoutingRulesRedirectRuleProperty struct {
	// Name of the host where requests are redirected.
	HostName *StringExpr `json:"HostName,omitempty"`

	// The HTTP redirect code to use on the response.
	HttpRedirectCode *StringExpr `json:"HttpRedirectCode,omitempty"`

	// The protocol to use in the redirect request.
	Protocol *StringExpr `json:"Protocol,omitempty"`

	// The object key prefix to use in the redirect request. For example, to
	// redirect requests for all pages with the prefix docs/ (objects in the
	// docs/ folder) to the documents/ prefix, you can set the
	// KeyPrefixEquals property in routing condition property to docs/, and
	// set the ReplaceKeyPrefixWith property to documents/.
	ReplaceKeyPrefixWith *StringExpr `json:"ReplaceKeyPrefixWith,omitempty"`

	// The specific object key to use in the redirect request. For example,
	// redirect request to error.html.
	ReplaceKeyWith *StringExpr `json:"ReplaceKeyWith,omitempty"`
}

// S3WebsiteConfigurationRoutingRulesRedirectRulePropertyList represents a list of S3WebsiteConfigurationRoutingRulesRedirectRuleProperty
type S3WebsiteConfigurationRoutingRulesRedirectRulePropertyList []S3WebsiteConfigurationRoutingRulesRedirectRuleProperty

// UnmarshalJSON sets the object from the provided JSON representation
func (l *S3WebsiteConfigurationRoutingRulesRedirectRulePropertyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := S3WebsiteConfigurationRoutingRulesRedirectRuleProperty{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = S3WebsiteConfigurationRoutingRulesRedirectRulePropertyList{item}
		return nil
	}
	list := []S3WebsiteConfigurationRoutingRulesRedirectRuleProperty{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = S3WebsiteConfigurationRoutingRulesRedirectRulePropertyList(list)
		return nil
	}
	return err
}

// S3WebsiteConfigurationRoutingRulesRoutingRuleConditionProperty represents Amazon S3 Website Configuration Routing Rules Routing Rule Condition Property
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-websiteconfiguration-routingrules-routingrulecondition.html
type S3WebsiteConfigurationRoutingRulesRoutingRuleConditionProperty struct {
	// Applies this redirect if the error code equals this value in the event
	// of an error.
	HttpErrorCodeReturnedEquals *StringExpr `json:"HttpErrorCodeReturnedEquals,omitempty"`

	// The object key name prefix when the redirect is applied. For example,
	// to redirect requests for ExamplePage.html, set the key prefix to
	// ExamplePage.html. To redirect request for all pages with the prefix
	// docs/, set the key prefix to docs/, which identifies all objects in
	// the docs/ folder.
	KeyPrefixEquals *StringExpr `json:"KeyPrefixEquals,omitempty"`
}

// S3WebsiteConfigurationRoutingRulesRoutingRuleConditionPropertyList represents a list of S3WebsiteConfigurationRoutingRulesRoutingRuleConditionProperty
type S3WebsiteConfigurationRoutingRulesRoutingRuleConditionPropertyList []S3WebsiteConfigurationRoutingRulesRoutingRuleConditionProperty

// UnmarshalJSON sets the object from the provided JSON representation
func (l *S3WebsiteConfigurationRoutingRulesRoutingRuleConditionPropertyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := S3WebsiteConfigurationRoutingRulesRoutingRuleConditionProperty{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = S3WebsiteConfigurationRoutingRulesRoutingRuleConditionPropertyList{item}
		return nil
	}
	list := []S3WebsiteConfigurationRoutingRulesRoutingRuleConditionProperty{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = S3WebsiteConfigurationRoutingRulesRoutingRuleConditionPropertyList(list)
		return nil
	}
	return err
}

// SNSSubscriptionProperty represents Amazon SNS Subscription Property Type
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-sns-subscription.html
type SNSSubscriptionProperty struct {
	// The subscription's endpoint (format depends on the protocol). For more
	// information, see the Subscribe Endpoint parameter in the Amazon Simple
	// Notification Service API Reference.
	Endpoint *StringExpr `json:"Endpoint,omitempty"`

	// The subscription's protocol. For more information, see the Subscribe
	// Protocol parameter in the Amazon Simple Notification Service API
	// Reference.
	Protocol *StringExpr `json:"Protocol,omitempty"`
}

// SNSSubscriptionPropertyList represents a list of SNSSubscriptionProperty
type SNSSubscriptionPropertyList []SNSSubscriptionProperty

// UnmarshalJSON sets the object from the provided JSON representation
func (l *SNSSubscriptionPropertyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := SNSSubscriptionProperty{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = SNSSubscriptionPropertyList{item}
		return nil
	}
	list := []SNSSubscriptionProperty{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = SNSSubscriptionPropertyList(list)
		return nil
	}
	return err
}

// SQSRedrivePolicy represents Amazon SQS RedrivePolicy
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-sqs-queues-redrivepolicy.html
type SQSRedrivePolicy struct {
	// The Amazon Resource Name (ARN) of the dead letter queue to which the
	// messages are sent to after the maxReceiveCount value has been
	// exceeded.
	DeadLetterTargetArn *StringExpr `json:"deadLetterTargetArn,omitempty"`

	// The number of times a message is delivered to the source queue before
	// being sent to the dead letter queue.
	MaxReceiveCount *IntegerExpr `json:"maxReceiveCount,omitempty"`
}

// SQSRedrivePolicyList represents a list of SQSRedrivePolicy
type SQSRedrivePolicyList []SQSRedrivePolicy

// UnmarshalJSON sets the object from the provided JSON representation
func (l *SQSRedrivePolicyList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := SQSRedrivePolicy{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = SQSRedrivePolicyList{item}
		return nil
	}
	list := []SQSRedrivePolicy{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = SQSRedrivePolicyList(list)
		return nil
	}
	return err
}

// WAFByteMatchSetByteMatchTuples represents AWS WAF ByteMatchSet ByteMatchTuples
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-waf-bytematchset-bytematchtuples.html
type WAFByteMatchSetByteMatchTuples struct {
	// The part of a web request that you want AWS WAF to search, such as a
	// specific header or a query string.
	FieldToMatch *WAFByteMatchSetByteMatchTuplesFieldToMatch `json:"FieldToMatch,omitempty"`

	// How AWS WAF finds matches within the web request part in which you are
	// searching. For valid values, see the PositionalConstraint content for
	// the ByteMatchTuple data type in the AWS WAF API Reference.
	PositionalConstraint *StringExpr `json:"PositionalConstraint,omitempty"`

	// The value that AWS WAF searches for. AWS CloudFormation base64 encodes
	// this value before sending it to AWS WAF.
	TargetString *StringExpr `json:"TargetString,omitempty"`

	// The base64-encoded value that AWS WAF searches for. AWS CloudFormation
	// sends this value to AWS WAF without encoding it.
	TargetStringBase64 *StringExpr `json:"TargetStringBase64,omitempty"`

	// Specifies how AWS WAF processes the target string value. Text
	// transformations eliminate some of the unusual formatting that
	// attackers use in web requests in an effort to bypass AWS WAF. If you
	// specify a transformation, AWS WAF transforms the target string value
	// before inspecting a web request for a match.
	TextTransformation *StringExpr `json:"TextTransformation,omitempty"`
}

// WAFByteMatchSetByteMatchTuplesList represents a list of WAFByteMatchSetByteMatchTuples
type WAFByteMatchSetByteMatchTuplesList []WAFByteMatchSetByteMatchTuples

// UnmarshalJSON sets the object from the provided JSON representation
func (l *WAFByteMatchSetByteMatchTuplesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := WAFByteMatchSetByteMatchTuples{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = WAFByteMatchSetByteMatchTuplesList{item}
		return nil
	}
	list := []WAFByteMatchSetByteMatchTuples{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = WAFByteMatchSetByteMatchTuplesList(list)
		return nil
	}
	return err
}

// WAFByteMatchSetByteMatchTuplesFieldToMatch represents AWS WAF ByteMatchSet ByteMatchTuples FieldToMatch
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-waf-bytematchset-bytematchtuples-fieldtomatch.html
type WAFByteMatchSetByteMatchTuplesFieldToMatch struct {
	// If you specify HEADER for the Type property, the name of the header
	// that AWS WAF searches for, such as User-Agent or Referer. If you
	// specify any other value for the Type property, do not specify this
	// property.
	Data *StringExpr `json:"Data,omitempty"`

	// The part of the web request in which AWS WAF searches for the target
	// string. For valid values, see FieldToMatch in the AWS WAF API
	// Reference.
	Type *StringExpr `json:"Type,omitempty"`
}

// WAFByteMatchSetByteMatchTuplesFieldToMatchList represents a list of WAFByteMatchSetByteMatchTuplesFieldToMatch
type WAFByteMatchSetByteMatchTuplesFieldToMatchList []WAFByteMatchSetByteMatchTuplesFieldToMatch

// UnmarshalJSON sets the object from the provided JSON representation
func (l *WAFByteMatchSetByteMatchTuplesFieldToMatchList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := WAFByteMatchSetByteMatchTuplesFieldToMatch{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = WAFByteMatchSetByteMatchTuplesFieldToMatchList{item}
		return nil
	}
	list := []WAFByteMatchSetByteMatchTuplesFieldToMatch{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = WAFByteMatchSetByteMatchTuplesFieldToMatchList(list)
		return nil
	}
	return err
}

// WAFIPSetIPSetDescriptors represents AWS WAF IPSet IPSetDescriptors
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-waf-ipset-ipsetdescriptors.html
type WAFIPSetIPSetDescriptors struct {
	// The IP address type, such as IPV4. For valid values, see the Type
	// contents of the IPSetDescriptor data type in the AWS WAF API
	// Reference.
	Type *StringExpr `json:"Type,omitempty"`

	// An IP address (in CIDR notation) that AWS WAF permits, blocks, or
	// counts. For example, to specify a single IP address such as
	// 192.0.2.44, specify 192.0.2.44/32. To specify a range of IP addresses
	// such as 192.0.2.0 to 192.0.2.255, specify 192.0.2.0/24.
	Value *StringExpr `json:"Value,omitempty"`
}

// WAFIPSetIPSetDescriptorsList represents a list of WAFIPSetIPSetDescriptors
type WAFIPSetIPSetDescriptorsList []WAFIPSetIPSetDescriptors

// UnmarshalJSON sets the object from the provided JSON representation
func (l *WAFIPSetIPSetDescriptorsList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := WAFIPSetIPSetDescriptors{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = WAFIPSetIPSetDescriptorsList{item}
		return nil
	}
	list := []WAFIPSetIPSetDescriptors{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = WAFIPSetIPSetDescriptorsList(list)
		return nil
	}
	return err
}

// WAFRulePredicates represents AWS WAF Rule Predicates
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-waf-rule-predicates.html
type WAFRulePredicates struct {
	// The unique identifier of a predicate, such as the ID of a ByteMatchSet
	// or IPSet.
	DataId *StringExpr `json:"DataId,omitempty"`

	// Whether to use the settings or the negated settings that you specified
	// in the ByteMatchSet, IPSet, SizeConstraintSet, SqlInjectionMatchSet,
	// or XssMatchSet objects.
	Negated *BoolExpr `json:"Negated,omitempty"`

	// The type of predicate in a rule, such as an IPSet (IPMatch). For valid
	// values, see the Type contents of the Predicate data type in the AWS
	// WAF API Reference.
	Type *StringExpr `json:"Type,omitempty"`
}

// WAFRulePredicatesList represents a list of WAFRulePredicates
type WAFRulePredicatesList []WAFRulePredicates

// UnmarshalJSON sets the object from the provided JSON representation
func (l *WAFRulePredicatesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := WAFRulePredicates{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = WAFRulePredicatesList{item}
		return nil
	}
	list := []WAFRulePredicates{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = WAFRulePredicatesList(list)
		return nil
	}
	return err
}

// WAFSizeConstraintSetSizeConstraint represents AWS WAF SizeConstraintSet SizeConstraint
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-waf-sizeconstraintset-sizeconstraint.html
type WAFSizeConstraintSetSizeConstraint struct {
	// The type of comparison that you want AWS WAF to perform. AWS WAF uses
	// this value in combination with the Size and FieldToMatch property
	// values to check if the size constraint is a match. For more
	// information and valid values, see the ComparisonOperator content for
	// the SizeConstraint data type in the AWS WAF API Reference.
	ComparisonOperator *StringExpr `json:"ComparisonOperator,omitempty"`

	// The part of a web request that you want AWS WAF to search, such as a
	// specific header or a query string.
	FieldToMatch *WAFSizeConstraintSetSizeConstraintFieldToMatch `json:"FieldToMatch,omitempty"`

	// The size in bytes that you want AWS WAF to compare against the size of
	// the specified FieldToMatch. AWS WAF uses Size in combination with the
	// ComparisonOperator and FieldToMatch property values to check if the
	// size constraint of a web request is a match. For more information and
	// valid values, see the Size content for the SizeConstraint data type in
	// the AWS WAF API Reference.
	Size *IntegerExpr `json:"Size,omitempty"`

	// Specifies how AWS WAF processes the FieldToMatch property before
	// inspecting a request for a match. Text transformations eliminate some
	// of the unusual formatting that attackers use in web requests in an
	// effort to bypass AWS WAF. If you specify a transformation, AWS WAF
	// transforms the FieldToMatch before inspecting a web request for a
	// match.
	TextTransformation *StringExpr `json:"TextTransformation,omitempty"`
}

// WAFSizeConstraintSetSizeConstraintList represents a list of WAFSizeConstraintSetSizeConstraint
type WAFSizeConstraintSetSizeConstraintList []WAFSizeConstraintSetSizeConstraint

// UnmarshalJSON sets the object from the provided JSON representation
func (l *WAFSizeConstraintSetSizeConstraintList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := WAFSizeConstraintSetSizeConstraint{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = WAFSizeConstraintSetSizeConstraintList{item}
		return nil
	}
	list := []WAFSizeConstraintSetSizeConstraint{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = WAFSizeConstraintSetSizeConstraintList(list)
		return nil
	}
	return err
}

// WAFSizeConstraintSetSizeConstraintFieldToMatch represents AWS WAF SizeConstraintSet SizeConstraint FieldToMatch
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-waf-sizeconstraintset-sizeconstraint-fieldtomatch.html
type WAFSizeConstraintSetSizeConstraintFieldToMatch struct {
	// If you specify HEADER for the Type property, the name of the header
	// that AWS WAF searches for, such as User-Agent or Referer. If you
	// specify any other value for the Type property, do not specify this
	// property.
	Data *StringExpr `json:"Data,omitempty"`

	// The part of the web request in which AWS WAF searches for the target
	// string. For valid values, see FieldToMatch in the AWS WAF API
	// Reference.
	Type *StringExpr `json:"Type,omitempty"`
}

// WAFSizeConstraintSetSizeConstraintFieldToMatchList represents a list of WAFSizeConstraintSetSizeConstraintFieldToMatch
type WAFSizeConstraintSetSizeConstraintFieldToMatchList []WAFSizeConstraintSetSizeConstraintFieldToMatch

// UnmarshalJSON sets the object from the provided JSON representation
func (l *WAFSizeConstraintSetSizeConstraintFieldToMatchList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := WAFSizeConstraintSetSizeConstraintFieldToMatch{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = WAFSizeConstraintSetSizeConstraintFieldToMatchList{item}
		return nil
	}
	list := []WAFSizeConstraintSetSizeConstraintFieldToMatch{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = WAFSizeConstraintSetSizeConstraintFieldToMatchList(list)
		return nil
	}
	return err
}

// WAFSqlInjectionMatchSetSqlInjectionMatchTuples represents AWS WAF SqlInjectionMatchSet SqlInjectionMatchTuples
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-waf-sqlinjectionmatchset-sqlinjectionmatchtuples.html
type WAFSqlInjectionMatchSetSqlInjectionMatchTuples struct {
	// The part of a web request that you want AWS WAF to search, such as a
	// specific header or a query string.
	FieldToMatch *WAFByteMatchSetByteMatchTuplesFieldToMatch `json:"FieldToMatch,omitempty"`

	// Text transformations eliminate some of the unusual formatting that
	// attackers use in web requests in an effort to bypass AWS WAF. If you
	// specify a transformation, AWS WAF transforms the target string value
	// before inspecting a web request for a match. For valid values, see the
	// TextTransformation content for the SqlInjectionMatchTuple data type in
	// the AWS WAF API Reference.
	TextTransformation *StringExpr `json:"TextTransformation,omitempty"`
}

// WAFSqlInjectionMatchSetSqlInjectionMatchTuplesList represents a list of WAFSqlInjectionMatchSetSqlInjectionMatchTuples
type WAFSqlInjectionMatchSetSqlInjectionMatchTuplesList []WAFSqlInjectionMatchSetSqlInjectionMatchTuples

// UnmarshalJSON sets the object from the provided JSON representation
func (l *WAFSqlInjectionMatchSetSqlInjectionMatchTuplesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := WAFSqlInjectionMatchSetSqlInjectionMatchTuples{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = WAFSqlInjectionMatchSetSqlInjectionMatchTuplesList{item}
		return nil
	}
	list := []WAFSqlInjectionMatchSetSqlInjectionMatchTuples{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = WAFSqlInjectionMatchSetSqlInjectionMatchTuplesList(list)
		return nil
	}
	return err
}

// WAFSqlInjectionMatchSetSqlInjectionMatchTuplesFieldToMatch represents AWS WAF SqlInjectionMatchSet SqlInjectionMatchTuples FieldToMatch
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-waf-sqlinjectionmatchset-sqlinjectionmatchtuples-fieldtomatch.html
type WAFSqlInjectionMatchSetSqlInjectionMatchTuplesFieldToMatch struct {
	// If you specify HEADER for the Type property, the name of the header
	// that AWS WAF searches for, such as User-Agent or Referer. If you
	// specify any other value for the Type property, do not specify this
	// property.
	Data *StringExpr `json:"Data,omitempty"`

	// The part of the web request in which AWS WAF searches for the target
	// string. For valid values, see FieldToMatch in the AWS WAF API
	// Reference.
	Type *StringExpr `json:"Type,omitempty"`
}

// WAFSqlInjectionMatchSetSqlInjectionMatchTuplesFieldToMatchList represents a list of WAFSqlInjectionMatchSetSqlInjectionMatchTuplesFieldToMatch
type WAFSqlInjectionMatchSetSqlInjectionMatchTuplesFieldToMatchList []WAFSqlInjectionMatchSetSqlInjectionMatchTuplesFieldToMatch

// UnmarshalJSON sets the object from the provided JSON representation
func (l *WAFSqlInjectionMatchSetSqlInjectionMatchTuplesFieldToMatchList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := WAFSqlInjectionMatchSetSqlInjectionMatchTuplesFieldToMatch{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = WAFSqlInjectionMatchSetSqlInjectionMatchTuplesFieldToMatchList{item}
		return nil
	}
	list := []WAFSqlInjectionMatchSetSqlInjectionMatchTuplesFieldToMatch{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = WAFSqlInjectionMatchSetSqlInjectionMatchTuplesFieldToMatchList(list)
		return nil
	}
	return err
}

// WAFXssMatchSetXssMatchTuple represents AWS WAF XssMatchSet XssMatchTuple
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-waf-xssmatchset-xssmatchtuple.html
type WAFXssMatchSetXssMatchTuple struct {
	// The part of a web request that you want AWS WAF to search, such as a
	// specific header or a query string.
	FieldToMatch *WAFXssMatchSetXssMatchTupleFieldToMatch `json:"FieldToMatch,omitempty"`

	// Specifies how AWS WAF processes the FieldToMatch property before
	// inspecting a request for a match. Text transformations eliminate some
	// of the unusual formatting that attackers use in web requests in an
	// effort to bypass AWS WAF. If you specify a transformation, AWS WAF
	// transforms theFieldToMatch parameter before inspecting a web request
	// for a match.
	TextTransformation *StringExpr `json:"TextTransformation,omitempty"`
}

// WAFXssMatchSetXssMatchTupleList represents a list of WAFXssMatchSetXssMatchTuple
type WAFXssMatchSetXssMatchTupleList []WAFXssMatchSetXssMatchTuple

// UnmarshalJSON sets the object from the provided JSON representation
func (l *WAFXssMatchSetXssMatchTupleList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := WAFXssMatchSetXssMatchTuple{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = WAFXssMatchSetXssMatchTupleList{item}
		return nil
	}
	list := []WAFXssMatchSetXssMatchTuple{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = WAFXssMatchSetXssMatchTupleList(list)
		return nil
	}
	return err
}

// WAFXssMatchSetXssMatchTupleFieldToMatch represents AWS WAF XssMatchSet XssMatchTuple FieldToMatch
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-waf-xssmatchset-xssmatchtuple-fieldtomatch.html
type WAFXssMatchSetXssMatchTupleFieldToMatch struct {
	// If you specify HEADER for the Type property, the name of the header
	// that AWS WAF searches for, such as User-Agent or Referer. If you
	// specify any other value for the Type property, do not specify this
	// property.
	Data *StringExpr `json:"Data,omitempty"`

	// The part of the web request in which AWS WAF searches for the target
	// string. For valid values, see FieldToMatch in the AWS WAF API
	// Reference.
	Type *StringExpr `json:"Type,omitempty"`
}

// WAFXssMatchSetXssMatchTupleFieldToMatchList represents a list of WAFXssMatchSetXssMatchTupleFieldToMatch
type WAFXssMatchSetXssMatchTupleFieldToMatchList []WAFXssMatchSetXssMatchTupleFieldToMatch

// UnmarshalJSON sets the object from the provided JSON representation
func (l *WAFXssMatchSetXssMatchTupleFieldToMatchList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := WAFXssMatchSetXssMatchTupleFieldToMatch{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = WAFXssMatchSetXssMatchTupleFieldToMatchList{item}
		return nil
	}
	list := []WAFXssMatchSetXssMatchTupleFieldToMatch{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = WAFXssMatchSetXssMatchTupleFieldToMatchList(list)
		return nil
	}
	return err
}

// WAFWebACLAction represents AWS WAF WebACL Action
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-waf-webacl-action.html
type WAFWebACLAction struct {
	// For actions that are associated with a rule, the action that AWS WAF
	// takes when a web request matches all conditions in a rule.
	Type *StringExpr `json:"Type,omitempty"`
}

// WAFWebACLActionList represents a list of WAFWebACLAction
type WAFWebACLActionList []WAFWebACLAction

// UnmarshalJSON sets the object from the provided JSON representation
func (l *WAFWebACLActionList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := WAFWebACLAction{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = WAFWebACLActionList{item}
		return nil
	}
	list := []WAFWebACLAction{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = WAFWebACLActionList(list)
		return nil
	}
	return err
}

// WAFWebACLRules represents AWS WAF WebACL Rules
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-waf-webacl-rules.html
type WAFWebACLRules struct {
	// The action that Amazon CloudFront (CloudFront) or AWS WAF takes when a
	// web request matches all conditions in the rule, such as allow, block,
	// or count the request.
	Action *WAFWebACLAction `json:"Action,omitempty"`

	// The order in which AWS WAF evaluates the rules in a web ACL. AWS WAF
	// evaluates rules with a lower value before rules with a higher value.
	// The value must be a unique integer. If you have multiple rules in a
	// web ACL, the priority numbers do not need to be consecutive.
	Priority *IntegerExpr `json:"Priority,omitempty"`

	// The ID of an AWS WAF rule to associate with a web ACL.
	RuleId *StringExpr `json:"RuleId,omitempty"`
}

// WAFWebACLRulesList represents a list of WAFWebACLRules
type WAFWebACLRulesList []WAFWebACLRules

// UnmarshalJSON sets the object from the provided JSON representation
func (l *WAFWebACLRulesList) UnmarshalJSON(buf []byte) error {
	// Cloudformation allows a single object when a list of objects is expected
	item := WAFWebACLRules{}
	if err := json.Unmarshal(buf, &item); err == nil {
		*l = WAFWebACLRulesList{item}
		return nil
	}
	list := []WAFWebACLRules{}
	err := json.Unmarshal(buf, &list)
	if err == nil {
		*l = WAFWebACLRulesList(list)
		return nil
	}
	return err
}

// NewResourceByType returns a new resource object correspoding with the provided type
func NewResourceByType(typeName string) ResourceProperties {
	switch typeName {
	case "AWS::ApiGateway::Account":
		return &ApiGatewayAccount{}
	case "AWS::ApiGateway::ApiKey":
		return &ApiGatewayApiKey{}
	case "AWS::ApiGateway::Authorizer":
		return &ApiGatewayAuthorizer{}
	case "AWS::ApiGateway::BasePathMapping":
		return &ApiGatewayBasePathMapping{}
	case "AWS::ApiGateway::ClientCertificate":
		return &ApiGatewayClientCertificate{}
	case "AWS::ApiGateway::Deployment":
		return &ApiGatewayDeployment{}
	case "AWS::ApiGateway::Method":
		return &ApiGatewayMethod{}
	case "AWS::ApiGateway::Model":
		return &ApiGatewayModel{}
	case "AWS::ApiGateway::Resource":
		return &ApiGatewayResource{}
	case "AWS::ApiGateway::RestApi":
		return &ApiGatewayRestApi{}
	case "AWS::ApiGateway::Stage":
		return &ApiGatewayStage{}
	case "AWS::ApiGateway::UsagePlan":
		return &ApiGatewayUsagePlan{}
	case "AWS::ApplicationAutoScaling::ScalableTarget":
		return &ApplicationAutoScalingScalableTarget{}
	case "AWS::ApplicationAutoScaling::ScalingPolicy":
		return &ApplicationAutoScalingScalingPolicy{}
	case "AWS::AutoScaling::AutoScalingGroup":
		return &AutoScalingAutoScalingGroup{}
	case "AWS::AutoScaling::LaunchConfiguration":
		return &AutoScalingLaunchConfiguration{}
	case "AWS::AutoScaling::LifecycleHook":
		return &AutoScalingLifecycleHook{}
	case "AWS::AutoScaling::ScalingPolicy":
		return &AutoScalingScalingPolicy{}
	case "AWS::AutoScaling::ScheduledAction":
		return &AutoScalingScheduledAction{}
	case "AWS::CertificateManager::Certificate":
		return &CertificateManagerCertificate{}
	case "AWS::CloudFormation::Authentication":
		return &CloudFormationAuthentication{}
	case "AWS::CloudFormation::CustomResource":
		return &CloudFormationCustomResource{}
	case "AWS::CloudFormation::Init":
		return &CloudFormationInit{}
	case "AWS::CloudFormation::Interface":
		return &CloudFormationInterface{}
	case "AWS::CloudFormation::Stack":
		return &CloudFormationStack{}
	case "AWS::CloudFormation::WaitCondition":
		return &CloudFormationWaitCondition{}
	case "AWS::CloudFormation::WaitConditionHandle":
		return &CloudFormationWaitConditionHandle{}
	case "AWS::CloudFront::Distribution":
		return &CloudFrontDistribution{}
	case "AWS::CloudTrail::Trail":
		return &CloudTrailTrail{}
	case "AWS::CloudWatch::Alarm":
		return &CloudWatchAlarm{}
	case "AWS::CodeCommit::Repository":
		return &CodeCommitRepository{}
	case "AWS::CodeDeploy::Application":
		return &CodeDeployApplication{}
	case "AWS::CodeDeploy::DeploymentConfig":
		return &CodeDeployDeploymentConfig{}
	case "AWS::CodeDeploy::DeploymentGroup":
		return &CodeDeployDeploymentGroup{}
	case "AWS::CodePipeline::CustomActionType":
		return &CodePipelineCustomAction{}
	case "AWS::CodePipeline::Pipeline":
		return &CodePipelinePipeline{}
	case "AWS::Config::ConfigRule":
		return &ConfigConfigRule{}
	case "AWS::Config::ConfigurationRecorder":
		return &ConfigConfigurationRecorder{}
	case "AWS::Config::DeliveryChannel":
		return &ConfigDeliveryChannel{}
	case "AWS::DataPipeline::Pipeline":
		return &DataPipelinePipeline{}
	case "AWS::DirectoryService::MicrosoftAD":
		return &DirectoryServiceMicrosoftAD{}
	case "AWS::DirectoryService::SimpleAD":
		return &DirectoryServiceSimpleAD{}
	case "AWS::DynamoDB::Table":
		return &DynamoDBTable{}
	case "AWS::EC2::CustomerGateway":
		return &EC2CustomerGateway{}
	case "AWS::EC2::DHCPOptions":
		return &EC2DHCPOptions{}
	case "AWS::EC2::EIP":
		return &EC2EIP{}
	case "AWS::EC2::EIPAssociation":
		return &EC2EIPAssociation{}
	case "AWS::EC2::FlowLog":
		return &EC2FlowLog{}
	case "AWS::EC2::Host":
		return &EC2Host{}
	case "AWS::EC2::Instance":
		return &EC2Instance{}
	case "AWS::EC2::InternetGateway":
		return &EC2InternetGateway{}
	case "AWS::EC2::NatGateway":
		return &EC2NatGateway{}
	case "AWS::EC2::NetworkAcl":
		return &EC2NetworkAcl{}
	case "AWS::EC2::NetworkAclEntry":
		return &EC2NetworkAclEntry{}
	case "AWS::EC2::NetworkInterface":
		return &EC2NetworkInterface{}
	case "AWS::EC2::NetworkInterfaceAttachment":
		return &EC2NetworkInterfaceAttachment{}
	case "AWS::EC2::PlacementGroup":
		return &EC2PlacementGroup{}
	case "AWS::EC2::Route":
		return &EC2Route{}
	case "AWS::EC2::RouteTable":
		return &EC2RouteTable{}
	case "AWS::EC2::SecurityGroup":
		return &EC2SecurityGroup{}
	case "AWS::EC2::SecurityGroupEgress":
		return &EC2SecurityGroupEgress{}
	case "AWS::EC2::SecurityGroupIngress":
		return &EC2SecurityGroupIngress{}
	case "AWS::EC2::SpotFleet":
		return &EC2SpotFleet{}
	case "AWS::EC2::Subnet":
		return &EC2Subnet{}
	case "AWS::EC2::SubnetNetworkAclAssociation":
		return &EC2SubnetNetworkAclAssociation{}
	case "AWS::EC2::SubnetRouteTableAssociation":
		return &EC2SubnetRouteTableAssociation{}
	case "AWS::EC2::Volume":
		return &EC2Volume{}
	case "AWS::EC2::VolumeAttachment":
		return &EC2VolumeAttachment{}
	case "AWS::EC2::VPC":
		return &EC2VPC{}
	case "AWS::EC2::VPCDHCPOptionsAssociation":
		return &EC2VPCDHCPOptionsAssociation{}
	case "AWS::EC2::VPCEndpoint":
		return &EC2VPCEndpoint{}
	case "AWS::EC2::VPCGatewayAttachment":
		return &EC2VPCGatewayAttachment{}
	case "AWS::EC2::VPCPeeringConnection":
		return &EC2VPCPeeringConnection{}
	case "AWS::EC2::VPNConnection":
		return &EC2VPNConnection{}
	case "AWS::EC2::VPNConnectionRoute":
		return &EC2VPNConnectionRoute{}
	case "AWS::EC2::VPNGateway":
		return &EC2VPNGateway{}
	case "AWS::EC2::VPNGatewayRoutePropagation":
		return &EC2VPNGatewayRoutePropagation{}
	case "AWS::ECR::Repository":
		return &ECRRepository{}
	case "AWS::ECS::Cluster":
		return &ECSCluster{}
	case "AWS::ECS::Service":
		return &ECSService{}
	case "AWS::ECS::TaskDefinition":
		return &ECSTaskDefinition{}
	case "AWS::EFS::FileSystem":
		return &EFSFileSystem{}
	case "AWS::EFS::MountTarget":
		return &EFSMountTarget{}
	case "AWS::ElastiCache::CacheCluster":
		return &ElastiCacheCacheCluster{}
	case "AWS::ElastiCache::ParameterGroup":
		return &ElastiCacheParameterGroup{}
	case "AWS::ElastiCache::ReplicationGroup":
		return &ElastiCacheReplicationGroup{}
	case "AWS::ElastiCache::SecurityGroup":
		return &ElastiCacheSecurityGroup{}
	case "AWS::ElastiCache::SecurityGroupIngress":
		return &ElastiCacheSecurityGroupIngress{}
	case "AWS::ElastiCache::SubnetGroup ":
		return &ElastiCacheSubnetGroup{}
	case "AWS::ElasticBeanstalk::Application":
		return &ElasticBeanstalkApplication{}
	case "AWS::ElasticBeanstalk::ApplicationVersion":
		return &ElasticBeanstalkApplicationVersion{}
	case "AWS::ElasticBeanstalk::ConfigurationTemplate":
		return &ElasticBeanstalkConfigurationTemplate{}
	case "AWS::ElasticBeanstalk::Environment":
		return &ElasticBeanstalkEnvironment{}
	case "AWS::ElasticLoadBalancing::LoadBalancer":
		return &ElasticLoadBalancingLoadBalancer{}
	case "AWS::ElasticLoadBalancingV2::Listener":
		return &ElasticLoadBalancingV2Listener{}
	case "AWS::ElasticLoadBalancingV2::ListenerRule":
		return &ElasticLoadBalancingV2ListenerRule{}
	case "AWS::ElasticLoadBalancingV2::LoadBalancer":
		return &ElasticLoadBalancingV2LoadBalancer{}
	case "AWS::ElasticLoadBalancingV2::TargetGroup":
		return &ElasticLoadBalancingV2TargetGroup{}
	case "AWS::Elasticsearch::Domain":
		return &ElasticsearchDomain{}
	case "AWS::EMR::Cluster":
		return &EMRCluster{}
	case "AWS::EMR::InstanceGroupConfig":
		return &EMRInstanceGroupConfig{}
	case "AWS::EMR::Step":
		return &EMRStep{}
	case "AWS::Events::Rule":
		return &EventsRule{}
	case "AWS::GameLift::Alias":
		return &GameLiftAlias{}
	case "AWS::GameLift::Build":
		return &GameLiftBuild{}
	case "AWS::GameLift::Fleet":
		return &GameLiftFleet{}
	case "AWS::IAM::AccessKey":
		return &IAMAccessKey{}
	case "AWS::IAM::Group":
		return &IAMGroup{}
	case "AWS::IAM::InstanceProfile":
		return &IAMInstanceProfile{}
	case "AWS::IAM::ManagedPolicy":
		return &IAMManagedPolicy{}
	case "AWS::IAM::Policy":
		return &IAMPolicy{}
	case "AWS::IAM::Role":
		return &IAMRole{}
	case "AWS::IAM::User":
		return &IAMUser{}
	case "AWS::IAM::UserToGroupAddition":
		return &IAMUserToGroupAddition{}
	case "AWS::IoT::Certificate":
		return &IoTCertificate{}
	case "AWS::IoT::Policy":
		return &IoTPolicy{}
	case "AWS::IoT::PolicyPrincipalAttachment":
		return &IoTPolicyPrincipalAttachment{}
	case "AWS::IoT::Thing":
		return &IoTThing{}
	case "AWS::IoT::ThingPrincipalAttachment":
		return &IoTThingPrincipalAttachment{}
	case "AWS::IoT::TopicRule":
		return &IoTTopicRule{}
	case "AWS::Kinesis::Stream":
		return &KinesisStream{}
	case "AWS::KinesisFirehose::DeliveryStream":
		return &KinesisFirehoseDeliveryStream{}
	case "AWS::KMS::Alias":
		return &KMSAlias{}
	case "AWS::KMS::Key":
		return &KMSKey{}
	case "AWS::Lambda::EventSourceMapping":
		return &LambdaEventSourceMapping{}
	case "AWS::Lambda::Alias":
		return &LambdaAlias{}
	case "AWS::Lambda::Function":
		return &LambdaFunction{}
	case "AWS::Lambda::Permission":
		return &LambdaPermission{}
	case "AWS::Lambda::Version":
		return &LambdaVersion{}
	case "AWS::Logs::Destination":
		return &LogsDestination{}
	case "AWS::Logs::LogGroup":
		return &LogsLogGroup{}
	case "AWS::Logs::LogStream":
		return &LogsLogStream{}
	case "AWS::Logs::MetricFilter":
		return &LogsMetricFilter{}
	case "AWS::Logs::SubscriptionFilter":
		return &LogsSubscriptionFilter{}
	case "AWS::OpsWorks::App":
		return &OpsWorksApp{}
	case "AWS::OpsWorks::ElasticLoadBalancerAttachment":
		return &OpsWorksElasticLoadBalancerAttachment{}
	case "AWS::OpsWorks::Instance":
		return &OpsWorksInstance{}
	case "AWS::OpsWorks::Layer":
		return &OpsWorksLayer{}
	case "AWS::OpsWorks::Stack":
		return &OpsWorksStack{}
	case "AWS::OpsWorks::UserProfile":
		return &OpsWorksUserProfile{}
	case "AWS::OpsWorks::Volume":
		return &OpsWorksVolume{}
	case "AWS::RDS::DBCluster":
		return &RDSDBCluster{}
	case "AWS::RDS::DBClusterParameterGroup":
		return &RDSDBClusterParameterGroup{}
	case "AWS::RDS::DBInstance":
		return &RDSDBInstance{}
	case "AWS::RDS::DBParameterGroup":
		return &RDSDBParameterGroup{}
	case "AWS::RDS::DBSecurityGroup":
		return &RDSDBSecurityGroup{}
	case "AWS::RDS::DBSecurityGroupIngress":
		return &RDSDBSecurityGroupIngress{}
	case "AWS::RDS::DBSubnetGroup":
		return &RDSDBSubnetGroup{}
	case "AWS::RDS::EventSubscription":
		return &RDSEventSubscription{}
	case "AWS::RDS::OptionGroup":
		return &RDSOptionGroup{}
	case "AWS::Redshift::Cluster":
		return &RedshiftCluster{}
	case "AWS::Redshift::ClusterParameterGroup":
		return &RedshiftClusterParameterGroup{}
	case "AWS::Redshift::ClusterSecurityGroup":
		return &RedshiftClusterSecurityGroup{}
	case "AWS::Redshift::ClusterSecurityGroupIngress":
		return &RedshiftClusterSecurityGroupIngress{}
	case "AWS::Redshift::ClusterSubnetGroup":
		return &RedshiftClusterSubnetGroup{}
	case "AWS::Route53::HealthCheck":
		return &Route53HealthCheck{}
	case "AWS::Route53::HostedZone":
		return &Route53HostedZone{}
	case "AWS::Route53::RecordSet":
		return &Route53RecordSet{}
	case "AWS::Route53::RecordSetGroup":
		return &Route53RecordSetGroup{}
	case "AWS::S3::Bucket":
		return &S3Bucket{}
	case "AWS::S3::BucketPolicy":
		return &S3BucketPolicy{}
	case "AWS::SDB::Domain":
		return &SDBDomain{}
	case "AWS::SNS::Subscription":
		return &SNSSubscription{}
	case "AWS::SNS::Topic":
		return &SNSTopic{}
	case "AWS::SNS::TopicPolicy":
		return &SNSTopicPolicy{}
	case "AWS::SQS::Queue":
		return &SQSQueue{}
	case "AWS::SQS::QueuePolicy":
		return &SQSQueuePolicy{}
	case "AWS::SSM::Document":
		return &SSMDocument{}
	case "AWS::WAF::ByteMatchSet":
		return &WAFByteMatchSet{}
	case "AWS::WAF::IPSet":
		return &WAFIPSet{}
	case "AWS::WAF::Rule":
		return &WAFRule{}
	case "AWS::WAF::SizeConstraintSet":
		return &WAFSizeConstraintSet{}
	case "AWS::WAF::SqlInjectionMatchSet":
		return &WAFSqlInjectionMatchSet{}
	case "AWS::WAF::WebACL":
		return &WAFWebACL{}
	case "AWS::WAF::XssMatchSet":
		return &WAFXssMatchSet{}
	case "AWS::WorkSpaces::Workspace":
		return &WorkSpacesWorkspace{}

	default:
		for _, eachProvider := range customResourceProviders {
			customType := eachProvider(typeName)
			if nil != customType {
				return customType
			}
		}
	}
	return nil
}
