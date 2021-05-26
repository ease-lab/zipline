module transport

go 1.16

replace (
	XDTgRPC_stream/plotter => ../plotter
	XDTprototype/dqp => ../dQP
	XDTprototype/proto/crossXDT => ../proto/crossXDT
	XDTprototype/proto/downXDT => ../proto/downXDT
	XDTprototype/proto/fnInvocation => ../proto/fnInvocation
	XDTprototype/proto/upXDT => ../proto/upXDT
	XDTprototype/sqp => ../sQP
	XDTprototype/transport => ./
)

require (
	github.com/sirupsen/logrus v1.8.1
	go.opentelemetry.io/otel v0.20.0
	go.opentelemetry.io/otel/exporters/trace/zipkin v0.20.0
	go.opentelemetry.io/otel/sdk v0.20.0
)
