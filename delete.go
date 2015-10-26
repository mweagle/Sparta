// Copyright (c) 2015 Matt Weagle <mweagle@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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
	exists, err := stackExists(serviceName, logger)
	if nil != err {
		return err
	}
	if exists {
		logger.Info("Stack exists: ", serviceName)
		params := &cloudformation.DeleteStackInput{
			StackName: aws.String(serviceName),
		}
		awsCloudFormation := cloudformation.New(awsConfig())
		resp, err := awsCloudFormation.DeleteStack(params)
		if nil != resp {
			logger.Info("Stack delete issued: ", resp)
		}
		return err
	} else {
		logger.Info("Stack does not exist: ", serviceName)
	}
	return nil
}
