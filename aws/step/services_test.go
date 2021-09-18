package step

import (
	"testing"

	gofecs "github.com/awslabs/goformation/v5/cloudformation/ecs"

	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
)

func TestFargateSNSServices(t *testing.T) {
	// Make the states
	fargateParams := FargateTaskParameters{
		Cluster:        "arn:aws:ecs:us-west-2:123123123123:cluster/StepFunctionsSample-ContainerTaskManagement08e32647-5862-4f61-a2a7-3443a2ef857d-ECSCluster-ZWJK3EFZ9T1H",
		TaskDefinition: "arn:aws:ecs:us-west-2:123123123123:task-definition/StepFunctionsSample-ContainerTaskManagement08e32647-5862-4f61-a2a7-3443a2ef857d-ECSTaskDefinition-UFPUM96E8JOQ:1",
		NetworkConfiguration: &FargateNetworkConfiguration{
			AWSVPCConfiguration: &gofecs.Service_AwsVpcConfiguration{
				Subnets: []string{
					"subnet-057bfcb4a52343473",
					"subnet-0f25a21f1251ecce5",
				},
				AssignPublicIp: "ENABLED",
			},
		},
	}
	fargateState := NewFargateTaskState("Run Fargate Task", fargateParams)

	snsSuccessParams := SNSTaskParameters{
		Message:  "AWS Fargate Task started by Step Functions succeeded 42",
		TopicArn: "arn:aws:sns:us-west-2:123123123123:StepFunctionsSample-ContainerTaskManagement08e32647-5862-4f61-a2a7-3443a2ef857d-SNSTopic-E8U58ADXVXRL",
	}
	snsSuccessState := NewSNSTaskState("Notify Success", snsSuccessParams)
	fargateState.Next(snsSuccessState)

	snsFailParams := SNSTaskParameters{
		Message:  "AWS Fargate Task started by Step Functions failed",
		TopicArn: "arn:aws:sns:us-west-2:123123123123:StepFunctionsSample-ContainerTaskManagement08e32647-5862-4f61-a2a7-3443a2ef857d-SNSTopic-E8U58ADXVXRL",
	}
	snsFailState := NewSNSTaskState("Notify Failure", snsFailParams)
	fargateState.WithCatchers(NewTaskCatch(
		snsFailState,
		StatesAll,
	))

	// Startup the machine.
	stateMachineName := spartaCF.UserScopedStackName("TestStepServicesMachine")
	startMachine := NewStateMachine(stateMachineName, fargateState)
	testStepProvision(t,
		nil,
		startMachine)
}
