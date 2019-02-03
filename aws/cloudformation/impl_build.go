// +build !lambdabinary

package cloudformation

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

func platformAccountUserName(awsSession *session.Session) (string, error) {
	iamSvc := iam.New(awsSession)
	userInfo, userInfoErr := iamSvc.GetUser(&iam.GetUserInput{})
	if userInfoErr != nil {
		return "", userInfoErr
	}
	userName := ""
	if userInfo.User.UserName != nil {
		userName = *userInfo.User.UserName
	}
	if len(userName) <= 0 && userInfo.User.UserId != nil {
		userName = *userInfo.User.UserId
	}
	if len(userName) <= 0 {
		return "", fmt.Errorf("failed to find valid user identifier from AWS")
	}
	return userName, nil
}
