package step

import (
	"math/rand"

	gocf "github.com/mweagle/go-cloudformation"
)

// FargateNetworkConfiguration contains the AWSVPCConfiguration
// information
type FargateNetworkConfiguration struct {
	AWSVPCConfiguration *gocf.ECSServiceAwsVPCConfiguration `json:"AwsvpcConfiguration"`
}

// FargateTaskParameters contains the information
// for a Fargate task
type FargateTaskParameters struct {
	Cluster              gocf.Stringable
	Group                string
	LaunchType           string
	NetworkConfiguration *FargateNetworkConfiguration
	Overrides            map[string]interface{}
	PlacementConstraints []map[string]string
	PlacementStrategy    []map[string]string
	PlatformVersion      string
	TaskDefinition       gocf.Stringable
}

// FargateTaskState represents a FargateTask
type FargateTaskState struct {
	BaseTask
	parameters FargateTaskParameters
}

// MarshalJSON for custom marshalling, since this will be stringified and we need it
// to turn into a stringified Ref:
// https://docs.aws.amazon.com/step-functions/latest/dg/connectors-ecs.html
func (fts *FargateTaskState) MarshalJSON() ([]byte, error) {
	additionalParams := fts.BaseTask.additionalParams()
	additionalParams["Resource"] = "arn:aws:states:::ecs:runTask.sync"

	parameterMap := map[string]interface{}{}
	if fts.parameters.Cluster != nil {
		parameterMap["Cluster"] = fts.parameters.Cluster
	}
	if fts.parameters.Group != "" {
		parameterMap["Group"] = fts.parameters.Group
	}
	if fts.parameters.LaunchType != "" {
		parameterMap["LaunchType"] = fts.parameters.LaunchType
	}
	if fts.parameters.NetworkConfiguration != nil {
		parameterMap["NetworkConfiguration"] = map[string]interface{}{
			"AwsvpcConfiguration": fts.parameters.NetworkConfiguration.AWSVPCConfiguration,
		}
	}
	if fts.parameters.Overrides != nil {
		parameterMap["Overrides"] = fts.parameters.Overrides
	}

	if fts.parameters.PlacementConstraints != nil {
		parameterMap["PlacementConstraints"] = fts.parameters.PlacementConstraints
	}

	if fts.parameters.PlacementStrategy != nil {
		parameterMap["PlacementStrategy"] = fts.parameters.PlacementStrategy
	}

	if fts.parameters.PlatformVersion != "" {
		parameterMap["PlatformVersion"] = fts.parameters.PlatformVersion
	}

	if fts.parameters.TaskDefinition != nil {
		parameterMap["TaskDefinition"] = fts.parameters.TaskDefinition
	}
	additionalParams["Parameters"] = parameterMap
	return fts.marshalStateJSON("Task", additionalParams)
}

// NewFargateTaskState returns an initialized FargateTaskState
func NewFargateTaskState(stateName string, parameters FargateTaskParameters) *FargateTaskState {
	ft := &FargateTaskState{
		BaseTask: BaseTask{
			baseInnerState: baseInnerState{
				name: stateName,
				id:   rand.Int63(),
			},
		},
		parameters: parameters,
	}
	return ft
}
