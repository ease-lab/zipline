package sdk

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	FnInvocationProto "github.com/ease-lab/vhive_stealth/examples/prototype/proto/FnInvocationProto"
	upXDT "github.com/ease-lab/vhive_stealth/examples/prototype/proto/upXDT"

	"google.golang.org/grpc"
)

type payload struct {
	FunctionName string
	Data         []byte
	Key          string
	isXDT        bool
}

func InvokeWithXDT(URL string, payloadByteArray []byte, chunk_size_in_bytes int) time.Duration {
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

	dataTransferDuration := PushData(key, payloadData, chunk_size_in_bytes)

	control_path_call(URL, serialisedPayload)

	return dataTransferDuration
}

func control_path_call(URL string, serialisedPayload []byte) {
	//gRPC call to the function in dQP responsible for fetching data for g(x)
	serverAddr := ":50006"

	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := FnInvocationProto.NewInvocationClient(conn)

	_, err = c.RouteInvocationCall(context.Background(), &FnInvocationProto.InvocationRequest{XdtJson: serialisedPayload})
	if err == nil {
		log.Printf("Fn invocation from source SDK successful")
	}
}

func PushData(key string, payload []byte, chunk_size_in_bytes int) time.Duration {

	serverAddr := ":50005"
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}
	start := time.Now()

	// create stream
	client := upXDT.NewStreamDataClient(conn)
	payload_size := len(payload)
	log.Printf("sending payload of size %d bytes", payload_size)
	stream, err := client.CollectData(context.Background())
	if err != nil {
		log.Fatalf("open stream error %v", err)
	}

	for currentByte := int(0); currentByte < payload_size; currentByte += chunk_size_in_bytes {

		if currentByte+chunk_size_in_bytes > payload_size {
			req := upXDT.Request{Chunk: payload[currentByte:payload_size], Key: key}
			if err := stream.Send(&req); err != nil {
				log.Printf("send error %v", err)
			}
			log.Printf("finishing request number : %d", currentByte)
		} else {
			req := upXDT.Request{Chunk: payload[currentByte : currentByte+chunk_size_in_bytes], Key: key}
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
