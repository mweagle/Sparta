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
func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World"))
}

////////////////////////////////////////////////////////////////////////////////
// Main
func main() {
	// Deploy it
	lambdaFn := sparta.HandleAWSLambda(sparta.LambdaName(helloWorld),
		http.HandlerFunc(helloWorld),
		sparta.IAMRoleDefinition{})
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

{{< highlight go >}}
go run main.go explore

INFO[0000] ========================================
INFO[0000] Welcome to MyHelloWorldStack                  GoVersion=go1.9 LinkFlags= Option=explore SpartaSHA=d3479d7 SpartaVersion=0.20.1 UTC="2017-10-05T02:46:11Z"
INFO[0000] ========================================
INFO[0000] The following URLs are available for testing.
INFO[0000] Hello_World                                   URL="http://localhost:9999/Hello_World"
INFO[0000] Functions can be invoked via application/json over POST
INFO[0000] 	curl -vs -X POST -H "Content-Type: application/json" --data @testEvent.json http://localhost:9999/Hello_World
INFO[0000] Where @testEvent.json is a local file with top level `context` and `event` properties:
INFO[0000] 	{"context": {}, "event": {}}
INFO[0000] Signaling parent process                      ParentPID=0
INFO[0000] Starting main server                          URL="http://localhost:9999"
{{< /highlight >}}

The _localhost_ server mirrors the contract between the NodeJS proxying tier and the **go** binary that is used in the AWS Lambda execution environment.

Per the instructions, let's create a _testEvent.json_ file with the required structure:

{{< highlight json >}}
{
  "context" : {},
  "event" : {}
}
{{< /highlight >}}

and post it:

{{< highlight bash >}}
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

{{< highlight go >}}
func TestExplore(t *testing.T) {
	// 1. Create the function(s) we want to test
	var lambdaFunctions []*LambdaAWSInfo
	lambdaFn := sparta.HandleAWSLambda(sparta.LambdaName(exploreTestHelloWorld),
		http.HandlerFunc(exploreTestHelloWorld),
		sparta.IAMRoleDefinition{})
	lambdaFunctions = append(lambdaFunctions, lambdaFn)

	// 2. Mock event specific data to send to the lambda function
	eventData := ArbitraryJSONObject{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3"}

	// 3. Make the request and confirm
	logger, _ := NewLogger("warning")
	ts := httptest.NewServer(NewServeMuxLambda(lambdaFunctions, logger))
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
