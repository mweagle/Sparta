var util = require('util')
var fs = require('fs')
var http = require('http')
var path = require('path')
var os = require('os')
var process = require('process')
var childProcess = require('child_process')
var spartaUtils = require('./sparta_utils')
var AWS = require('aws-sdk')
var awsConfig = new AWS.Config({})
var GOLANG_CONSTANTS = require('./golang-constants.json')

// TODO: See if https://forums.aws.amazon.com/message.jspa?messageID=633802
// has been updated with new information
process.env.PATH = process.env.PATH + ':/var/task'

// These two names will be dynamically reassigned during archive creation
var SPARTA_BINARY_NAME = 'Sparta.lambda.amd64'
var SPARTA_SERVICE_NAME = 'SpartaService'
// End dynamic reassignment

// This is where the binary will be extracted
var SPARTA_BINARY_PATH = path.join('/tmp', SPARTA_BINARY_NAME)
var MAXIMUM_RESPAWN_COUNT = 5

// Handle to the active golang process.
var golangProcess = null
var failCount = 0

var METRIC_NAMES = {
  CREATED: 'ProcessCreated',
  REUSED: 'ProcessReused',
  TERMINATED: 'ProcessTerminated'
}

var postRequestMetrics = function (path,
  startRemainingCountMillis,
  socketDuration,
  lambdaBodyLength,
  writeCompleteDuration,
  responseEndDuration) {
  var namespace = util.format('Sparta/%s', SPARTA_SERVICE_NAME)

  var params = {
    MetricData: [],
    Namespace: namespace
  }
  var dimensions = [
    {
      Name: 'Path',
      Value: path
    }
  ]
  // Log the uptime with every request...
  params.MetricData.push({
    MetricName: 'Uptime',
    Dimensions: dimensions,
    Unit: 'Seconds',
    Value: os.uptime()
  })
  params.MetricData.push({
    MetricName: 'StartRemainingTimeInMillis',
    Dimensions: dimensions,
    Unit: 'Milliseconds',
    Value: startRemainingCountMillis
  })
  params.MetricData.push({
    MetricName: 'LambdaResponseLength',
    Dimensions: dimensions,
    Unit: 'Bytes',
    Value: lambdaBodyLength
  })

  if (Array.isArray(socketDuration)) {
    params.MetricData.push({
      MetricName: util.format('OpenSocketDuration'),
      Dimensions: dimensions,
      Unit: 'Milliseconds',
      Value: Math.floor(socketDuration[0] / 1000 + socketDuration[1] * 1e9)
    })
  }

  if (Array.isArray(writeCompleteDuration)) {
    params.MetricData.push({
      MetricName: util.format('RequestCompleteDuration'),
      Dimensions: dimensions,
      Unit: 'Milliseconds',
      Value: Math.floor(writeCompleteDuration[0] / 1000 + writeCompleteDuration[1] * 1e9)
    })
  }

  if (Array.isArray(responseEndDuration)) {
    params.MetricData.push({
      MetricName: util.format('ResponseCompleteDuration'),
      Dimensions: dimensions,
      Unit: 'Milliseconds',
      Value: Math.floor(responseEndDuration[0] / 1000 + responseEndDuration[1] * 1e9)
    })
  }

  var cloudwatch = new AWS.CloudWatch(awsConfig)
  var onResult = function () {
    // NOP
  }
  cloudwatch.putMetricData(params, onResult)
}

function makeRequest (path, startRemainingCountMillis, event, context, lambdaCallback) {
  // http://docs.aws.amazon.com/lambda/latest/dg/nodejs-prog-model-context.html
  context.callbackWaitsForEmptyEventLoop = false

  // Let's track the request lifecycle
  var requestTime = process.hrtime()
  var lambdaBodyLength = 0
  var socketDuration = null
  var writeCompleteDuration = null
  var responseEndDuration = null

  var requestBody = {
    event: event,
    context: context
  }
  // If there is a request.event.body element, try and parse it to make
  // interacting with API Gateway a bit simpler.  The .body property
  // corresponds to the data shape set by the *.vtl templates
  if (requestBody && requestBody.event && requestBody.event.body) {
    try {
      requestBody.event.body = JSON.parse(requestBody.event.body)
    } catch (e) {}
  }
  var stringified = JSON.stringify(requestBody)
  var contentLength = Buffer.byteLength(stringified, 'utf-8')
  var options = {
    host: 'localhost',
    port: 9999,
    path: path,
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Content-Length': contentLength
    }
  }

  var onProxyComplete = function (err, response) {
    try {
      responseEndDuration = process.hrtime(requestTime)
      lambdaCallback(err, response)

      postRequestMetrics(path,
        startRemainingCountMillis,
        socketDuration,
        lambdaBodyLength,
        writeCompleteDuration,
        responseEndDuration)
    } catch (e) {
      spartaUtils.log({ERROR: e.toString()})
    }
  }

  var req = http.request(options, function (res) {
    res.setEncoding('utf8')
    var body = ''
    res.on('data', function (chunk) {
      body += chunk
    })
    res.on('end', function () {
      // Bridge the NodeJS and golang worlds by including the golang
      // HTTP status text in the error response if appropriate.  This enables
      // the API Gateway integration response to use standard golang StatusText regexp
      // matches to manage HTTP status codes.
      var responseData = {}
      var handlerError = (res.statusCode >= 400) ? new Error(body) : undefined
      if (handlerError) {
        responseData.code = res.statusCode
        responseData.status = GOLANG_CONSTANTS.HTTP_STATUS_TEXT[res.statusCode.toString()]
        responseData.headers = res.headers
        responseData.error = handlerError.toString()
      } else {
        responseData = body
        lambdaBodyLength = Buffer.byteLength(responseData, 'utf8')
        if (res.headers['content-type'] === 'application/json') {
          try {
            responseData = JSON.parse(body)
          } catch (e) {}
        }
      }
      var err = handlerError ? new Error(JSON.stringify(responseData)) : null
      var resp = handlerError ? null : responseData
      onProxyComplete(err, resp)
    })
  })
  req.once('socket', function (res) {
    socketDuration = process.hrtime(requestTime)
  })
  req.once('finish', function () {
    writeCompleteDuration = process.hrtime(requestTime)
  })
  req.once('error', function (e) {
    onProxyComplete(e, null)
  })
  req.write(stringified)
  req.end()
}

var postMetricCounter = function (metricName, userCallback) {
  var namespace = util.format('Sparta/%s', SPARTA_SERVICE_NAME)

  var params = {
    MetricData: [
      {
        MetricName: metricName,
        Unit: 'Count',
        Value: 1
      }
    ],
    Namespace: namespace
  }
  var cloudwatch = new AWS.CloudWatch(awsConfig)
  var onResult = function () {
    if (userCallback) {
      userCallback()
    }
  }
  cloudwatch.putMetricData(params, onResult)
}

// Move the file to /tmp to temporarily work around
// https://forums.aws.amazon.com/message.jspa?messageID=583910
var ensureGoLangBinary = function (callback) {
  try {
    fs.statSync(SPARTA_BINARY_PATH)
    setImmediate(callback, null)
  } catch (e) {
    var command = util.format('cp ./%s %s; chmod +x %s',
      SPARTA_BINARY_NAME,
      SPARTA_BINARY_PATH,
      SPARTA_BINARY_PATH)
    childProcess.exec(command, function (err, stdout) {
      if (err) {
        console.error(err)
        process.exit(1)
      } else {
        spartaUtils.log(stdout.toString('utf-8'))
      // Post the
      }
      callback(err, stdout)
    })
  }
}

var createForwarder = function (path) {
  var forwardToGolangProcess = function (event, context, callback, metricName, startRemainingCountMillisParam) {
    var startRemainingCountMillis = startRemainingCountMillisParam || context.getRemainingTimeInMillis()
    if (!golangProcess) {
      ensureGoLangBinary(function () {
        spartaUtils.log(util.format('Launching %s with args: execute --signal %d', SPARTA_BINARY_PATH, process.pid))
        golangProcess = childProcess.spawn(SPARTA_BINARY_PATH, ['execute', '--signal', process.pid], {})

        golangProcess.stdout.on('data', function (buf) {
          buf.toString('utf-8').split('\n').forEach(function (eachLine) {
            spartaUtils.log(eachLine)
          })
        })
        golangProcess.stderr.on('data', function (buf) {
          buf.toString('utf-8').split('\n').forEach(function (eachLine) {
            spartaUtils.log(eachLine)
          })
        })

        var terminationHandler = function (eventName) {
          return function (value) {
            var onPosted = function () {
              console.error(util.format('Sparta %s: %s\n', eventName.toUpperCase(), JSON.stringify(value)))
              failCount += 1
              if (failCount > MAXIMUM_RESPAWN_COUNT) {
                process.exit(1)
              }
              golangProcess = null
              forwardToGolangProcess(null,
                null,
                callback,
                METRIC_NAMES.TERMINATED,
                startRemainingCountMillis)
            }
            postMetricCounter(METRIC_NAMES.TERMINATED, onPosted)
          }
        }
        golangProcess.on('error', terminationHandler('error'))
        golangProcess.on('exit', terminationHandler('exit'))
        process.on('exit', function () {
          spartaUtils.log('Go process exited')
          if (golangProcess) {
            golangProcess.kill()
          }
        })
        var golangProcessReadyHandler = function () {
          spartaUtils.log('SIGUSR2 signal received')
          process.removeListener('SIGUSR2', golangProcessReadyHandler)
          forwardToGolangProcess(event,
            context,
            callback,
            METRIC_NAMES.CREATED,
            startRemainingCountMillis)
        }
        spartaUtils.log('Waiting for SIGUSR2 signal')
        process.on('SIGUSR2', golangProcessReadyHandler)
      })
    }
    else if (event && context) {
      postMetricCounter(metricName || METRIC_NAMES.REUSED)
      makeRequest(path, startRemainingCountMillis, event, context, callback)
    }
  }
  return forwardToGolangProcess
}

// Log the outputs
var envSettings = {
  aws_sdk: AWS.VERSION,
  node_js: process.version,
  os: {
    platform: os.platform(),
    release: os.release(),
    type: os.type(),
    uptime: os.uptime(),
    cpus: os.cpus(),
    totalmem: os.totalmem()
  }
}
spartaUtils.log(envSettings)

exports.main = createForwarder('/')

// Additional golang handlers to be dynamically appended below
