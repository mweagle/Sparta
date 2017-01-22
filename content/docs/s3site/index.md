+++
author = "Matt Weagle"
date = "2015-11-29T06:50:17"
title = "S3 Static Site with CORS"
tags = ["sparta"]
type = "doc"
+++

Sparta supports provisioning an S3-backed [static website](http://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteHosting.html) as part of provisioning.  We'll walk through provisioning a minimal [Bootstrap](http://getbootstrap.com) website that accesses API Gateway lambda functions provisioned by a single service in this example.

The source for this is the [SpartaHTML](https://github.com/mweagle/SpartaHTML) example application.

# Create the Lambda function

We'll start by creating a very simple lambda function:

{{< highlight go >}}
func helloWorld(event *json.RawMessage,
	context *sparta.LambdaContext,
	w http.ResponseWriter,
	logger *logrus.Logger) {
	logger.Info("Hello World: ", string(*event))
	fmt.Fprint(w, string(*event))
}
{{< /highlight >}}

This lambda function simply sends back the content of the inbound event.  See the [API Gateway example](/docs/apigateway/example1) for more information on the event contents.

# Create the API Gateway

The next step is to create an API Gateway instance and Stage, so that the API will be publicly available.

{{< highlight go >}}
apiStage := sparta.NewStage("v1")
apiGateway := sparta.NewAPIGateway("SpartaHTML", apiStage)
{{< /highlight >}}

Since we want to be able to access this API from another domain (the one provisioned by the S3 bucket), we'll need to [enable CORS](http://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-cors.html) as well:

{{< highlight go >}}
// Enable CORS s.t. the S3 site can access the resources
apiGateway.CORSEnabled = true
{{< /highlight >}}

Finally, we register the `helloWorld` lambda function with an API Gateway resource:

{{< highlight go >}}

func spartaLambdaFunctions(api *sparta.API) []*sparta.LambdaAWSInfo {
	var lambdaFunctions []*sparta.LambdaAWSInfo
	lambdaFn := sparta.NewLambda(sparta.IAMRoleDefinition{}, helloWorld, nil)

	if nil != api {
		apiGatewayResource, _ := api.NewResource("/hello", lambdaFn)
		_, err := apiGatewayResource.NewMethod("GET", http.StatusOK)
		if nil != err {
			panic("Failed to create /hello resource")
		}
	}
	return append(lambdaFunctions, lambdaFn)
}
{{< /highlight >}}


# Define the S3 Site

The next part is to define the S3 site resources via `sparta.NewS3Site(localFilePath)`.  The _localFilePath_ parameter
typically points to a directory, which will be:

  1. Recursively ZIP'd
  1. Posted to S3 alongside the Lambda code archive and CloudFormation Templates
  1. Dynamically unpacked by a CloudFormation CustomResource during `provision` to a new S3 bucket.

# Provision

Putting it all together, our `main()` function looks like:

{{< highlight go >}}

////////////////////////////////////////////////////////////////////////////////
// Main
func main() {
	// Register the function with the API Gateway
	apiStage := sparta.NewStage("v1")
	apiGateway := sparta.NewAPIGateway("SpartaHTML", apiStage)
	// Enable CORS s.t. the S3 site can access the resources
	apiGateway.CORSEnabled = true

	// Provision a new S3 bucket with the resources in the supplied subdirectory
	s3Site, _ := sparta.NewS3Site("./resources")

	// Deploy it
	sparta.Main("SpartaHTML",
		fmt.Sprintf("Sparta app that provisions a CORS-enabled API Gateway together with an S3 site"),
		spartaLambdaFunctions(apiGateway),
		apiGateway,
		s3Site)
}
{{< /highlight >}}

which can be provisioned using the standard [command line](/docs/commandline) option.

The _Outputs_ section of the `provision` command includes the hostname of our new S3 site:

{{< highlight nohighlight >}}
INFO[0114] Stack output        Description=API Gateway URL Key=APIGatewayURL Value=https://in8vahv6c8.execute-api.us-west-2.amazonaws.com/v1
INFO[0114] Stack output        Description=S3 website URL Key=S3SiteURL Value=http://spartahtml-site09b75dfd6a3e4d7e2167f6eca73957ee83-1c31huc6oly7k.s3-website-us-west-2.amazonaws.com
INFO[0114] Stack output        Description=Sparta Home Key=SpartaHome Value=https://github.com/mweagle/Sparta
INFO[0114] Stack output        Description=Sparta Version Key=SpartaVersion Value=0.1.0
INFO[0114] Stack provisioned   CreationTime=2015-12-15 17:25:11.323 +0000 UTC StackId=arn:aws:cloudformation:us-west-2:123412341234:stack/SpartaHTML/cb891ce0-a350-11e5-be26-507bfc8840a6 StackName=SpartaHTML
INFO[0114] Elapsed time        Seconds=114
{{< /highlight >}}

Open your browser to the `S3SiteURL` value (eg: _http://spartahtml-site09b75dfd6a3e4d7e2167f6eca73957ee83-1c31huc6oly7k.s3-website-us-west-2.amazonaws.com_) and view the deployed site.

# Discover

An open issue is how to communicate the dynamically assigned API Gateway hostname to the dynamically provisioned S3 site.

As part of expanding the ZIP archive to a target S3 bucket, Sparta also creates a _MANIFEST.json_ discovery file with discovery information. If your application has provisioned an APIGateway this JSON file will include that dynamically assigned URL as in:

  1. **MANIFEST.json**
{{< highlight json >}}
{
 "APIGatewayURL": {
  "Description": "API Gateway URL",
  "Value": "https://r3zq0apo1g.execute-api.us-west-2.amazonaws.com/v1"
 }
}
{{< /highlight >}}
