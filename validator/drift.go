package validator

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	sparta "github.com/mweagle/Sparta"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// DriftDetector is a detector that ensures that the service hasn't
// experienced configuration drift prior to being overwritten by a new provisioning
// step.
func DriftDetector(errorOnDrift bool) sparta.ServiceValidationHookHandler {

	driftDetector := func(context map[string]interface{},
		serviceName string,
		template *gocf.Template,
		S3Bucket string,
		S3Key string,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *logrus.Logger) error {
		// Create a cloudformation service.
		cfSvc := cloudformation.New(awsSession)
		detectStackDrift, detectStackDriftErr := cfSvc.DetectStackDrift(&cloudformation.DetectStackDriftInput{
			StackName: aws.String(serviceName),
		})
		if detectStackDriftErr != nil {
			// If it doesn't exist, then no worries...
			if strings.Contains(detectStackDriftErr.Error(), "does not exist") {
				return nil
			}
			return errors.Wrapf(detectStackDriftErr, "attempting to determine stack drift")
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
				logger.WithField("error", driftStatusErr).Warn("Failed to check Stack Drift")
			}
			if driftStatus != nil {
				switch *driftStatus.DetectionStatus {
				case "DETECTION_COMPLETE":
					detectionComplete = true
				default:
					logger.WithField("Status", *driftStatus.DetectionStatus).
						Info("Waiting for drift detection to complete")
					time.Sleep(11 * time.Second)
				}
			}
		}
		if !detectionComplete {
			return errors.Errorf("Stack drift detection did not complete in time")
		}

		golangFuncName := func(logicalResourceID string) string {
			templateRes, templateResExists := template.Resources[logicalResourceID]
			if !templateResExists {
				return ""
			}
			metadata := templateRes.Metadata
			if len(metadata) <= 0 {
				metadata = make(map[string]interface{}, 0)
			}
			golangFunc, golangFuncExists := metadata["golangFunc"]
			if !golangFuncExists {
				return ""
			}
			switch typedFunc := golangFunc.(type) {
			case string:
				return typedFunc
			default:
				return fmt.Sprintf("%#v", typedFunc)
			}
		}

		// Log the drifts
		logDrifts := func(stackResourceDrifts []*cloudformation.StackResourceDrift) {
			for _, eachDrift := range stackResourceDrifts {
				if len(eachDrift.PropertyDifferences) != 0 {
					for _, eachDiff := range eachDrift.PropertyDifferences {
						entry := logger.WithFields(logrus.Fields{
							"Resource":       *eachDrift.LogicalResourceId,
							"Actual":         *eachDiff.ActualValue,
							"Expected":       *eachDiff.ExpectedValue,
							"Relation":       *eachDiff.DifferenceType,
							"PropertyPath":   *eachDiff.PropertyPath,
							"LambdaFuncName": golangFuncName(*eachDrift.LogicalResourceId),
						})
						if errorOnDrift {
							entry.Error("Stack drift detected")
						} else {
							entry.Warn("Stack drift detected")
						}
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
				return errors.Wrapf(driftResultsErr, "attempting to describe stack drift")
			}
			stackResourceDrifts = append(stackResourceDrifts, driftResults.StackResourceDrifts...)
			if driftResults.NextToken == nil {
				break
			}
			loopCounter++
			// If there is more than 10 (1k total) something is seriously wrong...
			if loopCounter >= 10 {
				logDrifts(stackResourceDrifts)
				return errors.Errorf("Exceeded maximum number of Stack resource drifts: %d", len(stackResourceDrifts))
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
			return nil
		}
		return errors.Errorf("stack %s operation prevented due to stack drift", serviceName)
	}
	return sparta.ServiceValidationHookFunc(driftDetector)
}
