---
date: 2016-03-09T19:56:50+01:00
title: Slack SlashCommand
weight: 21
---

![SlashLogo](/images/apigateway/slack/slack_rgb.png)

In this example, we'll walk through creating a [Slack Slash Command](https://api.slack.com/slash-commands) service.  The source for
this is the [SpartaSlackbot](https://github.com/mweagle/SpartaSlackbot) repo.

Our initial command handler won't be very sophisticated, but will show the steps necessary to provision and configure a Sparta AWS Gateway-enabled Lambda function.

# Define the Lambda Function

This lambda handler is a bit more complicated than the other examples, primarily because of the [Slack Integration](https://api.slack.com/slash-commands) requirements.  The full source is:

```go
import (
	spartaAWSEvents "github.com/mweagle/Sparta/aws/events"
)
////////////////////////////////////////////////////////////////////////////////
// Hello world event handler
//
func helloSlackbot(ctx context.Context,
	apiRequest spartaAWSEvents.APIGatewayRequest) (map[string]interface{}, error) {
	logger, _ := ctx.Value(sparta.ContextKeyLogger).(*logrus.Logger)

	bodyParams, bodyParamsOk := apiRequest.Body.(map[string]interface{})
	if !bodyParamsOk {
		return nil, fmt.Errorf("Failed to type convert body. Type: %T", apiRequest.Body)
	}

	logger.WithFields(logrus.Fields{
		"BodyType":  fmt.Sprintf("%T", bodyParams),
		"BodyValue": fmt.Sprintf("%+v", bodyParams),
	}).Info("Slack slashcommand values")

	// 2. Create the response
	// Slack formatting:
	// https://api.slack.com/docs/formatting
	responseText := "Here's what I understood"
	for eachKey, eachParam := range bodyParams {
		responseText += fmt.Sprintf("\n*%s*: %+v", eachKey, eachParam)
	}

	// 4. Setup the response object:
	// https://api.slack.com/slash-commands, "Responding to a command"
	responseData := map[string]interface{}{
		"response_type": "in_channel",
		"text":          responseText,
		"mrkdwn":        true,
	}
	return responseData, nil
}
```

There are a couple of things to note in this code:

1. **Custom Event Type**
  - The inbound Slack `POST` request is `application/x-www-form-urlencoded` data.  This
	data is unmarshalled into the same _spartaAWSEvent.APIGatewayRequest_ using
	a customized [mapping template](https://github.com/mweagle/Sparta/blob/master/resources/provision/apigateway/inputmapping_formencoded.vtl).

1. **Response Formatting**
The lambda function extracts all Slack parameters and if defined, sends the `text` back with a bit of [Slack Message Formatting](https://api.slack.com/docs/formatting):

        ```go
				responseText := "Here's what I understood"
				for eachKey, eachParam := range bodyParams {
					responseText += fmt.Sprintf("\n*%s*: %+v", eachKey, eachParam)
				}
        ```

1. **Custom Response**
  - The Slack API expects a [JSON formatted response](https://api.slack.com/slash-commands), which is created in step 4:

        ```go
        responseData := sparta.ArbitraryJSONObject{
      		"response_type": "in_channel",
      		"text":          responseText,
      	}
        ```

# Create the API Gateway

With our lambda function defined, we need to setup an API Gateway so that it's publicly available:

```go
apiStage := sparta.NewStage("v1")
apiGateway := sparta.NewAPIGateway("SpartaSlackbot", apiStage)
```

The `apiStage` value implies that we want to deploy this API Gateway Rest API as part of Sparta's `provision` step.

# Create Lambda Binding & Resource

Next we create an `sparta.LambdaAWSInfo` struct that references the `s3ItemInfo` function:

```go
func spartaLambdaFunctions(api *sparta.API) []*sparta.LambdaAWSInfo {
	var lambdaFunctions []*sparta.LambdaAWSInfo
	lambdaFn := sparta.HandleAWSLambda(sparta.LambdaName(helloSlackbot),
		helloSlackbot,
		iamDynamicRole)

	if nil != api {
		apiGatewayResource, _ := api.NewResource("/slack", lambdaFn)
		_, err := apiGatewayResource.NewMethod("POST", http.StatusCreated)
		if nil != err {
			panic("Failed to create /hello resource")
		}
	}
	return append(lambdaFunctions, lambdaFn)
}
```

A few items to note here:

  * We're using an empty `sparta.IAMRoleDefinition{}` definition because our go lambda function doesn't access any additional AWS services.
  * Our lambda function will be accessible at the _/slack_ child path of the deployed API Gateway instance
  * Slack supports both [GET and POST](https://api.slack.com/slash-commands) integration types, but we're limiting our lambda function to `POST` only

# Provision

With everything configured, we then configure our `main()` function to forward to Sparta:

```go
func main() {
	// Register the function with the API Gateway
	apiStage := sparta.NewStage("v1")
	apiGateway := sparta.NewAPIGateway("SpartaSlackbot", apiStage)

	// Deploy it
	sparta.Main("SpartaSlackbot",
		fmt.Sprintf("Sparta app that responds to Slack commands"),
		spartaLambdaFunctions(apiGateway),
		apiGateway,
		nil)
}

```

and provision the service:

```nohighlight
S3_BUCKET=<MY_S3_BUCKETNAME> go run slack.go --level info provision
```

Look for the _Stack output_ section of the log, you'll need the **APIGatewayURL** value to configure Slack in the next step.

```nohighlight
INFO[0083] Stack output Description=API Gateway URL Key=APIGatewayURL Value=https://75mtsly44i.execute-api.us-west-2.amazonaws.com/v1
INFO[0083] Stack output Description=Sparta Home Key=SpartaHome Value=https://github.com/mweagle/Sparta
INFO[0083] Stack output Description=Sparta Version Key=SpartaVersion Value=0.1.3
```


# Configure Slack

At this point our lambda function is deployed and is available through the API Gateway (_https://75mtsly44i.execute-api.us-west-2.amazonaws.com/v1/slack_ in the current example).

The next step is to configure Slack with this custom integration:

  1. Visit https://slack.com/apps/build and choose the "Custom Integration" option:

    ![Custom integration](/images/apigateway/slack/customIntegration.jpg)

  1. On the next page, choose "Slash Commands":

    ![Slash Commands](/images/apigateway/slack/slashCommandMenu.jpg)

  1. The next screen is where you input the command that will trigger your lambda function.  Enter `/sparta`

    ![Slash Chose Command](/images/apigateway/slack/chooseCommand.jpg)

    - and click the "Add Slash Command Integration" button.

  1. Finally, scroll down the next page to the **Integration Settings** section and provide the API Gateway URL of your lambda function.

    ![Slash URL](/images/apigateway/slack/integrationSettings.jpg)

    * Leave the _Method_ field unchanged (it should be `POST`), to match how we configured the API Gateway entry above.

  1. Save it

    ![Save it](/images/apigateway/slack/saveIntegration.jpg)


There are additional Slash Command Integration options, but for this example the **URL** option is sufficient to trigger our command.

# Test

With everything configured, visit your team's Slack room and verify the integration via `/sparta` slash command:

![Sparta Response](/images/apigateway/slack/slackResponse.jpg)

# Cleaning Up

Before moving on, remember to decommission the service via:

```nohighlight
go run slack.go delete
```

# Wrapping Up

This example provides a good overview of Sparta & Slack integration, including how to handle external requests that are not `application/json` formatted.
