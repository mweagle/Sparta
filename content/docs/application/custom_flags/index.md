+++
author = "Matt Weagle"
date = "2016-06-09T17:46:33Z"
title = "Custom Flags"
tags = ["sparta"]
type = "doc"
+++

# Introduction

Some commands (eg: `provision`) may require additional options.  For instance, your application's provision logic may require VPC [subnets](https://aws.amazon.com/blogs/aws/new-access-resources-in-a-vpc-from-your-lambda-functions/) or EC2 [SSH Key Names](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html).  

The default Sparta command line option flags may be extended and validated by building on the exposed [Cobra](https://github.com/spf13/cobra) command objects.

## Adding Flags

To add a flag, use one of the [pflag](https://github.com/spf13/pflag) functions to register your custom flag with one of the standard [CommandLineOption](https://github.com/mweagle/Sparta/blob/master/sparta_main.go#L17) values.

For example:

{{< highlight go >}}

// SSHKeyName is the SSH KeyName to use when provisioning new EC2 instance
var SSHKeyName string

func main() {
  // And add the SSHKeyName option to the provision step
  sparta.CommandLineOptions.Provision.Flags().StringVarP(&SSHKeyName,
    "key",
    "k",
    "",
    "SSH Key Name to use for EC2 instances")
}
{{< /highlight >}}

## Validating Flags

Flags may be used to conditionalize which Sparta lambda functions are provided and/or their content.  In this case, your application may first need to parse and validate the command line input.  

To validate user input, define a [CommandLineOptionsHook](https://godoc.org/github.com/mweagle/Sparta#CommandLineOptionsHook) function and provide it to [sparta.ParseOptions](https://godoc.org/github.com/mweagle/Sparta#ParseOptions).  This function is called after _pflag_ bindings are invoked. 

The result of `ParseOptions` will be the value returned from your validation hook function. If there is an error, your application can exit with an application specific exit code.  For instance:

{{< highlight go >}}
// Define a validation hook s.t. we can verify the SSHKey is valid
validationHook := func(command *cobra.Command) error {
  if command.Name() == "provision" && len(options.SSHKeyName) <= 0 {
    return fmt.Errorf("SSHKeyName option is required")
  }
  return nil
  }
}
// What are the subnets?
parseErr := sparta.ParseOptions(validationHook)
if nil != parseErr {
  os.Exit(3)
}
{{< /highlight >}}

Sparta uses the [govalidator](https://github.com/asaskevich/govalidator/) package to simplify validating command line arguments.  See [sparta_main.go](https://github.com/mweagle/Sparta/blob/master/sparta_main.go) for an example.
