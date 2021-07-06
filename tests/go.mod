module github.com/ease-lab/vhive-xdt/integration_tests

go 1.16

replace (
	XDTgRPC_stream/plotter => ../plotter
	github.com/ease-lab/vhive-xdt/proto/crossXDT => ../proto/crossXDT
	github.com/ease-lab/vhive-xdt/proto/downXDT => ../proto/downXDT
	github.com/ease-lab/vhive-xdt/proto/upXDT => ../proto/upXDT
	github.com/ease-lab/vhive-xdt/queue-proxy/dQP => ../queue-proxy/dQP
	github.com/ease-lab/vhive-xdt/queue-proxy/sQP => ../queue-proxy/sQP
	github.com/ease-lab/vhive-xdt/sdk/golang => ../sdk/golang
	github.com/ease-lab/vhive-xdt/transport => ../transport
	github.com/ease-lab/vhive-xdt/utils => ../utils
)

require (
	XDTgRPC_stream/plotter v0.0.0-00010101000000-000000000000
	github.com/containerd/containerd v1.5.2
	github.com/ease-lab/vhive-xdt/queue-proxy/dQP v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/queue-proxy/sQP v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/sdk/golang v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/utils v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive/utils/tracing/go v0.0.0-20210624210547-e0cd5d053491
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/net v0.0.0-20210415231046-e915ea6b2b7d
	google.golang.org/grpc v1.38.0
	knative.dev/networking v0.0.0-20210512050647-ace2d3306f0b
	knative.dev/pkg v0.0.0-20210510175900-4564797bf3b7
	knative.dev/serving v0.23.1
)
