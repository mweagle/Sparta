package cloudformation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
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

// AddAutoIncrementingLambdaVersionResource inserts a new
// AWS::Lambda::Version resource into the template. It uses
// the existing CloudFormation template representation
// to determine the
func AddAutoIncrementingLambdaVersionResource(serviceName string,
	lambdaResourceName string,
	cfTemplate *gocf.Template,
	logger *logrus.Logger) error {

	stackTemplate, exists := cloudFormationStackTemplateMap[serviceName]
	if !exists {
		// Get the template
		sess, err := session.NewSession()
		if err != nil {
			fmt.Println("failed to create AWS session,", err)
			return err
		}
		logger.WithFields(logrus.Fields{
			"Service": serviceName,
		}).Info("Fetching existing Stack template for Lambda function versioning")

		svc := cloudformation.New(sess)
		params := &cloudformation.GetTemplateInput{
			StackName: aws.String(serviceName),
		}
		getTemplate, getTemplateErr := svc.GetTemplate(params)
		if nil != getTemplateErr {
			logger.WithFields(logrus.Fields{
				"GetTemplateErr": getTemplateErr,
			}).Debug("Error fetching current template")

			if strings.Contains(getTemplateErr.Error(), "does not exist") {
				cloudFormationStackTemplateMap[serviceName] = nil
				stackTemplate = nil
			} else {
				return getTemplateErr
			}
		} else {
			t := gocf.Template{}
			decodeErr := json.NewDecoder(strings.NewReader(*(getTemplate.TemplateBody))).Decode(&t)
			if nil != decodeErr {
				return decodeErr
			}
			cloudFormationStackTemplateMap[serviceName] = &t
			stackTemplate = &t
		}
	}

	lambdaVersionResourceName := func(versionIndex int) string {
		return sparta.CloudFormationResourceName(lambdaResourceName,
			"version",
			strconv.Itoa(versionIndex))
	}

	nextVersion := 1
	if nil != stackTemplate {
		// Copy all the existing resources starting with nextVersion
		for nextVersion >= 0 {
			testVersionName := lambdaVersionResourceName(nextVersion)
			existingResource, exists := stackTemplate.Resources[testVersionName]
			if exists {
				logger.WithFields(logrus.Fields{
					"Version":  nextVersion,
					"Resource": existingResource,
				}).Debug("Preserving Lambda version")

				cfTemplate.Resources[testVersionName] = existingResource
				nextVersion++
			} else {
				break
			}
		}
	}

	// Then add a new version resource
	versionResource := &gocf.LambdaVersion{
		FunctionName: gocf.GetAtt(lambdaResourceName, "Arn").String(),
	}
	cfTemplate.AddResource(lambdaVersionResourceName(nextVersion), versionResource)
	logger.WithFields(logrus.Fields{
		"Version":  nextVersion,
		"Function": lambdaResourceName,
	}).Info("Registering new AWS::Lambda::Version resource")
	return nil
}
