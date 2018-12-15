package step

import (
	"math/rand"

	gocf "github.com/mweagle/go-cloudformation"
)

// SageMakerTag represents a tag for a SageMaker task
type SageMakerTag struct {
	Key   string          `json:",omitempty"`
	Value gocf.Stringable `json:",omitempty"`
}

// SageMakerTrainingJobParameters are the parameters for a SageMaker Training
// Job as defined by https://docs.aws.amazon.com/sagemaker/latest/dg/API_CreateTrainingJob.html
type SageMakerTrainingJobParameters struct {
	AlgorithmSpecification map[string]interface{} `json:",omitempty"`
	HyperParameters        map[string]interface{} `json:",omitempty"`
	InputDataConfig        []interface{}          `json:",omitempty"`
	OutputDataConfig       map[string]interface{} `json:",omitempty"`
	ResourceConfig         map[string]interface{} `json:",omitempty"`
	RoleArn                gocf.Stringable        `json:",omitempty"`
	StoppingCondition      interface{}            `json:",omitempty"`
	Tags                   []SageMakerTag         `json:",omitempty"`
	TrainingJobName        string                 `json:",omitempty"`
	VpcConfig              map[string]interface{} `json:",omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
/*
  _____         _      _
 |_   _| _ __ _(_)_ _ (_)_ _  __ _
   | || '_/ _` | | ' \| | ' \/ _` |
   |_||_| \__,_|_|_||_|_|_||_\__, |
                             |___/
*/
////////////////////////////////////////////////////////////////////////////////

// SageMakerTrainingJob represents bindings for
// https://docs.aws.amazon.com/step-functions/latest/dg/connectors-sqs.html
type SageMakerTrainingJob struct {
	BaseTask
	parameters SageMakerTrainingJobParameters
}

// MarshalJSON for custom marshalling, since this will be stringified and we need it
// to turn into a stringified
// Ref: https://docs.aws.amazon.com/step-functions/latest/dg/connectors-sqs.html
func (smtj *SageMakerTrainingJob) MarshalJSON() ([]byte, error) {
	return smtj.BaseTask.marshalMergedParams("arn:aws:states:::sagemaker:createTrainingJob.sync",
		&smtj.parameters)
}

// NewSageMakerTrainingJob returns an initialized SQSTaskState
func NewSageMakerTrainingJob(stateName string,
	parameters SageMakerTrainingJobParameters) *SageMakerTrainingJob {
	smtj := &SageMakerTrainingJob{
		BaseTask: BaseTask{
			baseInnerState: baseInnerState{
				name: stateName,
				id:   rand.Int63(),
			},
		},
		parameters: parameters,
	}
	return smtj
}

////////////////////////////////////////////////////////////////////////////////
/*
_____                  __
|_   _| _ __ _ _ _  ___/ _|___ _ _ _ __
	| || '_/ _` | ' \(_-<  _/ _ \ '_| '  \
	|_||_| \__,_|_||_/__/_| \___/_| |_|_|_|
*/
////////////////////////////////////////////////////////////////////////////////

// SageMakerTransformJobParameters are the parameters for a SageMaker Training
// Job as defined by https://docs.aws.amazon.com/sagemaker/latest/dg/API_CreateTransformJob.html
type SageMakerTransformJobParameters struct {
	BatchStrategy           string                 `json:",omitempty"`
	Environment             map[string]string      `json:",omitempty"`
	MaxConcurrentTransforms int                    `json:",omitempty"`
	MaxPayloadInMB          int                    `json:",omitempty"`
	ModelName               string                 `json:",omitempty"`
	Tags                    []SageMakerTag         `json:",omitempty"`
	TransformInput          map[string]interface{} `json:",omitempty"`
	TransformJobName        string                 `json:",omitempty"`
	TransformOutput         map[string]interface{} `json:",omitempty"`
	TransformResources      map[string]interface{} `json:",omitempty"`
}

// SageMakerTransformJob represents bindings for
// https://docs.aws.amazon.com/step-functions/latest/dg/connectors-sqs.html
type SageMakerTransformJob struct {
	BaseTask
	parameters SageMakerTransformJobParameters
}

// MarshalJSON for custom marshalling, since this will be stringified and we need it
// to turn into a stringified
// Ref: https://docs.aws.amazon.com/step-functions/latest/dg/connectors-sqs.html
func (smtj *SageMakerTransformJob) MarshalJSON() ([]byte, error) {
	return smtj.BaseTask.marshalMergedParams("arn:aws:states:::sagemaker:createTransformJob.sync",
		&smtj.parameters)
}

// NewSageMakerTransformJob returns an initialized SQSTaskState
func NewSageMakerTransformJob(stateName string,
	parameters SageMakerTransformJobParameters) *SageMakerTransformJob {
	sns := &SageMakerTransformJob{
		BaseTask: BaseTask{
			baseInnerState: baseInnerState{
				name: stateName,
				id:   rand.Int63(),
			},
		},
		parameters: parameters,
	}
	return sns
}
