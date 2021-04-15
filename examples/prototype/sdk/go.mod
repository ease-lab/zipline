module sdk

go 1.15

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/SrcFnToQPProto => ../proto/SrcFnToQP

replace github.com/ease-lab/vhive_stealth/examples/gRPC_stream/plotter => ../plotter

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/CrossQPProto => ../proto/CrossQP

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/FnInvocationProto => ../proto/FnInvocation

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/QPToDstFnProto => ../proto/QPToDstFn

replace github.com/ease-lab/vhive_stealth/examples/prototype/sqp => ../sQP

replace github.com/ease-lab/vhive_stealth/examples/prototype/dqp => ../dQP

replace github.com/ease-lab/vhive_stealth/examples/prototype/sdk => ./

require (
	github.com/ease-lab/vhive_stealth/examples/gRPC_stream/plotter v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/dqp v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/FnInvocationProto v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/SrcFnToQPProto v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/sdk v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/sqp v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	google.golang.org/grpc v1.36.0
)
