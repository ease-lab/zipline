module main

go 1.15

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/SrcFnToQPProto => ../proto/SrcFnToQP

replace github.com/ease-lab/vhive_stealth/examples/gRPC_stream/plotter => ../plotter

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/CrossQPProto => ../proto/CrossQP

require (
	github.com/ease-lab/vhive_stealth/examples/gRPC_stream/plotter v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/CrossQPProto v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/SrcFnToQPProto v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	gonum.org/v1/gonum v0.9.1 // indirect
	google.golang.org/grpc v1.36.0
)
