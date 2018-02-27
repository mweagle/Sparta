// +build !lambdabinary

package sparta

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
)

const (
	nodeColorService     = "#EFEFEF"
	nodeColorEventSource = "#FBBB06"
	nodeColorLambda      = "#F58206"
	nodeColorAPIGateway  = "#06B5F5"
	nodeNameAPIGateway   = "API Gateway"
)

type templateResource struct {
	KeyName string
	Data    string
}

// RE for sanitizing golang/JS layer
var reSanitizeMermaidNodeName = regexp.MustCompile(`[\W\s]+`)
var reSanitizeMermaidLabelValue = regexp.MustCompile(`[\{\}\"\[\]']+`)

func mermaidNodeName(sourceName string) string {
	return reSanitizeMermaidNodeName.ReplaceAllString(sourceName, "x")
}

func mermaidLabelValue(labelText string) string {
	return reSanitizeMermaidLabelValue.ReplaceAllString(labelText, "")
}

func writeNode(writer io.Writer, nodeName string, nodeColor string, extraStyles string) {
	if "" != extraStyles {
		extraStyles = fmt.Sprintf(",%s", extraStyles)
	}
	sanitizedName := mermaidNodeName(nodeName)
	fmt.Fprintf(writer, "style %s fill:%s,stroke:#000,stroke-width:1px%s;\n", sanitizedName, nodeColor, extraStyles)
	fmt.Fprintf(writer, "%s[%s]\n", sanitizedName, mermaidLabelValue(nodeName))
}

func writeLink(writer io.Writer, fromNode string, toNode string, label string) {
	sanitizedFrom := mermaidNodeName(fromNode)
	sanitizedTo := mermaidNodeName(toNode)

	if "" != label {
		fmt.Fprintf(writer, "%s-- \"%s\" -->%s\n", sanitizedFrom, mermaidLabelValue(label), sanitizedTo)
	} else {
		fmt.Fprintf(writer, "%s-->%s\n", sanitizedFrom, sanitizedTo)
	}
}

func templateResourcesForKeys(resourceKeyNames []string, logger *logrus.Logger) []*templateResource {
	resources := make([]*templateResource, 0)

	for _, eachKey := range resourceKeyNames {
		resourcePath := fmt.Sprintf("/resources/describe/%s",
			strings.TrimLeft(eachKey, "/"))
		data, dataErr := _escFSString(false, resourcePath)
		if dataErr == nil {
			keyParts := strings.Split(resourcePath, "/")
			keyName := keyParts[len(keyParts)-1]
			resources = append(resources, &templateResource{
				KeyName: keyName,
				Data:    data,
			})
			logger.WithFields(logrus.Fields{
				"Path":    resourcePath,
				"KeyName": keyName,
			}).Debug("Embedded resource")

		} else {
			logger.WithFields(logrus.Fields{
				"Path": resourcePath,
			}).Warn("Failed to embed resource")
		}
	}
	return resources
}

func templateCSSFiles(logger *logrus.Logger) []*templateResource {
	cssFiles := []string{"mermaid-7.1.2/mermaid.css",
		"bootstrap-4.0.0/dist/css/bootstrap.min.css",
		"highlight.js/styles/xcode.css",
	}
	return templateResourcesForKeys(cssFiles, logger)
}

func templateJSFiles(logger *logrus.Logger) []*templateResource {
	jsFiles := []string{"jquery/jquery-3.3.1.min.js",
		"popper/popper.min.js",
		"mermaid-7.1.2/mermaid.js",
		"bootstrap-4.0.0/dist/js/bootstrap.min.js",
		"highlight.js/highlight.pack.js",
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

	var b bytes.Buffer

	// Setup the root object
	writeNode(&b,
		serviceName,
		nodeColorService,
		"color:white,font-weight:bold,stroke-width:4px")

	for _, eachLambda := range lambdaAWSInfos {
		// Create the node...
		writeNode(&b, eachLambda.lambdaFunctionName(), nodeColorLambda, "")
		writeLink(&b, eachLambda.lambdaFunctionName(), serviceName, "")

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
				writeNode(&b, name, nodeColor, "border-style:dotted")
				writeLink(&b, name, eachLambda.lambdaFunctionName(), strings.Replace(link, "\n", "<br><br>", -1))
			}
		}

		for _, eachEventSourceMapping := range eachLambda.EventSourceMappings {
			writeNode(&b, eachEventSourceMapping.EventSourceArn, nodeColorEventSource, "border-style:dotted")
			writeLink(&b, eachEventSourceMapping.EventSourceArn, eachLambda.lambdaFunctionName(), "")
		}
	}

	// API?
	if nil != api {
		// Create the APIGateway virtual node && connect it to the application
		writeNode(&b, nodeNameAPIGateway, nodeColorAPIGateway, "")

		for _, eachResource := range api.resources {
			for eachMethod := range eachResource.Methods {
				// Create the PATH node
				var nodeName = fmt.Sprintf("%s - %s", eachMethod, eachResource.pathPart)
				writeNode(&b, nodeName, nodeColorAPIGateway, "")
				writeLink(&b, nodeNameAPIGateway, nodeName, "")
				writeLink(&b, nodeName, eachResource.parentLambda.lambdaFunctionName(), "")
			}
		}
	}

	params := struct {
		SpartaVersion          string
		ServiceName            string
		ServiceDescription     string
		CloudFormationTemplate string
		CSSFiles               []*templateResource
		JSFiles                []*templateResource
		ImageMap               map[string]string
		MermaidData            string
	}{
		SpartaGitHash[0:8],
		serviceName,
		serviceDescription,
		cloudFormationTemplate.String(),
		templateCSSFiles(logger),
		templateJSFiles(logger),
		templateImageMap(logger),
		b.String(),
	}

	return tmpl.Execute(outputWriter, params)
}
