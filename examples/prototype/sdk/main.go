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

type payload struct {
	FunctionName string
	Data         []byte
	Key          string
	isXDT        bool
}

// Invoke the RPC call with XDT
func InvokeWithXDT(URL string, payloadByteArray []byte, chunkSizeInBytes int) time.Duration {
	now := time.Now()
	key := strconv.Itoa(int(now.UnixNano()))

	var xdtPayload payload
	if err := json.Unmarshal(payloadByteArray, &xdtPayload); err != nil {
		log.Fatal(err)
	}

	payloadData := xdtPayload.Data
	xdtPayload.Data = []byte("")
	xdtPayload.Key = key
	xdtPayload.isXDT = true

	serialisedPayload, _ := json.Marshal(xdtPayload)

	dataTransferDuration := PushData(key, payloadData, chunkSizeInBytes)

	fnInvocationCall(URL, serialisedPayload)

	return dataTransferDuration
}

// make fn invocation call with xdt payload
func fnInvocationCall(URL string, serialisedPayload []byte) {
	//gRPC call to the function in dQP responsible for fetching data for g(x)
	serverAddr := ":50006"

	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := fnInvocation.NewInvocationClient(conn)

	_, err = c.RouteInvocationCall(context.Background(), &fnInvocation.InvocationRequest{XdtJson: serialisedPayload})
	if err == nil {
		log.Printf("Fn invocation from source SDK successful")
	}
}

// push data to source QP
func PushData(key string, payload []byte, chunkSizeInBytes int) time.Duration {

	serverAddr := ":50005"
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}
	start := time.Now()

	// create stream
	client := upXDT.NewStreamDataClient(conn)
	payloadSize := len(payload)
	log.Printf("sending payload of size %d bytes", payloadSize)
	stream, err := client.CollectData(context.Background())
	if err != nil {
		log.Fatalf("open stream error %v", err)
	}

	for currentByte := int(0); currentByte < payloadSize; currentByte += chunkSizeInBytes {

		if currentByte+chunkSizeInBytes > payloadSize {
			req := upXDT.Request{Chunk: payload[currentByte:payloadSize], Key: key}
			if err := stream.Send(&req); err != nil {
				log.Printf("send error %v", err)
			}
			log.Printf("finishing request number : %d", currentByte)
		} else {
			req := upXDT.Request{Chunk: payload[currentByte : currentByte+chunkSizeInBytes], Key: key}
			if err := stream.Send(&req); err != nil {
				log.Printf("send error %v", err)
			}
			log.Printf("finishing request number : %d", currentByte)
		}

	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
	elapsed := time.Since(start)
	return elapsed
}
