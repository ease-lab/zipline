module main

go 1.15

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/SrcFnToQPProto => ../proto/SrcFnToQP

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/CrossQPProto => ../proto/CrossQP

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/FnInvocationProto => ../proto/FnInvocation

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/QPToDstFnProto => ../proto/QPToDstFn

require (
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/CrossQPProto v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/FnInvocationProto v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/QPToDstFnProto v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/SrcFnToQPProto v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.36.0
)
