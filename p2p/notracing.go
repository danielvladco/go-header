//go:build notracing

package p2p

import "go.opentelemetry.io/otel/trace/noop"

var tracer = noop.NewTracerProvider().Tracer("header/server")
