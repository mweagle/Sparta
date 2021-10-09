//go:build lambdabinary
// +build lambdabinary

package cloudformation

import (
	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
)

// https://blog.cloudflare.com/setting-go-variables-at-compile-time/

func platformUserName() string {
	return ""
}

func platformAccountUserName(awsConfig awsv2.Config) (string, error) {
	return "", nil
}
