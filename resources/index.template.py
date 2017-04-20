from __future__ import print_function
import pprint
import json
import sys
import traceback
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
        request = dict(event=event)
        contextDict = dict(
            functionName=context.function_name,
            functionVersion=context.function_version,
            invokedFunctionArn=context.invoked_function_arn,
            memoryLimitInMB=context.memory_limit_in_mb,
            awsRequestId=context.aws_request_id,
            logGroupName=context.log_group_name,
            logStreamName=context.log_stream_name
        )

        # Identity check...
        if getattr(context, "identity", None):
            contextDict["identity"] = dict(
                cognitoIdentityId=context.identity.cognito_identity_id,
                cognitoIdentityPoolId=context.identity.cognito_identity_pool_id
            )
        # Client context
        if getattr(context, "client_context", None):
            contextDict["client_context"] = dict(
                installation_id=context.client_context.installation_id,
                app_title=context.client_context.app_title,
                app_version_name=context.client_context.app_version_name,
                app_version_code=context.client_context.app_version_code,
                Custom=context.client_context.custom,
                env=context.client_context.env
            )

        # Update it
        request["context"] = contextDict
        memset(response_buffer, 0, MAX_RESPONSE_SIZE)
        memset(response_content_type_buffer, 0, MAX_RESPONSE_CONTENT_TYPE_SIZE)
        exitCode = c_int()

        credentials = get_credentials(get_session())
        bytesWritten = lib.Lambda(funcName.encode('ascii'),
                                    json.dumps(request).encode('ascii'),
                                    credentials.access_key.encode('ascii'),
                                    credentials.secret_key.encode('ascii'),
                                    credentials.token.encode('ascii'),
                                    byref(exitCode),
                                    response_content_type_buffer,
                                    MAX_RESPONSE_CONTENT_TYPE_SIZE,
                                    response_buffer,
                                    MAX_RESPONSE_SIZE-1)
        lowercase_content_type = response_content_type_buffer.value.lower()
        if "json" in lowercase_content_type.decode('utf-8'):
            return json.loads(response_buffer.value)
        elif "octet-stream" in lowercase_content_type.decode('utf-8'):
            return bytearray(response_buffer.value)
        elif "binary" in lowercase_content_type.decode('utf-8'):
            return bytearray(response_buffer.value)
        else:
            return response_buffer.value
    except:
        traceback.print_exc()
        print("Unexpected error:", sys.exc_info()[0])

## Insert auto generated code here...
{{ .PythonFunctions }}
