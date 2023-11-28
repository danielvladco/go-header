//go:build !notracing

package p2p

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("header/server")
