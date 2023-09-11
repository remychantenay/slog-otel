# slog-otel
[![Go Report Card](https://goreportcard.com/badge/github.com/remychantenay/slog-otel)](https://goreportcard.com/report/github.com/remychantenay/slog-otel)
[![codebeat badge](https://codebeat.co/badges/60d273d3-08e6-4f48-9c35-86ab75fc1924)](https://codebeat.co/projects/github-com-remychantenay-slog-otel-main)
[![GoDoc](https://godoc.org/github.com/remychantenay/slog-otel?status.svg)](https://godoc.org/github.com/remychantenay/slog-otel)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Go package that provides a custom handler for `log/slog` to ensures strong correlation between log records and Open-Telemetry spans by...

1. Adding span and trace IDs to the log record.
2. Adding context baggage members to the log record.
3. Adding log record as span event.
4. Adding log record attributes to the span event.
5. Setting span status based on slog record level (only if >= slog.LevelError).

## Usage
```go
// 1. Configure slog.
slog.SetDefault(slog.New(slogotel.OtelHandler{
	Next: slog.NewJSONHandler(os.Stdout, nil),
}))

// 2. Set up your logger.
logger := slog.Default()
logger = logger.With("component", "server")

// 3. Start your span.
ctx, span := tracer.Start(ctx, "operation-name")
defer span.End()

// 4. Log
logger.InfoContext(ctx, "Hello world!", "locale", "en_US")
```

#### Example
The following initial log:
```json
{
	"time": "2023-09-11T08:28:02.77215605Z",
	"level": "INFO",
	"component": "server",
	"msg": "Hello world!",
	"locale": "en_US"
}
```
... will be written as:
```json
{
	"time": "2023-09-11T08:28:02.77215605Z",
	"level": "INFO",
	"component": "server",
	"msg": "Hello world!",
	"locale": "en_US",
	"trace_id": "a9938fd7a6313e0f27f3fc87f574bff6",
	"span_id": "ed58f84d8971bf60"
}
```

and the related span will look like:
```json
 {
 	"Name": "GET /resources",
 	"SpanContext": {
 		"TraceID": "a9938fd7a6313e0f27f3fc87f574bff6",
 		"SpanID": "ed58f84d8971bf60",
 		...
 	},
 	"Parent": {
 		...
 	},
 	"SpanKind": 2,
 	"StartTime": "2023-09-11T08:28:02.761992425Z",
 	"EndTime": "2023-09-11T08:28:02.773230425Z",
 	"Attributes": [{
 			"Key": "http.method",
 			"Value": {
 				"Type": "STRING",
 				"Value": "GET"
 			}
 		},
		...
 	],
 	"Events": [{
 		"Name": "log_record",
 		"Attributes": [{
 				"Key": "msg",
 				"Value": {
 					"Type": "STRING",
 					"Value": "Hello world!"
 				}
 			},
 			{
 				"Key": "level",
 				"Value": {
 					"Type": "STRING",
 					"Value": "INFO"
 				}
 			},
 			{
 				"Key": "time",
 				"Value": {
 					"Type": "STRING",
 					"Value": "2023-09-11T08:28:02.77215605Z"
 				}
 			},
 			{
 				"Key": "locale",
 				"Value": {
 					"Type": "STRING",
 					"Value": "en_US"
 				}
 			},
 			{
 				"Key": "component",
 				"Value": {
 					"Type": "STRING",
 					"Value": "server"
 				}
 			}
 		],
 	}],
 	...
 }
```