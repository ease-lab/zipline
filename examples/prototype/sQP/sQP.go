// MIT License
//
// Copyright (c) 2021 Shyam Jesalpura and EASE lab
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package sqp

import (
	"github.com/ease-lab/vhive_stealth/examples/prototype/sdk"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"io"
	"sync"

	log "github.com/sirupsen/logrus"
	"net"

	"github.com/ease-lab/vhive_stealth/examples/prototype/proto/crossXDT"
	"github.com/ease-lab/vhive_stealth/examples/prototype/proto/upXDT"

	"google.golang.org/grpc"
)

// dataQueue stores a chan []bytes per payload addressed by transaction ID
var dataQueue sync.Map
// dataQueueSize stores a total number of chunks per payload addressed by transaction ID
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
			log.Infof("sQP: chunkTotal = %d",chunk.TotalChunks)
			dataQueueSize.Store(key,chunk.TotalChunks)
		}
		log.Infof("sQP: Enquing chunk number %d",chunkCount)
		channel <- chunk.Chunk
		chunkCount += 1
	}
}

// ServeData is the gRPC server to serve the available data to the dQP
func (s crossXDTServer) ServeData(in *crossXDT.Request, srv crossXDT.StreamData_ServeDataServer) error {

	log.Infof("sQP: DQP is fetching key: %s", in.Key)

	chunkCount := 0
	var channel chan []byte
	var chunkTotal int64

	// Check whether the first packet has been received at sQP or not
	for {
		if tmp,ok := dataQueueSize.Load(in.Key); ok {
			chunkTotal = tmp.(int64)
			log.Tracef("sQP: found chunkTotal %d for key %s",chunkTotal,in.Key)
			break
		}
	}
	// Check whether the channel has been created by the receiving function
	for {
		if tmp,ok := dataQueue.Load(in.Key); ok {
			log.Tracef("sQP: found channel for key %s",in.Key)
			channel = tmp.(chan []byte)
			break
		}
	}
	// Send packets from the channel one by one
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
}

// StartServer starts the SrcQP server
func StartServer(serverAddr string) {

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
