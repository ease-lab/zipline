module dqp

go 1.16

replace (
	XDTprototype/proto/crossXDT => ../proto/crossXDT
	XDTprototype/proto/downXDT => ../proto/downXDT
	XDTprototype/proto/fnInvocation => ../proto/fnInvocation
	XDTprototype/transport => ../transport
	XDTprototype/utils => ../utils
)

require (
	XDTprototype/proto/crossXDT v0.0.0-00010101000000-000000000000
	XDTprototype/proto/downXDT v0.0.0-00010101000000-000000000000
	XDTprototype/proto/fnInvocation v0.0.0-00010101000000-000000000000
	XDTprototype/transport v0.0.0-00010101000000-000000000000
	XDTprototype/utils v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.20.0
	google.golang.org/grpc v1.38.0
)
