// Reading and writing files are basic tasks needed for
// many Go programs. First we'll look at some examples of
// reading files.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	targetFile := os.Args[1]
	tags := os.Args[2:]

	// go run tries to compile the targetfile if it has the .go extension, so
	// we don't provide that on the command line and append it here.
	targetFile = fmt.Sprintf("%s.go", targetFile)

	absPath, err := filepath.Abs(targetFile)
	if nil != err {
		panic(err)
	}
	/* #nosec */
	fileContents, err := ioutil.ReadFile(absPath)
	if nil != err {
		panic(err)
	}
	tagString := strings.Join(tags, " ")
	fmt.Printf("Prepending tags: %s\n", tagString)

	// Include the #nosec directive to have gas ignore
	// the ignored error returns
	// https://github.com/GoASTScanner/gas
	updatedContents := fmt.Sprintf(`// +build lambdabinary
	
// lint:file-ignore U1000 Ignore all unused code, it's generated
/* #nosec */
	
%s`,
		tagString,
		fileContents)
	err = ioutil.WriteFile(absPath, []byte(updatedContents), 0644)
	if nil != err {
		panic(err)
	}
}
