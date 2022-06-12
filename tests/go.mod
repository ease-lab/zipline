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
	github.com/prometheus/statsd_exporter => github.com/prometheus/statsd_exporter v0.22.5
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp => go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.32.0
)

require (
	XDTgRPC_stream/plotter v0.0.0-00010101000000-000000000000
	cloud.google.com/go/container v1.2.0 // indirect
	cloud.google.com/go/monitoring v1.5.0 // indirect
	cloud.google.com/go/trace v1.2.0 // indirect
	github.com/containerd/containerd v1.6.2
	github.com/ease-lab/vSwarm/utils/tracing/go v0.0.0-20220609140039-b4da20ea6b36
	github.com/ease-lab/vhive-xdt/proto/downXDT v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/queue-proxy/dQP v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/queue-proxy/sQP v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/sdk/golang v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/utils v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.20.0
	golang.org/x/net v0.0.0-20220425223048-2871e0cb64e4
	google.golang.org/grpc v1.47.0
	knative.dev/networking v0.0.0-20210512050647-ace2d3306f0b
	knative.dev/pkg v0.0.0-20210510175900-4564797bf3b7
	knative.dev/serving v0.23.1
)
