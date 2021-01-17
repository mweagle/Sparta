package resources

import (
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/rs/zerolog"
)

// SESLambdaEventSourceResourceAction represents an SES rule action
// TODO - specialized types for Actions
type SESLambdaEventSourceResourceAction struct {
	ActionType       *gocf.StringExpr
	ActionProperties map[string]interface{}
}

func (action *SESLambdaEventSourceResourceAction) toReceiptAction(logger *zerolog.Logger) *ses.ReceiptAction {
	actionProperties := action.ActionProperties
	switch action.ActionType.Literal {
	case "LambdaAction":
		action := &ses.ReceiptAction{
			LambdaAction: &ses.LambdaAction{
				FunctionArn:    aws.String(actionProperties["FunctionArn"].(string)),
				InvocationType: aws.String("Event"),
			},
		}
		if val, exists := actionProperties["InvocationType"]; exists {
			action.LambdaAction.InvocationType = aws.String(val.(string))
		}
		if val, exists := actionProperties["TopicArn"]; exists {
			action.LambdaAction.TopicArn = aws.String(val.(string))
		}
		return action
	case "S3Action":
		action := &ses.ReceiptAction{
			S3Action: &ses.S3Action{
				BucketName: aws.String(actionProperties["BucketName"].(string)),
			},
		}
		if val, exists := actionProperties["KmsKeyArn"]; exists {
			action.S3Action.KmsKeyArn = aws.String(val.(string))
		}
		if val, exists := actionProperties["ObjectKeyPrefix"]; exists {
			action.S3Action.ObjectKeyPrefix = aws.String(val.(string))
		}
		if val, exists := actionProperties["TopicArn"]; exists {
			action.S3Action.TopicArn = aws.String(val.(string))
		}
		return action
	default:
		logger.Error().Msgf("No SESLmabdaEventSourceResourceAction marshaler found for action: %s", action.ActionType.Literal)
	}
	return nil
}

// SESLambdaEventSourceResourceRule stores settings necessary to configure an SES
// inbound rule
type SESLambdaEventSourceResourceRule struct {
	Name        *gocf.StringExpr
	Actions     []*SESLambdaEventSourceResourceAction
	ScanEnabled *gocf.BoolExpr `json:",omitempty"`
	Enabled     *gocf.BoolExpr `json:",omitempty"`
	Recipients  []*gocf.StringExpr
	TLSPolicy   *gocf.StringExpr `json:",omitempty"`
}

func ensureSESRuleSetName(ruleSetName string, svc *ses.SES, logger *zerolog.Logger) error {
	describeInput := &ses.DescribeReceiptRuleSetInput{
		RuleSetName: aws.String(ruleSetName),
	}
	var opError error
	describeRuleSet, describeRuleSetErr := svc.DescribeReceiptRuleSet(describeInput)
	if nil != describeRuleSetErr {
		if strings.Contains(describeRuleSetErr.Error(), "RuleSetDoesNotExist") {
			createRuleSet := &ses.CreateReceiptRuleSetInput{
				RuleSetName: aws.String(ruleSetName),
			}
			logger.Info().
				Interface("createRuleSet", createRuleSet).
				Msg("Creating Sparta SES Rule set")

			_, opError = svc.CreateReceiptRuleSet(createRuleSet)
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
	RuleSetName *gocf.StringExpr
	Rules       []*SESLambdaEventSourceResourceRule
}

// SESLambdaEventSourceResource handles configuring SES configuration
type SESLambdaEventSourceResource struct {
	gocf.CloudFormationCustomResource
	SESLambdaEventSourceResourceRequest
}

func (command SESLambdaEventSourceResource) updateSESRules(areRulesActive bool,
	session *session.Session,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {

	unmarshalErr := json.Unmarshal(event.ResourceProperties, &command)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	svc := ses.New(session)
	opError := ensureSESRuleSetName(command.RuleSetName.Literal, svc, logger)
	if nil == opError {
		for _, eachRule := range command.Rules {
			if areRulesActive {
				createReceiptRule := &ses.CreateReceiptRuleInput{
					RuleSetName: aws.String(command.RuleSetName.Literal),
					Rule: &ses.ReceiptRule{
						Name:        aws.String(eachRule.Name.Literal),
						Recipients:  make([]*string, 0),
						Actions:     make([]*ses.ReceiptAction, 0),
						ScanEnabled: aws.Bool(eachRule.ScanEnabled.Literal),
						TlsPolicy:   aws.String(eachRule.TLSPolicy.Literal),
						Enabled:     aws.Bool(eachRule.Enabled.Literal),
					},
				}
				for _, eachAction := range eachRule.Actions {
					createReceiptRule.Rule.Actions = append(createReceiptRule.Rule.Actions, eachAction.toReceiptAction(logger))
				}

				_, opError = svc.CreateReceiptRule(createReceiptRule)
			} else {
				// Delete them...
				deleteReceiptRule := &ses.DeleteReceiptRuleInput{
					RuleSetName: aws.String(command.RuleSetName.Literal),
					RuleName:    aws.String(eachRule.Name.Literal),
				}
				_, opError = svc.DeleteReceiptRule(deleteReceiptRule)
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
func (command SESLambdaEventSourceResource) Create(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateSESRules(true, awsSession, event, logger)
}

// Update implements the custom resource update operation
func (command SESLambdaEventSourceResource) Update(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateSESRules(true, awsSession, event, logger)
}

// Delete implements the custom resource delete operation
func (command SESLambdaEventSourceResource) Delete(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateSESRules(false, awsSession, event, logger)
}
