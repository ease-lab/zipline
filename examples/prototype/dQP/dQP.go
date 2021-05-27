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
	"golang.org/x/sync/errgroup"
	"io"
	"net"
	"sync"
	"time"

	"XDTprototype/commonUtils"
	"XDTprototype/proto/crossXDT"
	"XDTprototype/proto/downXDT"
	"XDTprototype/proto/fnInvocation"
	"XDTprototype/transport"
	"google.golang.org/grpc"
)

// bufferPool is responsible for managing bounded buffers of channels to store data
var bufferPool transport.BufferPool

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
			log.Infof("dQP: found chunkTotal %d for key %s at sQP", chunkTotal, in.Key)
			break
		}
	}

	for {
		select {
		case chunk := <-channel:
			resp := downXDT.Data{Chunk: chunk, TotalChunks: chunkTotal}
			if err := srv.Send(&resp); err != nil {
				log.Errorf("dQP: send error %v", err)
				return err
			}
			log.Debugf("dQP: Sending chunk : %d to DstFn", chunkCount)
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
		log.Error("dQP: RouteInvocation", err)
		return &fnInvocation.Empty{}, err
	}

	chunkSizeInBytes := commonUtils.LoadedConfig.ChunkSizeInBytes
	log.Infof("dQP: fetched data from sQP using key : %s", xdtPayload.Key)

	if commonUtils.LoadedConfig.Routing == commonUtils.CUT_THROUGH {
		log.Infof("dQP: [cut-through]: pulling data from sQP")
		go PullDataFromSrcQP(ctx, xdtPayload.Key, chunkSizeInBytes)
	} else if commonUtils.LoadedConfig.Routing == commonUtils.STORE_FORWARD {
		log.Infof("dQP: [Store & Forward]: pulling data from sQP")
		PullDataFromSrcQP(ctx, xdtPayload.Key, chunkSizeInBytes)
	}

	// route the invocation call to destination fn
	//  This timeout must be large enough for the request to complete
	timeoutDuration := time.Duration(commonUtils.LoadedConfig.RPCTimeoutDurationInMiliSecs) * time.Millisecond
	ctxx, cancel := context.WithTimeout(ctx, timeoutDuration)
	defer cancel()

	conn, err := grpc.DialContext(ctxx, commonUtils.LoadedConfig.DstServerAddr, commonUtils.GetGopts()...)
	if err != nil {
		log.Errorf("dQP: RouteInvocation: did not connect: %v", err)
		return &fnInvocation.Empty{}, err
	}

	c := downXDT.NewXDTtoFnClient(conn)
	_, err = c.XDTFnCall(ctx, &downXDT.InvocationRequest{XDTJSON: in.XDTJSON})
	if err != nil {
		log.Errorf("dQP: Fn invocation route unsuccessful")
		return &fnInvocation.Empty{}, err
	}
	log.Infof("dQP: Fn invocation route successful")

	errGroup, _ := errgroup.WithContext(ctx)
	errGroup.Go(func() error {
		err = conn.Close()
		if err != nil {
			log.Errorf("dQP: Error closing the connection to Dest")
			return err
		}
		return nil
	})

	return &fnInvocation.Empty{}, errGroup.Wait()
}

// PullDataFromSrcQP pulls data from src QP to dst QP
func PullDataFromSrcQP(ctx context.Context, key string, chunkSizeInBytes int) {
	//  This timeout must be large enough for the request to complete
	timeoutDuration := time.Duration(commonUtils.LoadedConfig.RPCTimeoutDurationInMiliSecs) * time.Millisecond
	ctxx, cancel := context.WithTimeout(ctx, timeoutDuration)
	defer cancel()

	conn, err := grpc.DialContext(ctxx, commonUtils.LoadedConfig.SQPServerAddr, commonUtils.GetGopts()...)
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
			log.Infof("dQP: %d chunks received at Dst", chunkCount)
			if commonUtils.LoadedConfig.Routing == commonUtils.STORE_FORWARD {
				bufferPool.StoreChannel(key, totalChunks, channel)
			}
			return
		}
		if err != nil {
			log.Errorf("dQP: receive error: %v", err)
		}
		log.Debugf("dQP: Received chunk no. %d", chunkCount)
		onlyOnce.Do(func() {
			totalChunks = chunk.TotalChunks
			log.Infof("dQP: requesting a new channel")
			if commonUtils.LoadedConfig.Routing == commonUtils.CUT_THROUGH {
				channel = bufferPool.CreateChannel()
				bufferPool.StoreChannel(key, chunk.TotalChunks, channel)
			} else if commonUtils.LoadedConfig.Routing == commonUtils.STORE_FORWARD {
				channel = bufferPool.CreateChannel()
			} else {
				log.Errorf("dQP: Invalid route type. Check config.json")
			}
			log.Infof("dQP: chunkTotal = %d", chunk.TotalChunks)
		})
		log.Debugf("dQP: Enquing chunk number %d", chunkCount)
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

	log.Infoln("dQP: start server")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("dQP: failed to serve: %v", err)
	}
}
