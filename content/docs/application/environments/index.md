+++
author = "Matt Weagle"
date = "2016-08-22T21:00:29"
title = "Environments"
tags = ["sparta"]
type = "doc"
+++

# Introduction

It's common for a single Sparta application to target multiple *environments*. For example:

  * Development
  * Staging
  * Production

Each environment is largely similar, but the application may need slightly different configuration in each context.

To support this, Sparta uses Go's [conditional compliation](http://dave.cheney.net/2013/10/12/how-to-use-conditional-compilation-with-the-go-build-tool) support to ensure that configuration information is validate at build time.  Conditional compilation is supported via the `--tags/-t` command line argument.

This example will work through the [SpartaConfig](https://github.com/mweagle/SpartaConfig) sample. The requirement is that each environment declare it's `Name` and also add that value as a Stack [Output](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/outputs-section-structure.html).

## Default Configuration

To start with, create the _default_ configuration. This is the configuration that Sparta uses when provisioning your Stack and defines the environment configuration contract.

### environments/default.go

```golang
// +build !staging,!production

package environments

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/session"
	gocf "github.com/crewjam/go-cloudformation"
	sparta "github.com/mweagle/Sparta"
)

// Name is the default configuration
const Name = ""
```
The important part is the set of excluded tags at the top of the file:

```golang
// +build !staging,!production
```

This ensures that the configuration is only eligible for compilation when Sparta goes to `provision` the service.

## Environment Configuration

The next steps are to define the environment-specific configuration files:

### environments/staging.go
```golang
// +build staging

package environments

import (
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/session"
	gocf "github.com/crewjam/go-cloudformation"
	sparta "github.com/mweagle/Sparta"
)

// Name is the production configuration
const Name = "staging"

```
### environments/production.go

```golang
// +build production

package environments

import (
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/session"
	gocf "github.com/crewjam/go-cloudformation"
	sparta "github.com/mweagle/Sparta"
)

// Name is the production configuration
const Name = "production"

```

These three files define the set of mutually-exclusive compile time files that represent environment targets.

## Segregating Services

Sparta uses the `sparta.Main(ServiceName, ...)` to define the AWS CloudFormation StackName which supports your application.  While the previous files represent different environments, they will collide at `provision` time as we have not specialized the service name.

We can specialize this by including the `buildTags` in the service name with:

```golang
fmt.Sprintf("SpartaHelloWorld-%s", sparta.OptionsGlobal.BuildTags),
```

Each time you run `provision` with a unique `--tags` value, a new CloudFormation stack will be created.


*NOTE*: This isn't something suitable for production, since there could be multiple `BuildTags`.

## Enforcing Environments

The final step is to add the environment name as a Stack Output. To annotate the stack with the output value, we'll register a [ServiceDecorator](https://godoc.org/github.com/mweagle/Sparta#ServiceDecoratorHook) and use the same conditional compilation support to compile the proper version.

The _main.go_ source file registers the workflow hook via:

```golang
hooks := &sparta.WorkflowHooks{
	Context:          map[string]interface{}{},
	ServiceDecorator: environments.ServiceDecoratorHook(sparta.OptionsGlobal.BuildTags),
}
```

Both _environments/staging.go_ and _environments/production.go_ define the same hook function:

```golang
func ServiceDecoratorHook(buildTags string) sparta.ServiceDecoratorHook {
	return func(context map[string]interface{},
		serviceName string,
		template *gocf.Template,
		S3Bucket string,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *logrus.Logger) error {
		template.Outputs["Environment"] = &gocf.Output{
			Description: "Sparta Config target environment",
			Value:       Name,
		}
		return nil
	}
}
```

The _environments/default.go_ definition is slightly different. The _default_ environment isn't one that our service should deploy to. It simply exists to make the initial Sparta build compile.  Build tags are applied to the *AWS Lambda* compiled binary that Sparta generates.

To prevent users from accidentally deploying to the "default" environment, the `BuildTags` are valdiated in the hook definition:

```golang
func ServiceDecoratorHook(buildTags string) sparta.ServiceDecoratorHook {
	return func(context map[string]interface{},
		serviceName string,
		template *gocf.Template,
		S3Bucket string,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *logrus.Logger) error {
		if len(buildTags) <= 0 {
			return fmt.Errorf("Please provide a --tags value for environment target")
		}
		return nil
	}
}
```



## Provisioning

Putting everything together, the `SpartaConfig` service can deploy to either environment:

**staging**

		`go run main.go provision --level info --s3Bucket $(S3_BUCKET) --noop --tags staging`

**production**

		`go run main.go provision --level info --s3Bucket $(S3_BUCKET) --noop --tags production`

Attempting to deploy to **default** will generate an error:

```bash
INFO[0000] Welcome to SpartaConfig-                      Go=go1.7.1 Option=provision SpartaVersion=0.9.2 UTC=2016-10-12T04:07:35Z
INFO[0000] Provisioning service                          BuildID=550c9e360426f48201c885c0abeb078dfc000a0a NOOP=true Tags=
INFO[0000] Verifying IAM Lambda execution roles
INFO[0000] IAM roles verified                            Count=1
INFO[0000] Running `go generate`
INFO[0000] Compiling binary                              Name=SpartaConfig_.lambda.amd64
INFO[0008] Executable binary size                        KB=15309 MB=14
INFO[0008] Creating ZIP archive for upload               TempName=/Users/mweagle/Documents/gopath/src/github.com/mweagle/SpartaConfig/SpartaConfig_104207098
INFO[0009] Registering Sparta function                   FunctionName=main.helloWorld
INFO[0009] Lambda function deployment package size       KB=4262 MB=4
INFO[0009] Bypassing bucket expiration policy check due to -n/-noop command line argument  BucketName=weagle
INFO[0009] Bypassing S3 upload due to -n/-noop command line argument  Bucket=weagle Key=SpartaConfig-/SpartaConfig_104207098
INFO[0009] Calling WorkflowHook                          WorkflowHook=github.com/mweagle/SpartaConfig/environments.ServiceDecoratorHook.func1 WorkflowHookContext=map[]
INFO[0009] Invoking rollback functions                   RollbackCount=0
ERRO[0009] Please provide a --tags value for environment target
Error: Please provide a --tags value for environment target
Usage:
  main provision [flags]

Flags:
  -i, --buildID string    Optional BuildID to use
  -s, --s3Bucket string   S3 Bucket to use for Lambda source
  -t, --tags string       Optional build tags to use for compilation

Global Flags:
  -l, --level string   Log level [panic, fatal, error, warn, info, debug] (default "info")
  -n, --noop           Dry-run behavior only (do not perform mutations)

ERRO[0009] Please provide a --tags value for environment target
exit status 1
```


## Notes

  - Call [ParseOptions](https://godoc.org/github.com/mweagle/Sparta#ParseOptions) to initialize  `sparta.OptionsGlobal.BuildTags` field for use in a service name definition.
  - An alternative approach is to define a custom [ArchiveHook](https://godoc.org/github.com/mweagle/Sparta#ArchiveHook) and inject custom configuration into the ZIP archive. This data is available at `Path.Join(env.LAMBDA_TASK_ROOT, ZIP_ARCHIVE_PATH)`