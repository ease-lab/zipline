package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"time"

	QPToDstFnProto "github.com/ease-lab/vhive_stealth/examples/prototype/proto/QPToDstFnProto"
	"google.golang.org/grpc"
)

type destination_server struct {
	// Embed the unimplemented server
	QPToDstFnProto.UnimplementedXDTtoFnServer
}

type payload struct {
	FunctionName string
	Data         []byte
	Key          string
}

var data_queue = make(map[string][]byte)

func (s destination_server) XDTFnCall(ctx context.Context, in *QPToDstFnProto.InvocationRequest) (*QPToDstFnProto.Empty, error) {

	log.Printf("destination received invocation call %s", in.XdtJson)

	var xdtPayload payload
	if err := json.Unmarshal(in.XdtJson, &xdtPayload); err != nil {
		log.Fatal(err)
	}

	key := xdtPayload.Key

	chunkSizeInBytes := 64 * 1024

	// fetch data from dQP
	fetchFromDQP(key, chunkSizeInBytes)
	return &QPToDstFnProto.Empty{}, nil
}

func fetchFromDQP(key string, chunkSizeInBytes int) (time.Duration, []byte) {
	serverAddr := ":50006"
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}
	start := time.Now()

	// create stream
	client := QPToDstFnProto.NewXDTtoFnClient(conn)
	in := &QPToDstFnProto.DataRequest{Key: key, ChunkSize: int64(chunkSizeInBytes)}
	stream, err := client.XDTDataServe(context.Background(), in)
	if err != nil {
		log.Fatalf("open stream error %v", err)
	}
	//receive data from source QP
	// push to data_queue
	packet_count := 1
	var payload []byte
	for {
		packet, err := stream.Recv()
		if err == io.EOF {
			elapsed := time.Since(start)
			log.Printf("Complete packet received at dQP")
			data_queue[key] = payload
			return elapsed, payload
		}
		if err != nil {
			log.Fatalf("receive error: %v", err)
		}
		log.Printf("Received chunk no. %d", packet_count)
		payload = append(payload, packet.Chunk...)
		packet_count += 1
	}
	return time.Duration(-1), []byte{}
}

func main() {

	// create listener for sdk
	lis_to_sdk, err := net.Listen("tcp", ":50007")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// create grpc server
	sdk_server := grpc.NewServer()
	// SrcFnToQPProto.RegisterStreamDataServer(sdk_server, control_call_server{})
	// crossQPProto.RegisterStreamDataServer(sdk_server, pull_server{})
	QPToDstFnProto.RegisterXDTtoFnServer(sdk_server, destination_server{})

	log.Println("start server")
	// and start...
	if err := sdk_server.Serve(lis_to_sdk); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
