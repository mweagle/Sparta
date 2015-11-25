var util = require('util');
var fs = require('fs');
var http = require('http');
var path = require('path');
var child_process = require('child_process');

//TODO: See if https://forums.aws.amazon.com/message.jspa?messageID=633802
// has been updated with new information
process.env.PATH = process.env.PATH + ':/var/task';


var SPARTA_BINARY_NAME = 'Sparta.lambda.amd64';
var SPARTA_BINARY_PATH = path.join('/tmp', SPARTA_BINARY_NAME);
var MAXIMUM_RESPAWN_COUNT = 5;

var PROXIED_MODULES = ['s3', 'sns', 'apigateway'];

var golangProcess = null;
var failCount = 0;

function makeRequest(path, event, context) {
  var requestBody = {
    event: event,
    context: context
  };

  var stringified = JSON.stringify(requestBody);
  var contentLength = Buffer.byteLength(stringified, 'utf-8');
  var options = {
    host: 'localhost',
    port: 9999,
    path: path,
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Content-Length': contentLength
    }
  };

  var req = http.request(options, function(res) {
    res.setEncoding('utf8');
    var body = '';
    res.on('data', function(chunk) {
      body += chunk;
    });
    res.on('end', function() {
      var err = (res.statusCode >= 400) ? new Error(body) : null;
      var doneValue = ((res.statusCode >= 200) && (res.statusCode <= 299)) ? body : null;
      try {
        // TODO: Check content-type before parse attempt
        if (doneValue)
        {
          doneValue = JSON.parse(doneValue);
        }
      } catch (e) {
        // NOP
      }
      context.done(err, doneValue);
    });
  });
  req.on('error', function(e) {
    context.done(e, null);
  });
  req.write(stringified);
  req.end();
}

var log = function(obj_or_string)
{
  if (typeof(obj_or_string) !== 'object')
  {
    obj_or_string = {msg: obj_or_string};
  }
  if (obj_or_string.stack)
  {
    console.error();
  }
  else
  {
    console.log(JSON.stringify(obj_or_string));
  }
};

// Move the file to /tmp to temporarily work around
// https://forums.aws.amazon.com/message.jspa?messageID=583910
var ensureGoLangBinary = function(callback)
{
    try
    {
      log('Checking for binary: ' + SPARTA_BINARY_PATH);
      fs.statSync(SPARTA_BINARY_PATH);
      setImmediate(callback, null);
    }
    catch (e)
    {
      log('Copying golang binary');
      var command = util.format('cp ./%s %s; chmod +x %s',
                                SPARTA_BINARY_NAME,
                                SPARTA_BINARY_PATH,
                                SPARTA_BINARY_PATH);
      child_process.exec(command, function (err, stdout) {
        if (err)
        {
          console.error(err);
          process.exit(1);
        }
        else
        {
          log(stdout.toString('utf-8'));
        }
        callback(err, stdout);
      });
    }
};

var createForwarder = function(path) {
  var forwardToGolangProcess = function(event, context)
  {
    if (!golangProcess) {
      ensureGoLangBinary(function() {
        golangProcess = child_process.spawn(SPARTA_BINARY_PATH, ['execute', '--signal', process.pid], {});

        golangProcess.stdout.on('data', function(buf) {
          // Just throw it away...
          console.log(buf.toString('utf-8'));
        });
        golangProcess.stderr.on('data', function(buf) {
          // Just throw it away...
          console.log(buf.toString('utf-8'));
        });

        var terminationHandler = function(eventName) {
          return function(value) {
            console.error(util.format('Sparta %s: %s\n', eventName.toUpperCase(), JSON.stringify(value)));
            failCount += 1;
            if (failCount > MAXIMUM_RESPAWN_COUNT) {
              process.exit(1);
            }
            golangProcess = null;
            forwardToGolangProcess(null, null);
          };
        };
        golangProcess.on('error', terminationHandler('error'));
        golangProcess.on('exit', terminationHandler('exit'));
        process.on('exit', function() {
          if (golangProcess) {
            golangProcess.kill();
          }
        });
        var golangProcessReadyHandler = function() {
          process.removeListener('SIGUSR2', golangProcessReadyHandler);
          forwardToGolangProcess(event, context);
        };
        process.on('SIGUSR2', golangProcessReadyHandler);
      });
    }
    else if (event && context)
    {
      makeRequest(path, event, context);
    }
  };
  return forwardToGolangProcess;
};

var sendResponse = function(event, context, e, results)
{
  try {
    var response = require('cfn-response');
    var data = {
      ERROR: e ? e.toString() : undefined,
      RESULTS: results || undefined
    };
    response.send(event, context, e ? response.FAILED : response.SUCCESS, data);
  }
  catch (eResponse) {
    // NOP
    console.log('ERROR sending response: ' + eResponse.toString());
  }
};

// CustomResource Configuration exports
PROXIED_MODULES.forEach(function (eachConfig) {
  var exportName = util.format('%sConfiguration', eachConfig);
  exports[exportName] = function(event, context)
  {
    try {
      var AWS = require('aws-sdk');
      var awsConfig = new AWS.Config({});

      // If the stack is in update mode, don't delegate
      var proxyTasks = [];
      proxyTasks.push(function (taskCB) {
        var params = {
          StackName: event.StackId
        };
        var awsConfig = new AWS.Config({});
        awsConfig.logger = console;
        var cloudFormation = new AWS.CloudFormation(awsConfig);
        cloudFormation.describeStacks(params, taskCB);
      });

      // Only delegate to the stack if the update is in progress.
      var onStackDescription = function(e, stackDescriptionResponse) {
        if (e)
        {
          sendResponse(event, context, e, null);
        }
        else {
          var stackDescription = stackDescriptionResponse.Stacks ? stackDescriptionResponse.Stacks[0] : {};
          var stackStatus = stackDescription.StackStatus || "";
          if (stackStatus !== "UPDATE_COMPLETE_CLEANUP_IN_PROGRESS")
          {
            try {
              console.log('APIGateway handling: ' + event.RequestType);
              console.log(util.format('Delegating [%s] to configurator: %s.\nEvent: %s',
                          event.RequestType, eachConfig,  JSON.stringify(event, null,  ' ')));
              var svc = require(util.format('./%s', eachConfig));
              svc.handler(event, context);
            } catch (e) {
              sendResponse(event, context, e, null);
            }
          } else {
            console.log('Bypassing configurator execution due to status: ' + stackStatus);
            sendResponse(event, context, e, "NOP");
          }
        }
      };
      // Get the current stack status
      var params = {
        StackName: event.StackId
      };
      var cloudFormation = new AWS.CloudFormation(awsConfig);
      cloudFormation.describeStacks(params, onStackDescription);
    } catch (e) {
      console.error('Failed to load configurator:' + e.toString());
      sendResponse(event, context, e, null);
    }
  };
});

exports.main = createForwarder('/');
// Additional golang handlers to be dynamically appended below
