// +build mage

// lint:file-ignore U1000 Ignore all  code, it's only for development

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
	"github.com/magefile/mage/sh" // mg contains helpful utility functions, like Deps
	"github.com/mholt/archiver"
	spartamage "github.com/mweagle/Sparta/magefile"
	"github.com/otiai10/copy"
	"github.com/pkg/browser"
	"github.com/pkg/errors"
)

const (
	localWorkDir      = "./.sparta"
	hugoVersion       = "0.79.0"
	archIconsRootPath = "resources/describe/AWS-Architecture-Icons_PNG"
	archIconsTreePath = "resources/describe/AWS-Architecture-Icons.tree.txt"
)

func xplatPath(pathParts ...string) string {
	return filepath.Join(pathParts...)
}

var (
	ignoreSubdirectoryPaths = []string{
		xplatPath(".vendor"),
		xplatPath(".sparta"),
		xplatPath(".vscode"),
		xplatPath("resources", "describe"),
		xplatPath("docs_source", "themes"),
	}
	hugoDocsSourcePath = xplatPath(".", "docs_source")
	hugoDocsPaths      = []string{
		hugoDocsSourcePath,
		xplatPath(".", "docs"),
	}
	hugoPath = filepath.Join(localWorkDir, "hugo")
	header   = strings.Repeat("-", 80)
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

func markdownSourceApply(commandParts ...string) error {
	return spartamage.ApplyToSource("md", ignoreSubdirectoryPaths, commandParts...)
}
func goSourceApply(commandParts ...string) error {
	return spartamage.ApplyToSource("go", ignoreSubdirectoryPaths, commandParts...)
}

func goFilteredSourceApply(ignorePatterns []string, commandParts ...string) error {
	ignorePatterns = append(ignorePatterns, ignoreSubdirectoryPaths...)
	return spartamage.ApplyToSource("go", ignorePatterns, commandParts...)
}

func gitCommit(shortVersion bool) (string, error) {
	args := []string{
		"rev-parse",
	}
	if shortVersion {
		args = append(args, "--short")
	}
	args = append(args, "HEAD")
	val, valErr := sh.Output("git", args...)
	return strings.TrimSpace(val), valErr
}

// EnsureCleanTree ensures that the git tree is clean
func EnsureCleanTree() error {
	cleanTreeScript := [][]string{
		// No dirty trees
		{"git", "diff", "--exit-code"},
	}
	return spartamage.Script(cleanTreeScript)
}

////////////////////////////////////////////////////////////////////////////////
// START - DOCUMENTATION
////////////////////////////////////////////////////////////////////////////////

// ensureWorkDir ensures that the scratch directory exists
func ensureWorkDir() error {
	return os.MkdirAll(localWorkDir, os.ModePerm)
}

func runHugoCommand(hugoCommandArgs ...string) error {
	absHugoPath, absHugoPathErr := filepath.Abs(hugoPath)
	if absHugoPathErr != nil {
		return absHugoPathErr
	}

	// Get the git short value
	gitSHA, gitSHAErr := gitCommit(true)
	if gitSHAErr != nil {
		return gitSHAErr
	}

	workDir, workDirErr := filepath.Abs(hugoDocsSourcePath)
	if workDirErr != nil {
		return workDirErr
	}
	var output io.Writer
	if mg.Verbose() {
		output = os.Stdout
	}
	cmd := exec.Command(absHugoPath, hugoCommandArgs...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("GIT_HEAD_COMMIT=%s", gitSHA))
	cmd.Stderr = os.Stderr
	cmd.Stdout = output
	cmd.Dir = workDir
	return cmd.Run()
}

func docsCopySourceTemplatesToDocs() error {
	outputDir := filepath.Join(".",
		"docs_source",
		"static",
		"source",
		"resources",
		"provision",
		"apigateway")
	rmErr := os.RemoveAll(outputDir)
	if rmErr != nil {
		return rmErr
	}
	// Create the directory
	createErr := os.MkdirAll(outputDir, os.ModePerm)
	if createErr != nil {
		return createErr
	}
	inputDir := filepath.Join(".", "resources", "provision", "apigateway")
	return copy.Copy(inputDir, outputDir)
}

// DocsInstallRequirements installs the required Hugo version
func DocsInstallRequirements() error {
	mg.SerialDeps(ensureWorkDir)

	// Is hugo already installed?
	spartamage.Log("Checking for Hugo version: %s", hugoVersion)

	hugoOutput, hugoOutputErr := sh.Output(hugoPath, "version")
	if hugoOutputErr == nil && strings.Contains(hugoOutput, hugoVersion) {
		spartamage.Log("Hugo version %s already installed at %s", hugoVersion, hugoPath)
		return nil
	}

	hugoArchiveName := ""
	switch runtime.GOOS {
	case "darwin":
		hugoArchiveName = "macOS-64bit.tar.gz"
	case "linux":
		hugoArchiveName = "Linux-64bit.tar.gz"
	default:
		hugoArchiveName = fmt.Sprintf("UNSUPPORTED_%s", runtime.GOOS)
	}

	hugoURL := fmt.Sprintf("https://github.com/gohugoio/hugo/releases/download/v%s/hugo_extended_%s_%s",
		hugoVersion,
		hugoVersion,
		hugoArchiveName)

	spartamage.Log("Installing Hugo from source: %s", hugoURL)
	outputArchive := filepath.Join(localWorkDir, "hugo.tar.gz")
	outputFile, outputErr := os.Create(outputArchive)
	if outputErr != nil {
		return outputErr
	}

	hugoResp, hugoRespErr := http.Get(hugoURL)
	if hugoRespErr != nil {
		return hugoRespErr
	}
	defer hugoResp.Body.Close()

	_, copyBytesErr := io.Copy(outputFile, hugoResp.Body)
	if copyBytesErr != nil {
		return copyBytesErr
	}
	// Great, go heads and untar it...
	unarchiver := archiver.NewTarGz()
	unarchiver.OverwriteExisting = true
	untarErr := unarchiver.Unarchive(outputArchive, localWorkDir)
	if untarErr != nil {
		return untarErr
	}
	versionScript := [][]string{
		{hugoPath, "version"},
	}
	return spartamage.Script(versionScript)
}

// DocsBuild builds the public documentation site in the /docs folder
func DocsBuild() error {
	cleanDocsDirectory := func() error {
		docsDir, docsDirErr := filepath.Abs("docs")
		if docsDirErr != nil {
			return docsDirErr
		}
		spartamage.Log("Cleaning output directory: %s", docsDir)
		return os.RemoveAll(docsDir)
	}

	mg.SerialDeps(DocsInstallRequirements,
		cleanDocsDirectory,
		docsCopySourceTemplatesToDocs)
	return runHugoCommand()
}

// DocsCommit builds and commits the current
// documentation with an autogenerated comment
func DocsCommit() error {
	mg.SerialDeps(DocsBuild)

	commitNoMessageScript := make([][]string, 0)
	for _, eachPath := range hugoDocsPaths {
		commitNoMessageScript = append(commitNoMessageScript,
			[]string{"git", "add", "--all", eachPath},
		)
	}
	commitNoMessageScript = append(commitNoMessageScript,
		[]string{"git", "commit", "-m", `"Documentation updates"`},
	)
	return spartamage.Script(commitNoMessageScript)
}

// DocsEdit starts a Hugo server and hot reloads the documentation at http://localhost:1313
func DocsEdit() error {
	mg.SerialDeps(DocsInstallRequirements,
		docsCopySourceTemplatesToDocs)

	editCommandArgs := []string{
		"server",
		"--disableFastRender",
		"--watch",
		"--forceSyncStatic",
		"--verbose",
	}
	go func() {
		spartamage.Log("Waiting for docs to build...")
		time.Sleep(3 * time.Second)
		browser.OpenURL("http://localhost:1313")
	}()
	return runHugoCommand(editCommandArgs...)
}

////////////////////////////////////////////////////////////////////////////////
// END - DOCUMENTATION
////////////////////////////////////////////////////////////////////////////////

// GenerateAutomaticCode is the handler that runs the codegen part of things
func GenerateAutomaticCode() error {
	// First one is the embedded metric format
	// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch_Embedded_Metric_Format_Specification.html
	args := []string{"aws/cloudwatch/emf.schema.json",
		"--capitalization",
		"AWS",
		"--capitalization",
		"emf",
		"--output",
		"aws/cloudwatch/emf.go",
		"--package",
		"cloudwatch",
	}
	if mg.Verbose() {
		args = append(args, "--verbose")
	}
	return sh.Run("gojsonschema", args...)
}

// GenerateBuildInfo creates the automatic buildinfo.go file so that we can
// stamp the SHA into the binaries we build...
func GenerateBuildInfo() error {
	mg.SerialDeps(EnsureCleanTree)

	// The first thing we need is the `git` SHA
	gitSHA, gitSHAErr := gitCommit(false)
	if gitSHAErr != nil {
		return errors.Wrapf(gitSHAErr, "Failed to get git commit SHA")
	}

	// Super = update the buildinfo data
	buildInfoTemplate := `package sparta

// THIS FILE IS AUTOMATICALLY GENERATED
// DO NOT EDIT
// CREATED: %s

// SpartaGitHash is the commit hash of this Sparta library
const SpartaGitHash = "%s"
`
	updatedInfo := fmt.Sprintf(buildInfoTemplate, time.Now().UTC(), gitSHA)
	// Write it to the output location...
	writeErr := ioutil.WriteFile("./buildinfo.go", []byte(updatedInfo), os.ModePerm)

	if writeErr != nil {
		return writeErr
	}
	commitGenerateCommands := [][]string{
		{"git", "diff"},
		{"git", "commit", "-a", "-m", `"Autogenerated build info"`},
	}
	return spartamage.Script(commitGenerateCommands)

}

// GenerateConstants runs the set of commands that update the embedded CONSTANTS
// for both local and AWS Lambda execution
func GenerateConstants() error {
	generateCommands := [][]string{
		// Remove the tree output
		{"rm",
			"-fv",
			xplatPath(archIconsTreePath),
		},
		//Create the embedded version
		{"esc",
			"-o",
			"./CONSTANTS.go",
			"-private",
			"-pkg",
			"sparta",
			"./resources"},
		//Create a secondary CONSTANTS_AWSBINARY.go file with empty content.
		{"esc",
			"-o",
			"./CONSTANTS_AWSBINARY.go",
			"-private",
			"-pkg",
			"sparta",
			"./resources/awsbinary/README.md"},
		//The next step will insert the
		// build tags at the head of each file so that they are mutually exclusive
		{"go",
			"run",
			"./cmd/insertTags/main.go",
			"./CONSTANTS",
			"!lambdabinary"},
		{"go",
			"run",
			"./cmd/insertTags/main.go",
			"./CONSTANTS_AWSBINARY",
			"lambdabinary"},
		// Create the tree output
		{"tree",
			"-Q",
			"-o",
			xplatPath(archIconsTreePath),
			xplatPath(archIconsRootPath),
		},
		{"git",
			"commit",
			"-a",
			"-m",
			"Autogenerated constants"},
	}
	return spartamage.Script(generateCommands)
}

// EnsurePrealloc ensures that slices that could be preallocated are enforced
func EnsurePrealloc() error {
	// Super run some commands
	preallocCommand := [][]string{
		{"prealloc", "-set_exit_status", "./..."},
	}
	return spartamage.Script(preallocCommand)
}

// CIBuild is the task to build in the context of  CI pipeline
func CIBuild() error {
	mg.SerialDeps(EnsureCIBuildEnvironment,
		Build,
		Test)
	return nil
}

// EnsureMarkdownSpelling ensures that all *.MD files are checked for common
// spelling mistakes
func EnsureMarkdownSpelling() error {
	return markdownSourceApply("misspell", "-error")
}

// EnsureSpelling ensures that there are no misspellings in the source
func EnsureSpelling() error {
	ignoreFiles := []string{
		"CONSTANTS*",
	}
	goSpelling := func() error {
		return goFilteredSourceApply(ignoreFiles, "misspell", "-error")
	}
	mg.SerialDeps(
		goSpelling,
		EnsureMarkdownSpelling)
	return nil
}

// EnsureVet ensures that the source has been `go vet`ted
func EnsureVet() error {
	verboseFlag := ""
	if mg.Verbose() {
		verboseFlag = "-v"
	}
	vetCommand := [][]string{
		{"go", "vet", verboseFlag, "./..."},
	}
	return spartamage.Script(vetCommand)
}

// EnsureLint ensures that the source is `golint`ed
func EnsureLint() error {
	return goSourceApply("golint")
}

// EnsureGoFmt ensures that the source is `gofmt -s` is empty
func EnsureGoFmt() error {

	ignoreGlobs := append(ignoreSubdirectoryPaths,
		"CONSTANTS.go",
		"CONSTANTS_AWSBINARY.go")
	return spartamage.ApplyToSource("go", ignoreGlobs, "gofmt", "-s", "-d")
}

// EnsureFormatted ensures that the source code is formatted with goimports
func EnsureFormatted() error {
	cmd := exec.Command("goimports", "-e", "-d", ".")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	if stdout.String() != "" {
		if mg.Verbose() {
			log.Print(stdout.String())
		}
		return errors.New("`goimports -e -d .` found import errors. Run `goimports -e -w .` to fix them")
	}
	return nil
}

// EnsureStaticChecks ensures that the source code passes static code checks
func EnsureStaticChecks() error {
	// https://staticcheck.io/
	excludeChecks := "-exclude=G204,G505,G401,G404,G601"
	staticCheckErr := sh.Run("staticcheck",
		"github.com/mweagle/Sparta/...")
	if staticCheckErr != nil {
		return staticCheckErr
	}
	// https://github.com/securego/gosec
	if mg.Verbose() {
		return sh.Run("gosec",
			excludeChecks,
			"./...")
	}
	return sh.Run("gosec",
		excludeChecks,
		"-quiet",
		"./...")
}

// LogCodeMetrics ensures that the source code is formatted with goimports
func LogCodeMetrics() error {
	return sh.Run("gocloc", ".")
}

// EnsureAllPreconditions ensures that the source passes *ALL* static `ensure*`
// precondition steps
func EnsureAllPreconditions() error {
	mg.SerialDeps(
		EnsureVet,
		EnsureLint,
		EnsureGoFmt,
		EnsureFormatted,
		EnsureStaticChecks,
		EnsureSpelling,
		EnsurePrealloc,
	)
	return nil
}

// EnsureCIBuildEnvironment is the command that sets up the CI
// environment to run the build.
func EnsureCIBuildEnvironment() error {
	// Super run some commands
	ciCommands := [][]string{
		{"go", "version"},
	}
	return spartamage.Script(ciCommands)
}

// Build the application
func Build() error {
	mg.Deps(EnsureAllPreconditions)
	return sh.Run("go", "build", ".")
}

// Clean the working directory
func Clean() error {
	cleanCommands := [][]string{
		{"go", "clean", "."},
		{"rm", "-rf", "./graph.html"},
		{"rsync", "-a", "--quiet", "--remove-source-files", "./vendor/", "$GOPATH/src"},
	}
	return spartamage.Script(cleanCommands)
}

// Describe runs the `TestDescribe` test to generate a describe HTML output
// file at graph.html
func Describe() error {
	describeCommands := [][]string{
		{"rm", "-rf", "./graph.html"},
		{"go", "test", "-v", "-run", "TestDescribe"},
	}
	return spartamage.Script(describeCommands)
}

// Publish the latest source
func Publish() error {
	mg.SerialDeps(DocsBuild,
		DocsCommit,
		GenerateBuildInfo)

	describeCommands := [][]string{
		{"git", "push", "origin"},
	}
	return spartamage.Script(describeCommands)
}

// UnitTest only runs the unit tests
func UnitTest() error {
	verboseFlag := ""
	if mg.Verbose() {
		verboseFlag = "-v"
	}
	testCommand := [][]string{
		{"go", "test", verboseFlag, "-cover", "-race", "./..."},
	}
	return spartamage.Script(testCommand)
}

// Test runs the Sparta tests
func Test() error {
	mg.SerialDeps(
		EnsureAllPreconditions,
	)
	verboseFlag := ""
	if mg.Verbose() {
		verboseFlag = "-v"
	}
	testCommand := [][]string{
		{"go", "test", verboseFlag, "-cover", "-race", "./..."},
	}
	return spartamage.Script(testCommand)
}

// TestCover runs the test and opens up the resulting report
func TestCover() error {
	mg.SerialDeps(
		EnsureAllPreconditions,
	)
	coverageReport := fmt.Sprintf("%s/cover.out", localWorkDir)
	testCoverCommands := [][]string{
		{"go", "test", fmt.Sprintf("-coverprofile=%s", coverageReport), "."},
		{"go", "tool", "cover", fmt.Sprintf("-html=%s", coverageReport)},
		{"rm", coverageReport},
		{"open", fmt.Sprintf("%s/cover.html", localWorkDir)},
	}
	return spartamage.Script(testCoverCommands)
}

// CompareAgainstMasterBranch is a convenience function to show the comparisons
// of the current pushed branch against the master branch
func CompareAgainstMasterBranch() error {
	// Get the current branch, open a browser
	// to the change...
	// The first thing we need is the `git` branch
	gitInfo, gitInfoErr := sh.Output("git", "rev-parse", "--abbrev-ref", "HEAD")
	if gitInfoErr != nil {
		return gitInfoErr
	}
	stdOutResult := strings.TrimSpace(gitInfo)
	githubURL := fmt.Sprintf("https://github.com/mweagle/Sparta/compare/master...%s", stdOutResult)
	return browser.OpenURL(githubURL)
}
