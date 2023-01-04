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
	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"context"
	"github.com/ease-lab/vhive-xdt/proto/crossXDT"
	"github.com/ease-lab/vhive-xdt/proto/downXDT"
	"google.golang.org/grpc/metadata"
	"net"
	"sync"

	"github.com/ease-lab/vhive-xdt/utils"

	log "github.com/sirupsen/logrus"
)

var packetMap = sync.Map{}

type DownXDTServer struct {
}

// XDTDataServe is a gRPC server to serve data to the DstFn
func (s DownXDTServer) XDTDataServe(ctx context.Context, req downXDT.XDTtoFn_xDTDataServe) error {

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
	log.Infof("dQP: data being fetched by DstFn using key : %s", key)

	var payloadData []byte
	// Check whether the first packet has been received at dQP or not
	for {
		if value, ok := packetMap.Load(key); ok {
			payloadData = value.([]byte)
			log.Infof("dQP: found chunkTotal %v for key %s at dQP", ok, key)
			break
		}
	}
	err = res.SetPayload(payloadData)
	if err != nil {
		log.Fatalf("[dQP]: error setting response %v", err)
		packetMap.Delete(key)
		return err
	}
	packetMap.Delete(key)
	return nil

	//for {
	//	select {
	//	case chunk := <-channel:
	//		resp := downXDT.Data{Chunk: chunk, TotalChunks: chunkTotal}
	//		if err := srv.Send(&resp); err != nil {
	//			log.Errorf("dQP: send error %v", err)
	//			log.Infof("[dQP] freeing channel %s due to send error", in.Key)
	//			bufferPool.FreeChannel(in.Key)
	//			return err
	//		}
	//		log.Debugf("dQP: Sending chunk : %d to DstFn", chunkCount)
	//		chunkCount += 1
	//	default:
	//		if chunkTotal == int64(chunkCount) {
	//			log.Infof("[dQP] transfer to DST complete, freeing channel %s ", in.Key)
	//			bufferPool.FreeChannel(in.Key)
	//			return nil
	//		}
	//	}
	//}

}

// PullDataFromSrcQP pulls data from src QP to dst QP
func PullDataFromSrcQP(ctx context.Context) error {

	headers, _ := metadata.FromOutgoingContext(ctx)
	key := headers["key"][0]
	sQPAddr := headers["sqp_addr"][0]
	//routing := headers["routing"][0]

	conn, err := net.Dial("tcp", sQPAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	capconn := rpc.NewConn(rpc.NewStreamTransport(conn), nil)
	client := crossXDT.StreamData(capconn.Bootstrap(ctx))

	// Okay! Let's make an RPC call!  Remember:  RPC is performed simply by
	// calling a's methods.
	//
	// There are a couple of interesting things to note here:
	//  1. We pass a callback function to set parameters on the RPC call.  If the
	//     call takes no arguments, you MAY pass nil.
	//  2. We return a Future type, representing the in-flight RPC call.  As with
	//     the earlier call to Bootstrap, a's methods do not block.  They instead
	//     return a future that eventually resolves with the RPC results. We also
	//     return a release function, which MUST be called when you're done with
	//     the RPC call and its results.
	// do release
	f, _ := client.ServeData(ctx, func(ps crossXDT.StreamData_serveData_Params) error {
		ps.SetKey(key)
		return nil
	})

	// You can do other things while the RPC call is in-flight.  Everything
	// is asynchronous. For simplicity, we're going to block until the call
	// completes.
	res, err := f.Struct()
	if err != nil {
		log.Fatal(err)
	}

	// Lastly, let's print the result.  Recall that 'product' is the name of
	// the return value that we defined in the schema file.
	payload, err := res.Payload()
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("[dqp]: storing payload for key %s", key)
	packetMap.Store(key, payload)
	log.Info("dqp: Received", payload[0:9], payload[len(payload)-9:])
	return nil
}

// StartServer starts DstQP server
func StartServer(config utils.Config) {

	lis, err := net.Listen("tcp", config.DQPServerPort)
	if err != nil {
		log.Fatalf("dQP: failed to listen: %v", err)
	}
	server := DownXDTServer{}
	for {
		client := downXDT.XDTtoFn_ServerToClient(&server)
		// accept connection
		conn, err := lis.Accept()
		if err != nil {
			log.Fatal(err)
		}
		capnpconn := rpc.NewConn(rpc.NewStreamTransport(conn), &rpc.Options{
			// The BootstrapClient is the RPC interface that will be made available
			// to the remote endpoint by default.  In this case, Arith.
			BootstrapClient: capnp.Client(client),
		})
		// Block until the connection terminates.
		select {
		case <-capnpconn.Done():
			capnpconn.Close()
			continue
		}
	}
}
