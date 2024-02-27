package slogotel

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	// TraceIDKey is the key used by the Otel handler
	// to inject the trace ID in the log record.
	TraceIDKey = "trace_id"
	// SpanIDKey is the key used by the Otel handler
	// to inject the span ID in the log record.
	SpanIDKey = "span_id"
	// SpanEventKey is the key used by the Otel handler
	// to inject the log record in the recording span, as a span event.
	SpanEventKey = "log_record"
)

// OtelHandler is an implementation of slog's Handler interface.
// Its role is to ensure correlation between logs and OTel spans
// by:
//
// 1. Adding otel span and trace IDs to the log record.
// 2. Adding otel context baggage members to the log record.
// 3. Setting slog record as otel span event.
// 4. Adding slog record attributes to the otel span event.
// 5. Setting span status based on slog record level (only if >= slog.LevelError).
type OtelHandler struct {
	// Next represents the next handler in the chain.
	Next slog.Handler
	// NoBaggage determines whether to add context baggage members to the log record.
	NoBaggage bool
}

// HandlerFn defines the handler used by slog.Handler as return value.
type HandlerFn func(slog.Handler) slog.Handler

// NewOtelHandler creates and returns a new OtelHandler to use with log/slog.
func NewOtelHandler() HandlerFn {
	return func(next slog.Handler) slog.Handler {
		return &OtelHandler{
			Next: next,
		}
	}
}

// Handle handles the provided log record and adds correlation between a slog record and an Open-Telemetry span.
func (h OtelHandler) Handle(ctx context.Context, record slog.Record) error {
	if ctx == nil {
		return h.Next.Handle(ctx, record)
	}

	if !h.NoBaggage {
		// Adding context baggage members to log record.
		b := baggage.FromContext(ctx)
		for _, m := range b.Members() {
			record.AddAttrs(slog.String(m.Key(), m.Value()))
		}
	}

	span := trace.SpanFromContext(ctx)
	if span == nil || !span.IsRecording() {
		return h.Next.Handle(ctx, record)
	}

	// Adding log info to span event.
	eventAttrs := make([]attribute.KeyValue, 0, record.NumAttrs())
	eventAttrs = append(eventAttrs, attribute.String(slog.MessageKey, record.Message))
	eventAttrs = append(eventAttrs, attribute.String(slog.LevelKey, record.Level.String()))
	eventAttrs = append(eventAttrs, attribute.String(slog.TimeKey, record.Time.Format(time.RFC3339Nano)))

	record.Attrs(func(attr slog.Attr) bool {
		otelAttr := h.slogAttrToOtelAttr(attr)
		if otelAttr.Valid() {
			eventAttrs = append(eventAttrs, otelAttr)
		}

		return true
	})

	span.AddEvent(SpanEventKey, trace.WithAttributes(eventAttrs...))

	// Adding span info to log record.
	spanContext := span.SpanContext()
	if spanContext.HasTraceID() {
		traceID := spanContext.TraceID().String()
		record.AddAttrs(slog.String(TraceIDKey, traceID))
	}

	if spanContext.HasSpanID() {
		spanID := spanContext.SpanID().String()
		record.AddAttrs(slog.String(SpanIDKey, spanID))
	}

	// Setting span status if the log is an error.
	// Purposely leaving as codes.Unset (default) otherwise.
	if record.Level >= slog.LevelError {
		span.SetStatus(codes.Error, record.Message)
	}

	return h.Next.Handle(ctx, record)
}

// WithAttrs returns a new Otel whose attributes consists of handler's attributes followed by attrs.
func (h OtelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return OtelHandler{
		Next: h.Next.WithAttrs(attrs),
		NoBaggage: h.NoBaggage,
	}
}

// WithGroup returns a new Otel with a group, provided the group's name.
func (h OtelHandler) WithGroup(name string) slog.Handler {
	return OtelHandler{
		Next: h.Next.WithGroup(name),
		NoBaggage: h.NoBaggage,
	}
}

// Enabled reports whether the logger emits log records at the given context and level.
// Note: We handover the decision down to the next handler.
func (h OtelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Next.Enabled(ctx, level)
}

// slogAttrToOtelAttr converts a slog attribute to an OTel one.
// Note: returns an empty attribute if the provided slog attribute is empty.
func (h OtelHandler) slogAttrToOtelAttr(attr slog.Attr, groupKeys ...string) attribute.KeyValue {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return attribute.KeyValue{}
	}

	key := func(k string, prefixes ...string) string {
		for _, prefix := range prefixes {
			k = fmt.Sprintf("%s.%s", prefix, k)
		}

		return k
	}(attr.Key, groupKeys...)

	value := attr.Value.Resolve()

	switch attr.Value.Kind() {
	case slog.KindBool:
		return attribute.Bool(key, value.Bool())
	case slog.KindFloat64:
		return attribute.Float64(key, value.Float64())
	case slog.KindInt64:
		return attribute.Int64(key, value.Int64())
	case slog.KindString:
		return attribute.String(key, value.String())
	case slog.KindTime:
		return attribute.String(key, value.Time().Format(time.RFC3339Nano))
	case slog.KindGroup:
		groupAttrs := value.Group()
		if len(groupAttrs) == 0 {
			return attribute.KeyValue{}
		}

		for _, groupAttr := range groupAttrs {
			return h.slogAttrToOtelAttr(groupAttr, append(groupKeys, key)...)
		}
	case slog.KindAny:
		switch v := attr.Value.Any().(type) {
		case []string:
			return attribute.StringSlice(key, v)
		case []int:
			return attribute.IntSlice(key, v)
		case []int64:
			return attribute.Int64Slice(key, v)
		case []float64:
			return attribute.Float64Slice(key, v)
		case []bool:
			return attribute.BoolSlice(key, v)
		default:
			return attribute.KeyValue{}
		}
	default:
		return attribute.KeyValue{}
	}

	return attribute.KeyValue{}
}
