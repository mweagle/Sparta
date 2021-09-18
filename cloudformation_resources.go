package sparta

import (
	"bytes"
	"fmt"
	jmesPath "github.com/jmespath/go-jmespath"
	"reflect"
	"strings"
	"text/template"

	gof "github.com/awslabs/goformation/v5/cloudformation"
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
	resource, resourceErr := _escFSString(false, "/resources/cloudformation-schema.json")
	if resourceErr != nil {
		return nil, resourceErr
	}

	var jsonData interface{}
	unmarshalErr := json.Unmarshal([]byte(resource), &jsonData)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	// Issue the JMES query to find this resource in the schema...
	jmesQuery = fmt.Sprintf("keys(Resources.\"%s\".Attributes)", resource.AWSCloudFormationType)
	result, resultErr = jmesPath.search(jmesQuery, jsonData)
	if resultErr != nil {
		return nil
	}
	typedArr, typedArrErr := result.([]string)
	return typedArr, typedArrErr
}

func newCloudFormationResource(resourceType string, logger *zerolog.Logger) (gof.Resource, error) {
	/*
		TODO - implmement
		esProps := gocf.NewResourceByType(resourceType)
		if nil == resProps {

			logger.Fatal().
				Str("Type", resourceType).
				Msg("Failed to create CloudFormation CustomResource!")

			return nil, fmt.Errorf("unsupported CustomResourceType: %s", resourceType)
		}
		return resProps, nil
	*/
	return nil, nil
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

func safeAppendDependency(resource gof.Resource, dependencyName string) {

	val := reflect.ValueOf(resource).Elem()
	dependsOnField := val.FieldByName("AWSCloudFormationDependsOn")
	if dependsOnField.IsValid() && dependsOnField.CanConvert(dependsOnInterface) {
		dependsArray := dependsOnField.Interface().([]string)
		if dependsArray == nil {
			dependsArray = []string{}
		}
		dependsArray = append(dependsArray, dependencyName)
	}
}

func safeMetadataInsert(resource gof.Resource, key string, value interface{}) {
	val := reflect.ValueOf(resource).Elem()
	metadataField := val.FieldByName("AWSCloudFormationMetadata")
	if metadataField.IsValid() && metadataField.CanConvert(metadataInterface) {
		metadataMap := metadataField.Interface().(map[string]interface{})
		if metadataMap == nil {
			metadataMap = make(map[string]interface{})
		}
		metadataMap[key] = value
	}
}
