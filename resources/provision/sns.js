//http://docs.aws.amazon.com/lambda/latest/dg/nodejs-prog-model-context.html
//http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-lambda-function-code.html
var cfnResponse = require('cfn-response');
var AWS = require('aws-sdk');
var awsConfig = new AWS.Config({});
// awsConfig.logger = console;

console.log('NodeJS v.' + process.version + ', AWS SDK v.' + AWS.VERSION);

exports.handler = function(event, context) {
  var responseData = {};
  console.log('SNS handler');
  try {
    var sns = new AWS.SNS(awsConfig);
    var mode = (event && event.ResourceProperties) ?
                event.ResourceProperties.Mode : '';

    var topicArn = (event && event.ResourceProperties) ? event.ResourceProperties.TopicArn : null;
    var lambdaArn = (event && event.ResourceProperties) ?
      event.ResourceProperties.LambdaTarget : null;
    var subscriptionArn =  (event && event.ResourceProperties) ?
      event.ResourceProperties.SubscriptionArn : null;

    var onResult = function(e, response) {
      responseData.error = e ? e.toString() : undefined;
      var status = e ? cfnResponse.FAILED : cfnResponse.SUCCESS;
      if (response && response.SubscriptionArn) {
        // Outputs for the confirmation invocation
        responseData.SubscriptionArn = response.SubscriptionArn;
      }
      cfnResponse.send(event, context, status, responseData);
    };

    if (mode === 'Subscribe' && event.RequestType !== 'Delete') {
      var params = {
        Protocol: 'lambda',
        TopicArn: topicArn,
        Endpoint: lambdaArn
      };
      console.log('Subscribing: ' + JSON.stringify(params, null, ' '));
      sns.subscribe(params, onResult);
    } else if (mode === 'Unsubscribe' && event.RequestType === 'Delete' && subscriptionArn) {
      console.log('Unsubscribing: ' + subscriptionArn);
      sns.unsubscribe({SubscriptionArn: subscriptionArn}, onResult);
    }
    else {
      // Nada
      onResult(null, {});
    }
  } catch (e) {
    responseData.error = e.toString();
    cfnResponse.send(event, context, cfnResponse.FAILED, responseData);
  }
};