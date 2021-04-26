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

var dataQueue = make(map[string][]byte)

type crossXDTServer struct {
	crossXDT.UnimplementedStreamDataServer
}

type upXDTServer struct {
	upXDT.UnimplementedStreamDataServer
}

// to be called by SrcFn to push data to sQP
func (s upXDTServer) SendData(srv upXDT.StreamData_SendDataServer) error {
	chunkCount := 0
	//var payload []byte
	var key string
	for {
		chunk, err := srv.Recv()
		if err == io.EOF {
			log.Infof("%d chunks received at sQP",chunkCount)
			return srv.SendAndClose(&upXDT.Empty{})
		}
		if err != nil {
			log.Fatalf("receive error: %v", err)
		}
		key = chunk.Key
		log.Tracef("Key received: %s in chunk %d", key, chunkCount)
		dataQueue[key+";"+strconv.Itoa(chunkCount)] = chunk.Chunk
		chunkCount += 1
	}
	return nil
}

// gRPC server to serve the available data to the dQP
func (s crossXDTServer) ServeData(in *crossXDT.Request, srv crossXDT.StreamData_ServeDataServer) error {

	log.Infof("fetch key: %d from sQP", in.Key)

	chunkCount := 0
	for {
		chunk, ok := dataQueue[in.Key+";"+strconv.Itoa(chunkCount)]
		if !ok {
			break
		}
		log.Tracef("chunk fetched from sQP using key %s",in.Key+";"+strconv.Itoa(chunkCount))
		resp := crossXDT.Response{Chunk:chunk }
		if err := srv.Send(&resp); err != nil {
			log.Fatalf("send error %v", err)
		}
		log.Tracef("finishing request number : %d", chunkCount)
		chunkCount +=1
	}
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
