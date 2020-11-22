package sparta

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	gocf "github.com/mweagle/go-cloudformation"
	"github.com/rs/zerolog"
)

// Utility function to marshal an interface
func marshalInterface(item interface{}) interface{} {
	if item != nil {
		return item
	}
	return nil
}

// Utility function to marshal an int
func marshalInt(intVal int64) *gocf.IntegerExpr {
	if intVal != 0 {
		return gocf.Integer(intVal)
	}
	return nil
}

// Utility function to marshal a string
func marshalString(stringVal string) *gocf.StringExpr {
	if stringVal != "" {
		return gocf.String(stringVal)
	}
	return nil
}

func marshalStringExpr(stringExpr gocf.Stringable) *gocf.StringExpr {
	if stringExpr != nil {
		return stringExpr.String()
	}
	return nil
}

// Utility function to marshal a string lsit
func marshalStringList(stringVals []string) *gocf.StringListExpr {
	if len(stringVals) != 0 {
		stringableList := make([]gocf.Stringable, len(stringVals))
		for eachIndex, eachStringVal := range stringVals {
			stringableList[eachIndex] = gocf.String(eachStringVal)
		}
		return gocf.StringList(stringableList...)
	}
	return nil
}

// Utility function to marshal a boolean
func marshalBool(boolValue bool) *gocf.BoolExpr {
	if !boolValue {
		return gocf.Bool(boolValue)
	}
	return nil
}

// resourceOutputs is responsible for returning the conditional
// set of CloudFormation outputs for a given resource type.
func resourceOutputs(resourceName string,
	resource gocf.ResourceProperties,
	logger *zerolog.Logger) ([]string, error) {

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
		logger.Warn().
			Str("ResourceType", fmt.Sprintf("%T", typedResource)).
			Msg("Discovery information for dependency not yet implemented")
	}
	return outputProps, nil
}

func newCloudFormationResource(resourceType string, logger *zerolog.Logger) (gocf.ResourceProperties, error) {
	resProps := gocf.NewResourceByType(resourceType)
	if nil == resProps {

		logger.Fatal().
			Str("Type", resourceType).
			Msg("Failed to create CloudFormation CustomResource!")

		return nil, fmt.Errorf("unsupported CustomResourceType: %s", resourceType)
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
	logger *zerolog.Logger) ([]byte, error) {

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
	quotedAttrs := make([]string, len(resourceOutputs))
	for eachIndex, eachOutput := range resourceOutputs {
		quotedAttrs[eachIndex] = fmt.Sprintf(`"%s" :"{ "Fn::GetAtt" : [ "%s", "%s" ] }"`,
			eachOutput,
			logicalResourceName,
			eachOutput)
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
