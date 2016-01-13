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
