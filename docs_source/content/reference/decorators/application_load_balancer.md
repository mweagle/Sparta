---
date: 2019-08-10 14:59:12
title: Application Load Balancer
weight: 10
alwaysopen: false
---
The [ApplicationLoadBalancerDecorator](https://godoc.org/github.com/mweagle/Sparta/decorator#ApplicationLoadBalancerDecorator) allows you to expose lambda functions as [Application Load Balancer targets](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/lambda-functions.html). 

This can be useful to provide HTTP(S) access to one or more Lambda functions without requiring an API-Gateway service.

## Lambda Function

Application Load Balancer (ALB) lambda targets must satisfy a prescribed Lambda signature: 

```go
import (
  awsEvents "github.com/aws/aws-lambda-go/events"
)

func(context.Context, awsEvents.ALBTargetGroupRequest) awsEvents.ALBTargetGroupResponse
```

See the [ALBTargetGroupRequest](https://godoc.org/github.com/aws/aws-lambda-go/events#ALBTargetGroupRequest) and [ALBTargetGroupResponse](https://godoc.org/github.com/aws/aws-lambda-go/events#ALBTargetGroupResponse) _godoc_ entries for more information.

An example ALB-eligible target function might look like:

```go
// ALB eligible lambda function
func helloNewWorld(ctx context.Context,
  albEvent awsEvents.ALBTargetGroupRequest) (awsEvents.ALBTargetGroupResponse, error) {

  return awsEvents.ALBTargetGroupResponse{
    StatusCode:        200,
    StatusDescription: fmt.Sprintf("200 OK"),
    Body:              "Some other handler",
    IsBase64Encoded:   false,
    Headers:           map[string]string{},
  }, nil
}
```

Once you've defined your ALB-compatible functions, the next step is to register them with the decorator responsible for configuring them as ALB listener targets.

## Definition

The `ApplicationLoadBalancerDecorator` satisfies the [ServiceDecoratorHookHandler](https://godoc.org/github.com/mweagle/Sparta#ServiceDecoratorHookHandler) interface and adds a set of CloudFormation Resources to support properly publishing your Lambda functions.

Since this access path requires an Application Load Balancer, the first step is to define the SecurityGroup associated with the ALB so that incoming requests can be accepted.

The following definition will create an Security Group that accepts public traffic on port `80`:

```go
  sgResName := sparta.CloudFormationResourceName("ALBSecurityGroup", "ALBSecurityGroup")
  sgRes := &gocf.EC2SecurityGroup{
    GroupDescription: gocf.String("ALB Security Group"),
    SecurityGroupIngress: &gocf.EC2SecurityGroupIngressPropertyList{
      gocf.EC2SecurityGroupIngressProperty{
        IPProtocol: gocf.String("tcp"),
        FromPort:   gocf.Integer(80),
        ToPort:     gocf.Integer(80),
        CidrIP:     gocf.String("0.0.0.0/0"),
      },
    },
  }
```

The subnets for our Application Load Balancer are supplied as an environment variable (__TEST_SUBNETS__) of the form `id1,id2`:

```go
  subnetList := strings.Split(os.Getenv("TEST_SUBNETS"), ",")
  subnetIDs := make([]gocf.Stringable, len(subnetList))
  for eachIndex, eachSubnet := range subnetList {
    subnetIDs[eachIndex] = gocf.String(eachSubnet)
  }
```

The next step is to define the ALB and associate it with both the account Subnets and SecurityGroup we just defined:

```go
  alb := &gocf.ElasticLoadBalancingV2LoadBalancer{
    Subnets:        gocf.StringList(subnetIDs...),
    SecurityGroups: gocf.StringList(gocf.GetAtt(sgResName, "GroupId")),
  }
```

This `ElasticLoadBalancingV2LoadBalancer` instance is provided to `NewApplicationLoadBalancerDecorator` to create the decorator that will annotate the CloudFormation
template with the required resources.

```go
albDecorator, albDecoratorErr := spartaDecorators.NewApplicationLoadBalancerDecorator(alb,
    80,
    "HTTP",
    lambdaFn)
```

The `NewApplicationLoadBalancerDecorator` accepts four arguments:

- The `ElasticLoadBalancingV2LoadBalancer` that handles this service's incoming requests
- The port (`80`) that incoming requests will be accepted
- The protocol (`HTTP`) for incoming requests
- The default _*sparta.LambdaAWSInfo_ instance (`lambdaFn`) to use as the ALB's [DefaultAction](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-elasticloadbalancingv2-listener.html#cfn-elasticloadbalancingv2-listener-defaultactions) handler in case no other conditional target matches the incoming request.

## Conditional Targets

Services may expose more than one Lambda function on that same port using multiple [ListenerRule](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-elasticloadbalancingv2-listenerrule.html) entries.

For instance, to register a second lambda function `lambdaFn2` with the same Application Load Balancer at the _/newhello_ path, add a `ConditionalEntry` as in:

```go
  albDecorator.AddConditionalEntry(gocf.ElasticLoadBalancingV2ListenerRuleRuleCondition{
    Field: gocf.String("path-pattern"),
    PathPatternConfig: &gocf.ElasticLoadBalancingV2ListenerRulePathPatternConfig{
      Values: gocf.StringList(gocf.String("/newhello*")),
    },
  }, lambdaFn2)
```

This will create a rule that associates the _/newhello*_ path with `lambdaFn2`. Requests that do not match the incoming path will fallback to the default handler (`lambdaFn`).
See the [RuleCondition](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-elasticloadbalancingv2-listenerrule-conditions.html) documentation
for the full set of conditions that can be expressed.

## Additional Resources

The next step is to ensure that the Security Group that we associated with our ALB is included in the final template. This is done by including it in the `ApplicationLoadBalancerDecorator.Resources` map which allows you to provide additional CloudFormation resources that should be included in the final template:

```go
// Finally, tell the ALB decorator we have some additional resources that need to be
// included in the CloudFormation template
albDecorator.Resources[sgResName] = sgRes
```

## Workflow Hooks

With the decorator fully configured, the final step is to provide it as part of the WorkflowHooks struct:

```go
  // Supply it to the WorkflowHooks and get going...
  workflowHooks := &sparta.WorkflowHooks{
    ServiceDecorators: []sparta.ServiceDecoratorHookHandler{
      albDecorator,
    },
  }

  err := sparta.MainEx(awsName,
    "Simple Sparta application that demonstrates how to make Lambda functions an ALB Target",
    lambdaFunctions,
    nil,
    nil,
    workflowHooks,
    false)
```

## Output

As part of the provisioning workflow, the `ApplicationLoadBalancerDecorator` will include the Application Load Balancer discovery information in the Outputs section as in:

```plain
INFO[0056] Stack Outputs ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬
INFO[0056]     ApplicationLoadBalancerDNS80              Description="ALB DNSName (port: 80, protocol: HTTP)" Value=MyALB-ELBv2-44R3J0MV1D37-943334334.us-west-2.elb.amazonaws.com
INFO[0056]     ApplicationLoadBalancerName80             Description="ALB Name (port: 80, protocol: HTTP)" Value=MyALB-ELBv2-44R3J0MV1D37
INFO[0056]     ApplicationLoadBalancerURL80              Description="ALB URL (port: 80, protocol: HTTP)" Value="http://MyALB-ELBv2-44R3J0MV1D37-943334334.us-west-2.elb.amazonaws.com:80"
```

## Testing

Using _curl_ we can verify the newly provisioned Application Load Balancer behavior. The default lambda function echoes the incoming request and is available at the ALB URL basepath:

```plain
curl http://MyALB-ELBv2-44R3J0MV1D37-943334334.us-west-2.elb.amazonaws.com
```

returns

```json
{
  "httpMethod": "GET",
  "path": "/",
  "headers": {
    "accept": "*/*",
    "host": "MyALB-ELBv2-44R3J0MV1D37-943334334.us-west-2.elb.amazonaws.com",
    "user-agent": "curl/7.54.0",
    "x-amzn-trace-id": "Root=1-5d507bf4-ca98d1ad44ac0fe56ec6a9ae",
    "x-forwarded-for": "24.17.9.178",
    "x-forwarded-port": "80",
    "x-forwarded-proto": "http"
  },
  "requestContext": {
    "elb": {
      "targetGroupArn": "arn:aws:elasticloadbalancing:us-west-2:123412341234:targetgroup/MyALB-ALBDe-1OJX6J3VGX369/1dab61286efaebb6"
    }
  },
  "isBase64Encoded": false,
  "body": ""
}
```

The conditional lambda function behavior is exposed at _/newhello_ as in:

```plain
curl http://MyALB-ELBv2-44R3J0MV1D37-943334334.us-west-2.elb.amazonaws.com/newhello
Some other handler
```
