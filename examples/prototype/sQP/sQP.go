package sqp

import (
	"io"
	//"math"
	"net"
	"strconv"

	log "github.com/sirupsen/logrus"

	crossXDT "github.com/ease-lab/vhive_stealth/examples/prototype/proto/crossXDT"
	upXDT "github.com/ease-lab/vhive_stealth/examples/prototype/proto/upXDT"

	"google.golang.org/grpc"
)

// 1562 chunks of size 64KB are required to store 100 MB
var dataQueue = make(map[string][]byte)

type crossXDTServer struct {
	crossXDT.UnimplementedStreamDataServer
}

type upXDTServer struct {
	upXDT.UnimplementedStreamDataServer
}

// to be called by SrcFn to push data to sQP
func (s upXDTServer) SendData(srv upXDT.StreamData_SendDataServer) error {
	packetCount := 0
	//var payload []byte
	var key string
	for {
		packet, err := srv.Recv()
		if err == io.EOF {
			log.Infof("%d chunks received at sQP",packetCount)
			//dataQueue[key] = payload
			return srv.SendAndClose(&upXDT.Empty{})
		}
		if err != nil {
			log.Fatalf("receive error: %v", err)
		}
		key = packet.Key
		log.Tracef("Key received: %s in chunk %d", key, packetCount)
		//payload = append(payload, packet.Chunk...
		//log.Infof("storing chunk using key %s",key+";"+strconv.Itoa(packetCount))
		dataQueue[key+";"+strconv.Itoa(packetCount)] = packet.Chunk
		packetCount += 1
	}
	return nil
}

// gRPC server to serve the available data to the dQP
func (s crossXDTServer) ServeData(in *crossXDT.Request, srv crossXDT.StreamData_ServeDataServer) error {

	log.Infof("fetch key: %d from sQP", in.Key)

	packetCount := 0
	for {
		chunk, ok := dataQueue[in.Key+";"+strconv.Itoa(packetCount)]
		if !ok {
			break
		}
		log.Tracef("chunk fetched from sQP using key %s",in.Key+";"+strconv.Itoa(packetCount))
		resp := crossXDT.Response{Chunk:chunk }
		if err := srv.Send(&resp); err != nil {
			log.Fatalf("send error %v", err)
		}
		log.Tracef("finishing request number : %d", packetCount)
		packetCount +=1
	}
	//blob := dataQueue[in.Key]
	//blobLength := int64(len(blob))
	//for currentByte := int64(0); currentByte < blobLength; currentByte += in.ChunkSize {
	//
	//	if currentByte+in.ChunkSize > blobLength {
	//		resp := crossXDT.Response{Chunk: blob[currentByte:blobLength]}
	//		if err := srv.Send(&resp); err != nil {
	//			log.Fatalf("send error %v", err)
	//		}
	//		log.Tracef("finishing request number : %d", currentByte)
	//	} else {
	//		resp := crossXDT.Response{Chunk: blob[currentByte : currentByte+in.ChunkSize]}
	//		if err := srv.Send(&resp); err != nil {
	//			log.Fatalf("send error %v", err)
	//		}
	//		log.Tracef("finishing request number : %d", currentByte)
	//	}
	//
	//}
	return nil
}

// start SrcQP server
func StartServer(serverAddr string) {

	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	upXDT.RegisterStreamDataServer(server, upXDTServer{})
	crossXDT.RegisterStreamDataServer(server, crossXDTServer{})

	log.Println("start server")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
