var response = require('cfn-response');
var AWS = require('aws-sdk');
console.log('AWS SDK v.' + AWS.VERSION);
exports.handler = function(event, context) {
  var responseData = {};
  console.log('SNS handler');
  try
  {
    // var onResult = function(e, data)
    // {
    //   console.log('Response received');
    //   responseData.SNS = data;
    //   responseData.error = e;
    //   response.send(event, context, response.SUCCESS, responseData);
    // };
    // var sns = new AWS.SNS();
    // var arn = (event && event.ResourceProperties && event.ResourceProperties.Permission) ?
    //     event.ResourceProperties.Permission.TopicArn : 'INVALID';
    //   var params = {
    //     TopicArn: arn
    //   };
    throw new Error('SNS configuration not yet implemented - calling SNS APIs from Lambda hangs CustomResource creation');
    //sns.getTopicAttributes(params, onResult);
  }
  catch (e)
  {
    responseData.event = event;
    responseData.error = e.toString();
    console.log('Failed to create SNS object');
    response.send(event, context, response.SUCCESS, responseData);
  }
};