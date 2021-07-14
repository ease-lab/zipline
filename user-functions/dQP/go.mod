module github.com/ease-lab/vhive-xdt/user-functions/dQP

go 1.16

replace (
	github.com/ease-lab/vhive-xdt/proto/crossXDT => ../../proto/crossXDT
	github.com/ease-lab/vhive-xdt/proto/downXDT => ../../proto/downXDT
	github.com/ease-lab/vhive-xdt/queue-proxy/dQP => ../../queue-proxy/dQP
	github.com/ease-lab/vhive-xdt/transport => ../../transport
	github.com/ease-lab/vhive-xdt/utils => ../../utils
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp => go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.20.0
)

require (
	github.com/containerd/containerd v1.5.2
	github.com/ease-lab/vhive-xdt/queue-proxy/dQP v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/utils v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive/utils/tracing/go v0.0.0-20210623161727-460bac97d8c0
	github.com/sirupsen/logrus v1.8.1
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e
	google.golang.org/grpc v1.38.0
	knative.dev/pkg v0.0.0-20210625194144-4cdacd04734a
	knative.dev/serving v0.24.0
)
