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

package sdk

import (
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"io"
	"net"
	"sync"

	"XDTprototype/proto/downXDT"

	"google.golang.org/grpc"
)

type downXDTServer struct {
	downXDT.UnimplementedXDTtoFnServer
}

var DestinationHandler func([]byte)

// XDTFnCall is to be called by dQP to invoke DstFn
func (s downXDTServer) XDTFnCall(ctx context.Context, in *downXDT.InvocationRequest) (*downXDT.Empty, error) {

	log.Infof("DST: received invocation call %s", in.XDTJSON)

	var xdtPayload Payload
	if err := json.Unmarshal(in.XDTJSON, &xdtPayload); err != nil {
		log.Fatal(err)
	}

	key := xdtPayload.Key

	chunkSizeInBytes := LoadedConfig.ChunkSizeInBytes

	// fetch data from dQP
	payloadBytes := FetchFromDQP(ctx, key, chunkSizeInBytes)
	//call destination function
	DestinationHandler(payloadBytes)
	return &downXDT.Empty{}, nil
}

// FetchFromDQP fetches data from dQP to DstFn
func FetchFromDQP(ctx context.Context, key string, chunkSizeInBytes int) []byte {
	serverAddr := LoadedConfig.DQPServerAddr
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()))
	if err != nil {
		log.Fatalf("DST: can not connect with server %v", err)
	}

	client := downXDT.NewXDTtoFnClient(conn)
	in := &downXDT.DataRequest{Key: key, ChunkSize: int64(chunkSizeInBytes)}
	stream, err := client.XDTDataServe(ctx, in)
	if err != nil {
		log.Fatalf("DST: open stream error %v", err)
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
			log.Trace(payloadBytes[0:9], payloadBytes[byteCount-9:byteCount])
			return payloadBytes[:byteCount]
		}
		if err != nil {
			log.Fatalf("DST: receive error: %v", err)
		}
		log.Tracef("DST: Received chunk no. %d", chunkCount)
		onlyOnce.Do(func() {
			totalChunks = chunk.TotalChunks
			log.Infof("DST: creating a new buffer")
			payloadBytes = make([]byte, LoadedConfig.StAndFwBufferSize*LoadedConfig.ChunkSizeInBytes)
			log.Infof("sQP: chunkTotal = %d", totalChunks)
		})
		log.Infof("DST: appending chunk number %d", chunkCount)
		copy(payloadBytes[byteCount:], chunk.Chunk)
		byteCount += len(chunk.Chunk)
		chunkCount += 1
	}

}

// StartDstServer starts DstQP server
func StartDstServer(serverAddr string, handler func([]byte)) {

	DestinationHandler = handler

	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("DST: failed to listen: %v", err)
	}
	server := grpc.NewServer(grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()))
	downXDT.RegisterXDTtoFnServer(server, downXDTServer{})

	log.Println("DST: start server")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("DST: failed to serve: %v", err)
	}
}
