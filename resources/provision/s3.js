// Manages an S3 buckets notification sources:
// http://docs.aws.amazon.com/AWSJavaScriptSDK/latest/AWS/S3.html#putBucketNotificationConfiguration-property
// http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-lambda-function-code.html
var response = require('cfn-response');
var AWS = require('aws-sdk');
var awsConfig = new AWS.Config({});
//awsConfig.logger = console;

var s3 = new AWS.S3(awsConfig);
var cloudFormation = new AWS.CloudFormation(awsConfig);

exports.handler = function(event, context) {
  var data = {};
  var props = event.ResourceProperties;
  var oldProps = event.OldResourceProperties || {};
  var bucketName = (props.BucketArn || '').split(':').pop();

  var onEnd = function(error, returnValue) {
    data.Error = error || undefined;
    data.Result = returnValue || undefined;
    response.send(event, context, data.Error ? response.FAILED : response.SUCCESS, data);
  };
  var onResponse = function(error, s3Config) {
    if (error) {
      onEnd(error);
    } else {
      var funcs = {};
      var delArns = [];
      var propsArn = props.LambdaTarget;
      if (oldProps.LambdaTarget && oldProps.LambdaTarget !== props.LambdaTarget) {
        delArns.push(oldProps.LambdaTarget);
      }
      if (event.RequestType === 'Delete') {
        delArns.push(props.LambdaTarget);
      } else if (delArns.indexOf(propsArn) === -1) {
        funcs[propsArn] = props.Permission;
        funcs[propsArn].LambdaFunctionArn = propsArn;
      }
      (s3Config.LambdaFunctionConfigurations || []).forEach(function(iConf) {
        var arnKey = iConf.LambdaFunctionArn;
        if (delArns.indexOf(arnKey) === -1) {
          funcs[arnKey] = iConf;
        }
      });
      s3Config.LambdaFunctionConfigurations = Object.keys(funcs).map(function(e) {
        return funcs[e];
      });
      // Put it back
      var logMsg = {
        Del: delArns,
        Event: event,
        Config: s3Config,
        Type: event.RequestType
      };
      console.log('Result: ' + JSON.stringify(logMsg));
      s3.putBucketNotificationConfiguration({
        Bucket: bucketName,
        NotificationConfiguration: s3Config
      }, onEnd);
    }
  };

  if (event.RequestType === 'Delete') {
    var onDescribeStacks = function(e, del) {
      if (e) {
        onEnd(e);
      } else {
        var stackStatus = del.Stacks[0] ? del.Stacks[0].StackStatus : '';
        if (stackStatus !== 'UPDATE_COMPLETE_CLEANUP_IN_PROGRESS') {
          s3.getBucketNotificationConfiguration({
            Bucket: bucketName
          }, onResponse);
        } else {
          onEnd(null, del.Stacks[0]);
        }
      }
    };
    cloudFormation.describeStacks({
      StackName: event.StackId
    }, onDescribeStacks);
  } else {
    s3.getBucketNotificationConfiguration({
      Bucket: bucketName
    }, onResponse);
  }
};
