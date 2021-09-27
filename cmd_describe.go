//go:build !lambdabinary
// +build !lambdabinary

package sparta

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"text/template"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// workflowHooksDescriptionNodes returns the set of []*DescriptionInfo
// entries that summarizes the WorkflowNodes
func workflowHooksDescriptionNodes(serviceName string, hooks *WorkflowHooks) ([]*DescriptionInfo, error) {
	if hooks == nil {
		return nil, nil
	}
	workflowDescInfo := make([]*DescriptionInfo, 0)
	for _, eachServiceDecorator := range hooks.ServiceDecorators {
		describable, isDescribable := eachServiceDecorator.(Describable)
		if isDescribable {
			descInfos, descInfosErr := describable.Describe(serviceName)
			if descInfosErr != nil {
				return nil, descInfosErr
			}
			workflowDescInfo = append(workflowDescInfo, descInfos)
		}
	}
	return workflowDescInfo, nil
}

// Describe produces a graphical representation of a service's Lambda and data sources.  Typically
// automatically called as part of a compiled golang binary via the `describe` command
// line option.
func Describe(serviceName string,
	serviceDescription string,
	lambdaAWSInfos []*LambdaAWSInfo,
	api APIGateway,
	site *S3Site,
	s3BucketName string,
	buildTags string,
	linkerFlags string,
	outputWriter io.Writer,
	workflowHooks *WorkflowHooks,
	logger *zerolog.Logger) error {

	// Multiwriter
	templateFile, templateFileErr := templateOutputFile(optionsProvision.OutputDir,
		serviceName)
	if templateFileErr != nil {
		return templateFileErr
	}

	var cloudFormationTemplate bytes.Buffer
	multiWriter := io.MultiWriter(templateFile, &cloudFormationTemplate)

	buildErr := Build(true,
		serviceName,
		serviceDescription,
		lambdaAWSInfos,
		api,
		site,
		false,
		"BUILD_ID",
		"",
		ScratchDirectory,
		buildTags,
		linkerFlags,
		multiWriter,
		workflowHooks,
		logger)
	closeErr := templateFile.Close()
	if closeErr != nil {
		logger.Warn().
			Err(closeErr).
			Msg("Failed to close template file handle")
	}
	if buildErr != nil {
		return buildErr
	}

	tmpl, err := template.New("description").Parse(embeddedMustString("resources/describe/template.html"))
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

	fullIconPath := func(descriptionNode *DescriptionIcon) string {
		// Use an empty PNG if we don't have an image
		if descriptionNode == nil {
			// Because the style uses data(image) we need to ensure that
			// empty nodes have some sort of image, else the Cytoscape JS
			// won't render
			return "AWS-Architecture-Assets/Default/empty-image.png"
		}
		return fmt.Sprintf("AWS-Architecture-Assets/%s/%s",
			descriptionNode.Category,
			descriptionNode.Name)
	}

	// Setup the root object
	writeErr := describer.writeNodeWithParent(serviceName,
		nodeColorService,
		fullIconPath(&DescriptionIcon{
			Category: "Res_Management-Governance",
			Name:     "Res_48_Light/Res_AWS-CloudFormation_Stack_48_Light.png",
		}),
		"",
		labelWeightBold)
	if writeErr != nil {
		return writeErr
	}

	parentMap := make(map[string]bool)
	writeNodes := func(parent string, descriptionNodes []*DescriptionTriplet) error {
		if parent != "" {
			_, exists := parentMap[parent]
			if !exists {
				writeErr = describer.writeNodeWithParent(parent,
					"#FF0000",
					fullIconPath(nil),
					"",
					labelWeightBold)
				if writeErr != nil {
					return writeErr
				}
				parentMap[parent] = true
			}
		}

		for _, eachDescNode := range descriptionNodes {
			descDisplayInfo := eachDescNode.DisplayInfo
			if descDisplayInfo == nil {
				descDisplayInfo = &DescriptionDisplayInfo{}
			}
			writeErr = describer.writeNodeWithParent(eachDescNode.SourceNodeName,
				descDisplayInfo.SourceNodeColor,
				fullIconPath(descDisplayInfo.SourceIcon),
				parent,
				labelWeightNormal)
			if writeErr != nil {
				return writeErr
			}
			writeErr = describer.writeEdge(eachDescNode.SourceNodeName,
				eachDescNode.TargetNodeName,
				eachDescNode.ArcLabel)
			if writeErr != nil {
				return writeErr
			}
		}
		return nil
	}

	for _, eachLambda := range lambdaAWSInfos {
		descriptionNodes, descriptionNodesErr := eachLambda.Description(serviceName)
		if descriptionNodesErr != nil {
			return descriptionNodesErr
		}
		writeErr := writeNodes("Lambdas", descriptionNodes)
		if writeErr != nil {
			return writeErr
		}
	}
	// The API needs to know how to describe itself. So for that it needs an object that
	// encapsulates writing the nodes and links...so let's go ahead
	// and make that object, then supply it to the Describe interface function

	// API?
	if nil != api {
		descriptionInfo, descriptionInfoErr := api.Describe(serviceName)
		if descriptionInfoErr != nil {
			return descriptionInfoErr
		}
		writeErr := writeNodes(descriptionInfo.Name, descriptionInfo.Nodes)
		if writeErr != nil {
			return writeErr
		}
	}
	// What about everything else...
	workflowDescription, workflowDescriptionErr := workflowHooksDescriptionNodes(serviceName, workflowHooks)
	if workflowDescriptionErr != nil {
		return workflowDescriptionErr
	}
	for _, eachWorkflowDesc := range workflowDescription {
		groupName := eachWorkflowDesc.Name
		if groupName == "" {
			groupName = "WorkflowHooks"
		}
		workflowDescriptionErr = writeNodes(groupName, eachWorkflowDesc.Nodes)
		if workflowDescriptionErr != nil {
			return workflowDescriptionErr
		}
	}

	// Write it out...
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
