// Manages an S3 buckets notification sources:
// http://docs.aws.amazon.com/AWSJavaScriptSDK/latest/AWS/S3.html#putBucketNotificationConfiguration-property
// http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-lambda-function-code.html
var response = require('cfn-response');
var AWS = require('aws-sdk');
var awsConfig = new AWS.Config({});
// awsConfig.logger = console;
console.log('NodeJS v.' + process.version + ', AWS SDK v.' + AWS.VERSION);
var s3 = new AWS.S3(awsConfig);
exports.handler = function(event, context) {
  var responseData = {};
  try {
    var onUpdateConfigResponse = function(e, updateResponse) {
      responseData.error = e ? e.toString() : undefined;
      responseData.update = updateResponse ? updateResponse : undefined;
      response.send(event, context, e ? response.FAILED : response.SUCCESS, responseData);
    };

    var onResponse = function(e, configResponse) {
      if (e) {
        responseData.Error = e.toString();
        response.send(event, context, response.FAILED, responseData);
      } else if (event.ResourceProperties) {
        var props = event.ResourceProperties;
        var addIDs = (event.RequestType !== 'Delete') ? [props.LambdaTarget] : [];
        var removeArns = [];
        if (event.OldResourceProperties && event.OldResourceProperties.LambdaTarget) {
          removeArns.push(event.OldResourceProperties.LambdaTarget);
        }
        if (event.RequestType === 'Delete') {
          removeArns.push(props.LambdaTarget);
        }
        var lambdas = configResponse.LambdaFunctionConfigurations || [];
        addIDs.forEach(function() {
          lambdas.push(props.Permission);
          lambdas[lambdas.length - 1].LambdaFunctionArn = props.LambdaTarget;
        });
        var pruned = {};
        lambdas.forEach(function(eachConfig) {
          var arnKey = eachConfig.LambdaFunctionArn;
          if (removeArns.indexOf(arnKey) === -1) {
            pruned[arnKey] = eachConfig;
          }
        });
        configResponse.LambdaFunctionConfigurations = Object.keys(pruned).map(function (e) {return pruned[e];});
        // Put it back
        var logMsg = {
          Remove: removeArns,
          Add: addIDs,
          Event: event,
          Config: configResponse,
          Type: event.RequestType
        };
        console.log('S3 Config: ' + JSON.stringify(logMsg));
        s3.putBucketNotificationConfiguration({
          Bucket: event.ResourceProperties.Bucket,
          NotificationConfiguration: configResponse
        }, onUpdateConfigResponse);
      } else {
        response.send(event, context, response.FAILED, {
          'Error': 'Invalid props'
        });
      }
    };
    var params = {
      Bucket: event.ResourceProperties.Bucket
    };
    s3.getBucketNotificationConfiguration(params, onResponse);
  } catch (e) {
    responseData.awsLoaded = e.toString();
    response.send(event, context, response.FAILED, responseData);
  }
};