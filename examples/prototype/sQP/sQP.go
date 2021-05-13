package sqp

import (
	sdk "github.com/ease-lab/vhive_stealth/examples/prototype/sdk"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"io"
	"sync"

	log "github.com/sirupsen/logrus"
	"net"

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

// SendData is called by SrcFn to push data to sQP
func (s upXDTServer) SendData(srv upXDT.StreamData_SendDataServer) error {
	chunkCount := 0
	var key string
	var channel chan []byte
	for {
		chunk, err := srv.Recv()
		if err == io.EOF {
			log.Infof("sQP: %d chunks received",chunkCount)
			if sdk.LoadedConfig.Routing == "S&F" {
				dataQueue.Store(key, channel)
			}
			return srv.SendAndClose(&upXDT.Empty{})
		}
		if err != nil {
			log.Fatalf("sQP: receive error: %v", err)
		}
		key = chunk.Key
		log.Tracef("sQP: Key received: %s in chunk %d", key, chunkCount)
		if _,ok := dataQueueSize.Load(key); !ok {
			log.Infof("sQP: creating a new channel")
			if sdk.LoadedConfig.Routing == "CT" {
				channel = make(chan []byte, sdk.LoadedConfig.BufferSize)
				dataQueue.Store(key, channel)
			}else if sdk.LoadedConfig.Routing == "S&F" {
				channel = make(chan []byte, sdk.LoadedConfig.StAndFwBufferSize)
			}else {
				log.Errorf("sQP: Invalid route type. Check config.json")
			}
			log.Infof("sQP: chunkTotal = %d",chunk.ChunkTotal)
			dataQueueSize.Store(key,chunk.ChunkTotal)
		}
		log.Infof("sQP: Enquing chunk number %d",chunkCount)
		channel <- chunk.Chunk
		chunkCount += 1
	}
	return nil
}

// ServeData is the gRPC server to serve the available data to the dQP
func (s crossXDTServer) ServeData(in *crossXDT.Request, srv crossXDT.StreamData_ServeDataServer) error {

	log.Infof("sQP: DQP is fetching key: %s", in.Key)

	chunkCount := 0
	var channel chan []byte
	var chunkTotal int64
	for {
		if tmp,ok := dataQueueSize.Load(in.Key); ok {
			chunkTotal = tmp.(int64)
			log.Tracef("sQP: found chunkTotal %d for key %s",chunkTotal,in.Key)
			break
		}
	}
	for {
		if tmp,ok := dataQueue.Load(in.Key); ok {
			log.Tracef("sQP: found channel for key %s",in.Key)
			channel = tmp.(chan []byte)
			break
		}
	}

	for {
		select {
		case chunk := <-channel:
			resp := crossXDT.Response{Chunk: chunk, ChunkTotal: chunkTotal}
			if err := srv.Send(&resp); err != nil {
				log.Fatalf("sQP: send error %v", err)
			}
			log.Infof("sQP: pushing chunk no. %d to dQP", chunkCount)
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

// StartServer starts the SrcQP server
func StartServer(serverAddr string) {

	//shutdown := sdk.InitTracer()
	//defer shutdown()

	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("sQP: failed to listen: %v", err)
	}

	server := grpc.NewServer(grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()))
	upXDT.RegisterStreamDataServer(server, upXDTServer{})
	crossXDT.RegisterStreamDataServer(server, crossXDTServer{})

	log.Println("sQP: start server")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("sQP: failed to serve: %v", err)
	}

}
