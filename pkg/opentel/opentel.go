package opentel

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
)

const name = "github.com/TrungNNg/og-tag/opentel"

var (
	Tracer        = otel.Tracer(name)
	shutdownFuncs []func(context.Context) error
)

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// setup propagator, trace exporter and trace provider
func SetupOTelSDK() {
	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace exporter and trace provider.
	tracerProvider, err := newTracerProvider()
	if err != nil {
		slog.Error("could not setup otel trace exporter, trace provider")
		os.Exit(1)
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)
}

// Shutdown call shutdowns functions in shutdownFuncs for cleanup
func Shutdown(ctx context.Context) error {
	var err error
	for _, fn := range shutdownFuncs {
		err = errors.Join(err, fn(ctx))
	}
	shutdownFuncs = nil
	return err
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTracerProvider() (*trace.TracerProvider, error) {
	traceExporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
	)
	return tracerProvider, nil
}
