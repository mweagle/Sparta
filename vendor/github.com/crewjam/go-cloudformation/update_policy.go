package cloudformation

// UpdatePolicy represents UpdatePolicy Attribute
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-attribute-updatepolicy.html
type UpdatePolicy struct {
	AutoScalingRollingUpdate   *UpdatePolicyAutoScalingRollingUpdate   `json:"AutoScalingRollingUpdate,omitempty"`
	AutoScalingScheduledAction *UpdatePolicyAutoScalingScheduledAction `json:"AutoScalingScheduledAction,omitempty"`
}

// UpdatePolicyAutoScalingRollingUpdate represents an AutoScalingRollingUpdate
//
// You can use the AutoScalingRollingUpdate policy to specify how AWS CloudFormation handles rolling updates for a particular resource.
type UpdatePolicyAutoScalingRollingUpdate struct {
	// The maximum number of instances that are terminated at a given time.
	MaxBatchSize *IntegerExpr `json:"MaxBatchSize,omitempty"`

	// The minimum number of instances that must be in service within the Auto Scaling group while obsolete instances are being terminated.
	MinInstancesInService *IntegerExpr `json:"MinInstancesInService,omitempty"`

	// The percentage of instances in an Auto Scaling rolling update that must signal success for an update to succeed. You can specify a value from 0 to 100. AWS CloudFormation rounds to the nearest tenth of a percent. For example, if you update five instances with a minimum successful percentage of 50, three instances must signal success.
	// If an instance doesn't send a signal within the specified pause time, AWS CloudFormation assumes the instance did not successfully update.
	// If you specify this property, you must enable the WaitOnResourceSignals property.
	MinSuccessfulInstancesPercent *IntegerExpr `json:"MinSuccessfulInstancesPercent,omitempty"`

	// The amount of time to pause after AWS CloudFormation makes a change to the Auto Scaling group before making additional changes to a resource. For example, the amount of time to pause before adding or removing instances when scaling up or terminating instances in an Auto Scaling group.
	//
	// If you specify the WaitOnResourceSignals property, the amount of time to wait until the Auto Scaling group receives the required number of valid signals. If the pause time is exceeded before theAuto Scaling group receives the required number of signals, the update times out and fails. For best results, specify a period of time that gives your instances plenty of time to get up and running. In the event of a rollback, a shorter pause time can cause update rollback failures.
	//
	// The value must be in ISO8601 duration format, in the form: "PT#H#M#S", where each # is the number of hours, minutes, and/or seconds, respectively. The maximum amount of time that can be specified for the pause time is one hour ("PT1H").
	//
	// Default: PT0S (zero seconds). If the WaitOnResourceSignals property is set to true, the default is PT5M.
	PauseTime *StringExpr `json:"PauseTime,omitempty"`

	// The Auto Scaling processes to suspend during a stack update. Suspending processes is useful when you don't want Auto Scaling to potentially interfere with a stack update. For example, you can suspend process so that no alarms are triggered during an update. For valid values, see SuspendProcesses in the Auto Scaling API Reference.
	SuspendProcesses *StringListExpr `json:"SuspendProcesses,omitempty"`

	// Indicates whether the Auto Scaling group waits on signals during an update. AWS CloudFormation suspends the update of an Auto Scaling group after any new Amazon EC2 instances are launched into the group. AWS CloudFormation must receive a signal from each new instance within the specified pause time before AWS CloudFormation continues the update. You can use the cfn-signal helper script or SignalResource API to signal the Auto Scaling group. This property is useful when you want to ensure instances have completed installing and configuring applications before the Auto Scaling group update proceeds.
	WaitOnResourceSignals *BoolExpr `json:"WaitOnResourceSignals,omitempty"`
}

// UpdatePolicyAutoScalingScheduledAction represents an AutoScalingScheduledAction object
//
// When the AWS::AutoScaling::AutoScalingGroup resource has an associated scheduled action, the AutoScalingScheduledAction policy describes how AWS CloudFormation handles updates for the MinSize, MaxSize, and DesiredCapacity properties..
//
// With scheduled actions, the group size properties (minimum size, maximum size, and desired capacity) of an Auto Scaling group can change at any time. Whenever you update a stack with an Auto Scaling group and scheduled action, AWS CloudFormation always sets group size property values of your Auto Scaling group to the values that are defined in the AWS::AutoScaling::AutoScalingGroup resource of your template, even if a scheduled action is in effect. However, you might not want AWS CloudFormation to change any of the group size property values, such as when you have a scheduled action in effect. You can use the AutoScalingScheduledAction update policy to prevent AWS CloudFormation from changing the min size, max size, or desired capacity unless you modified the individual values in your template.
//
type UpdatePolicyAutoScalingScheduledAction struct {
	// During a stack update, indicates whether AWS CloudFormation ignores any group size property differences between your current Auto Scaling group and the Auto Scaling group that is described in the AWS::AutoScaling::AutoScalingGroup resource of your template. However, if you modified any group size property values in your template, AWS CloudFormation will always use the modified values and update your Auto Scaling group.
	IgnoreUnmodifiedGroupSizeProperties *BoolExpr `json:"IgnoreUnmodifiedGroupSizeProperties,omitempty"`
}
