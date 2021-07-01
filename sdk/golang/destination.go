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

package golang

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"sync"

	"github.com/ease-lab/vhive-xdt/proto/downXDT"
	"github.com/ease-lab/vhive-xdt/utils"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"google.golang.org/grpc"
)

type downXDTServer struct {
	downXDT.UnimplementedXDTtoFnServer
}

var config utils.Config
var DestinationHandler func([]byte)

// XDTFnCall is to be called by dQP to invoke DstFn
func (s downXDTServer) XDTFnCall(ctx context.Context, in *downXDT.InvocationRequest) (*downXDT.Empty, error) {

	log.Infof("DST: received invocation call %s", in.XDTJSON)

	var xdtPayload utils.Payload
	if err := json.Unmarshal(in.XDTJSON, &xdtPayload); err != nil {
		log.Error("DST: XDTFnCall", err)
		return &downXDT.Empty{}, err
	}

	key := xdtPayload.Key

	chunkSizeInBytes := config.ChunkSizeInBytes

	// fetch data from dQP
	payloadBytes, err := FetchFromDQP(ctx, key, config.DQPServerHostname+config.DQPServerPort, chunkSizeInBytes)
	if err != nil {
		log.Errorf("DST: FetchFromDQP failed %v", err)
		return &downXDT.Empty{}, err
	}

	//call destination function
	DestinationHandler(payloadBytes)
	return &downXDT.Empty{}, nil
}

// FetchFromDQP fetches data from dQP to DstFn
func FetchFromDQP(ctx context.Context, key string, dQPAddr string, chunkSizeInBytes int) ([]byte, error) {

	conn, err := utils.GetGRPCConn(ctx, dQPAddr, false)
	if err != nil {
		log.Errorf("DST: can not connect with server %v", err)
		return []byte{}, err
	}

	client := downXDT.NewXDTtoFnClient(conn)
	in := &downXDT.DataRequest{Key: key, ChunkSize: int64(chunkSizeInBytes)}
	stream, err := client.XDTDataServe(ctx, in)
	if err != nil {
		log.Errorf("DST: open stream error %v", err)
		return []byte{}, err
	}

	chunkCount := 0
	byteCount := 0
	var onlyOnce sync.Once
	var totalChunks int64
	var payloadBytes []byte
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			log.Infof("DST: Received %d chunks at DstFn with first/last bytes as:", chunkCount)
			log.Info(payloadBytes[0:9], payloadBytes[byteCount-9:byteCount])
			return payloadBytes[:byteCount], nil
		}
		if err != nil {
			log.Errorf("DST: receive error: %v", err)
			return []byte{}, err
		}
		log.Debugf("DST: Received chunk no. %d", chunkCount)
		onlyOnce.Do(func() {
			totalChunks = chunk.TotalChunks
			log.Infof("DST: creating a new buffer")
			payloadBytes = make([]byte, utils.LoadConfig.StAndFwBufferSize*utils.LoadConfig.ChunkSizeInBytes)
			log.Infof("DST: chunkTotal = %d", totalChunks)
		})
		log.Debugf("DST: appending chunk number %d", chunkCount)
		copy(payloadBytes[byteCount:], chunk.Chunk)
		byteCount += len(chunk.Chunk)
		chunkCount += 1
	}
}

// StartDstServer starts DstQP server
func StartDstServer(receivedConfig utils.Config, handler func([]byte)) {

	config = receivedConfig
	DestinationHandler = handler

	lis, err := net.Listen("tcp", config.DstServerPort)
	if err != nil {
		log.Fatalf("DST: failed to listen: %v", err)
	}
	var server *grpc.Server
	if config.TracingEnabled {
		server = grpc.NewServer(grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
			grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()))
	} else {
		server = grpc.NewServer()
	}
	downXDT.RegisterXDTtoFnServer(server, downXDTServer{})

	log.Infoln("DST: start server")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("DST: failed to serve: %v", err)
	}
}
