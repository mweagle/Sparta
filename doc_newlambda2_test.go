package sparta

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"net/http"
)

func lambdaHelloWorld(event *LambdaEvent, context *LambdaContext, w *http.ResponseWriter, logger *logrus.Logger) {
	fmt.Fprintf(*w, "Hello World!")
}

func ExampleNewLambda_iAMRoleDefinition() {
	roleDefinition := sparta.IAMRoleDefinition{}
	roleDefinition.Privileges = append(roleDefinition.Privileges, sparta.IAMRolePrivilege{
		Actions: []string{"s3:GetObject",
			"s3:PutObject"},
		Resource: "arn:aws:s3:::*",
	})
	helloWorldLambda := NewLambda(sparta.IAMRoleDefinition{}, lambdaHelloWorld, nil)
	if nil != helloWorldLambda {
		fmt.Printf("Failed to create new Lambda function")
	}
}
