package slogotel_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"go.opentelemetry.io/otel/baggage"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	slogotel "github.com/remychantenay/slog-otel"
)

var (
	loggerWithOtel *slog.Logger
	logger         *slog.Logger
	traceExporter  *tracetest.InMemoryExporter
	tracer         trace.Tracer
)

func init() {
	logger = setUploggerWithoutOtelHandler()
	loggerWithOtel = setUploggerWithOtelHandler()
	traceExporter, tracer = setUpBenchmarkTracer()
}

func setUploggerWithOtelHandler() *slog.Logger {
	var buffer bytes.Buffer
	return slog.New(slogotel.OtelHandler{
		Next: slog.NewJSONHandler(&buffer, nil),
	})
}

func setUploggerWithoutOtelHandler() *slog.Logger {
	var buffer bytes.Buffer
	return slog.New(slog.NewJSONHandler(&buffer, nil))
}

func setUpBenchmarkTracer() (*tracetest.InMemoryExporter, trace.Tracer) {
	exporter := tracetest.NewInMemoryExporter()
	traceProvider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	tracer := traceProvider.Tracer("benchmark-tracer")

	return exporter, tracer
}

func BenchmarkJSONHandler_SimpleLog(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	logger.InfoContext(ctx, "Hello world!")
}

func BenchmarkOtelHandler_SimpleLog(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			loggerWithOtel.InfoContext(ctx, "Hello world!")
		}
	})
}

func BenchmarkJSONHandler_WithSpan(b *testing.B) {
	traceExporter.Reset()

	m1, _ := baggage.NewMember("key1b", "value1b")
	m2, _ := baggage.NewMember("key2b", "value2b")
	bag, _ := baggage.New(m1, m2)
	ctx := baggage.ContextWithBaggage(context.Background(), bag)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx, span := tracer.Start(ctx, "operation-name")
			defer span.End()

			group1 := slog.Group("group_1", "key_1", "value_1")
			group2 := slog.Group("group_2", "key_2", "value_2")
			logger.InfoContext(ctx, "Hello world!",
				"key1", "value1",
				"key2", 42.0,
				"key3", 42,
				"key4", true,
				"key5", time.Now(),
				group1,
				group2,
			)
		}
	})
}

func BenchmarkOtelHandler_WithSpan(b *testing.B) {
	traceExporter.Reset()

	m1, _ := baggage.NewMember("key1b", "value1b")
	m2, _ := baggage.NewMember("key2b", "value2b")
	bag, _ := baggage.New(m1, m2)
	ctx := baggage.ContextWithBaggage(context.Background(), bag)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx, span := tracer.Start(ctx, "operation-name")
			defer span.End()

			group1 := slog.Group("group_1", "key_1", "value_1")
			group2 := slog.Group("group_2", "key_2", "value_2")
			loggerWithOtel.InfoContext(ctx, "Hello world!",
				"key1", "value1",
				"key2", 42.0,
				"key3", 42,
				"key4", true,
				"key5", time.Now(),
				group1,
				group2,
			)
		}
	})
}
