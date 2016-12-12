// +build lambdabinary

package sparta

// Provides NOP implementations for functions that do not need to execute
// in the Lambda context

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"io"
	"syscall"
)

// Delete is not available in the AWS Lambda binary
func Delete(serviceName string, logger *logrus.Logger) error {
	logger.Error("Delete() not supported in AWS Lambda binary")
	return errors.New("Delete not supported for this binary")
}

// Provision is not available in the AWS Lambda binary
func Provision(noop bool,
	serviceName string,
	serviceDescription string,
	lambdaAWSInfos []*LambdaAWSInfo,
	api *API,
	site *S3Site,
	s3Bucket string,
	buildID string,
	codePipelineTrigger string,
	buildTags string,
	linkerFlags string,
	writer io.Writer,
	workflowHooks *WorkflowHooks,
	logger *logrus.Logger) error {
	logger.Error("Deploy() not supported in AWS Lambda binary")
	return errors.New("Deploy not supported for this binary")
}

// Describe is not available in the AWS Lambda binary
func Describe(serviceName string,
	serviceDescription string,
	lambdaAWSInfos []*LambdaAWSInfo,
	api *API,
	site *S3Site,
	s3BucketName string,
	buildTags string,
	linkerFlags string,
	outputWriter io.Writer,
	workflowHooks *WorkflowHooks,
	logger *logrus.Logger) error {
	logger.Error("Describe() not supported in AWS Lambda binary")
	return errors.New("Describe not supported for this binary")
}

// Explore is not available in the AWS Lambda binary
func Explore(lambdaAWSInfos []*LambdaAWSInfo,
	port int,
	logger *logrus.Logger) error {
	logger.Error("Explore() not supported in AWS Lambda binary")
	return errors.New("Explore not supported for this binary")
}

// Support Windows development, by only requiring `syscall` in the compiled
// linux binary.  THere is a NOP impl over in sparta_xplatbuild that doesn't
// include the lambdabinary flag
func platformKill(parentProcessPID int) {
	syscall.Kill(parentProcessPID, syscall.SIGUSR2)
}

// RegisterCodePipelineEnvironment is not available during lambda execution
func RegisterCodePipelineEnvironment(environmentName string, environmentVariables map[string]string) error {
	return nil
}
