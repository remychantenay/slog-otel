# slog-otel
[![Go Report Card](https://goreportcard.com/badge/github.com/remychantenay/slog-otel)](https://goreportcard.com/report/github.com/remychantenay/slog-otel)
[![codebeat badge](https://codebeat.co/badges/33ebce8f-9681-4c9c-8c43-f9ab4f197d9e)](https://codebeat.co/projects/github-com-remychantenay-slog-otel-main)
[![GoDoc](https://godoc.org/github.com/remychantenay/slog-otel?status.svg)](https://godoc.org/github.com/remychantenay/slog-otel)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Go package that provides an implementation of `log/slog`'s [Handler interface](https://pkg.go.dev/log/slog#Handler) that ensures a strong correlation between log records and [Open-Telemetry spans](https://opentelemetry.io/docs/concepts/signals/traces/#spans) by...

1. Adding [span and trace IDs](https://opentelemetry.io/docs/concepts/signals/traces/#span-context) to the log record.
2. Adding context [baggage](https://opentelemetry.io/docs/concepts/signals/baggage/) members to the log record (can be disabled).
3. Adding log record as [span event](https://opentelemetry.io/docs/concepts/signals/traces/#span-events) (can be disabled).
4. Adding log record attributes to the span event (can be disabled).
5. Setting [span status](https://opentelemetry.io/docs/concepts/signals/traces/#span-status) based on slog record level (only if >= slog.LevelError).

## Usage
```go
import (
    "context"
    "log/slog"

    "go.opentelemetry.io/otel/baggage"
    "go.opentelemetry.io/otel/trace"

    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    slogotel "github.com/remychantenay/slog-otel"
)

// 1. Configure slog.
slog.SetDefault(slog.New(slogotel.OtelHandler{
    Next: slog.NewJSONHandler(os.Stdout, nil),
}))

// 2. Set up your logger.
logger := slog.Default()
logger = logger.With("component", "server")

// 3. (Optional) Add baggage to your context.
m1, _ := baggage.NewMember("key_1", "value_1")
m2, _ := baggage.NewMember("key_2", "value_2")
bag, _ := baggage.New(m1, m2)
ctx := baggage.ContextWithBaggage(context.Background(), bag)

// 4. Start your span.
ctx, span := tracer.Start(ctx, "operation-name")
defer span.End()

// 5. Log.
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
    "span_id": "ed58f84d8971bf60",
    "key_1": "value_1",
    "key_2": "value_2"
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

## License
Apache License Version 2.0
