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

package dqp

import (
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"io"
	"net"
	"sync"

	"github.com/ease-lab/vhive_stealth/examples/prototype/proto/crossXDT"
	"github.com/ease-lab/vhive_stealth/examples/prototype/proto/downXDT"
	"github.com/ease-lab/vhive_stealth/examples/prototype/proto/fnInvocation"
	"github.com/ease-lab/vhive_stealth/examples/prototype/sdk"

	"google.golang.org/grpc"
)

// bufferPool is responsible for managing bounded buffers of channels to store data
var bufferPool sdk.BufferPool

type fnInvocationServer struct {
	fnInvocation.UnimplementedInvocationServer
}

type downXDTServer struct {
	downXDT.UnimplementedXDTtoFnServer
}

type payload struct {
	FunctionName string
	Data         []byte
	Key          string
}

// XDTDataServe is a gRPC server to serve data to the DstFn
func (s downXDTServer) XDTDataServe(in *downXDT.DataRequest, srv downXDT.XDTtoFn_XDTDataServeServer) error {

	log.Infof("dQP: data being fetched by DstFn using key : %s", in.Key)

	chunkCount := 0
	var channel chan []byte
	var chunkTotal int64
	// Check whether the first packet has been received at dQP or not
	for {
		if channel, chunkTotal = bufferPool.GetChannel(in.Key); channel != nil {
			log.Tracef("dQP: found chunkTotal %d for key %s at sQP", chunkTotal, in.Key)
			break
		}
	}

	for {
		select {
		case chunk := <-channel:
			resp := downXDT.Data{Chunk: chunk}
			if err := srv.Send(&resp); err != nil {
				log.Fatalf("dQP: send error %v", err)
			}
			log.Tracef("dQP: Sending chunk : %d to DstFn", chunkCount)
			chunkCount += 1
		default:
			if chunkTotal == int64(chunkCount) {
				bufferPool.FreeChannel(in.Key)
				return nil
			}
		}
	}
}

// RouteInvocation is a gRPC server to route the function call from SrcFn to the DstFn
func (s fnInvocationServer) RouteInvocation(ctx context.Context, in *fnInvocation.InvocationRequest) (*fnInvocation.Empty, error) {

	log.Infof("dQP: received serialised json: %s", in.XDTJSON)

	var xdtPayload payload
	if err := json.Unmarshal(in.XDTJSON, &xdtPayload); err != nil {
		log.Error(err)
	}

	log.Infof("dQP: fetching data from sQP using key : %s", xdtPayload.Key)
	chunkSizeInBytes := sdk.LoadedConfig.ChunkSizeInBytes

	if sdk.LoadedConfig.Routing == "CT" {
		log.Infof("dQP: CT: pulling data from sQP")
		go PullDataFromSrcQP(ctx, xdtPayload.Key, chunkSizeInBytes)
	} else if sdk.LoadedConfig.Routing == "S&F" {
		log.Infof("dQP: S&F: pulling data from sQP")
		PullDataFromSrcQP(ctx, xdtPayload.Key, chunkSizeInBytes)
	}

	// route the invocation call to destination fn
	serverAddr := sdk.LoadedConfig.DstServerAddr
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()))
	if err != nil {
		log.Errorf("did not connect: %v", err)
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			log.Errorf("dQP: Error closing the connection to Dest")
		}
	}()

	c := downXDT.NewXDTtoFnClient(conn)
	_, err = c.XDTFnCall(ctx, &downXDT.InvocationRequest{XDTJSON: in.XDTJSON})
	if err != nil {
		log.Infof("dQP: Fn invocation route unsuccessful")
	}
	log.Infof("dQP: Fn invocation route successful")
	return &fnInvocation.Empty{}, nil
}

// PullDataFromSrcQP pulls data from src QP to dst QP
func PullDataFromSrcQP(ctx context.Context, key string, chunkSizeInBytes int) {

	serverAddr := sdk.LoadedConfig.SQPServerAddr
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()))
	if err != nil {
		log.Errorf("dQP: can not connect with server %v", err)
	}

	client := crossXDT.NewStreamDataClient(conn)
	in := &crossXDT.Request{Key: key, ChunkSize: int64(chunkSizeInBytes)}
	stream, err := client.ServeData(ctx, in)
	if err != nil {
		log.Errorf("dQP: open stream error %v", err)
	}

	chunkCount := 0
	var onlyOnce sync.Once
	var channel chan []byte
	var totalChunks int64
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			log.Tracef("dQP: %d chunks received at Dst", chunkCount)
			if sdk.LoadedConfig.Routing == "S&F" {
				bufferPool.StoreChannel(key, totalChunks, channel)
			}
			return
		}
		if err != nil {
			log.Errorf("dQP: receive error: %v", err)
		}
		log.Tracef("dQP: Received chunk no. %d", chunkCount)
		onlyOnce.Do(func() {
			totalChunks = chunk.TotalChunks
			log.Infof("dQP: requesting a new channel")
			if sdk.LoadedConfig.Routing == "CT" {
				channel = bufferPool.CreateChannel()
				bufferPool.StoreChannel(key, chunk.TotalChunks, channel)
			} else if sdk.LoadedConfig.Routing == "S&F" {
				channel = bufferPool.CreateChannel()
			} else {
				log.Errorf("dQP: Invalid route type. Check config.json")
			}
			log.Infof("dQP: chunkTotal = %d", chunk.TotalChunks)
		})
		log.Infof("dQP: Enquing chunk number %d", chunkCount)
		channel <- chunk.Chunk
		chunkCount += 1
	}
}

// StartServer starts DstQP server
func StartServer(serverAddr string) {

	bufferPool.Init()

	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("dQP: failed to listen: %v", err)
	}

	server := grpc.NewServer(grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()))
	downXDT.RegisterXDTtoFnServer(server, downXDTServer{})
	fnInvocation.RegisterInvocationServer(server, fnInvocationServer{})

	log.Println("dQP: start server")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("dQP: failed to serve: %v", err)
	}
}
