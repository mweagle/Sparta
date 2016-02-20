// Manages a CloudWatch Logs event source:
//
//
var _ = require('underscore');
var async = require('async');
var cfnResponse = require('./cfn-response');
var AWS = require('aws-sdk');
var awsConfig = new AWS.Config({});
//awsConfig.logger = console;

var convergeLogFiltersDelete = function(cloudWatchLogs, filtersMap, callback) {

  var filterIterator = function(filterDef, filterName, iterCB) {
    var idempotentDelete = function(e, results) {
      if (e) {
        if (e.toString().indexOf('ResourceNotFoundException: ') >= 0) {
          console.log('Target parent rule does not exist, ignoring error');
          e = null;
        }
      }
      iterCB(e, results);
    };

    var params = {
      filterName: filterName,
      logGroupName: filterDef.LogGroupName
    };
    cloudWatchLogs.deleteSubscriptionFilter(params, idempotentDelete);
  };
  async.forEachOf(filtersMap || {}, filterIterator, callback);
};


var convergeLogFiltersCreate = function(cloudWatchLogs, filtersMap, lambdaTargetArn, callback) {

  var filterIterator = function(filterDef, filterName, iterCB) {
    var params = {
      destinationArn: lambdaTargetArn,
      filterName: filterName,
      filterPattern: filterDef.FilterPattern,
      logGroupName: filterDef.LogGroupName
    };
    cloudWatchLogs.putSubscriptionFilter(params, iterCB);
  };
  async.forEachOf(filtersMap || {}, filterIterator, callback);
};

exports.handler = function(event, context) {
  var responseData = {};
  try {
    var cloudWatchLogs = new AWS.CloudWatchLogs(awsConfig);
    var oldProps = event.OldResourceProperties || {};
    var oldFilters = oldProps.Filters || {};

    // New rules?
    var newProps = event.ResourceProperties || {};
    var newFilters = newProps.Filters || {};
    var newLambdaTarget = newProps.LambdaTarget || "";

    // If this is an update, delete existing rules...
    var tasks = [];

    // Delete any old rules
    tasks[0] = _.partial(convergeLogFiltersDelete, cloudWatchLogs, oldFilters);

    // Create any new ones?
    if (event.RequestType !== 'Delete') {
      tasks[1] = _.partial(convergeLogFiltersCreate, cloudWatchLogs, newFilters, newLambdaTarget);
    } else {
      tasks[1] = _.partial(convergeLogFiltersDelete, cloudWatchLogs, newFilters);
    }
    var onResult = function(e, response) {
      responseData.error = e ? e.toString() : undefined;
      var status = e ? cfnResponse.FAILED : cfnResponse.SUCCESS;
      if (!e && response) {
        responseData.CloudWatchLogsResult = response;
      }
      cfnResponse.send(event, context, status, responseData);
    };
    async.series(tasks, onResult);
  } catch (e) {
    responseData.error = e.toString();
    cfnResponse.send(event, context, cfnResponse.FAILED, responseData);
  }
};
