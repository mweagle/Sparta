package resources

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2SES "github.com/aws/aws-sdk-go-v2/service/ses"
	awsv2SESTypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	gof "github.com/awslabs/goformation/v5/cloudformation"
	"github.com/rs/zerolog"
)

// SESLambdaEventSourceResourceAction represents an SES rule action
// TODO - specialized types for Actions
type SESLambdaEventSourceResourceAction struct {
	ActionType       string
	ActionProperties map[string]interface{}
}

func (action *SESLambdaEventSourceResourceAction) toReceiptAction(logger *zerolog.Logger) *awsv2SESTypes.ReceiptAction {
	actionProperties := action.ActionProperties
	switch action.ActionType {
	case "LambdaAction":
		action := &awsv2SESTypes.ReceiptAction{
			LambdaAction: &awsv2SESTypes.LambdaAction{
				FunctionArn:    awsv2.String(actionProperties["FunctionArn"].(string)),
				InvocationType: awsv2SESTypes.InvocationTypeEvent,
			},
		}
		if val, exists := actionProperties["InvocationType"]; exists {
			action.LambdaAction.InvocationType = awsv2SESTypes.InvocationType(val.(string))
		}
		if val, exists := actionProperties["TopicArn"]; exists {
			action.LambdaAction.TopicArn = awsv2.String(val.(string))
		}
		return action
	case "S3Action":
		action := &awsv2SESTypes.ReceiptAction{
			S3Action: &awsv2SESTypes.S3Action{
				BucketName: awsv2.String(actionProperties["BucketName"].(string)),
			},
		}
		if val, exists := actionProperties["KmsKeyArn"]; exists {
			action.S3Action.KmsKeyArn = awsv2.String(val.(string))
		}
		if val, exists := actionProperties["ObjectKeyPrefix"]; exists {
			action.S3Action.ObjectKeyPrefix = awsv2.String(val.(string))
		}
		if val, exists := actionProperties["TopicArn"]; exists {
			action.S3Action.TopicArn = awsv2.String(val.(string))
		}
		return action
	default:
		logger.Error().Msgf("No SESLmabdaEventSourceResourceAction marshaler found for action: %s", action.ActionType)
	}
	return nil
}

func toBool(s string) bool {
	tVal, tValErr := strconv.ParseBool(s)
	return (tVal && tValErr == nil)
}

// SESLambdaEventSourceResourceRule stores settings necessary to configure an SES
// inbound rule. Boolean types are strings to workaround
// https://forums.aws.amazon.com/thread.jspa?threadID=302268
type SESLambdaEventSourceResourceRule struct {
	Name        string
	Actions     []*SESLambdaEventSourceResourceAction
	ScanEnabled string `json:",omitempty"`
	Enabled     string `json:",omitempty"`
	Recipients  []string
	TLSPolicy   string `json:",omitempty"`
}

func ensureSESRuleSetName(ruleSetName string, svc *awsv2SES.Client, logger *zerolog.Logger) error {
	describeInput := &awsv2SES.DescribeReceiptRuleSetInput{
		RuleSetName: awsv2.String(ruleSetName),
	}
	var opError error
	describeRuleSet, describeRuleSetErr := svc.DescribeReceiptRuleSet(context.Background(), describeInput)
	if nil != describeRuleSetErr {
		if strings.Contains(describeRuleSetErr.Error(), "RuleSetDoesNotExist") {
			createRuleSet := &awsv2SES.CreateReceiptRuleSetInput{
				RuleSetName: awsv2.String(ruleSetName),
			}
			logger.Info().
				Interface("createRuleSet", createRuleSet).
				Msg("Creating Sparta SES Rule set")

			_, opError = svc.CreateReceiptRuleSet(context.Background(), createRuleSet)
		}
	} else {
		logger.Info().
			Interface("describeRuleSet", describeRuleSet).
			Msg("SES Rule Set already exists")
	}
	return opError
}

// SESLambdaEventSourceResourceRequest defines the request properties to configure
// SES
type SESLambdaEventSourceResourceRequest struct {
	CustomResourceRequest
	RuleSetName string
	Rules       []*SESLambdaEventSourceResourceRule
}

// SESLambdaEventSourceResource handles configuring SES configuration
type SESLambdaEventSourceResource struct {
	gof.CustomResource
}

func (command SESLambdaEventSourceResource) updateSESRules(areRulesActive bool,
	awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {

	request := SESLambdaEventSourceResourceRequest{}
	unmarshalErr := json.Unmarshal(event.ResourceProperties, &request)
	if unmarshalErr != nil {
		logger.Warn().
			Interface("REQUEST", event.ResourceProperties).
			Msg("SES Request")
		return nil, unmarshalErr
	}

	svc := awsv2SES.NewFromConfig(awsConfig)
	opError := ensureSESRuleSetName(request.RuleSetName, svc, logger)
	if nil == opError {
		for _, eachRule := range request.Rules {
			if areRulesActive {
				createReceiptRule := &awsv2SES.CreateReceiptRuleInput{
					RuleSetName: awsv2.String(request.RuleSetName),
					Rule: &awsv2SESTypes.ReceiptRule{
						Name:        awsv2.String(eachRule.Name),
						Recipients:  make([]string, 0),
						Actions:     make([]awsv2SESTypes.ReceiptAction, 0),
						ScanEnabled: toBool(eachRule.ScanEnabled),
						TlsPolicy:   awsv2SESTypes.TlsPolicy(eachRule.TLSPolicy),
						Enabled:     toBool(eachRule.Enabled),
					},
				}
				for _, eachAction := range eachRule.Actions {
					receiptAction := eachAction.toReceiptAction(logger)
					if receiptAction != nil {
						createReceiptRule.Rule.Actions = append(createReceiptRule.Rule.Actions,
							*receiptAction)
					}
				}

				_, opError = svc.CreateReceiptRule(context.Background(), createReceiptRule)
			} else {
				// Delete them...
				deleteReceiptRule := &awsv2SES.DeleteReceiptRuleInput{
					RuleSetName: awsv2.String(request.RuleSetName),
					RuleName:    awsv2.String(eachRule.Name),
				}
				_, opError = svc.DeleteReceiptRule(context.Background(), deleteReceiptRule)
			}
			if nil != opError {
				return nil, opError
			}
		}
	}
	return nil, opError
}

// IAMPrivileges returns the IAM privs for this custom action
func (command *SESLambdaEventSourceResource) IAMPrivileges() []string {
	return []string{"ses:CreateReceiptRuleSet",
		"ses:CreateReceiptRule",
		"ses:DeleteReceiptRule",
		"ses:DeleteReceiptRuleSet",
		"ses:DescribeReceiptRuleSet"}
}

// Create implements the custom resource create operation
func (command SESLambdaEventSourceResource) Create(ctx context.Context, awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateSESRules(true, awsConfig, event, logger)
}

// Update implements the custom resource update operation
func (command SESLambdaEventSourceResource) Update(ctx context.Context, awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateSESRules(true, awsConfig, event, logger)
}

// Delete implements the custom resource delete operation
func (command SESLambdaEventSourceResource) Delete(ctx context.Context, awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateSESRules(false, awsConfig, event, logger)
}
