package dqp

import (
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"net"

	crossXDT "github.com/ease-lab/vhive_stealth/examples/prototype/proto/crossXDT"
	downXDT "github.com/ease-lab/vhive_stealth/examples/prototype/proto/downXDT"
	fnInvocation "github.com/ease-lab/vhive_stealth/examples/prototype/proto/fnInvocation"
	sdk "github.com/ease-lab/vhive_stealth/examples/prototype/sdk"

	"google.golang.org/grpc"
)

var dataQueue = make(map[string]chan []byte)
var dataQueueSize = make(map[string]int)

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
	var channel chan []byte
	var ok bool
	for {
		if channel,ok = dataQueue[in.Key]; ok {
			break
		}
	}

	for {
		select {
		case chunk := <-channel:
			resp := downXDT.Data{Chunk:chunk }
			if err := srv.Send(&resp); err != nil {
				log.Fatalf("send error %v", err)
			}
			log.Tracef("Sending chunk : %d to DstFn", chunkCount)
			chunkCount+=1
		default:
			if totalChunks, ok := dataQueueSize[in.Key]; ok && totalChunks == chunkCount {
				return nil
			}
		}
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

	go PullDataFromSrcQP(xdtPayload.Key, chunkSizeInBytes)

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
func PullDataFromSrcQP(key string, chunkSizeInBytes int) {

	serverAddr := config.SQPServerAddr
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Errorf("can not connect with server %v", err)
	}

	//ctx,cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	ctx := context.Background()
	client := crossXDT.NewStreamDataClient(conn)
	in := &crossXDT.Request{Key: key, ChunkSize: int64(chunkSizeInBytes)}
	stream, err := client.ServeData(ctx, in)
	if err != nil {
		log.Errorf("open stream error %v", err)
	}

	chunkCount := 0
	var channel chan []byte
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			log.Tracef("%d chunks received at Dst",chunkCount)
			dataQueueSize[key] = chunkCount
			//log.Trace(dataQueue[key+";0"][0:9],dataQueue[key+";"+strconv.Itoa(chunkCount-1)][len(dataQueue[key+";"+strconv.Itoa(chunkCount-1)])-9:])
			return
		}
		if err != nil {
			log.Errorf("receive error: %v", err)
		}
		log.Tracef("Received chunk no. %d at dQP", chunkCount)
		if _,ok := dataQueue[key]; !ok {
			log.Infof("creating a new channel at dQP")
			channel = make(chan []byte, config.ChannelBufferSize)
			dataQueue[key] = channel
		}
		log.Infof("Enquing chunk number %d at dQP",chunkCount)
		channel <- chunk.Chunk
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
