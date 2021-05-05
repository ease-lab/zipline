package sdk

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	downXDT "github.com/ease-lab/vhive_stealth/examples/prototype/proto/downXDT"

	"google.golang.org/grpc"
)

var dataQueue sync.Map
var dataQueueSize sync.Map

// to be called by dQP to invoke DstFn
func (s downXDTServer) XDTFnCall(ctx context.Context, in *downXDT.InvocationRequest) (*downXDT.Empty, error) {

	log.Infof("DST: received invocation call %s", in.XdtJson)

	var xdtPayload Payload
	if err := json.Unmarshal(in.XdtJson, &xdtPayload); err != nil {
		log.Fatal(err)
	}

	key := xdtPayload.Key

	chunkSizeInBytes := LoadedConfig.ChunkSizeInBytes

	// fetch data from dQP
	FetchFromDQP(key, chunkSizeInBytes)
	return &downXDT.Empty{}, nil
}

// fetch data from dQP to DstFn
func FetchFromDQP(key string, chunkSizeInBytes int) (time.Duration, int) {
	serverAddr := LoadedConfig.DQPServerAddr
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("DST: can not connect with server %v", err)
	}
	start := time.Now()

	client := downXDT.NewXDTtoFnClient(conn)
	in := &downXDT.DataRequest{Key: key, ChunkSize: int64(chunkSizeInBytes)}
	stream, err := client.XDTDataServe(context.Background(), in)
	if err != nil {
		log.Fatalf("DST: open stream error %v", err)
	}

	chunkCount := 0
	var channel chan []byte
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			elapsed := time.Since(start)
			log.Infof("DST: Received %d chunks at DstFn with first/last bytes as:",chunkCount)
			//log.Trace(dataQueue[key+";0"][0:9],dataQueue[key+";"+strconv.Itoa(chunkCount-1)][len(dataQueue[key+";"+strconv.Itoa(chunkCount-1)])-9:])
			return elapsed, chunkCount
		}
		if err != nil {
			log.Fatalf("DST: receive error: %v", err)
		}
		log.Tracef("DST: Received chunk no. %d", chunkCount)
		if _,ok := dataQueue.Load(key); !ok {
			log.Infof("DST: creating a new channel")
			channel = make(chan []byte, 1600)
			dataQueue.Store(key, channel)
			log.Infof("DST: chunkTotal = %d",chunk.ChunkTotal)
			dataQueueSize.Store(key,chunk.ChunkTotal)
		}
		log.Infof("DST: Enquing chunk number %d",chunkCount)
		channel <- chunk.Chunk
		chunkCount += 1
	}

}

// start DstQP server
func StartDstServer(serverAddr string) {

	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("DST: failed to listen: %v", err)
	}

	server := grpc.NewServer()
	downXDT.RegisterXDTtoFnServer(server, downXDTServer{})

	log.Println("DST: start server")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("DST: failed to serve: %v", err)
	}
}
