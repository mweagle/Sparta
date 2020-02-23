// +build !lambdabinary

package sparta

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"
	"time"

	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Describe produces a graphical representation of a service's Lambda and data sources.  Typically
// automatically called as part of a compiled golang binary via the `describe` command
// line option.
func Describe(serviceName string,
	serviceDescription string,
	lambdaAWSInfos []*LambdaAWSInfo,
	api APIGateway,
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
	buildID, buildIDErr := provisionBuildID("none", logger)
	if buildIDErr != nil {
		buildID = fmt.Sprintf("%d", time.Now().Unix())
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
		buildID,
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

	// Setup the describer
	describer := descriptionWriter{
		nodes:  make([]*cytoscapeNode, 0),
		logger: logger,
	}

	// Instead of inline mermaid stuff, we're going to stuff raw
	// json through. We can also include AWS images in the icon
	// using base64/encoded:
	// Example: https://cytoscape.github.io/cytoscape.js-tutorial-demo/datasets/social.json
	// Use the "fancy" CSS:
	// https://github.com/cytoscape/cytoscape.js-tutorial-demo/blob/gh-pages/stylesheets/fancy.json
	// Which is dynamically updated at: https://cytoscape.github.io/cytoscape.js-tutorial-demo/

	// Setup the root object
	writeErr := describer.writeNode(serviceName,
		nodeColorService,
		"AWS-Architecture-Icons_SVG_20200131/SVG Light/Management & Governance/AWS-CloudFormation_Stack_light-bg.svg")
	if writeErr != nil {
		return writeErr
	}
	for _, eachLambda := range lambdaAWSInfos {
		// Other cytoscape nodes
		// Create the node...
		writeErr = describer.writeNode(eachLambda.lambdaFunctionName(),
			nodeColorLambda,
			"AWS-Architecture-Icons_SVG_20200131/SVG Light/Mobile/Amazon-API-Gateway_light-bg.svg")
		if writeErr != nil {
			return writeErr
		}
		writeErr = describer.writeEdge(eachLambda.lambdaFunctionName(),
			serviceName,
			"")
		if writeErr != nil {
			return writeErr
		}
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
				if nodeColor == "" {
					nodeColor = nodeColorEventSource
				}

				writeErr = describer.writeNode(name,
					nodeColor,
					iconForAWSResource(eachNode.Name))
				if writeErr != nil {
					return writeErr
				}
				writeErr = describer.writeEdge(
					name,
					eachLambda.lambdaFunctionName(),
					link)
				if writeErr != nil {
					return writeErr
				}
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
			writeErr = describer.writeNode(nodeName,
				nodeColorEventSource,
				iconForAWSResource(dynamicArn))
			if writeErr != nil {
				return writeErr
			}
			writeErr = describer.writeEdge(nodeName,
				eachLambda.lambdaFunctionName(),
				"")
			if writeErr != nil {
				return writeErr
			}
		}
	}
	// The API needs to know how to describe itself. So for that it needs an object that
	// encapsulates writing the nodes and links...so let's go ahead
	// and make that object, then supply it to the Describe interface function

	// API?
	if nil != api {
		// TODO - delegate
		writeErr := api.Describe(&describer)
		if writeErr != nil {
			return writeErr
		}
	}
	cytoscapeBytes, cytoscapeBytesErr := json.MarshalIndent(describer.nodes, "", " ")
	if cytoscapeBytesErr != nil {
		return errors.Wrapf(cytoscapeBytesErr, "Failed to marshal cytoscape data")
	}
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
