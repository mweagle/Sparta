package cloudformation

import (
	"bytes"
	"encoding/json"
	"fmt"
	gocf "github.com/crewjam/go-cloudformation"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"text/template"
)

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
func (converter *templateConverter) toExpressionSlice(input interface{}) ([]string, error) {
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

func (converter *templateConverter) parseData() *templateConverter {
	if converter.conversionError != nil {
		return converter
	}
	reAWSProp := regexp.MustCompile("\\{\\s*\"([Ref|Fn\\:\\:\\w+])")
	splitData := strings.Split(converter.expandedTemplate, "\n")

	for eachLineIndex, eachLine := range splitData {
		curContents := eachLine
		for {
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
				tail := curContents[0:]
				for closingTokenIndex := strings.Index(tail, "}"); closingTokenIndex >= 0; closingTokenIndex = strings.Index(tail, "}") {

					testBlock := tail[0 : closingTokenIndex+1]
					curContents = tail[closingTokenIndex+1:]
					err := json.Unmarshal([]byte(testBlock), &parsed)
					if err != nil {
						break
					} else {
						for eachKey, eachValue := range parsed {
							switch eachKey {
							case "Ref":
								converter.contents = append(converter.contents, gocf.Ref(eachValue.(string)))
							case "Fn::GetAtt":
								attrValues, attrValuesErr := converter.toExpressionSlice(eachValue)
								if nil != attrValuesErr {
									converter.conversionError = attrValuesErr
									return converter
								}
								if len(attrValues) != 2 {
									converter.conversionError = fmt.Errorf("Invalid params for Fn::GetAtt: %s", eachValue)
									return converter
								}
								converter.contents = append(converter.contents, gocf.GetAtt(attrValues[0], attrValues[1]))
							case "Fn::FindInMap":
								attrValues, attrValuesErr := converter.toExpressionSlice(eachValue)
								if nil != attrValuesErr {
									converter.conversionError = attrValuesErr
									return converter
								}
								if len(attrValues) != 3 {
									converter.conversionError = fmt.Errorf("Invalid params for Fn::FindInMap: %s", eachValue)
									return converter
								}
								converter.contents = append(converter.contents, gocf.FindInMap(attrValues[0], gocf.String(attrValues[1]), gocf.String(attrValues[2])))
							default:
								converter.conversionError = fmt.Errorf("Unsupported AWS Function detected: %s", testBlock)
								return converter
							}
						}
						tail = tail[closingTokenIndex+1:]
					}
				}
				if len(parsed) <= 0 {
					// We never did find the end...
					converter.conversionError = fmt.Errorf("Invalid CloudFormation JSON expression on line: %s", eachLine)
					return converter
				}
			}
			// No match, just include it iff there is another line afterwards
			newlineValue := ""
			if eachLineIndex < (len(splitData) - 1) {
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
