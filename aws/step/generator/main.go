package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"text/template"
)

type ChoiceRule struct {
	Name         string
	VariableType string
}

var choiceStatePrelude = `
// Code generated by github.com/mweagle/Sparta/aws/step/generator/main.go. DO NOT EDIT.

package step

import (
	"encoding/json"
	"time"
)

/*******************************************************************************
   ___ ___  __  __ ___  _   ___ ___ ___  ___  _  _ ___
  / __/ _ \|  \/  | _ \/_\ | _ \_ _/ __|/ _ \| \| / __|
 | (_| (_) | |\/| |  _/ _ \|   /| |\__ \ (_) |    \__ \
  \___\___/|_|  |_|_|/_/ \_\_|_\___|___/\___/|_|\_|___/

/******************************************************************************/

// For path based selectors see the
// JSONPath: https://github.com/NodePrime/jsonpath
// documentation

`
var choiceStateTemplate = `
////////////////////////////////////////////////////////////////////////////////
// {{.Name}}
////////////////////////////////////////////////////////////////////////////////

// {{.Name}} comparison
type {{.Name}} struct {
	Comparison
	Variable string
	Value    {{.VariableType}}
}

// MarshalJSON for custom marshalling
func (cmp *{{.Name}}) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable       string
		{{.Name}} 		 {{.VariableType}}
	}{
		Variable:       cmp.Variable,
		{{.Name}}: 			cmp.Value,
	})
}
`

type choiceRuleVariableDef struct {
	Variable string
}
type choiceDefinitions struct {
	Choices map[string]choiceRuleVariableDef `json:"choices"`
}

func main() {
	// Params are the path to the input definitions and the
	// output path.
	if len(os.Args) != 3 {
		fmt.Printf("Please provide path to the source definition as arg1, path to the output file as arg2\n")
		os.Exit(1)
	}
	inputDefFile := os.Args[1]
	outputSourceFile := os.Args[2]
	/* #nosec G304 */
	inputFileBytes, inputFileBytesErr := ioutil.ReadFile(inputDefFile)
	if inputFileBytesErr != nil {
		fmt.Printf("Failed to read %s: %v\n", inputDefFile, inputFileBytesErr)
		os.Exit(1)
	}
	var choiceDefs choiceDefinitions
	unmarshalErr := json.Unmarshal(inputFileBytes, &choiceDefs)
	if unmarshalErr != nil {
		fmt.Printf("Failed to unmarshal %s: %v", inputDefFile, unmarshalErr)
		os.Exit(1)
	}

	// Rip the definitions and write each one
	ruleTemplate, ruleTemplateErr := template.New("ruleTemplate").Parse(choiceStateTemplate)
	if ruleTemplateErr != nil {
		fmt.Printf("Failed to parse template: %v\n", ruleTemplateErr)
		os.Exit(1)
	}

	outputSource, outputSourceErr := os.Create(outputSourceFile)
	if outputSourceErr != nil {
		fmt.Printf("Failed to open %s: %v\n", outputSourceFile, outputSourceErr)
		os.Exit(1)
	}
	/* #nosec */
	defer func() {
		closeErr := outputSource.Close()
		if closeErr != nil {
			fmt.Printf("Failed to close output stream: %#v", closeErr)
		}
	}()
	_, writeErr := io.WriteString(outputSource, choiceStatePrelude)
	if writeErr != nil {
		fmt.Printf("Failed to write: %v\n", writeErr)
		os.Exit(1)
	}
	for eachRuleName, eachRuleDef := range choiceDefs.Choices {
		templateParams := ChoiceRule{
			Name:         eachRuleName,
			VariableType: eachRuleDef.Variable,
		}
		executeErr := ruleTemplate.Execute(outputSource, templateParams)
		if executeErr != nil {
			fmt.Printf("Failed to execute template: %v\n", executeErr)
			os.Exit(1)
		}
	}
	fmt.Printf("All done!\n")
}
