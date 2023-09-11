/*
Package slogotel provides a custom handler for `log/slog` to ensures strong correlation between log records and Open-Telemetry spans.

# Usage

	slog.SetDefault(slog.New(slogotel.OtelHandler{
		Next: slog.NewJSONHandler(os.Stdout, nil),
	}))

	logger := slog.Default()
	logger = logger.With("component", "server")

	ctx, span := tracer.Start(ctx, "operation-name")
	defer span.End()

	logger.InfoContext(ctx, "Hello world!", "locale", "en_US")
*/
package slogotel
