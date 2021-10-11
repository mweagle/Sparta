package decorator

import (
	"fmt"
	"strings"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	gof "github.com/awslabs/goformation/v5/cloudformation"
	gofelbv2 "github.com/awslabs/goformation/v5/cloudformation/elasticloadbalancingv2"
	goflambda "github.com/awslabs/goformation/v5/cloudformation/lambda"
	sparta "github.com/mweagle/Sparta"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type targetGroupEntry struct {
	conditions []gofelbv2.ListenerRule_RuleCondition
	lambdaFn   *sparta.LambdaAWSInfo
	priority   int
}

// ApplicationLoadBalancerDecorator is an instance of a service decorator that
// handles registering Lambda functions with an Application Load Balancer.
type ApplicationLoadBalancerDecorator struct {
	alb                  *gofelbv2.LoadBalancer
	port                 int
	protocol             string
	defaultLambdaHandler *sparta.LambdaAWSInfo
	targets              []*targetGroupEntry
	Resources            map[string]gof.Resource
}

// LogicalResourceName returns the CloudFormation resource name of the primary
// ALB
func (albd *ApplicationLoadBalancerDecorator) LogicalResourceName() string {
	return sparta.CloudFormationResourceName("ELBv2Resource", "ELBv2Resource")
}

// AddConditionalEntry adds a new lambda target that is conditionally routed
// to depending on the condition value.
func (albd *ApplicationLoadBalancerDecorator) AddConditionalEntry(condition gofelbv2.ListenerRule_RuleCondition,
	lambdaFn *sparta.LambdaAWSInfo) *ApplicationLoadBalancerDecorator {

	return albd.AddConditionalEntryWithPriority(condition, 0, lambdaFn)
}

// AddConditionalEntryWithPriority adds a new lambda target that is conditionally routed
// to depending on the condition value using the user supplied priority value
func (albd *ApplicationLoadBalancerDecorator) AddConditionalEntryWithPriority(condition gofelbv2.ListenerRule_RuleCondition,
	priority int,
	lambdaFn *sparta.LambdaAWSInfo) *ApplicationLoadBalancerDecorator {

	return albd.AddMultiConditionalEntryWithPriority([]gofelbv2.ListenerRule_RuleCondition{condition},
		priority,
		lambdaFn)
}

// AddMultiConditionalEntry adds a new lambda target that is conditionally routed
// to depending on the multi condition value.
func (albd *ApplicationLoadBalancerDecorator) AddMultiConditionalEntry(conditions []gofelbv2.ListenerRule_RuleCondition,
	lambdaFn *sparta.LambdaAWSInfo) *ApplicationLoadBalancerDecorator {

	return albd.AddMultiConditionalEntryWithPriority(conditions, 0, lambdaFn)
}

// AddMultiConditionalEntryWithPriority adds a new lambda target that is conditionally routed
// to depending on the multi condition value with the given priority index
func (albd *ApplicationLoadBalancerDecorator) AddMultiConditionalEntryWithPriority(conditions []gofelbv2.ListenerRule_RuleCondition,
	priority int,
	lambdaFn *sparta.LambdaAWSInfo) *ApplicationLoadBalancerDecorator {

	// Add a version resource to the lambda so that we target that resource...
	albd.targets = append(albd.targets, &targetGroupEntry{
		conditions: conditions,
		priority:   priority,
		lambdaFn:   lambdaFn,
	})
	return albd
}

// DecorateService satisfies the ServiceDecoratorHookHandler interface
func (albd *ApplicationLoadBalancerDecorator) DecorateService(context map[string]interface{},
	serviceName string,
	template *gof.Template,
	S3Bucket string,
	S3Key string,
	buildID string,
	awsConfig awsv2.Config,
	noop bool,
	logger *zerolog.Logger) error {

	portScopedResourceName := func(prefix string, parts ...string) string {
		return sparta.CloudFormationResourceName(fmt.Sprintf("%s%d", prefix, albd.port),
			parts...)
	}

	////////////////////////////////////////////////////////////////////////////
	// Closure to manage the permissions, version, and alias resources needed
	// for each lambda target group
	//
	visitedLambdaFuncs := make(map[string]bool)
	ensureLambdaPreconditions := func(lambdaFn *sparta.LambdaAWSInfo, dependentResource gof.Resource) error {
		_, exists := visitedLambdaFuncs[lambdaFn.LogicalResourceName()]
		if exists {
			return nil
		}
		// Add the lambda permission
		albPermissionResourceName := portScopedResourceName("ALBPermission", lambdaFn.LogicalResourceName())
		lambdaInvokePermission := &goflambda.Permission{
			Action:       "lambda:InvokeFunction",
			FunctionName: gof.GetAtt(lambdaFn.LogicalResourceName(), "Arn"),
			Principal:    sparta.ElasticLoadBalancingPrincipal,
		}
		template.Resources[albPermissionResourceName] = lambdaInvokePermission

		// The stable alias resource and unstable, retained version resource
		aliasResourceName := portScopedResourceName("ALBAlias", lambdaFn.LogicalResourceName())
		versionResourceName := portScopedResourceName("ALBVersion", lambdaFn.LogicalResourceName(), buildID)

		versionResource := &goflambda.Version{
			FunctionName: gof.GetAtt(lambdaFn.LogicalResourceName(), "Arn"),
		}
		versionResource.AWSCloudFormationDeletionPolicy = "Retain"
		template.Resources[versionResourceName] = versionResource

		// Add the alias that binds the lambda to the version...
		aliasResource := &goflambda.Alias{
			FunctionVersion: gof.GetAtt(versionResourceName, "Version"),
			FunctionName:    gof.Ref(lambdaFn.LogicalResourceName()),
			Name:            "live",
		}
		// One time only
		aliasResource.AWSCloudFormationDependsOn = append(aliasResource.AWSCloudFormationDependsOn,
			albPermissionResourceName,
			versionResourceName,
			aliasResourceName)
		template.Resources[aliasResourceName] = aliasResource

		visitedLambdaFuncs[lambdaFn.LogicalResourceName()] = true
		return nil
	}

	////////////////////////////////////////////////////////////////////////////
	// START
	//
	// Add the alb. We'll link each target group inside the loop...
	template.Resources[albd.LogicalResourceName()] = albd.alb
	//	albRes := template.AddResource(albd.LogicalResourceName(), albd.alb)
	defaultListenerResName := portScopedResourceName("ALBListener", "DefaultListener")
	defaultTargetGroupResName := portScopedResourceName("ALBDefaultTarget", albd.defaultLambdaHandler.LogicalResourceName())

	// Create the default lambda target group...
	defaultTargetGroupRes := &gofelbv2.TargetGroup{
		TargetType: "lambda",
		Targets: []gofelbv2.TargetGroup_TargetDescription{
			gofelbv2.TargetGroup_TargetDescription{
				Id: gof.GetAtt(albd.defaultLambdaHandler.LogicalResourceName(), "Arn"),
			},
		},
	}
	// Add it...
	template.Resources[defaultTargetGroupResName] = defaultTargetGroupRes

	// Then create the ELB listener with the default entry. We'll add the conditional
	// lambda targets after this...
	listenerRes := &gofelbv2.Listener{
		LoadBalancerArn: gof.Ref(albd.LogicalResourceName()),
		Port:            albd.port,
		Protocol:        albd.protocol,
		DefaultActions: []gofelbv2.Listener_Action{
			{
				TargetGroupArn: gof.Ref(defaultTargetGroupResName),
				Type:           "forward",
			},
		},
	}
	listenerRes.AWSCloudFormationDependsOn = []string{defaultTargetGroupResName}
	template.Resources[defaultListenerResName] = listenerRes

	// Make sure this is all hooked up
	ensureErr := ensureLambdaPreconditions(albd.defaultLambdaHandler, listenerRes)
	if ensureErr != nil {
		return errors.Wrapf(ensureErr, "Failed to create precondition resources for Lambda TargetGroup")
	}
	// Finally, ensure that each lambdaTarget has a single InvokePermission permission
	// set so that the ALB can actually call them...
	for eachIndex, eachTarget := range albd.targets {
		// Create a new TargetGroup for this lambda function
		conditionalLambdaTargetGroupResName := portScopedResourceName("ALBTargetCond",
			eachTarget.lambdaFn.LogicalResourceName())
		conditionalLambdaTargetGroup := &gofelbv2.TargetGroup{
			TargetType: "lambda",
			Targets: []gofelbv2.TargetGroup_TargetDescription{
				{
					Id: gof.GetAtt(eachTarget.lambdaFn.LogicalResourceName(), "Arn"),
				},
			},
		}
		// Add it...
		template.Resources[conditionalLambdaTargetGroupResName] = conditionalLambdaTargetGroup
		// Create the stable alias resource resource....
		preconditionErr := ensureLambdaPreconditions(eachTarget.lambdaFn, conditionalLambdaTargetGroup)
		if preconditionErr != nil {
			return errors.Wrapf(preconditionErr, "Failed to create precondition resources for Lambda TargetGroup")
		}

		// Priority is either user defined or the current slice index
		rulePriority := eachTarget.priority
		if rulePriority <= 0 {
			rulePriority = int(1 + eachIndex)
		}

		// Now create the rule that conditionally routes to this Lambda, in priority order...
		listenerRule := &gofelbv2.ListenerRule{
			Actions: []gofelbv2.ListenerRule_Action{
				{
					TargetGroupArn: gof.Ref(conditionalLambdaTargetGroupResName),
					Type:           "forward",
				},
			},
			Conditions:  eachTarget.conditions,
			ListenerArn: gof.Ref(defaultListenerResName),
			Priority:    rulePriority,
		}
		// Add the rule...
		listenerRuleResName := portScopedResourceName("ALBRule",
			eachTarget.lambdaFn.LogicalResourceName(),
			fmt.Sprintf("%d", eachIndex))

		// Add the resource
		listenerRule.AWSCloudFormationDependsOn = []string{conditionalLambdaTargetGroupResName}
		template.Resources[listenerRuleResName] = listenerRule
	}
	// Add any other CloudFormation resources, in any order
	for eachKey, eachResource := range albd.Resources {
		// All the secondary resources are dependencies for the ALB
		albd.alb.AWSCloudFormationDependsOn = append(albd.alb.AWSCloudFormationDependsOn,
			eachKey)
		template.Resources[eachKey] = eachResource
	}
	portOutputName := func(prefix string) string {
		return fmt.Sprintf("%s%d", prefix, albd.port)
	}
	albOutput := func(label string, value interface{}) gof.Output {
		return gof.Output{
			Description: fmt.Sprintf("%s (port: %d, protocol: %s)", label, albd.port, albd.protocol),
			Value:       value,
		}
	}
	// Add the output to the template
	template.Outputs[portOutputName("ApplicationLoadBalancerDNS")] = albOutput(
		"ALB DNSName",
		gof.GetAtt(albd.LogicalResourceName(), "DNSName"))

	template.Outputs[portOutputName("ApplicationLoadBalancerName")] = albOutput(
		"ALB Name",
		gof.GetAtt(albd.LogicalResourceName(), "LoadBalancerName"))

	template.Outputs[portOutputName("ApplicationLoadBalancerURL")] = albOutput(
		"ALB URL",
		gof.Join("", []string{
			strings.ToLower(albd.protocol),
			"://",
			gof.GetAtt(albd.LogicalResourceName(), "DNSName"),
			fmt.Sprintf(":%d", albd.port),
		}))

	return nil
}

// NewApplicationLoadBalancerDecorator returns an application load balancer
// decorator that allows one or more lambda functions to be marked
// as ALB targets
func NewApplicationLoadBalancerDecorator(alb *gofelbv2.LoadBalancer,
	port int,
	protocol string,
	defaultLambdaHandler *sparta.LambdaAWSInfo) (*ApplicationLoadBalancerDecorator, error) {
	return &ApplicationLoadBalancerDecorator{
		alb:                  alb,
		port:                 port,
		protocol:             protocol,
		defaultLambdaHandler: defaultLambdaHandler,
		targets:              make([]*targetGroupEntry, 0),
		Resources:            make(map[string]gof.Resource),
	}, nil
}
