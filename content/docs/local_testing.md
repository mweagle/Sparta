---
date: 2016-03-09T19:56:50+01:00
title: Local Testing
weight: 10
menu:
  main:
    parent: Documentation
    identifier: local-testing
    weight: 0
---
While developing Sparta lambda functions it may be useful to test them locally without needing to `provision` each new code change.  Sparta supports _localhost_ testing in two different ways:

  - The `explore` command line option
  -  `httptest.NewServer` for _go test_ style testing

# Example

For this example, let's define a simple Sparta application:

{{< highlight go >}}
package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	sparta "github.com/mweagle/Sparta"
)

////////////////////////////////////////////////////////////////////////////////
// Hello world event handler
//
func helloWorld(event *json.RawMessage,
	context *sparta.LambdaContext,
	w http.ResponseWriter,
	logger *logrus.Logger) {

	fmt.Fprint(w, "Hello World")
}

////////////////////////////////////////////////////////////////////////////////
// Main
func main() {
	// Deploy it
	lambdaFn := sparta.NewLambda(sparta.IAMRoleDefinition{}, helloWorld, nil)
	var lambdaFunctions []*sparta.LambdaAWSInfo
	lambdaFunctions = append(lambdaFunctions, lambdaFn)

	sparta.Main("SpartaExplore",
		fmt.Sprintf("Test explore command"),
		lambdaFunctions,
		nil,
		nil)
}
{{< /highlight >}}

# Command Line Testing

With our application defined, let's run it:

{{< highlight nohighlight>}}
go run main.go explore

INFO[0000] Welcome to Sparta                             Option=explore TS=2016-01-31T18:05:19Z Version=0.3.0
INFO[0000] --------------------------------------------------------------------------------
INFO[0000] The following URLs are available for testing.
INFO[0000] main.helloWorld                               URL=http://localhost:9999/main.helloWorld
INFO[0000] Functions can be invoked via application/json over POST
INFO[0000] 	curl -vs -X POST -H "Content-Type: application/json" --data @testEvent.json http://localhost:9999/main.helloWorld
INFO[0000] Where @testEvent.json is a local file with top level `context` and `event` properties:
INFO[0000] 	{context: {}, event: {}}
INFO[0000] Starting Sparta server                        URL=http://localhost:9999
{{< /highlight >}}

The _localhost_ server mirrors the contract between the NodeJS proxying tier and the **Go** binary that is used in the AWS Lambda execution environment.

Per the instructions, let's create a _testEvent.json_ file with the required structure:

{{< highlight json >}}
{
  "context" : {},
  "event" : {}
}
{{< /highlight >}}

and post it:

{{< highlight >}}
curl -vs -X POST -H "Content-Type: application/json" --data @testEvent.json http://localhost:9999/main.helloWorld

*   Trying ::1...
* Connected to localhost (::1) port 9999 (#0)
> POST /main.helloWorld HTTP/1.1
> Host: localhost:9999
> User-Agent: curl/7.43.0
> Accept: */*
> Content-Type: application/json
> Content-Length: 33
>
* upload completely sent off: 33 out of 33 bytes
< HTTP/1.1 200 OK
< Date: Mon, 21 Dec 2015 03:13:33 GMT
< Content-Length: 11
< Content-Type: text/plain; charset=utf-8
<
* Connection #0 to host localhost left intact
Hello World
{{< /highlight >}}

Our lambda function (which at this point is just an HTTP handler) was successfully called and responded successfully.

While this approach does work, it's not a scalable approach to writing automated tests.

# <code>httptest</code> support

To integrate with the existing [go test](https://golang.org/pkg/testing/) command, Sparta includes two functions:

  1. `NewLambdaHTTPHandler` : Creates an [httptest.NewServer](https://golang.org/pkg/net/http/httptest/#NewServer)-compliant `http.Handler` value.
  1. `Sparta/explore.NewRequest` : Creates a mock JSON object with optional user-defined *event* data.

To show this in action, let's walk through how Sparta [does this](https://github.com/mweagle/Sparta/blob/master/explore_test.go):

{{< highlight golang>}}

func TestExplore(t *testing.T) {
	// 1. Create the function(s) we want to test
	var lambdaFunctions []*LambdaAWSInfo
	lambdaFn := NewLambda(IAMRoleDefinition{}, exploreTestHelloWorld, nil)
	lambdaFunctions = append(lambdaFunctions, lambdaFn)

	// 2. Mock event specific data to send to the lambda function
	eventData := ArbitraryJSONObject{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3"}

	// 3. Make the request and confirm
	logger, _ := NewLogger("warning")
	ts := httptest.NewServer(NewLambdaHTTPHandler(lambdaFunctions, logger))
	defer ts.Close()
	resp, err := explore.NewRequest(lambdaFn.URLPath(), eventData, ts.URL)
	if err != nil {
		t.Fatal(err.Error())
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	t.Log("Status: ", resp.Status)
	t.Log("Headers: ", resp.Header)
	t.Log("Body: ", string(body))
}
{{< /highlight >}}

The test function above has the following features:

  1. Create a slice of one or more Sparta lambda functions to test
  1. Optionally define *event* data to use in the test
  1. Create a new `httptest.NewServer`
  1. Issue the test case request with `Sparta/explore.NewRequest`
  1. Validate the test results

These _localhost_ tests will be executed as part of your application's normal `go test` lifecycle.

# Notes
  * APIGateway requests can be tested using mock payloads via [Sparta/explore.NewAPIGatewayRequest](https://github.com/mweagle/Sparta/blob/master/explore/explore.go#L103)
  * Ensure your localhost AWS credentials have sufficient privileges to access any AWS services.
    * Sparta does not provision `IAM::Role` resources for local testing.
  * Localhost testing is not a substitute for CI/CD pipelines.
    - [You build it, you run it](https://queue.acm.org/detail.cfm?id=1142065)
    - [Velocity & Volume](https://youtu.be/wyWI3gLpB8o)
