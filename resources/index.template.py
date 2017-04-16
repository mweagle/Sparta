from __future__ import print_function
import pprint
import json
import sys
import os
from ctypes import *
from botocore.credentials import get_credentials
from botocore.session import get_session

lib = cdll.LoadLibrary("{{ .LibraryName }}")
lib.Lambda.argtypes = [c_char_p,
                        c_char_p,
                        c_char_p,
                        c_char_p,
                        c_char_p,
                        POINTER(c_int),
                        c_char_p,
                        c_int,
                        c_char_p,
                        c_int]
lib.Lambda.restype = c_int

################################################################################
# AWS Lambda limits
# Ref: http://docs.aws.amazon.com/lambda/latest/dg/limits.html
################################################################################
MAX_RESPONSE_SIZE = 6 * 1024 * 1024
response_buffer = create_string_buffer(MAX_RESPONSE_SIZE)
MAX_RESPONSE_CONTENT_TYPE_SIZE = 1024
response_content_type_buffer = create_string_buffer(MAX_RESPONSE_CONTENT_TYPE_SIZE)

def lambda_handler(funcName, event, context):
    try:
        # Need to marshall the string into something we can get to in the
        # Go universe, so for that we can just get a struct
        # with the context. The event content can be passed in as a
        # raw char pointer.
        request = {
            "event" : event,
            "context" : {}
        }
        contextDict = {}
        contextDict["functionName"] = context.function_name
        contextDict["functionVersion"] = context.function_version
        contextDict["invokedFunctionArn"] = context.invoked_function_arn
        contextDict["memoryLimitInMB"] = context.memory_limit_in_mb
        contextDict["awsRequestId"] = context.aws_request_id
        contextDict["logGroupName"] = context.log_group_name
        contextDict["logStreamName"] = context.log_stream_name

        # Identity check...
        if context.identity is not None:
            contextDict["identity"] = {
                "cognitoIdentityId" : context.identity.cognito_identity_id,
                "cognitoIdentityPoolId" : context.identity.cognito_identity_pool_id
            }
        # Client context
        if context.client_context is not None:
            destClientContext = {}
            srcClientContext = context.client_context
            destClientContext["installation_id"] = context.client_context.installation_id
            destClientContext["app_title"] = context.client_context.app_title
            destClientContext["app_version_name"] = context.client_context.app_version_name
            destClientContext["app_version_code"] = context.client_context.app_version_code
            destClientContext["Custom"] = context.client_context.custom
            destClientContext["env"] = context.client_context.env
            contextDict["client_context"] = destClientContext

        # Update it
        request["context"] = contextDict
        memset(response_buffer, 0, MAX_RESPONSE_SIZE)
        memset(response_content_type_buffer, 0, MAX_RESPONSE_CONTENT_TYPE_SIZE)
        exitCode = c_int()

        credentials = get_credentials(get_session())
        bytesWritten = lib.Lambda(funcName.encode('utf-8'),
                                    json.dumps(request),
                                    credentials.access_key,
                                    credentials.secret_key,
                                    credentials.token,
                                    byref(exitCode),
                                    response_content_type_buffer,
                                    MAX_RESPONSE_CONTENT_TYPE_SIZE,
                                    response_buffer,
                                    MAX_RESPONSE_SIZE-1)
        lowercase_content_type = response_content_type_buffer.value.lower()
        if "json" in lowercase_content_type:
            return json.loads(response_buffer.value)
        elif "octet-stream" in lowercase_content_type:
            return bytearray(response_buffer.value)
        elif "binary" in lowercase_content_type:
            return bytearray(response_buffer.value)
        else:
            return response_buffer.value
    except:
        print("Unexpected error:", sys.exc_info()[0])

## Insert auto generated code here...
{{ .PythonFunctions }}