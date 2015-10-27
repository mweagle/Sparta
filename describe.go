// Copyright (c) 2015 Matt Weagle <mweagle@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// +build !lambdabinary

package sparta

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"io"
	"strconv"
	"text/template"
)

const DESCRIPTION_TEMPLATE = `<!doctype html>
<html>
<head>
  <title>{{ .ServiceName }}</title>

	<style>
	{{ .MermaidCSS }}
	</style>

	<script>
	{{ .MermaidJS}}

	mermaid.initialize({startOnLoad:true});

	</script>
</head>
<body>
	<h3>{{ .ServiceName }} - {{ .ServiceDescription }}</h3>
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

func writenode(writer io.Writer, nodeName string) {
	fmt.Fprintf(writer, "%s[%s]\n", nodeName, nodeName)
}

func writelink(writer io.Writer, fromNode string, toNode string) {
	fmt.Fprintf(writer, "%s-->%s\n", fromNode, toNode)
}

// Produces a graphical representation of your service's Lambda and data sources.  Typically
// automatically called as part of a compiled golang binary via the `describe` command
// line option.
func Describe(serviceName string, serviceDescription string, lambdaAWSInfos []*LambdaAWSInfo, outputWriter io.Writer, logger *logrus.Logger) error {

	tmpl, err := template.New("description").Parse(DESCRIPTION_TEMPLATE)
	if err != nil {
		return errors.New(err.Error())
	}

	var b bytes.Buffer

	// Setup the root object
	writenode(&b, serviceName)
	fmt.Fprintf(&b, "style %s fill:#f9f,stroke:#333,stroke-width:4px;\n", serviceName)

	for _, eachLambda := range lambdaAWSInfos {
		logger.Debug("Appending: ", eachLambda.lambdaFnName)
		// Create the node...
		writenode(&b, eachLambda.lambdaFnName)
		writelink(&b, eachLambda.lambdaFnName, serviceName)

		// Create permission & event mappings
		// functions declared in this
		for _, eachPermission := range eachLambda.Permissions {
			nodeName := *eachPermission.Principal
			if "" != *eachPermission.SourceArn {
				nodeName = *eachPermission.SourceArn
			}
			writenode(&b, nodeName)
			writelink(&b, nodeName, eachLambda.lambdaFnName)
		}

		for _, eachEventSourceMapping := range eachLambda.EventSourceMappings {
			nodeName := *eachEventSourceMapping.EventSourceArn
			writenode(&b, nodeName)
			writelink(&b, nodeName, eachLambda.lambdaFnName)
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
