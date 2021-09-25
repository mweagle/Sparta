//go:build lambdabinary
// +build lambdabinary

package cloudformation

import "github.com/aws/aws-sdk-go/aws/session"

// https://blog.cloudflare.com/setting-go-variables-at-compile-time/

func platformUserName() string {
	return ""
}

func platformAccountUserName(awsSession *session.Session) (string, error) {
	return "", nil
}
