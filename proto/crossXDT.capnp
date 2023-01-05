using Go = import "/go.capnp";
@0x85d3acc39d94e0f8;
$Go.package("crossXDT");
$Go.import("github.com/ease-lab/vhive-xdt/proto/crossXDT");

# Declare the serveData capability
interface StreamData {
	serveData @0 (key :Text) -> (payload :Data);
	serveBroadcastData @0 (key :Text) -> (payload :Data);
}