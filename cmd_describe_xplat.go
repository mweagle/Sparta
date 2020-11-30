package sparta

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Utility discovery information that is necessary for compilation
// in both local and AWS Binary mode

const (
	nodeColorService     = "#720502"
	nodeColorEventSource = "#BF2803"
	nodeColorLambda      = "#F35B05"
	nodeColorAPIGateway  = "#06B5F5"
	nodeNameAPIGateway   = "API Gateway"

	labelWeightNormal = "normal"
	labelWeightBold   = "bolder"
	defaultImagePath  = "/resources/describe/AWS-Architecture-Icons_PNG/PNG Light/_General/General@4x.png"
)

// This is the `go` type that's shuttled through the JSON data
// and parsed by the sparta.js script that's executed in the browser
type cytoscapeData struct {
	ID               string `json:"id"`
	Parent           string `json:"parent"`
	Image            string `json:"image"`
	Width            string `json:"width"`
	Height           string `json:"height"`
	BackgroundColor  string `json:"backgroundColor,omitempty"`
	Source           string `json:"source,omitempty"`
	Target           string `json:"target,omitempty"`
	Label            string `json:"label,omitempty"`
	LabelWeight      string `json:"labelWeight,omitempty"`
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

type descriptionWriter struct {
	nodes  []*cytoscapeNode
	logger *zerolog.Logger
}

func (dw *descriptionWriter) writeNodeWithParent(nodeName string,
	nodeColor string,
	nodeImage string,
	nodeParent string,
	labelWeight string) error {

	nodeID, nodeErr := cytoscapeNodeID(nodeName)
	if nodeErr != nil {
		return errors.Wrapf(nodeErr,
			"Failed to create nodeID for entry: %s",
			nodeName)
	}
	parentID := ""
	if nodeParent != "" {
		tmpParentID, tmpParentIDErr := cytoscapeNodeID(nodeParent)
		if tmpParentIDErr != nil {
			return errors.Wrapf(nodeErr,
				"Failed to create nodeID for entry: %s",
				nodeParent)
		}
		parentID = tmpParentID
	}
	if labelWeight == "" {
		labelWeight = labelWeightNormal
	}
	appendNode := &cytoscapeNode{
		Data: cytoscapeData{
			ID:          nodeID,
			Parent:      parentID,
			Width:       "128",
			Height:      "128",
			Label:       strings.Trim(nodeName, "\""),
			LabelWeight: labelWeight,
		},
	}
	if nodeImage != "" {
		resourceItem := templateResourceForKey(nodeImage, dw.logger)
		if resourceItem != nil {
			appendNode.Data.Image = fmt.Sprintf("data:image/png;base64,%s",
				base64.StdEncoding.EncodeToString([]byte(resourceItem.Data)))
		}
	}
	dw.nodes = append(dw.nodes, appendNode)
	return nil
}

// func (dw *descriptionWriter) writeNode(nodeName string,
// 	nodeColor string,
// 	nodeImage string) error {
// 	return dw.writeNodeWithParent(nodeName, nodeColor, nodeImage, "")
// }

func (dw *descriptionWriter) writeEdge(fromNode string,
	toNode string,
	label string) error {

	nodeSource, nodeSourceErr := cytoscapeNodeID(fromNode)
	if nodeSourceErr != nil {
		return errors.Wrapf(nodeSourceErr,
			"Failed to create nodeID for entry: %s",
			fromNode)
	}
	nodeTarget, nodeTargetErr := cytoscapeNodeID(toNode)
	if nodeTargetErr != nil {
		return errors.Wrapf(nodeSourceErr,
			"Failed to create nodeID for entry: %s",
			toNode)
	}

	dw.nodes = append(dw.nodes, &cytoscapeNode{
		Data: cytoscapeData{
			ID:     fmt.Sprintf("%d", rand.Uint64()),
			Source: nodeSource,
			Target: nodeTarget,
			Label:  label,
		},
	})
	return nil
}

func templateResourceForKey(resourceKeyName string, logger *zerolog.Logger) *templateResource {
	var resource *templateResource
	resourcePath := fmt.Sprintf("/resources/describe/%s",
		strings.TrimLeft(resourceKeyName, "/"))

	keyParts := strings.Split(resourcePath, "/")
	keyName := keyParts[len(keyParts)-1]

	data, dataErr := _escFSString(false, resourcePath)
	if dataErr == nil {
		resource = &templateResource{
			KeyName: keyName,
			Data:    data,
		}
		logger.Debug().
			Str("Path", resourcePath).
			Str("KeyName", keyName).
			Msg("Embedded resource")

	} else {
		logger.Warn().
			Str("Path", resourcePath).
			Msg("Failed to find resource. Using default image.")

		data, dataErr = _escFSString(false, defaultImagePath)
		if dataErr == nil {
			resource = &templateResource{
				KeyName: keyName,
				Data:    data,
			}
		} else {
			logger.Error().
				Str("Path", defaultImagePath).
				Err(dataErr).
				Msg("Failed to find default resource")
		}
	}
	return resource
}

func templateResourcesForKeys(resourceKeyNames []string, logger *zerolog.Logger) []*templateResource {
	var resources []*templateResource

	for _, eachKey := range resourceKeyNames {
		loadedResource := templateResourceForKey(eachKey, logger)
		if loadedResource != nil {
			resources = append(resources, loadedResource)
		}
	}
	return resources
}

func templateCSSFiles(logger *zerolog.Logger) []*templateResource {
	cssFiles := []string{"bootstrap-4.0.0/dist/css/bootstrap.min.css",
		"highlight.js/styles/xcode.css",
	}
	return templateResourcesForKeys(cssFiles, logger)
}

func templateJSFiles(logger *zerolog.Logger) []*templateResource {
	jsFiles := []string{"jquery/jquery-3.3.1.min.js",
		"popper/popper.min.js",
		"bootstrap-4.0.0/dist/js/bootstrap.min.js",
		"highlight.js/highlight.pack.js",
		"dagre-0.8.4/dist/dagre.min.js",
		"cytoscape.js/dist/cytoscape.min.js",
		"cytoscape.js-dagre/cytoscape-dagre.js",
		"sparta.js",
	}
	return templateResourcesForKeys(jsFiles, logger)
}

func templateImageMap(logger *zerolog.Logger) map[string]string {
	images := []string{"SpartaHelmet256.png",
		"AWS-Architecture-Icons_PNG/PNG Light/Compute/AWS-Lambda_Lambda-Function_light-bg@4x.png",
		"AWS-Architecture-Icons_PNG/PNG Light/Management & Governance/AWS-CloudFormation_light-bg@4x.png",
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
func iconForAWSResource(rawEmitter interface{}) *DescriptionIcon {
	jsonBytes, jsonBytesErr := json.Marshal(rawEmitter)
	if jsonBytesErr != nil {
		jsonBytes = make([]byte, 0)
	}
	canonicalRaw := strings.ToLower(string(jsonBytes))
	iconMappings := map[string]*DescriptionIcon{
		"dynamodb": {
			Category: "Database",
			Name:     "Amazon-DynamoDB@4x.png",
		},
		"sqs": {
			Category: "Application Integration",
			Name:     "Amazon-Simple-Queue-Service-SQS@4x.png",
		},
		"sns": {
			Category: "Application Integration",
			Name:     "Amazon-Simple-Notification-Service-SNS@4x.png",
		},
		"cloudwatch": {
			Category: "Management & Governance",
			Name:     "Amazon-Simple-Notification-Service-SNS@4x.png",
		},
		"kinesis": {
			Category: "Analytics",
			Name:     "Amazon-Kinesis@4x.png",
		},
		"s3": {
			Category: "Storage",
			Name:     "Amazon-Simple-Storage-Service-S3@4x.png",
		},
		"codecommit": {
			Category: "Developer Tools",
			Name:     "AWS-CodeCommit_light-bg.svg",
		},
	}
	// Return it if we have it...
	for eachKey, eachIcon := range iconMappings {
		if strings.Contains(canonicalRaw, eachKey) {
			return eachIcon
		}
	}
	return nil
}

// DescriptionIcon is the struct that contains the category & icon
// to use when rendering a ndoe
type DescriptionIcon struct {
	Category string
	Name     string
}

// DescriptionDisplayInfo encapsulates information that is for display only
type DescriptionDisplayInfo struct {
	SourceNodeColor string
	SourceIcon      *DescriptionIcon
}

// DescriptionTriplet is a node that should be included in the final
// describe output.
type DescriptionTriplet struct {
	SourceNodeName string
	ArcLabel       string
	DisplayInfo    *DescriptionDisplayInfo
	TargetNodeName string
}

// DescriptionInfo is the set of information that represents a DescribeableResource
type DescriptionInfo struct {
	Name  string
	Nodes []*DescriptionTriplet
}

// Describable represents the interface for something that can
// provide a description
type Describable interface {
	Describe(targetNodeName string) (*DescriptionInfo, error)
}
