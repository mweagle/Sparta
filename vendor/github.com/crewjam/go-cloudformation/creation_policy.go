package cloudformation

// CreationPolicy represents CreationPolicy Attribute
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-attribute-creationpolicy.html
type CreationPolicy struct {
	ResourceSignal *CreationPolicyResourceSignal `json:"ResourceSignal,omitempty"`
}

// CreationPolicyResourceSignal represents a CreationPolicy ResourceSignal
//
// see http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-attribute-creationpolicy.html
type CreationPolicyResourceSignal struct {
	// The number of success signals AWS CloudFormation must receive before it sets the resource status as CREATE_COMPLETE. If the resource receives a failure signal or doesn't receive the specified number of signals before the timeout period expires, the resource creation fails and AWS CloudFormation rolls the stack back.
	Count *IntegerExpr `json:"Count,omitempty"`

	// The length of time that AWS CloudFormation waits for the number of signals that was specified in the Count property. The timeout period starts after AWS CloudFormation starts creating the resource, and the timeout expires no sooner than the time you specify but can occur shortly thereafter. The maximum time that you can specify is 12 hours.
	//
	// The value must be in ISO8601 duration format, in the form: "PT#H#M#S", where each # is the number of hours, minutes, and seconds, respectively. For best results, specify a period of time that gives your instances plenty of time to get up and running. A shorter timeout can cause a rollback.
	Timeout *StringExpr `json:"Timeout,omitempty"`
}
