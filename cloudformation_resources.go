package sparta

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	gocf "github.com/mweagle/go-cloudformation"
	"github.com/sirupsen/logrus"
)

// resourceOutputs is responsible for returning the conditional
// set of CloudFormation outputs for a given resource type.
func resourceOutputs(resourceName string,
	resource gocf.ResourceProperties,
	logger *logrus.Logger) ([]string, error) {

	outputProps := []string{}
	switch typedResource := resource.(type) {
	case gocf.IAMRole:
		// NOP
	case *gocf.DynamoDBTable:
		if typedResource.StreamSpecification != nil {
			outputProps = append(outputProps, "StreamArn")
		}
	case gocf.DynamoDBTable:
		if typedResource.StreamSpecification != nil {
			outputProps = append(outputProps, "StreamArn")
		}
	case gocf.KinesisStream,
		*gocf.KinesisStream:
		outputProps = append(outputProps, "Arn")
	case gocf.Route53RecordSet,
		*gocf.Route53RecordSet:
		// NOP
	case gocf.S3Bucket,
		*gocf.S3Bucket:
		outputProps = append(outputProps, "DomainName", "WebsiteURL")
	case gocf.SNSTopic,
		*gocf.SNSTopic:
		outputProps = append(outputProps, "TopicName")
	case gocf.SQSQueue,
		*gocf.SQSQueue:
		outputProps = append(outputProps, "Arn", "QueueName")
	default:
		logger.WithFields(logrus.Fields{
			"ResourceType": fmt.Sprintf("%T", typedResource),
		}).Warn("Discovery information for dependency not yet implemented")
	}
	return outputProps, nil
}

func newCloudFormationResource(resourceType string, logger *logrus.Logger) (gocf.ResourceProperties, error) {
	resProps := gocf.NewResourceByType(resourceType)
	if nil == resProps {
		logger.WithFields(logrus.Fields{
			"Type": resourceType,
		}).Fatal("Failed to create CloudFormation CustomResource!")
		return nil, fmt.Errorf("Unsupported CustomResourceType: %s", resourceType)
	}
	return resProps, nil
}

type discoveryDataTemplate struct {
	ResourceID         string
	ResourceType       string
	ResourceProperties string
}

var discoveryDataForResourceDependency = `
	{
		"ResourceID" : "<< .ResourceID >>",
		"ResourceRef" : "{"Ref":"<< .ResourceID >>"}",
		"ResourceType" : "<< .ResourceType >>",
		"Properties" : {
			<< .ResourceProperties >>
		}
	}
`

func discoveryResourceInfoForDependency(cfTemplate *gocf.Template,
	logicalResourceName string,
	logger *logrus.Logger) ([]byte, error) {

	item, ok := cfTemplate.Resources[logicalResourceName]
	if !ok {
		return nil, nil
	}
	resourceOutputs, resourceOutputsErr := resourceOutputs(logicalResourceName,
		item.Properties,
		logger)
	if resourceOutputsErr != nil {
		return nil, resourceOutputsErr
	}
	// Template data
	templateData := &discoveryDataTemplate{
		ResourceID:   logicalResourceName,
		ResourceType: item.Properties.CfnResourceType(),
	}
	var quotedAttrs []string
	for _, eachOutput := range resourceOutputs {
		quotedAttrs = append(quotedAttrs,
			fmt.Sprintf(`"%s" :"{ "Fn::GetAtt" : [ "%s", "%s" ] }"`,
				eachOutput,
				logicalResourceName,
				eachOutput))
	}
	templateData.ResourceProperties = strings.Join(quotedAttrs, ",")

	// Create the data that can be stuffed into Environment
	discoveryTemplate, discoveryTemplateErr := template.New("discoveryResourceData").
		Delims("<<", ">>").
		Parse(discoveryDataForResourceDependency)
	if nil != discoveryTemplateErr {
		return nil, discoveryTemplateErr
	}

	var templateResults bytes.Buffer
	evalResultErr := discoveryTemplate.Execute(&templateResults, templateData)
	return templateResults.Bytes(), evalResultErr
}
func safeAppendDependency(resource *gocf.Resource, dependencyName string) {
	if nil == resource.DependsOn {
		resource.DependsOn = []string{}
	}
	resource.DependsOn = append(resource.DependsOn, dependencyName)
}
func safeMetadataInsert(resource *gocf.Resource, key string, value interface{}) {
	if nil == resource.Metadata {
		resource.Metadata = make(map[string]interface{})
	}
	resource.Metadata[key] = value
}
