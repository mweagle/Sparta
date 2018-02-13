package sparta

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"text/template"

	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	cloudformationresources "github.com/mweagle/Sparta/aws/cloudformation/resources"
	spartaIAM "github.com/mweagle/Sparta/aws/iam"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	// ScratchDirectory is the cwd relative path component
	// where intermediate build artifacts are created
	ScratchDirectory = ".sparta"
)

// PushSourceConfigurationActions map stores common IAM Policy Actions for Lambda
// push-source configuration management.
// The configuration is handled by CustomResources inserted into the generated
// CloudFormation template.
var PushSourceConfigurationActions = struct {
	SNSLambdaEventSource            []string
	S3LambdaEventSource             []string
	SESLambdaEventSource            []string
	CloudWatchLogsLambdaEventSource []string
}{
	SNSLambdaEventSource: []string{"sns:ConfirmSubscription",
		"sns:GetTopicAttributes",
		"sns:ListSubscriptionsByTopic",
		"sns:Subscribe",
		"sns:Unsubscribe"},
	S3LambdaEventSource: []string{"s3:GetBucketLocation",
		"s3:GetBucketNotification",
		"s3:PutBucketNotification",
		"s3:GetBucketNotificationConfiguration",
		"s3:PutBucketNotificationConfiguration"},
	SESLambdaEventSource: []string{"ses:CreateReceiptRuleSet",
		"ses:CreateReceiptRule",
		"ses:DeleteReceiptRule",
		"ses:DeleteReceiptRuleSet",
		"ses:DescribeReceiptRuleSet"},
	CloudWatchLogsLambdaEventSource: []string{"logs:DescribeSubscriptionFilters",
		"logs:DeleteSubscriptionFilter",
		"logs:PutSubscriptionFilter",
	},
}

// This is a literal version of the DiscoveryInfo struct.
var discoveryData = `
{
	"ResourceID": "<< .TagLogicalResourceID >>",
	"Region": "{"Ref" : "AWS::Region"}",
	"StackID": "{"Ref" : "AWS::StackId"}",
	"StackName": "{"Ref" : "AWS::StackName"}",
	"Resources":{<<range $eachDepResource, $eachOutputString := .Resources>>
		"<< $eachDepResource >>" : << $eachOutputString >><< trailingComma >><<end>>
	}
}`

//
type discoveryDataTemplateData struct {
	TagLogicalResourceID string
	Resources            map[string]string
}

func runOSCommand(cmd *exec.Cmd, logger *logrus.Logger) error {
	logger.WithFields(logrus.Fields{
		"Arguments": cmd.Args,
		"Dir":       cmd.Dir,
		"Path":      cmd.Path,
		"Env":       cmd.Env,
	}).Debug("Running Command")
	outputWriter := logger.Writer()
	defer outputWriter.Close()
	cmd.Stdout = outputWriter
	cmd.Stderr = outputWriter
	return cmd.Run()
}

func lambdaFunctionEnvironment(userEnvMap map[string]*gocf.StringExpr,
	resourceID string,
	deps map[string]string,
	logger *logrus.Logger) (*gocf.LambdaFunctionEnvironment, error) {
	// Merge everything, add the deps
	envMap := make(map[string]interface{})
	for eachKey, eachValue := range userEnvMap {
		envMap[eachKey] = eachValue
	}
	discoveryInfo, discoveryInfoErr := discoveryInfoForResource(resourceID, deps)
	if discoveryInfoErr != nil {
		return nil, errors.Wrapf(discoveryInfoErr, "Failed to calculate dependency info")
	}
	envMap[envVarLogLevel] = logger.Level.String()
	envMap[envVarDiscoveryInformation] = discoveryInfo
	return &gocf.LambdaFunctionEnvironment{
		Variables: envMap,
	}, nil
}

func discoveryInfoForResource(resID string, deps map[string]string) (*gocf.StringExpr, error) {
	discoveryDataTemplateData := &discoveryDataTemplateData{
		TagLogicalResourceID: resID,
		Resources:            deps,
	}
	totalDeps := len(deps)
	var templateFuncMap = template.FuncMap{
		// The name "inc" is what the function will be called in the template text.
		"trailingComma": func() string {
			totalDeps--
			if totalDeps > 0 {
				return ","
			}
			return ""
		},
	}

	discoveryTemplate, discoveryTemplateErr := template.New("discoveryData").
		Delims("<<", ">>").
		Funcs(templateFuncMap).
		Parse(discoveryData)
	if nil != discoveryTemplateErr {
		return nil, discoveryTemplateErr
	}

	var templateResults bytes.Buffer
	evalResultErr := discoveryTemplate.Execute(&templateResults, discoveryDataTemplateData)
	if nil != evalResultErr {
		return nil, evalResultErr
	}
	templateReader := bytes.NewReader(templateResults.Bytes())
	templateExpr, templateExprErr := spartaCF.ConvertToTemplateExpression(templateReader, nil)
	if templateExprErr != nil {
		return nil, templateExprErr
	}
	return gocf.Base64(templateExpr), nil
}

func awsPrincipalToService(awsPrincipalName string) string {
	return strings.ToUpper(strings.SplitN(awsPrincipalName, ".:-", 2)[0])
}

// ensureCustomResourceHandler handles ensuring that the custom resource responsible
// for supporting the operation is actually part of this stack.
func ensureCustomResourceHandler(serviceName string,
	binaryName string,
	customResourceTypeName string,
	sourceArn *gocf.StringExpr,
	dependsOn []string,
	template *gocf.Template,
	S3Bucket string,
	S3Key string,
	logger *logrus.Logger) (string, error) {

	// Prefix
	prefixName := fmt.Sprintf("%s-Sparta-CFRes", serviceName)
	subscriberHandlerName := CloudFormationResourceName(prefixName, customResourceTypeName)

	//////////////////////////////////////////////////////////////////////////////
	// IAM Role definition
	iamResourceName, err := ensureIAMRoleForCustomResource(customResourceTypeName,
		sourceArn,
		template,
		logger)
	if nil != err {
		return "", errors.Wrapf(err,
			"Failed to ensure IAM Role for custom resource: %s",
			customResourceTypeName)
	}
	iamRoleRef := gocf.GetAtt(iamResourceName, "Arn")
	_, exists := template.Resources[subscriberHandlerName]
	if !exists {
		logger.WithFields(logrus.Fields{
			"Service": customResourceTypeName,
		}).Debug("Including Lambda CustomResource for AWS Service")

		configuratorDescription := customResourceDescription(serviceName,
			customResourceTypeName)

		//////////////////////////////////////////////////////////////////////////////
		// Custom Resource Lambda Handler

		// The handler name is the resource name & will be set in the
		// env block
		lambdaFunctionName := awsLambdaFunctionName(customResourceTypeName)

		// Don't forget the discovery info...
		lambdaEnv, lambdaEnvErr := lambdaFunctionEnvironment(nil,
			customResourceTypeName,
			nil,
			logger)
		if lambdaEnvErr != nil {
			return "", errors.Wrapf(lambdaEnvErr, "Failed to create environment for required custom resource")
		}
		customResourceHandlerDef := gocf.LambdaFunction{
			Code: &gocf.LambdaFunctionCode{
				S3Bucket: gocf.String(S3Bucket),
				S3Key:    gocf.String(S3Key),
			},
			Runtime:      gocf.String(GoLambdaVersion),
			Description:  gocf.String(configuratorDescription),
			Handler:      gocf.String(binaryName),
			Role:         iamRoleRef,
			Timeout:      gocf.Integer(30),
			FunctionName: lambdaFunctionName.String(),
			// DISPATCH INFORMATION
			Environment: lambdaEnv,
		}

		cfResource := template.AddResource(subscriberHandlerName, customResourceHandlerDef)
		if nil != dependsOn && (len(dependsOn) > 0) {
			cfResource.DependsOn = append(cfResource.DependsOn, dependsOn...)
		}
	}
	return subscriberHandlerName, nil
}

// ensureIAMRoleForCustomResource ensures that the single IAM::Role for a single
// AWS principal (eg, s3.*.*) exists, and includes statements for the given
// sourceArn.  Sparta uses a single IAM::Role for the CustomResource configuration
// lambda, which is the union of all Arns in the application.
func ensureIAMRoleForCustomResource(awsPrincipalName string,
	sourceArn *gocf.StringExpr,
	template *gocf.Template,
	logger *logrus.Logger) (string, error) {

	var principalActions []string
	switch awsPrincipalName {
	case cloudformationresources.SNSLambdaEventSource:
		principalActions = PushSourceConfigurationActions.SNSLambdaEventSource
	case cloudformationresources.S3LambdaEventSource:
		principalActions = PushSourceConfigurationActions.S3LambdaEventSource
	case cloudformationresources.SESLambdaEventSource:
		principalActions = PushSourceConfigurationActions.SESLambdaEventSource
	case cloudformationresources.CloudWatchLogsLambdaEventSource:
		principalActions = PushSourceConfigurationActions.CloudWatchLogsLambdaEventSource
	default:
		return "", errors.Errorf("Unsupported principal for IAM role creation: %s", awsPrincipalName)
	}

	// What's the stable IAMRoleName?
	resourceBaseName := fmt.Sprintf("CFResIAMRole%s", awsPrincipalToService(awsPrincipalName))
	stableRoleName := CloudFormationResourceName(resourceBaseName, awsPrincipalName)

	// Ensure it exists, then check to see if this Source ARN is already specified...
	// Checking equality with Stringable?

	// Create a new Role
	var existingIAMRole *gocf.IAMRole
	existingResource, exists := template.Resources[stableRoleName]
	logger.WithFields(logrus.Fields{
		"PrincipalActions": principalActions,
		"SourceArn":        sourceArn,
	}).Debug("Ensuring IAM Role results")

	if !exists {
		// Insert the IAM role here.  We'll walk the policies data in the next section
		// to make sure that the sourceARN we have is in the list
		statements := CommonIAMStatements.Core

		iamPolicyList := gocf.IAMRolePolicyList{}
		iamPolicyList = append(iamPolicyList,
			gocf.IAMRolePolicy{
				PolicyDocument: ArbitraryJSONObject{
					"Version":   "2012-10-17",
					"Statement": statements,
				},
				PolicyName: gocf.String(fmt.Sprintf("%sPolicy", stableRoleName)),
			},
		)

		existingIAMRole = &gocf.IAMRole{
			AssumeRolePolicyDocument: AssumePolicyDocument,
			Policies:                 &iamPolicyList,
		}
		template.AddResource(stableRoleName, existingIAMRole)

		// Create a new IAM Role resource
		logger.WithFields(logrus.Fields{
			"RoleName": stableRoleName,
		}).Debug("Inserting IAM Role")
	} else {
		existingIAMRole = existingResource.Properties.(*gocf.IAMRole)
	}

	// Walk the existing statements
	if nil != existingIAMRole.Policies {
		for _, eachPolicy := range *existingIAMRole.Policies {
			policyDoc := eachPolicy.PolicyDocument.(ArbitraryJSONObject)
			statements := policyDoc["Statement"]
			for _, eachStatement := range statements.([]spartaIAM.PolicyStatement) {
				if sourceArn.String() == eachStatement.Resource.String() {

					logger.WithFields(logrus.Fields{
						"RoleName":  stableRoleName,
						"SourceArn": sourceArn.String(),
					}).Debug("SourceArn already exists for IAM Policy")
					return stableRoleName, nil
				}
			}
		}

		logger.WithFields(logrus.Fields{
			"RoleName": stableRoleName,
			"Action":   principalActions,
			"Resource": sourceArn,
		}).Debug("Inserting Actions for configuration ARN")

		// Add this statement to the first policy, iff the actions are non-empty
		if len(principalActions) > 0 {
			rootPolicy := (*existingIAMRole.Policies)[0]
			rootPolicyDoc := rootPolicy.PolicyDocument.(ArbitraryJSONObject)
			rootPolicyStatements := rootPolicyDoc["Statement"].([]spartaIAM.PolicyStatement)
			rootPolicyDoc["Statement"] = append(rootPolicyStatements, spartaIAM.PolicyStatement{
				Effect:   "Allow",
				Action:   principalActions,
				Resource: sourceArn,
			})
		}

		return stableRoleName, nil
	}
	return "", errors.Errorf("Unable to find Policies entry for IAM role: %s", stableRoleName)
}

func systemGoVersion(logger *logrus.Logger) (string, error) {
	runtimeVersion := runtime.Version()
	// Get the golang version from the output:
	// Matts-MBP:Sparta mweagle$ go version
	// go version go1.8.1 darwin/amd64
	golangVersionRE := regexp.MustCompile(`go(\d+\.\d+(\.\d+)?)`)
	matches := golangVersionRE.FindStringSubmatch(runtimeVersion)
	if len(matches) > 2 {
		return matches[1], nil
	}
	logger.WithFields(logrus.Fields{
		"Output": runtimeVersion,
	}).Warn("Unable to find Golang version using RegExp - using current version")
	return runtimeVersion, nil
}
