module sdk

go 1.16

replace (
	XDTgRPC_stream/plotter => ../plotter
	XDTprototype/dqp => ../dQP
	XDTprototype/proto/crossXDT => ../proto/crossXDT
	XDTprototype/proto/downXDT => ../proto/downXDT
	XDTprototype/proto/fnInvocation => ../proto/fnInvocation
	XDTprototype/proto/upXDT => ../proto/upXDT
	XDTprototype/sdk => ./
	XDTprototype/sqp => ../sQP
)

require (
	XDTgRPC_stream/plotter v0.0.0-00010101000000-000000000000
	XDTprototype/dqp v0.0.0-00010101000000-000000000000
	XDTprototype/proto/downXDT v0.0.0-00010101000000-000000000000
	XDTprototype/proto/fnInvocation v0.0.0-00010101000000-000000000000
	XDTprototype/proto/upXDT v0.0.0-00010101000000-000000000000
	XDTprototype/sdk v0.0.0-00010101000000-000000000000
	XDTprototype/sqp v0.0.0-00010101000000-000000000000
	github.com/hjson/hjson-go v3.1.0+incompatible
	github.com/sirupsen/logrus v1.8.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.20.0
	go.opentelemetry.io/otel v0.20.0
	go.opentelemetry.io/otel/exporters/trace/zipkin v0.20.0
	go.opentelemetry.io/otel/sdk v0.20.0
	google.golang.org/grpc v1.37.0
	muzzammil.xyz/jsonc v0.0.0-20201229145248-615b0916ca38 // indirect
)
