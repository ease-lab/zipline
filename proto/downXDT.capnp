using Go = import "/go.capnp";
@0x85d3acc39d94e0f8;
$Go.package("downXDT");
$Go.import("github.com/ease-lab/vhive-xdt/proto/downXDT");

# Declare the serveData capability
interface XDTtoFn {
	xDTDataServe @0 (key :Text) -> (payload :Data);
}