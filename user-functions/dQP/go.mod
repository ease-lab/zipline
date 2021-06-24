module github.com/ease-lab/xdt/user-functions/dQP

go 1.16

replace (
	github.com/ease-lab/xdt/dQP => ../../dQP
	github.com/ease-lab/xdt/proto/crossXDT => ../../proto/crossXDT
	github.com/ease-lab/xdt/proto/downXDT => ../../proto/downXDT
	github.com/ease-lab/xdt/proto/fnInvocation => ../../proto/fnInvocation
	github.com/ease-lab/xdt/transport => ../../transport
	github.com/ease-lab/xdt/utils => ../../utils
)

require (
	github.com/containerd/containerd v1.5.2
	github.com/ease-lab/vhive/utils/tracing/go v0.0.0-20210623161727-460bac97d8c0
	github.com/ease-lab/xdt/dQP v0.0.0-00010101000000-000000000000
	github.com/ease-lab/xdt/utils v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
)
