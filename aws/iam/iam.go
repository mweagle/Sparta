package iam

import (
	iamtypes "github.com/mweagle/Sparta/aws/iam/builder/types"
)

// PolicyStatement represents an entry in an IAM policy document
type PolicyStatement struct {
	Effect    string
	Action    []string
	Resource  string                 `json:",omitempty"`
	Principal *iamtypes.IAMPrincipal `json:",omitempty"`
	Condition interface{}            `json:",omitempty"`
}

// AssumeRolePolicyDocumentForServicePrincipal returns the document
// for the given service principal
func AssumeRolePolicyDocumentForServicePrincipal(principal string) interface{} {
	return map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []interface{}{
			map[string]interface{}{
				"Effect": "Allow",
				"Action": "sts:AssumeRole",
				"Principal": map[string]interface{}{
					"Service": principal,
				},
			},
		},
	}
}
