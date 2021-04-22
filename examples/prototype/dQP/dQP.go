package dqp

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"time"

	crossXDT "github.com/ease-lab/vhive_stealth/examples/prototype/proto/crossXDT"
	downXDT "github.com/ease-lab/vhive_stealth/examples/prototype/proto/downXDT"
	fnInvocation "github.com/ease-lab/vhive_stealth/examples/prototype/proto/fnInvocation"
	sdk "github.com/ease-lab/vhive_stealth/examples/prototype/sdk"

	"google.golang.org/grpc"
)

var dataQueue = make(map[string][]byte)

type fnInvocationServer struct {
	fnInvocation.UnimplementedInvocationServer
}

type downXDTServer struct {
	downXDT.UnimplementedXDTtoFnServer
}

type payload struct {
	FunctionName string
	Data         []byte
	Key          string
}

var config = sdk.LoadConfig("../config.json")

// gRPC server to serve data to the DstFn
func (s downXDTServer) XDTDataServe(in *downXDT.DataRequest, srv downXDT.XDTtoFn_XDTDataServeServer) error {

	log.Printf("fetch key : %d", in.Key)

	blob := dataQueue[in.Key]
	blobLength := int64(len(blob))
	for currentByte := int64(0); currentByte < blobLength; currentByte += in.ChunkSize {

		if currentByte+in.ChunkSize > blobLength {
			resp := downXDT.Data{Chunk: blob[currentByte:blobLength]}
			if err := srv.Send(&resp); err != nil {
				log.Printf("send error %v", err)
			}
			log.Printf("finishing request number : %d", currentByte)
		} else {
			resp := downXDT.Data{Chunk: blob[currentByte : currentByte+in.ChunkSize]}
			if err := srv.Send(&resp); err != nil {
				log.Printf("send error %v", err)
			}
			log.Printf("finishing request number : %d", currentByte)
		}

	}
	return nil
}

// gRPC server to route the function call fron SrcFn to the DstFn
func (s fnInvocationServer) RouteInvocation(ctx context.Context, in *fnInvocation.InvocationRequest) (*fnInvocation.Empty, error) {

	log.Printf("received serialised json : %s", in.XdtJson)

	var xdtPayload payload
	if err := json.Unmarshal(in.XdtJson, &xdtPayload); err != nil {
		log.Fatal(err)
	}

	log.Printf("fetching data using key : %d", xdtPayload.Key)
	chunkSizeInBytes := config.ChunkSizeInBytes
	duration, payloadData := PullDataFromSrcQP(xdtPayload.Key, chunkSizeInBytes)
	log.Printf("pulled data from sQP in %s", duration)
	dataQueue[xdtPayload.Key] = payloadData

	// route the invocation call to destnation fn
	serverAddr := ":50007"
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := downXDT.NewXDTtoFnClient(conn)
	_, err = c.XDTFnCall(context.Background(), &downXDT.InvocationRequest{XdtJson: in.XdtJson})
	if err == nil {
		log.Printf("Fn invocation route at dQP successful")
	}

	return &fnInvocation.Empty{}, nil
}

// pull data from src QP to dst QP
func PullDataFromSrcQP(key string, chunkSizeInBytes int) (time.Duration, []byte) {

	serverAddr := ":50005"
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}
	start := time.Now()

	client := crossXDT.NewStreamDataClient(conn)
	in := &crossXDT.Request{Key: key, ChunkSize: int64(chunkSizeInBytes)}
	stream, err := client.ServeData(context.Background(), in)
	if err != nil {
		log.Fatalf("open stream error %v", err)
	}

	packetCount := 1
	var payload []byte
	for {
		packet, err := stream.Recv()
		if err == io.EOF {
			elapsed := time.Since(start)
			log.Printf("Complete packet received")
			dataQueue[key] = payload
			return elapsed, payload
		}
		if err != nil {
			log.Fatalf("receive error: %v", err)
		}
		log.Printf("Received chunk no. %d", packetCount)
		payload = append(payload, packet.Chunk...)
		packetCount += 1
	}
	return time.Duration(-1), []byte{}
}

// start DstQP server
func StartServer(serverAddr string) {

	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	downXDT.RegisterXDTtoFnServer(server, downXDTServer{})
	fnInvocation.RegisterInvocationServer(server, fnInvocationServer{})

	log.Println("start server")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
