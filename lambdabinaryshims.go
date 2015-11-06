// +build lambdabinary

package sparta

// Provides NOP implementations for functions that do not need to execute
// in the Lambda context

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"io"
)

func Delete(serviceName string, logger *logrus.Logger) error {
	logger.Error("Delete() not supported in AWS Lambda binary")
	return errors.New("Delete not supported for this binary")
}

func Provision(serviceName string, serviceDescription string, lambdaAWSInfos []*LambdaAWSInfo, s3Bucket string, logger *logrus.Logger) error {
	logger.Error("Deploy() not supported in AWS Lambda binary")
	return errors.New("Deploy not supported for this binary")

}
func Describe(serviceName string, serviceDescription string, lambdaAWSInfos []*LambdaAWSInfo, outputWriter io.Writer, logger *logrus.Logger) error {
	logger.Error("Describe() not supported in AWS Lambda binary")
	return errors.New("Describe not supported for this binary")
}

func Explore(serviceName string, logger *logrus.Logger) error {
	logger.Error("Explore() not supported in AWS Lambda binary")
	return errors.New("Explore not supported for this binary")
}
