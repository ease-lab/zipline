module dqp

go 1.15

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/upXDT => ../proto/upXDT

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/crossXDT => ../proto/crossXDT

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/fnInvocation => ../proto/fnInvocation

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/downXDT => ../proto/downXDT

replace github.com/ease-lab/vhive_stealth/examples/prototype/sdk => ../sdk

replace github.com/ease-lab/vhive_stealth/examples/prototype/dqp => ../dQP

replace github.com/ease-lab/vhive_stealth/examples/prototype/sqp => ../sQP

replace github.com/ease-lab/vhive_stealth/examples/gRPC_stream/plotter => ../plotter

require (
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/crossXDT v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/downXDT v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/fnInvocation v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/sdk v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.20.0
	google.golang.org/grpc v1.37.0
)
