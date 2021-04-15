module gx

go 1.15

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/upXDT => ../proto/upXDT

replace github.com/ease-lab/vhive_stealth/examples/gRPC_stream/plotter => ../plotter

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/FnInvocationProto => ../proto/FnInvocation

replace github.com/ease-lab/vhive_stealth/examples/prototype/proto/downXDT => ../proto/downXDT

require (
	github.com/ease-lab/vhive_stealth/examples/prototype/proto/downXDT v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859 // indirect
	golang.org/x/sys v0.0.0-20210304124612-50617c2ba197 // indirect
	golang.org/x/text v0.3.5 // indirect
	google.golang.org/grpc v1.36.0
)
