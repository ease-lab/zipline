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
	"io"
	"net"
	"sync"

	"google.golang.org/grpc/metadata"

	"github.com/ease-lab/vhive-xdt/proto/downXDT"
	"github.com/ease-lab/vhive-xdt/utils"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"google.golang.org/grpc"
)

type downXDTServer struct {
	config  utils.Config
	handler func([]byte)
	client  downXDT.XDTtoFnClient
	downXDT.UnimplementedXDTtoFnServer
}

// XDTFnCall is to be called by dQP to invoke DstFn
func (s downXDTServer) XDTFnCall(ctx context.Context, in *downXDT.InvocationRequest) (*downXDT.InvocationResponse, error) {

	log.Infof("DST: received invocation call %s", in.XDTJSON)

	headers, ok := metadata.FromIncomingContext(ctx)

	if ok && headers["is_xdt"][0] == "true" {
		key := headers["key"][0]
		log.Infof("DST: using %s routing", headers["routing"][0])

		// fetch data from dQP
		payloadBytes, err := FetchFromDQP(ctx, key, s.client, s.config)
		if err != nil {
			log.Errorf("DST: FetchFromDQP failed %v", err)
			return &downXDT.InvocationResponse{}, err
		}

		//call destination function
		s.handler(payloadBytes)
	}

	return &downXDT.InvocationResponse{}, nil
}

// FetchFromDQP fetches data from dQP to DstFn
func FetchFromDQP(ctx context.Context, key string, client downXDT.XDTtoFnClient, config utils.Config) ([]byte, error) {

	in := &downXDT.DataRequest{Key: key}
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
			payloadBytes = make([]byte, config.StAndFwBufferSize*config.ChunkSizeInBytes)
			log.Infof("DST: chunkTotal = %d", totalChunks)
		})
		log.Debugf("DST: appending chunk number %d", chunkCount)
		copy(payloadBytes[byteCount:], chunk.Chunk)
		byteCount += len(chunk.Chunk)
		chunkCount += 1
	}
}

// StartDstServer starts DstQP server
func StartDstServer(config utils.Config, handler func([]byte)) {

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
	// connect to DQP
	conn, err := utils.GetGRPCConn(context.Background(), config.DQPServerHostname+config.DQPServerPort, false)
	if err != nil {
		log.Fatalf("DST: can not connect with dQP server %v", err)
	}
	client := downXDT.NewXDTtoFnClient(conn)

	s := downXDTServer{}
	s.config = config
	s.handler = handler
	s.client = client
	downXDT.RegisterXDTtoFnServer(server, s)

	log.Infoln("DST: start server")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("DST: failed to serve: %v", err)
	}
}
