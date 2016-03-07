var crypto = require('crypto');
var util = require('util');
var fs = require('fs');
var response = require('./cfn-response');
var _ = require('underscore');
var async = require('async');
var AWS = require('aws-sdk');
var awsConfig = new AWS.Config({});
//awsConfig.logger = console;

var sparta_utils = require('./sparta_utils');
var toBoolean = sparta_utils.toBoolean;
var apigateway = new AWS.APIGateway(awsConfig);
var lambda = new AWS.Lambda(awsConfig);

var RE_STATEMENT_ALREADY_EXISTS = /ResourceConflictException.*already exists/;

var cachedIntegrationDefaultResponseTemplates = null;

////////////////////////////////////////////////////////////////////////////////
// UTILITY FUNCTIONS
var logResults = function(msgText, e, results) {
  var msg = {
    ERROR: e || undefined,
    RESULTS: results || undefined,
    MESSAGE: msgText
  };
  sparta_utils.log(msg);
};

var statementID = function(lambdaArn) {
  var shasum = crypto.createHash('sha1');
  shasum.update(lambdaArn);
  return util.format('Sparta%s', shasum.digest('hex'));
};

var lamdbdaURI = function(lambdaArn) {
  return util.format('arn:aws:apigateway:%s:lambda:path/2015-03-31/functions/%s/invocations',
    lambda.config.region,
    lambdaArn);
};

var accumulatedStackLambdas = function(resourcesRoot, accumulator) {
  // If this is the API root node, then be a bit flexible
  accumulator = accumulator || [];

  var apiChildren = resourcesRoot.APIResources || {};
  _.each(apiChildren, function(eachValue /*, eachKey */ ) {
    if (eachValue.LambdaArn) {
      accumulator.push(eachValue.LambdaArn);
    }
  });
  var children = resourcesRoot.Children || {};
  _.each(children, function(eachValue /*, eachKey */ ) {
    accumulatedStackLambdas(eachValue, accumulator);
  });
  return accumulator;
};

////////////////////////////////////////////////////////////////////////////////
// BEGIN - DELETE API FUNCTIONS
var ensureLambdaPermissionsDeleted = function(lambdaFunctionArns, callback) {
  var cleanupIterator = function(eachArn, iterCB) {
    var onCleanup = function(e, result) {
      // logResults('removePermission result', null, {
      //   ERROR: e,
      //   RESULTS: result,
      //   ARN: eachArn
      // });
      // If there's an error
      sparta_utils.idempotentDeleteHandler("ResourceNotFoundException", iterCB)(e, result);
    };
    try {
      var params = {
        FunctionName: eachArn,
        StatementId: statementID(eachArn)
      };
      lambda.removePermission(params, onCleanup);
    } catch (e) {
      logResults('Failed to remove permission', e, null);
      setImmediate(onCleanup, e, {});
    }
  };
  async.eachSeries(lambdaFunctionArns, cleanupIterator, callback);
};

var cleanupLambdaPermissions = function(apiResources, callback) {
  var lambdaArns = accumulatedStackLambdas(apiResources || []);
  ensureLambdaPermissionsDeleted(lambdaArns, function() {
    // Ignore the results - best effort
    callback(null, true);
  });
};

var deleteResource = function(restAPIInfo, resourceInfo, callback) {
  var onDelete = function(e, results) {
    if (e) {
      callback(e, null);
    }
    else
    {
      var params = {
        resourceId: resourceInfo.id,
        restApiId: restAPIInfo.id
      };
      var idempotentDelete = sparta_utils.idempotentDeleteHandler("NotFoundException", callback);
      apigateway.deleteResource(params, idempotentDelete);
    }
  };

  if (resourceInfo && !_.isEmpty(resourceInfo.resourceMethods)) {
    var iterDeleteMethod = function(value, methodName, iterCB) {
      var params = {
        httpMethod: methodName,
        resourceId: resourceInfo.id,
        restApiId: restAPIInfo.id
      };

      // Need to describe the method to get the resource entry
      apigateway.deleteMethod(params, iterCB);
    };
    async.forEachOfSeries(resourceInfo.resourceMethods,
                          iterDeleteMethod,
                          onDelete);
  } else {
    setImmediate(onDelete, null, true);
  }
};

var ensureResourcesDeleted = function(restAPIInfo, apiResources, callback) {
  // Describe the rest API, for each resource, remove all methods
  // then remove the resource
  var waterfallTasks = [];

  if (_.isObject(restAPIInfo) && !_.isEmpty(restAPIInfo.id)) {
    waterfallTasks[0] = function(waterfallCB) {
      var params = {
        restApiId: restAPIInfo.id
      };
      apigateway.getResources(params, waterfallCB);
    };

    waterfallTasks[1] = function(restAPIResources, waterfallCB) {
      var iterDelete = function(resourceInfo, cb) {

        // logResults('delete resource state', null, {
        //   REST_INFO: restAPIInfo,
        //   RESOURCE_INFO: resourceInfo
        // });
        deleteResource(restAPIInfo, resourceInfo, cb);
      }
      /**
      "items": [
                {
                 "id": "xc8qp9f7ig",
                 "path": "/"
                },
                {
                 "id": "xuvta4",
                 "parentId": "xc8qp9f7ig",
                 "pathPart": "versions",
                 "path": "/versions",
                 "resourceMethods": {
                  "GET": {},
                  "OPTIONS": {}
                 }
                }
      */
      // logResults('API Gateway resources', null, {
      //   RESOURCES: restAPIResources
      // });
      // Filter the resources so that we only need to delete the
      // top level children, excluding the root path
      var itemsToDelete = _.filter(restAPIResources.items,
                                  function (eachItem) {
                                    var slashCount = (eachItem.path.match(/\//g) || []).length;
                                    return (slashCount==1 && eachItem.path !== '/');
                                  });
      async.eachSeries(itemsToDelete, iterDelete, waterfallCB);
    };
    var terminus = function(e, results) {
      if (e) {
        callback(e, null);
      } else {
        // Best effort permissions cleanup
        cleanupLambdaPermissions(apiResources, callback);
      }
    };
    async.waterfall(waterfallTasks, terminus);
  } else {
    setImmediate(callback, null, true);
  }
};

var ensureAPIDeleted = function(restAPIInfo, apiResources, callback) {
  // After the API is deleted, give a best effort attempt to
  // cleanup the permissions
  var onAPIDeleted = function(e, results) {
    if (e) {
      terminus(e, results);
    } else {
      cleanupLambdaPermissions(apiResources, callback);
    }
  };

  // If the API exists, cleanup leaves to root
  var params = {
    restApiId: restAPIInfo.id
  };
  apigateway.deleteRestApi(params, onAPIDeleted);
};

// END - DELETE API FUNCTIONS
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// BEGIN - CREATE API FUNCTIONS

var ensureRestAPI = function(resourceProperties, callback) {
  var waterfall = [];
  var apiProps = resourceProperties.API || {};

  // Get all the APIs
  waterfall.push(function(cb) {
    apigateway.getRestApis({}, cb);
  });

  waterfall.push(function(restAPIs, cb) {
    var matchingAPI = _.find(restAPIs.items || [], function(eachRestAPI) {
      return eachRestAPI.name === apiProps.Name;
    });
    cb(null, matchingAPI);
  });

  // If we didn't find one, create it...
  var terminus = function(e, matchingAPI) {
    if (e || matchingAPI) {
      callback(e, e ? null : matchingAPI);
    } else {
      var apiProps = resourceProperties.API || {};
      var params = {
        name: apiProps.Name,
        cloneFrom: apiProps.CloneFrom || undefined,
        description: apiProps.Description || undefined
      };

      apigateway.createRestApi(params, callback);
    }
  };
  async.waterfall(waterfall, terminus);
};

var ensureLambdaPermissionCreated = function(lambdaArn, resourceMethodDefinition, rolePolicyCache, callback) {
  var addPermissionParams = {
    Action: 'lambda:InvokeFunction',
    FunctionName: lambdaArn,
    Principal: 'apigateway.amazonaws.com',
    StatementId: statementID(lambdaArn),
  };
  var cachedValues = rolePolicyCache[lambdaArn] || {};
  var matching = _.find(Object.keys(cachedValues), function(eachKey) {
    return cachedValues[eachKey].Sid === addPermissionParams.StatementId;
  });
  if (matching) {
    setImmediate(callback, null, {});
  } else {
    // Add it and cache it...
    var creationTasks = {};
    creationTasks.add = function(asyncCB) {
      var onAddPermission = function(e, result) {
        if (e && RE_STATEMENT_ALREADY_EXISTS.test(e.toString())) {
          logResults('Statement already exists', null, e.toString());
          e = null;
        }
        asyncCB(e, result);
      };
      lambda.addPermission(addPermissionParams, onAddPermission);
    };
    creationTasks.cache = ['add'];
    creationTasks.cache.push(function(asyncCB) {
      var getPolicyParams = {
        FunctionName: lambdaArn
      };
      lambda.getPolicy(getPolicyParams, asyncCB);
    });
    var terminus = function(e, results) {
      if (!e && results.cache) {
        try {
          rolePolicyCache[lambdaArn] = JSON.parse(results.cache.Policy);
          logResults('Cached IAM Role', null, {
            ARN: lambdaArn,
            POLICY: rolePolicyCache[lambdaArn]
          });
        } catch (eParse) {
          e = eParse;
        }
      }
      callback(e, results);
    };
    async.auto(creationTasks, terminus);
  }
};

var ensureAPIResourceMethodsCreated = function(restApiId, awsResourceId, APIDefinition, rolePolicyCache, createdCB) {
  // Iterator applied to each member of the methodOpParams// object
  var methodCreationIterator = function(lambdaArn, methodName, methodDef, cb) {
    var creationTasks = {};
    // Parameters common to all Method-related API calls
    var methodOpParams = function(apiSpecificParams) {
      return _.extend({
        httpMethod: methodDef.HTTPMethod,
        resourceId: awsResourceId,
        restApiId: restApiId,
      }, apiSpecificParams || {});
    };

    // 1. Create the Method entry
    // Create the method
    creationTasks.putMethod = function(asyncCB) {
      // Ensure the request params are booleans
      var requestParams = _.reduce(methodDef.Parameters || {},
        function(memo, eachParam, eachKey) {
          memo[eachKey] = toBoolean(eachParam);
          return memo;
        }, {});

      // TODO: Support Model creation
      var params = methodOpParams({
        authorizationType: methodDef.AuthorizationType || "NONE",
        apiKeyRequired: toBoolean(methodDef.APIKeyRequired),
        requestParameters: requestParams
          /*,
                 requestModels: methodDef.RequestModels || {},*/
      });
      apigateway.putMethod(params, asyncCB);
    };


    // 1b. load the default ResponseTemplates
    creationTasks.defaultIntegrationResponseTemplates = function(asyncCB) {
      try {
        // Load each file and transform that into a content-type to VTL mapping
        if (!cachedIntegrationDefaultResponseTemplates) {
          var mappingTemplates = {
            json: fs.readFileSync('./apigateway/inputmapping_json.vtl', {
              encoding: 'utf-8'
            }),
            default: fs.readFileSync('./apigateway/inputmapping_json.vtl', {
              encoding: 'utf-8'
            }),
          };
          cachedIntegrationDefaultResponseTemplates = {
            'application/json': mappingTemplates.json,
            'text/plain': mappingTemplates.default,
            'application/x-www-form-urlencoded': mappingTemplates.default,
            'multipart/form-data': mappingTemplates.default
          };
        }
        setImmediate(asyncCB, null, cachedIntegrationDefaultResponseTemplates);
      } catch (e) {
        setImmediate(asyncCB, e, null);
      }
    };

    var putMethodResponseTask = function(statusCode, parameters, models) {
      return function(taskCB) {
        var responseModels = _.reduce(models,
          function(memo, eachModelDef, eachContentType) {
            memo[eachContentType] = eachModelDef.Name;
            return memo;
          }, {});
        var responseParameters = _.reduce(parameters || {},
          function(memo, eachParam, eachKey) {
            memo[eachKey] = toBoolean(eachParam);
            return memo;
          }, {});
        var params = methodOpParams({
          statusCode: statusCode.toString(),
          responseParameters: responseParameters,
          responseModels: responseModels
        });
        apigateway.putMethodResponse(params, taskCB);
      };
    };

    // 2. Create the Method response, which is a map of status codes to
    // response objects
    creationTasks.putMethodResponse = Object.keys(creationTasks);
    creationTasks.putMethodResponse.push(function(asyncCB) {
      var putMethodResponseTasks = [];
      var responses = methodDef.Responses || {};
      _.each(responses, function(eachResponseObject, eachResponseStatus) {
        var models = eachResponseObject.Models || {};
        var parameters = eachResponseObject.Parameters || {};
        putMethodResponseTasks.push(putMethodResponseTask(eachResponseStatus, parameters, models));
      });

      // Run them...
      async.series(putMethodResponseTasks, asyncCB);
    });

    // 3. Create the Method integration
    // Create the method integration
    var putIntegrationTask = function(statusCode, selectionPattern, parameters, templates) {
      return function(taskCB) {
        var params = methodOpParams({
          statusCode: statusCode.toString(),
          selectionPattern: selectionPattern || undefined,
          responseParameters: parameters || {},
          responseTemplates: templates || {}
        });
        apigateway.putIntegrationResponse(params, taskCB);
      };
    };

    creationTasks.putIntegration = Object.keys(creationTasks);
    creationTasks.putIntegration.push(function(asyncCB, context) {
      var integration = methodDef.Integration || {};
      var opParams = {
        type: integration.Type || 'AWS'
      };
      var requestTemplates = _.isEmpty(integration.RequestTemplates) ?
        context.defaultIntegrationResponseTemplates :
        integration.RequestTemplates;

      switch (opParams.type) {
        case 'AWS':
          opParams.cacheKeyParameters = [];
          opParams.requestTemplates = requestTemplates;
          opParams.uri = lamdbdaURI(lambdaArn);
          opParams.integrationHttpMethod = 'POST';
          break;
        case 'MOCK':
          {
            opParams.cacheKeyParameters = [];
            opParams.requestTemplates = requestTemplates;
            opParams.integrationHttpMethod = 'MOCK';
            break;
          }
        default:
          {
            logResults('Unsupported API Gateway type: ' + opParams.type);
          }
      }
      var params = methodOpParams(opParams);
      apigateway.putIntegration(params, asyncCB);
    });

    // 4. Create the integration response
    // The integration responses
    creationTasks.putIntegrationResponse = Object.keys(creationTasks);
    creationTasks.putIntegrationResponse.push(function(asyncCB) {
      var integration = methodDef.Integration || {};
      var responses = integration.Responses || {};
      var putIntegrationResponseTasks = [];
      _.each(responses,
        function(eachResponse, eachStatusCode) {
          putIntegrationResponseTasks.push(putIntegrationTask(eachStatusCode,
            eachResponse.SelectionPattern,
            eachResponse.Parameters,
            eachResponse.Templates));
        });
      async.series(putIntegrationResponseTasks, asyncCB);

    });

    // 5. Punch a hole into the Lambda s.t. this Arn has permission to invoke the function
    // Related: https://forums.aws.amazon.com/message.jspa?messageID=678324
    creationTasks.addPermission = Object.keys(creationTasks);
    creationTasks.addPermission.push(function(asyncCB /*, context */ ) {
      try {
        ensureLambdaPermissionCreated(lambdaArn, methodDef, rolePolicyCache, asyncCB);
      } catch (e) {
        logResults('Failed to addPermission', e, methodDef);
        setImmediate(asyncCB, e, null);
      }
    });

    // When we're done, describe everything to see what it looks like
    creationTasks.methodDescription = Object.keys(creationTasks);
    creationTasks.methodDescription.push(function(asyncCB) {
      apigateway.getMethod(methodOpParams({}), asyncCB);
    });

    // Wrap it up
    var terminus = function(e, createResults) {
      logResults('methodCreationIterator results', e, createResults);
      cb(e, createResults);
    };
    async.auto(creationTasks, terminus);
  };

  // Start the iteration, which requires the Lambda ARN
  // Create the HTTP methods for this item.
  var lambdaArn = APIDefinition.LambdaArn;
  var methods = Object.keys(APIDefinition.Methods);
  async.eachSeries(methods, function(eachMethod, seriesCB) {
    methodCreationIterator(lambdaArn, eachMethod, APIDefinition.Methods[eachMethod], seriesCB);
  }, createdCB);
};

var ensureResourcesCreated = function(restAPIInfo, resourceProperties, callback ) {
  var restApiId = restAPIInfo.id || "";

  var tasks = [];
  // Get the current resources
  tasks.push(function(cb) {
    var params = {
      restApiId: restApiId,
      limit: "100",
    };
    apigateway.getResources(params, cb);
  });

  // Turn them into a {path, resourceID} map
  tasks.push(function(getResults, cb) {
    var resourceIndex = {};
    if (getResults && getResults.items) {
      resourceIndex = _.reduce(getResults.items,
        function(memo, eachItem) {
          memo[eachItem.path] = eachItem.id;
          return memo;
        }, {});
    }
    setImmediate(cb, null, resourceIndex);
  });

  // Create all the resources in the custom data
  tasks.push(function(resourceIndex, taskCB) {
    var lambdaRolePolicyCache = {};

    var workerError = null;

    //////////////////////////////////////////////////////////////////////////
    // The queue worker for resources visited as the
    // visitor descends the "API" property
    var processResourceEntry = function(taskData, processCB) {
      var rootObject = taskData.definition;
      var parentResourceId = taskData.parentId;

      ////////////////////////////////////////////////////////////////////////
      var onProcessComplete = function(e, processTaskResults) {
        workerError = e;
        if (e) {
          logResults('Failed to create resource', e);
        }
        if (!workerError) {
          // Push the parent ID into the child
          var children = rootObject.Children || {};
          var childKeys = Object.keys(children);
          childKeys.forEach(function(eachKey) {
            var task = {
              definition: children[eachKey],
              parentId: processTaskResults.createResource.id
            };
            workerQueue.push(task);
          });
        }
        processCB(workerError);
      };

      ////////////////////////////////////////////////////////////////////////
      // Make sure the PathComponent is already in the resourceIndex
      var processTasks = {};
      processTasks.createResource = function(asyncCB) {
        // If there is a parentId, then create the child resource
        // for this path
        if (parentResourceId) {
          // Create the resource...
          var params = {
            parentId: parentResourceId,
            pathPart: rootObject.PathComponent,
            restApiId: restApiId
          };
          apigateway.createResource(params, asyncCB);
        } else {
          // No need to create a child resource for "/" path
          setImmediate(asyncCB, null, {
            id: resourceIndex["/"]
          });
        }
      };

      ////////////////////////////////////////////////////////////////////////
      // Create the Methods
      processTasks.createMethods = ['createResource'];
      processTasks.createMethods.push(function(asyncCB, context) {
        var createResourceResponse = context.createResource || {};
        logResults('createResource response', null, createResourceResponse);

        // The API resources will be created a of the root resource
        // or the previously created resource id subpath component
        var resourceId = createResourceResponse.id || resourceIndex["/"];
        var apiResources = rootObject.APIResources || {};
        var apiKeys = Object.keys(apiResources);
        var onAPIResourcesComplete = function(e /*, results */ ) {
          asyncCB(e, e ? null : resourceId);
        };
        async.eachSeries(apiKeys, function(eachKey, itorCB) {
          ensureAPIResourceMethodsCreated(restApiId, resourceId, apiResources[eachKey], lambdaRolePolicyCache, itorCB);
        }, onAPIResourcesComplete);
      });
      async.auto(processTasks, onProcessComplete);
    };

    // Setup the queue to descend
    var workerQueue = async.queue(processResourceEntry, 1);
    workerQueue.drain = function() {
      taskCB(workerError, true);
    };
    var apiDefinition = resourceProperties.API || {};
    var rootResourceDefinition = apiDefinition.Resources || {};
    workerQueue.push({
      definition: rootResourceDefinition,
      parentId: null
    });
  });
  async.waterfall(tasks, callback);
};

var ensureDeployment = function(restAPIInfo, resourceProperties, callback) {

  var restApiId = restAPIInfo.id || "";

  var apiDefinition = resourceProperties.API || {};
  var stageDefinition = apiDefinition.Stage || {};
  if (stageDefinition.Name) {
    var deployTasks = [];
    deployTasks.push(function(taskCB) {
      var params = {
        restApiId: restApiId,
        stageName: stageDefinition.Name,
        cacheClusterEnabled: ("true" === stageDefinition.CacheClusterEnabled),
        cacheClusterSize: _.isEmpty(stageDefinition.CacheClusterSize) ? undefined : stageDefinition.CacheClusterSize,
        stageDescription: stageDefinition.Description || '',
        variables: stageDefinition.Variables || {}
      };
      logResults('Creating deployment', null, params);
      var terminus = function(e, results) {
        if (!e && results) {
          results.URL = util.format('https://%s.execute-api.%s.amazonaws.com/%s',
            restApiId,
            lambda.config.region,
            stageDefinition.Name);
        }
        taskCB(e, results);
      };
      apigateway.createDeployment(params, terminus);
    });
    async.waterfall(deployTasks, callback);
  } else {
    // No stage requested
    logResults('Stage not requested', null, restApiId);
    setImmediate(callback, null, restApiId);
  }
};
exports.handler = function(event, context) {
  console.log('Request type: ' + event.RequestType);

  var data = {};

  var onComplete = function(error, returnValue) {
    sparta_utils.log({
      ERROR: error || undefined,
      RESULTS: returnValue || undefined,
      MESSAGE: 'API Gateway JS results'
    });

    data.Error = error || undefined;
    data.Result = returnValue || undefined;
    data.URL = (data.Result && data.Result.ensureDeployment) ? data.Result.ensureDeployment.URL : "";

    try {
      response.send(event, context, data.Error ? response.FAILED : response.SUCCESS, data);
    } catch (e) {
      logResults('ALL DONE', error, returnValue);
    }
  };

  if (event.ResourceProperties) {

    var resourceProps = event.ResourceProperties || {};
    var apiProps = resourceProps.API || {};

    var oldResourceProps = event.OldResourceProperties || {};
    var oldAPIProps = oldResourceProps.API || {};

    var tasks = {};

    // Issue https://github.com/mweagle/Sparta/issues/6
    // Given the API name, find the RestAPI that owns the
    // APIGateway and try to reuse it s.t. AWS assigned domain
    // name is stable

    tasks.ensureRestAPI = _.partial(ensureRestAPI, resourceProps);

    // // If this is an update and the name has changed, delete the old one
    // if (!_.isEmpty(oldAPIProps) && (oldAPIProps.Name !== apiProps.Name)) {
    //   tasks.ensureOldRestAPI = ['ensureRestAPI',
    //     _.partial(ensureRestAPI, oldResourceProps)
    //   ];
    //
    //   tasks.ensureOldAPIDeleted = ['ensureOldRestAPI',
    //     function(taskCB, context) {
    //       ensureAPIDeleted(context.ensureOldRestAPI, oldAPIProps.Resources, taskCB);
    //     }
    //   ];
    // }
    //
    switch (event.RequestType) {
      case 'Delete':
        {
          tasks.ensureDeleted = ['ensureRestAPI',
            function(taskCB, context) {
              ensureAPIDeleted(context.ensureRestAPI, apiProps.Resources, taskCB);
            }
          ];
          break;
        }
      default:
        {
          // Delete the legacy resources for this API
          tasks.deleteResources = ['ensureRestAPI',
            function(taskCB, context) {
              ensureResourcesDeleted(context.ensureRestAPI, apiProps.Resources, taskCB);
            }
          ];

          // Create the new resources
          tasks.ensureResources = ['deleteResources',
            function(taskCB, context) {
              ensureResourcesCreated(context.ensureRestAPI, resourceProps, taskCB);
            }
          ];

          tasks.ensureDeployment = ['ensureResources',
            function(taskCB, context) {
              ensureDeployment(context.ensureRestAPI, resourceProps, taskCB);
            }
          ];
        }
    }
    async.auto(tasks, onComplete);
  } else {
    logResults('Resource properties not found');
    response.send(event, context, response.SUCCESS, data);
  }
};
