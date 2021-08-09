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
	"strings"
	"sync"

	"github.com/ease-lab/vhive-xdt/proto/crossXDT"

	"google.golang.org/grpc/metadata"

	"github.com/ease-lab/vhive-xdt/proto/downXDT"
	"github.com/ease-lab/vhive-xdt/utils"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"google.golang.org/grpc"
)

type downXDTServer struct {
	config  utils.Config
	handler func([]byte) ([]byte, bool)
	client  downXDT.XDTtoFnClient
	downXDT.UnimplementedXDTtoFnServer
}

// XDTFnCall is to be called by dQP to invoke DstFn
func (s downXDTServer) XDTFnCall(ctx context.Context, in *downXDT.InvocationRequest) (*downXDT.InvocationResponse, error) {

	log.Infof("DST: received invocation call %s", in.XDTJSON)

	headers, ok := metadata.FromIncomingContext(ctx)
	var message []byte
	if ok && headers["is_xdt"][0] == "true" {
		key := headers["key"][0]
		log.Infof("DST: using %s routing", headers["routing"][0])

		// fetch data from dQP
		payloadBytes, err := FetchFromDQP(ctx, key, s.client)
		if err != nil {
			log.Errorf("DST: FetchFromDQP failed %v", err)
			return &downXDT.InvocationResponse{}, err
		}

		//call destination function
		message, ok = s.handler(payloadBytes)
	}

	return &downXDT.InvocationResponse{Message: message, Ok: ok}, nil
}

// FetchFromDQP fetches data from dQP to DstFn
func FetchFromDQP(ctx context.Context, key string, client downXDT.XDTtoFnClient) ([]byte, error) {
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
			payloadBytes = make([]byte, int(totalChunks)*len(chunk.Chunk))
			log.Infof("DST: chunkTotal = %d", totalChunks)
		})
		log.Debugf("DST: appending chunk number %d", chunkCount)
		copy(payloadBytes[byteCount:], chunk.Chunk)
		byteCount += len(chunk.Chunk)
		chunkCount += 1
	}
}

// Get pulls payload from DQP server using the key
func Get(ctx context.Context, capability string, config utils.Config) ([]byte, error) {
	log.Infof("attempting Get using capability %s", capability)
	key := capability
	splitString := strings.SplitN(capability, "|", 2)
	sQPAddr := splitString[1]
	httpMetadata := map[string]string{
		"is_xdt":   "true",
		"key":      key,
		"sqp_addr": sQPAddr,
		"routing":  config.Routing,
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.New(httpMetadata))
	conn, err := utils.GetGRPCConn(ctx, sQPAddr, true)
	if err != nil {
		log.Errorf("DST: can not connect with SQP server %v", err)
		return nil, err
	}

	client := crossXDT.NewStreamDataClient(conn)
	in := &crossXDT.Request{Key: key}
	stream, err := client.ServeData(ctx, in)
	if err != nil {
		log.Errorf("dST: open stream error %v", err)
		return nil, err
	}

	chunkCount := 0
	byteCount := 0
	var onlyOnce sync.Once
	var payloadBytes []byte
	var totalChunks int64
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			log.Infof("DST: Received %d chunks at DstFn with first/last bytes as:", chunkCount)
			log.Info(payloadBytes[0:9], payloadBytes[byteCount-9:byteCount])
			return payloadBytes[:byteCount], nil
		}
		if err != nil {
			log.Errorf("dQP: receive error: %v", err)
			return nil, err
		}
		log.Debugf("dQP: Received chunk no. %d", chunkCount)
		onlyOnce.Do(func() {
			totalChunks = chunk.TotalChunks
			log.Infof("DST: creating a new buffer")
			payloadBytes = make([]byte, totalChunks*int64(config.ChunkSizeInBytes))
			log.Infof("DST: chunkTotal = %d", totalChunks)
		})
		log.Debugf("DST: appending chunk number %d", chunkCount)
		copy(payloadBytes[byteCount:], chunk.Chunk)
		byteCount += len(chunk.Chunk)
		chunkCount += 1
	}
}

// BroadcastGet pulls payload from DQP server using the key
func BroadcastGet(ctx context.Context, capability string, config utils.Config) ([]byte, error) {
	log.Infof("attempting Get using capability %s", capability)
	key := capability
	splitString := strings.SplitN(capability, "|", 2)
	sQPAddr := splitString[1]
	httpMetadata := map[string]string{
		"is_xdt":   "true",
		"key":      key,
		"sqp_addr": sQPAddr,
		"routing":  config.Routing,
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.New(httpMetadata))
	conn, err := utils.GetGRPCConn(ctx, sQPAddr, true)
	if err != nil {
		log.Errorf("DST: can not connect with SQP server %v", err)
		return nil, err
	}

	client := crossXDT.NewStreamDataClient(conn)
	in := &crossXDT.BroadcastRequest{Key: key, ChunkSizeInBytes: int64(config.ChunkSizeInBytes)}
	stream, err := client.ServeBroadcastData(ctx, in)
	if err != nil {
		log.Errorf("dST: open stream error %v", err)
		return nil, err
	}

	chunkCount := 0
	byteCount := 0
	var onlyOnce sync.Once
	var payloadBytes []byte
	var totalChunks int64
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			log.Infof("DST: Received %d chunks at DstFn with first/last bytes as:", chunkCount)
			log.Info(payloadBytes[0:9], payloadBytes[byteCount-9:byteCount])
			return payloadBytes[:byteCount], nil
		}
		if err != nil {
			log.Errorf("DST: receive error: %v", err)
			return nil, err
		}
		log.Debugf("DST: Received chunk no. %d", chunkCount)
		onlyOnce.Do(func() {
			totalChunks = chunk.TotalChunks
			log.Infof("DST: creating a new buffer")
			payloadBytes = make([]byte, totalChunks*int64(config.ChunkSizeInBytes))
			log.Infof("DST: chunkTotal = %d", totalChunks)
		})
		log.Debugf("DST: appending chunk number %d", chunkCount)
		copy(payloadBytes[byteCount:], chunk.Chunk)
		byteCount += len(chunk.Chunk)
		chunkCount += 1
	}
}

// StartDstServer starts DstQP server
func StartDstServer(config utils.Config, handler func([]byte) ([]byte, bool)) {

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
