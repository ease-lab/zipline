package dqp

import (
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"

	crossXDT "github.com/ease-lab/vhive_stealth/examples/prototype/proto/crossXDT"
	downXDT "github.com/ease-lab/vhive_stealth/examples/prototype/proto/downXDT"
	fnInvocation "github.com/ease-lab/vhive_stealth/examples/prototype/proto/fnInvocation"
	sdk "github.com/ease-lab/vhive_stealth/examples/prototype/sdk"

	"google.golang.org/grpc"
)

var dataQueue sync.Map
var dataQueueSize sync.Map

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



// gRPC server to serve data to the DstFn
func (s downXDTServer) XDTDataServe(in *downXDT.DataRequest, srv downXDT.XDTtoFn_XDTDataServeServer) error {

	log.Infof("dQP: data being fetched by DstFn using key : %s", in.Key)

	chunkCount := 0
	var channel chan []byte
	var chunkTotal int64
	for {
		log.Tracef("dQP: finding channel for key %s",in.Key)
		if tmp,ok := dataQueue.Load(in.Key); ok {
			log.Tracef("dQP: found channel for key %s",in.Key)
			channel = tmp.(chan []byte)
			break
		}
	}
	for {
		if tmp,ok := dataQueueSize.Load(in.Key); ok {
			chunkTotal = tmp.(int64)
			log.Tracef("dQP: found chunkTotal %d for key %s",chunkTotal,in.Key)
			break
		}
	}

	for {
		select {
		case chunk := <-channel:
			resp := downXDT.Data{Chunk:chunk }
			if err := srv.Send(&resp); err != nil {
				log.Fatalf("dQP: send error %v", err)
			}
			log.Tracef("dQP: Sending chunk : %d to DstFn", chunkCount)
			chunkCount+=1
		default:
			if chunkTotal == int64(chunkCount) {
				dataQueue.Delete(in.Key)
				dataQueueSize.Delete(in.Key)
				close(channel)
				return nil
			}
		}
	}
	return nil
}

// gRPC server to route the function call from SrcFn to the DstFn
func (s fnInvocationServer) RouteInvocation(ctx context.Context, in *fnInvocation.InvocationRequest) (*fnInvocation.Empty, error) {

	log.Infof("dQP: received serialised json: %s", in.XDTJSON)

	var xdtPayload payload
	if err := json.Unmarshal(in.XDTJSON, &xdtPayload); err != nil {
		log.Error(err)
	}

	log.Infof("dQP: fetching data from sQP using key : %s", xdtPayload.Key)
	chunkSizeInBytes := sdk.LoadedConfig.ChunkSizeInBytes

	if sdk.LoadedConfig.Routing == "CT" {
		log.Infof("dQP: CT: pulling data from sQP")
		go PullDataFromSrcQP(xdtPayload.Key, chunkSizeInBytes)
	}else if sdk.LoadedConfig.Routing == "S&F"{
		log.Infof("dQP: S&F: pulling data from sQP")
		PullDataFromSrcQP(xdtPayload.Key, chunkSizeInBytes)
	}

	// route the invocation call to destination fn
	serverAddr := sdk.LoadedConfig.DstServerAddr
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Errorf("did not connect: %v", err)
	}
	defer conn.Close()

	c := downXDT.NewXDTtoFnClient(conn)
	_, err = c.XDTFnCall(context.Background(), &downXDT.InvocationRequest{XdtJson: in.XDTJSON})
	if err != nil {
		log.Infof("dQP: Fn invocation route unsuccessful")
	}
	log.Infof("dQP: Fn invocation route successful")
	return &fnInvocation.Empty{}, nil
}

// pull data from src QP to dst QP
func PullDataFromSrcQP(key string, chunkSizeInBytes int) {

	serverAddr := sdk.LoadedConfig.SQPServerAddr
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Errorf("dQP: can not connect with server %v", err)
	}

	//ctx,cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	ctx := context.Background()
	client := crossXDT.NewStreamDataClient(conn)
	in := &crossXDT.Request{Key: key, ChunkSize: int64(chunkSizeInBytes)}
	stream, err := client.ServeData(ctx, in)
	if err != nil {
		log.Errorf("dQP: open stream error %v", err)
	}

	chunkCount := 0
	var channel chan []byte
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			log.Tracef("dQP: %d chunks received at Dst",chunkCount)
			//log.Trace(dataQueue[key+";0"][0:9],dataQueue[key+";"+strconv.Itoa(chunkCount-1)][len(dataQueue[key+";"+strconv.Itoa(chunkCount-1)])-9:])
			if sdk.LoadedConfig.Routing == "S&F" {
				dataQueue.Store(key, channel)
			}
			return
		}
		if err != nil {
			log.Errorf("dQP: receive error: %v", err)
		}
		log.Tracef("dQP: Received chunk no. %d", chunkCount)
		if _,ok := dataQueueSize.Load(key); !ok {
			log.Infof("dQP: creating a new channel")
			if sdk.LoadedConfig.Routing == "CT" {
				channel = make(chan []byte, sdk.LoadedConfig.BufferSize)
				dataQueue.Store(key, channel)
			}else if sdk.LoadedConfig.Routing == "S&F" {
				channel = make(chan []byte, sdk.LoadedConfig.StAndFwBufferSize)
			}else {
				log.Errorf("dQP: Invalid route type. Check config.json")
			}
			log.Infof("dQP: chunkTotal = %d",chunk.ChunkTotal)
			dataQueueSize.Store(key,chunk.ChunkTotal)
		}
		log.Infof("dQP: Enquing chunk number %d",chunkCount)
		channel <- chunk.Chunk
		chunkCount += 1
	}
}

// start DstQP server
func StartServer(serverAddr string) {

	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("dQP: failed to listen: %v", err)
	}

	server := grpc.NewServer()
	downXDT.RegisterXDTtoFnServer(server, downXDTServer{})
	fnInvocation.RegisterInvocationServer(server, fnInvocationServer{})

	log.Println("dQP: start server")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("dQP: failed to serve: %v", err)
	}

}
