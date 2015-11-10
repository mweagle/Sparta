// +build !lambdabinary

package sparta

import (
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// Delete the provided serviceName.  Failing to delete a non-existent
// service is not considered an error.  Note that the delete does
func Delete(serviceName string, logger *logrus.Logger) error {
	session := awsSession(logger)
	awsCloudFormation := cloudformation.New(session)

	exists, err := stackExists(serviceName, awsCloudFormation, logger)
	if nil != err {
		return err
	}
	if exists {
		logger.Info("Stack exists: ", serviceName)
		params := &cloudformation.DeleteStackInput{
			StackName: aws.String(serviceName),
		}
		resp, err := awsCloudFormation.DeleteStack(params)
		if nil != resp {
			logger.Info("Stack delete issued: ", resp)
		}
		return err
	}

	logger.Info("Stack does not exist: ", serviceName)
	return nil
}
