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
	"encoding/json"
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

  <script type="text/javascript">
  {{ .VisJS }}
  </script>
  <style> {{ .VisCSS }} </style>
  <style type="text/css">
    #lambdaCloud {
      border: 4px solid lightgray;

	    max-width: 100%;
	    max-height: 100%;
	    bottom: 0;
	    left: 0;
	    margin: auto;
	    overflow: auto;
	    position: fixed;
	    right: 0;
	    top: 0;

    }
  </style>
</head>
<body>

	<h3>{{ .ServiceName }} - {{ .ServiceDescription }}</h3>
	<div id="lambdaCloud"></div>

<script type="text/javascript">

  // create a network
  var container = document.getElementById('lambdaCloud');
  var data = {
    nodes: {{ .Nodes }},
    edges: {{ .Edges }}
  };
 	var options = {
      layout: {
      		randomSeed:3,
          hierarchical: {
              direction: "LR",
              sortMethod: "directed"
          }
      },
			edges: {
			       smooth: true,
			       arrows: {to : true }
			}
	};
  var network = new vis.Network(container, data, options);
</script>

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

// Produces a graphical representation of your service's Lambda and data sources.  Typically
// automatically called as part of a compiled golang binary via the `describe` command
// line option.
func Describe(serviceName string, serviceDescription string, lambdaAWSInfos []*LambdaAWSInfo, outputWriter io.Writer, logger *logrus.Logger) error {

	tmpl, err := template.New("description").Parse(DESCRIPTION_TEMPLATE)
	if err != nil {
		return errors.New(err.Error())
	}

	// Create the graph
	nodes := make([]*ArbitraryJSONObject, 0)
	edges := make([]*ArbitraryJSONObject, 0)

	// Add the root object
	rootObjectID := serviceName
	rootObjectJSON := &ArbitraryJSONObject{
		"id":    rootObjectID,
		"label": rootObjectID,
		"shape": "star",
		"size":  30,
		"font": ArbitraryJSONObject{
			"size": 30,
		},
	}
	nodes = append(nodes, rootObjectJSON)

	for eachGroup, eachLambda := range lambdaAWSInfos {
		logger.Info("Appending: ", eachLambda.lambdaFnName)
		nodes = append(nodes, nodeObject(eachLambda.lambdaFnName, "box", eachGroup))
		edges = append(edges, edgeObject(eachLambda.lambdaFnName, rootObjectID, ""))

		// Create permission & event mappings
		// functions declared in this
		for _, eachPermission := range eachLambda.Permissions {
			nodeName := *eachPermission.Principal
			if "" != *eachPermission.SourceArn {
				nodeName = *eachPermission.SourceArn
			}
			nodes = append(nodes, nodeObject(nodeName, "dot", eachGroup))
			edges = append(edges, edgeObject(nodeName, eachLambda.lambdaFnName, *eachPermission.Action))
		}

		for _, eachEventSourceMapping := range eachLambda.EventSourceMappings {
			nodeName := *eachEventSourceMapping.EventSourceArn
			nodes = append(nodes, nodeObject(nodeName, "dot", eachGroup))
			edgeLabel := fmt.Sprintf("StartingPosition=%s", *eachEventSourceMapping.StartingPosition)
			edges = append(edges, edgeObject(nodeName, eachLambda.lambdaFnName, edgeLabel))
		}
	}
	logger.Info("Node count: ", len(nodes))
	nodeContent, err := json.Marshal(nodes)
	if err != nil {
		return errors.New("Failed to Marshal node content: " + err.Error())
	}
	edgeContent, err := json.Marshal(edges)
	if err != nil {
		return errors.New("Failed to Marshal edge template: " + err.Error())
	}

	params := struct {
		ServiceName        string
		ServiceDescription string
		VisJS              string
		VisCSS             string
		Nodes              string
		Edges              string
	}{
		serviceName,
		serviceDescription,
		FSMustString(false, "/resources/vis/vis.min.js"),
		FSMustString(false, "/resources/vis/vis.min.css"),
		string(nodeContent),
		string(edgeContent),
	}
	return tmpl.Execute(outputWriter, params)
}
