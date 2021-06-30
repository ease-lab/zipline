module github.com/ease-lab/xdt/user-functions/gx

go 1.16

replace (
	github.com/ease-lab/xdt/proto/downXDT => ../../proto/downXDT
	github.com/ease-lab/xdt/proto/upXDT => ../../proto/upXDT
	github.com/ease-lab/xdt/sdk => ../../sdk
	github.com/ease-lab/xdt/utils => ../../utils
)

require (
	github.com/containerd/containerd v1.5.2
	github.com/ease-lab/xdt/sdk v0.0.0-00010101000000-000000000000
	github.com/ease-lab/xdt/utils v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
)
