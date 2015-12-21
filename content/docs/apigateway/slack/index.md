+++
author = "Matt Weagle"
date = "2015-11-29T06:50:17"
title = "Slack"
tags = ["sparta"]
type = "doc"
+++

## <a href="{{< relref "#intro" >}}">Slack SlashCommand</a>

In this example, we'll walk through creating a [Slack Slash Command](https://api.slack.com/slash-commands) service.  The source for this is the [SpartaSlackbot](https://github.com/mweagle/SpartaSlackbot) repo.

Our initial command handler won't be very sophisticated, but will show the steps necessary to provision and configure a Sparta AWS Gateway-enabled Lambda function.  

## <a href="{{< relref "#lambda" >}}">Define the Lambda Function</a>

This lambda handler is a bit more complicated than the other examples, primarily because of the [Slack Integration](https://api.slack.com/slash-commands) requirements.  The full source is:

{{< highlight go >}}

////////////////////////////////////////////////////////////////////////////////
// Hello world event handler
//
func helloSlackbot(event *json.RawMessage,
	context *sparta.LambdaContext,
	w http.ResponseWriter,
	logger *logrus.Logger) {

	// 1. Unmarshal the primary event
	var lambdaEvent slackLambdaJSONEvent
	err := json.Unmarshal([]byte(*event), &lambdaEvent)
	if err != nil {
		logger.Error("Failed to unmarshal event data: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// 2. Conditionally unmarshal to get the Slack text.  See
	// https://api.slack.com/slash-commands
	// for the value name list
	requestParams := url.Values{}
	if bodyData, ok := lambdaEvent.Body.(string); ok {
		requestParams, err = url.ParseQuery(bodyData)
		if err != nil {
			logger.Error("Failed to parse query: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		logger.WithFields(logrus.Fields{
			"Values": requestParams,
		}).Info("Slack slashcommand values")
	} else {
		logger.Info("Event body empty")
	}

	// 3. Create the response
	// Slack formatting:
	// https://api.slack.com/docs/formatting
	responseText := "You talkin to me?"
	for _, eachLine := range requestParams["text"] {
		responseText += fmt.Sprintf("\n>>> %s", eachLine)
	}

	// 4. Setup the response object:
	// https://api.slack.com/slash-commands, "Responding to a command"
	responseData := sparta.ArbitraryJSONObject{
		"response_type": "in_channel",
		"text":          responseText,
	}
	// 5. Send it off
	responseBody, err := json.Marshal(responseData)
	if err != nil {
		logger.Error("Failed to marshal response: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fmt.Fprint(w, string(responseBody))
}
{{< /highlight >}}

There are a couple of things to note in this code:

1. **Custom Event Type**
  - The inbound Slack `POST` request is `application/x-www-form-urlencoded` data.  However, our [integration mapping](https://github.com/mweagle/Sparta/blob/master/resources/provision/apigateway/inputmapping_default.vtl) mediates the API Gateway HTTPS request, transforming the public request into an integration request.  The integration mapping wraps the raw `POST` body with the mapping envelope (so that we can access [identity information, HTTP headers, etc.](/docs/apigateway/example1)), which produces an inbound JSON request that includes a **Body** parameter.  The **Body** string value is the raw inbound `POST` data.  Since it's `application/x-www-form-urlencoded`, to get the actual parameters we need to parse it:

    ```javascript
    if bodyData, ok := lambdaEvent.Body.(string); ok {
      requestParams, err = url.ParseQuery(bodyData)
    ```

  - The lambda function extracts all Slack parameters and if defined, sends the `text` back with a bit of [Slack Message Formatting](https://api.slack.com/docs/formatting) (and some attitude, to be honest about it):

    ```javascript
    responseText := "You talkin to me?"
    for _, eachLine := range requestParams["text"] {
      responseText += fmt.Sprintf("\n>>> %s", eachLine)
    }
    ```

1. **Custom Response**
  - The Slack API expects a [JSON formatted response](https://api.slack.com/slash-commands), which is created in step 4:

  ```javascript
  responseData := sparta.ArbitraryJSONObject{
		"response_type": "in_channel",
		"text":          responseText,
	}
  ```

### <a href="{{< relref "#example2API" >}}">Create the API Gateway</a>

With our lambda function defined, we need to setup an API Gateway so that it's publicly available:

{{< highlight go >}}
apiStage := sparta.NewStage("v1")
apiGateway := sparta.NewAPIGateway("SpartaSlackbot", apiStage)
{{< /highlight >}}

The `apiStage` value implies that we want to deploy this API Gateway Rest API as part of Sparta's `provision` step.  

### <a href="{{< relref "#example2Resource" >}}">Create Lambda Binding & Resource</a>

Next we create an `sparta.LambdaAWSInfo` struct that references the `s3ItemInfo` function:

{{< highlight go >}}
func spartaLambdaFunctions(api *sparta.API) []*sparta.LambdaAWSInfo {
	var lambdaFunctions []*sparta.LambdaAWSInfo
	lambdaFn := sparta.NewLambda(sparta.IAMRoleDefinition{}, helloSlackbot, nil)

	if nil != api {
		apiGatewayResource, _ := api.NewResource("/slack", lambdaFn)
		_, err := apiGatewayResource.NewMethod("POST")
		if nil != err {
			panic("Failed to create /hello resource")
		}
	}
	return append(lambdaFunctions, lambdaFn)
}
{{< /highlight >}}

A few items to note here:

  * We're using an empty `sparta.IAMRoleDefinition{}` definition because our go lambda function doesn't access any additional AWS services.
  * Our lambda function will be accessible at the _/slack_ child path of the deployed API Gateway instance
  * Slack supports both [GET and POST](https://api.slack.com/slash-commands) integration types, but we're limiting our lambda function to `POST` only

### <a href="{{< relref "#provision" >}}">Provision</a>

With everything configured, we then configure our `main()` function to forward to Sparta:

{{< highlight go >}}
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

{{< /highlight >}}

and provision the service:

{{< highlight nohighlight >}}
S3_BUCKET=<MY_S3_BUCKETNAME> go run slack.go --level info provision
{{< /highlight >}}

Look for the _Stack output_ section of the log, you'll need the **APIGatewayURL** value to configure Slack in the next step.

{{< highlight nohighlight >}}
INFO[0083] Stack output Description=API Gateway URL Key=APIGatewayURL Value=https://75mtsly44i.execute-api.us-west-2.amazonaws.com/v1
INFO[0083] Stack output Description=Sparta Home Key=SpartaHome Value=https://github.com/mweagle/Sparta
INFO[0083] Stack output Description=Sparta Version Key=SpartaVersion Value=0.1.3
{{< /highlight >}}


### <a href="{{< relref "#configureSlack" >}}">Configure Slack</a>

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

### <a href="{{< relref "#test" >}}">Test</a>

With everything configured, visit your team's Slack room and verify the integration via `/sparta` slash command:

![Sparta Response](/images/apigateway/slack/slackResponse.jpg)    
