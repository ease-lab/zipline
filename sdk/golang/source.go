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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/atomic"

	tracing "github.com/ease-lab/vSwarm/utils/tracing/go"

	"google.golang.org/grpc/metadata"

	"github.com/ease-lab/vhive-xdt/proto/crossXDT"
	"github.com/ease-lab/vhive-xdt/proto/downXDT"

	"github.com/ease-lab/vhive-xdt/utils"

	log "github.com/sirupsen/logrus"

	"net"

	"github.com/ease-lab/vhive-xdt/proto/upXDT"
	"google.golang.org/grpc"
)

type crossXDTServer struct {
	payloadDataMap *sync.Map
	config         utils.Config
	crossXDT.UnimplementedStreamDataServer
}

type XDTclient struct {
	atom           atomic.Uint64
	config         utils.Config
	client         upXDT.StreamDataClient
	addr           string
	payloadDataMap sync.Map
	crossXDTserver crossXDTServer
}

func NewXDTclient(config utils.Config) (*XDTclient, error) {
	var xdtClient XDTclient
	xdtClient.config = config
	sQPAddr := config.SQPServerHostname + config.SQPServerPort
	conn, err := utils.GetGRPCConn(context.Background(), sQPAddr, false)
	if err != nil {
		log.Errorf("SRC: can not connect to SQP %v", err)
		return &xdtClient, err
	}
	xdtClient.client = upXDT.NewStreamDataClient(conn)
	xdtClient.atom.Store(0)
	xdtClient.addr = utils.FetchSelfIP() + config.SQPServerPort

	if config.NoCopy {
		xdtClient.addr = utils.FetchSelfIP() + config.SrcServerPort
		log.Infof("[src] starting the host server")
		xdtClient.crossXDTserver = crossXDTServer{payloadDataMap: &xdtClient.payloadDataMap, config: config}
		go xdtClient.crossXDTserver.ServeData(config.SrcServerPort)
	}

	return &xdtClient, nil
}

func (x *XDTclient) splitPayload(xdtPayload *utils.Payload) (string, []byte) {
	key := fmt.Sprintf("%s|%s", strconv.FormatUint(x.atom.Inc(), 10), x.addr)
	log.Infof("XDT invoke called with payload size %d", len(xdtPayload.Data))

	payloadData := xdtPayload.Data
	log.Info(payloadData[0:9], payloadData[len(payloadData)-9:])
	xdtPayload.Data = []byte("")
	return key, payloadData
}

func (x *XDTclient) serve(key string, payloadData []byte) string {

	x.payloadDataMap.Store(key, payloadData)

	return "bla"
}

// ServeData is the gRPC server to serve the available data to the dQP
func (s crossXDTServer) ServeData(serverPort string) error {

	lis, err := net.Listen("tcp", serverPort)
	if err != nil {
		log.Fatalf("src: failed to listen: %v", err)
	}

	for {
		// accept connection
		conn, err := lis.Accept()
		if err != nil {
			log.Fatalf("src: failed to accept connection: %v", err)
		}
		go func(conn net.Conn) {
			defer conn.Close()
			//Get the key
			reader := bufio.NewReader(conn)
			var err error
			key, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}
			key = strings.TrimSuffix(key, "\n")
			payloadDataInterface, ok := (*s.payloadDataMap).LoadAndDelete(key)
			if !ok {
				log.Errorf("[src] error getting payload from the map")
			}
			payloadData := payloadDataInterface.([]byte)
			log.Infof("src: dQP is fetching key: %s", key)
			_, err = conn.Write(payloadData)
			if err != nil {
				log.Errorf("[src] error sending bytes to dQP: %v", err)
			}
		}(conn)
	}
	return nil
}

// ServeBroadcastData is the gRPC server to serve the available data to the dQP
func (s crossXDTServer) ServeBroadcastData(in *crossXDT.BroadcastRequest, srv crossXDT.StreamData_ServeBroadcastDataServer) error {

	log.Infof("sQP: dQP is fetching key: %s", in.Key)

	payloadDataInterface, ok := (*s.payloadDataMap).Load(in.Key)
	if !ok {
		return nil
	}

	payloadData := payloadDataInterface.([]byte)
	log.Infof("src: dQP is fetching key: %s", in.Key)
	payloadSize := len(payloadData)
	log.Infof("Transfering %d bytes to dQP", payloadSize)
	chunkSizeInBytes := s.config.ChunkSizeInBytes
	chunkTotal := len(payloadData) / chunkSizeInBytes
	if len(payloadData)%chunkSizeInBytes != 0 {
		chunkTotal += 1
	}

	for currentByte := 0; currentByte < payloadSize; currentByte += chunkSizeInBytes {

		if currentByte+chunkSizeInBytes > payloadSize {
			resp := crossXDT.Response{Chunk: payloadData[currentByte:payloadSize], TotalChunks: int64(chunkTotal)}
			if err := srv.Send(&resp); err != nil {
				log.Errorf("send error %v", err)
				return err
			}
			log.Debugf("finishing request number : %d", currentByte)
		} else {
			resp := crossXDT.Response{Chunk: payloadData[currentByte : currentByte+chunkSizeInBytes], TotalChunks: int64(chunkTotal)}
			if err := srv.Send(&resp); err != nil {
				log.Errorf("send error %v", err)
				return err

			}
			log.Debugf("finishing request number : %d", currentByte)
		}

	}
	return nil
}

// Put uploads the data to sQP and returns key and sQP address
func (x *XDTclient) Put(ctx context.Context, payload []byte) (string, error) {

	key, _ := x.splitPayload(&utils.Payload{Data: payload})
	var payloadLocation string
	if x.config.NoCopy {
		payloadLocation = x.config.SrcServerHostname + x.config.SrcServerPort
		x.serve(key, payload)
	} else {
		payloadLocation = x.config.SQPServerHostname + x.config.SQPServerPort
	}

	httpMetadata := map[string]string{
		"is_xdt":                "true",
		"key":                   key,
		"sqp_addr":              payloadLocation,
		"routing":               x.config.Routing,
		"payload_size_in_bytes": strconv.Itoa(len(payload)),
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.New(httpMetadata))
	//  This timeout must be large enough for the request to complete
	timeoutDuration := time.Duration(x.config.RPCTimeoutDuration) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeoutDuration)
	defer cancel()

	span := tracing.Span{SpanName: "PutXDT", TracerName: "PutXDT-Tracer"}
	ctx = span.StartSpan(ctx)
	defer span.EndSpan()

	if !x.config.NoCopy {
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
	}
	return key, nil
}

// BroadcastPut uploads the data to sQP and returns key and sQP address
func (x *XDTclient) BroadcastPut(ctx context.Context, payload []byte) (string, error) {
	var payloadLocation string

	key, _ := x.splitPayload(&utils.Payload{Data: payload})
	if x.config.NoCopy {
		payloadLocation = x.config.SrcServerHostname + x.config.SrcServerPort
		x.serve(key, payload)
	} else {
		payloadLocation = x.config.SQPServerHostname + x.config.SQPServerPort
	}

	httpMetadata := map[string]string{
		"is_xdt":                "true",
		"key":                   key,
		"sqp_addr":              payloadLocation,
		"routing":               x.config.Routing,
		"payload_size_in_bytes": strconv.Itoa(len(payload)),
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.New(httpMetadata))
	//  This timeout must be large enough for the request to complete
	timeoutDuration := time.Duration(x.config.RPCTimeoutDuration) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeoutDuration)
	defer cancel()

	span := tracing.Span{SpanName: "PutXDT", TracerName: "PutXDT-Tracer"}
	ctx = span.StartSpan(ctx)
	defer span.EndSpan()

	if !x.config.NoCopy {
		errorPushData := make(chan error, 1)
		go func() { errorPushData <- x.PushBroadcastData(ctx, key, payload) }()

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
	}
	return key, nil
}

// ServeAndInvoke invokes the RPC call with and serves the object for DstQP to pull
func (x *XDTclient) ServeAndInvoke(ctx context.Context, URL string, xdtPayload utils.Payload) ([]byte, bool, error) {

	srcAddr := x.addr
	key, payloadData := x.splitPayload(&xdtPayload)
	serialisedPayload, err := json.Marshal(xdtPayload)
	if err != nil {
		return nil, false, err
	}
	httpMetadata := map[string]string{
		"is_xdt":                "true",
		"key":                   key,
		"sqp_addr":              srcAddr,
		"routing":               x.config.Routing,
		"payload_size_in_bytes": strconv.Itoa(len(payloadData)),
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.New(httpMetadata))
	//  This timeout must be large enough for the request to complete
	timeoutDuration := time.Duration(x.config.RPCTimeoutDuration) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeoutDuration)
	defer cancel()

	span := tracing.Span{SpanName: "InvokeWithXDT", TracerName: "InvokeWithXDT-Tracer"}
	ctx = span.StartSpan(ctx)
	defer span.EndSpan()

	x.serve(key, payloadData)

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
			log.Errorf("SDK: ServeAndInvokeWithXDT: fnInvocationCall failed: %v", err)
			return nil, false, err
		}
	}

	response := <-responseChannel
	return response.Message, response.Ok, nil

}

// Invoke invokes the RPC call with proper version
func (x *XDTclient) Invoke(ctx context.Context, URL string, xdtPayload utils.Payload) ([]byte, bool, error) {
	if x.config.NoCopy {
		return x.ServeAndInvoke(ctx, URL, xdtPayload)
	} else {
		return x.InvokeWithCopy(ctx, URL, xdtPayload)
	}
}

// InvokeWithCopy invokes the RPC call with XDT with copy
func (x *XDTclient) InvokeWithCopy(ctx context.Context, URL string, xdtPayload utils.Payload) ([]byte, bool, error) {

	sQPAddr := x.addr
	key, payloadData := x.splitPayload(&xdtPayload)
	serialisedPayload, err := json.Marshal(xdtPayload)
	if err != nil {
		return nil, false, err
	}

	httpMetadata := map[string]string{
		"is_xdt":                "true",
		"key":                   key,
		"sqp_addr":              sQPAddr,
		"routing":               x.config.Routing,
		"payload_size_in_bytes": strconv.Itoa(len(payloadData)),
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
func (x *XDTclient) PushData(ctx context.Context, key string, payload []byte) error {

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

// PushBroadcastData to source QP
func (x *XDTclient) PushBroadcastData(ctx context.Context, key string, payload []byte) error {

	payloadSize := len(payload)
	log.Infof("Transfering %d bytes to sQP", payloadSize)
	stream, err := x.client.BroadcastUpload(ctx)
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
