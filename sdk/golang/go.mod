module github.com/ease-lab/vhive-xdt/sdk/golang

go 1.18

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

require (
	github.com/containerd/containerd v1.6.2 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/openzipkin/zipkin-go v0.4.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.opentelemetry.io/otel v1.6.3 // indirect
	go.opentelemetry.io/otel/exporters/zipkin v1.6.3 // indirect
	go.opentelemetry.io/otel/sdk v1.6.3 // indirect
	go.opentelemetry.io/otel/trace v1.6.3 // indirect
	golang.org/x/net v0.0.0-20220425223048-2871e0cb64e4 // indirect
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220414192740-2d67ff6cf2b4 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
)
