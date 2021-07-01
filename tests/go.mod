module github.com/ease-lab/xdt/integration_tests

go 1.16

replace (
	XDTgRPC_stream/plotter => ../plotter
	github.com/ease-lab/xdt/dQP => ../dQP
	github.com/ease-lab/xdt/proto/crossXDT => ../proto/crossXDT
	github.com/ease-lab/xdt/proto/downXDT => ../proto/downXDT
	github.com/ease-lab/xdt/proto/upXDT => ../proto/upXDT
	github.com/ease-lab/xdt/sQP => ../sQP
	github.com/ease-lab/xdt/sdk/go_sdk => ../sdk/go_sdk
	github.com/ease-lab/xdt/transport => ../transport
	github.com/ease-lab/xdt/utils => ../utils
)

require (
	XDTgRPC_stream/plotter v0.0.0-00010101000000-000000000000
	github.com/containerd/containerd v1.5.2
	github.com/ease-lab/vhive/utils/tracing/go v0.0.0-20210624210547-e0cd5d053491
	github.com/ease-lab/xdt/dQP v0.0.0-00010101000000-000000000000
	github.com/ease-lab/xdt/sQP v0.0.0-00010101000000-000000000000
	github.com/ease-lab/xdt/sdk/go_sdk v0.0.0-00010101000000-000000000000
	github.com/ease-lab/xdt/utils v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/net v0.0.0-20210415231046-e915ea6b2b7d
	knative.dev/pkg v0.0.0-20210510175900-4564797bf3b7
	knative.dev/serving v0.23.1
)
