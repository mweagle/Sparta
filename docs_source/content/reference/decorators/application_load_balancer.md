---
date: 2019-08-10 14:59:12
title: Application Load Balancer
weight: 10
alwaysopen: false
---
The [ApplicationLoadBalancerDecorator](https://godoc.org/github.com/mweagle/Sparta/decorator#ApplicationLoadBalancerDecorator) allows your service to register your lambda functions as [Application Load Balancer targets](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/lambda-functions.html).

This can be useful to expose one or more Lambda functions to the public internet without requiring an API-Gateway configuration. See 

## Lambda Function

The lambda function target for an ALB request is required to have a specific signature: `func(context.Context, awsEvents.ALBTargetGroupRequest) awsEvents.ALBTargetGroupResponse`

See the [ALBTargetGroupRequest](https://godoc.org/github.com/aws/aws-lambda-go/events#ALBTargetGroupRequest) and [ALBTargetGroupResponse](https://godoc.org/github.com/aws/aws-lambda-go/events#ALBTargetGroupResponse) _godoc_ entries for more information.

An example ALB target function might look like:

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

For an ALB that accepts all HTTP traffic on port `80`,  is done via code similar to:

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

The next step is to define the ALB and associate it with both the account Subnets and SecurityGroup we just defined:

```go
  alb := &gocf.ElasticLoadBalancingV2LoadBalancer{
    Subnets:        gocf.StringList(subnetIDs...),
    SecurityGroups: gocf.StringList(gocf.GetAtt(sgResName, "GroupId")),
  }
```

With this definition, we can create a decorator as in:

```go
albDecorator, albDecoratorErr := spartaDecorators.NewApplicationLoadBalancerDecorator(alb,
    80,
    "HTTP",
    lambdaFn)
```

The `NewApplicationLoadBalancerDecorator` accepts four arguments:

- The ApplicationLoadBalancer that will handle this service's incoming requests
- The port (`80`) that incoming requests will be accepted
- The protocol (`HTTP`) for incoming requests
- The default _*sparta.LambdaAWSInfo_ instance (`lambdaFn`) to use as the ALB's [DefaultAction](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-elasticloadbalancingv2-listener.html#cfn-elasticloadbalancingv2-listener-defaultactions) handler in case no other conditional target matches the incoming request.

## Conditional Targets

Services may expose more than a single Lambda function on the same port using [ListenerRule](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-elasticloadbalancingv2-listenerrule.html) entries. 

For instance, to expose a second lambda function `lambdaFn2` on the same listener but on a specific path, register a `ConditionalEntry` as in:

```go
  albDecorator.AddConditionalEntry(gocf.ElasticLoadBalancingV2ListenerRuleRuleCondition{
    Field: gocf.String("path-pattern"),
    PathPatternConfig: &gocf.ElasticLoadBalancingV2ListenerRulePathPatternConfig{
      Values: gocf.StringList(gocf.String("/newhello*")),
    },
  }, lambdaFn2)
```
This will create a rule that associates the _/newhello*_ path with `lambdaFn2`. Requests that do not match the incoming path will fallback to the default handler (`lambdaFn`).

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

The `ApplicationLoadBalancerDecorator` includes the Application Load Balancer discovery information in the Outputs section as in:

```plain
INFO[0156] Stack Outputs ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬
INFO[0156]     ApplicationLoadBalancerDNS80              Description="DNS value of the ALB" Value=MyALB-ELBv2-44R3J0MV1D37-943334334.us-west-2.elb.amazonaws.com
INFO[0156]     ApplicationLoadBalancerName80             Description="Name of the ALB" Value=MyALB-ELBv2-44R3J0MV1D37
INFO[0156]     ApplicationLoadBalancerURL80              Description="URL value of the ALB" Value="http://MyALB-ELBv2-44R3J0MV1D37-943334334.us-west-2.elb.amazonaws.com:80"
```
