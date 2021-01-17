---
date: 2020-12-31 21:48:47
title: OCI
weight: 800
---

As of December 2020, Lambda functions can also be packaged as [OCI](https://docs.aws.amazon.com/lambda/latest/dg/images-create.html) compatible container
images. Your Sparta application can leverage this capability by providing
a _dockerFile_ argument to `provision`. The contents of your _Dockerfile_ must
produce a container with a reserved label so that Sparta can determine what
ECR instance to push your container.

Sparta docker build [arguments](https://docs.aws.amazon.com/lambda/latest/dg/images-create.html):

- **AWS_ACCOUNT_ID** : The AWS account ID used to build or provision
- **AWS_REGION** : The target deploy region
- **SPARTA_BINARY** : The relative path to the Sparta binary
- **SPARTA_ECR_LABEL_NAME** : The reserved label name that your build must use in order for Sparta to determine the ECR to push the image.
- **SPARTA_BUILD_ID** : The current build ID

A simple HelloWorld _Dockerfile_ is as follows:

```dockerfile
FROM public.ecr.aws/lambda/go:1
ARG AWS_ACCOUNT_ID
ARG AWS_REGION
ARG SPARTA_BINARY
ARG SPARTA_ECR_LABEL_NAME
ARG SPARTA_BUILD_ID
LABEL ${SPARTA_ECR_LABEL_NAME}=${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/spartaoci:${SPARTA_BUILD_ID}
COPY ${SPARTA_BINARY} /var/task/SpartaOCI
WORKDIR /var/task
CMD [ "SpartaOCI" ]
```

Providing the reserved **LABEL** value allows Sparta to locally tag, authenticate
with the ECR, push the image, and finally use the uploaded CodeURI as a
provisioning argument:

```text
...
01 Jan 21 20:27 PST |INFO | Successfully built 95a7695342c5                  io=stdout
01 Jan 21 20:27 PST |INFO | Successfully tagged sparta/myocistack-123412341234:58e1819785b3a40e1642879c60b0716dad873021 io=stdout
01 Jan 21 20:27 PST |INFO | '123412341234.dkr.ecr.us-west-2.amazonaws.com/spartaoci:58e1819785b3a40e1642879c60b0716dad873021' io=stdout
01 Jan 21 20:27 PST |INFO | Provisioning service                             InPlaceUpdates=false NOOP=false Params={"ArtifactS3Bucket":"weagle"} Tags={"io:sparta:buildId":"58e1819785b3a40e1642879c60b0716dad873021"} Template=.sparta/MyOCIStack_123412341234-cftemplate.json
01 Jan 21 20:27 PST |INFO | Checking S3 region                               Bucket=weagle CredentialsRegion=us-west-2 Region=us-west-2
01 Jan 21 20:27 PST |INFO | Checking S3 versioning policy                    Bucket=weagle Region=us-west-2 VersioningEnabled=true
01 Jan 21 20:27 PST |INFO | Pushing local image to ECR                       Tag=123412341234.dkr.ecr.us-west-2.amazonaws.com/spartaoci:58e1819785b3a40e1642879c60b0716dad873021
01 Jan 21 20:27 PST |INFO | Uploading                                        Bucket=weagle Key=MyOCIStack-123412341234/MyOCIStack_123412341234-cftemplate.json Path=.sparta/MyOCIStack_123412341234-cftemplate.json Size="2.6 kB"
01 Jan 21 20:27 PST |INFO | The push refers to repository [123412341234.dkr.ecr.us-west-2.amazonaws.com/spartaoci] io=stdout
01 Jan 21 20:27 PST |INFO | 4a307d712ef3: Preparing                          io=stdout
01 Jan 21 20:27 PST |INFO | c9041e2fe22e: Preparing                          io=stdout
...
```

## Notes

- Providing both a _dockerFile_ argument and an _ArchiveHook_ WorkflowHook value will produce a runtime error.
