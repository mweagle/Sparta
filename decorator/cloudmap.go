package decorator

import (
	"encoding/json"
	"fmt"
	"time"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	gof "github.com/awslabs/goformation/v5/cloudformation"
	goflambda "github.com/awslabs/goformation/v5/cloudformation/lambda"
	gofservicediscovery "github.com/awslabs/goformation/v5/cloudformation/servicediscovery"
	sparta "github.com/mweagle/Sparta/v3"
	spartaCF "github.com/mweagle/Sparta/v3/aws/cloudformation"
	spartaIAM "github.com/mweagle/Sparta/v3/aws/iam/builder"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// NewCloudMapServiceDecorator returns an instance of CloudMapServiceDecorator
// which can be used to publish information into CloudMap
func NewCloudMapServiceDecorator(namespaceID string,
	serviceName string) (*CloudMapServiceDecorator, error) {
	if namespaceID == "" ||
		serviceName == "" {
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
	namespaceID string
	serviceName string
	Description string
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
	jsonData, jsonDataErr := json.Marshal(cmsd.serviceName)
	if jsonDataErr != nil {
		jsonData = []byte(fmt.Sprintf("%#v", cmsd.serviceName))
	}
	return sparta.CloudFormationResourceName("CloudMap", string(jsonData))
}

// DecorateService satisfies the ServiceDecoratorHookHandler interface
func (cmsd *CloudMapServiceDecorator) DecorateService(context map[string]interface{},
	serviceName string,
	template *gof.Template,
	S3Bucket string,
	S3Key string,
	buildID string,
	awsConfig awsv2.Config,
	noop bool,
	logger *zerolog.Logger) error {

	// Create the service entry
	serviceDiscoveryResourceName := cmsd.LogicalResourceName()
	serviceDiscoveryResource := &gofservicediscovery.Service{
		NamespaceId: cmsd.namespaceID,
		Name:        cmsd.serviceName,
		HealthCheckCustomConfig: &gofservicediscovery.Service_HealthCheckCustomConfig{
			FailureThreshold: 1,
		},
	}
	if cmsd.Description != "" {
		serviceDiscoveryResource.Description = cmsd.Description
	}
	template.Resources[serviceDiscoveryResourceName] = serviceDiscoveryResource

	// Then for each template decorator, apply it
	for _, eachAttributeSet := range cmsd.servicePublishers {
		resourceName := sparta.CloudFormationResourceName("CloudMapRes", fmt.Sprintf("%v", eachAttributeSet))
		resource := &gofservicediscovery.Instance{
			InstanceAttributes: eachAttributeSet,
			ServiceId:          gof.Ref(serviceDiscoveryResourceName),
			//InstanceID:         gof.String(eachLookupName),
		}
		template.Resources[resourceName] = resource
	}
	return nil
}

func (cmsd *CloudMapServiceDecorator) publish(lookupName string,
	resourceName string,
	resource gof.Resource,
	userAttributes map[string]interface{},
	logger *zerolog.Logger) error {

	_, exists := cmsd.servicePublishers[lookupName]
	if exists {
		return errors.Errorf("CloudMap discovery info for lookup name `%s` is already defined. Instance names must e unique", lookupName)
	}

	attributes := make(map[string]interface{})
	attributes["Ref"] = gof.Ref(resourceName)
	attributes["Type"] = resource.AWSCloudFormationType()
	outputs, outputsErr := spartaCF.ResourceOutputs(resourceName, resource, logger)
	if outputsErr == nil {
		for _, eachAttribute := range outputs {
			attributes[eachAttribute] = gof.GetAtt(resourceName, eachAttribute)
		}
	}

	for eachKey, eachValue := range userAttributes {
		attributes[eachKey] = eachValue
	}
	attributes["Name"] = lookupName
	cmsd.servicePublishers[lookupName] = attributes
	return nil
}

// PublishResource publishes the known outputs and attributes for the
// given Resource instance
func (cmsd *CloudMapServiceDecorator) PublishResource(lookupName string,
	resourceName string,
	resource gof.Resource,
	addditionalProperties map[string]interface{},
	logger *zerolog.Logger) error {

	return cmsd.publish(lookupName,
		resourceName,
		resource,
		addditionalProperties,
		logger)
}

//PublishLambda publishes the known outputs for the given sparta
//AWS Lambda function
func (cmsd *CloudMapServiceDecorator) PublishLambda(lookupName string,
	lambdaInfo *sparta.LambdaAWSInfo,
	additionalAttributes map[string]interface{},
	logger *zerolog.Logger) error {
	lambdaEntry := &goflambda.Function{}

	return cmsd.publish(lookupName,
		lambdaInfo.LogicalResourceName(),
		lambdaEntry,
		additionalAttributes,
		logger)
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
		lambdaOptions.Environment = make(map[string]string)
	}
	lambdaOptions.Environment[EnvVarCloudMapNamespaceID] = cmsd.namespaceID
	lambdaOptions.Environment[EnvVarCloudMapServiceID] = gof.Ref(cmsd.LogicalResourceName())

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
