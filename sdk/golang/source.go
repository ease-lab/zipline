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
	"fmt"
	"strconv"
	"time"

	tracing "github.com/ease-lab/vhive/utils/tracing/go"

	"google.golang.org/grpc/metadata"

	"github.com/ease-lab/vhive-xdt/proto/downXDT"

	"github.com/ease-lab/vhive-xdt/utils"

	log "github.com/sirupsen/logrus"

	"github.com/ease-lab/vhive-xdt/proto/upXDT"
	"google.golang.org/grpc"
)

type XDTclient struct {
	config utils.Config
	client upXDT.StreamDataClient
}

func NewXDTclient(config utils.Config) (XDTclient, error) {
	var xdtClient XDTclient
	xdtClient.config = config
	sQPAddr := config.SQPServerHostname + config.SQPServerPort
	conn, err := utils.GetGRPCConn(context.Background(), sQPAddr, false)
	if err != nil {
		log.Errorf("SRC: can not connect to SQP %v", err)
		return xdtClient, err
	}

	xdtClient.client = upXDT.NewStreamDataClient(conn)
	return xdtClient, nil
}

func splitPayload(xdtPayload *utils.Payload) (string, []byte) {
	now := time.Now()
	key := strconv.Itoa(int(now.UnixNano()))
	log.Infof("XDT invoke called with payload size %d", len(xdtPayload.Data))

	payloadData := xdtPayload.Data
	log.Info(payloadData[0:9], payloadData[len(payloadData)-9:])
	xdtPayload.Data = []byte("")
	return key, payloadData
}

// Put uploads the data to sQP and returns key and sQP address
func (x XDTclient) Put(ctx context.Context, payload []byte) (string, error) {
	sQPAddr := x.config.SQPServerHostname + x.config.SQPServerPort
	key, _ := splitPayload(&utils.Payload{Data: payload})

	httpMetadata := map[string]string{
		"is_xdt":   "true",
		"key":      key,
		"sqp_addr": sQPAddr,
		"routing":  x.config.Routing,
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.New(httpMetadata))
	//  This timeout must be large enough for the request to complete
	timeoutDuration := time.Duration(x.config.RPCTimeoutDuration) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeoutDuration)
	defer cancel()

	span := tracing.Span{SpanName: "PutXDT", TracerName: "PutXDT-Tracer"}
	ctx = span.StartSpan(ctx)
	defer span.EndSpan()

	errorPushData := make(chan error, 1)
	go func() { errorPushData <- x.PushData(ctx, key, payload) }()

	select {
	case <-ctx.Done():
		<-errorPushData // Wait for f to return.
		return "", ctx.Err()
	case err := <-errorPushData:
		if err != nil {
			log.Errorf("SDK: [Store & Forward] Push data failed")
			return "", err
		}
	}
	return fmt.Sprintf("%s|%s", key, sQPAddr), nil
}

// Invoke invokes the RPC call with XDT
func (x XDTclient) Invoke(ctx context.Context, URL string, xdtPayload utils.Payload) ([]byte, bool, error) {

	sQPAddr := x.config.SQPServerHostname + x.config.SQPServerPort
	key, payloadData := splitPayload(&xdtPayload)
	serialisedPayload, err := json.Marshal(xdtPayload)
	if err != nil {
		return nil, false, err
	}

	httpMetadata := map[string]string{
		"is_xdt":   "true",
		"key":      key,
		"sqp_addr": sQPAddr,
		"routing":  x.config.Routing,
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.New(httpMetadata))
	//  This timeout must be large enough for the request to complete
	timeoutDuration := time.Duration(x.config.RPCTimeoutDuration) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeoutDuration)
	defer cancel()

	span := tracing.Span{SpanName: "InvokeWithXDT", TracerName: "InvokeWithXDT-Tracer"}
	ctx = span.StartSpan(ctx)
	defer span.EndSpan()

	errorPushData := make(chan error, 1)
	go func() { errorPushData <- x.PushData(ctx, key, payloadData) }()
	if x.config.Routing == utils.STORE_FORWARD {
		log.Info("SDK: using store & forward routing")
		select {
		case <-ctx.Done():
			<-errorPushData // Wait for f to return.
			return nil, false, ctx.Err()
		case err := <-errorPushData:
			if err != nil {
				log.Errorf("SDK: [Store & Forward] Push data failed")
				return nil, false, err
			}
		}
	}

	errorFnInvocationCall := make(chan error, 1)
	responseChannel := make(chan *downXDT.InvocationResponse, 1)
	go func() {
		response, err := fnInvocationCall(ctx, URL, serialisedPayload)
		errorFnInvocationCall <- err
		responseChannel <- response
	}()
	select {
	case <-ctx.Done():
		<-errorFnInvocationCall
		return nil, false, ctx.Err()
	case err := <-errorFnInvocationCall:
		if err != nil {
			log.Errorf("SDK: InvokeWithXDT: fnInvocationCall failed: %v", err)
			return nil, false, err
		}
	}

	if x.config.Routing == utils.CUT_THROUGH {
		log.Info("SDK: using cut through routing")
		// Wait for completion and return the first error (if any)
		select {
		case <-ctx.Done():
			<-errorPushData
			return nil, false, ctx.Err()
		case err := <-errorPushData:
			if err != nil {
				log.Errorf("SDK: [Cut Through] Push data failed")
				return nil, false, err
			}
			response := <-responseChannel
			return response.Message, response.Ok, nil
		}
	} else {
		response := <-responseChannel
		return response.Message, response.Ok, nil
	}
}

// fnInvocationCall makes fn invocation call to dQP with xdt payload
func fnInvocationCall(ctx context.Context, URL string, serialisedPayload []byte) (*downXDT.InvocationResponse, error) {

	errorChannel := make(chan error, 1)
	responseChannel := make(chan *downXDT.InvocationResponse, 1)

	go func() {
		conn, err := grpc.DialContext(ctx, URL, utils.GetGopts()...)
		if err != nil {
			errorChannel <- err
			return
		}
		c := downXDT.NewXDTtoFnClient(conn)
		log.Infof("SDK: Fn invocation start")
		response, err := c.XDTFnCall(ctx, &downXDT.InvocationRequest{XDTJSON: serialisedPayload})
		if err != nil {
			errorChannel <- err
			return
		}
		// need some help in closing this connection in case of an error.
		err = conn.Close()
		if err != nil {
			log.Errorf("dQP: Error closing the connection to Dest")
			errorChannel <- err
			return
		}
		errorChannel <- nil
		responseChannel <- response
	}()
	select {
	case <-ctx.Done():
		log.Errorf("SDK: context expired at fnInvocationCall")
		<-errorChannel
		return nil, ctx.Err()
	case err := <-errorChannel:
		if err != nil {
			log.Errorf("SDK: Fn invocation failed: %v", err)
			return nil, err
		}
		log.Infof("SDK: Fn invocation successful")
		return <-responseChannel, nil
	}
}

// PushData to source QP
func (x XDTclient) PushData(ctx context.Context, key string, payload []byte) error {

	payloadSize := len(payload)
	log.Infof("Transfering %d bytes to sQP", payloadSize)
	stream, err := x.client.SendData(ctx)
	if err != nil {
		log.Errorf("open stream error %v", err)
		return err
	}
	chunkSizeInBytes := x.config.ChunkSizeInBytes
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
	log.Infof("SDK: data push successful")
	return nil
}
