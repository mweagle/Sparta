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

package sparta

// NOTE: your application MUST use `package main` and define a `main()` function.  The
// example text is to make the documentation compatible with godoc.

func helloWorld(event sparta.LambdaEvent, context sparta.LambdaContext, w http.ResponseWriter) {
	fmt.Fprintf(w, "Hello World!")
}

// Should be main() in your application
//
// The compiled application then supports self-deployment (and update), description,
// and execution via command line arguments:
//
// $ ./MyApplication
//
// Usage: MyApplication [global options] <verb> [verb options]
//
// Global options:
//         -l, --level    Log level [panic, fatal, error, warn, info, debug] (default: info)
//         -h, --help     Show this help
//
// Verbs:
//     provision:
//         -b, --s3Bucket S3 Bucket to use for Lambda source (*)
//     describe:
//         -o, --out      Output file for HTML description (*)
//     execute:
//         -p, --port     Alternative port for HTTP binding
//         -s, --signal   Process ID to signal with SIGUSR2 once ready
func ExampleMain() {
	var lambdaFunctions []*sparta.LambdaAWSInfo
	helloWorldLambda := sparta.NewLambda("PreexistingAWSLambdaRoleName", helloWorld, nil)
	lambdaFunctions = append(lambdaFunctions, helloWorldLambda)
	sparta.Main("HelloWorldLambdaService", "Description for Hello World Lambda", lambdaFunctions)
}
