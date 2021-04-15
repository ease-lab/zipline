package dqp

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"time"

	crossQPProto "github.com/ease-lab/vhive_stealth/examples/prototype/proto/CrossQPProto"
	FnInvocationProto "github.com/ease-lab/vhive_stealth/examples/prototype/proto/FnInvocationProto"
	QPToDstFnProto "github.com/ease-lab/vhive_stealth/examples/prototype/proto/QPToDstFnProto"
	SrcFnToQPProto "github.com/ease-lab/vhive_stealth/examples/prototype/proto/SrcFnToQPProto"

	"google.golang.org/grpc"
)

var data_queue = make(map[string][]byte)

type pull_server struct {
	crossQPProto.UnimplementedStreamDataServer
}

type push_server struct {
	SrcFnToQPProto.UnimplementedStreamDataServer
}

type control_call_server struct {
	FnInvocationProto.UnimplementedInvocationServer
}

type xdt_to_dst struct {
	QPToDstFnProto.UnimplementedXDTtoFnServer
}

type payload struct {
	FunctionName string
	Data         []byte
	Key          string
}

// gRPC server to serve the available data to the DstFn
func (s xdt_to_dst) XDTDataServe(in *QPToDstFnProto.DataRequest, srv QPToDstFnProto.XDTtoFn_XDTDataServeServer) error {
	//serve data to the dQP
	log.Printf("fetch key : %d", in.Key)

	blob := data_queue[in.Key]
	blob_length := int64(len(blob))
	for currentByte := int64(0); currentByte < blob_length; currentByte += in.ChunkSize {

		if currentByte+in.ChunkSize > blob_length {
			resp := QPToDstFnProto.Data{Chunk: blob[currentByte:blob_length]}
			if err := srv.Send(&resp); err != nil {
				log.Printf("send error %v", err)
			}
			log.Printf("finishing request number : %d", currentByte)
		} else {
			resp := QPToDstFnProto.Data{Chunk: blob[currentByte : currentByte+in.ChunkSize]}
			if err := srv.Send(&resp); err != nil {
				log.Printf("send error %v", err)
			}
			log.Printf("finishing request number : %d", currentByte)
		}

	}
	return nil
}

// gRPC server to serve the available data to the dQP
func (s control_call_server) RouteInvocationCall(ctx context.Context, in *FnInvocationProto.InvocationRequest) (*FnInvocationProto.Empty, error) {

	log.Printf("received serialised json : %s", in.XdtJson)

	var xdtPayload payload
	if err := json.Unmarshal(in.XdtJson, &xdtPayload); err != nil {
		log.Fatal(err)
	}

	log.Printf("fetching data using key : %d", xdtPayload.Key)

	chunkSizeInBytes := 64 * 1024

	duration, payloadData := PullDataFromSrcQP(xdtPayload.Key, chunkSizeInBytes)

	log.Printf("pulled data from sQP in %s", duration)

	data_queue[xdtPayload.Key] = payloadData

	// send the invocation call to destnation fn

	serverAddr := ":50007"

	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := QPToDstFnProto.NewXDTtoFnClient(conn)

	_, err = c.XDTFnCall(context.Background(), &QPToDstFnProto.InvocationRequest{XdtJson: in.XdtJson})
	if err == nil {
		log.Printf("Fn invocation route at dQP successful")
	}

	return &FnInvocationProto.Empty{}, nil
}

// gRPC server for sQP to serve the available data to the dQP
func (s pull_server) ServeData(in *crossQPProto.Request, srv crossQPProto.StreamData_ServeDataServer) error {
	//serve data to the dQP
	log.Printf("fetch key : %d", in.Key)

	blob := data_queue[in.Key]
	blob_length := int64(len(blob))
	for currentByte := int64(0); currentByte < blob_length; currentByte += in.ChunkSize {

		if currentByte+in.ChunkSize > blob_length {
			resp := crossQPProto.Response{Chunk: blob[currentByte:blob_length]}
			if err := srv.Send(&resp); err != nil {
				log.Printf("send error %v", err)
			}
			log.Printf("finishing request number : %d", currentByte)
		} else {
			resp := crossQPProto.Response{Chunk: blob[currentByte : currentByte+in.ChunkSize]}
			if err := srv.Send(&resp); err != nil {
				log.Printf("send error %v", err)
			}
			log.Printf("finishing request number : %d", currentByte)
		}

	}
	return nil
}

// pullDataFromSrcQP pull data from src QP to dst QP
func PullDataFromSrcQP(key string, chunk_size_in_bytes int) (time.Duration, []byte) {

	serverAddr := ":50005"
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}
	start := time.Now()

	// create stream
	client := crossQPProto.NewStreamDataClient(conn)
	in := &crossQPProto.Request{Key: key, ChunkSize: int64(chunk_size_in_bytes)}
	stream, err := client.ServeData(context.Background(), in)
	if err != nil {
		log.Fatalf("open stream error %v", err)
	}

	packet_count := 1
	var payload []byte
	for {
		packet, err := stream.Recv()
		if err == io.EOF {
			elapsed := time.Since(start)
			log.Printf("Complete packet received")
			// push to data_queue
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

func StartServer(serverAddr string) {

	// create listener for sdk
	lis_to_sdk, err := net.Listen("tcp", "serverAddr")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// create grpc server
	sdk_server := grpc.NewServer()
	QPToDstFnProto.RegisterXDTtoFnServer(sdk_server, xdt_to_dst{})
	FnInvocationProto.RegisterInvocationServer(sdk_server, control_call_server{})

	log.Println("start server")
	// and start...
	if err := sdk_server.Serve(lis_to_sdk); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
