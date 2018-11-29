---
date: 2018-01-22 21:49:38
title: CloudFront Distribution
weight: 10
alwaysopen: false
---

The [CloudFrontDistributionDecorator](https://godoc.org/github.com/mweagle/Sparta/decorator#CloudFrontSiteDistributionDecorator) associates a CloudFront Distribution with your S3-backed website. It is implemented as a [ServiceDecoratorHookHandler](https://godoc.org/github.com/mweagle/Sparta#ServiceDecoratorHookHandler) as a single service can only provision one CloudFront distribution.

Sample usage:

```go

////////////////////////////////////////////////////////////////////////////////
// CloudFront settings
const subdomain = "mySiteSubdomain"

// The domain managed by Route53.
const domainName = "myRoute53ManagedDomain.net"

// The site will be available at
// https://mySiteSubdomain.myRoute53ManagedDomain.net

// The S3 bucketname must match the subdomain.domain
// name pattern to serve as a CloudFront Distribution target
var bucketName = fmt.Sprintf("%s.%s", subdomain, domainName)

func distroHooks(s3Site *sparta.S3Site) *sparta.WorkflowHooks {

  // Commented out demonstration of how to front the site
  // with a CloudFront distribution.
  // Note that provisioning a distribution will incur additional
  // costs
  hooks := &sparta.WorkflowHooks{}
  siteHookDecorator := spartaDecorators.CloudFrontSiteDistributionDecorator(s3Site,
    subdomain,
    domainName,
    gocf.String(os.Getenv("SPARTA_ACM_CLOUDFRONT_ARN")))
  hooks.ServiceDecorators = []sparta.ServiceDecoratorHookHandler{
    siteHookDecorator,
  }
  return hooks
}

```
