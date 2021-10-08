//go:build !lambdabinary
// +build !lambdabinary

package cloudformation

import (
	"context"
	"fmt"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2IAM "github.com/aws/aws-sdk-go-v2/service/iam"
)

func platformAccountUserName(awsConfig awsv2.Config) (string, error) {
	iamSvc := awsv2IAM.NewFromConfig(awsConfig)
	userInfo, userInfoErr := iamSvc.GetUser(context.Background(), &awsv2IAM.GetUserInput{})
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
