// +build !lambdabinary

package sparta

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/Sirupsen/logrus"
)

func writenode(writer io.Writer, nodeName string, nodeColor string) {
	fmt.Fprintf(writer, "style %s fill:#%s,stroke:#000,stroke-width:1px;\n", nodeName, nodeColor)
	fmt.Fprintf(writer, "%s[%s]\n", nodeName, nodeName)
}

func writelink(writer io.Writer, fromNode string, toNode string, label string) {
	if "" != label {
		fmt.Fprintf(writer, "%s-- \"%s\" -->%s\n", fromNode, label, toNode)
	} else {
		fmt.Fprintf(writer, "%s-->%s\n", fromNode, toNode)
	}
}
func describeAPI() string {
	return ""
}

// Describe produces a graphical representation of a service's Lambda and data sources.  Typically
// automatically called as part of a compiled golang binary via the `describe` command
// line option.
func Describe(serviceName string, serviceDescription string, lambdaAWSInfos []*LambdaAWSInfo, api *API, outputWriter io.Writer, logger *logrus.Logger) error {
	var cloudFormationTemplate bytes.Buffer
	err := Provision(true, serviceName, serviceDescription, lambdaAWSInfos, api, "S3Bucket", &cloudFormationTemplate, logger)
	if nil != err {
		return err
	}
	/*
		// Export the template and insert it into an HTML page.  Let the page do the work...

		for _, eachEntry := range ctx.lambdaAWSInfos {
			err := eachEntry.export(ctx.s3Bucket, s3Key, ctx.lambdaIAMRoleNameMap, ctx.cloudformationResources, ctx.cloudformationOutputs, ctx.logger)
			if nil != err {
				return nil, err
			}
		}
		// If there's an API gateway definition, provision custom resources
		// and IAM role to
		if nil != ctx.api {
			ctx.api.export(ctx.s3Bucket, s3Key, ctx.lambdaIAMRoleNameMap, ctx.cloudformationResources, ctx.cloudformationOutputs, ctx.logger)
		}
	*/
	tmpl, err := template.New("description").Parse(FSMustString(false, "/resources/describe/template.html"))
	if err != nil {
		return errors.New(err.Error())
	}

	var b bytes.Buffer

	// Setup the root object
	writenode(&b, serviceName, "2AF1EA")

	for _, eachLambda := range lambdaAWSInfos {
		// Create the node...
		writenode(&b, eachLambda.lambdaFnName, "00A49F")
		writelink(&b, eachLambda.lambdaFnName, serviceName, "")

		// Create permission & event mappings
		// functions declared in this
		for _, eachPermission := range eachLambda.Permissions {
			name, link := eachPermission.descriptionInfo()

			// Style it to have the Amazon color
			writenode(&b, name, "F1702A")
			writelink(&b, name, eachLambda.lambdaFnName, strings.Replace(link, " ", "<br>", -1))
		}

		for _, eachEventSourceMapping := range eachLambda.EventSourceMappings {
			nodeName := *eachEventSourceMapping.EventSourceArn
			writenode(&b, nodeName, "F1702A")
			writelink(&b, nodeName, eachLambda.lambdaFnName, "")
		}
	}
	// If there's an API, then output that as well...

	// params := struct {
	// 	ServiceName        string
	// 	ServiceDescription string
	// 	MermaidCSS         string
	// 	MermaidJS          string
	// 	MermaidData        string
	// }{
	// 	serviceName,
	// 	serviceDescription,
	// 	FSMustString(false, "/resources/bootstrap-3.3.5/css/bootstrap.css"),
	// 	FSMustString(false, "/resources/mermaid/mermaid.css"),
	// 	FSMustString(false, "/resources/mermaid/mermaid.min.js"),
	// 	FSMustString(false, "/resources/jquery/jquery-2.1.4.min.js"),
	// 	FSMustString(false, "/resources/bootstrap-3.3.5/js/bootstrap.js"),
	// 	b.String(),
	// }
	params := struct {
		SpartaVersion          string
		ServiceName            string
		ServiceDescription     string
		CloudFormationTemplate string
		BootstrapCSS           string
		MermaidCSS             string
		HighlightsCSS          string
		JQueryJS               string
		BootstrapJS            string
		MermaidJS              string
		HighlightsJS           string
		MermaidData            string
	}{
		SpartaVersion,
		serviceName,
		serviceDescription,
		cloudFormationTemplate.String(),
		FSMustString(false, "/resources/bootstrap/css/bootstrap.min.css"),
		FSMustString(false, "/resources/mermaid/mermaid.css"),
		FSMustString(false, "/resources/highlights/styles/vs.css"),
		FSMustString(false, "/resources/jquery/jquery-2.1.4.min.js"),
		FSMustString(false, "/resources/bootstrap/js/bootstrap.min.js"),
		FSMustString(false, "/resources/mermaid/mermaid.min.js"),
		FSMustString(false, "/resources/highlights/highlight.pack.js"),
		b.String(),
	}

	return tmpl.Execute(outputWriter, params)
}
