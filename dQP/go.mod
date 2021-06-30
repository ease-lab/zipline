module github.com/ease-lab/xdt/dQP

go 1.16

replace (
	github.com/ease-lab/xdt/proto/crossXDT => ../proto/crossXDT
	github.com/ease-lab/xdt/proto/downXDT => ../proto/downXDT
	github.com/ease-lab/xdt/transport => ../transport
	github.com/ease-lab/xdt/utils => ../utils
)

require (
	cloud.google.com/go v0.54.0 // indirect
	github.com/ease-lab/xdt/proto/crossXDT v0.0.0-00010101000000-000000000000
	github.com/ease-lab/xdt/proto/downXDT v0.0.0-00010101000000-000000000000
	github.com/ease-lab/xdt/transport v0.0.0-00010101000000-000000000000
	github.com/ease-lab/xdt/utils v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.20.0
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110 // indirect
	golang.org/x/sys v0.0.0-20210324051608-47abb6519492 // indirect
	golang.org/x/tools v0.0.0-20210106214847-113979e3529a // indirect
	google.golang.org/genproto v0.0.0-20201110150050-8816d57aaa9a // indirect
	google.golang.org/grpc v1.38.0
)
