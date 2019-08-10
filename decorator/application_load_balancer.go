package decorator

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	sparta "github.com/mweagle/Sparta"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type targetGroupEntry struct {
	conditions *gocf.ElasticLoadBalancingV2ListenerRuleRuleConditionList
	lambdaFn   *sparta.LambdaAWSInfo
}

// ApplicationLoadBalancerDecorator is an instance of a service decorator that
// handles registering Lambda functions with an Application Load Balancer.
type ApplicationLoadBalancerDecorator struct {
	alb                  *gocf.ElasticLoadBalancingV2LoadBalancer
	port                 int64
	protocol             string
	defaultLambdaHandler *sparta.LambdaAWSInfo
	targets              []*targetGroupEntry
	Resources            map[string]gocf.ResourceProperties
}

// LogicalResourceName returns the CloudFormation resource name of the primary
// ALB
func (albd *ApplicationLoadBalancerDecorator) LogicalResourceName() string {
	return sparta.CloudFormationResourceName("ELBv2Resource", "ELBv2Resource")
}

// AddConditionalEntry adds a new lambda target that is conditionally routed
// to depending on the condition value.
func (albd *ApplicationLoadBalancerDecorator) AddConditionalEntry(condition gocf.ElasticLoadBalancingV2ListenerRuleRuleCondition,
	lambdaFn *sparta.LambdaAWSInfo) *ApplicationLoadBalancerDecorator {

	// Add a version resource to the lambda so that we target that resource...
	albd.targets = append(albd.targets, &targetGroupEntry{
		conditions: &gocf.ElasticLoadBalancingV2ListenerRuleRuleConditionList{condition},
		lambdaFn:   lambdaFn,
	})
	return albd
}

// AddMultiConditionalEntry adds a new lambda target that is conditionally routed
// to depending on the multi condition value.
func (albd *ApplicationLoadBalancerDecorator) AddMultiConditionalEntry(conditions *gocf.ElasticLoadBalancingV2ListenerRuleRuleConditionList,
	lambdaFn *sparta.LambdaAWSInfo) *ApplicationLoadBalancerDecorator {

	// Add a version resource to the lambda so that we target that resource...
	albd.targets = append(albd.targets, &targetGroupEntry{
		conditions: conditions,
		lambdaFn:   lambdaFn,
	})
	return albd
}

// DecorateService satisfies the ServiceDecoratorHookHandler interface
func (albd *ApplicationLoadBalancerDecorator) DecorateService(context map[string]interface{},
	serviceName string,
	template *gocf.Template,
	S3Bucket string,
	S3Key string,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *logrus.Logger) error {

	portScopedResourceName := func(prefix string, parts ...string) string {
		return sparta.CloudFormationResourceName(fmt.Sprintf("%s%d", prefix, albd.port),
			parts...)
	}

	////////////////////////////////////////////////////////////////////////////
	// Closure to manage the permissions, version, and alias resources needed
	// for each lambda target group
	//
	visitedLambdaFuncs := make(map[string]bool)
	ensureLambdaPreconditions := func(lambdaFn *sparta.LambdaAWSInfo, dependentResource *gocf.Resource) error {
		_, exists := visitedLambdaFuncs[lambdaFn.LogicalResourceName()]
		if exists {
			return nil
		}
		// Add the lambda permission
		albPermissionResourceName := portScopedResourceName("ALBPermission", lambdaFn.LogicalResourceName())
		lambdaInvokePermission := &gocf.LambdaPermission{
			Action:       gocf.String("lambda:InvokeFunction"),
			FunctionName: gocf.GetAtt(lambdaFn.LogicalResourceName(), "Arn"),
			Principal:    gocf.String(sparta.ElasticLoadBalancingPrincipal),
		}
		template.AddResource(albPermissionResourceName, lambdaInvokePermission)
		// The stable alias resource and unstable, retained version resource
		aliasResourceName := portScopedResourceName("ALBAlias", lambdaFn.LogicalResourceName())
		versionResourceName := portScopedResourceName("ALBVersion", lambdaFn.LogicalResourceName(), buildID)

		versionResource := &gocf.LambdaVersion{
			FunctionName: gocf.GetAtt(lambdaFn.LogicalResourceName(), "Arn").String(),
		}
		lambdaVersionRes := template.AddResource(versionResourceName, versionResource)
		lambdaVersionRes.DeletionPolicy = "Retain"

		// Add the alias that binds the lambda to the version...
		aliasResource := &gocf.LambdaAlias{
			FunctionVersion: gocf.GetAtt(versionResourceName, "Version").String(),
			FunctionName:    gocf.Ref(lambdaFn.LogicalResourceName()).String(),
			Name:            gocf.String("live"),
		}
		template.AddResource(aliasResourceName, aliasResource)
		// One time only
		dependentResource.DependsOn = append(dependentResource.DependsOn,
			albPermissionResourceName,
			versionResourceName,
			aliasResourceName)
		visitedLambdaFuncs[lambdaFn.LogicalResourceName()] = true
		return nil
	}

	////////////////////////////////////////////////////////////////////////////
	// START
	//
	// Add the alb. We'll link each target group inside the loop...
	albRes := template.AddResource(albd.LogicalResourceName(), albd.alb)
	defaultListenerResName := portScopedResourceName("ALBListener", "DefaultListener")
	defaultTargetGroupResName := portScopedResourceName("ALBDefaultTarget", albd.defaultLambdaHandler.LogicalResourceName())

	// Create the default lambda target group...
	defaultTargetGroupRes := &gocf.ElasticLoadBalancingV2TargetGroup{
		TargetType: gocf.String("lambda"),
		Targets: &gocf.ElasticLoadBalancingV2TargetGroupTargetDescriptionList{
			gocf.ElasticLoadBalancingV2TargetGroupTargetDescription{
				ID: gocf.GetAtt(albd.defaultLambdaHandler.LogicalResourceName(), "Arn").String(),
			},
		},
	}
	// Add it...
	targetGroupRes := template.AddResource(defaultTargetGroupResName, defaultTargetGroupRes)

	// Then create the ELB listener with the default entry. We'll add the conditional
	// lambda targets after this...
	listenerRes := &gocf.ElasticLoadBalancingV2Listener{
		LoadBalancerArn: gocf.Ref(albd.LogicalResourceName()).String(),
		Port:            gocf.Integer(albd.port),
		Protocol:        gocf.String(albd.protocol),
		DefaultActions: &gocf.ElasticLoadBalancingV2ListenerActionList{
			gocf.ElasticLoadBalancingV2ListenerAction{
				TargetGroupArn: gocf.Ref(defaultTargetGroupResName).String(),
				Type:           gocf.String("forward"),
			},
		},
	}
	defaultListenerRes := template.AddResource(defaultListenerResName, listenerRes)
	defaultListenerRes.DependsOn = append(defaultListenerRes.DependsOn, defaultTargetGroupResName)

	// Make sure this is all hooked up
	ensureErr := ensureLambdaPreconditions(albd.defaultLambdaHandler, targetGroupRes)
	if ensureErr != nil {
		return errors.Wrapf(ensureErr, "Failed to create precondition resources for Lambda TargetGroup")
	}
	// Finally, ensure that each lambdaTarget has a single InvokePermission permission
	// set so that the ALB can actually call them...
	for eachIndex, eachTarget := range albd.targets {
		// Create a new TargetGroup for this lambda function
		conditionalLambdaTargetGroupResName := portScopedResourceName("ALBTargetCond",
			eachTarget.lambdaFn.LogicalResourceName())
		conditionalLambdaTargetGroup := &gocf.ElasticLoadBalancingV2TargetGroup{
			TargetType: gocf.String("lambda"),
			Targets: &gocf.ElasticLoadBalancingV2TargetGroupTargetDescriptionList{
				gocf.ElasticLoadBalancingV2TargetGroupTargetDescription{
					ID: gocf.GetAtt(eachTarget.lambdaFn.LogicalResourceName(), "Arn").String(),
				},
			},
		}
		// Add it...
		targetGroupRes := template.AddResource(conditionalLambdaTargetGroupResName, conditionalLambdaTargetGroup)
		// Create the stable alias resource resource....
		preconditionErr := ensureLambdaPreconditions(eachTarget.lambdaFn, targetGroupRes)
		if preconditionErr != nil {
			return errors.Wrapf(preconditionErr, "Failed to create precondition resources for Lambda TargetGroup")
		}

		// The ALB depends on it...
		//defaultListenerRes.DependsOn = append(defaultListenerRes.DependsOn, conditionalLambdaTargetGroupResName)

		// Now create the rule that conditionally routes to this Lambda, in priority order...
		listenerRule := &gocf.ElasticLoadBalancingV2ListenerRule{
			Actions: &gocf.ElasticLoadBalancingV2ListenerRuleActionList{
				gocf.ElasticLoadBalancingV2ListenerRuleAction{
					TargetGroupArn: gocf.Ref(conditionalLambdaTargetGroupResName).String(),
					Type:           gocf.String("forward"),
				},
			},
			Conditions:  eachTarget.conditions,
			ListenerArn: gocf.Ref(defaultListenerResName).String(),
			Priority:    gocf.Integer(int64(1 + eachIndex)),
		}
		// Add the rule...
		listenerRuleResName := portScopedResourceName("ALBRule",
			eachTarget.lambdaFn.LogicalResourceName(),
			fmt.Sprintf("%d", eachIndex))

		// Add the resource
		listenerRes := template.AddResource(listenerRuleResName, listenerRule)
		listenerRes.DependsOn = append(listenerRes.DependsOn, conditionalLambdaTargetGroupResName)
	}
	// Add any other CloudFormation resources, in any order
	for eachKey, eachResource := range albd.Resources {
		template.AddResource(eachKey, eachResource)
		// All the secondary resources are dependencies for the ALB
		albRes.DependsOn = append(albRes.DependsOn, eachKey)
	}
	portOutputName := func(prefix string) string {
		return fmt.Sprintf("%s%d", prefix, albd.port)
	}
	// Add the output to the template
	template.Outputs[portOutputName("ApplicationLoadBalancerDNS")] = &gocf.Output{
		Description: "DNS value of the ALB",
		Value:       gocf.GetAtt(albd.LogicalResourceName(), "DNSName"),
	}
	template.Outputs[portOutputName("ApplicationLoadBalancerName")] = &gocf.Output{
		Description: "Name of the ALB",
		Value:       gocf.GetAtt(albd.LogicalResourceName(), "LoadBalancerName"),
	}
	template.Outputs[portOutputName("ApplicationLoadBalancerURL")] = &gocf.Output{
		Description: "URL value of the ALB",
		Value: gocf.Join("",
			gocf.String(strings.ToLower(albd.protocol)),
			gocf.String("://"),
			gocf.GetAtt(albd.LogicalResourceName(), "DNSName"),
			gocf.String(fmt.Sprintf(":%d", albd.port))),
	}
	return nil
}

// NewApplicationLoadBalancerDecorator returns an application load balancer
// decorator that allows one or more lambda functions to be marked
// as ALB targets
func NewApplicationLoadBalancerDecorator(alb *gocf.ElasticLoadBalancingV2LoadBalancer,
	port int64,
	protocol string,
	defaultLambdaHandler *sparta.LambdaAWSInfo) (*ApplicationLoadBalancerDecorator, error) {
	return &ApplicationLoadBalancerDecorator{
		alb:                  alb,
		port:                 port,
		protocol:             protocol,
		defaultLambdaHandler: defaultLambdaHandler,
		targets:              make([]*targetGroupEntry, 0),
		Resources:            make(map[string]gocf.ResourceProperties),
	}, nil
}
