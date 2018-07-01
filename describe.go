// +build !lambdabinary

package sparta

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"text/template"

	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	"github.com/sirupsen/logrus"
)

const (
	nodeColorService     = "#720502"
	nodeColorEventSource = "#BF2803"
	nodeColorLambda      = "#F35B05"
	nodeColorAPIGateway  = "#06B5F5"
	nodeNameAPIGateway   = "API Gateway"
)

type cytoscapeData struct {
	ID               string `json:"id"`
	Image            string `json:"image"`
	BackgroundColor  string `json:"backgroundColor,omitempty"`
	Source           string `json:"source,omitempty"`
	Target           string `json:"target,omitempty"`
	Label            string `json:"label,omitempty"`
	DegreeCentrality int    `json:"degreeCentrality"`
}
type cytoscapeNode struct {
	Data    cytoscapeData `json:"data"`
	Classes string        `json:"classes,omitempty"`
}
type templateResource struct {
	KeyName string
	Data    string
}

func cytoscapeNodeID(rawData interface{}) (string, error) {
	bytes, bytesErr := json.Marshal(rawData)
	if bytesErr != nil {
		return "", bytesErr
	}
	hash := sha1.New()
	_, writeErr := hash.Write(bytes)
	if writeErr != nil {
		return "", writeErr
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func writeNode(nodes *[]*cytoscapeNode,
	nodeName string,
	nodeColor string,
	nodeImage string,
	logger *logrus.Logger) {
	nodeID, _ := cytoscapeNodeID(nodeName)
	appendNode := &cytoscapeNode{
		Data: cytoscapeData{
			ID:    nodeID,
			Label: strings.Trim(nodeName, "\""),
		},
	}
	if nodeImage != "" {
		resourceItem := templateResourceForKey(nodeImage, logger)
		if resourceItem != nil {
			appendNode.Data.Image = fmt.Sprintf("data:image/svg+xml;base64,%s",
				base64.StdEncoding.EncodeToString([]byte(resourceItem.Data)))
		}
	}
	*nodes = append(*nodes, appendNode)

}

func writeLink(nodes *[]*cytoscapeNode, fromNode string, toNode string, label string) {
	nodeSource, _ := cytoscapeNodeID(fromNode)
	nodeTarget, _ := cytoscapeNodeID(toNode)

	*nodes = append(*nodes, &cytoscapeNode{
		Data: cytoscapeData{
			ID:     fmt.Sprintf("%d", rand.Uint64()),
			Source: nodeSource,
			Target: nodeTarget,
			Label:  label,
		},
	})
}
func templateResourceForKey(resourceKeyName string, logger *logrus.Logger) *templateResource {
	var resource *templateResource
	resourcePath := fmt.Sprintf("/resources/describe/%s",
		strings.TrimLeft(resourceKeyName, "/"))
	data, dataErr := _escFSString(false, resourcePath)
	if dataErr == nil {
		keyParts := strings.Split(resourcePath, "/")
		keyName := keyParts[len(keyParts)-1]
		resource = &templateResource{
			KeyName: keyName,
			Data:    data,
		}
		logger.WithFields(logrus.Fields{
			"Path":    resourcePath,
			"KeyName": keyName,
		}).Debug("Embedded resource")

	} else {
		logger.WithFields(logrus.Fields{
			"Path": resourcePath,
		}).Warn("Failed to embed resource")
	}
	return resource
}
func templateResourcesForKeys(resourceKeyNames []string, logger *logrus.Logger) []*templateResource {
	resources := make([]*templateResource, 0)

	for _, eachKey := range resourceKeyNames {
		loadedResource := templateResourceForKey(eachKey, logger)
		if loadedResource != nil {
			resources = append(resources, loadedResource)
		}
	}
	return resources
}

func templateCSSFiles(logger *logrus.Logger) []*templateResource {
	cssFiles := []string{"bootstrap-4.0.0/dist/css/bootstrap.min.css",
		"highlight.js/styles/xcode.css",
	}
	return templateResourcesForKeys(cssFiles, logger)
}

func templateJSFiles(logger *logrus.Logger) []*templateResource {
	jsFiles := []string{"jquery/jquery-3.3.1.min.js",
		"popper/popper.min.js",
		"bootstrap-4.0.0/dist/js/bootstrap.min.js",
		"highlight.js/highlight.pack.js",
		"dagre-0.8.2/dagre-0.8.2/dist/dagre.js",
		"cytoscapejs/cytoscape.js",
		"cytoscape.js-dagre-2.2.1/cytoscape.js-dagre-2.2.1/cytoscape-dagre.js",
		"sparta.js",
	}
	return templateResourcesForKeys(jsFiles, logger)
}

func templateImageMap(logger *logrus.Logger) map[string]string {
	images := []string{"SpartaHelmet256.png",
		"AWSIcons/Compute/Compute_AWSLambda_LambdaFunction.svg",
		"AWSIcons/Management Tools/ManagementTools_AWSCloudFormation.svg",
	}
	resources := templateResourcesForKeys(images, logger)
	imageMap := make(map[string]string)
	for _, eachResource := range resources {
		imageMap[eachResource.KeyName] = base64.StdEncoding.EncodeToString([]byte(eachResource.Data))
	}
	return imageMap
}

// TODO - this should really be smarter, including
// looking at the referred resource to understand it's
// type
func iconForAWSResource(rawEmitter interface{}) string {
	jsonBytes, _ := json.Marshal(rawEmitter)
	canonicalRaw := strings.ToLower(string(jsonBytes))
	if strings.Contains(canonicalRaw, "dynamodb") {
		return "AWSIcons/Database/Database_AmazonDynamoDB.svg"
	}
	if strings.Contains(canonicalRaw, "sqs") {
		return "AWSIcons/Messaging/Messaging_AmazonSQS.svg"
	}
	if strings.Contains(canonicalRaw, "sns") {
		return "AWSIcons/Messaging/Messaging_AmazonSNS_topic.svg"
	}
	if strings.Contains(canonicalRaw, "cloudwatch") {
		return "AWSIcons/Management Tools/ManagementTools_AmazonCloudWatch.svg"
	}
	if strings.Contains(canonicalRaw, "kinesis") {
		return "AWSIcons/Analytics/Analytics_AmazonKinesis.svg"
	}
	if strings.Contains(canonicalRaw, "s3") {
		return "AWSIcons/Storage/Storage_AmazonS3_bucket.svg"
	}
	return "AWSIcons/General/General_AWScloud.svg"
}

// Describe produces a graphical representation of a service's Lambda and data sources.  Typically
// automatically called as part of a compiled golang binary via the `describe` command
// line option.
func Describe(serviceName string,
	serviceDescription string,
	lambdaAWSInfos []*LambdaAWSInfo,
	api *API,
	s3Site *S3Site,
	s3BucketName string,
	buildTags string,
	linkFlags string,
	outputWriter io.Writer,
	workflowHooks *WorkflowHooks,
	logger *logrus.Logger) error {

	validationErr := validateSpartaPreconditions(lambdaAWSInfos, logger)
	if validationErr != nil {
		return validationErr
	}

	var cloudFormationTemplate bytes.Buffer
	err := Provision(true,
		serviceName,
		serviceDescription,
		lambdaAWSInfos,
		api,
		s3Site,
		s3BucketName,
		false,
		false,
		"N/A",
		"",
		buildTags,
		linkFlags,
		&cloudFormationTemplate,
		workflowHooks,
		logger)
	if nil != err {
		return err
	}

	tmpl, err := template.New("description").Parse(_escFSMustString(false, "/resources/describe/template.html"))
	if err != nil {
		return errors.New(err.Error())
	}

	cytoscapeElements := make([]*cytoscapeNode, 0)

	// Instead of inline mermaid stuff, we're going to stuff raw
	// json through. We can also include AWS images in the icon
	// using base64/encoded:
	// Example: https://cytoscape.github.io/cytoscape.js-tutorial-demo/datasets/social.json
	// Use the "fancy" CSS:
	// https://github.com/cytoscape/cytoscape.js-tutorial-demo/blob/gh-pages/stylesheets/fancy.json
	// Which is dynamically updated at: https://cytoscape.github.io/cytoscape.js-tutorial-demo/

	// Setup the root object
	writeNode(&cytoscapeElements,
		serviceName,
		nodeColorService,
		"AWSIcons/Management Tools/ManagementTools_AWSCloudFormation_stack.svg",
		logger)

	for _, eachLambda := range lambdaAWSInfos {
		// Other cytoscape nodes
		// Create the node...
		writeNode(&cytoscapeElements,
			eachLambda.lambdaFunctionName(),
			nodeColorLambda,
			"AWSIcons/Compute/Compute_AWSLambda.svg",
			logger)
		writeLink(&cytoscapeElements,
			eachLambda.lambdaFunctionName(),
			serviceName,
			"")

		// Create permission & event mappings
		// functions declared in this
		for _, eachPermission := range eachLambda.Permissions {
			nodes, err := eachPermission.descriptionInfo()
			if nil != err {
				return err
			}

			for _, eachNode := range nodes {
				name := strings.TrimSpace(eachNode.Name)
				link := strings.TrimSpace(eachNode.Relation)
				// Style it to have the Amazon color
				nodeColor := eachNode.Color
				if "" == nodeColor {
					nodeColor = nodeColorEventSource
				}

				writeNode(&cytoscapeElements,
					name,
					nodeColor,
					iconForAWSResource(eachNode.Name),
					logger)
				writeLink(&cytoscapeElements,
					name,
					eachLambda.lambdaFunctionName(),
					link)
			}
		}
		for index, eachEventSourceMapping := range eachLambda.EventSourceMappings {
			dynamicArn := spartaCF.DynamicValueToStringExpr(eachEventSourceMapping.EventSourceArn)
			jsonBytes, jsonBytesErr := json.Marshal(dynamicArn)
			if jsonBytesErr != nil {
				jsonBytes = []byte(fmt.Sprintf("%s-EventSourceMapping[%d]",
					eachLambda.lambdaFunctionName(),
					index))
			}
			nodeName := string(jsonBytes)
			writeNode(&cytoscapeElements,
				nodeName,
				nodeColorEventSource,
				iconForAWSResource(dynamicArn),
				logger)
			writeLink(&cytoscapeElements,
				nodeName,
				eachLambda.lambdaFunctionName(),
				"")
		}
	}

	// API?
	if nil != api {
		// Create the APIGateway virtual node && connect it to the application
		writeNode(&cytoscapeElements,
			nodeNameAPIGateway,
			nodeColorAPIGateway,
			"AWSIcons/Application Services/ApplicationServices_AmazonAPIGateway.svg",
			logger)
		for _, eachResource := range api.resources {
			for eachMethod := range eachResource.Methods {
				// Create the PATH node
				var nodeName = fmt.Sprintf("%s - %s", eachMethod, eachResource.pathPart)
				writeNode(&cytoscapeElements,
					nodeName,
					nodeColorAPIGateway,
					"AWSIcons/General/General_Internet.svg",
					logger)
				writeLink(&cytoscapeElements,
					nodeNameAPIGateway,
					nodeName,
					"")
				writeLink(&cytoscapeElements,
					nodeName,
					eachResource.parentLambda.lambdaFunctionName(),
					"")
			}
		}
	}
	cytoscapeBytes, _ := json.MarshalIndent(cytoscapeElements, "", " ")
	params := struct {
		SpartaVersion          string
		ServiceName            string
		ServiceDescription     string
		CloudFormationTemplate string
		CSSFiles               []*templateResource
		JSFiles                []*templateResource
		ImageMap               map[string]string
		CytoscapeData          interface{}
	}{
		SpartaGitHash[0:8],
		serviceName,
		serviceDescription,
		cloudFormationTemplate.String(),
		templateCSSFiles(logger),
		templateJSFiles(logger),
		templateImageMap(logger),
		string(cytoscapeBytes),
	}
	return tmpl.Execute(outputWriter, params)
}
