module github.com/ease-lab/xdt/queue-proxy/sQP

go 1.16

replace (
	github.com/ease-lab/xdt/proto/crossXDT => ../../proto/crossXDT
	github.com/ease-lab/xdt/proto/upXDT => ../../proto/upXDT
	github.com/ease-lab/xdt/transport => ../../transport
	github.com/ease-lab/xdt/utils => ../../utils
)

require (
	github.com/ease-lab/xdt/proto/crossXDT v0.0.0-00010101000000-000000000000
	github.com/ease-lab/xdt/proto/upXDT v0.0.0-00010101000000-000000000000
	github.com/ease-lab/xdt/transport v0.0.0-00010101000000-000000000000
	github.com/ease-lab/xdt/utils v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.20.0
	google.golang.org/grpc v1.38.0
)
