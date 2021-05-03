package sqp

import (
	sdk "github.com/ease-lab/vhive_stealth/examples/prototype/sdk"
	"io"
	"sync"

	//"math"
	"net"
	"strconv"

	log "github.com/sirupsen/logrus"

	crossXDT "github.com/ease-lab/vhive_stealth/examples/prototype/proto/crossXDT"
	upXDT "github.com/ease-lab/vhive_stealth/examples/prototype/proto/upXDT"

	"google.golang.org/grpc"
)

var dataQueue sync.Map
var dataQueueSize sync.Map

type crossXDTServer struct {
	crossXDT.UnimplementedStreamDataServer
}

type upXDTServer struct {
	upXDT.UnimplementedStreamDataServer
}

// to be called by SrcFn to push data to sQP
func (s upXDTServer) SendData(srv upXDT.StreamData_SendDataServer) error {
	chunkCount := 0
	var key string
	var channel chan []byte
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
		if _,ok := dataQueue.Load(key); !ok {
			log.Infof("creating a new channel at sQP")
			channel = make(chan []byte, sdk.LoadedConfig.BufferSize)
			dataQueue.Store(key, channel)
			log.Infof("chunkTotal = %d",chunk.ChunkTotal)
			dataQueueSize.Store(key,chunk.ChunkTotal)
		}
		log.Infof("Enquing chunk number %d at sQP",chunkCount)
		channel <- chunk.Chunk
		chunkCount += 1
	}
	return nil
}

// gRPC server to serve the available data to the dQP
func (s crossXDTServer) ServeData(in *crossXDT.Request, srv crossXDT.StreamData_ServeDataServer) error {

	log.Infof("fetch key: %s from sQP", in.Key)

	chunkCount := 0
	var channel chan []byte
	var chunkTotal int64
	for {
		if tmp,ok := dataQueue.Load(in.Key); ok {
			log.Tracef("found channel for key %s",in.Key)
			channel = tmp.(chan []byte)
			break
		}
	}
	for {
		if tmp,ok := dataQueueSize.Load(in.Key); ok {
			chunkTotal = tmp.(int64)
			log.Tracef("found chunkTotal %d for key %s",chunkTotal,in.Key)
			break
		}
	}

	for {
		select {
		case chunk := <-channel:
			log.Tracef("chunk fetched from sQP using key %s", in.Key+";"+strconv.Itoa(chunkCount))
			resp := crossXDT.Response{Chunk: chunk, ChunkTotal: chunkTotal}
			if err := srv.Send(&resp); err != nil {
				log.Fatalf("send error %v", err)
			}
			log.Infof("pushing chunk no. %d to dQP", chunkCount)
			chunkCount += 1
		default:
			if chunkTotal == int64(chunkCount) {
				dataQueue.Delete(in.Key)
				dataQueueSize.Delete(in.Key)
				close(channel)
				return nil
			}
		}
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
