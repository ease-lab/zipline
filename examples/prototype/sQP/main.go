package main

import (
	"context"
	"crypto/rand"
	"io"
	"log"
	"net"
	"time"

	crossQPProto "github.com/ease-lab/vhive_stealth/examples/prototype/proto/CrossQPProto"
	SrcFnToQPProto "github.com/ease-lab/vhive_stealth/examples/prototype/proto/SrcFnToQPProto"

	"google.golang.org/grpc"
)

var data_queue = make(map[string][]byte)

type pull_server struct {
	// Embed the unimplemented server
	crossQPProto.UnimplementedStreamDataServer
}

type push_server struct {
	// Embed the unimplemented server
	SrcFnToQPProto.UnimplementedStreamDataServer
}

// to be called by client to push data
func (s push_server) CollectData(srv SrcFnToQPProto.StreamData_CollectDataServer) error {
	//receive data from user-container
	// push to data_queue
	packet_count := 1
	var payload []byte
	var key string
	for {
		packet, err := srv.Recv()
		if err == io.EOF {
			log.Printf("Complete packet received")
			data_queue[key] = payload
			return srv.SendAndClose(&SrcFnToQPProto.Empty{})
		}
		if err != nil {
			log.Fatalf("receive error: %v", err)
		}
		key = packet.Key
		log.Printf("Key received: %s in chunk %d", key, packet_count)
		payload = append(payload, packet.Chunk...)
		packet_count += 1
	}
	return nil
}

// gRPC server to serve the available data to the dQP
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

func pullData(key string, chunk_size_in_bytes int) (time.Duration, []byte) {

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
	//receive data from source QP
	// push to data_queue
	packet_count := 1
	var payload []byte
	for {
		packet, err := stream.Recv()
		if err == io.EOF {
			elapsed := time.Since(start)
			log.Printf("Complete packet received")
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

	payload_data := make([]byte, 10*1024*1024) // 10MiB
	//create random blob
	rand.Read(payload_data)

	data_queue["123456789"] = payload_data
	// create listener for sdk
	lis_to_sdk, err := net.Listen("tcp", ":50005")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// create grpc server
	sdk_server := grpc.NewServer()
	SrcFnToQPProto.RegisterStreamDataServer(sdk_server, push_server{})
	crossQPProto.RegisterStreamDataServer(sdk_server, pull_server{})

	log.Println("start server")
	// and start...
	if err := sdk_server.Serve(lis_to_sdk); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
