//http://docs.aws.amazon.com/lambda/latest/dg/nodejs-prog-model-context.html
//http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-lambda-function-code.html
var cfnResponse = require('cfn-response');
var AWS = require('aws-sdk');
var awsConfig = new AWS.Config({});
var path = require('path');
var util = require('util');
var fs = require('fs');
var _ = require('underscore');
var os = require('os');
var AdmZip = require('adm-zip');
var async = require('async');
var mime = require('mime-types');

//awsConfig.logger = console;

console.log('NodeJS v.' + process.version + ', AWS SDK v.' + AWS.VERSION);

/**
Three properties in the event:
  SourceBucket
  SourceKey
  TargetBucket
*/
var pushContent = function(event, callback) {
  var s3 = new AWS.S3(awsConfig);
  var eventProps = event.ResourceProperties || {};
  var localFile = path.join(os.tmpdir(), util.format('s3site-%s.zip', process.hrtime().join()));
  var tasks = [];

  tasks[0] = function(taskCB)
  {
    var params = {Bucket: eventProps.SourceBucket, Key: eventProps.SourceKey};
    var file = fs.createWriteStream(localFile, { flags: 'w', encoding: null, mode: 0666 });
    s3.getObject(params).
      on('httpData',
          function(chunk) {
            file.write(chunk);
        }).
      on('httpDone',
          function() {
              file.end();
              file.on('finish', function() {
                taskCB(null, 'Source downloaded');
              });
          }).
      on('error',
          function(error) {
            taskCB(error);
          }).
    send();
  };
  tasks[1] = function(taskCB)
  {
    // TODO - figure out the name of the S3 buccket that actually was created.
    var targetBucket = eventProps.TargetBucket.split(':').pop();

    // Upload all the files...
    var errors = [];
    var uploadCount = 0;
    var uploadQueue = async.queue(function(zipEntry, queueCB) {
      uploadCount += 1;
      // TODO - retry PUT
      var onResult = function(e) {
        if (e) {
          errors.push(e);
          queueCB(e);
        } else {
          queueCB(null);
        }
      };
      var params = {
        Bucket: targetBucket,
        Key: zipEntry.entryName,
        Body: zip.readFile(zipEntry),
        ContentType: mime.lookup(zipEntry.entryName) || 'application/octet-stream'
      };
      s3.putObject(params, onResult);
    });

    // assign a callback
    uploadQueue.drain = function() {
      var e = _.isEmpty(errors) ? null : new Error(errors.join(';'));
      var result = {
        'TotalFiles' : uploadCount
      };
      taskCB(e, result);
    };

    // reading archives
    var zip = new AdmZip(localFile);
    zip.getEntries().forEach(function(zipEntry) {
        uploadQueue.push(zipEntry);
    });
  };
  async.series(tasks, callback);
};

var deleteContent = function(event, callback) {
  var s3 = new AWS.S3(awsConfig);
  var eventProps = event.ResourceProperties || {};
  var targetBucket = eventProps.TargetBucket.split(':').pop();
  var errors = [];

  var deleteQueue = async.queue(function (s3ListItemResponse, queueCB) {
    var params = {
      Bucket: targetBucket,
      Key: s3ListItemResponse.Key
    };
    var onResult = function(e)
    {
      if (e) {
        errors.push(e);
      }
      queueCB(e);
    };
    s3.deleteObject(params, onResult);
  });
  // assign a callback
  deleteQueue.drain = function() {
    // Only exit if we're drained and there is no outstanding
    // request with an iterator
    if (_.isEmpty(marker)) {
      callback(null, null);
    } else {
      console.log('Waiting for next return: ' + marker);
    }
  };

  // Wait until we actually get some data
  deleteQueue.pause();
  var marker = null;
  var baseListParams = {
    Bucket: targetBucket,
    MaxKeys: 128,
  };

  var onListObjects = function(e, results) {
    if (deleteQueue.paused) {
      deleteQueue.resume();
    }
    marker = null;
    if (e) {
      callback(e);
    }
    else {
      var contents = results.Contents || [];
      if (contents.length <= 0)
      {
        // Nothing to do  - the queue will drain when it's done...
      }
      else {
        var lastObj = contents[contents.length-1] || {};
        marker = results.IsTruncated ? lastObj.Key : null;
        contents.forEach(function (eachEntry) {
          deleteQueue.push(eachEntry);
        });
      }
    }
  };
  var listObjects = function() {
    var params = baseListParams;
    if (marker)
    {
      params.Marker = marker;
    }
    s3.listObjects(params, onListObjects);
  };
  listObjects();
};

exports.handler = function(event, context) {
  var responseData = {};
  try {
    // Download the ZIP archive from the bucket, decompress it, and push
    // everything to the target bucket...go

    // If this is a delete request, then
    var fn = (event.RequestType !== 'Delete') ? pushContent : deleteContent;
    var onComplete = function(e, results)
    {
      responseData.error = e ? e.toString() : undefined;
      responseData.results = results || undefined;
      var status = e ? cfnResponse.FAILED : cfnResponse.SUCCESS;
      cfnResponse.send(event, context, status, responseData);
    };
    fn(event, onComplete);
  } catch (e) {
    responseData.error = e.toString();
    cfnResponse.send(event, context, cfnResponse.FAILED, responseData);
  }
};
