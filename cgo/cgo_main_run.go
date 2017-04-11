// +build !lambdabinary

package cgo

import (
	"bytes"
	"fmt"
	"github.com/mweagle/Sparta"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"os"
)

func cgoMain(callerFile string,
	serviceName string,
	serviceDescription string,
	lambdaAWSInfos []*sparta.LambdaAWSInfo,
	api *sparta.API,
	site *sparta.S3Site,
	workflowHooks *sparta.WorkflowHooks) error {

	// So this depends on being able to rewrite the main() function...

	// Read the main() input
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, callerFile, nil, 0)
	if err != nil {
		fmt.Printf("ParseFileErr: %s", err.Error())
		return err
	}

	// Add the imports that we need in the walkers text
	astutil.AddImport(fset, file, "unsafe")

	// Great, now change main to init()
	for _, eachVisitor := range visitors() {
		ast.Walk(eachVisitor, file)
	}

	// Now save the file as a string and run it through
	// the text transformers
	var byteWriter bytes.Buffer
	transformErr := printer.Fprint(&byteWriter, fset, file)
	if nil != transformErr {
		return transformErr
	}
	updatedSource := byteWriter.String()
	for _, eachTransformer := range transformers() {
		transformedSource, transformedSourceErr := eachTransformer(updatedSource)
		if nil != transformedSourceErr {
			return transformedSourceErr
		}
		updatedSource = transformedSource
	}
	// The temporary file is the input file, with a suffix
	rewrittenFilepath := fmt.Sprintf("%s.rewritten.go", callerFile)
	swappedFilepath := fmt.Sprintf("%s.sparta.og", callerFile)
	renameErr := os.Rename(callerFile, swappedFilepath)
	if nil != renameErr {
		fmt.Printf("Failed to backup source: %s", renameErr.Error())
		return renameErr
	}
	defer os.Rename(swappedFilepath, callerFile)

	outputFile, outputFileErr := os.Create(rewrittenFilepath)
	if nil != outputFileErr {
		fmt.Printf("Failed to create output file: %s", outputFileErr.Error())
		return outputFileErr
	}
	defer outputFile.Close()
	defer os.Remove(rewrittenFilepath)

	// Save the updated contents
	_, writtenErr := outputFile.WriteString(updatedSource)
	if nil != writtenErr {
		return writtenErr
	}
	// Great, let's go ahead and do the build.
	mainErr := sparta.MainEx(serviceName,
		serviceDescription,
		lambdaAWSInfos,
		api,
		site,
		workflowHooks,
		true)

	// TODO - if there's an error, rename the
	// file to something we can debug

	return mainErr
}
