package decorator

const (
	// KeyCloudMapRef is the name of the property that contains the `Ref`
	// output from the CloudFormation resource
	KeyCloudMapRef = "Ref"
	// KeyCloudMapType is the name of the property that contains the CloudFormation
	// resource type of the published resource
	KeyCloudMapType = "Type"
	// KeyCloudMapResourceName is the logical CloudFormation resource name
	KeyCloudMapResourceName = "ResourceName"
)

const (
	// EnvVarCloudMapNamespaceID contains the CloudMap namespaceID that was
	// registered in this stack. This serviceID enables your lambda function
	// to call the https://docs.aws.amazon.com/sdk-for-go/api/service/servicediscovery
	// for listing or discovering instanes
	EnvVarCloudMapNamespaceID = "SPARTA_CLOUDMAP_NAMESPACE_ID"
	// EnvVarCloudMapServiceID contains the CloudMap serviceID that was
	// registered in this stack. This serviceID enables your lambda function
	// to call the https://docs.aws.amazon.com/sdk-for-go/api/service/servicediscovery
	// for listing or discovering instanes
	EnvVarCloudMapServiceID = "SPARTA_CLOUDMAP_SERVICE_ID"
)
