// +build !lambdabinary

package sparta

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	spartaAWS "github.com/mweagle/Sparta/aws"
	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	"github.com/sirupsen/logrus"
)

// Delete the provided serviceName.  Failing to delete a non-existent
// service is not considered an error.  Note that the delete does
func Delete(serviceName string, logger *logrus.Logger) error {
	session := spartaAWS.NewSession(logger)
	awsCloudFormation := cloudformation.New(session)

	exists, err := spartaCF.StackExists(serviceName, session, logger)
	if nil != err {
		return err
	}
	logger.WithFields(logrus.Fields{
		"Exists": exists,
		"Name":   serviceName,
	}).Info("Stack existence check")

	if exists {

		params := &cloudformation.DeleteStackInput{
			StackName: aws.String(serviceName),
		}
		resp, err := awsCloudFormation.DeleteStack(params)
		if nil != resp {
			logger.WithFields(logrus.Fields{
				"Response": resp,
			}).Info("Delete request submitted")
		}
		return err
	}
	logger.Info("Stack does not exist")
	return nil
}
