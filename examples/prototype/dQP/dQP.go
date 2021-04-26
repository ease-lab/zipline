package dqp

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	crossXDT "github.com/ease-lab/vhive_stealth/examples/prototype/proto/crossXDT"
	downXDT "github.com/ease-lab/vhive_stealth/examples/prototype/proto/downXDT"
	fnInvocation "github.com/ease-lab/vhive_stealth/examples/prototype/proto/fnInvocation"
	sdk "github.com/ease-lab/vhive_stealth/examples/prototype/sdk"

	"google.golang.org/grpc"
)

var dataQueue = make(map[string][]byte)

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

var config = sdk.LoadConfig("../config.json")

// gRPC server to serve data to the DstFn
func (s downXDTServer) XDTDataServe(in *downXDT.DataRequest, srv downXDT.XDTtoFn_XDTDataServeServer) error {

	log.Infof("fetching from dQP using key : %s", in.Key)

	chunkCount := 0

	for {
		chunk, ok := dataQueue[in.Key+";"+strconv.Itoa(chunkCount)]
		if !ok {
			break
		}
		resp := downXDT.Data{Chunk:chunk }
		if err := srv.Send(&resp); err != nil {
			log.Fatalf("send error %v", err)
		}
		log.Tracef("finishing request number : %d", chunkCount)
		chunkCount+=1
	}

	return nil
}

// gRPC server to route the function call from SrcFn to the DstFn
func (s fnInvocationServer) RouteInvocation(ctx context.Context, in *fnInvocation.InvocationRequest) (*fnInvocation.Empty, error) {

	log.Infof("received serialised json at dQP: %s", in.XdtJson)

	var xdtPayload payload
	if err := json.Unmarshal(in.XdtJson, &xdtPayload); err != nil {
		log.Error(err)
	}

	log.Infof("fetching data from sQP using key : %s", xdtPayload.Key)
	chunkSizeInBytes := config.ChunkSizeInBytes
	duration, payloadCount := PullDataFromSrcQP(xdtPayload.Key, chunkSizeInBytes)
	log.Infof("pulled %d chunks from sQP in %s",payloadCount, duration)

	// route the invocation call to destination fn
	serverAddr := config.DstServerAddr
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Errorf("did not connect: %v", err)
	}
	defer conn.Close()

	c := downXDT.NewXDTtoFnClient(conn)
	_, err = c.XDTFnCall(context.Background(), &downXDT.InvocationRequest{XdtJson: in.XdtJson})
	if err == nil {
		log.Infof("Fn invocation route at dQP successful")
	}

	return &fnInvocation.Empty{}, nil
}

// pull data from src QP to dst QP
func PullDataFromSrcQP(key string, chunkSizeInBytes int) (time.Duration, int) {

	serverAddr := config.SQPServerAddr
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Errorf("can not connect with server %v", err)
	}
	start := time.Now()

	client := crossXDT.NewStreamDataClient(conn)
	in := &crossXDT.Request{Key: key, ChunkSize: int64(chunkSizeInBytes)}
	stream, err := client.ServeData(context.Background(), in)
	if err != nil {
		log.Errorf("open stream error %v", err)
	}

	chunkCount := 0
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			elapsed := time.Since(start)
			log.Tracef("Complete packet received")
			log.Trace(dataQueue[key+";0"][0:9],dataQueue[key+";"+strconv.Itoa(chunkCount-1)][len(dataQueue[key+";"+strconv.Itoa(chunkCount-1)])-9:])
			//dataQueue[key] = payload
			return elapsed, chunkCount
		}
		if err != nil {
			log.Errorf("receive error: %v", err)
		}
		log.Tracef("Received chunk no. %d", chunkCount)
		dataQueue[key+";"+strconv.Itoa(chunkCount)] = chunk.Chunk
		chunkCount += 1
	}
}

// start DstQP server
func StartServer(serverAddr string) {

	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	downXDT.RegisterXDTtoFnServer(server, downXDTServer{})
	fnInvocation.RegisterInvocationServer(server, fnInvocationServer{})

	log.Println("start server")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
