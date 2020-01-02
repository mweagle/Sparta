package system

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func ensureMainEntrypoint(logger *logrus.Logger) error {
	// Don't do this for "go test" runs
	if flag.Lookup("test.v") != nil {
		logger.Debug("Skipping main() check for test")
		return nil
	}

	fset := token.NewFileSet()
	packageMap, parseErr := parser.ParseDir(fset, ".", nil, parser.PackageClauseOnly)
	if parseErr != nil {
		return errors.Errorf("Failed to parse source input: %s", parseErr.Error())
	}
	logger.WithFields(logrus.Fields{
		"SourcePackages": packageMap,
	}).Debug("Checking working directory")

	// If there isn't a main defined, we're in the wrong directory..
	mainPackageCount := 0
	for eachPackage := range packageMap {
		if eachPackage == "main" {
			mainPackageCount++
		}
	}
	if mainPackageCount <= 0 {
		unlikelyBinaryErr := fmt.Errorf("error: It appears your application's `func main() {}` is not in the current working directory. Please run this command in the same directory as `func main() {}`")
		return unlikelyBinaryErr
	}
	return nil
}

// GoVersion returns the configured go version for this system
func GoVersion(logger *logrus.Logger) (string, error) {
	runtimeVersion := runtime.Version()
	// Get the golang version from the output:
	// Matts-MBP:Sparta mweagle$ go version
	// go version go1.8.1 darwin/amd64
	golangVersionRE := regexp.MustCompile(`go(\d+\.\d+(\.\d+)?)`)
	matches := golangVersionRE.FindStringSubmatch(runtimeVersion)
	if len(matches) > 2 {
		return matches[1], nil
	}
	logger.WithFields(logrus.Fields{
		"Output": runtimeVersion,
	}).Warn("Unable to find Golang version using RegExp - using current version")
	return runtimeVersion, nil
}

// GoPath returns either $GOPATH or the new $HOME/go path
// introduced with Go 1.8
func GoPath() string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		home := os.Getenv("HOME")
		gopath = filepath.Join(home, "go")
	}
	return gopath
}

// BuildGoBinary is a helper to build a go binary with the given options
func BuildGoBinary(serviceName string,
	executableOutput string,
	useCGO bool,
	buildID string,
	userSuppliedBuildTags string,
	linkFlags string,
	noop bool,
	logger *logrus.Logger) error {

	// Before we do anything, let's make sure there's a `main` package in this directory.
	ensureMainPackageErr := ensureMainEntrypoint(logger)
	if ensureMainPackageErr != nil {
		return ensureMainPackageErr
	}
	// Go generate
	cmd := exec.Command("go", "generate")
	if logger.Level == logrus.DebugLevel {
		cmd = exec.Command("go", "generate", "-v", "-x")
	}
	cmd.Env = os.Environ()
	commandString := fmt.Sprintf("%s", cmd.Args)
	logger.Info(fmt.Sprintf("Running `%s`", strings.Trim(commandString, "[]")))
	goGenerateErr := RunOSCommand(cmd, logger)
	if nil != goGenerateErr {
		return goGenerateErr
	}
	// TODO: Smaller binaries via linker flags
	// Ref: https://blog.filippo.io/shrink-your-go-binaries-with-this-one-weird-trick/
	noopTag := ""
	if noop {
		noopTag = "noop "
	}

	buildTags := []string{
		"lambdabinary",
		"linux",
	}
	if noopTag != "" {
		buildTags = append(buildTags, noopTag)
	}
	if userSuppliedBuildTags != "" {
		userBuildTagsParts := strings.Split(userSuppliedBuildTags, " ")
		for _, eachTag := range userBuildTagsParts {
			buildTags = append(buildTags, eachTag)
		}
	}
	userBuildFlags := []string{"-tags", strings.Join(buildTags, " ")}

	// Append all the linker flags
	// Stamp the service name into the binary
	// We need to stamp the servicename into the aws binary so that if the user
	// chose some type of dynamic stack name at provision time, the name
	// we use at execution time has that value. This is necessary because
	// the function dispatch logic uses the AWS_LAMBDA_FUNCTION_NAME environment
	// variable to do the lookup. And in effect, this value has to be unique
	// across an account, since functions cannot have the same name
	// Custom flags for the binary
	linkerFlags := map[string]string{
		"StampedServiceName": serviceName,
		"StampedBuildID":     buildID,
	}
	for eachFlag, eachValue := range linkerFlags {
		linkFlags = fmt.Sprintf("%s -s -w -X github.com/mweagle/Sparta.%s=%s",
			linkFlags,
			eachFlag,
			eachValue)
	}
	linkFlags = strings.TrimSpace(linkFlags)
	if len(linkFlags) != 0 {
		userBuildFlags = append(userBuildFlags, "-ldflags", linkFlags)
	}
	// If this is CGO, do the Docker build if we're doing an actual
	// provision. Otherwise use the "normal" build to keep things
	// a bit faster.
	var cmdError error
	if useCGO {
		currentDir, currentDirErr := os.Getwd()
		if nil != currentDirErr {
			return currentDirErr
		}
		gopathVersion, gopathVersionErr := GoVersion(logger)
		if nil != gopathVersionErr {
			return gopathVersionErr
		}

		gopath := GoPath()
		containerGoPath := "/usr/src/gopath"
		// Get the package path in the current directory
		// so that we can it to the container path
		packagePath := strings.TrimPrefix(currentDir, gopath)
		volumeMountMapping := fmt.Sprintf("%s:%s", gopath, containerGoPath)
		containerSourcePath := fmt.Sprintf("%s%s", containerGoPath, packagePath)

		// If there's one from the environment, use that...
		// TODO

		// Otherwise, make one...

		// Any CGO paths?
		cgoLibPath := fmt.Sprintf("%s/cgo/lib", containerSourcePath)
		cgoIncludePath := fmt.Sprintf("%s/cgo/include", containerSourcePath)

		// Pass any SPARTA_* prefixed environment variables to the docker build
		//
		goosTarget := os.Getenv("SPARTA_GOOS")
		if goosTarget == "" {
			goosTarget = "linux"
		}
		goArch := os.Getenv("SPARTA_GOARCH")
		if goArch == "" {
			goArch = "amd64"
		}
		spartaEnvVars := []string{
			// "-e",
			// fmt.Sprintf("GOPATH=%s", containerGoPath),
			"-e",
			fmt.Sprintf("GOOS=%s", goosTarget),
			"-e",
			fmt.Sprintf("GOARCH=%s", goArch),
			"-e",
			fmt.Sprintf("CGO_LDFLAGS=-L%s", cgoLibPath),
			"-e",
			fmt.Sprintf("CGO_CFLAGS=-I%s", cgoIncludePath),
		}
		// User vars
		for _, eachPair := range os.Environ() {
			if strings.HasPrefix(eachPair, "SPARTA_") {
				spartaEnvVars = append(spartaEnvVars, "-e", eachPair)
			}
		}
		dockerBuildArgs := []string{
			"run",
			"--rm",
			"-v",
			volumeMountMapping,
			"-w",
			containerSourcePath}
		dockerBuildArgs = append(dockerBuildArgs, spartaEnvVars...)
		dockerBuildArgs = append(dockerBuildArgs,
			fmt.Sprintf("golang:%s", gopathVersion),
			"go",
			"build",
			"-o",
			executableOutput,
			"-buildmode=default",
		)
		dockerBuildArgs = append(dockerBuildArgs, userBuildFlags...)
		cmd = exec.Command("docker", dockerBuildArgs...)
		cmd.Env = os.Environ()
		logger.WithFields(logrus.Fields{
			"Name": executableOutput,
			"Args": dockerBuildArgs,
		}).Info("Building `cgo` library in Docker")
		cmdError = RunOSCommand(cmd, logger)

		// If this succeeded, let's find the .h file and move it into the scratch
		// Try to keep things tidy...
		if nil == cmdError {
			soExtension := filepath.Ext(executableOutput)
			headerFilepath := fmt.Sprintf("%s.h", strings.TrimSuffix(executableOutput, soExtension))
			_, headerFileErr := os.Stat(headerFilepath)
			if nil == headerFileErr {
				targetPath, targetPathErr := TemporaryFile(".sparta", filepath.Base(headerFilepath))
				if nil != targetPathErr {
					headerFileErr = targetPathErr
				} else {
					headerFileErr = os.Rename(headerFilepath, targetPath.Name())
				}
			}
			if nil != headerFileErr {
				logger.WithFields(logrus.Fields{
					"Path": headerFilepath,
				}).Warn("Failed to move .h file to scratch directory")
			}
		}
	} else {
		// Build the regular version
		buildArgs := []string{
			"build",
			"-o",
			executableOutput,
		}
		// Debug flags?
		if logger.Level == logrus.DebugLevel {
			buildArgs = append(buildArgs, "-v")
		}
		buildArgs = append(buildArgs, userBuildFlags...)
		buildArgs = append(buildArgs, ".")
		cmd = exec.Command("go", buildArgs...)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "GOOS=linux", "GOARCH=amd64")
		logger.WithFields(logrus.Fields{
			"Name": executableOutput,
		}).Info("Compiling binary")
		cmdError = RunOSCommand(cmd, logger)
	}
	return cmdError
}

// TemporaryFile creates a stable temporary filename in the current working
// directory
func TemporaryFile(scratchDir string, name string) (*os.File, error) {
	workingDir, err := os.Getwd()
	if nil != err {
		return nil, err
	}

	// Use a stable temporary name
	temporaryPath := filepath.Join(workingDir, scratchDir, name)
	buildDir := filepath.Dir(temporaryPath)
	mkdirErr := os.MkdirAll(buildDir, os.ModePerm)
	if nil != mkdirErr {
		return nil, mkdirErr
	}

	tmpFile, err := os.Create(temporaryPath)
	if err != nil {
		return nil, errors.New("Failed to create temporary file: " + err.Error())
	}

	return tmpFile, nil
}
