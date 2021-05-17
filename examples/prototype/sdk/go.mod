module sdk

go 1.16

replace (
	github.com/ease-lab/vhive_stealth/examples/gRPC_stream/plotter => ../plotter
	github.com/ease-lab/vhive_stealth/examples/prototype/dqp => ../dQP
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/crossXDT => ../proto/crossXDT
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/downXDT => ../proto/downXDT
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/fnInvocation => ../proto/fnInvocation
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/upXDT => ../proto/upXDT
	github.com/ease-lab/vhive_stealth/examples/prototype/sdk => ./
	github.com/ease-lab/vhive_stealth/examples/prototype/sqp => ../sQP
)

require (
	github.com/ease-lab/vhive_stealth/examples/gRPC_stream/plotter v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/dqp v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/downXDT v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/fnInvocation v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/upXDT v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/sdk v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/sqp v0.0.0-00010101000000-000000000000
	github.com/hjson/hjson-go v3.1.0+incompatible
	github.com/sirupsen/logrus v1.8.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.20.0
	go.opentelemetry.io/otel v0.20.0
	go.opentelemetry.io/otel/exporters/trace/zipkin v0.20.0
	go.opentelemetry.io/otel/sdk v0.20.0
	google.golang.org/grpc v1.37.0
	muzzammil.xyz/jsonc v0.0.0-20201229145248-615b0916ca38 // indirect
)
