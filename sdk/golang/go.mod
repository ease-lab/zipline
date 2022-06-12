module github.com/ease-lab/vhive-xdt/sdk/golang

go 1.17

replace (
	github.com/ease-lab/vhive-xdt/proto/crossXDT => ./../../proto/crossXDT
	github.com/ease-lab/vhive-xdt/proto/downXDT => ./../../proto/downXDT
	github.com/ease-lab/vhive-xdt/proto/upXDT => ./../../proto/upXDT
	github.com/ease-lab/vhive-xdt/utils => ./../../utils
)

require (
	github.com/ease-lab/vSwarm/utils/tracing/go v0.0.0-20220609140039-b4da20ea6b36
	github.com/ease-lab/vhive-xdt/proto/crossXDT v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/proto/downXDT v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/proto/upXDT v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/utils v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.31.0
	go.uber.org/atomic v1.9.0
	google.golang.org/grpc v1.47.0
)
