module github.com/ease-lab/vhive-xdt/user-functions/fx

go 1.18

replace (
	github.com/ease-lab/vhive-xdt/proto/crossXDT => ../../proto/crossXDT
	github.com/ease-lab/vhive-xdt/proto/downXDT => ../../proto/downXDT
	github.com/ease-lab/vhive-xdt/proto/upXDT => ../../proto/upXDT
	github.com/ease-lab/vhive-xdt/sdk/golang => ../../sdk/golang
	github.com/ease-lab/vhive-xdt/utils => ../../utils
)

require (
	github.com/containerd/containerd v1.6.4
	github.com/ease-lab/vSwarm/utils/tracing/go v0.0.0-20220523084245-7d0affaa96fd
	github.com/ease-lab/vhive-xdt/sdk/golang v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/utils v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive/examples/protobuf/helloworld v0.0.0-20220523121411-8e88e386ae97
	github.com/sirupsen/logrus v1.8.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.32.0
	google.golang.org/grpc v1.46.2
)
