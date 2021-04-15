module dqp

go 1.15

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/upXDT => ../proto/upXDT

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/crossXDT => ../proto/crossXDT

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/fnInvocation => ../proto/fnInvocation

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/downXDT => ../proto/downXDT

require (
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/crossXDT v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/fnInvocation v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/downXDT v0.0.0-00010101000000-000000000000
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/upXDT v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.36.0
)
