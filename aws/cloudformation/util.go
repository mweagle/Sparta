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

// ConvertToTemplateExpression transforms the templateDataReader contents into
// an Fn::Join- compatible representation for template serialization.
func ConvertToTemplateExpression(templateDataReader io.Reader, additionalUserTemplateProperties map[string]interface{}) (*gocf.StringExpr, error) {

	templateDataBytes, templateDataErr := ioutil.ReadAll(templateDataReader)
	if nil != templateDataErr {
		return nil, templateDataErr
	}
	templateData := string(templateDataBytes)

	toExpressionSlice := func(input interface{}) ([]string, error) {
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
	parsedTemplate, templateErr := template.New("CloudFormation").Parse(templateData)
	if nil != templateErr {
		return nil, fmt.Errorf("Failed to parse template: %s", templateErr.Error())
	}
	output := &bytes.Buffer{}
	executeErr := parsedTemplate.Execute(output, additionalUserTemplateProperties)
	if nil != executeErr {
		return nil, fmt.Errorf("Failed to execute template: %s", executeErr.Error())
	}
	reAWSProp := regexp.MustCompile("\\{\\s*\"([Ref|Fn\\:\\:\\w+])")
	splitData := strings.Split(output.String(), "\n")
	var contents []gocf.Stringable
	for eachLineIndex, eachLine := range splitData {
		curContents := eachLine
		for {
			matchInfo := reAWSProp.FindStringSubmatchIndex(curContents)
			if nil != matchInfo {
				// If there's anything at the head, push it.
				if matchInfo[0] != 0 {
					head := curContents[0:matchInfo[0]]
					contents = append(contents, gocf.String(fmt.Sprintf("%s", head)))
					curContents = curContents[len(head):len(curContents)]
				}
				// There's at least one match...find the closing brace...
				var parsed map[string]interface{}
				tail := curContents[0:len(curContents)]
				for closingTokenIndex := strings.Index(tail, "}"); closingTokenIndex >= 0; closingTokenIndex = strings.Index(tail, "}") {

					testBlock := tail[0 : closingTokenIndex+1]
					curContents = tail[closingTokenIndex+1 : len(tail)]
					err := json.Unmarshal([]byte(testBlock), &parsed)
					if err != nil {
						break
					} else {
						for eachKey, eachValue := range parsed {
							switch eachKey {
							case "Ref":
								contents = append(contents, gocf.Ref(eachValue.(string)))
							case "Fn::GetAtt":
								attrValues, attrValuesErr := toExpressionSlice(eachValue)
								if nil != attrValuesErr {
									return nil, attrValuesErr
								}
								if len(attrValues) != 2 {
									return nil, fmt.Errorf("Invalid params for Fn::GetAtt: %s", eachValue)
								}
								contents = append(contents, gocf.GetAtt(attrValues[0], attrValues[1]))
							case "Fn::FindInMap":
								attrValues, attrValuesErr := toExpressionSlice(eachValue)
								if nil != attrValuesErr {
									return nil, attrValuesErr
								}
								if len(attrValues) != 3 {
									return nil, fmt.Errorf("Invalid params for Fn::FindInMap: %s", eachValue)
								}
								contents = append(contents, gocf.FindInMap(attrValues[0], gocf.String(attrValues[1]), gocf.String(attrValues[2])))
							default:
								return nil, fmt.Errorf("Unsupported AWS Function detected: %s", testBlock)
							}
						}
						tail = tail[closingTokenIndex+1 : len(tail)]
					}
				}
				if len(parsed) <= 0 {
					// We never did find the end...
					return nil, fmt.Errorf("Invalid CloudFormation JSON expression on line: %s", eachLine)
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
				contents = append(contents, gocf.String(appendLine))
			}
			break
		}
	}
	return gocf.Join("", contents...), nil
}
