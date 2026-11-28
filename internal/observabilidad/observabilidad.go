package observabilidad

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Configurar las trazas del proceso y regresar
// El destino sale de OTEL_EXPORTER_OTLP_ENDPOINT
func Iniciar(ctx context.Context, servicio, instancia string) (func(context.Context) error, error) {
	// Mandar los spans al colector por OTLP sobre gRPC
	exportador, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("no se pudo crear el exportador de trazas: %w", err)
	}

	// Etiquetar cada span con el servicio y el pod que lo generó
	recurso := resource.NewSchemaless(
		attribute.String("service.name", servicio),
		attribute.String("service.instance.id", instancia),
	)

	// Juntar los spans y mandarlos en lotes
	proveedor := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exportador),
		sdktrace.WithResource(recurso),
	)

	// Registrar el proveedor y el text map propagator
	otel.SetTracerProvider(proveedor)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return proveedor.Shutdown, nil
}
