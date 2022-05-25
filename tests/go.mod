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
	github.com/containerd/containerd v1.6.4
	github.com/ease-lab/vSwarm/utils/tracing/go v0.0.0-20220523084245-7d0affaa96fd
	github.com/ease-lab/vhive-xdt/proto/downXDT v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/queue-proxy/dQP v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/queue-proxy/sQP v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/sdk/golang v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/utils v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.31.0
	golang.org/x/net v0.0.0-20220524220425-1d687d428aca
	google.golang.org/grpc v1.46.2
	knative.dev/networking v0.0.0-20220524205304-22d1b933cf73
	knative.dev/pkg v0.0.0-20220524202603-19adf798efb8
	knative.dev/serving v0.31.0
)
