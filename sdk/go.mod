module github.com/ease-lab/xdt/sdk

go 1.16

replace (
	XDTgRPC_stream/plotter => ../plotter
	github.com/ease-lab/xdt/dQP => ../dQP
	github.com/ease-lab/xdt/proto/crossXDT => ../proto/crossXDT
	github.com/ease-lab/xdt/proto/downXDT => ../proto/downXDT
	github.com/ease-lab/xdt/proto/fnInvocation => ../proto/fnInvocation
	github.com/ease-lab/xdt/proto/upXDT => ../proto/upXDT
	github.com/ease-lab/xdt/sQP => ../sQP
	github.com/ease-lab/xdt/tracing => ../tracing
	github.com/ease-lab/xdt/transport => ../transport
	github.com/ease-lab/xdt/utils => ../utils
)

require (
	XDTgRPC_stream/plotter v0.0.0-00010101000000-000000000000
	github.com/ease-lab/xdt/dQP v0.0.0-00010101000000-000000000000
	github.com/ease-lab/xdt/proto/downXDT v0.0.0-00010101000000-000000000000
	github.com/ease-lab/xdt/proto/fnInvocation v0.0.0-00010101000000-000000000000
	github.com/ease-lab/xdt/proto/upXDT v0.0.0-00010101000000-000000000000
	github.com/ease-lab/xdt/sQP v0.0.0-00010101000000-000000000000
	github.com/ease-lab/xdt/tracing v0.0.0-00010101000000-000000000000
	github.com/ease-lab/xdt/utils v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.20.0
	google.golang.org/grpc v1.38.0
)
