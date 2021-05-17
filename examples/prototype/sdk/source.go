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
	"google.golang.org/grpc/metadata"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ease-lab/vhive_stealth/examples/prototype/proto/fnInvocation"
	"github.com/ease-lab/vhive_stealth/examples/prototype/proto/upXDT"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"google.golang.org/grpc"
)

// InvokeWithXDT invokes the RPC call with XDT
func InvokeWithXDT(URL string, xdtPayload Payload, chunkSizeInBytes int) {

	if LoadedConfig.TracingEnabled {
		shutdown := InitTracer()
		defer shutdown()
	}

	log.Infof("SDK: XDT invoke start")
	now := time.Now()
	key := strconv.Itoa(int(now.UnixNano()))
	log.Infof("XDT invoke called with payload size %d", len(xdtPayload.Data))

	md := metadata.Pairs(
		"timestamp", time.Now().Format(time.StampNano),
		"transaction-id", key,
	)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	payloadData := xdtPayload.Data
	log.Info(payloadData[0:9],payloadData[len(payloadData)-9:])
	xdtPayload.Data = []byte("")
	xdtPayload.Key = key
	xdtPayload.IsXDT = true

	serialisedPayload, _ := json.Marshal(xdtPayload)

	if LoadedConfig.Routing == "S&F" {
		log.Info("SDK: using store & forward routing")
		PushData(ctx, key, payloadData, chunkSizeInBytes)
	}else if  LoadedConfig.Routing == "CT" {
		log.Info("SDK: using cut through routing")
		go PushData(ctx, key, payloadData, chunkSizeInBytes)
	}

	fnInvocationCall(ctx, URL, serialisedPayload)
}

// fnInvocationCall makes fn invocation call to dQP with xdt payload
func fnInvocationCall(ctx context.Context, URL string, serialisedPayload []byte) {

	conn, err := grpc.Dial(URL, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer func (){
		err = conn.Close()
		if err != nil {
			log.Errorf("dQP: Error closing the connection to Dest")
		}
	}()

	c := fnInvocation.NewInvocationClient(conn)

	log.Infof("SDK: Fn invocation start")
	_, err = c.RouteInvocation(ctx, &fnInvocation.InvocationRequest{XDTJSON: serialisedPayload})
	if err == nil {
		log.Infof("SDK: Fn invocation successful")
	}
}

// PushData to source QP
func PushData(ctx context.Context, key string, payload []byte, chunkSizeInBytes int) {

	serverAddr := LoadedConfig.SQPServerAddr
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()))
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}

	client := upXDT.NewStreamDataClient(conn)
	payloadSize := len(payload)
	log.Infof("Transfering %d bytes to sQP", payloadSize)
	stream, err := client.SendData(ctx)
	if err != nil {
		log.Fatalf("open stream error %v", err)
	}

	chunkTotal := len(payload)/chunkSizeInBytes
	if len(payload)%chunkSizeInBytes!=0 {
		chunkTotal+=1
	}

	for currentByte := 0; currentByte < payloadSize; currentByte += chunkSizeInBytes {

		if currentByte+chunkSizeInBytes > payloadSize {
			req := upXDT.Request{Chunk: payload[currentByte:payloadSize], Key: key, TotalChunks: int64(chunkTotal)}
			if err := stream.Send(&req); err != nil {
				log.Fatalf("send error %v", err)
			}
			log.Tracef("finishing request number : %d", currentByte)
		} else {
			req := upXDT.Request{Chunk: payload[currentByte : currentByte+chunkSizeInBytes], Key: key, TotalChunks: int64(chunkTotal)}
			if err := stream.Send(&req); err != nil {
				log.Fatalf("send error %v", err)
			}
			log.Tracef("finishing request number : %d", currentByte)
		}

	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
}
