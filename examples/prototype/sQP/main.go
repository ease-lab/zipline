package sqp

import (
	"io"
	"log"
	"net"

	crossXDT "github.com/ease-lab/vhive_stealth/examples/prototype/proto/crossXDT"
	upXDT "github.com/ease-lab/vhive_stealth/examples/prototype/proto/upXDT"

	"google.golang.org/grpc"
)

var dataQueue = make(map[string][]byte)

type crossXDTServer struct {
	crossXDT.UnimplementedStreamDataServer
}

type upXDTServer struct {
	upXDT.UnimplementedStreamDataServer
}

// to be called by SrcFn to push data to sQP
func (s upXDTServer) CollectData(srv upXDT.StreamData_CollectDataServer) error {
	packetCount := 1
	var payload []byte
	var key string
	for {
		packet, err := srv.Recv()
		if err == io.EOF {
			log.Printf("Complete packet received")
			// push to dataQueue
			dataQueue[key] = payload
			return srv.SendAndClose(&upXDT.Empty{})
		}
		if err != nil {
			log.Fatalf("receive error: %v", err)
		}
		key = packet.Key
		log.Printf("Key received: %s in chunk %d", key, packetCount)
		payload = append(payload, packet.Chunk...)
		packetCount += 1
	}
	return nil
}

// gRPC server to serve the available data to the dQP
func (s crossXDTServer) ServeData(in *crossXDT.Request, srv crossXDT.StreamData_ServeDataServer) error {

	log.Printf("fetch key : %d", in.Key)

	blob := dataQueue[in.Key]
	blobLength := int64(len(blob))
	for currentByte := int64(0); currentByte < blobLength; currentByte += in.ChunkSize {

		if currentByte+in.ChunkSize > blobLength {
			resp := crossXDT.Response{Chunk: blob[currentByte:blobLength]}
			if err := srv.Send(&resp); err != nil {
				log.Printf("send error %v", err)
			}
			log.Printf("finishing request number : %d", currentByte)
		} else {
			resp := crossXDT.Response{Chunk: blob[currentByte : currentByte+in.ChunkSize]}
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
	server := grpc.NewServer()
	upXDT.RegisterStreamDataServer(server, upXDTServer{})
	crossXDT.RegisterStreamDataServer(server, crossXDTServer{})

	log.Println("start server")
	// and start...
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
