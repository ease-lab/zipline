module github.com/ease-lab/vhive-xdt/sdk/golang

go 1.18

replace (
	github.com/ease-lab/vhive-xdt/proto/crossXDT => ./../../proto/crossXDT
	github.com/ease-lab/vhive-xdt/proto/downXDT => ./../../proto/downXDT
	github.com/ease-lab/vhive-xdt/proto/upXDT => ./../../proto/upXDT
	github.com/ease-lab/vhive-xdt/utils => ./../../utils
)

require (
	capnproto.org/go/capnp/v3 v3.0.0-alpha.18
	github.com/ease-lab/vSwarm/utils/tracing/go v0.0.0-20220609140039-b4da20ea6b36
	github.com/ease-lab/vhive-xdt/proto/crossXDT v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/proto/downXDT v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/utils v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.46.0
	go.uber.org/atomic v1.9.0
	google.golang.org/grpc v1.59.0
)

require (
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/containerd/containerd v1.6.38 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/openzipkin/zipkin-go v0.4.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.opentelemetry.io/otel v1.21.0 // indirect
	go.opentelemetry.io/otel/exporters/zipkin v1.6.3 // indirect
	go.opentelemetry.io/otel/metric v1.21.0 // indirect
	go.opentelemetry.io/otel/sdk v1.21.0 // indirect
	go.opentelemetry.io/otel/trace v1.21.0 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231002182017-d307bd883b97 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)
