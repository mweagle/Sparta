package decorator

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	sparta "github.com/mweagle/Sparta"
	spartaIAM "github.com/mweagle/Sparta/aws/iam/builder"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// NewCloudMapServiceDecorator returns an instance of CloudMapServiceDecorator
// which can be used to publish information into CloudMap
func NewCloudMapServiceDecorator(namespaceID gocf.Stringable,
	serviceName gocf.Stringable) (*CloudMapServiceDecorator, error) {
	if namespaceID == nil ||
		serviceName == nil {
		return nil,
			errors.Errorf("Both namespaceID and serviceName must not be nil for CloudMapServiceDecorator")
	}
	decorator := &CloudMapServiceDecorator{
		namespaceID:       namespaceID,
		serviceName:       serviceName,
		servicePublishers: make(map[string]interface{}),
		nonce:             fmt.Sprintf("%d", time.Now().Unix()),
	}

	return decorator, nil
}

// CloudMapServiceDecorator is an instance of a service decorator that
// publishes CloudMap info
type CloudMapServiceDecorator struct {
	namespaceID gocf.Stringable
	serviceName gocf.Stringable
	Description gocf.Stringable
	nonce       string
	// This is a list of decorators that handle the publishing of
	// individual lambda and resources. That means we can't apply this until all the
	// individual decorators are complete...
	servicePublishers map[string]interface{}
}

// LogicalResourceName returns the CloudFormation Logical resource
// name that can be used to get information about the generated
// CloudFormation resource
func (cmsd *CloudMapServiceDecorator) LogicalResourceName() string {
	jsonData, jsonDataErr := cmsd.serviceName.String().MarshalJSON()
	if jsonDataErr != nil {
		jsonData = []byte(fmt.Sprintf("%#v", cmsd.serviceName))
	}
	return sparta.CloudFormationResourceName("CloudMap", string(jsonData))
}

// DecorateService satisfies the ServiceDecoratorHookHandler interface
func (cmsd *CloudMapServiceDecorator) DecorateService(context map[string]interface{},
	serviceName string,
	template *gocf.Template,
	S3Bucket string,
	S3Key string,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *logrus.Logger) error {

	// Create the service entry
	serviceDiscoveryResourceName := cmsd.LogicalResourceName()
	serviceDiscoveryResource := &gocf.ServiceDiscoveryService{
		NamespaceID: cmsd.namespaceID.String(),
		Name:        cmsd.serviceName.String(),
		HealthCheckCustomConfig: &gocf.ServiceDiscoveryServiceHealthCheckCustomConfig{
			FailureThreshold: gocf.Integer(1),
		},
	}
	if cmsd.Description != nil {
		serviceDiscoveryResource.Description = cmsd.Description.String()
	}
	template.AddResource(serviceDiscoveryResourceName, serviceDiscoveryResource)

	// Then for each template decorator, apply it
	for _, eachAttributeSet := range cmsd.servicePublishers {
		resourceName := sparta.CloudFormationResourceName("CloudMapRes", fmt.Sprintf("%v", eachAttributeSet))
		resource := &gocf.ServiceDiscoveryInstance{
			InstanceAttributes: eachAttributeSet,
			ServiceID:          gocf.Ref(serviceDiscoveryResourceName).String(),
			//InstanceID:         gocf.String(eachLookupName),
		}
		template.AddResource(resourceName, resource)
	}
	return nil
}

func (cmsd *CloudMapServiceDecorator) publish(lookupName string,
	resourceName string,
	resource gocf.ResourceProperties,
	userAttributes map[string]interface{}) error {

	_, exists := cmsd.servicePublishers[lookupName]
	if exists {
		return errors.Errorf("CloudMap discovery info for lookup name `%s` is already defined. Instance names must e unique", lookupName)
	}

	attributes := make(map[string]interface{})
	attributes["Ref"] = gocf.Ref(resourceName)
	attributes["Type"] = resource.CfnResourceType()
	for _, eachAttribute := range resource.CfnResourceAttributes() {
		attributes[eachAttribute] = gocf.GetAtt(resourceName, eachAttribute)
	}
	for eachKey, eachValue := range userAttributes {
		attributes[eachKey] = eachValue
	}
	attributes["Name"] = lookupName
	cmsd.servicePublishers[lookupName] = attributes
	return nil
}

// PublishResource publishes the known outputs and attributes for the
// given ResourceProperties instance
func (cmsd *CloudMapServiceDecorator) PublishResource(lookupName string,
	resourceName string,
	resource gocf.ResourceProperties,
	addditionalProperties map[string]interface{}) error {

	return cmsd.publish(lookupName,
		resourceName,
		resource,
		addditionalProperties)
}

//PublishLambda publishes the known outputs for the given sparta
//AWS Lambda function
func (cmsd *CloudMapServiceDecorator) PublishLambda(lookupName string,
	lambdaInfo *sparta.LambdaAWSInfo,
	additionalAttributes map[string]interface{}) error {
	lambdaEntry := &gocf.LambdaFunction{}

	return cmsd.publish(lookupName,
		lambdaInfo.LogicalResourceName(),
		lambdaEntry,
		additionalAttributes)
}

// EnableDiscoverySupport enables the IAM privs for the CloudMap ServiceID
// created by this stack as well as any additional serviceIDs
func (cmsd *CloudMapServiceDecorator) EnableDiscoverySupport(lambdaInfo *sparta.LambdaAWSInfo,
	additionalServiceIDs ...string) error {

	// Update the environment
	lambdaOptions := lambdaInfo.Options
	if lambdaOptions == nil {
		lambdaOptions = &sparta.LambdaFunctionOptions{}
	}
	lambdaEnvVars := lambdaOptions.Environment
	if lambdaEnvVars == nil {
		lambdaOptions.Environment = make(map[string]*gocf.StringExpr)
	}
	lambdaOptions.Environment[EnvVarCloudMapNamespaceID] = cmsd.namespaceID.String()
	lambdaOptions.Environment[EnvVarCloudMapServiceID] = gocf.Ref(cmsd.LogicalResourceName()).String()

	// Ref: https://docs.aws.amazon.com/IAM/latest/UserGuide/list_awscloudmap.html
	// arn:aws:servicediscovery:<region>:<account-id>:<resource-type>/<resource_name>.
	// arn:${Partition}:servicediscovery:${Region}:${Account}:service/${ServiceName}
	// DiscoverInstance, GetInstance, ListInstances, GetService
	privGlobal := spartaIAM.Allow(
		"servicediscovery:DiscoverInstances",
		"servicediscovery:GetNamespace",
		"servicediscovery:ListInstances").ForResource().
		Literal("*").
		ToPrivilege()
	lambdaInfo.RoleDefinition.Privileges = append(lambdaInfo.RoleDefinition.Privileges, privGlobal)

	privScoped := spartaIAM.Allow("servicediscovery:GetService").ForResource().
		Literal("arn:aws:servicediscovery:").
		Region(":").
		AccountID(":").
		Literal("service/").
		Ref(cmsd.LogicalResourceName(), "").
		ToPrivilege()

	lambdaInfo.RoleDefinition.Privileges = append(lambdaInfo.RoleDefinition.Privileges,
		privScoped)

	for _, eachServiceVal := range additionalServiceIDs {
		privScoped := spartaIAM.Allow("servicediscovery:GetService").ForResource().
			Literal("arn:aws:servicediscovery:").
			Region(":").
			AccountID(":").
			Literal("service/").
			Literal(eachServiceVal).
			ToPrivilege()
		lambdaInfo.RoleDefinition.Privileges = append(lambdaInfo.RoleDefinition.Privileges,
			privScoped)
	}
	return nil
}
