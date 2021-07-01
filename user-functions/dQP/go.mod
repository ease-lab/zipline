module github.com/ease-lab/vhive-xdt/user-functions/dQP

go 1.16

replace (
	github.com/ease-lab/vhive-xdt/proto/crossXDT => ../../proto/crossXDT
	github.com/ease-lab/vhive-xdt/proto/downXDT => ../../proto/downXDT
	github.com/ease-lab/vhive-xdt/queue-proxy/dQP => ../../queue-proxy/dQP
	github.com/ease-lab/vhive-xdt/transport => ../../transport
	github.com/ease-lab/vhive-xdt/utils => ../../utils
)

require (
	github.com/containerd/containerd v1.5.2
	github.com/ease-lab/vhive-xdt/queue-proxy/dQP v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive-xdt/utils v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive/utils/tracing/go v0.0.0-20210623161727-460bac97d8c0
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e
	knative.dev/pkg v0.0.0-20210625194144-4cdacd04734a
	knative.dev/serving v0.23.1
)
