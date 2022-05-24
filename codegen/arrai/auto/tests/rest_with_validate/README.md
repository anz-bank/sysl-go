## Purpose

Test of codegen for validation of request/response types and its integration with struct tags.

Sysl-go supports the `validate` annotation that can be used to validate field values of request and response objects.

#### Example

For example, the following Sysl specification:

```sysl
App:
  /query/{age <: int}/{height <: int} [validate="age:min=0,max=100 height:min=0"]:
    GET:
      return ok <: Person

  !type Person:
    name <: string
    age <: int [validate="min=0,max=100"]
    height <: int [validate="min=0"]
    contact <: string [validate="email"]
```

Results in the following generated types:

```go
type GetQueryRequest struct {
    Age    int64 `validate:"min=0,max=100"`
    Height int64 `validate:"min=0"`
}

type Person struct {
    Name    string `json:"name"`
    Age     int64 `json:"age" validate:"min=0,max=100"`
    Height  int64 `json:"height" validate:"min=0"`
    Contact int64 `json:"contact" validate:"email"`
}
```

### Validation

Internally, Sysl-go uses [validate](https://godoc.org/github.com/go-playground/validator/v10) struct tags set against the generated types.

### Request parameters

To validate REST request parameters, the `validate` annotation on the endpoint contains a string in the form `name:tags` where `name` is the name of the request parameter and `tags` is the value of the [validate](https://godoc.org/github.com/go-playground/validator/v10) struct tag:

```sysl
/query/{age <: int} [validate="age:min=0,max=100"]
```

To define values for multiple request parameters, individual `name:tags` definitions are separated by a single space:

```sysl
/query/{age <: int}/{height <: int} [validate="age:min=0,max=100 height:min=0"]
```

In the event that the tag itself contains a space, the `validate` annotation can be provided with a list of `name:tags` values:

```sysl
/query/{age <: int}/{height <: int} [validate=["age:oneof=0 1 2", "height:min=0"]]
```

### Request and response objects

To validate request and response objects, the `validate` annotation on the corresponding property contains the value of the [validate](https://godoc.org/github.com/go-playground/validator/v10) struct tag:

```sysl
type GetQueryRequest struct {
    Age    int64 `validate:"min=0,max=100"`
    Height int64 `validate:"min=0"`
}
```

In some instances it is desirable to permit invalid response objects to be returned. 
To do this, annotate the endpoint with `permit_invalid_response`.
Invalid response objects are permitted to be returned but will be logged as invalid:

```sysl
/query/{age <: int} [~permit_invalid_response]
```

Annotating the endpoint method is also supported in the instance that this behaviour should only affect specific methods: 

```sysl
/query/{age <: int}:
  GET [~permit_invalid_response]:
    ...
  POST:
    ...
```

### Missing required request parameters

Unfortunately request objects were not automatically validated for missing required parameters, 
and due to not wanting to possibly break existing applications by adding new validations, an option was added to be able to turn on this validation by annotating with `validate` on the entire app or individual types.

You can either turn it on for the entire app:
```sysl
App [~validate]:
  ...
```
Or for individual types:
```sysl
App:
  /query:
    POST (Body <: Name [mediatype="application/json", ~body]):
      return ok <: Person

  !type Name [~validate]:
    name <: string

  !type Person:
    name <: string
    age <: int [validate="min=0,max=100"]
    height <: int [validate="min=0"]
    contact <: string [validate="email"]
```
