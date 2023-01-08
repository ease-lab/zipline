@0x85d3acc39d94e0f8;

# Declare the serveData capability
interface StreamData {
	serveData @0 (key :Text) -> (payload :Data);
	serveBroadcastData @1 (key :Text) -> (payload :Data);
}