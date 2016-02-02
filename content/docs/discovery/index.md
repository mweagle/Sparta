+++
author = "Matt Weagle"
date = "2015-11-29T06:50:17"
title = "Discovery Service"
tags = ["sparta"]
type = "doc"
+++

<span class="label label-warning">TODO: Discovery Service documentation</span>

Source [link](https://github.com/mweagle/SpartaApplication/blob/master/application.go#L263).

# Raw Data

S3 Dynamic info
```json

{
    "Configuration": {
        "S3DynamicBucket5de4436284814c262da3b904c1f3fc73b23cea00": {
            "DomainName": "spartaapplication-s3dynamicbucket5de4436284814c26-ll4cejoliisg.s3.amazonaws.com",
            "Ref": "spartaapplication-s3dynamicbucket5de4436284814c26-ll4cejoliisg",
            "Type": "AWS::S3::Bucket",
            "WebsiteURL": "http://spartaapplication-s3dynamicbucket5de4436284814c26-ll4cejoliisg.s3-website-us-west-2.amazonaws.com"
        },
        "aws:cloudformation:stack-id": "arn:aws:cloudformation:us-west-2:123412341234:stack/SpartaApplication/c25fab50-c904-11e5-acca-503f20f2ade6",
        "aws:cloudformation:stack-name": "SpartaApplication",
        "golangFunc": "main.echoS3DynamicBucketEvent",
        "sparta:cloudformation:region": "us-west-2"
    },
    "Event": "{\"key3\":\"value3\",\"key2\":\"value2\",\"key1\":\"value1\"}",
    "RequestID": "638ab6d2-c906-11e5-ac8b-9ff428a230ba",
    "level": "info",
    "msg": "Request received",
    "time": "2016-02-01T17:08:20Z"
}
```

SES Configuration
```json
{
    "Configuration": {
        "SESMessageStoreBucketa622fdfda5789d596c08c79124f12b978b3da772": {
            "DomainName": "spartaapplication-sesmessagestorebucketa622fdfda5-1b8t1fol64if3.s3.amazonaws.com",
            "Ref": "spartaapplication-sesmessagestorebucketa622fdfda5-1b8t1fol64if3",
            "Tags": [
                {
                    "Key": "sparta:logicalBucketName",
                    "Value": "Special"
                }
            ],
            "Type": "AWS::S3::Bucket",
            "WebsiteURL": "http://spartaapplication-sesmessagestorebucketa622fdfda5-1b8t1fol64if3.s3-website-us-west-2.amazonaws.com"
        },
        "aws:cloudformation:stack-id": "arn:aws:cloudformation:us-west-2:123412341234:stack/SpartaApplication/c25fab50-c904-11e5-acca-503f20f2ade6",
        "aws:cloudformation:stack-name": "SpartaApplication",
        "golangFunc": "main.echoSESEvent",
        "sparta:cloudformation:region": "us-west-2"
    },
    "Error": null,
    "level": "info",
    "msg": "Discovery results",
    "time": "2016-02-01T17:10:27Z"
}
```
