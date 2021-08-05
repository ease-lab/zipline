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

package dQP

import (
	"context"
	"io"
	"net"
	"sync"

	"google.golang.org/grpc/metadata"

	"github.com/ease-lab/vhive-xdt/proto/crossXDT"
	"github.com/ease-lab/vhive-xdt/proto/downXDT"

	"github.com/ease-lab/vhive-xdt/transport"
	"github.com/ease-lab/vhive-xdt/utils"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

// bufferPool is responsible for managing bounded buffers of channels to store data
var bufferPool transport.BufferPool

type downXDTServer struct {
	downXDT.UnimplementedXDTtoFnServer
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
				log.Infof("[dQP] freeing channel %s due to send error", in.Key)
				bufferPool.FreeChannel(in.Key)
				return err
			}
			log.Debugf("dQP: Sending chunk : %d to DstFn", chunkCount)
			chunkCount += 1
		default:
			if chunkTotal == int64(chunkCount) {
				log.Infof("[dQP] transfer to DST complete, freeing channel %s ", in.Key)
				bufferPool.FreeChannel(in.Key)
				return nil
			}
		}
	}
}

// PullDataFromSrcQP pulls data from src QP to dst QP
func PullDataFromSrcQP(ctx context.Context) error {

	headers, _ := metadata.FromOutgoingContext(ctx)
	key := headers["key"][0]
	sQPAddr := headers["sqp_addr"][0]
	routing := headers["routing"][0]
	conn, err := utils.GetGRPCConn(ctx, sQPAddr, true)
	if err != nil {
		log.Errorf("SRC: can not connect with server %v", err)
		return err
	}

	client := crossXDT.NewStreamDataClient(conn)
	in := &crossXDT.Request{Key: key}
	stream, err := client.ServeData(ctx, in)
	if err != nil {
		log.Errorf("dQP: open stream error %v", err)
		return err
	}

	chunkCount := 0
	var onlyOnce sync.Once
	var channel chan []byte
	var totalChunks int64
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			log.Infof("dQP: %d chunks received", chunkCount)
			if routing == utils.STORE_FORWARD {
				bufferPool.StoreChannel(key, totalChunks, channel)
			}
			return nil
		}
		if err != nil {
			log.Errorf("dQP: receive error: %v", err)
			return err
		}
		log.Debugf("dQP: Received chunk no. %d", chunkCount)
		onlyOnce.Do(func() {
			totalChunks = chunk.TotalChunks
			log.Debugf("dQP: requesting a new channel")
			if routing == utils.CUT_THROUGH {
				channel = bufferPool.CreateChannel()
				bufferPool.StoreChannel(key, chunk.TotalChunks, channel)
			} else if routing == utils.STORE_FORWARD {
				channel = bufferPool.CreateChannel()
			} else {
				log.Errorf("dQP: Invalid route type %s. Check config", routing)
			}
			log.Debugf("dQP: channel allocated")
			log.Infof("dQP: chunkTotal = %d", chunk.TotalChunks)
		})
		log.Debugf("dQP: Enquing chunk number %d", chunkCount)
		channel <- chunk.Chunk
		chunkCount += 1
	}
}

// StartServer starts DstQP server
func StartServer(config utils.Config) {

	bufferPool.Init(config)

	lis, err := net.Listen("tcp", config.DQPServerPort)
	if err != nil {
		log.Fatalf("dQP: failed to listen: %v", err)
	}
	var server *grpc.Server
	if config.TracingEnabled {
		server = grpc.NewServer(grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
			grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()))
	} else {
		server = grpc.NewServer()
	}
	downXDT.RegisterXDTtoFnServer(server, downXDTServer{})

	log.Infoln("dQP: start server")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("dQP: failed to serve: %v", err)
	}
}
