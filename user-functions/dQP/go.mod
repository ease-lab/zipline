module github.com/ease-lab/vhive-xdt/user-functions/dQP

go 1.18

replace (
	github.com/ease-lab/vhive-xdt/proto/crossXDT => ../../proto/crossXDT
	github.com/ease-lab/vhive-xdt/proto/downXDT => ../../proto/downXDT
	github.com/ease-lab/vhive-xdt/queue-proxy/dQP => ../../queue-proxy/dQP
	github.com/ease-lab/vhive-xdt/transport => ../../transport
	github.com/ease-lab/vhive-xdt/utils => ../../utils
)

require (
	github.com/containerd/containerd v1.6.4
	github.com/ease-lab/vSwarm/utils/tracing/go v0.0.0-20220523084245-7d0affaa96fd
	github.com/ease-lab/vhive-xdt/queue-proxy/dQP v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/utils v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.32.0
	golang.org/x/net v0.0.0-20220524220425-1d687d428aca
	google.golang.org/grpc v1.46.2
	knative.dev/pkg v0.0.0-20220525153005-18f69958870f
)

require golang.org/x/tools v0.1.9 // indirect
