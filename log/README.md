<h1>Logging</h1>

- [Overview](#overview)
- [Logged Events](#logged-events)
  - [Error Events](#error-events)
  - [Info Events](#info-events)
  - [Debug Events](#debug-events)
  - [Event Fields](#event-fields)
- [Usage](#usage)
- [Framework](#framework)
- [External Configuration](#external-configuration)
- [Custom Configuration](#custom-configuration)
- [Native Support](#native-support)
- [Legacy Support](#legacy-support)
  - [Logrus](#logrus)
  - [Pkg logger](#pkg-logger)

# Overview 

Sysl-go comes equipped with flexible, out-of-the-box logging support.
The following sections describe what gets logged and how the logger can be configured and utilised within custom code.

# Logged Events

Sysl-go logs the following major categories and fields:

## Error Events

Sysl-go logs all errors encountered within the running on an application.
Examples include timeout errors, marshalling errors and errors encountered in custom code.

## Info Events

Sysl-go logs all information for the purpose of understanding the state of the application and its requests.
Examples include the server configuration on startup and the state of upstream and downstream requests.

## Debug Events

Sysl-go logs verbose information for the purpose of debugging.
Examples include JWT authorisation failures and request and response payloads (see [External Configuration](#external-configuration) below).

## Event Fields

Sysl-go includes attaches additional field values to some of its logs.
The following notable fields are attached:

| Field | Details |  
|---|---|  
| `traceid` | HTTP: Identifier retrieved from `RequestID` header or uniqely generated for the purpose of tracing a request. |
| `remote` | HTTP: The remote address of a downstream request. |
| `latency` | The time required to fulfil a request. | 

## Errors

Sysl-go logs all errors encountered within the running on an application.

# Usage

Sysl-go utilises the `context.Context` to store the logger for use throughout the application.
Anywhere that a context can be found, a log can be sent to the centralised logger:

```go
import ( "github.com/anz-bank/sysl-go/log" )

log.Info(ctx, "Hello world")
```

In some instances it is desirable to attach additional information to all logs that utilise a given context:

```go
ctx = log.WithStr(ctx, "server_name", "great_pineapple")
log.Info(ctx, "Server started") // Log includes server name
...
ctx = log.WithInt(ctx, "request_id", request.id)
log.Info(ctx, "Request received") // Log includes both server name and request id
``` 

# Framework

Sysl-go respects that different teams want to use different logging solutions and that Sysl-go shouldn't prevent you from doing as such.
In order to achieve this, the logging framework within Sysl-go is designed around the `log.Logger` interface that acts as a wrapper around concrete logging implementations.

Out-of-the-box Sysl-go supports the following logging implementations:
- [Logrus](https://github.com/sirupsen/logrus)
- [Pkg](https://github.com/anz-bank/pkg/tree/master/log)
- [ZeroPkg](https://github.com/anz-bank/pkg/tree/master/logging)

By default, the [Pkg](https://github.com/anz-bank/pkg/tree/master/log) logger is used within Sysl-go.
To use a different logger or to configure the logger beyond the log level, see [Custom Configuration](#custom-configuration) below.

# External Configuration

Sysl-go applications are configured through an external configuration file.
The logger can be configured to record logs at or above a particular level:

```yaml
library:
  log:
    level: debug # one of error, info, debug
```

Another configurable value provides the ability to log the contents of requests and responses:
```yaml
library:
  log:
    logPayload: true # include payload contents in log messages
```

# Custom Configuration

By default, the [Pkg](https://github.com/anz-bank/pkg/tree/master/log) logger is used within Sysl-go.
To use a different logger or to configure the logger beyond the log level, add a custom `Logger` hook to the `Hooks` structure that builds and returns a customised logger.

The following example demonstrates the use of a customised Logrus logger:

```go
func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return example.NewServer(ctx,
		func(ctx context.Context, config AppConfig) (*example.ServiceInterface, *core.Hooks, error) {
			return &gateway.ServiceInterface{
					...
				}, &core.Hooks{
				    Logger: func() log.Logger {
                        logger := logrus.New()
                        logger.Formatter = &logrus.JSONFormatter{}
                        return anzlog.NewLogrusLogger(logger)
                    }
                }, nil
		},
	)
}
```

The returned logger will have the log level applied from the [configuration file](#configuration-file) by Sysl-go internally.

# Native Support

Some logging implementations provide their own mechanism to interact with the `context.Context`.
For example, the [Pkg](https://github.com/anz-bank/pkg/tree/master/log) logger can be used directly:

```go
import ( pkglog "github.com/anz-bank/pkg/log" )
ctx = pkglog.With("key", "value").Onto(ctx)
pkglog.Info(ctx, "Hello world")
```

The Sysl-go logging framework understands this and so the implementation works seamlessly whether you access the logger natively or through the wrapper:

```go
import ( 
    pkglog "github.com/anz-bank/pkg/log" 
    "github.com/anz-bank/sysl-go/log" 
)
ctx = pkglog.With("key", "value").Onto(ctx) // Put a key/value into the context
pkglog.Info(ctx, "Native") // Native call, includes key/value pair
log.Info(ctx, "Wrapped") // Wrapped call, also includes key/value pair
```

# Legacy Support

Sysl-go has gone through two iterations of logging. 
The first iteration enforced the use of [Logrus](https://github.com/sirupsen/logrus).
The second iteration replaced Logrus with the [Pkg](https://github.com/anz-bank/pkg/tree/master/log) logger, providing a hook mechanism to route logs back through Logrus for backwards compatibility.

With the introduction of the new logging framework, backwards compatible support of both Logrus and the Pkg logger is provided seamlessly, however upgrading to the approach described in the [Custom Configuration](#custom-configuration) is recommended:

## Logrus

```go
import ( 
    "github.com/sirupsen/logrus" 
    "github.com/anz-bank/sysl-go/log" 
)
ctx := common.LoggerToContext(ctx, logrus.New(), nil) // Put the logger in the context (the legacy approach)
example.NewServer(ctx, ...) // Initialise the server
...
logger := common.GetLoggerFromContext(ctx) // Retrieve the logger from the context
logger.Info(ctx, "Native") // Log natively
log.Info(ctx, "Wrapped") // Log using the wrapper (uses the same Logrus instance)
```

## Pkg logger

```go
import ( 
    pkglog "github.com/anz-bank/pkg/log" 
    "github.com/anz-bank/sysl-go/log" 
)
ctx = pkglog.With("key", "value").Onto(ctx) // Initialise the pkg logger (the legacy approach)
example.NewServer(ctx, ...) // Initialise the server
...
pkglog.Info(ctx, "Native")  // Log natively
log.Info(ctx, "Wrapped")  // Log using the wrapper (uses the same pkg values)
```
