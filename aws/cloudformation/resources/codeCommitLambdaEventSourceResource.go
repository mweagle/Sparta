package resources

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codecommit"
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
	session *session.Session,
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
	codeCommitSvc := codecommit.New(session)

	// Get the current subscriptions...
	ccInput := &codecommit.GetRepositoryTriggersInput{
		RepositoryName: aws.String(request.RepositoryName),
	}
	triggers, triggersErr := codeCommitSvc.GetRepositoryTriggers(ccInput)
	if triggersErr != nil {
		return nil, triggersErr
	}

	// Find the lambda ARN for this function...
	putTriggers := make([]*codecommit.RepositoryTrigger, 0)
	var preexistingTrigger *codecommit.RepositoryTrigger
	for _, eachTrigger := range triggers.Triggers {
		// Treat the preexisting one specially
		if *eachTrigger.DestinationArn == request.LambdaTargetArn {
			preexistingTrigger = eachTrigger
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

	reqBranches := make([]*string, len(request.Branches))
	for idx, eachBranch := range request.Branches {
		reqBranches[idx] = aws.String(eachBranch)
	}
	reqEvents := make([]*string, len(request.Events))
	for idx, eachEvent := range request.Events {
		reqEvents[idx] = aws.String(eachEvent)
	}
	if len(reqEvents) <= 0 {
		logger.Info().Msg("No events found. Defaulting to `all`.")
		reqEvents = []*string{
			aws.String("all"),
		}
	}
	if isTargetActive && preexistingTrigger == nil {
		// Add one...
		putTriggers = append(putTriggers, &codecommit.RepositoryTrigger{
			DestinationArn: aws.String(request.LambdaTargetArn),
			Name:           aws.String(request.TriggerName),
			Branches:       reqBranches,
			Events:         reqEvents,
		})
	} else if !isTargetActive {
		// It's already removed...
	} else if isTargetActive && preexistingTrigger != nil {
		putTriggers = append(putTriggers, preexistingTrigger)
	}
	// Put it back...
	putTriggersInput := &codecommit.PutRepositoryTriggersInput{
		RepositoryName: aws.String(request.RepositoryName),
		Triggers:       putTriggers,
	}
	putTriggersResp, putTriggersRespErr := codeCommitSvc.PutRepositoryTriggers(putTriggersInput)
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
func (command CodeCommitLambdaEventSourceResource) Create(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateRegistration(true, awsSession, event, logger)
}

// Update implements the custom resource update operation
func (command CodeCommitLambdaEventSourceResource) Update(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateRegistration(true, awsSession, event, logger)
}

// Delete implements the custom resource delete operation
func (command CodeCommitLambdaEventSourceResource) Delete(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateRegistration(false, awsSession, event, logger)
}
