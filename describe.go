// +build !lambdabinary

package sparta

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/template"

	"github.com/Sirupsen/logrus"
)

const descriptionTemplate = `<!doctype html>
<html>
<head>
  <title>{{ .ServiceName }}</title>

	<style>
	{{ .MermaidCSS }}
	</style>


  <style>
    body {
      background-color: #F5F5F5;
      font-family: "-apple-system", Menlo, Arial, Helvetica, sans-serif;
      font-size: smaller;
    }
    h2 {
      font-variant: small-caps;
    }
  </style>
	<script>
	{{ .MermaidJS}}

	mermaid.initialize({startOnLoad:true,
										htmlLabels: true,
									  flowchart:{
									     useMaxWidth: true
									  }
										});

	</script>
</head>
<body>
	<h2> {{ .ServiceName }} </h2>
	<h5> {{ .ServiceDescription }}</h5>
	<hr />
	<div class="mermaid">
		%% Example code
		graph LR

    {{ .MermaidData}}
	</div>
</body>
</html>
`

func nodeObject(text string, shape string, group int) *ArbitraryJSONObject {
	return &ArbitraryJSONObject{
		"id":    text,
		"label": text,
		"shape": shape,
		"group": strconv.Itoa(group),
	}
}

func edgeObject(from string, to string, edgeLabel string) *ArbitraryJSONObject {
	return &ArbitraryJSONObject{
		"from":  from,
		"to":    to,
		"label": edgeLabel,
		"color": &ArbitraryJSONObject{
			"inherit": "both",
		},
	}
}

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

// Describe produces a graphical representation of a service's Lambda and data sources.  Typically
// automatically called as part of a compiled golang binary via the `describe` command
// line option.
func Describe(serviceName string, serviceDescription string, lambdaAWSInfos []*LambdaAWSInfo, outputWriter io.Writer, logger *logrus.Logger) error {

	tmpl, err := template.New("description").Parse(descriptionTemplate)
	if err != nil {
		return errors.New(err.Error())
	}

	var b bytes.Buffer

	// Setup the root object
	writenode(&b, serviceName, "2AF1EA")

	for _, eachLambda := range lambdaAWSInfos {
		logger.Debug("Appending: ", eachLambda.lambdaFnName)
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
	params := struct {
		ServiceName        string
		ServiceDescription string
		MermaidCSS         string
		MermaidJS          string
		MermaidData        string
	}{
		serviceName,
		serviceDescription,
		FSMustString(false, "/resources/mermaid/mermaid.css"),
		FSMustString(false, "/resources/mermaid/mermaid.min.js"),
		b.String(),
	}
	return tmpl.Execute(outputWriter, params)
}
