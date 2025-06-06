package metrics

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

func NewMeter(ctx context.Context) (metric.Meter, error) {
	exp, err := stdoutmetric.New(stdoutmetric.WithWriter(os.Stderr))
	if err != nil {
		return nil, err
	}

	recources, err := resource.New(
		ctx,
		resource.WithProcessCommandArgs(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceName("isucon14_benchmarker"),
		),
	)
	if err != nil {
		return nil, err
	}

	reader := metricsdk.NewPeriodicReader(exp, metricsdk.WithInterval(3*time.Second))

	provider := metricsdk.NewMeterProvider(
		metricsdk.WithResource(recources),
		metricsdk.WithReader(reader),
	)
	otel.SetMeterProvider(provider)

	return otel.Meter("isucon14_benchmarker"), nil
}
