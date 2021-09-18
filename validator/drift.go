package validator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	gof "github.com/awslabs/goformation/v5/cloudformation"
	goflambda "github.com/awslabs/goformation/v5/cloudformation/lambda"
	sparta "github.com/mweagle/Sparta"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// DriftDetector is a detector that ensures that the service hasn't
// experienced configuration drift prior to being overwritten by a new provisioning
// step.
func DriftDetector(errorOnDrift bool) sparta.ServiceValidationHookHandler {

	driftDetector := func(ctx context.Context,
		serviceName string,
		template *gof.Template,
		lambdaFunctionCode *goflambda.Function_Code,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *zerolog.Logger) (context.Context, error) {
		// Create a cloudformation service.
		cfSvc := cloudformation.New(awsSession)
		detectStackDrift, detectStackDriftErr := cfSvc.DetectStackDrift(&cloudformation.DetectStackDriftInput{
			StackName: aws.String(serviceName),
		})
		if detectStackDriftErr != nil {
			// If it doesn't exist, then no worries...
			if strings.Contains(detectStackDriftErr.Error(), "does not exist") {
				return ctx, nil
			}
			return ctx, errors.Wrapf(detectStackDriftErr, "attempting to determine stack drift")
		}

		// Poll until it's done...
		describeDriftDetectionStatus := &cloudformation.DescribeStackDriftDetectionStatusInput{
			StackDriftDetectionId: detectStackDrift.StackDriftDetectionId,
		}
		detectionComplete := false

		// Put a limit on the detection
		for i := 0; i <= 30 && !detectionComplete; i++ {
			driftStatus, driftStatusErr := cfSvc.DescribeStackDriftDetectionStatus(describeDriftDetectionStatus)
			if driftStatusErr != nil {
				logger.Warn().
					Err(driftStatusErr).
					Msg("Failed to check Stack Drift")
			}
			if driftStatus != nil {
				switch *driftStatus.DetectionStatus {
				case "DETECTION_COMPLETE":
					detectionComplete = true
				default:
					logger.Info().
						Str("Status", *driftStatus.DetectionStatus).
						Msg("Waiting for drift detection to complete")

					time.Sleep(11 * time.Second)
				}
			}
		}
		if !detectionComplete {
			return ctx, errors.Errorf("Stack drift detection did not complete in time")
		}

		golangFuncName := func(logicalResourceID string) string {
			templateRes, templateResExists := template.Resources[logicalResourceID]
			if !templateResExists {
				return ""
			}
			typedRes, typedResOk := templateRes.(*goflambda.Function)
			funcName := fmt.Sprintf("ResourceID: %s", logicalResourceID)
			if typedResOk && typedRes.AWSCloudFormationMetadata != nil {
				funcName = fmt.Sprintf("%#v", typedRes.AWSCloudFormationMetadata["golangFunc"])
			}
			return funcName
		}

		// Log the drifts
		logDrifts := func(stackResourceDrifts []*cloudformation.StackResourceDrift) {
			for _, eachDrift := range stackResourceDrifts {
				if len(eachDrift.PropertyDifferences) != 0 {
					for _, eachDiff := range eachDrift.PropertyDifferences {
						var loggerEntry *zerolog.Event
						if errorOnDrift {
							loggerEntry = logger.Error()
						} else {
							loggerEntry = logger.Warn()
						}

						loggerEntry.
							Str("Resource", *eachDrift.LogicalResourceId).
							Str("Actual", *eachDiff.ActualValue).
							Str("Expected", *eachDiff.ExpectedValue).
							Str("Relation", *eachDiff.DifferenceType).
							Str("PropertyPath", *eachDiff.PropertyPath).
							Str("LambdaFuncName", golangFuncName(*eachDrift.LogicalResourceId)).
							Msg("Stack drift detected")
					}
				}
			}
		}

		// Utility function to fetch all the drifts
		stackResourceDrifts := make([]*cloudformation.StackResourceDrift, 0)
		input := &cloudformation.DescribeStackResourceDriftsInput{
			MaxResults: aws.Int64(100),
			StackName:  aws.String(serviceName),
		}
		// There can't be more than 200 resources in the template
		// https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/cloudformation-limits.html
		loopCounter := 0
		for {
			driftResults, driftResultsErr := cfSvc.DescribeStackResourceDrifts(input)
			if driftResultsErr != nil {
				return ctx, errors.Wrapf(driftResultsErr, "attempting to describe stack drift")
			}
			stackResourceDrifts = append(stackResourceDrifts, driftResults.StackResourceDrifts...)
			if driftResults.NextToken == nil {
				break
			}
			loopCounter++
			// If there is more than 10 (1k total) something is seriously wrong...
			if loopCounter >= 10 {
				logDrifts(stackResourceDrifts)
				return ctx, errors.Errorf("Exceeded maximum number of Stack resource drifts: %d", len(stackResourceDrifts))
			}

			input = &cloudformation.DescribeStackResourceDriftsInput{
				MaxResults: aws.Int64(100),
				StackName:  aws.String(serviceName),
				NextToken:  driftResults.NextToken,
			}
		}

		// Log them
		logDrifts(stackResourceDrifts)
		if len(stackResourceDrifts) == 0 || !errorOnDrift {
			return ctx, nil
		}
		return ctx, errors.Errorf("stack %s operation prevented due to stack drift", serviceName)
	}
	return sparta.ServiceValidationHookFunc(driftDetector)
}
