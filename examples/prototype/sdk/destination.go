package sdk

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"time"

	log "github.com/sirupsen/logrus"

	downXDT "github.com/ease-lab/vhive_stealth/examples/prototype/proto/downXDT"

	"google.golang.org/grpc"
)

// to be called by dQP to invoke DstFn
func (s downXDTServer) XDTFnCall(ctx context.Context, in *downXDT.InvocationRequest) (*downXDT.Empty, error) {

	log.Infof("destination received invocation call %s", in.XdtJson)

	var xdtPayload payload
	if err := json.Unmarshal(in.XdtJson, &xdtPayload); err != nil {
		log.Fatal(err)
	}

	key := xdtPayload.Key

	chunkSizeInBytes := config.ChunkSizeInBytes

	// fetch data from dQP
	FetchFromDQP(key, chunkSizeInBytes)
	return &downXDT.Empty{}, nil
}

// fetch data from dQP to DstFn
func FetchFromDQP(key string, chunkSizeInBytes int) (time.Duration, []byte) {
	serverAddr := ":50006"
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}
	start := time.Now()

	client := downXDT.NewXDTtoFnClient(conn)
	in := &downXDT.DataRequest{Key: key, ChunkSize: int64(chunkSizeInBytes)}
	stream, err := client.XDTDataServe(context.Background(), in)
	if err != nil {
		log.Fatalf("open stream error %v", err)
	}

	packetCount := 1
	var payload []byte
	for {
		packet, err := stream.Recv()
		if err == io.EOF {
			elapsed := time.Since(start)
			log.Infof("Complete packet received at dQP")
			dataQueue[key] = payload
			return elapsed, payload
		}
		if err != nil {
			log.Fatalf("receive error: %v", err)
		}
		log.Tracef("Received chunk no. %d", packetCount)
		payload = append(payload, packet.Chunk...)
		packetCount += 1
	}
}

// start DstQP server
func StartDstServer(serverAddr string) {

	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	downXDT.RegisterXDTtoFnServer(server, downXDTServer{})

	log.Println("start server")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
