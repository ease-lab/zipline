module github.com/ease-lab/vhive-xdt/user-functions/fx

go 1.23.0

replace (
	github.com/ease-lab/vhive-xdt/proto/crossXDT => ../../proto/crossXDT
	github.com/ease-lab/vhive-xdt/proto/downXDT => ../../proto/downXDT
	github.com/ease-lab/vhive-xdt/proto/upXDT => ../../proto/upXDT
	github.com/ease-lab/vhive-xdt/sdk/golang => ../../sdk/golang
	github.com/ease-lab/vhive-xdt/utils => ../../utils
)

require (
	github.com/containerd/containerd v1.7.29
	github.com/ease-lab/vSwarm/utils/tracing/go v0.0.0-20220609140039-b4da20ea6b36
	github.com/ease-lab/vhive-xdt/sdk/golang v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/utils v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive/examples/protobuf/helloworld v0.0.0-20210624210547-e0cd5d053491
	github.com/sirupsen/logrus v1.9.3
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.46.1
	google.golang.org/grpc v1.59.0
)

require (
	capnproto.org/go/capnp/v3 v3.0.0-alpha.18 // indirect
	cloud.google.com/go/compute v1.23.3 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/ease-lab/vhive-xdt/proto/crossXDT v0.0.0-00010101000000-000000000000 // indirect
	github.com/ease-lab/vhive-xdt/proto/downXDT v0.0.0-00010101000000-000000000000 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
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
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240401170217-c3f982113cda // indirect
	google.golang.org/protobuf v1.35.2 // indirect
)
