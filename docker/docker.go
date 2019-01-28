package docker

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/mweagle/Sparta/system"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS
////////////////////////////////////////////////////////////////////////////////

const (
	// BinaryNameArgument is the argument provided to docker build that
	// supplies the local statically built Go binary
	BinaryNameArgument = "SPARTA_DOCKER_BINARY"
)

// BuildDockerImageWithFlags is an extended version of BuildDockerImage that includes
// support for build time tags
func BuildDockerImageWithFlags(serviceName string,
	dockerFilepath string,
	dockerTags map[string]string,
	buildTags string,
	linkFlags string,
	logger *logrus.Logger) error {

	// BEGIN DOCKER PRECONDITIONS
	// Ensure that serviceName and tags are lowercase to make Docker happy
	var dockerErrors []string
	for eachKey, eachValue := range dockerTags {
		if eachKey != strings.ToLower(eachKey) ||
			eachValue != strings.ToLower(eachValue) {
			dockerErrors = append(dockerErrors, fmt.Sprintf("--tag %s:%s MUST be lower case", eachKey, eachValue))
		}
	}

	if len(dockerErrors) > 0 {
		return errors.Errorf("Docker build errors: %s", strings.Join(dockerErrors[:], ", "))
	}
	// BEGIN Informational - output the docker version...
	dockerVersionCmd := exec.Command("docker", "-v")
	dockerVersionCmdErr := system.RunOSCommand(dockerVersionCmd, logger)
	if dockerVersionCmdErr != nil {
		return errors.Wrapf(dockerVersionCmdErr, "Attempting to get docker version")
	}
	// END Informational - output the docker version...

	// END DOCKER PRECONDITIONS

	// Compile this binary for minimal Docker size
	// https://blog.codeship.com/building-minimal-docker-containers-for-go-applications/
	currentTime := time.Now().UnixNano()
	executableOutput := fmt.Sprintf("%s-%d-docker.lambda.amd64", serviceName, currentTime)
	buildErr := system.BuildGoBinary(serviceName,
		executableOutput,
		false,
		fmt.Sprintf("%d", currentTime),
		buildTags,
		linkFlags,
		false,
		logger)
	if buildErr != nil {
		return errors.Wrapf(buildErr, "Attempting to build Docker binary")
	}
	defer func() {
		removeErr := os.Remove(executableOutput)
		if removeErr != nil {
			logger.WithFields(logrus.Fields{
				"Path":  executableOutput,
				"Error": removeErr,
			}).Warn("Failed to delete temporary Docker binary")
		}
	}()

	// ARG SPARTA_DOCKER_BINARY reference s.t. we can supply the binary
	// name to the build..
	// We need to build the static binary s.t. we can add it to the Docker container...
	// Build the image...
	dockerArgs := []string{
		"build",
		"--build-arg",
		fmt.Sprintf("%s=%s", BinaryNameArgument, executableOutput),
	}

	if dockerFilepath != "" {
		dockerArgs = append(dockerArgs, "--file", dockerFilepath)
	}
	// Add the latest tag
	// dockerArgs = append(dockerArgs, "--tag", fmt.Sprintf("sparta/%s:latest", serviceName))
	logger.WithFields(logrus.Fields{
		"Tags": dockerTags,
	}).Info("Creating Docker image")

	for eachKey, eachValue := range dockerTags {
		dockerArgs = append(dockerArgs, "--tag", fmt.Sprintf("%s:%s",
			strings.ToLower(eachKey),
			strings.ToLower(eachValue)))
	}

	dockerArgs = append(dockerArgs, ".")
	dockerCmd := exec.Command("docker", dockerArgs...)
	return system.RunOSCommand(dockerCmd, logger)
}

// BuildDockerImage creates the smallest docker image for this Golang binary
// using the serviceName as the image name and including the supplied tags
func BuildDockerImage(serviceName string,
	dockerFilepath string,
	tags map[string]string,
	logger *logrus.Logger) error {

	return BuildDockerImageWithFlags(serviceName,
		dockerFilepath,
		tags,
		"",
		"",
		logger)
}

// PushDockerImageToECR pushes a local Docker image to an ECR repository
func PushDockerImageToECR(localImageTag string,
	ecrRepoName string,
	awsSession *session.Session,
	logger *logrus.Logger) (string, error) {

	stsSvc := sts.New(awsSession)
	ecrSvc := ecr.New(awsSession)

	// 1. Get the caller identity s.t. we can get the ECR URL which includes the
	// account name
	stsIdentityOutput, stsIdentityErr := stsSvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if stsIdentityErr != nil {
		return "", errors.Wrapf(stsIdentityErr, "Attempting to get AWS caller identity")
	}

	// 2. Create the URL to which we're going to do the push
	localImageTagParts := strings.Split(localImageTag, ":")
	ecrTagValue := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s:%s",
		*stsIdentityOutput.Account,
		*awsSession.Config.Region,
		ecrRepoName,
		localImageTagParts[len(localImageTagParts)-1])

	// 3. Tag the local image with the ECR tag
	dockerTagCmd := exec.Command("docker", "tag", localImageTag, ecrTagValue)
	dockerTagCmdErr := system.RunOSCommand(dockerTagCmd, logger)
	if dockerTagCmdErr != nil {
		return "", errors.Wrapf(dockerTagCmdErr, "Attempting to tag Docker image")
	}

	// 4. Push the image - if that fails attempt to reauthorize with the docker
	// client and try again
	var pushError error
	dockerPushCmd := exec.Command("docker", "push", ecrTagValue)
	pushError = system.RunOSCommand(dockerPushCmd, logger)
	if pushError != nil {
		logger.WithFields(logrus.Fields{
			"Error": pushError,
		}).Info("ECR push failed - reauthorizing")
		ecrAuthTokenResult, ecrAuthTokenResultErr := ecrSvc.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
		if ecrAuthTokenResultErr != nil {
			pushError = ecrAuthTokenResultErr
		} else {
			authData := ecrAuthTokenResult.AuthorizationData[0]
			authToken, authTokenErr := base64.StdEncoding.DecodeString(*authData.AuthorizationToken)
			if authTokenErr != nil {
				pushError = authTokenErr
			} else {
				authTokenString := string(authToken)
				authTokenParts := strings.Split(authTokenString, ":")
				dockerURL := fmt.Sprintf("https://%s.dkr.ecr.%s.amazonaws.com",
					*stsIdentityOutput.Account,
					*awsSession.Config.Region)
				dockerLoginCmd := exec.Command("docker",
					"login",
					"-u",
					authTokenParts[0],
					"--password-stdin",
					dockerURL)
				dockerLoginCmd.Stdout = os.Stdout
				dockerLoginCmd.Stdin = bytes.NewReader([]byte(fmt.Sprintf("%s\n", authTokenParts[1])))
				dockerLoginCmd.Stderr = os.Stderr
				dockerLoginCmdErr := system.RunOSCommand(dockerLoginCmd, logger)
				if dockerLoginCmdErr != nil {
					pushError = dockerLoginCmdErr
				} else {
					// Try it again...
					dockerRetryPushCmd := exec.Command("docker", "push", ecrTagValue)
					dockerRetryPushCmdErr := system.RunOSCommand(dockerRetryPushCmd, logger)
					pushError = dockerRetryPushCmdErr
				}
			}
		}
	}
	if pushError != nil {
		pushError = errors.Wrapf(pushError, "Attempting to push Docker image")
	}
	return ecrTagValue, pushError
}
