package cloudformation

import (
	"bytes"
	"context"
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

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2S3Manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	awsv2CF "github.com/aws/aws-sdk-go-v2/service/cloudformation"
	awsv2CFTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	awsv2S3 "github.com/aws/aws-sdk-go-v2/service/s3"
	gof "github.com/awslabs/goformation/v5/cloudformation"
	gofiam "github.com/awslabs/goformation/v5/cloudformation/iam"
	"github.com/briandowns/spinner"
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
	contents         []string
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
					converter.contents = append(converter.contents, head)
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
								converter.contents = append(converter.contents,
									gof.Join("", []string{
										"\"",
										parsedContents,
										"\""}))
							} else {
								converter.contents = append(converter.contents, parsedContents)
							}
							curContents = curContents[indexPos+1:]
							if len(curContents) <= 0 && (eachLineIndex < (splitDataLineCount - 1)) {
								converter.contents = append(converter.contents, "\n")
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
				converter.contents = append(converter.contents, appendLine)
				break
			}
		}
	}
	return converter
}

func (converter *templateConverter) results() (string, error) {
	if nil != converter.conversionError {
		return "", converter.conversionError
	}
	return gof.Join("", converter.contents), nil
}

// END - templateConverter

func cloudformationPollingDelay() time.Duration {
	return time.Duration(3+rand.Int31n(5)) * time.Second
}

func updateStackViaChangeSet(ctx context.Context,
	serviceName string,
	cfTemplate *gof.Template,
	cfTemplateURL string,
	stackParameters map[string]string,
	awsTags map[string]string,
	awsCloudFormation *awsv2CF.Client,
	logger *zerolog.Logger) error {

	// Create a change set name...
	changeSetRequestName := ResourceName(fmt.Sprintf("%sChangeSet", serviceName))
	_, changesErr := CreateStackChangeSet(ctx,
		changeSetRequestName,
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
	executeChangeSetInput := awsv2CF.ExecuteChangeSetInput{
		ChangeSetName: awsv2.String(changeSetRequestName),
		StackName:     awsv2.String(serviceName),
	}
	executeChangeSetOutput, executeChangeSetError := awsCloudFormation.ExecuteChangeSet(ctx,
		&executeChangeSetInput)

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

func parseFnJoinExpr(data map[string]interface{}) (string, error) {
	if len(data) <= 0 {
		return "", fmt.Errorf("data for FnJoinExpr is empty")
	}
	for eachKey, eachValue := range data {
		switch eachKey {
		case "Ref":
			return gof.Ref(eachValue.(string)), nil
		case "Fn::GetAtt":
			attrValues, attrValuesErr := toExpressionSlice(eachValue)
			if nil != attrValuesErr {
				return "", attrValuesErr
			}
			if len(attrValues) != 2 {
				return "", fmt.Errorf("invalid params for Fn::GetAtt: %s", eachValue)
			}
			return gof.GetAtt(attrValues[0], attrValues[1]), nil
		case "Fn::FindInMap":
			attrValues, attrValuesErr := toExpressionSlice(eachValue)
			if nil != attrValuesErr {
				return "", attrValuesErr
			}
			if len(attrValues) != 3 {
				return "", fmt.Errorf("invalid params for Fn::FindInMap: %s", eachValue)
			}
			return gof.FindInMap(attrValues[0],
				attrValues[1], attrValues[2]), nil
		}
	}
	return "", fmt.Errorf("unsupported AWS Function detected: %#v", data)
}

func stackCapabilities(template *gof.Template) []awsv2CFTypes.Capability {
	capabilitiesMap := make(map[awsv2CFTypes.Capability]bool)

	// Only require IAM capability if the definition requires it.
	for _, eachResource := range template.Resources {
		if eachResource.AWSCloudFormationType() == "AWS::IAM::Role" {
			capabilitiesMap["CAPABILITY_IAM"] = true
			switch typedResource := eachResource.(type) {
			case *gofiam.Role:
				capabilitiesMap[awsv2CFTypes.CapabilityCapabilityNamedIam] = (typedResource.RoleName != "")
			}
		}
	}
	capabilities := make([]awsv2CFTypes.Capability, len(capabilitiesMap))
	capabilitiesIndex := 0
	for eachKey := range capabilitiesMap {
		capabilities[capabilitiesIndex] = eachKey
		capabilitiesIndex++
	}
	return capabilities
}

////////////////////////////////////////////////////////////////////////////////
// Public
////////////////////////////////////////////////////////////////////////////////

// DynamicValueToStringExpr is a DRY function to type assert
// a potentiall dynamic value into a string
// satisfying type
func DynamicValueToStringExpr(dynamicValue interface{}) string {
	var stringExpr string

	switch typedValue := dynamicValue.(type) {
	case string:
		stringExpr = typedValue
	default:
		panic(fmt.Sprintf("Unsupported dynamic value type: %+v", typedValue))
	}
	return stringExpr
}

// S3AllKeysArnForBucket returns a CloudFormation-compatible Arn expression
// (string or Ref) for all bucket keys (`/*`).  The bucket
// parameter may be either a string or an interface{} ("Ref: "myResource")
// value
func S3AllKeysArnForBucket(bucket interface{}) string {
	arnParts := []string{
		"arn:aws:s3:::",
		DynamicValueToStringExpr(bucket),
		"/*",
	}
	return gof.Join("", arnParts)
}

// S3ArnForBucket returns a CloudFormation-compatible Arn expression
// (string or Ref) suitable for template reference.  The bucket
// parameter may be either a string or an interface{} ("Ref: "myResource")
// value
func S3ArnForBucket(bucket interface{}) string {
	arnParts := []string{
		"arn:aws:s3:::",
		DynamicValueToStringExpr(bucket),
	}
	return gof.Join("", arnParts)
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
	additionalUserTemplateProperties map[string]interface{}) (string, error) {
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
	additionalUserTemplateProperties map[string]interface{}) (string, error) {
	converter := &templateConverter{
		templateReader:          templateData,
		additionalTemplateProps: additionalUserTemplateProperties,
		doQuote:                 true,
	}
	return converter.expandTemplate().parseData().results()
}

// StackEvents returns the slice of awsv2CF.StackEvents for the given stackID or stackName
func StackEvents(ctx context.Context,
	stackID string,
	eventFilterLowerBoundInclusive time.Time,
	awsConfig awsv2.Config) ([]awsv2CFTypes.StackEvent, error) {

	cfService := awsv2CF.NewFromConfig(awsConfig)
	var events []awsv2CFTypes.StackEvent

	nextToken := ""
	for {
		params := &awsv2CF.DescribeStackEventsInput{
			StackName: awsv2.String(stackID),
		}
		if len(nextToken) > 0 {
			params.NextToken = awsv2.String(nextToken)
		}

		resp, err := cfService.DescribeStackEvents(ctx, params)
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
	stackInfo           *awsv2CFTypes.Stack
}

// WaitForStackOperationComplete is a blocking, polling based call that
// periodically fetches the stackID set of events and uses the state value
// to determine if an operation is complete
func WaitForStackOperationComplete(ctx context.Context,
	stackID string,
	pollingMessage string,
	awsCloudFormation *awsv2CF.Client,
	logger *zerolog.Logger) (*WaitForStackOperationCompleteResult, error) {

	result := &WaitForStackOperationCompleteResult{}

	// Startup a spinner...
	charSetIndex := 39
	terminalType := os.Getenv("LC_TERMINAL")
	logger.Debug().
		Str("terminal", terminalType).
		Msg("Terminal type")

	if strings.Contains(terminalType, "iTerm") {
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
	describeStacksInput := &awsv2CF.DescribeStacksInput{
		StackName: awsv2.String(stackID),
	}
	cliSpinnerStarted := false
	startTime := time.Now()

	for waitComplete := false; !waitComplete; {
		// Startup the spinner if needed...
		deltaTime := time.Since(startTime)
		if !cliSpinnerStarted {
			cliSpinner.Start()
			defer cliSpinner.Stop()
			cliSpinnerStarted = true
		}
		spinnerText := fmt.Sprintf(" %s (elapsed: %s)",
			pollingMessage,
			deltaTime)
		cliSpinner.Suffix = spinnerText

		// Then sleep and figure out if things are done...
		sleepDuration := time.Duration(11+rand.Int31n(13)) * time.Second
		time.Sleep(sleepDuration)

		describeStacksOutput, err := awsCloudFormation.DescribeStacks(ctx, describeStacksInput)

		if nil != err {
			// TODO - add retry iff we're RateExceeded due to collective access
			return nil, err
		}
		if len(describeStacksOutput.Stacks) <= 0 {
			return nil, fmt.Errorf("failed to enumerate stack info: %v", *describeStacksInput.StackName)
		}

		result.stackInfo = &describeStacksOutput.Stacks[0]
		switch result.stackInfo.StackStatus {
		case awsv2CFTypes.StackStatusCreateComplete,
			awsv2CFTypes.StackStatusUpdateComplete:
			result.operationSuccessful = true
			waitComplete = true
		case
			// Include DeleteComplete as new provisions will automatically rollback
			awsv2CFTypes.StackStatusDeleteComplete,
			awsv2CFTypes.StackStatusCreateFailed,
			awsv2CFTypes.StackStatusDeleteFailed,
			awsv2CFTypes.StackStatusRollbackFailed,
			awsv2CFTypes.StackStatusRollbackComplete,
			awsv2CFTypes.StackStatusUpdateRollbackComplete:
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
func UploadTemplate(ctx context.Context,
	serviceName string,
	cfTemplate *gof.Template,
	s3Bucket string,
	s3KeyName string,
	awsConfig awsv2.Config,
	logger *zerolog.Logger) (string, error) {

	logger.Info().
		Str("Key", s3KeyName).
		Str("Bucket", s3Bucket).
		Msg("Uploading CloudFormation template")

	s3Client := awsv2S3.NewFromConfig(awsConfig)
	s3Uploader := awsv2S3Manager.NewUploader(s3Client)

	// Serialize the template and upload it
	cfTemplateJSON, err := json.Marshal(cfTemplate)
	if err != nil {
		return "", errors.Wrap(err, "Failed to Marshal CloudFormation template")
	}

	// Upload the actual CloudFormation template to S3 to maximize the template
	// size limit
	// Ref: http://docs.aws.amazon.com/AWSCloudFormation/latest/APIReference/API_CreateStack.html
	contentBody := string(cfTemplateJSON)
	uploadInput := &awsv2S3.PutObjectInput{
		Bucket:      &s3Bucket,
		Key:         &s3KeyName,
		ContentType: awsv2.String("application/json"),
		Body:        strings.NewReader(contentBody),
	}
	templateUploadResult, templateUploadResultErr := s3Uploader.Upload(ctx, uploadInput)
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
func StackExists(ctx context.Context,
	stackNameOrID string,
	awsConfig awsv2.Config,
	logger *zerolog.Logger) (bool, error) {
	cf := awsv2CF.NewFromConfig(awsConfig)

	describeStacksInput := &awsv2CF.DescribeStacksInput{
		StackName: awsv2.String(stackNameOrID),
	}
	describeStacksOutput, err := cf.DescribeStacks(ctx, describeStacksInput)

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
func CreateStackChangeSet(ctx context.Context,
	changeSetRequestName string,
	serviceName string,
	cfTemplate *gof.Template,
	templateURL string,
	stackParameters map[string]string,
	stackTags map[string]string,
	awsCloudFormation *awsv2CF.Client,
	logger *zerolog.Logger) (*awsv2CF.DescribeChangeSetOutput, error) {

	cloudFormationParameters := make([]awsv2CFTypes.Parameter,
		0,
		len(stackParameters))
	for eachKey, eachValue := range stackParameters {
		cloudFormationParameters = append(cloudFormationParameters, awsv2CFTypes.Parameter{
			ParameterKey:   awsv2.String(eachKey),
			ParameterValue: awsv2.String(eachValue),
		})
	}

	capabilities := stackCapabilities(cfTemplate)
	changeSetInput := &awsv2CF.CreateChangeSetInput{
		Capabilities:  capabilities,
		ChangeSetName: awsv2.String(changeSetRequestName),
		ClientToken:   awsv2.String(changeSetRequestName),
		Description:   awsv2.String(fmt.Sprintf("Change set for service: %s", serviceName)),
		StackName:     awsv2.String(serviceName),
		TemplateURL:   awsv2.String(templateURL),
		Parameters:    cloudFormationParameters,
	}
	if len(stackTags) != 0 {
		awsTags := []awsv2CFTypes.Tag{}
		for eachKey, eachValue := range stackTags {
			awsTags = append(awsTags,
				awsv2CFTypes.Tag{
					Key:   awsv2.String(eachKey),
					Value: awsv2.String(eachValue),
				})
		}
		changeSetInput.Tags = awsTags
	}
	_, changeSetError := awsCloudFormation.CreateChangeSet(ctx, changeSetInput)
	if nil != changeSetError {
		return nil, changeSetError
	}

	logger.Info().
		Str("StackName", serviceName).
		Msg("Issued CreateChangeSet request")

	describeChangeSetInput := awsv2CF.DescribeChangeSetInput{
		ChangeSetName: awsv2.String(changeSetRequestName),
		StackName:     awsv2.String(serviceName),
	}

	var describeChangeSetOutput *awsv2CF.DescribeChangeSetOutput

	// Loop, with a total timeout of 3 minutes
	startTime := time.Now()
	changeSetStabilized := false
	for !changeSetStabilized {
		sleepDuration := cloudformationPollingDelay()
		time.Sleep(sleepDuration)

		changeSetOutput, describeChangeSetError := awsCloudFormation.DescribeChangeSet(ctx, &describeChangeSetInput)

		if nil != describeChangeSetError {
			return nil, describeChangeSetError
		}
		describeChangeSetOutput = changeSetOutput
		// The current status of the change set, such as CREATE_IN_PROGRESS, CREATE_COMPLETE,
		// or FAILED.
		if nil != describeChangeSetOutput {
			switch describeChangeSetOutput.Status {
			case awsv2CFTypes.ChangeSetStatusCreateInProgress:
				// If this has taken more than 3 minutes, then that's an error
				elapsedTime := time.Since(startTime)
				if elapsedTime > cloudformationPollingTimeout {
					return nil, fmt.Errorf("failed to finalize ChangeSet within window: %s", elapsedTime.String())
				}
			case awsv2CFTypes.ChangeSetStatusCreateComplete:
				changeSetStabilized = true
			case awsv2CFTypes.ChangeSetStatusFailed:
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
		_, deleteChangeSetResultErr := DeleteChangeSet(ctx,
			serviceName,
			changeSetRequestName,
			awsCloudFormation)
		return nil, deleteChangeSetResultErr
	}
	return describeChangeSetOutput, nil
}

// DeleteChangeSet is a utility function that attempts to delete
// an existing CloudFormation change set, with a bit of retry
// logic in case of EC
func DeleteChangeSet(ctx context.Context,
	stackName string,
	changeSetRequestName string,
	awsCloudFormation *awsv2CF.Client) (*awsv2CF.DeleteChangeSetOutput, error) {

	// Delete request...
	deleteChangeSetInput := awsv2CF.DeleteChangeSetInput{
		ChangeSetName: awsv2.String(changeSetRequestName),
		StackName:     awsv2.String(stackName),
	}

	startTime := time.Now()
	for {
		elapsedTime := time.Since(startTime)

		deleteChangeSetResults, deleteChangeSetResultErr := awsCloudFormation.DeleteChangeSet(ctx, &deleteChangeSetInput)
		if nil == deleteChangeSetResultErr {
			return deleteChangeSetResults, nil
		} else if strings.Contains(deleteChangeSetResultErr.Error(), string(awsv2CFTypes.StackStatusCreateInProgress)) {
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
func ListStacks(ctx context.Context,
	awsConfig awsv2.Config,
	maxReturned int,
	stackFilters ...awsv2CFTypes.StackStatus) ([]awsv2CFTypes.StackSummary, error) {

	listStackInput := &awsv2CF.ListStacksInput{
		StackStatusFilter: []awsv2CFTypes.StackStatus{},
	}
	for _, eachFilter := range stackFilters {
		listStackInput.StackStatusFilter = append(listStackInput.StackStatusFilter, awsv2CFTypes.StackStatus(eachFilter))
	}
	cloudformationSvc := awsv2CF.NewFromConfig(awsConfig)
	accumulator := []awsv2CFTypes.StackSummary{}
	for {
		listResult, listResultErr := cloudformationSvc.ListStacks(ctx, listStackInput)
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
func ConvergeStackState(ctx context.Context,
	serviceName string,
	cfTemplate *gof.Template,
	templateURL string,
	stackParameters map[string]string,
	tags map[string]string,
	startTime time.Time,
	operationTimeout time.Duration,
	awsConfig awsv2.Config,
	outputsDividerChar string,
	dividerWidth int,
	logger *zerolog.Logger) (*awsv2CFTypes.Stack, error) {

	// Create the parameter values.
	logEntry := logger.Info()
	for eachKey, eachValue := range stackParameters {
		logEntry = logEntry.Str(fmt.Sprintf("Parameter: %s", eachKey), eachValue)
	}
	logEntry.Interface("Tags", tags).
		Str("Name", serviceName).
		Msg("Stack configuration")

	cloudformationSvc := awsv2CF.NewFromConfig(awsConfig)
	// Update the tags
	exists, existsErr := StackExists(ctx, serviceName, awsConfig, logger)
	if nil != existsErr {
		return nil, existsErr
	}
	stackID := ""
	if exists {
		updateErr := updateStackViaChangeSet(ctx,
			serviceName,
			cfTemplate,
			templateURL,
			stackParameters,
			tags,
			cloudformationSvc,
			logger)

		if nil != updateErr {
			return nil, updateErr
		}
		stackID = serviceName
	} else {
		var cloudFormationParameters []awsv2CFTypes.Parameter
		for eachKey, eachValue := range stackParameters {
			cloudFormationParameters = append(cloudFormationParameters, awsv2CFTypes.Parameter{
				ParameterKey:   awsv2.String(eachKey),
				ParameterValue: awsv2.String(eachValue),
			})
		}
		awsTags := []awsv2CFTypes.Tag{}
		if nil != tags {
			for eachKey, eachValue := range tags {
				awsTags = append(awsTags,
					awsv2CFTypes.Tag{
						Key:   awsv2.String(eachKey),
						Value: awsv2.String(eachValue),
					})
			}
		}

		// Create stack
		createStackInput := &awsv2CF.CreateStackInput{
			StackName:        awsv2.String(serviceName),
			TemplateURL:      awsv2.String(templateURL),
			TimeoutInMinutes: awsv2.Int32(int32(operationTimeout.Minutes())),
			OnFailure:        awsv2CFTypes.OnFailureDelete,
			Capabilities:     stackCapabilities(cfTemplate),
			Parameters:       cloudFormationParameters,
		}

		if len(awsTags) != 0 {
			createStackInput.Tags = awsTags
		}

		logger.Info().
			Interface("StackInput", createStackInput).
			Msg("Create stack input")

		createStackResponse, createStackResponseErr := cloudformationSvc.CreateStack(ctx, createStackInput)
		if nil != createStackResponseErr {
			return nil, createStackResponseErr
		}
		logger.Info().
			Str("StackID", *createStackResponse.StackId).
			Msg("Creating stack")
		stackID = *createStackResponse.StackId
	}

	// Wait for the operation to succeed
	pollingMessage := "Waiting for CloudFormation operation to complete"
	convergeResult, convergeErr := WaitForStackOperationComplete(ctx,
		stackID,
		pollingMessage,
		cloudformationSvc,
		logger)
	if nil != convergeErr {
		return nil, convergeErr
	}
	// Get the events and assemble them into either errors to output
	// or summary information
	resourceMetrics := make(map[string]*resourceProvisionMetrics)
	errorMessages := []string{}
	events, err := StackEvents(ctx, stackID, startTime, awsConfig)
	if nil != err {
		return nil, fmt.Errorf("failed to retrieve stack events: %s", err.Error())
	}

	for _, eachEvent := range events {
		switch eachEvent.ResourceStatus {
		case awsv2CFTypes.ResourceStatusCreateFailed,
			awsv2CFTypes.ResourceStatusDeleteFailed,
			awsv2CFTypes.ResourceStatusUpdateFailed:
			errMsg := fmt.Sprintf("\tError ensuring %s (%s): %s",
				*eachEvent.ResourceType,
				*eachEvent.LogicalResourceId,
				*eachEvent.ResourceStatusReason)
			// Only append if the resource failed because something else failed
			// and this resource was canceled.
			if !strings.Contains(errMsg, "cancelled") {
				errorMessages = append(errorMessages, errMsg)
			}
		case awsv2CFTypes.ResourceStatusCreateInProgress,
			awsv2CFTypes.ResourceStatusUpdateInProgress:
			existingMetric, existingMetricExists := resourceMetrics[*eachEvent.LogicalResourceId]
			if !existingMetricExists {
				existingMetric = &resourceProvisionMetrics{}
			}
			existingMetric.resourceType = *eachEvent.ResourceType
			existingMetric.logicalResourceID = *eachEvent.LogicalResourceId
			existingMetric.startTime = *eachEvent.Timestamp
			resourceMetrics[*eachEvent.LogicalResourceId] = existingMetric
		case awsv2CFTypes.ResourceStatusCreateComplete,
			awsv2CFTypes.ResourceStatusUpdateComplete:
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
				Str("Value", *eachOutput.OutputValue).
				Str("Description", *eachOutput.Description).
				Msgf("    %s", *eachOutput.OutputKey)
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
	awsConfig awsv2.Config) (string, error) {
	awsName, awsNameErr := platformAccountUserName(awsConfig)
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
