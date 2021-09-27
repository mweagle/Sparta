package hook

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws/session"
	sparta "github.com/mweagle/Sparta"
	"github.com/mweagle/Sparta/system"
	"github.com/rs/zerolog"
)

// UPXDockerFile is the docker file used to build the local
// UPX image to do the compression.
const UPXDockerFile = `
FROM alpine:edge
RUN apk add --no-cache upx=3.96-r1
ENTRYPOINT [ "/usr/bin/upx" ]
`

// PostBuildUPXCompressHook returns a WorkflowHookHandler that handles
// UPX compressing the `go` binary containing our functions
func PostBuildUPXCompressHook(dockerImageName string) sparta.WorkflowHookHandler {
	upxHook := func(ctx context.Context,
		serviceName string,
		S3Bucket string,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *zerolog.Logger) (context.Context, error) {
		logger.Info().
			Str("DockerImage", dockerImageName).
			Msg("Compressing binary with UPX")

		outputDir := ctx.Value(sparta.ContextKeyBuildOutputDir).(string)
		outputBinaryName := ctx.Value(sparta.ContextKeyBuildBinaryName).(string)
		absPath, _ := filepath.Abs(outputDir)
		outputBinaryPath := filepath.Join(absPath, outputBinaryName)

		outputCompressedBinaryName := fmt.Sprintf("%s.upx", outputBinaryName)
		outputCompressedBinaryPath := filepath.Join(absPath, outputCompressedBinaryName)
		_, statErr := os.Stat(outputCompressedBinaryPath)
		if !os.IsNotExist(statErr) {
			removeErr := os.Remove(outputCompressedBinaryPath)
			if removeErr != nil {
				logger.Warn().
					Err(removeErr).
					Msg("Failed to delete existing file")
			}
		}

		// Run the UPX packer against the binary with a new name,
		// then overwrite the existing one...
		commandLineArgs := []string{
			"run",
			"--rm",
			"-w",
			"/var/tmp",
			"-v",
			fmt.Sprintf("%s:/var/tmp", absPath),
			dockerImageName,
			"--best",
			"--lzma",
			"-o",
			outputCompressedBinaryName,
			outputBinaryName,
		}
		dockerCompressCmd := exec.Command("docker", commandLineArgs...)
		dockerCompressCmdErr := system.RunOSCommand(dockerCompressCmd, logger)
		if dockerCompressCmdErr != nil {
			return ctx, dockerCompressCmdErr
		}
		return ctx, os.Rename(outputCompressedBinaryPath, outputBinaryPath)
	}
	return sparta.WorkflowHookFunc(upxHook)
}
