package step

import (
	"math/rand"

	gocf "github.com/mweagle/go-cloudformation"
)

// FargateNetworkConfiguration contains the AWSVPCConfiguration
// information
type FargateNetworkConfiguration struct {
	AWSVPCConfiguration *gocf.ECSServiceAwsVPCConfiguration `json:"AwsvpcConfiguration,omitempty"`
}

// FargateTaskParameters contains the information
// for a Fargate task
type FargateTaskParameters struct {
	Cluster              gocf.Stringable              `json:",omitempty"`
	Group                string                       `json:",omitempty"`
	LaunchType           string                       `json:",omitempty"`
	NetworkConfiguration *FargateNetworkConfiguration `json:",omitempty"`
	Overrides            map[string]interface{}       `json:",omitempty"`
	PlacementConstraints []map[string]string          `json:",omitempty"`
	PlacementStrategy    []map[string]string          `json:",omitempty"`
	PlatformVersion      string                       `json:",omitempty"`
	TaskDefinition       gocf.Stringable              `json:",omitempty"`
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
	return fts.BaseTask.marshalMergedParams("arn:aws:states:::ecs:runTask.sync",
		&fts.parameters)
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
