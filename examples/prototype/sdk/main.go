package sdk

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	downXDT "github.com/ease-lab/vhive_stealth/examples/prototype/proto/downXDT"
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

type downXDTServer struct {
	downXDT.UnimplementedXDTtoFnServer
}

var dataQueue = make(map[string][]byte)

// Invoke the RPC call with XDT
func InvokeWithXDT(URL string, payloadByteArray []byte, chunkSizeInBytes int) time.Duration {
	now := time.Now()
	key := strconv.Itoa(int(now.UnixNano()))

	var xdtPayload payload
	if err := json.Unmarshal(payloadByteArray, &xdtPayload); err != nil {
		log.Fatal(err)
	}

	log.Printf("XDT invoke called with payload size %d", len(xdtPayload.Data))

	payloadData := xdtPayload.Data
	xdtPayload.Data = []byte("")
	xdtPayload.Key = key
	xdtPayload.isXDT = true

	serialisedPayload, _ := json.Marshal(xdtPayload)

	_ = PushData(key, payloadData, chunkSizeInBytes)

	fnInvocationCall(URL, serialisedPayload)

	elapsed := time.Since(now)
	return elapsed
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

// to be called by dQP to invoke remote fn
func (s downXDTServer) XDTFnCall(ctx context.Context, in *downXDT.InvocationRequest) (*downXDT.Empty, error) {

	log.Printf("destination received invocation call %s", in.XdtJson)

	var xdtPayload payload
	if err := json.Unmarshal(in.XdtJson, &xdtPayload); err != nil {
		log.Fatal(err)
	}

	key := xdtPayload.Key

	chunkSizeInBytes := 64 * 1024

	// fetch data from dQP
	FetchFromDQP(key, chunkSizeInBytes)
	return &downXDT.Empty{}, nil
}

// fetech data from dQP
func FetchFromDQP(key string, chunkSizeInBytes int) (time.Duration, []byte) {
	serverAddr := ":50006"
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}
	start := time.Now()

	// create stream
	client := downXDT.NewXDTtoFnClient(conn)
	in := &downXDT.DataRequest{Key: key, ChunkSize: int64(chunkSizeInBytes)}
	stream, err := client.XDTDataServe(context.Background(), in)
	if err != nil {
		log.Fatalf("open stream error %v", err)
	}
	//receive data from source QP
	// push to dataQueue
	packetCount := 1
	var payload []byte
	for {
		packet, err := stream.Recv()
		if err == io.EOF {
			elapsed := time.Since(start)
			log.Printf("Complete packet received at dQP")
			dataQueue[key] = payload
			return elapsed, payload
		}
		if err != nil {
			log.Fatalf("receive error: %v", err)
		}
		log.Printf("Received chunk no. %d", packetCount)
		payload = append(payload, packet.Chunk...)
		packetCount += 1
	}
	return time.Duration(-1), []byte{}
}

func StartDstServer(serverAddr string, handler func([]byte)) {

	// create listener for sdk
	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// create grpc server
	server := grpc.NewServer()
	downXDT.RegisterXDTtoFnServer(server, downXDTServer{})

	log.Println("start server")
	// and start...
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
