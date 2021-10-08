//go:build lambdabinary
// +build lambdabinary

package cloudformation

import "github.com/aws/aws-sdk-go-v2/aws/session"

// https://blog.cloudflare.com/setting-go-variables-at-compile-time/

func platformUserName() string {
	return ""
}

func platformAccountUserName(awsConfig aws.Config) (string, error) {
	return "", nil
}
