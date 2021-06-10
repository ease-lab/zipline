module main

go 1.16

replace (
	XDTgRPC_stream/plotter => ../../plotter
	XDTprototype/utils => ../../utils
	XDTprototype/dqp => ../../dQP
	XDTprototype/proto/crossXDT => ../../proto/crossXDT
	XDTprototype/proto/downXDT => ../../proto/downXDT
	XDTprototype/proto/fnInvocation => ../../proto/fnInvocation
	XDTprototype/proto/upXDT => ../../proto/upXDT
	XDTprototype/sdk => ../../sdk
	XDTprototype/sqp => ../../sQP
	XDTprototype/tracing => ../../tracing
	XDTprototype/transport => ../../transport
)

require (
	XDTprototype/utils v0.0.0-00010101000000-000000000000
	XDTprototype/sdk v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
)
