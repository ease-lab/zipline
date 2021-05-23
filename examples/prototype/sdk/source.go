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
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/metadata"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"XDTprototype/proto/fnInvocation"
	"XDTprototype/proto/upXDT"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"google.golang.org/grpc"
)

// InvokeWithXDT invokes the RPC call with XDT
func InvokeWithXDT(URL string, xdtPayload Payload, chunkSizeInBytes int) error {

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
	log.Info(payloadData[0:9], payloadData[len(payloadData)-9:])
	xdtPayload.Data = []byte("")
	xdtPayload.Key = key
	xdtPayload.IsXDT = true

	serialisedPayload, err := json.Marshal(xdtPayload)
	if err != nil {
		return err
	}

	errGroup, _ := errgroup.WithContext(ctx)

	if LoadedConfig.Routing == STORE_FORWARD {
		log.Info("SDK: using store & forward routing")
		err := PushData(ctx, key, payloadData, chunkSizeInBytes)
		if err != nil {
			log.Errorf("SDK: [Store & Forward] Push data failed")
			return err
		}
	} else if LoadedConfig.Routing == CUT_THROUGH {
		log.Info("SDK: using cut through routing")
		errGroup.Go(func() error {
			err := PushData(ctx, key, payloadData, chunkSizeInBytes)
			if err != nil {
				log.Errorf("SDK: [cut-through] Push data failed")
				return err
			}
			return nil
		})
	}

	if err := fnInvocationCall(ctx, URL, serialisedPayload); err != nil {
		log.Errorf("SDK: InvokeWithXDT: fnInvocationCall failed")
		return err
	}
	// Wait for completion and return the first error (if any)
	return errGroup.Wait()
}

// fnInvocationCall makes fn invocation call to dQP with xdt payload
func fnInvocationCall(ctx context.Context, URL string, serialisedPayload []byte) error {

	conn, err := grpc.Dial(URL, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()))
	if err != nil {
		log.Errorf("did not connect: %v", err)
		return err
	}
	errGroup, _ := errgroup.WithContext(ctx)

	c := fnInvocation.NewInvocationClient(conn)

	log.Infof("SDK: Fn invocation start")
	_, err = c.RouteInvocation(ctx, &fnInvocation.InvocationRequest{XDTJSON: serialisedPayload})
	if err != nil {
		log.Errorf("SDK: Fn invocation failed")
		return err
	}
	log.Infof("SDK: Fn invocation successful")

	errGroup.Go(func() error {
		err = conn.Close()
		if err != nil {
			log.Errorf("dQP: Error closing the connection to Dest")
			return err
		}
		return nil
	})
	return errGroup.Wait()
}

// PushData to source QP
func PushData(ctx context.Context, key string, payload []byte, chunkSizeInBytes int) error {

	serverAddr := LoadedConfig.SQPServerAddr
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()))
	if err != nil {
		log.Errorf("can not connect with server %v", err)
		return err
	}

	client := upXDT.NewStreamDataClient(conn)
	payloadSize := len(payload)
	log.Infof("Transfering %d bytes to sQP", payloadSize)
	stream, err := client.SendData(ctx)
	if err != nil {
		log.Errorf("open stream error %v", err)
		return err
	}

	chunkTotal := len(payload) / chunkSizeInBytes
	if len(payload)%chunkSizeInBytes != 0 {
		chunkTotal += 1
	}

	for currentByte := 0; currentByte < payloadSize; currentByte += chunkSizeInBytes {

		if currentByte+chunkSizeInBytes > payloadSize {
			req := upXDT.Request{Chunk: payload[currentByte:payloadSize], Key: key, TotalChunks: int64(chunkTotal)}
			if err := stream.Send(&req); err != nil {
				log.Errorf("send error %v", err)
				return err
			}
			log.Debugf("finishing request number : %d", currentByte)
		} else {
			req := upXDT.Request{Chunk: payload[currentByte : currentByte+chunkSizeInBytes], Key: key, TotalChunks: int64(chunkTotal)}
			if err := stream.Send(&req); err != nil {
				log.Errorf("send error %v", err)
				return err

			}
			log.Debugf("finishing request number : %d", currentByte)
		}

	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		log.Errorf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
		return err
	}
	return nil
}
