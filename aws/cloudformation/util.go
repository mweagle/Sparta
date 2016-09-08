package cloudformation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/lambda"
	gocf "github.com/crewjam/go-cloudformation"
	sparta "github.com/mweagle/Sparta"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

var cloudFormationStackTemplateMap map[string]*gocf.Template

func init() {
	cloudFormationStackTemplateMap = make(map[string]*gocf.Template, 0)
}

////////////////////////////////////////////////////////////////////////////////
// Private
////////////////////////////////////////////////////////////////////////////////
func toExpressionSlice(input interface{}) ([]string, error) {
	var expressions []string
	slice, sliceOK := input.([]interface{})
	if !sliceOK {
		return nil, fmt.Errorf("Failed to convert to slice")
	}
	for _, eachValue := range slice {
		switch str := eachValue.(type) {
		case string:
			expressions = append(expressions, str)
		}
	}
	return expressions, nil
}
func parseFnJoinExpr(data map[string]interface{}) (*gocf.StringExpr, error) {
	if len(data) <= 0 {
		return nil, fmt.Errorf("FnJoinExpr data is empty")
	}
	for eachKey, eachValue := range data {
		switch eachKey {
		case "Ref":
			return gocf.Ref(eachValue.(string)).String(), nil
		case "Fn::GetAtt":
			attrValues, attrValuesErr := toExpressionSlice(eachValue)
			if nil != attrValuesErr {
				return nil, attrValuesErr
			}
			if len(attrValues) != 2 {
				return nil, fmt.Errorf("Invalid params for Fn::GetAtt: %s", eachValue)
			}
			return gocf.GetAtt(attrValues[0], attrValues[1]).String(), nil
		case "Fn::FindInMap":
			attrValues, attrValuesErr := toExpressionSlice(eachValue)
			if nil != attrValuesErr {
				return nil, attrValuesErr
			}
			if len(attrValues) != 3 {
				return nil, fmt.Errorf("Invalid params for Fn::FindInMap: %s", eachValue)
			}
			return gocf.FindInMap(attrValues[0], gocf.String(attrValues[1]), gocf.String(attrValues[2])), nil
		}
	}
	return nil, fmt.Errorf("Unsupported AWS Function detected: %#v", data)
}

////////////////////////////////////////////////////////////////////////////////
// Public
////////////////////////////////////////////////////////////////////////////////

// S3AllKeysArnForBucket returns a CloudFormation-compatible Arn expression
// (string or Ref) for all bucket keys (`/*`).  The bucket
// parameter may be either a string or an interface{} ("Ref: "myResource")
// value
func S3AllKeysArnForBucket(bucket interface{}) *gocf.StringExpr {
	arnParts := []gocf.Stringable{gocf.String("arn:aws:s3:::")}

	switch bucket.(type) {
	case string:
		// Don't be smart if the Arn value is a user supplied literal
		arnParts = append(arnParts, gocf.String(bucket.(string)))
	case *gocf.StringExpr:
		arnParts = append(arnParts, bucket.(*gocf.StringExpr))
	case gocf.RefFunc:
		arnParts = append(arnParts, bucket.(gocf.RefFunc).String())
	default:
		panic(fmt.Sprintf("Unsupported SourceArn value type: %+v", bucket))
	}
	arnParts = append(arnParts, gocf.String("/*"))
	return gocf.Join("", arnParts...).String()
}

// S3ArnForBucket returns a CloudFormation-compatible Arn expression
// (string or Ref) suitable for template reference.  The bucket
// parameter may be either a string or an interface{} ("Ref: "myResource")
// value
func S3ArnForBucket(bucket interface{}) *gocf.StringExpr {
	arnParts := []gocf.Stringable{gocf.String("arn:aws:s3:::")}

	switch bucket.(type) {
	case string:
		// Don't be smart if the Arn value is a user supplied literal
		arnParts = append(arnParts, gocf.String(bucket.(string)))
	case *gocf.StringExpr:
		arnParts = append(arnParts, bucket.(*gocf.StringExpr))
	case gocf.RefFunc:
		arnParts = append(arnParts, bucket.(gocf.RefFunc).String())
	default:
		panic(fmt.Sprintf("Unsupported SourceArn value type: %+v", bucket))
	}
	return gocf.Join("", arnParts...).String()
}

// MapToResourceTags transforms a go map[string]string to a CloudFormation-compliant
// Tags representation.  See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-resource-tags.html
func MapToResourceTags(tagMap map[string]string) []interface{} {
	var tags []interface{}
	for eachKey, eachValue := range tagMap {
		tags = append(tags, map[string]interface{}{
			"Key":   eachKey,
			"Value": eachValue,
		})
	}
	return tags
}

// Struct to encapsulate transforming data into
type templateConverter struct {
	templateReader          io.Reader
	additionalTemplateProps map[string]interface{}
	// internals
	expandedTemplate string
	contents         []gocf.Stringable
	conversionError  error
}

func (converter *templateConverter) expandTemplate() *templateConverter {
	if nil != converter.conversionError {
		return converter
	}
	templateDataBytes, templateDataErr := ioutil.ReadAll(converter.templateReader)
	if nil != templateDataErr {
		converter.conversionError = templateDataErr
		return converter
	}
	templateData := string(templateDataBytes)

	parsedTemplate, templateErr := template.New("CloudFormation").Parse(templateData)
	if nil != templateErr {
		converter.conversionError = templateDataErr
		return converter
	}
	output := &bytes.Buffer{}
	executeErr := parsedTemplate.Execute(output, converter.additionalTemplateProps)
	if nil != executeErr {
		converter.conversionError = executeErr
		return converter
	}
	converter.expandedTemplate = output.String()
	return converter
}

func (converter *templateConverter) parseData() *templateConverter {
	if converter.conversionError != nil {
		return converter
	}
	reAWSProp := regexp.MustCompile("\\{\\s*\"([Ref|Fn\\:\\:\\w+])")
	splitData := strings.Split(converter.expandedTemplate, "\n")
	splitDataLineCount := len(splitData)

	for eachLineIndex, eachLine := range splitData {
		curContents := eachLine
		for len(curContents) != 0 {

			matchInfo := reAWSProp.FindStringSubmatchIndex(curContents)
			if nil != matchInfo {
				// If there's anything at the head, push it.
				if matchInfo[0] != 0 {
					head := curContents[0:matchInfo[0]]
					converter.contents = append(converter.contents, gocf.String(fmt.Sprintf("%s", head)))
					curContents = curContents[len(head):]
				}

				// There's at least one match...find the closing brace...
				var parsed map[string]interface{}
				for indexPos, eachChar := range curContents {
					if string(eachChar) == "}" {
						testBlock := curContents[0 : indexPos+1]
						err := json.Unmarshal([]byte(testBlock), &parsed)
						if err == nil {
							parsedContents, parsedContentsErr := parseFnJoinExpr(parsed)
							if nil != parsedContentsErr {
								converter.conversionError = parsedContentsErr
								return converter
							}
							converter.contents = append(converter.contents, parsedContents)
							curContents = curContents[indexPos+1:]
							if len(curContents) <= 0 && (eachLineIndex < (splitDataLineCount - 1)) {
								converter.contents = append(converter.contents, gocf.String("\n"))
							}
							break
						}
					}
				}
				if nil == parsed {
					// We never did find the end...
					converter.conversionError = fmt.Errorf("Invalid CloudFormation JSON expression on line: %s", eachLine)
					return converter
				}
			} else {
				// No match, just include it iff there is another line afterwards
				newlineValue := ""
				if eachLineIndex < (splitDataLineCount - 1) {
					newlineValue = "\n"
				}
				// Always include a newline at a minimum
				appendLine := fmt.Sprintf("%s%s", curContents, newlineValue)
				if len(appendLine) != 0 {
					converter.contents = append(converter.contents, gocf.String(appendLine))
				}
				break
			}
		}
	}
	return converter
}

func (converter *templateConverter) results() (*gocf.StringExpr, error) {
	if nil != converter.conversionError {
		return nil, converter.conversionError
	}
	return gocf.Join("", converter.contents...), nil
}

// ConvertToTemplateExpression transforms the templateData contents into
// an Fn::Join- compatible representation for template serialization.
// The templateData contents may include both golang text/template properties
// and single-line JSON Fn::Join supported serializations.
func ConvertToTemplateExpression(templateData io.Reader, additionalUserTemplateProperties map[string]interface{}) (*gocf.StringExpr, error) {
	converter := &templateConverter{
		templateReader:          templateData,
		additionalTemplateProps: additionalUserTemplateProperties,
	}
	return converter.expandTemplate().parseData().results()
}

func existingStackTemplate(serviceName string,
	session *session.Session,
	logger *logrus.Logger) (*gocf.Template, error) {
	template, templateExists := cloudFormationStackTemplateMap[serviceName]
	if !templateExists {
		templateParams := &cloudformation.GetTemplateInput{
			StackName: aws.String(serviceName),
		}
		logger.WithFields(logrus.Fields{
			"Service": serviceName,
		}).Info("Fetching existing CloudFormation template")

		cloudformationSvc := cloudformation.New(session)
		rawTemplate, rawTemplateErr := cloudformationSvc.GetTemplate(templateParams)
		if nil != rawTemplateErr {
			if strings.Contains(rawTemplateErr.Error(), "does not exist") {
				template = nil
			} else {
				return nil, rawTemplateErr
			}
		} else {
			t := gocf.Template{}
			jsonDecodeErr := json.NewDecoder(strings.NewReader(*rawTemplate.TemplateBody)).Decode(&t)
			if nil != jsonDecodeErr {
				return nil, jsonDecodeErr
			}
			template = &t
		}
		cloudFormationStackTemplateMap[serviceName] = template
	} else {
		logger.WithFields(logrus.Fields{
			"Service": serviceName,
		}).Debug("Using cached CloudFormation Template resources")
	}

	return template, nil
}

func existingLambdaResourceVersions(serviceName string,
	lambdaResourceName string,
	session *session.Session,
	logger *logrus.Logger) (*lambda.ListVersionsByFunctionOutput, error) {

	errorIsNotExist := func(apiError error) bool {
		return apiError != nil && strings.Contains(apiError.Error(), "does not exist")
	}

	logger.WithFields(logrus.Fields{
		"ResourceName": lambdaResourceName,
	}).Info("Fetching existing function versions")

	cloudFormationSvc := cloudformation.New(session)
	describeParams := &cloudformation.DescribeStackResourceInput{
		StackName:         aws.String(serviceName),
		LogicalResourceId: aws.String(lambdaResourceName),
	}
	describeResponse, describeResponseErr := cloudFormationSvc.DescribeStackResource(describeParams)
	logger.WithFields(logrus.Fields{
		"Response":    describeResponse,
		"ResponseErr": describeResponseErr,
	}).Debug("Describe response")
	if errorIsNotExist(describeResponseErr) {
		return nil, nil
	} else if describeResponseErr != nil {
		return nil, describeResponseErr
	}

	listVersionsParams := &lambda.ListVersionsByFunctionInput{
		FunctionName: describeResponse.StackResourceDetail.PhysicalResourceId,
		MaxItems:     aws.Int64(128),
	}
	lambdaSvc := lambda.New(session)
	listVersionsResp, listVersionsRespErr := lambdaSvc.ListVersionsByFunction(listVersionsParams)
	if errorIsNotExist(listVersionsRespErr) {
		return nil, nil
	} else if listVersionsRespErr != nil {
		return nil, listVersionsRespErr
	}
	logger.WithFields(logrus.Fields{
		"Response":    listVersionsResp,
		"ResponseErr": listVersionsRespErr,
	}).Debug("ListVersionsByFunction")
	return listVersionsResp, nil
}

// AutoIncrementingLambdaVersionInfo is dynamically populated during
// a call AddAutoIncrementingLambdaVersionResource. The VersionHistory
// is a map of published versions to their CloudFormation resource names
type AutoIncrementingLambdaVersionInfo struct {
	// The version that will be published as part of this operation
	CurrentVersion int
	// The CloudFormation resource name that defines the
	// AWS::Lambda::Version resource to be included with this operation
	CurrentVersionResourceName string
	// The version history that maps a published version value
	// to its CloudFormation resource name. Used for defining lagging
	// indicator Alias values
	VersionHistory map[int]string
}

// AddAutoIncrementingLambdaVersionResource inserts a new
// AWS::Lambda::Version resource into the template. It uses
// the existing CloudFormation template representation
// to determine the version index to append. The returned
// map is from `versionIndex`->`CloudFormationResourceName`
// to support second-order AWS::Lambda::Alias records on a
// per-version level
func AddAutoIncrementingLambdaVersionResource(serviceName string,
	lambdaResourceName string,
	cfTemplate *gocf.Template,
	logger *logrus.Logger) (*AutoIncrementingLambdaVersionInfo, error) {

	// Get the template
	session, sessionErr := session.NewSession()
	if sessionErr != nil {
		return nil, sessionErr
	}

	// Get the current template - for each version we find in the version listing
	// we look up the actual CF resource and copy it into this template
	existingStackDefinition, existingStackDefinitionErr := existingStackTemplate(serviceName,
		session,
		logger)
	if nil != existingStackDefinitionErr {
		return nil, existingStackDefinitionErr
	}

	// TODO - fetch the template and look up the resources
	existingVersions, existingVersionsErr := existingLambdaResourceVersions(serviceName,
		lambdaResourceName,
		session,
		logger)
	if nil != existingVersionsErr {
		return nil, existingVersionsErr
	}

	// Initialize the auto incrementing version struct
	autoIncrementingLambdaVersionInfo := AutoIncrementingLambdaVersionInfo{
		CurrentVersion:             0,
		CurrentVersionResourceName: "",
		VersionHistory:             make(map[int]string, 0),
	}

	lambdaVersionResourceName := func(versionIndex int) string {
		return sparta.CloudFormationResourceName(lambdaResourceName,
			"version",
			strconv.Itoa(versionIndex))
	}

	if nil != existingVersions {
		// Add the CloudFormation resource
		logger.WithFields(logrus.Fields{
			"VersionCount": len(existingVersions.Versions) - 1, // Ignore $LATEST
			"ResourceName": lambdaResourceName,
		}).Info("Total number of published versions")

		for _, eachEntry := range existingVersions.Versions {
			versionIndex, versionIndexErr := strconv.Atoi(*eachEntry.Version)
			if nil == versionIndexErr {
				// Find the existing resource...
				versionResourceName := lambdaVersionResourceName(versionIndex)
				if nil == existingStackDefinition {
					return nil, fmt.Errorf("Unable to find exising Version resource in nil Template")
				}
				cfResourceDefinition, cfResourceDefinitionExists := existingStackDefinition.Resources[versionResourceName]
				if !cfResourceDefinitionExists {
					return nil, fmt.Errorf("Unable to find exising Version resource (Resource: %s, Version: %d) in template",
						versionResourceName,
						versionIndex)
				}
				cfTemplate.Resources[versionResourceName] = cfResourceDefinition
				// Add the CloudFormation resource
				logger.WithFields(logrus.Fields{
					"Version":      versionIndex,
					"ResourceName": versionResourceName,
				}).Debug("Preserving Lambda version")

				// Store the state, tracking the latest version
				autoIncrementingLambdaVersionInfo.VersionHistory[versionIndex] = versionResourceName
				if versionIndex > autoIncrementingLambdaVersionInfo.CurrentVersion {
					autoIncrementingLambdaVersionInfo.CurrentVersion = versionIndex
				}
			}
		}
	}

	// Bump the version and add a new entry...
	autoIncrementingLambdaVersionInfo.CurrentVersion++
	versionResource := &gocf.LambdaVersion{
		FunctionName: gocf.GetAtt(lambdaResourceName, "Arn").String(),
	}
	autoIncrementingLambdaVersionInfo.CurrentVersionResourceName = lambdaVersionResourceName(autoIncrementingLambdaVersionInfo.CurrentVersion)
	cfTemplate.AddResource(autoIncrementingLambdaVersionInfo.CurrentVersionResourceName, versionResource)

	// Log the version we're about to publish...
	logger.WithFields(logrus.Fields{
		"ResourceName": lambdaResourceName,
		"StackVersion": autoIncrementingLambdaVersionInfo.CurrentVersion,
	}).Info("Inserting new version resource")

	return &autoIncrementingLambdaVersionInfo, nil
}
