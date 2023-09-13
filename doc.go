/*
Package slogotel provides a custom handler for `log/slog` to ensures strong correlation between log records and Open-Telemetry spans.

# Usage

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
	tracer := sdktrace.NewTracerProvider().Tracer("server")
	ctx, span := tracer.Start(ctx, "operation-name")
	defer span.End()

	// 5. Log.
	logger.InfoContext(ctx, "Hello world!", "locale", "en_US")
*/
package slogotel
