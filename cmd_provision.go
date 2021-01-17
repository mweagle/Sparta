package sparta

import (
	"bytes"
	"fmt"
	"reflect"
	"text/template"

	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	cfCustomResources "github.com/mweagle/Sparta/aws/cloudformation/resources"
	spartaIAM "github.com/mweagle/Sparta/aws/iam"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const (
	// ScratchDirectory is the cwd relative path component
	// where intermediate build artifacts are created
	ScratchDirectory = ".sparta"
	// EnvVarCustomResourceTypeName is the environment variable
	// name that stores the CustomResource TypeName that should be
	// instantiated
	EnvVarCustomResourceTypeName = "SPARTA_CUSTOM_RESOURCE_TYPE"
)

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

func lambdaFunctionEnvironment(userEnvMap map[string]*gocf.StringExpr,
	resourceID string,
	deps map[string]string,
	logger *zerolog.Logger) (*gocf.LambdaFunctionEnvironment, error) {
	// Merge everything, add the deps
	envMap := make(map[string]interface{})
	for eachKey, eachValue := range userEnvMap {
		envMap[eachKey] = eachValue
	}
	discoveryInfo, discoveryInfoErr := discoveryInfoForResource(resourceID, deps)
	if discoveryInfoErr != nil {
		return nil, errors.Wrapf(discoveryInfoErr, "Failed to calculate dependency info")
	}
	envMap[envVarLogLevel] = logger.GetLevel().String()
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
		// The name "trailingComma" is what the function will be called in the template text.
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

// EnsureCustomResourceHandler handles ensuring that the custom resource responsible
// for supporting the operation is actually part of this stack. The returned
// string value is the CloudFormation resource name that implements this
// resource. The customResourceCloudFormationTypeName must have already
// been registered with gocf and implement the resources.CustomResourceCommand
// interface
func EnsureCustomResourceHandler(serviceName string,
	customResourceCloudFormationTypeName string,
	sourceArn *gocf.StringExpr,
	dependsOn []string,
	template *gocf.Template,
	lambdaFunctionCode *gocf.LambdaFunctionCode,
	logger *zerolog.Logger) (string, error) {

	// Ok, we need a way to round trip this type as the AWS lambda function name.
	// The problem with this is that the full CustomResource::Type value isn't an
	// AWS Lambda friendly name. We want to do this so that in the AWS lambda handler
	// we can attempt to instantiate a new CustomAction resource, typecast it to a
	// CustomResourceCommand type and then apply the workflow. Doing this means
	// we can decouple the lookup logic for custom resource...

	resource := gocf.NewResourceByType(customResourceCloudFormationTypeName)
	if resource == nil {
		return "", errors.Errorf("Unable to create custom resource handler of type: %v", customResourceCloudFormationTypeName)
	}
	command, commandOk := resource.(cfCustomResources.CustomResourceCommand)
	if !commandOk {
		return "", errors.Errorf("Cannot type assert resource type %s to CustomResourceCommand", customResourceCloudFormationTypeName)
	}

	// Prefix
	commandType := reflect.TypeOf(command)
	customResourceTypeName := fmt.Sprintf("%T", command)
	prefixName := fmt.Sprintf("%s-CFRes", serviceName)
	subscriberHandlerName := CloudFormationResourceName(prefixName, customResourceTypeName)

	//////////////////////////////////////////////////////////////////////////////
	// IAM Role definition
	iamResourceName, err := ensureIAMRoleForCustomResource(command,
		sourceArn,
		template,
		logger)
	if nil != err {
		return "", errors.Wrapf(err,
			"Failed to ensure IAM Role for custom resource: %T",
			command)
	}
	iamRoleRef := gocf.GetAtt(iamResourceName, "Arn")
	_, exists := template.Resources[subscriberHandlerName]
	if exists {
		return subscriberHandlerName, nil
	}

	// Encode the resourceType...
	configuratorDescription := customResourceDescription(serviceName, customResourceTypeName)

	//////////////////////////////////////////////////////////////////////////////
	// Custom Resource Lambda Handler
	// Insert it into the template resources...
	logger.Info().
		Str("CloudFormationResourceType", customResourceCloudFormationTypeName).
		Str("Resource", customResourceTypeName).
		Str("TypeOf", commandType.String()).
		Msg("Including Lambda CustomResource")

	// Don't forget the discovery info...
	userDispatchMap := map[string]*gocf.StringExpr{
		EnvVarCustomResourceTypeName: gocf.String(customResourceCloudFormationTypeName),
	}
	lambdaEnv, lambdaEnvErr := lambdaFunctionEnvironment(userDispatchMap,
		customResourceTypeName,
		nil,
		logger)
	if lambdaEnvErr != nil {
		return "", errors.Wrapf(lambdaEnvErr, "Failed to create environment for required custom resource")
	}
	// Add the special key that's the custom resource type name
	customResourceHandlerDef := gocf.LambdaFunction{
		Code:        lambdaFunctionCode,
		Runtime:     gocf.String(string(Go1LambdaRuntime)),
		Description: gocf.String(configuratorDescription),
		Handler:     gocf.String(SpartaBinaryName),
		Role:        iamRoleRef,
		Timeout:     gocf.Integer(30),
		// Let AWS assign a name here...
		//		FunctionName: lambdaFunctionName.String(),
		// DISPATCH INFORMATION
		Environment: lambdaEnv,
	}
	if lambdaFunctionCode.ImageURI != nil {
		customResourceHandlerDef.PackageType = gocf.String("Image")
	} else {
		customResourceHandlerDef.Runtime = gocf.String(string(Go1LambdaRuntime))
		customResourceHandlerDef.Handler = gocf.String(SpartaBinaryName)
	}

	cfResource := template.AddResource(subscriberHandlerName, customResourceHandlerDef)
	if nil != dependsOn && (len(dependsOn) > 0) {
		cfResource.DependsOn = append(cfResource.DependsOn, dependsOn...)
	}
	return subscriberHandlerName, nil
}

// ensureIAMRoleForCustomResource ensures that the single IAM::Role for a single
// AWS principal (eg, s3.*.*) exists, and includes statements for the given
// sourceArn.  Sparta uses a single IAM::Role for the CustomResource configuration
// lambda, which is the union of all Arns in the application.
func ensureIAMRoleForCustomResource(command cfCustomResources.CustomResourceCommand,
	sourceArn *gocf.StringExpr,
	template *gocf.Template,
	logger *zerolog.Logger) (string, error) {

	// What's the stable IAMRoleName?
	commandName := fmt.Sprintf("%T", command)
	resourceBaseName := fmt.Sprintf("CFResIAMRole%s", commandName)
	stableRoleName := CloudFormationResourceName(resourceBaseName, resourceBaseName)

	// Is it a privileged command?
	var privileges []string
	privilegedCommand, privilegedCommandOk := command.(cfCustomResources.CustomResourcePrivilegedCommand)
	if privilegedCommandOk {
		privileges = privilegedCommand.IAMPrivileges()
	}

	// Ensure it exists, then check to see if this Source ARN is already specified...
	// Checking equality with Stringable?

	// Create a new Role
	var existingIAMRole *gocf.IAMRole
	existingResource, exists := template.Resources[stableRoleName]
	logger.Debug().
		Interface("PrincipalActions", privileges).
		Interface("SourceArn", sourceArn).
		Msg("Ensuring IAM Role results")

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
		logger.Debug().
			Str("RoleName", stableRoleName).
			Msg("Inserting IAM Role")
	} else {
		existingIAMRole = existingResource.Properties.(*gocf.IAMRole)
	}

	// ARNs are only required if there are non-empty privileges associated
	// with the command
	if sourceArn == nil {
		if len(privileges) != 0 {
			return "", errors.Errorf("CustomResource %s requires a SourceARN to apply it's %d principle actions",
				commandName,
				len(privileges))
		}
		return stableRoleName, nil
	}
	// Walk the existing statements
	if nil != existingIAMRole.Policies {
		for _, eachPolicy := range *existingIAMRole.Policies {
			policyDoc := eachPolicy.PolicyDocument.(ArbitraryJSONObject)
			statements := policyDoc["Statement"]
			for _, eachStatement := range statements.([]spartaIAM.PolicyStatement) {
				if sourceArn.String() == eachStatement.Resource.String() {

					logger.Debug().
						Str("RoleName", stableRoleName).
						Interface("SourceArn", sourceArn.String()).
						Msg("SourceArn already exists for IAM Policy")
					return stableRoleName, nil
				}
			}
		}

		logger.Debug().
			Str("RoleName", stableRoleName).
			Interface("Action", privileges).
			Interface("Resource", sourceArn).
			Msg("Inserting Actions for configuration ARN")

		// Add this statement to the first policy, iff the actions are non-empty
		if len(privileges) > 0 {
			rootPolicy := (*existingIAMRole.Policies)[0]
			rootPolicyDoc := rootPolicy.PolicyDocument.(ArbitraryJSONObject)
			rootPolicyStatements := rootPolicyDoc["Statement"].([]spartaIAM.PolicyStatement)
			rootPolicyDoc["Statement"] = append(rootPolicyStatements,
				spartaIAM.PolicyStatement{
					Effect:   "Allow",
					Action:   privileges,
					Resource: sourceArn,
				})
		}
		return stableRoleName, nil
	}
	return "", errors.Errorf("Unable to find Policies entry for IAM role: %s", stableRoleName)
}
