module github.com/ease-lab/vhive-xdt/user-functions/fx

go 1.16

replace (
	github.com/ease-lab/vhive-xdt/proto/downXDT => ../../proto/downXDT
	github.com/ease-lab/vhive-xdt/proto/upXDT => ../../proto/upXDT
	github.com/ease-lab/vhive-xdt/sdk/go_sdk => ../../sdk/go_sdk
	github.com/ease-lab/vhive-xdt/utils => ../../utils
)

require (
	github.com/containerd/containerd v1.5.2
	github.com/ease-lab/vhive-xdt/sdk/go_sdk v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/utils v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive/examples/protobuf/helloworld v0.0.0-20210624210547-e0cd5d053491
	github.com/sirupsen/logrus v1.8.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.20.0
	google.golang.org/grpc v1.38.0
)
