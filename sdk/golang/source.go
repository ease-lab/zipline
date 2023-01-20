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
	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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
	xdtClient.atom.Store(0)
	//start server for serving payloads
	xdtClient.addr = utils.FetchSelfIP() + config.SrcServerPort
	log.Infof("[src] starting the host server")
	lis, err := net.Listen("tcp", config.SrcServerPort)
	if err != nil {
		log.Fatalf("src: failed to listen: %v", err)
	}
	xdtClient.crossXDTserver = crossXDTServer{payloadDataMap: &xdtClient.payloadDataMap, config: config}
	go func() {
		for {
			client := crossXDT.StreamData_ServerToClient(xdtClient.crossXDTserver)

			conn, err := lis.Accept()
			if err != nil {
				log.Fatal(err)
			}
			go func() {
				capnpconn := rpc.NewConn(rpc.NewStreamTransport(conn), &rpc.Options{
					// The BootstrapClient is the RPC interface that will be made available
					// to the remote endpoint by default.  In this case, Arith.
					BootstrapClient: capnp.Client(client),
				})
				// Block until the connection terminates.
				select {
				case <-capnpconn.Done():
					capnpconn.Close()
					return
				}
			}()
		}
	}()

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
func (s crossXDTServer) ServeData(ctx context.Context, req crossXDT.StreamData_serveData) error {

	res, err := req.AllocResults() // allocate the results struct
	if err != nil {
		log.Fatalf("[src]: error allocating response %v", err)
		return err
	}
	key, err := req.Args().Key()
	if err != nil {
		log.Fatalf("[src]: error getting key %v", err)
		return err
	}

	payloadDataInterface, ok := (*s.payloadDataMap).LoadAndDelete(key)
	if !ok {
		return nil
	}
	payloadData := payloadDataInterface.([]byte)
	log.Infof("src: dQP is fetching key: %s", key)
	err = res.SetPayload(payloadData)
	if err != nil {
		log.Fatalf("[src]: error setting response %v", err)
		return err
	}
	return nil
}

// ServeBroadcastData is the gRPC server to serve the available data to the dQP
func (s crossXDTServer) ServeBroadcastData(ctx context.Context, req crossXDT.StreamData_serveBroadcastData) error {

	res, err := req.AllocResults() // allocate the results struct
	if err != nil {
		log.Fatalf("[src]: error allocating response %v", err)
		return err
	}
	key, err := req.Args().Key()
	if err != nil {
		log.Fatalf("[src]: error getting key %v", err)
		return err
	}

	payloadDataInterface, ok := (*s.payloadDataMap).Load(key)
	if !ok {
		return nil
	}
	payloadData := payloadDataInterface.([]byte)
	log.Infof("src: dQP is fetching key: %s", key)
	err = res.SetPayload(payloadData)
	if err != nil {
		log.Fatalf("[src]: error setting response %v", err)
		return err
	}
	return nil
}

// Put adds the src server and returns key and src address
func (x *XDTclient) Put(ctx context.Context, payload []byte) (string, error) {
	key, _ := x.splitPayload(&utils.Payload{Data: payload})
	x.serve(key, payload)
	return key, nil
}

// BroadcastPut adds the src server and returns key and src address
func (x *XDTclient) BroadcastPut(ctx context.Context, payload []byte) (string, error) {
	return x.Put(ctx, payload)
}

// Invoke invokes the RPC call with proper version
func (x *XDTclient) Invoke(ctx context.Context, URL string, xdtPayload utils.Payload) ([]byte, bool, error) {
	srcAddr := x.addr
	key, payloadData := x.splitPayload(&xdtPayload)
	serialisedPayload, err := json.Marshal(xdtPayload)
	if err != nil {
		return nil, false, err
	}

	httpMetadata := map[string]string{
		"is_xdt":   "true",
		"key":      key,
		"src_addr": srcAddr,
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
