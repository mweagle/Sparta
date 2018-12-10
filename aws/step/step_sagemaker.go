package step

import (
	"math/rand"

	gocf "github.com/mweagle/go-cloudformation"
)

// SageMakerTag represents a tag for a SageMaker task
type SageMakerTag struct {
	Key   string
	Value gocf.Stringable
}

// SageMakerTrainingJobParameters are the parameters for a SageMaker Training
// Job as defined by https://docs.aws.amazon.com/sagemaker/latest/dg/API_CreateTrainingJob.html
type SageMakerTrainingJobParameters struct {
	AlgorithmSpecification map[string]interface{}
	HyperParameters        map[string]interface{}
	InputDataConfig        []interface{}
	OutputDataConfig       map[string]interface{}
	ResourceConfig         map[string]interface{}
	RoleArn                gocf.Stringable
	StoppingCondition      interface{}
	Tags                   []SageMakerTag
	TrainingJobName        string
	VpcConfig              map[string]interface{}
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
	additionalParams := smtj.BaseTask.additionalParams()
	additionalParams["Resource"] = "arn:aws:states:::sagemaker:createTrainingJob.sync"
	parameterMap := map[string]interface{}{}
	if smtj.parameters.AlgorithmSpecification != nil {
		parameterMap["AlgorithmSpecification"] = smtj.parameters.AlgorithmSpecification
	}
	if smtj.parameters.HyperParameters != nil {
		parameterMap["HyperParameters"] = smtj.parameters.HyperParameters
	}
	if smtj.parameters.InputDataConfig != nil {
		parameterMap["InputDataConfig"] = smtj.parameters.InputDataConfig
	}
	if smtj.parameters.OutputDataConfig != nil {
		parameterMap["OutputDataConfig"] = smtj.parameters.OutputDataConfig
	}
	if smtj.parameters.ResourceConfig != nil {
		parameterMap["ResourceConfig"] = smtj.parameters.ResourceConfig
	}
	if smtj.parameters.RoleArn != nil {
		parameterMap["RoleArn"] = smtj.parameters.RoleArn
	}
	if smtj.parameters.StoppingCondition != nil {
		parameterMap["StoppingCondition"] = smtj.parameters.StoppingCondition
	}
	if smtj.parameters.Tags != nil {
		parameterMap["Tags"] = smtj.parameters.Tags
	}
	if smtj.parameters.TrainingJobName != "" {
		parameterMap["TrainingJobName"] = smtj.parameters.TrainingJobName
	}
	if smtj.parameters.VpcConfig != nil {
		parameterMap["VpcConfig"] = smtj.parameters.VpcConfig
	}
	additionalParams["Parameters"] = parameterMap
	return smtj.marshalStateJSON("Task", additionalParams)
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
	BatchStrategy           string
	Environment             map[string]string
	MaxConcurrentTransforms int
	MaxPayloadInMB          int
	ModelName               string
	Tags                    []SageMakerTag
	TransformInput          map[string]interface{}
	TransformJobName        string
	TransformOutput         map[string]interface{}
	TransformResources      map[string]interface{}
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

	additionalParams := smtj.BaseTask.additionalParams()
	additionalParams["Resource"] = "arn:aws:states:::sagemaker:createTransformJob.sync"
	parameterMap := map[string]interface{}{}

	if smtj.parameters.BatchStrategy != "" {
		parameterMap["BatchStrategy"] = smtj.parameters.BatchStrategy
	}
	if smtj.parameters.Environment != nil {
		parameterMap["Environment"] = smtj.parameters.Environment
	}
	if smtj.parameters.MaxConcurrentTransforms != 0 {
		parameterMap["MaxConcurrentTransforms"] = smtj.parameters.MaxConcurrentTransforms
	}
	if smtj.parameters.MaxPayloadInMB != 0 {
		parameterMap["MaxPayloadInMB"] = smtj.parameters.MaxPayloadInMB
	}
	if smtj.parameters.ModelName != "" {
		parameterMap["ModelName"] = smtj.parameters.ModelName
	}
	if smtj.parameters.Tags != nil {
		parameterMap["Tags"] = smtj.parameters.Tags
	}
	if smtj.parameters.TransformInput != nil {
		parameterMap["TransformInput"] = smtj.parameters.TransformInput
	}
	if smtj.parameters.TransformJobName != "" {
		parameterMap["TransformJobName"] = smtj.parameters.TransformJobName
	}
	if smtj.parameters.TransformOutput != nil {
		parameterMap["TransformOutput"] = smtj.parameters.TransformOutput
	}
	if smtj.parameters.TransformResources != nil {
		parameterMap["TransformResources"] = smtj.parameters.TransformResources
	}
	additionalParams["Parameters"] = parameterMap
	return smtj.marshalStateJSON("Task", additionalParams)
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
