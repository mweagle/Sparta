---
date: 2018-11-28 20:03:46
title: REST Service
weight: 10
---

The [rest](github.com/mweagle/Sparta/archetype/rest) package provides convenience functions
to define a serverless [REST style](https://en.wikipedia.org/wiki/Representational_state_transfer) service.

The package uses three concepts:

- Routes: URL paths that resolve to a single `go` struct
- Resources: `go` structs that optionally define HTTP method (`GET`, `POST`, etc.).
- [ResourceDefinition](https://godoc.org/github.com/mweagle/Sparta/archetype/rest#ResourceDefinition): an interface that `go` structs must implement in order to support resource-based registration.

## Routes

Routes are similar many HTTP-routing libraries. They support [path parameters](https://docs.aws.amazon.com/apigateway/latest/developerguide/integrating-api-with-aws-services-lambda.html#api-as-lambda-proxy-expose-get-method-with-path-parameters-to-call-lambda-function).

## Resources

Resources are the targets of Routes. There is a one to one mapping of URL Routes to `go` structs.
These `struct` types must define one or more member functions that comply with the
[valid function signatures](https://docs.aws.amazon.com/lambda/latest/dg/go-programming-model-handler-types.html)
for AWS Lambda.

For example:

```go
import (
  spartaAPIGateway "github.com/mweagle/Sparta/aws/apigateway"
)
// TodoItemResource is the /todo/{id} resource
type TodoItemResource struct {
  spartaAccessor.S3Accessor
}

func (svc *TodoItemResource) Get(ctx context.Context,
  apigRequest TodoRequest) (interface{}, error) {
  // ...
  return spartaAPIGateway.NewResponse(http.StatusOK, "All good!"), nil
}
```

As the resource will be exposed over API-Gateway, the return type must be a
struct type created by `spartaAPIGateway.NewResponse` so that the API-Gateway
integration mappings can properly transform the response.

## Resource Definition

The last component is to bind the Routes and Resources together by implementing
the [ResourceDefinition](https://godoc.org/github.com/mweagle/Sparta/archetype/rest#ResourceDefinition)
interface. This interface defines a single function that supplies the
binding information.

For instance, the `TodoItemResource` type might define a REST resource like:

```go
// ResourceDefinition returns the Sparta REST definition for the Todo item
func (svc *TodoItemResource) ResourceDefinition() (spartaREST.ResourceDefinition, error) {

  return spartaREST.ResourceDefinition{
    URL: "/todo/{id}",
    MethodHandlers: spartaREST.MethodHandlerMap{
      // GET
      http.MethodGet: spartaREST.NewMethodHandler(svc.Get, http.StatusOK).
        Options(&sparta.LambdaFunctionOptions{
          MemorySize: 128,
          Timeout:    10,
        }).
        StatusCodes(http.StatusInternalServerError).
        Privileges(svc.S3Accessor.KeysPrivilege("s3:GetObject"),
          svc.S3Accessor.BucketPrivilege("s3:ListBucket")),
    },
  }, nil
}
```

The `ResourceDefinition` function returns a struct that defines the:

- Route (`URL`)
- MethodHandlers (`GET`, `POST`, etc.)

and for each MethodHandler, the:

- HTTP verb to struct function mapping
- Optional [LambdaFunctionOptions](https://godoc.org/github.com/mweagle/Sparta#LambdaFunctionOptions)
- Expected HTTP status codes (`StatusCodes`)
  - Limiting the number of allowed status codes reduces API Gateway creation time
- Additional IAM `Privileges` needed for this method

## Registration

With the REST resource providing its API-Gateway binding information, the final step
is to supply the `ResourceDefinition` implementing instance to
[RegisterResource](https://godoc.org/github.com/mweagle/Sparta/archetype/rest#RegisterResource) and
return the set of extracted `*LambdaAWSInfo` structs:

```go
  todoItemResource := &todoResources.TodoItemResource{}
  registeredFuncs, registeredFuncsErr := spartaREST.RegisterResource(api, todoItemResource)
```

See the [SpartaTodoBackend](https://github.com/mweagle/SpartaTodoBackend) for a complete example that
implements the [TodoBackend](https://todobackend.com) API in a completely serverless way!