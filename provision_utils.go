package sparta

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	gocf "github.com/mweagle/go-cloudformation"

	"github.com/Sirupsen/logrus"
)

const salt = "213EA743-A98F-499D-8FEF-B87015FE13E7"

// PushSourceConfigurationActions map stores common IAM Policy Actions for Lambda
// push-source configuration management.
// The configuration is handled by CustomResources inserted into the generated
// CloudFormation template.
var PushSourceConfigurationActions = map[string][]string{
	"s3.amazonaws.com": {"s3:GetBucketLocation",
		"s3:GetBucketNotification",
		"s3:PutBucketNotification",
		"s3:GetBucketNotificationConfiguration",
		"s3:PutBucketNotificationConfiguration"},
	"sns.amazonaws.com": {"sns:ConfirmSubscription",
		"sns:GetTopicAttributes",
		"sns:Subscribe",
		"sns:Unsubscribe"},
	"ses.amazonaws.com": {"ses:CreateReceiptRuleSet",
		"ses:CreateReceiptRule",
		"ses:DeleteReceiptRule",
		"ses:DeleteReceiptRuleSet",
		"ses:DescribeReceiptRuleSet"},
	"apigateway.amazonaws.com": {"apigateway:*",
		"lambda:AddPermission",
		"lambda:RemovePermission",
		"lambda:GetPolicy"},
}

func nodeJSHandlerName(jsBaseFilename string) string {
	return fmt.Sprintf("index.%sConfiguration", jsBaseFilename)
}

func awsPrincipalToService(awsPrincipalName string) string {
	return strings.ToUpper(strings.SplitN(awsPrincipalName, ".", 2)[0])
}

func ensureConfiguratorLambdaResource(awsPrincipalName string,
	sourceArn interface{},
	dependsOn []string,
	template *gocf.Template,
	S3Bucket string,
	S3Key string,
	logger *logrus.Logger) (string, error) {

	// AWS service basename
	awsServiceName := awsPrincipalToService(awsPrincipalName)
	configuratorExportName := strings.ToLower(awsServiceName)

	// Create a unique name that we can use for the configuration info
	keyName, err := json.Marshal(ArbitraryJSONObject{
		"Principal": awsPrincipalName,
		"Arn":       sourceArn,
		"DependsOn": dependsOn,
	})
	if err != nil {
		logger.Error("Failed to create configurator resource name: ", err.Error())
		return "", err
	}
	subscriberHandlerName := CloudFormationResourceName(fmt.Sprintf("%sSubscriber",
		awsServiceName), string(keyName))

	//////////////////////////////////////////////////////////////////////////////
	// IAM Role definition
	principalActions, exists := PushSourceConfigurationActions[awsPrincipalName]
	if !exists {
		return "", fmt.Errorf("Unsupported principal for IAM role creation: %s", awsPrincipalName)
	}
	// Create a Role that enables this resource management
	iamResourceName, err := ensureIAMRoleResource(principalActions, sourceArn, template, logger)
	if nil != err {
		return "", err
	}
	iamRoleRef := gocf.GetAtt(iamResourceName, "Arn")
	_, exists = template.Resources[subscriberHandlerName]
	if !exists {
		logger.WithFields(logrus.Fields{
			"Service": awsServiceName,
		}).Debug("Including Lambda CustomResource for AWS Service")

		configuratorDescription := fmt.Sprintf("Sparta managed Lambda CustomResource to configure %s service",
			awsServiceName)

		//////////////////////////////////////////////////////////////////////////////
		// Custom Resource Lambda Handler
		// NOTE: This brittle function name has an analog in ./resources/index.js b/c the
		// AWS Lamba execution treats the entire ZIP file as a module.  So all module exports
		// need to be forwarded through the module's index.js file.
		handlerName := nodeJSHandlerName(configuratorExportName)
		logger.Debug("Lambda Configuration handler: ", handlerName)

		customResourceHandlerDef := gocf.LambdaFunction{
			Code: &gocf.LambdaFunctionCode{
				S3Bucket: gocf.String(S3Bucket),
				S3Key:    gocf.String(S3Key),
			},
			Description: gocf.String(configuratorDescription),
			Handler:     gocf.String(handlerName),
			Role:        iamRoleRef,
			Runtime:     gocf.String("nodejs"),
			Timeout:     gocf.Integer(30),
		}
		cfResource := template.AddResource(subscriberHandlerName, customResourceHandlerDef)
		if nil != dependsOn && (len(dependsOn) > 0) {
			cfResource.DependsOn = append(cfResource.DependsOn, dependsOn...)
		}

	}
	return subscriberHandlerName, nil
}

func ensureIAMRoleResource(principalActions []string,
	sourceArn interface{},
	template *gocf.Template,
	logger *logrus.Logger) (string, error) {

	// Create a new Role
	hash := sha1.New()
	hash.Write([]byte(fmt.Sprintf("%v%s", sourceArn, salt)))
	roleName := fmt.Sprintf("ConfigIAMRole%s", hex.EncodeToString(hash.Sum(nil)))

	existingResource, exists := template.Resources[roleName]
	logger.WithFields(logrus.Fields{
		"PrincipalActions": principalActions,
		"SourceArn":        sourceArn,
		"Exists":           exists,
	}).Debug("Ensuring IAM Role results")

	// If it exists, make sure these permissions are enabled on it...
	if nil != existingResource {
		statementExists := false
		iamRole, ok := existingResource.Properties.(*gocf.IAMRole)
		if !ok {
			return "", errors.New("Failed to convert to IAM::Role type")
		}
		if nil != iamRole.Policies {
			for _, eachPolicy := range *iamRole.Policies {
				policyDoc := eachPolicy.PolicyDocument.(ArbitraryJSONObject)
				statements := policyDoc["Statement"]
				for _, eachStatement := range statements.([]ArbitraryJSONObject) {
					if eachStatement["Resource"] == sourceArn {
						statementExists = true
					}
				}
			}
		}

		if !statementExists {
			if nil == iamRole.Policies {
				iamRole.Policies = &gocf.IAMPoliciesList{}
			}
			// Just push a new one...
			// Policies *IAMPoliciesList
			// IAMPoliciesList []IAMPolicies
			updatedPolicies := append(*iamRole.Policies, gocf.IAMPolicies{
				PolicyDocument: ArbitraryJSONObject{
					"Statement": []ArbitraryJSONObject{
						{
							"Effect":   "Allow",
							"Action":   principalActions,
							"Resource": sourceArn,
						},
					},
				},
			})
			iamRole.Policies = &updatedPolicies
		}
		logger.Debug("Using prexisting IAM Role: " + roleName)
		return roleName, nil
	}

	// Create a new IAM Role resource
	logger.WithFields(logrus.Fields{
		"RoleName": roleName,
		"Actions":  principalActions,
	}).Debug("Inserting IAM Role")

	// Provision a new one and add it...
	statements := CommonIAMStatements["core"]
	statements = append(statements, ArbitraryJSONObject{
		"Effect":   "Allow",
		"Action":   principalActions,
		"Resource": sourceArn,
	})

	iamPolicy := gocf.IAMRole{
		AssumeRolePolicyDocument: AssumePolicyDocument,
		Policies: &gocf.IAMPoliciesList{
			gocf.IAMPolicies{
				PolicyDocument: ArbitraryJSONObject{
					"Version":   "2012-10-17",
					"Statement": statements,
				},
				PolicyName: gocf.String(CloudFormationResourceName("Config", fmt.Sprintf("%v", sourceArn))),
			},
		},
	}
	template.AddResource(roleName, iamPolicy)
	return roleName, nil
}
