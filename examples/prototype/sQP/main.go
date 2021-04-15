package sqp

import (
	"io"
	"log"
	"net"

	crossQPProto "github.com/ease-lab/vhive_stealth/examples/prototype/proto/CrossQPProto"
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

// to be called by SrcFn to push data to sQP
func (s push_server) CollectData(srv SrcFnToQPProto.StreamData_CollectDataServer) error {
	packet_count := 1
	var payload []byte
	var key string
	for {
		packet, err := srv.Recv()
		if err == io.EOF {
			log.Printf("Complete packet received")
			// push to data_queue
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

func StartServer(serverAddr string) {

	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// create grpc server
	sdk_server := grpc.NewServer()
	SrcFnToQPProto.RegisterStreamDataServer(sdk_server, push_server{})
	crossQPProto.RegisterStreamDataServer(sdk_server, pull_server{})

	log.Println("start server")
	// and start...
	if err := sdk_server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
