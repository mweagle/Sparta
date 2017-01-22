package cloudformationresources

import (
	"strings"

	gocf "github.com/crewjam/go-cloudformation"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

// SESLambdaEventSourceResourceAction represents an SES rule action
// TODO - specialized types for Actions
type SESLambdaEventSourceResourceAction struct {
	ActionType       *gocf.StringExpr
	ActionProperties map[string]interface{}
}

func (action *SESLambdaEventSourceResourceAction) toReceiptAction(logger *logrus.Logger) *ses.ReceiptAction {
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
		logger.Error("No SESLmabdaEventSourceResourceAction marshaler found for action: " + action.ActionType.Literal)
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

func ensureSESRuleSetName(ruleSetName string, svc *ses.SES, logger *logrus.Logger) error {
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
			logger.WithFields(logrus.Fields{
				"createRuleSet": createRuleSet,
			}).Info("Creating Sparta SES Rule set")
			_, opError = svc.CreateReceiptRuleSet(createRuleSet)
		}
	} else {
		logger.WithFields(logrus.Fields{
			"describeRuleSet": describeRuleSet,
		}).Info("Sparta SES Rule Set already exists")
	}
	return opError
}

// SESLambdaEventSourceResource handles configuring SES configuration
type SESLambdaEventSourceResource struct {
	GoAWSCustomResource
	RuleSetName *gocf.StringExpr
	Rules       []*SESLambdaEventSourceResourceRule
}

func (command SESLambdaEventSourceResource) updateSESRules(areRulesActive bool,
	session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {

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

func (command SESLambdaEventSourceResource) create(session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {
	return command.updateSESRules(true, session, logger)
}

func (command SESLambdaEventSourceResource) update(session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {
	return command.updateSESRules(true, session, logger)
}

func (command SESLambdaEventSourceResource) delete(session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {
	return command.updateSESRules(false, session, logger)
}
