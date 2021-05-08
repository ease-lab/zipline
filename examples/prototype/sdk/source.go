package sdk

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	fnInvocation "github.com/ease-lab/vhive_stealth/examples/prototype/proto/fnInvocation"
	upXDT "github.com/ease-lab/vhive_stealth/examples/prototype/proto/upXDT"

	"google.golang.org/grpc"
)

// Invoke the RPC call with XDT
func InvokeWithXDT(URL string, xdtPayload Payload, chunkSizeInBytes int) {

	log.Infof("SDK: XDT invoke start")
	now := time.Now()
	key := strconv.Itoa(int(now.UnixNano()))
	log.Infof("XDT invoke called with payload size %d", len(xdtPayload.Data))

	payloadData := xdtPayload.Data
	log.Info(payloadData[0:9],payloadData[len(payloadData)-9:])
	xdtPayload.Data = []byte("")
	xdtPayload.Key = key
	xdtPayload.IsXDT = true

	serialisedPayload, _ := json.Marshal(xdtPayload)

	if LoadedConfig.Routing == "S&F" {
		log.Info("SDK: using store & forward routing")
		PushData(key, payloadData, chunkSizeInBytes)
	}else if  LoadedConfig.Routing == "CT" {
		log.Info("SDK: using cut through routing")
		go PushData(key, payloadData, chunkSizeInBytes)
	}

	fnInvocationCall(URL, serialisedPayload)
}

// make fn invocation call to dQP with xdt payload
func fnInvocationCall(URL string, serialisedPayload []byte) {

	serverAddr := LoadedConfig.DQPServerAddr
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := fnInvocation.NewInvocationClient(conn)

	log.Infof("SDK: Fn invocation start")
	_, err = c.RouteInvocation(context.Background(), &fnInvocation.InvocationRequest{XDTJSON: serialisedPayload})
	if err == nil {
		log.Infof("SDK: Fn invocation successful")
	}
}

// push data to source QP
func PushData(key string, payload []byte, chunkSizeInBytes int) {

	serverAddr := LoadedConfig.SQPServerAddr
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}

	client := upXDT.NewStreamDataClient(conn)
	payloadSize := len(payload)
	log.Infof("Transfering %d bytes to sQP", payloadSize)
	stream, err := client.SendData(context.Background())
	if err != nil {
		log.Fatalf("open stream error %v", err)
	}

	chunkTotal := len(payload)/chunkSizeInBytes
	if len(payload)%chunkSizeInBytes!=0 {
		chunkTotal+=1
	}

	for currentByte := 0; currentByte < payloadSize; currentByte += chunkSizeInBytes {

		if currentByte+chunkSizeInBytes > payloadSize {
			req := upXDT.Request{Chunk: payload[currentByte:payloadSize], Key: key, ChunkTotal: int64(chunkTotal)}
			if err := stream.Send(&req); err != nil {
				log.Fatalf("send error %v", err)
			}
			log.Tracef("finishing request number : %d", currentByte)
		} else {
			req := upXDT.Request{Chunk: payload[currentByte : currentByte+chunkSizeInBytes], Key: key, ChunkTotal: int64(chunkTotal)}
			if err := stream.Send(&req); err != nil {
				log.Fatalf("send error %v", err)
			}
			log.Tracef("finishing request number : %d", currentByte)
		}

	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
}
