module fx

go 1.16

replace (
	github.com/ease-lab/vhive_stealth/examples/gRPC_stream/plotter => ../../plotter
	github.com/ease-lab/vhive_stealth/examples/prototype/dqp => ../../dQP
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/crossXDT => ../../proto/crossXDT
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/downXDT => ../../proto/downXDT
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/fnInvocation => ../../proto/fnInvocation
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/upXDT => ../../proto/upXDT
	github.com/ease-lab/vhive_stealth/examples/prototype/sdk => ../../sdk
	github.com/ease-lab/vhive_stealth/examples/prototype/sqp => ../../sQP
)

require (
	github.com/ease-lab/vhive_stealth/examples/prototype/sdk v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
)
