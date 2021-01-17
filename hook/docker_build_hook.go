package hook

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/session"
	sparta "github.com/mweagle/Sparta"
	spartaDocker "github.com/mweagle/Sparta/docker"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/rs/zerolog"
)

// BuildDockerImageHook is the WorkflowHookHandler responsible for running
// a docker build with the given path, working directory, and provided tags...
func BuildDockerImageHook(dockerFilePath string,
	dockerWorkingDirectory string,
	dockerTags map[string]string) sparta.WorkflowHookHandler {
	dockerBuild := func(ctx context.Context,
		serviceName string,
		S3Bucket gocf.Stringable,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *zerolog.Logger) (context.Context, error) {

		buildErr := spartaDocker.BuildDockerImageInDirectoryWithFlags(serviceName,
			dockerFilePath,
			dockerWorkingDirectory,
			dockerTags,
			"",
			"",
			logger)
		return ctx, buildErr
	}
	return sparta.WorkflowHookFunc(dockerBuild)
}
