+++
author = "Matt Weagle"
date = "2015-11-29T06:50:17"
title = "Alternative Topologies"
tags = ["sparta"]
type = "doc"
+++

# Introduction

At a broad level, AWS Lambda represents a new level of compute abstraction for services. Developers don't immediately concern themselves with HA topologies, configuration management, capacity planning, or many of the other areas traditionally handled by operations. These are handled by the vendor supplied execution environment.

However, Lambda is a relatively new technology and is not ideally suited to certain types of tasks.  For example, given the current [Lambda limits](http://docs.aws.amazon.com/lambda/latest/dg/limits.html), the following task types might better be handled by "legacy" AWS services:

  * Long running tasks 
  * Tasks with significant disk space requirements
  * Large HTTP(S) I/O tasks
  
It may also make sense to integrate EC2 when:

  * Applications are being gradually decomposed into Lambda functions
  * Latency-sensitive request paths can't afford [cold container](https://aws.amazon.com/blogs/compute/container-reuse-in-lambda/) startup times
  * Price/performance justifies using EC2 
  * Using EC2 as a failover for system-wide Lambda outages

For such cases, Sparta supports running the exact same binary on EC2.  This section describes how to create a single Sparta service that publishes a function via AWS Lambda _and_ EC2 as part of the same application codebase. It's based on the [SpartaOmega](https://github.com/mweagle/SpartaOmega) project.

# Mixed Topology

Deploying your application to a mixed topology is accomplished by combining existing Sparta features. There is no "make mixed" command line option. 

## Add Custom Command Line Option

The first step is to add a [custom command line option](/docs/application/custom_commands). This command option will be used when your binary is running in "mixed topology" mode.  The SpartaOmega project starts up a localhost HTTP server, so we'll add a `httpServer` command line option with:

{{< highlight go >}}
// Custom command to startup a simple HelloWorld HTTP server
httpServerCommand := &cobra.Command{
  Use:   "httpServer",
  Short: "Sample HelloWorld HTTP server",
  Long:  `Sample HelloWorld HTTP server that binds to port: ` + HTTPServerPort,
  RunE: func(cmd *cobra.Command, args []string) error {
    http.HandleFunc("/", helloWorldResource)
    return http.ListenAndServe(fmt.Sprintf(":%d", HTTPServerPort), nil)
  },
}
sparta.CommandLineOptions.Root.AddCommand(httpServerCommand)
{{< /highlight >}}

Our command doesn't accept any additional flags. If your command needs additional user flags, consider adding a [ParseOptions](https://godoc.org/github.com/mweagle/Sparta#ParseOptions) call to validate they are properly set.

## Create CloudInit Userdata

The next step is to write a [user-data](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/user-data.html) script that will be used to configure your EC2 instance(s) at startup. Your script is likely to differ from the one below, but at a minimum it will include code to download and unzip the archive containing your Sparta binary. 

{{< highlight bash >}}
#!/bin/bash -xe
SPARTA_OMEGA_BINARY_PATH=/home/ubuntu/{{ .ServiceName }}.lambda.amd64

################################################################################
# 
# Tested on Ubuntu 16.04
#
# AMI: ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-20160516.1 (ami-06b94666)
if [ ! -f "/home/ubuntu/userdata.sh" ]
then
  curl -vs http://169.254.169.254/latest/user-data -o /home/ubuntu/userdata.sh
  chmod +x /home/ubuntu/userdata.sh
fi

# Install everything
service supervisor stop || apt-get install supervisor -y
apt-get update -y 
apt-get upgrade -y 
apt-get install supervisor awscli unzip git -y

################################################################################
# Our own binary
aws s3 cp s3://{{ .S3Bucket }}/{{ .S3Key }} /home/ubuntu/application.zip
unzip -o /home/ubuntu/application.zip -d /home/ubuntu
chmod +x $SPARTA_OMEGA_BINARY_PATH

################################################################################
# SUPERVISOR
# REF: http://supervisord.org/
# Cleanout secondary directory
mkdir -pv /etc/supervisor/conf.d
  
SPARTA_OMEGA_SUPERVISOR_CONF="[program:spartaomega]
command=$SPARTA_OMEGA_BINARY_PATH httpServer
numprocs=1
directory=/tmp
priority=999
autostart=true
autorestart=unexpected
startsecs=10
startretries=3
exitcodes=0,2
stopsignal=TERM
stopwaitsecs=10
stopasgroup=false
killasgroup=false
user=ubuntu
stdout_logfile=/var/log/spartaomega.log
stdout_logfile_maxbytes=1MB
stdout_logfile_backups=10
stdout_capture_maxbytes=1MB
stdout_events_enabled=false
redirect_stderr=false
stderr_logfile=spartaomega.err.log
stderr_logfile_maxbytes=1MB
stderr_logfile_backups=10
stderr_capture_maxbytes=1MB
stderr_events_enabled=false
"
echo "$SPARTA_OMEGA_SUPERVISOR_CONF" > /etc/supervisor/conf.d/spartaomega.conf

# Patch up the directory
chown -R ubuntu:ubuntu /home/ubuntu

# Startup Supervisor
service supervisor restart || service supervisor start
{{< /highlight >}}

The script uses the command line option (`command=$SPARTA_OMEGA_BINARY_PATH httpServer`) that was defined in the first step.

It also uses the `S3Bucket` and `S3Key` properties that Sparta creates during the build and provides to your decorator function (next section).

### Notes

The script is using [text/template](https://golang.org/pkg/text/template/) markup to expand properties known at build time.  Because this content will be parsed by [ConvertToTemplateExpression](https://godoc.org/github.com/mweagle/Sparta/aws/cloudformation#ConvertToTemplateExpression) (next section), it's also possible to use [Fn::Join](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-join.html) compatible JSON serializations (single line only) to reference properties that are known only during CloudFormation provision time.  

For example, if we were also provisioning a PostgreSQL instance and needed to dynamically discover the endpoint address, a shell script variable could be assigned via:

{{< highlight bash >}}
POSTGRES_ADDRESS={ "Fn::GetAtt" : [ "{{ .DBInstanceResourceName }}" , "Endpoint.Address" ] }
{{< /highlight >}}

This expression combines both a build-time variable (`DBInstanceResourceName`: the CloudFormation resource name) and a provision time one (`Endpoint.Address`: dynamically assigned by the CloudFormation [RDS Resource](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-getatt.html)).

## Decorate Toplogy

The final step is to use a [TemplateDecorator](https://godoc.org/github.com/mweagle/Sparta#TemplateDecorator) to tie everything together. A decorator can annotate the CloudFormation template with any supported [go-cloudformation](https://github.com/crewjam/go-cloudformation) resource.  For this example, we'll create a single AutoScalingGroup and EC2 instance that's bootstrapped with our custom _userdata.sh_ script.

{{< highlight go >}}

// The CloudFormation template decorator that inserts all the other
// AWS components we need to support this deployment...
func lambdaDecorator(customResourceAMILookupName string) sparta.TemplateDecorator {

  return func(serviceName string,
    lambdaResourceName string,
    lambdaResource gocf.LambdaFunction,
    resourceMetadata map[string]interface{},
    S3Bucket string,
    S3Key string,
    template *gocf.Template,
    logger *logrus.Logger) error {

    // Create the launch configuration with Metadata to download the ZIP file, unzip it & launch the
    // golang binary...
    ec2SecurityGroupResourceName := sparta.CloudFormationResourceName("SpartaOmegaSecurityGroup",
      "SpartaOmegaSecurityGroup")
    asgLaunchConfigurationName := sparta.CloudFormationResourceName("SpartaOmegaASGLaunchConfig",
      "SpartaOmegaASGLaunchConfig")
    asgResourceName := sparta.CloudFormationResourceName("SpartaOmegaASG",
      "SpartaOmegaASG")
    ec2InstanceRoleName := sparta.CloudFormationResourceName("SpartaOmegaEC2InstanceRole",
      "SpartaOmegaEC2InstanceRole")
    ec2InstanceProfileName := sparta.CloudFormationResourceName("SpartaOmegaEC2InstanceProfile",
      "SpartaOmegaEC2InstanceProfile")

    //////////////////////////////////////////////////////////////////////////////
    // 1 - Create the security group for the SpartaOmega EC2 instance
    ec2SecurityGroup := &gocf.EC2SecurityGroup{
      GroupDescription: gocf.String("SpartaOmega Security Group"),
      SecurityGroupIngress: &gocf.EC2SecurityGroupRuleList{
        gocf.EC2SecurityGroupRule{
          CidrIp:     gocf.String("0.0.0.0/0"),
          IpProtocol: gocf.String("tcp"),
          FromPort:   gocf.Integer(HTTPServerPort),
          ToPort:     gocf.Integer(HTTPServerPort),
        },
        gocf.EC2SecurityGroupRule{
          CidrIp:     gocf.String("0.0.0.0/0"),
          IpProtocol: gocf.String("tcp"),
          FromPort:   gocf.Integer(22),
          ToPort:     gocf.Integer(22),
        },
      },
    }
    template.AddResource(ec2SecurityGroupResourceName, ec2SecurityGroup)
    //////////////////////////////////////////////////////////////////////////////
    // 2 - Create the ASG and associate the userdata with the EC2 init
    // EC2 Instance Role...
    statements := sparta.CommonIAMStatements.Core

    // Add the statement that allows us to fetch the S3 object with this compiled
    // binary
    statements = append(statements, spartaIAM.PolicyStatement{
      Effect:   "Allow",
      Action:   []string{"s3:GetObject"},
      Resource: gocf.String(fmt.Sprintf("arn:aws:s3:::%s/%s", S3Bucket, S3Key)),
    })
    iamPolicyList := gocf.IAMPoliciesList{}
    iamPolicyList = append(iamPolicyList,
      gocf.IAMPolicies{
        PolicyDocument: sparta.ArbitraryJSONObject{
          "Version":   "2012-10-17",
          "Statement": statements,
        },
        PolicyName: gocf.String("EC2Policy"),
      },
    )
    ec2InstanceRole := &gocf.IAMRole{
      AssumeRolePolicyDocument: sparta.AssumePolicyDocument,
      Policies:                 &iamPolicyList,
    }
    template.AddResource(ec2InstanceRoleName, ec2InstanceRole)

    // Create the instance profile
    ec2InstanceProfile := &gocf.IAMInstanceProfile{
      Path:  gocf.String("/"),
      Roles: []gocf.Stringable{gocf.Ref(ec2InstanceRoleName).String()},
    }
    template.AddResource(ec2InstanceProfileName, ec2InstanceProfile)

    //Now setup the properties map, expand the userdata, and attach it...
    userDataProps := map[string]interface{}{
      "S3Bucket":    S3Bucket,
      "S3Key":       S3Key,
      "ServiceName": serviceName,
    }

    userDataTemplateInput, userDataTemplateInputErr := resources.FSString(false, "/resources/source/userdata.sh")
    if nil != userDataTemplateInputErr {
      return userDataTemplateInputErr
    }
    templateReader := strings.NewReader(userDataTemplateInput)
    userDataExpression, userDataExpressionErr := spartaCF.ConvertToTemplateExpression(templateReader, 
                                                                                      userDataProps)
    if nil != userDataExpressionErr {
      return userDataExpressionErr
    }

    logger.WithFields(logrus.Fields{
      "Parameters": userDataProps,
      "Expanded":   userDataExpression,
    }).Debug("Expanded userdata")

    asgLaunchConfigurationResource := &gocf.AutoScalingLaunchConfiguration{
      ImageId:            gocf.GetAtt(customResourceAMILookupName, "HVM"),
      InstanceType:       gocf.String("t2.micro"),
      KeyName:            gocf.String(SSHKeyName),
      IamInstanceProfile: gocf.Ref(ec2InstanceProfileName).String(),
      UserData:           gocf.Base64(userDataExpression),
      SecurityGroups:     gocf.StringList(gocf.GetAtt(ec2SecurityGroupResourceName, "GroupId")),
    }
    launchConfigResource := template.AddResource(asgLaunchConfigurationName,
      asgLaunchConfigurationResource)
    launchConfigResource.DependsOn = append(launchConfigResource.DependsOn,
      customResourceAMILookupName)

    // Create the ASG
    asgResource := &gocf.AutoScalingAutoScalingGroup{
      // Empty Region is equivalent to all region AZs
      // Ref: http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-getavailabilityzones.html
      AvailabilityZones:       gocf.GetAZs(gocf.String("")),
      LaunchConfigurationName: gocf.Ref(asgLaunchConfigurationName).String(),
      MaxSize:                 gocf.String("1"),
      MinSize:                 gocf.String("1"),
    }
    template.AddResource(asgResourceName, asgResource)
    return nil
  }
}
{{< /highlight >}}

There are a few things to point out in this function:

  * **Security Groups** - The decorator adds an ingress rule so that the endpoint is publicly accessible:
{{< highlight go >}}
    gocf.EC2SecurityGroupRule{
          CidrIp:     gocf.String("0.0.0.0/0"),
          IpProtocol: gocf.String("tcp"),
          FromPort:   gocf.Integer(HTTPServerPort),
          ToPort:     gocf.Integer(HTTPServerPort),
        }
{{< /highlight >}}
  * **IAM Role** - In order to download the S3 archive, the EC2 IAM Policy includes a custom privilege:
{{< highlight go >}}
      statements = append(statements, spartaIAM.PolicyStatement{
      Effect:   "Allow",
      Action:   []string{"s3:GetObject"},
      Resource: gocf.String(fmt.Sprintf("arn:aws:s3:::%s/%s", S3Bucket, S3Key)),
    })
{{< /highlight >}}
  * **UserData Marshaling** - Marshaling the _userdata.sh_ script is handled by `ConvertToTemplateExpression`:
{{< highlight go >}}
    // Now setup the properties map, expand the userdata, and attach it...
    userDataProps := map[string]interface{}{
      "S3Bucket":    S3Bucket,
      "S3Key":       S3Key,
      "ServiceName": serviceName,
    }
    // ...
    templateReader := strings.NewReader(userDataTemplateInput)
    userDataExpression, userDataExpressionErr := spartaCF.ConvertToTemplateExpression(templateReader, 
                                                                                      userDataProps)
    // ...
    asgLaunchConfigurationResource := &gocf.AutoScalingLaunchConfiguration{
      // ...
      UserData:           gocf.Base64(userDataExpression),
      // ...
    }
{{< /highlight >}}
  * **Custom Command Line Flags** - To externalize the SSH Key Name, the binary expects a [custom flag](/docs/application/custom_flags) (not shown above):
{{< highlight go >}}
  // And add the SSHKeyName option to the provision step
  sparta.CommandLineOptions.Provision.Flags().StringVarP(&SSHKeyName,
    "key",
    "k",
    "",
    "SSH Key Name to use for EC2 instances")
{{< /highlight >}}
  This value is used as an input to the AutoScalingLaunchConfiguration value:
{{< highlight go >}}
    asgLaunchConfigurationResource := &gocf.AutoScalingLaunchConfiguration{
      // ...
      KeyName:            gocf.String(SSHKeyName),
      // ...
    }
{{< /highlight >}}

# Result

Deploying your Go application using a mixed topology enables your "Lambda" endpoint to be addressable via AWS Lambda and standard HTTP.

## HTTP Access

{{< highlight bash >}}

$ curl -vs http://ec2-52-26-146-138.us-west-2.compute.amazonaws.com:9999/
*   Trying 52.26.146.138...
* Connected to ec2-52-26-146-138.us-west-2.compute.amazonaws.com (52.26.146.138) port 9999 (#0)
> GET / HTTP/1.1
> Host: ec2-52-26-146-138.us-west-2.compute.amazonaws.com:9999
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Date: Fri, 10 Jun 2016 14:58:15 GMT
< Content-Length: 29
< Content-Type: text/plain; charset=utf-8
<
* Connection #0 to host ec2-52-26-146-138.us-west-2.compute.amazonaws.com left intact
Hello world from SpartaOmega!
{{< /highlight >}}

## Lambda Access

![Lambda](/images/alternative_topology/lambda.jpg)

# Conclusion

Mixed topology deployment is a powerful feature that enables your application to choose the right set of resources.  It provides a way for services to non-destructively migrate to AWS Lambda or shift existing Lambda workloads to alternative compute resources.  

# Notes
  - _userdata.sh_ isn't sufficient to reconfigure in response to CloudFormation update events.  Production systems should also include [cfn-hup](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/cfn-hup.html) listeners.
  - Production deployments may consider [CodeDeploy](https://aws.amazon.com/codedeploy/) to assist in HA binary rollover.
  - Forwarding [CloudWatch Logs](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/WhatIsCloudWatchLogs.html) is not handled by this sample. 
  - Consider using HTTPS & [Let's Encrypt](https://ivopetkov.com/b/let-s-encrypt-on-ec2/) on your EC2 instances.
