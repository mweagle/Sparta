package resources

import (
	"context"
	"encoding/json"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2CodeCommit "github.com/aws/aws-sdk-go-v2/service/codecommit"
	awsv2CodeCommitTypes "github.com/aws/aws-sdk-go-v2/service/codecommit/types"

	gof "github.com/awslabs/goformation/v5/cloudformation"
	"github.com/rs/zerolog"
)

// CodeCommitLambdaEventSourceResourceRequest defines the request properties to configure
// SNS
type CodeCommitLambdaEventSourceResourceRequest struct {
	CustomResourceRequest
	LambdaTargetArn string
	RepositoryName  string
	TriggerName     string
	Events          []string `json:",omitempty"`
	Branches        []string `json:",omitempty"`
}

// CodeCommitLambdaEventSourceResource is a simple POC showing how to create custom resources
type CodeCommitLambdaEventSourceResource struct {
	gof.CustomResource
}

func (command CodeCommitLambdaEventSourceResource) updateRegistration(isTargetActive bool,
	awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	request := CodeCommitLambdaEventSourceResourceRequest{}
	unmarshalErr := json.Unmarshal(event.ResourceProperties, &request)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}
	logger.Info().
		Interface("Event", command).
		Msg("CodeCommit Custom Resource info")

	// We need the repo in here...
	codeCommitSvc := awsv2CodeCommit.NewFromConfig(awsConfig)

	// Get the current subscriptions...
	ccInput := &awsv2CodeCommit.GetRepositoryTriggersInput{
		RepositoryName: awsv2.String(request.RepositoryName),
	}
	triggers, triggersErr := codeCommitSvc.GetRepositoryTriggers(context.Background(), ccInput)
	if triggersErr != nil {
		return nil, triggersErr
	}

	// Find the lambda ARN for this function...
	putTriggers := make([]awsv2CodeCommitTypes.RepositoryTrigger, 0)
	var preexistingTrigger *awsv2CodeCommitTypes.RepositoryTrigger
	for _, eachTrigger := range triggers.Triggers {
		// Treat the preexisting one specially
		if *eachTrigger.DestinationArn == request.LambdaTargetArn {
			preexistingTrigger = &eachTrigger
		} else {
			putTriggers = append(putTriggers, eachTrigger)
		}
	}

	// Just log it...
	logger.Info().
		Str("RepositoryName", request.RepositoryName).
		Interface("Trigger", preexistingTrigger).
		Interface("LambdaArn", request.LambdaTargetArn).
		Msg("Current CodeCommit trigger status")

	reqBranches := make([]string, len(request.Branches))
	for idx, eachBranch := range request.Branches {
		reqBranches[idx] = eachBranch
	}
	reqEvents := make([]awsv2CodeCommitTypes.RepositoryTriggerEventEnum, len(request.Events))
	for idx, eachEvent := range request.Events {
		reqEvents[idx] = awsv2CodeCommitTypes.RepositoryTriggerEventEnum(eachEvent)
	}
	if len(reqEvents) <= 0 {
		logger.Info().Msg("No events found. Defaulting to `all`.")
		reqEvents = []awsv2CodeCommitTypes.RepositoryTriggerEventEnum{
			awsv2CodeCommitTypes.RepositoryTriggerEventEnumAll,
		}
	}
	if isTargetActive && preexistingTrigger == nil {
		// Add one...
		putTriggers = append(putTriggers, awsv2CodeCommitTypes.RepositoryTrigger{
			DestinationArn: awsv2.String(request.LambdaTargetArn),
			Name:           awsv2.String(request.TriggerName),
			Branches:       reqBranches,
			Events:         reqEvents,
		})
	} else if !isTargetActive {
		// It's already removed...
	} else if isTargetActive {
		putTriggers = append(putTriggers, *preexistingTrigger)
	}
	// Put it back...
	putTriggersInput := &awsv2CodeCommit.PutRepositoryTriggersInput{
		RepositoryName: awsv2.String(request.RepositoryName),
		Triggers:       putTriggers,
	}
	putTriggersResp, putTriggersRespErr := codeCommitSvc.PutRepositoryTriggers(context.Background(), putTriggersInput)
	// Just log it...
	logger.Info().
		Interface("Response", putTriggersResp).
		Interface("Error", putTriggersRespErr).
		Msg("CodeCommit PutRepositoryTriggers")

	return nil, putTriggersRespErr
}

// IAMPrivileges returns the IAM privs for this custom action
func (command *CodeCommitLambdaEventSourceResource) IAMPrivileges() []string {
	return []string{"codecommit:GetRepositoryTriggers",
		"codecommit:PutRepositoryTriggers"}
}

// Create implements the custom resource create operation
func (command CodeCommitLambdaEventSourceResource) Create(awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateRegistration(true, awsConfig, event, logger)
}

// Update implements the custom resource update operation
func (command CodeCommitLambdaEventSourceResource) Update(awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateRegistration(true, awsConfig, event, logger)
}

// Delete implements the custom resource delete operation
func (command CodeCommitLambdaEventSourceResource) Delete(awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateRegistration(false, awsConfig, event, logger)
}
