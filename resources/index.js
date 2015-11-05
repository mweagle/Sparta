var util = require('util');
var fs = require('fs');
var http = require('http');
var path = require('path');
var child_process = require('child_process');

//https://forums.aws.amazon.com/message.jspa?messageID=633802
process.env.PATH = process.env.PATH + ':/var/task';

var SPARTA_BINARY_NAME = 'Sparta.lambda.amd64';
var SPARTA_BINARY_PATH = path.join('/tmp', SPARTA_BINARY_NAME);
var MAXIMUM_RESPAWN_COUNT = 5;

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

exports.main = createForwarder('/');
// Additional golang handlers to be dynamically appended below