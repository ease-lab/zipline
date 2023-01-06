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
	"capnproto.org/go/capnp/v3/rpc"
	"context"
	"github.com/ease-lab/vhive-xdt/proto/crossXDT"
	"google.golang.org/grpc/metadata"
	"net"
	"strings"

	"github.com/ease-lab/vhive-xdt/proto/downXDT"
	"github.com/ease-lab/vhive-xdt/utils"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"google.golang.org/grpc"
)

type downXDTServer struct {
	config  utils.Config
	handler func([]byte) ([]byte, bool)
	conn    *rpc.Conn
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
		payloadBytes, err := FetchData(ctx, key, s.conn)
		if err != nil {
			log.Errorf("DST: FetchData failed %v", err)
			return &downXDT.InvocationResponse{}, err
		}

		//call destination function
		message, ok = s.handler(payloadBytes)
	}

	return &downXDT.InvocationResponse{Message: message, Ok: ok}, nil
}

// FetchData fetches data from dQP to DstFn
func FetchData(ctx context.Context, key string, capconn *rpc.Conn) ([]byte, error) {
	client := crossXDT.StreamData(capconn.Bootstrap(context.Background()))
	f, _ := client.ServeData(ctx, func(ps crossXDT.StreamData_serveData_Params) error {
		ps.SetKey(key)
		return nil
	})
	//defer release()

	res, err := f.Struct()
	if err != nil {
		log.Fatal(err)
	}
	payload, err := res.Payload()
	if err != nil {
		log.Fatal(err)
	}

	log.Info("DST: ", payload[0:9], payload[len(payload)-9:])
	return payload, nil
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

	conn, err := net.Dial("tcp", sQPAddr)
	if err != nil {
		log.Fatal(err)
	}
	capconn := rpc.NewConn(rpc.NewStreamTransport(conn), nil)
	return FetchData(ctx, key, capconn)
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

	conn, err := net.Dial("tcp", sQPAddr)
	if err != nil {
		log.Fatal(err)
	}
	capconn := rpc.NewConn(rpc.NewStreamTransport(conn), nil)
	client := crossXDT.StreamData(capconn.Bootstrap(context.Background()))
	f, _ := client.ServeBroadcastData(ctx, func(ps crossXDT.StreamData_serveBroadcastData_Params) error {
		ps.SetKey(key)
		return nil
	})
	//defer release()

	res, err := f.Struct()
	if err != nil {
		log.Fatal(err)
	}
	payload, err := res.Payload()
	if err != nil {
		log.Fatal(err)
	}

	log.Info("DST: ", payload[0:9], payload[len(payload)-9:])
	return payload, nil
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

	conn, err := net.Dial("tcp", config.DQPServerHostname+config.DQPServerPort)
	if err != nil {
		log.Fatal(err)
	}
	capconn := rpc.NewConn(rpc.NewStreamTransport(conn), nil)

	s := downXDTServer{}
	s.config = config
	s.handler = handler
	s.conn = capconn
	downXDT.RegisterXDTtoFnServer(server, s)

	log.Infoln("DST: start server")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("DST: failed to serve: %v", err)
	}
}
