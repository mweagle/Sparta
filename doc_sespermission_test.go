package sparta

import (
	"context"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

func sesLambdaProcessor(ctx context.Context,
	props map[string]interface{}) (map[string]interface{}, error) {
	lambdaCtx, _ := lambdacontext.FromContext(ctx)
	Logger().Info().
		Str("RequestID", lambdaCtx.AwsRequestID).
		Interface("Properties", props).
		Msg("Lambda event")
	return props, nil
}

func ExampleSESPermission_messageBody() {
	var lambdaFunctions []*LambdaAWSInfo
	// Define the IAM role
	roleDefinition := IAMRoleDefinition{}
	sesLambda, _ := NewAWSLambda(LambdaName(sesLambdaProcessor),
		sesLambdaProcessor,
		roleDefinition)

	// Add a Permission s.t. the Lambda function is automatically invoked
	// in response to inbound email
	lambdaSESPermission := SESPermission{
		BasePermission: BasePermission{
			SourceArn: "*",
		},
		InvocationType: "Event",
	}
	// Store the message body in a newly provisioned S3 bucket
	bodyStorage, _ := lambdaSESPermission.NewMessageBodyStorageResource("MessageBody")
	lambdaSESPermission.MessageBodyStorage = bodyStorage

	// Add some custom ReceiptRules.
	lambdaSESPermission.ReceiptRules = append(lambdaSESPermission.ReceiptRules,
		ReceiptRule{
			Name:       "Default",
			Recipients: []string{},
			TLSPolicy:  "Optional",
		})
	sesLambda.Permissions = append(sesLambda.Permissions, lambdaSESPermission)

	lambdaFunctions = append(lambdaFunctions, sesLambda)
	Main("SESLambdaApp", "Registers for SES events and saves the MessageBody", lambdaFunctions, nil, nil)
}

func ExampleSESPermission_headersOnly() {
	var lambdaFunctions []*LambdaAWSInfo
	// Define the IAM role
	roleDefinition := IAMRoleDefinition{}
	sesLambda, _ := NewAWSLambda(LambdaName(sesLambdaProcessor),
		sesLambdaProcessor,
		roleDefinition)

	// Add a Permission s.t. the Lambda function is automatically invoked
	// in response to inbound email
	lambdaSESPermission := SESPermission{
		BasePermission: BasePermission{
			SourceArn: "*",
		},
		InvocationType: "Event",
	}
	// Add some custom ReceiptRules.  Rules will be inserted (evaluated) in their
	// array rank order.
	lambdaSESPermission.ReceiptRules = make([]ReceiptRule, 0)
	lambdaSESPermission.ReceiptRules = append(lambdaSESPermission.ReceiptRules,
		ReceiptRule{
			Name:       "Special",
			Recipients: []string{"somebody@mydomain.io"},
			TLSPolicy:  "Optional",
		})

	lambdaSESPermission.ReceiptRules = append(lambdaSESPermission.ReceiptRules,
		ReceiptRule{
			Name:       "Default",
			Recipients: []string{},
			TLSPolicy:  "Optional",
		})
	sesLambda.Permissions = append(sesLambda.Permissions, lambdaSESPermission)

	lambdaFunctions = append(lambdaFunctions, sesLambda)
	Main("SESLambdaApp", "Registers for SES events", lambdaFunctions, nil, nil)
}
