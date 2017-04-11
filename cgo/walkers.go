package cgo

import (
	"fmt"
	"go/ast"
	"regexp"
)

var cgoImports = `// #include <stdio.h>
// #include <stdlib.h>
// #include <string.h>
// #include <stdio.h>
import "C"
`

const cgoExports = `

//export Lambda
func Lambda(functionName *C.char,
	requestJSON *C.char,
	exitCode *C.int,
	responseContentTypeBuffer *C.char,
	responseContentTypeLen int,
	responseBuffer *C.char,
	responseBufferContentLen int) int {

	inputFunction := C.GoString(functionName)
	inputRequest := C.GoString(requestJSON)
	spartaResp, spartaRespHeaders, responseErr := cgo.LambdaHandler(inputFunction, inputRequest)
	lambdaExitCode := 0
	var pyResponseBufer []byte
	if nil != responseErr {
		lambdaExitCode = 1
		pyResponseBufer = []byte(responseErr.Error())
	} else {
		pyResponseBufer = spartaResp
	}


	// Copy content type
	contentTypeHeader := spartaRespHeaders.Get("Content-Type")

	// If there is no header, assume it's json
	if "" == contentTypeHeader {
		contentTypeHeader = "application/json"
	}
	if "" != contentTypeHeader {
		responseContentTypeBytes := C.CBytes([]byte(contentTypeHeader))
		defer C.free(responseContentTypeBytes)
		copyContentTypeBytesLen := len(contentTypeHeader)
		if (copyContentTypeBytesLen > responseContentTypeLen) {
			copyContentTypeBytesLen = responseContentTypeLen
		}
		C.memcpy(unsafe.Pointer(responseContentTypeBuffer),
			unsafe.Pointer(responseContentTypeBytes),
			C.size_t(copyContentTypeBytesLen))
	}

	// Copy response body
	copyBytesLen := len(pyResponseBufer)
	if copyBytesLen > responseBufferContentLen {
		copyBytesLen = responseBufferContentLen
	}
	responseBytes := C.CBytes(pyResponseBufer)
	defer C.free(responseBytes)
	C.memcpy(unsafe.Pointer(responseBuffer),
		unsafe.Pointer(responseBytes),
		C.size_t(copyBytesLen))
	*exitCode = C.int(lambdaExitCode)
	return copyBytesLen
}

func main() {
	// NOP
}
`

var packageRegexp = regexp.MustCompile("(?m)^package.*[\r\n]{1,2}")

// First thing we need to do is change the main() function
// to be an init() function
type mainRewriteVisitor struct {
}

func (v *mainRewriteVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch t := node.(type) {
	case *ast.FuncDecl:
		if t.Name.Name == "main" {
			t.Name = ast.NewIdent("init")
		}
	}
	return v
}

func visitors() []ast.Visitor {
	return []ast.Visitor{
		&mainRewriteVisitor{},
	}
}

type transformer func(inputText string) (string, error)

func cgoImportsTransformer(inputText string) (string, error) {
	matchIndex := packageRegexp.FindStringIndex(inputText)
	if nil == matchIndex {
		return "", fmt.Errorf("Failed to find package statement")
	}
	// Great, append the cgo header
	return fmt.Sprintf("%s%s%s",
		inputText[0:matchIndex[1]],
		cgoImports,
		inputText[matchIndex[1]:]), nil
}

func cgoExportsTransformer(inputText string) (string, error) {
	return fmt.Sprintf("%s\n%s", inputText, cgoExports), nil
}

func transformers() []transformer {
	return []transformer{
		cgoImportsTransformer,
		cgoExportsTransformer,
	}
}
