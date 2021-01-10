package cloudformation

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/briandowns/spinner"
	humanize "github.com/dustin/go-humanize"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

//var cacheLock sync.Mutex

func init() {
	rand.Seed(time.Now().Unix())
}

// RE to ensure CloudFormation compatible resource names
// Issue: https://github.com/mweagle/Sparta/issues/8
// Ref: http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/resources-section-structure.html
var reCloudFormationInvalidChars = regexp.MustCompile("[^A-Za-z0-9]+")

// maximum amount of time allowed for polling CloudFormation
var cloudformationPollingTimeout = 3 * time.Minute

////////////////////////////////////////////////////////////////////////////////
// Private
////////////////////////////////////////////////////////////////////////////////

// If the platform specific implementation of user.Current()
// isn't available, go get something that's a "stable" user
// name
func defaultUserName() string {
	userName := os.Getenv("USER")
	if userName == "" {
		userName = os.Getenv("USERNAME")
	}
	if userName == "" {
		userName = fmt.Sprintf("user%d", os.Getuid())
	}
	return userName
}

type resourceProvisionMetrics struct {
	resourceType      string
	logicalResourceID string
	startTime         time.Time
	endTime           time.Time
	elapsed           time.Duration
}

// BEGIN - templateConverter
// Struct to encapsulate transforming data into
type templateConverter struct {
	templateReader          io.Reader
	additionalTemplateProps map[string]interface{}
	// internals
	doQuote          bool
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
	// TODO - parse this better... 🤔
	// First see if it's JSON...if so, just walk it
	reAWSProp := regexp.MustCompile("\\{\\s*\"\\s*(Ref|Fn::GetAtt|Fn::FindInMap)")
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
					converter.contents = append(converter.contents, gocf.String(head))
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
							if converter.doQuote {
								converter.contents = append(converter.contents, gocf.Join("",
									gocf.String("\""),
									parsedContents,
									gocf.String("\"")))
							} else {
								converter.contents = append(converter.contents, parsedContents)
							}
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
					converter.conversionError = fmt.Errorf("invalid CloudFormation JSON expression on line: %s", eachLine)
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
				converter.contents = append(converter.contents, gocf.String(appendLine))
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

// END - templateConverter

func cloudformationPollingDelay() time.Duration {
	return time.Duration(3+rand.Int31n(5)) * time.Second
}

func updateStackViaChangeSet(serviceName string,
	cfTemplate *gocf.Template,
	cfTemplateURL string,
	stackParameters map[string]string,
	awsTags map[string]string,
	awsCloudFormation *cloudformation.CloudFormation,
	logger *zerolog.Logger) error {

	// Create a change set name...
	changeSetRequestName := ResourceName(fmt.Sprintf("%sChangeSet", serviceName))
	_, changesErr := CreateStackChangeSet(changeSetRequestName,
		serviceName,
		cfTemplate,
		cfTemplateURL,
		stackParameters,
		awsTags,
		awsCloudFormation,
		logger)
	if nil != changesErr {
		return changesErr
	}

	//////////////////////////////////////////////////////////////////////////////
	// Apply the change
	executeChangeSetInput := cloudformation.ExecuteChangeSetInput{
		ChangeSetName: aws.String(changeSetRequestName),
		StackName:     aws.String(serviceName),
	}
	executeChangeSetOutput, executeChangeSetError := awsCloudFormation.ExecuteChangeSet(&executeChangeSetInput)

	logger.Debug().
		Interface("ExecuteChangeSetOutput", executeChangeSetOutput).
		Msg("ExecuteChangeSet result")

	if nil == executeChangeSetError {
		logger.Info().
			Str("StackName", serviceName).
			Msg("Issued ExecuteChangeSet request")
	}
	return executeChangeSetError

}

// func existingLambdaResourceVersions(serviceName string,
// 	lambdaResourceName string,
// 	session *session.Session,
// 	logger *zerolog.Logger) (*lambda.ListVersionsByFunctionOutput, error) {

// 	errorIsNotExist := func(apiError error) bool {
// 		return apiError != nil && strings.Contains(apiError.Error(), "does not exist")
// 	}

// 	logger.WithFields(logrus.Fields{
// 		"ResourceName": lambdaResourceName,
// 	}).Info("Fetching existing function versions")

// 	cloudFormationSvc := cloudformation.New(session)
// 	describeParams := &cloudformation.DescribeStackResourceInput{
// 		StackName:         aws.String(serviceName),
// 		LogicalResourceId: aws.String(lambdaResourceName),
// 	}
// 	describeResponse, describeResponseErr := cloudFormationSvc.DescribeStackResource(describeParams)
// 	logger.WithFields(logrus.Fields{
// 		"Response":    describeResponse,
// 		"ResponseErr": describeResponseErr,
// 	}).Debug("Describe response")
// 	if errorIsNotExist(describeResponseErr) {
// 		return nil, nil
// 	} else if describeResponseErr != nil {
// 		return nil, describeResponseErr
// 	}

// 	listVersionsParams := &lambda.ListVersionsByFunctionInput{
// 		FunctionName: describeResponse.StackResourceDetail.PhysicalResourceId,
// 		MaxItems:     aws.Int64(128),
// 	}
// 	lambdaSvc := lambda.New(session)
// 	listVersionsResp, listVersionsRespErr := lambdaSvc.ListVersionsByFunction(listVersionsParams)
// 	if errorIsNotExist(listVersionsRespErr) {
// 		return nil, nil
// 	} else if listVersionsRespErr != nil {
// 		return nil, listVersionsRespErr
// 	}
// 	logger.WithFields(logrus.Fields{
// 		"Response":    listVersionsResp,
// 		"ResponseErr": listVersionsRespErr,
// 	}).Debug("ListVersionsByFunction")
// 	return listVersionsResp, nil
// }

func toExpressionSlice(input interface{}) ([]string, error) {
	var expressions []string
	slice, sliceOK := input.([]interface{})
	if !sliceOK {
		return nil, fmt.Errorf("failed to convert to slice")
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
		return nil, fmt.Errorf("data for FnJoinExpr is empty")
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
				return nil, fmt.Errorf("invalid params for Fn::GetAtt: %s", eachValue)
			}
			return gocf.GetAtt(attrValues[0], attrValues[1]).String(), nil
		case "Fn::FindInMap":
			attrValues, attrValuesErr := toExpressionSlice(eachValue)
			if nil != attrValuesErr {
				return nil, attrValuesErr
			}
			if len(attrValues) != 3 {
				return nil, fmt.Errorf("invalid params for Fn::FindInMap: %s", eachValue)
			}
			return gocf.FindInMap(attrValues[0], gocf.String(attrValues[1]), gocf.String(attrValues[2])), nil
		}
	}
	return nil, fmt.Errorf("unsupported AWS Function detected: %#v", data)
}

func stackCapabilities(template *gocf.Template) []*string {
	capabilitiesMap := make(map[string]bool)

	// Only require IAM capability if the definition requires it.
	for _, eachResource := range template.Resources {
		if eachResource.Properties.CfnResourceType() == "AWS::IAM::Role" {
			capabilitiesMap["CAPABILITY_IAM"] = true
			switch typedResource := eachResource.Properties.(type) {
			case gocf.IAMRole:
				capabilitiesMap["CAPABILITY_NAMED_IAM"] = (typedResource.RoleName != nil)
			case *gocf.IAMRole:
				capabilitiesMap["CAPABILITY_NAMED_IAM"] = (typedResource.RoleName != nil)
			}
		}
	}
	capabilities := make([]*string, len(capabilitiesMap))
	capabilitiesIndex := 0
	for eachKey := range capabilitiesMap {
		capabilities[capabilitiesIndex] = aws.String(eachKey)
		capabilitiesIndex++
	}
	return capabilities
}

////////////////////////////////////////////////////////////////////////////////
// Public
////////////////////////////////////////////////////////////////////////////////

// DynamicValueToStringExpr is a DRY function to type assert
// a potentiall dynamic value into a gocf.Stringable
// satisfying type
func DynamicValueToStringExpr(dynamicValue interface{}) gocf.Stringable {
	var stringExpr gocf.Stringable
	switch typedValue := dynamicValue.(type) {
	case string:
		stringExpr = gocf.String(typedValue)
	case *gocf.StringExpr:
		stringExpr = typedValue
	case gocf.Stringable:
		stringExpr = typedValue.String()
	default:
		panic(fmt.Sprintf("Unsupported dynamic value type: %+v", typedValue))
	}
	return stringExpr
}

// S3AllKeysArnForBucket returns a CloudFormation-compatible Arn expression
// (string or Ref) for all bucket keys (`/*`).  The bucket
// parameter may be either a string or an interface{} ("Ref: "myResource")
// value
func S3AllKeysArnForBucket(bucket interface{}) *gocf.StringExpr {
	arnParts := []gocf.Stringable{
		gocf.String("arn:aws:s3:::"),
		DynamicValueToStringExpr(bucket),
		gocf.String("/*"),
	}
	return gocf.Join("", arnParts...).String()
}

// S3ArnForBucket returns a CloudFormation-compatible Arn expression
// (string or Ref) suitable for template reference.  The bucket
// parameter may be either a string or an interface{} ("Ref: "myResource")
// value
func S3ArnForBucket(bucket interface{}) *gocf.StringExpr {
	arnParts := []gocf.Stringable{
		gocf.String("arn:aws:s3:::"),
		DynamicValueToStringExpr(bucket),
	}
	return gocf.Join("", arnParts...).String()
}

// MapToResourceTags transforms a go map[string]string to a CloudFormation-compliant
// Tags representation.  See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-resource-tags.html
func MapToResourceTags(tagMap map[string]string) []interface{} {
	tags := make([]interface{}, len(tagMap))
	tagsIndex := 0
	for eachKey, eachValue := range tagMap {
		tags[tagsIndex] = map[string]interface{}{
			"Key":   eachKey,
			"Value": eachValue,
		}
		tagsIndex++
	}
	return tags
}

// ConvertToTemplateExpression transforms the templateData contents into
// an Fn::Join- compatible representation for template serialization.
// The templateData contents may include both golang text/template properties
// and single-line JSON Fn::Join supported serializations.
func ConvertToTemplateExpression(templateData io.Reader,
	additionalUserTemplateProperties map[string]interface{}) (*gocf.StringExpr, error) {
	converter := &templateConverter{
		templateReader:          templateData,
		additionalTemplateProps: additionalUserTemplateProperties,
	}
	return converter.expandTemplate().parseData().results()
}

// ConvertToInlineJSONTemplateExpression transforms the templateData contents into
// an Fn::Join- compatible inline JSON representation for template serialization.
// The templateData contents may include both golang text/template properties
// and single-line JSON Fn::Join supported serializations.
func ConvertToInlineJSONTemplateExpression(templateData io.Reader,
	additionalUserTemplateProperties map[string]interface{}) (*gocf.StringExpr, error) {
	converter := &templateConverter{
		templateReader:          templateData,
		additionalTemplateProps: additionalUserTemplateProperties,
		doQuote:                 true,
	}
	return converter.expandTemplate().parseData().results()
}

// StackEvents returns the slice of cloudformation.StackEvents for the given stackID or stackName
func StackEvents(stackID string,
	eventFilterLowerBoundInclusive time.Time,
	awsSession *session.Session) ([]*cloudformation.StackEvent, error) {

	cfService := cloudformation.New(awsSession)
	var events []*cloudformation.StackEvent

	nextToken := ""
	for {
		params := &cloudformation.DescribeStackEventsInput{
			StackName: aws.String(stackID),
		}
		if len(nextToken) > 0 {
			params.NextToken = aws.String(nextToken)
		}

		resp, err := cfService.DescribeStackEvents(params)
		if nil != err {
			return nil, err
		}
		for _, eachEvent := range resp.StackEvents {
			if eachEvent.Timestamp.Equal(eventFilterLowerBoundInclusive) ||
				eachEvent.Timestamp.After(eventFilterLowerBoundInclusive) {
				events = append(events, eachEvent)
			}
		}
		if nil == resp.NextToken {
			break
		} else {
			nextToken = *resp.NextToken
		}
	}
	return events, nil
}

// WaitForStackOperationCompleteResult encapsulates the stackInfo
// following a WaitForStackOperationComplete call
type WaitForStackOperationCompleteResult struct {
	operationSuccessful bool
	stackInfo           *cloudformation.Stack
}

// WaitForStackOperationComplete is a blocking, polling based call that
// periodically fetches the stackID set of events and uses the state value
// to determine if an operation is complete
func WaitForStackOperationComplete(stackID string,
	pollingMessage string,
	awsCloudFormation *cloudformation.CloudFormation,
	logger *zerolog.Logger) (*WaitForStackOperationCompleteResult, error) {

	result := &WaitForStackOperationCompleteResult{}

	// Startup a spinner...
	charSetIndex := 39
	if strings.Contains(os.Getenv("LC_TERMINAL"), "iTerm") {
		charSetIndex = 39 // WAS 7 to handle iTerm improperly updating
	}
	cliSpinner := spinner.New(spinner.CharSets[charSetIndex],
		333*time.Millisecond)
	spinnerErr := cliSpinner.Color("red", "bold")
	if spinnerErr != nil {
		logger.Warn().
			Err(spinnerErr).
			Msg("Failed to set spinner color")
	}

	// Poll for the current stackID state, and
	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackID),
	}
	cliSpinnerStarted := false
	startTime := time.Now()

	for waitComplete := false; !waitComplete; {
		// Startup the spinner if needed...

		if !cliSpinnerStarted {
			cliSpinner.Start()
			defer cliSpinner.Stop()
			cliSpinnerStarted = true
		}
		spinnerText := fmt.Sprintf(" %s (requested: %s)",
			pollingMessage,
			humanize.Time(startTime))
		cliSpinner.Suffix = spinnerText

		// Then sleep and figure out if things are done...
		sleepDuration := time.Duration(11+rand.Int31n(13)) * time.Second
		time.Sleep(sleepDuration)

		describeStacksOutput, err := awsCloudFormation.DescribeStacks(describeStacksInput)
		if nil != err {
			// TODO - add retry iff we're RateExceeded due to collective access
			return nil, err
		}
		if len(describeStacksOutput.Stacks) <= 0 {
			return nil, fmt.Errorf("failed to enumerate stack info: %v", *describeStacksInput.StackName)
		}
		result.stackInfo = describeStacksOutput.Stacks[0]
		switch *(result.stackInfo).StackStatus {
		case cloudformation.StackStatusCreateComplete,
			cloudformation.StackStatusUpdateComplete:
			result.operationSuccessful = true
			waitComplete = true
		case
			// Include DeleteComplete as new provisions will automatically rollback
			cloudformation.StackStatusDeleteComplete,
			cloudformation.StackStatusCreateFailed,
			cloudformation.StackStatusDeleteFailed,
			cloudformation.StackStatusRollbackFailed,
			cloudformation.StackStatusRollbackComplete,
			cloudformation.StackStatusUpdateRollbackComplete:
			result.operationSuccessful = false
			waitComplete = true
		default:
			// If this is JSON output, just do the normal thing
			// NOP
		}
	}
	return result, nil
}

// StableResourceName returns a stable resource name
func StableResourceName(value string) string {
	return ResourceName(value, value)
}

// ResourceName returns a name suitable as a logical
// CloudFormation resource value.  See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/resources-section-structure.html
// for more information.  The `prefix` value should provide a hint as to the
// resource type (eg, `SNSConfigurator`, `ImageTranscoder`).  Note that the returned
// name is not content-addressable.
func ResourceName(prefix string, parts ...string) string {
	hash := sha1.New()
	_, writeErr := hash.Write([]byte(prefix))
	if writeErr != nil {
		fmt.Printf("Failed to write to hash: " + writeErr.Error())
	}
	if len(parts) <= 0 {
		randValue := rand.Int63()
		_, writeErr = hash.Write([]byte(strconv.FormatInt(randValue, 10)))
		if writeErr != nil {
			fmt.Printf("Failed to write to hash: " + writeErr.Error())
		}
	} else {
		for _, eachPart := range parts {
			_, writeErr = hash.Write([]byte(eachPart))
			if writeErr != nil {
				fmt.Printf("Failed to write to hash: " + writeErr.Error())
			}
		}
	}
	resourceName := fmt.Sprintf("%s%s", prefix, hex.EncodeToString(hash.Sum(nil)))

	// Ensure that any non alphanumeric characters are replaced with ""
	return reCloudFormationInvalidChars.ReplaceAllString(resourceName, "x")
}

// UploadTemplate marshals the given cfTemplate and uploads it to the
// supplied bucket using the given KeyName
func UploadTemplate(serviceName string,
	cfTemplate *gocf.Template,
	s3Bucket string,
	s3KeyName string,
	awsSession *session.Session,
	logger *zerolog.Logger) (string, error) {

	logger.Info().
		Str("Key", s3KeyName).
		Str("Bucket", s3Bucket).
		Msg("Uploading CloudFormation template")

	s3Uploader := s3manager.NewUploader(awsSession)

	// Serialize the template and upload it
	cfTemplateJSON, err := json.Marshal(cfTemplate)
	if err != nil {
		return "", errors.Wrap(err, "Failed to Marshal CloudFormation template")
	}

	// Upload the actual CloudFormation template to S3 to maximize the template
	// size limit
	// Ref: http://docs.aws.amazon.com/AWSCloudFormation/latest/APIReference/API_CreateStack.html
	contentBody := string(cfTemplateJSON)
	uploadInput := &s3manager.UploadInput{
		Bucket:      &s3Bucket,
		Key:         &s3KeyName,
		ContentType: aws.String("application/json"),
		Body:        strings.NewReader(contentBody),
	}
	templateUploadResult, templateUploadResultErr := s3Uploader.Upload(uploadInput)
	if nil != templateUploadResultErr {
		return "", templateUploadResultErr
	}

	// Be transparent
	logger.Info().
		Str("URL", templateUploadResult.Location).
		Msg("Template uploaded")
	return templateUploadResult.Location, nil
}

// StackExists returns whether the given stackName or stackID currently exists
func StackExists(stackNameOrID string, awsSession *session.Session, logger *zerolog.Logger) (bool, error) {
	cf := cloudformation.New(awsSession)

	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackNameOrID),
	}
	describeStacksOutput, err := cf.DescribeStacks(describeStacksInput)

	logger.Debug().
		Interface("DescribeStackOutput", describeStacksOutput).
		Msg("DescribeStackOutput results")

	exists := false
	if err != nil {
		logger.Debug().
			Err(err).
			Msg("DescribeStackOutput")

		// If the stack doesn't exist, then no worries
		if strings.Contains(err.Error(), "does not exist") {
			exists = false
		} else {
			return false, err
		}
	} else {
		exists = true
	}
	return exists, nil
}

// CreateStackChangeSet returns the DescribeChangeSetOutput
// for a given stack transformation
func CreateStackChangeSet(changeSetRequestName string,
	serviceName string,
	cfTemplate *gocf.Template,
	templateURL string,
	stackParameters map[string]string,
	stackTags map[string]string,
	awsCloudFormation *cloudformation.CloudFormation,
	logger *zerolog.Logger) (*cloudformation.DescribeChangeSetOutput, error) {

	cloudFormationParameters := make([]*cloudformation.Parameter,
		0,
		len(stackParameters))
	for eachKey, eachValue := range stackParameters {
		cloudFormationParameters = append(cloudFormationParameters, &cloudformation.Parameter{
			ParameterKey:   aws.String(eachKey),
			ParameterValue: aws.String(eachValue),
		})
	}

	capabilities := stackCapabilities(cfTemplate)
	changeSetInput := &cloudformation.CreateChangeSetInput{
		Capabilities:  capabilities,
		ChangeSetName: aws.String(changeSetRequestName),
		ClientToken:   aws.String(changeSetRequestName),
		Description:   aws.String(fmt.Sprintf("Change set for service: %s", serviceName)),
		StackName:     aws.String(serviceName),
		TemplateURL:   aws.String(templateURL),
		Parameters:    cloudFormationParameters,
	}
	if len(stackTags) != 0 {
		awsTags := []*cloudformation.Tag{}
		for eachKey, eachValue := range stackTags {
			awsTags = append(awsTags,
				&cloudformation.Tag{
					Key:   aws.String(eachKey),
					Value: aws.String(eachValue),
				})
		}
		changeSetInput.Tags = awsTags
	}
	_, changeSetError := awsCloudFormation.CreateChangeSet(changeSetInput)
	if nil != changeSetError {
		return nil, changeSetError
	}

	logger.Info().
		Str("StackName", serviceName).
		Msg("Issued CreateChangeSet request")

	describeChangeSetInput := cloudformation.DescribeChangeSetInput{
		ChangeSetName: aws.String(changeSetRequestName),
		StackName:     aws.String(serviceName),
	}

	var describeChangeSetOutput *cloudformation.DescribeChangeSetOutput

	// Loop, with a total timeout of 3 minutes
	startTime := time.Now()
	changeSetStabilized := false
	for !changeSetStabilized {
		sleepDuration := cloudformationPollingDelay()
		time.Sleep(sleepDuration)

		changeSetOutput, describeChangeSetError := awsCloudFormation.DescribeChangeSet(&describeChangeSetInput)

		if nil != describeChangeSetError {
			return nil, describeChangeSetError
		}
		describeChangeSetOutput = changeSetOutput
		// The current status of the change set, such as CREATE_IN_PROGRESS, CREATE_COMPLETE,
		// or FAILED.
		if nil != describeChangeSetOutput {
			switch *describeChangeSetOutput.Status {
			case "CREATE_IN_PROGRESS":
				// If this has taken more than 3 minutes, then that's an error
				elapsedTime := time.Since(startTime)
				if elapsedTime > cloudformationPollingTimeout {
					return nil, fmt.Errorf("failed to finalize ChangeSet within window: %s", elapsedTime.String())
				}
			case "CREATE_COMPLETE":
				changeSetStabilized = true
			case "FAILED":
				return nil, fmt.Errorf("failed to create ChangeSet: %#v", *describeChangeSetOutput)
			}
		}
	}

	logger.Debug().
		Interface("ChangeSetInput", changeSetInput).
		Interface("DescribeChangeSetOutput", describeChangeSetOutput).
		Msg("DescribeChangeSet result")

	//////////////////////////////////////////////////////////////////////////////
	// If there aren't any changes, then skip it...
	if len(describeChangeSetOutput.Changes) <= 0 {
		logger.Info().
			Str("StackName", serviceName).
			Msg("No changes detected for service")

		// Delete it...
		_, deleteChangeSetResultErr := DeleteChangeSet(serviceName,
			changeSetRequestName,
			awsCloudFormation)
		return nil, deleteChangeSetResultErr
	}
	return describeChangeSetOutput, nil
}

// DeleteChangeSet is a utility function that attempts to delete
// an existing CloudFormation change set, with a bit of retry
// logic in case of EC
func DeleteChangeSet(stackName string,
	changeSetRequestName string,
	awsCloudFormation *cloudformation.CloudFormation) (*cloudformation.DeleteChangeSetOutput, error) {

	// Delete request...
	deleteChangeSetInput := cloudformation.DeleteChangeSetInput{
		ChangeSetName: aws.String(changeSetRequestName),
		StackName:     aws.String(stackName),
	}

	startTime := time.Now()
	for {
		elapsedTime := time.Since(startTime)

		deleteChangeSetResults, deleteChangeSetResultErr := awsCloudFormation.DeleteChangeSet(&deleteChangeSetInput)
		if nil == deleteChangeSetResultErr {
			return deleteChangeSetResults, nil
		} else if strings.Contains(deleteChangeSetResultErr.Error(), "CREATE_IN_PROGRESS") {
			if elapsedTime > cloudformationPollingTimeout {
				return nil, fmt.Errorf("failed to delete ChangeSet within timeout window: %s", elapsedTime.String())
			}
			sleepDuration := cloudformationPollingDelay()
			time.Sleep(sleepDuration)
		} else {
			return nil, deleteChangeSetResultErr
		}
	}
}

// ListStacks returns a slice of stacks that meet the given filter.
func ListStacks(session *session.Session,
	maxReturned int,
	stackFilters ...string) ([]*cloudformation.StackSummary, error) {

	listStackInput := &cloudformation.ListStacksInput{
		StackStatusFilter: []*string{},
	}
	for _, eachFilter := range stackFilters {
		listStackInput.StackStatusFilter = append(listStackInput.StackStatusFilter, aws.String(eachFilter))
	}
	cloudformationSvc := cloudformation.New(session)
	accumulator := []*cloudformation.StackSummary{}
	for {
		listResult, listResultErr := cloudformationSvc.ListStacks(listStackInput)
		if listResultErr != nil {
			return nil, listResultErr
		}
		accumulator = append(accumulator, listResult.StackSummaries...)
		if len(accumulator) >= maxReturned || listResult.NextToken == nil {
			return accumulator, nil
		}
		listStackInput.NextToken = listResult.NextToken
	}
}

// ConvergeStackState ensures that the serviceName converges to the template
// state defined by cfTemplate. This function establishes a polling loop to determine
// when the stack operation has completed.
func ConvergeStackState(serviceName string,
	cfTemplate *gocf.Template,
	templateURL string,
	stackParameters map[string]string,
	tags map[string]string,
	startTime time.Time,
	operationTimeout time.Duration,
	awsSession *session.Session,
	outputsDividerChar string,
	dividerWidth int,
	logger *zerolog.Logger) (*cloudformation.Stack, error) {

	logger.Info().
		Interface("Parameters", stackParameters).
		Interface("Tags", tags).
		Str("Name", serviceName).
		Msg("Stack configuration")

	awsCloudFormation := cloudformation.New(awsSession)
	// Create the parameter values.
	// Update the tags
	exists, existsErr := StackExists(serviceName, awsSession, logger)
	if nil != existsErr {
		return nil, existsErr
	}
	stackID := ""
	if exists {
		updateErr := updateStackViaChangeSet(serviceName,
			cfTemplate,
			templateURL,
			stackParameters,
			tags,
			awsCloudFormation,
			logger)

		if nil != updateErr {
			return nil, updateErr
		}
		stackID = serviceName
	} else {
		var cloudFormationParameters []*cloudformation.Parameter
		for eachKey, eachValue := range stackParameters {
			cloudFormationParameters = append(cloudFormationParameters, &cloudformation.Parameter{
				ParameterKey:   aws.String(eachKey),
				ParameterValue: aws.String(eachValue),
			})
		}
		awsTags := []*cloudformation.Tag{}
		if nil != tags {
			for eachKey, eachValue := range tags {
				awsTags = append(awsTags,
					&cloudformation.Tag{
						Key:   aws.String(eachKey),
						Value: aws.String(eachValue),
					})
			}
		}
		// Create stack
		createStackInput := &cloudformation.CreateStackInput{
			StackName:        aws.String(serviceName),
			TemplateURL:      aws.String(templateURL),
			TimeoutInMinutes: aws.Int64(int64(operationTimeout.Minutes())),
			OnFailure:        aws.String(cloudformation.OnFailureDelete),
			Capabilities:     stackCapabilities(cfTemplate),
			Parameters:       cloudFormationParameters,
		}
		if len(awsTags) != 0 {
			createStackInput.Tags = awsTags
		}
		createStackResponse, createStackResponseErr := awsCloudFormation.CreateStack(createStackInput)
		if nil != createStackResponseErr {
			return nil, createStackResponseErr
		}
		logger.Info().
			Str("StackID", *createStackResponse.StackId).
			Msg("Creating stack")
		for eachKey, eachVal := range stackParameters {
			logger.Info().
				Str(eachKey, eachVal).
				Msg("Stack parameter")

		}
		stackID = *createStackResponse.StackId
	}
	// Wait for the operation to succeed
	pollingMessage := "Waiting for CloudFormation operation to complete"
	convergeResult, convergeErr := WaitForStackOperationComplete(stackID,
		pollingMessage,
		awsCloudFormation,
		logger)
	if nil != convergeErr {
		return nil, convergeErr
	}
	// Get the events and assemble them into either errors to output
	// or summary information
	resourceMetrics := make(map[string]*resourceProvisionMetrics)
	errorMessages := []string{}
	events, err := StackEvents(stackID, startTime, awsSession)
	if nil != err {
		return nil, fmt.Errorf("failed to retrieve stack events: %s", err.Error())
	}

	for _, eachEvent := range events {
		switch *eachEvent.ResourceStatus {
		case cloudformation.ResourceStatusCreateFailed,
			cloudformation.ResourceStatusDeleteFailed,
			cloudformation.ResourceStatusUpdateFailed:
			errMsg := fmt.Sprintf("\tError ensuring %s (%s): %s",
				aws.StringValue(eachEvent.ResourceType),
				aws.StringValue(eachEvent.LogicalResourceId),
				aws.StringValue(eachEvent.ResourceStatusReason))
			// Only append if the resource failed because something else failed
			// and this resource was canceled.
			if !strings.Contains(errMsg, "cancelled") {
				errorMessages = append(errorMessages, errMsg)
			}
		case cloudformation.ResourceStatusCreateInProgress,
			cloudformation.ResourceStatusUpdateInProgress:
			existingMetric, existingMetricExists := resourceMetrics[*eachEvent.LogicalResourceId]
			if !existingMetricExists {
				existingMetric = &resourceProvisionMetrics{}
			}
			existingMetric.resourceType = *eachEvent.ResourceType
			existingMetric.logicalResourceID = *eachEvent.LogicalResourceId
			existingMetric.startTime = *eachEvent.Timestamp
			resourceMetrics[*eachEvent.LogicalResourceId] = existingMetric
		case cloudformation.ResourceStatusCreateComplete,
			cloudformation.ResourceStatusUpdateComplete:
			existingMetric, existingMetricExists := resourceMetrics[*eachEvent.LogicalResourceId]
			if !existingMetricExists {
				existingMetric = &resourceProvisionMetrics{}
			}
			existingMetric.logicalResourceID = *eachEvent.LogicalResourceId
			existingMetric.endTime = *eachEvent.Timestamp
			resourceMetrics[*eachEvent.LogicalResourceId] = existingMetric
		default:
			// NOP
		}
	}

	// If it didn't work, then output some failure information
	if !convergeResult.operationSuccessful {
		for _, eachError := range errorMessages {
			logger.Error().Err(errors.New(eachError)).Msg("Stack provisioning error")
		}
		return nil, fmt.Errorf("failed to provision: %s", serviceName)
	}

	// Rip through the events so that we can output exactly how long it took to
	// update each resource
	resourceStats := make([]*resourceProvisionMetrics, len(resourceMetrics))
	resourceStatIndex := 0
	for _, eachResource := range resourceMetrics {
		eachResource.elapsed = eachResource.endTime.Sub(eachResource.startTime)
		resourceStats[resourceStatIndex] = eachResource
		resourceStatIndex++
	}
	// Create a slice with them all, sorted by total elapsed mutation time
	sort.Slice(resourceStats, func(i, j int) bool {
		return resourceStats[i].elapsed > resourceStats[j].elapsed
	})

	// Output the sorted time it took to create the necessary resources...
	outputHeader := "CloudFormation Metrics "
	suffix := strings.Repeat(outputsDividerChar, dividerWidth-len(outputHeader))
	logger.Info().Msgf("%s%s", outputHeader, suffix)

	for _, eachResourceStat := range resourceStats {
		logger.Info().
			Str("Resource", eachResourceStat.logicalResourceID).
			Str("Type", eachResourceStat.resourceType).
			Dur("Duration", eachResourceStat.elapsed).
			Msg("   Operation duration")
	}
	if nil != convergeResult.stackInfo.Outputs {
		// Add a nice divider if there are Stack specific output
		outputHeader := "Stack Outputs "
		suffix := strings.Repeat(outputsDividerChar, dividerWidth-len(outputHeader))
		logger.Info().Msgf("%s%s", outputHeader, suffix)

		for _, eachOutput := range convergeResult.stackInfo.Outputs {
			logger.Info().
				Str("Value", aws.StringValue(eachOutput.OutputValue)).
				Str("Description", aws.StringValue(eachOutput.Description)).
				Msgf("    %s", aws.StringValue(eachOutput.OutputKey))
		}
	}
	return convergeResult.stackInfo, nil
}

// UserAccountScopedStackName returns a CloudFormation stack
// name that takes into account the current username that is
//associated with the supplied AWS credentials
/*
A stack name can contain only alphanumeric characters
(case sensitive) and hyphens. It must start with an alphabetic
\character and cannot be longer than 128 characters.
*/
func UserAccountScopedStackName(basename string,
	awsSession *session.Session) (string, error) {
	awsName, awsNameErr := platformAccountUserName(awsSession)
	if awsNameErr != nil {
		return "", awsNameErr
	}
	userName := strings.Replace(awsName, " ", "-", -1)
	userName = strings.Replace(userName, ".", "-", -1)
	return fmt.Sprintf("%s-%s", basename, userName), nil
}

// UserScopedStackName returns a CloudFormation stack
// name that takes into account the current username
/*
A stack name can contain only alphanumeric characters
(case sensitive) and hyphens. It must start with an alphabetic
\character and cannot be longer than 128 characters.
*/
func UserScopedStackName(basename string) string {
	platformUserName := platformUserName()
	if platformUserName == "" {
		return basename
	}
	userName := strings.Replace(platformUserName, " ", "-", -1)
	userName = strings.Replace(userName, ".", "-", -1)
	return fmt.Sprintf("%s-%s", basename, userName)
}
