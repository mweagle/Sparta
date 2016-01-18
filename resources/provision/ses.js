var cfnResponse = require('./cfn-response');
var AWS = require('aws-sdk');
var awsConfig = new AWS.Config({});
//awsConfig.logger = console;

var _ = require('underscore');
var async = require('async');
var toBoolean = require('./sparta_utils').toBoolean;

var SPARTA_RULE_SET_NAME = 'SpartaRuleSet';

// Ensure that the SpartaRuleSet actually exists
var ensureSpartaRuleSet = function(ses, callback) {
  var ruleSetParams = {
    RuleSetName: SPARTA_RULE_SET_NAME
  };
  var onResult = function(e /*, result */ ) {
    if (e) {
      if (e.toString().indexOf('RuleSetDoesNotExist') >= 0) {
        ses.createReceiptRuleSet(ruleSetParams, callback);
      } else {
        callback(e, null);
      }
    } else {
      // All good, keep going...
      callback(null, null);
    }
  };
  ses.describeReceiptRuleSet(ruleSetParams, onResult);
};

var convergeRuleSetStateDelete = function(ses, rules, callback) {
  var ruleIterator = function(rule, iterCB) {
    // What are we supposed to do?
    var deleteParams = {
      RuleSetName: SPARTA_RULE_SET_NAME,
      RuleName: rule.Name
    };
    ses.deleteReceiptRule(deleteParams, iterCB);
  };
  async.eachSeries(rules, ruleIterator, callback);
};

var convergeRuleSetStateCreate = function(ses, rules, callback) {
  var ruleIterator = function(rule, iterCB) {
    var createParams = {
      RuleSetName: SPARTA_RULE_SET_NAME,
      Rule: rule
    };
    // Ensure boolean types
    createParams.Rule.Enabled = toBoolean(createParams.Rule.Enabled);
    createParams.Rule.ScanEnabled = toBoolean(createParams.Rule.ScanEnabled);
    ses.createReceiptRule(createParams, iterCB);
  };
  // We're going to reverse the rules so that iff we are creating,
  // we can leave the "After" field blank and rules will be inserted in the
  // proper rank order
  rules.reverse();
  async.eachSeries(rules, ruleIterator, callback);
};

exports.handler = function(event, context) {
  var responseData = {};
  try {
    var sns = new AWS.SES(awsConfig);
    var oldProps = event.OldResourceProperties || {};
    var oldRules = oldProps.Rules || [];

    // New rules?
    var newProps = event.ResourceProperties || {};
    var newRules = newProps.Rules || [];

    // Domain doesn't matter - we just need to update the 'SpartaRuleSet'.  The
    // golang JSON serialization scopes the Rule names with the servicename
    // to handle collisions under the same ruleset name
    var tasks = [];

    // First verify it's there.
    tasks[0] = _.partial(ensureSpartaRuleSet, sns);

    // Then delete any of the old rules
    tasks[1] = _.partial(convergeRuleSetStateDelete, sns, oldRules);
    // And delete any of the new rules
    tasks[2] = _.partial(convergeRuleSetStateDelete, sns, newRules);

    // Create any new ones?
    if (event.RequestType !== 'Delete') {
      tasks[3] = _.partial(convergeRuleSetStateCreate, sns, newRules);
    }
    var onResult = function(e, response) {
      responseData.error = e ? e.toString() : undefined;
      var status = e ? cfnResponse.FAILED : cfnResponse.SUCCESS;
      if (!e && response) {
        responseData.SESResults = response;
      }
      cfnResponse.send(event, context, status, responseData);
    };
    async.series(tasks, onResult);
  } catch (e) {
    responseData.error = e.toString();
    cfnResponse.send(event, context, cfnResponse.FAILED, responseData);
  }
};
