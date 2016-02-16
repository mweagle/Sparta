var _ = require('underscore');

module.exports.toBoolean = function(value) {
  var bValue = value;
  if (_.isString(bValue)) {
    switch (bValue.toLowerCase().trim()) {
      case "true":
      case "1":
        bValue = true;
        break;
      case "false":
      case "0":
      case null:
        bValue = false;
        break;
      default:
        bValue = false;
    }
  }
  return bValue;
};

module.exports.idempotentDeleteHandler = function(successString, cb) {
  return function(e, results) {
    if (e) {
      if (e.toString().indexOf(successString) >= 0) {
        e = null;
      }
    }
    cb(e, results || true);
  }
};
module.exports.cfnResponseLocalTesting = function() {
  console.log('Using local CFN response object');
  return {
    FAILED : 'FAILED',
    SUCCESS: 'SUCCESS',
    send: function(event, context, status, responseData) {
        var msg = {
          event: event,
          context: context,
          result: status,
          responseData: responseData
        };
        console.log(JSON.stringify(msg, null, ' '));
    }
  };
};
