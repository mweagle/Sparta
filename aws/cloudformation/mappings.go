package cloudformation

// Constants for https://docs.aws.amazon.com/general/latest/gr/rande.html

const (
	// Endpoint property
	Endpoint = "endpoint"
	// Protocol property
	Protocol = "protocol"
	// HostedZoneID property
	HostedZoneID = "hostedZoneID"
)

// APIGatewayMapping is the mapping for APIGateway settings
var APIGatewayMapping = map[string]interface{}{
	"us-east-2": map[string]string{
		Endpoint:     "apigateway.us-east-2.amazonaws.com",
		Protocol:     "HTTPS",
		HostedZoneID: "ZOJJZC49E0EPZ",
	},
	"us-east-1": map[string]string{
		Endpoint:     "apigateway.us-east-1.amazonaws.com",
		Protocol:     "HTTPS",
		HostedZoneID: "Z1UJRXOUMOOFQ8",
	},
	"us-west-1": map[string]string{
		Endpoint:     "apigateway.us-west-1.amazonaws.com",
		Protocol:     "HTTPS",
		HostedZoneID: "Z2MUQ32089INYE",
	},
	"us-west-2": map[string]string{
		Endpoint:     "apigateway.us-west-2.amazonaws.com",
		Protocol:     "HTTPS",
		HostedZoneID: "Z2OJLYMUO9EFXC",
	},
	"ap-south-1": map[string]string{
		Endpoint:     "apigateway.ap-south-1.amazonaws.com",
		Protocol:     "HTTPS",
		HostedZoneID: "Z3VO1THU9YC4UR",
	},
	"ap-northeast-3": map[string]string{
		Endpoint:     "apigateway.ap-northeast-3.amazonaws.com",
		Protocol:     "HTTPS",
		HostedZoneID: "Z2YQB5RD63NC85",
	},
	"ap-northeast-2": map[string]string{
		Endpoint:     "apigateway.ap-northeast-2.amazonaws.com",
		Protocol:     "HTTPS",
		HostedZoneID: "Z20JF4UZKIW1U8",
	},
	"ap-southeast-1": map[string]string{
		Endpoint:     "apigateway.ap-southeast-1.amazonaws.com",
		Protocol:     "HTTPS",
		HostedZoneID: "ZL327KTPIQFUL",
	},
	"ap-southeast-2": map[string]string{
		Endpoint:     "apigateway.ap-southeast-2.amazonaws.com",
		Protocol:     "HTTPS",
		HostedZoneID: "Z2RPCDW04V8134",
	},
	"ap-northeast-1": map[string]string{
		Endpoint:     "apigateway.ap-northeast-1.amazonaws.com",
		Protocol:     "HTTPS",
		HostedZoneID: "Z1YSHQZHG15GKL",
	},
	"ca-central-1": map[string]string{
		Endpoint:     "apigateway.ca-central-1.amazonaws.com",
		Protocol:     "HTTPS",
		HostedZoneID: "Z19DQILCV0OWEC",
	},
	"cn-north-1": map[string]string{
		Endpoint:     "apigateway.cn-north-1.amazonaws.com.cn",
		Protocol:     "HTTPS",
		HostedZoneID: "",
	},
	"cn-northwest-1": map[string]string{
		Endpoint:     "apigateway.cn-northwest-1.amazonaws.com.cn",
		Protocol:     "HTTPS",
		HostedZoneID: "",
	},
	"eu-central-1": map[string]string{
		Endpoint:     "apigateway.eu-central-1.amazonaws.com",
		Protocol:     "HTTPS",
		HostedZoneID: "Z1U9ULNL0V5AJ3",
	},
	"eu-west-1": map[string]string{
		Endpoint:     "apigateway.eu-west-1.amazonaws.com",
		Protocol:     "HTTPS",
		HostedZoneID: "ZLY8HYME6SFDD",
	},
	"eu-west-2": map[string]string{
		Endpoint:     "apigateway.eu-west-2.amazonaws.com",
		Protocol:     "HTTPS",
		HostedZoneID: "ZJ5UAJN8Y3Z2Q",
	},
	"eu-west-3": map[string]string{
		Endpoint:     "apigateway.eu-west-3.amazonaws.com",
		Protocol:     "HTTPS",
		HostedZoneID: "Z3KY65QIEKYHQQ",
	},
	"eu-north-1": map[string]string{
		Endpoint:     "apigateway.eu-north-1.amazonaws.com",
		Protocol:     "HTTPS",
		HostedZoneID: "Z2YB950C88HT6D",
	},
	"sa-east-1": map[string]string{
		Endpoint:     "apigateway.sa-east-1.amazonaws.com",
		Protocol:     "HTTPS",
		HostedZoneID: "ZCMLWB8V5SYIT",
	},
}
