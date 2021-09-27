package sparta

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	gof "github.com/awslabs/goformation/v5/cloudformation"
	"github.com/jmespath/go-jmespath"
	cwCustomProvider "github.com/mweagle/Sparta/aws/cloudformation/provider"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var metadataInterface = reflect.TypeOf(map[string]interface{}{})
var dependsOnInterface = reflect.TypeOf([]string{})

// resourceOutputs is responsible for returning the conditional
// set of CloudFormation outputs for a given resource type. These are
// produced from the schema
func resourceOutputs(resourceName string,
	resource gof.Resource,
	logger *zerolog.Logger) ([]string, error) {

	// Get the schema
	schemaDef, schemaDefErr := embeddedString("resources/cloudformation-schema.json")
	if schemaDefErr != nil {
		return nil, schemaDefErr
	}

	var rawData interface{}
	unmarshalErr := json.Unmarshal([]byte(schemaDef), &rawData)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	// Issue the JMES query to find this resource in the schema...
	jmesQuery := fmt.Sprintf("ResourceTypes.\"%s\".Attributes", resource.AWSCloudFormationType())
	result, resultErr := jmespath.Search(jmesQuery, rawData)
	if resultErr != nil {
		return nil, resultErr
	}

	resultMap, resultMapOk := result.(map[string]interface{})
	if !resultMapOk {
		// If this a custom resource, there are no outputs...
		if resource.AWSCloudFormationType() == "AWS::CloudFormation::CustomResource" {
			return nil, nil
		}
		return nil, errors.Errorf("Failed to extract outputs for resource type: %s", resource.AWSCloudFormationType())
	}

	vals := []string{}
	for eachKey := range resultMap {
		vals = append(vals, eachKey)
	}
	return vals, nil
}

func newCloudFormationResource(resourceType string, logger *zerolog.Logger) (gof.Resource, error) {
	resProps, _ := cwCustomProvider.NewCloudFormationCustomResource(resourceType, logger)
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

func discoveryResourceInfoForDependency(cfTemplate *gof.Template,
	logicalResourceName string,
	logger *zerolog.Logger) ([]byte, error) {

	item, ok := cfTemplate.Resources[logicalResourceName]
	if !ok {
		return nil, nil
	}
	resourceOutputs, resourceOutputsErr := resourceOutputs(logicalResourceName,
		item,
		logger)
	if resourceOutputsErr != nil {
		return nil, resourceOutputsErr
	}
	// Template data
	templateData := &discoveryDataTemplate{
		ResourceID:   logicalResourceName,
		ResourceType: item.AWSCloudFormationType(),
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

func safeAppendDependency(resource gof.Resource, dependencyName string) error {

	val := reflect.ValueOf(resource).Elem()
	dependsOnField := val.FieldByName("AWSCloudFormationDependsOn")
	if dependsOnField.IsValid() && dependsOnField.CanConvert(dependsOnInterface) {
		dependsArray := dependsOnField.Interface().([]string)
		if dependsArray == nil {
			dependsArray = []string{}
		}
		dependsArray = append(dependsArray, dependencyName)
		reflectMapVal := reflect.ValueOf(dependsArray)
		dependsOnField.Set(reflectMapVal)
		return nil
	}
	return errors.Errorf("Failed to set Dependencies for resource: %v", resource)
}

func safeMetadataInsert(resource gof.Resource, key string, value interface{}) error {
	val := reflect.ValueOf(resource).Elem()
	metadataField := val.FieldByName("AWSCloudFormationMetadata")
	if metadataField.IsValid() && metadataField.CanConvert(metadataInterface) {
		metadataMap := metadataField.Interface().(map[string]interface{})
		if metadataMap == nil {
			metadataMap = make(map[string]interface{})
		}
		metadataMap[key] = value
		reflectMapVal := reflect.ValueOf(metadataMap)
		metadataField.Set(reflectMapVal)
		return nil
	}
	return errors.Errorf("Failed to set Metadata for resource: %v", resource)
}
