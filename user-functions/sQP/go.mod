module github.com/ease-lab/xdt/user-functions/sQP

go 1.16

replace (
	github.com/ease-lab/xdt/proto/crossXDT => ../../proto/crossXDT
	github.com/ease-lab/xdt/proto/upXDT => ../../proto/upXDT
	github.com/ease-lab/xdt/sQP => ../../sQP
	github.com/ease-lab/xdt/tracing => ../../tracing
	github.com/ease-lab/xdt/transport => ../../transport
	github.com/ease-lab/xdt/utils => ../../utils
)

require (
	github.com/ease-lab/xdt/sQP v0.0.0-00010101000000-000000000000
	github.com/ease-lab/xdt/tracing v0.0.0-00010101000000-000000000000
	github.com/ease-lab/xdt/utils v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
)
