module gx

go 1.15

replace github.com/ease-lab/vhive_stealth/examples/prototype/sdk => ../../sdk

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/upXDT => ../../proto/upXDT

replace github.com/ease-lab/vhive_stealth/examples/gRPC_stream/plotter => ../../plotter

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/crossXDT => ../../proto/crossXDT

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/fnInvocation => ../../proto/fnInvocation

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/downXDT => ../../proto/downXDT

replace github.com/ease-lab/vhive_stealth/examples/prototype/sqp => ../../sQP

replace github.com/ease-lab/vhive_stealth/examples/prototype/dqp => ../../dQP

require (
	github.com/ease-lab/vhive_stealth/examples/prototype/sdk v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
)
