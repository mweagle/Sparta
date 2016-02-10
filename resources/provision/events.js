// Manages an CloudWatchEvents notification sources:
// https://aws.amazon.com/blogs/aws/new-cloudwatch-events-track-and-respond-to-changes-to-your-aws-resources/
// http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/WhatIsCloudWatchEvents.html
var _ = require('underscore');
var async = require('async');
var cfnResponse = require('./cfn-response');

var AWS = require('aws-sdk');
var awsConfig = new AWS.Config({});
awsConfig.logger = console;

var convergeRuleSetStateDelete = function(cwEvents, rules, callback) {
  var ruleIterator = function(ruleDef, ruleName, iterCB) {

    // This is a bit brittle, since we use the rulename as the params.Ids in
    // convergeRuleSetStateCreate() and depend on that correspondence here.
    var ruleTasks = [];
    ruleTasks[0] = function(seriesCB) {
      var params = {
        Ids: [ruleName],
        Rule: ruleName
      };
      var idempotentDelete = function(e, results) {
        if (e) {
          if (e.toString().indexOf('ResourceNotFoundException: Rule') >= 0) {
            console.log('Target parent rule does not exist, ignoring error');
            e = null;
          }
        }
        seriesCB(e, results);
      };
      cwEvents.removeTargets(params, idempotentDelete);
    };
    ruleTasks[1] = function(seriesCB) {
      var deleteParams = {
        Name: ruleName
      };
      cwEvents.deleteRule(deleteParams, seriesCB);
    };


    async.series(ruleTasks, iterCB);
  };
  async.forEachOf(rules || {}, ruleIterator, callback);
};

var convergeRuleSetStateCreate = function(cwEvents, rules, lambdaTargetArn, callback) {
  var ruleIterator = function(ruleDef, ruleName, iterCB) {
    var ruleTasks = [];

    ruleTasks[0] = function(seriesCB) {
      var params = {
          Name: ruleName,
          Description: ruleDef.Description || "",
          EventPattern: _.isEmpty(ruleDef.EventPattern) ? undefined : ruleDef.EventPattern,
          ScheduleExpression: _.isEmpty(ruleDef.ScheduleExpression) ? undefined : ruleDef.ScheduleExpression,
          State: 'ENABLED'
        };
        cwEvents.putRule(params, seriesCB);
    };

    ruleTasks[1] = function(seriesCB) {
      var rulesTarget = ruleDef.RuleTarget || {};
      var params = {
        Rule: ruleName,
        Targets: [
          {
            Arn: lambdaTargetArn,
            Id: ruleName,
            Input: rulesTarget.Input || undefined,
            InputPath: rulesTarget.InputPath || undefined
          },
        ]
      };
      cwEvents.putTargets(params, seriesCB);
    };
    async.series(ruleTasks, iterCB);
  };
  async.forEachOf(rules || {}, ruleIterator, callback);
};

exports.handler = function(event, context) {
  var responseData = {};
  try {
    var cloudwatchEvents = new AWS.CloudWatchEvents(awsConfig);
    var oldProps = event.OldResourceProperties || {};
    var oldRules = oldProps.Rules || [];

    // New rules?
    var newProps = event.ResourceProperties || {};
    var newRules = newProps.Rules || [];
    var newLambdaTarget = newProps.LambdaTarget || "";

    // If this is an update, delete existing rules...
    var tasks = [];

    // Delete any old rules
    tasks[0] = _.partial(convergeRuleSetStateDelete, cloudwatchEvents, oldRules);

    // Create any new ones?
    if (event.RequestType !== 'Delete') {
      tasks[1] = _.partial(convergeRuleSetStateCreate, cloudwatchEvents, newRules, newLambdaTarget);
    } else {
      tasks[1] = _.partial(convergeRuleSetStateDelete, cloudwatchEvents, newRules);
    }
    var onResult = function(e, response) {
      responseData.error = e ? e.toString() : undefined;
      var status = e ? cfnResponse.FAILED : cfnResponse.SUCCESS;
      if (!e && response) {
        responseData.CloudWatchEventsResults = response;
      }
      cfnResponse.send(event, context, status, responseData);
    };
    async.series(tasks, onResult);
  } catch (e) {
    responseData.error = e.toString();
    cfnResponse.send(event, context, cfnResponse.FAILED, responseData);
  }
};
