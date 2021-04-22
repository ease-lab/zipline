package sdk

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"os"
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

type Config struct {
	ChunkSizeInBytes int
}

type downXDTServer struct {
	downXDT.UnimplementedXDTtoFnServer
}

var dataQueue = make(map[string][]byte)

var config = LoadConfig("../config.json")

func LoadConfig(file string) Config {
	log.Debugf("Opening JSON file with config: %s\n", file)
	jsonFile, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatal(err)
	}

	var config Config

	json.Unmarshal(byteValue, &config)

	return config
}

// Invoke the RPC call with XDT
func InvokeWithXDT(URL string, payloadByteArray []byte, chunkSizeInBytes int) {

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
}

// make fn invocation call to dQP with xdt payload
func fnInvocationCall(URL string, serialisedPayload []byte) {

	serverAddr := ":50006"
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := fnInvocation.NewInvocationClient(conn)

	_, err = c.RouteInvocation(context.Background(), &fnInvocation.InvocationRequest{XdtJson: serialisedPayload})
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

	client := upXDT.NewStreamDataClient(conn)
	payloadSize := len(payload)
	log.Printf("sending payload of size %d bytes", payloadSize)
	stream, err := client.SendData(context.Background())
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

// to be called by dQP to invoke DstFn
func (s downXDTServer) XDTFnCall(ctx context.Context, in *downXDT.InvocationRequest) (*downXDT.Empty, error) {

	log.Printf("destination received invocation call %s", in.XdtJson)

	var xdtPayload payload
	if err := json.Unmarshal(in.XdtJson, &xdtPayload); err != nil {
		log.Fatal(err)
	}

	key := xdtPayload.Key

	chunkSizeInBytes := config.ChunkSizeInBytes

	// fetch data from dQP
	FetchFromDQP(key, chunkSizeInBytes)
	return &downXDT.Empty{}, nil
}

// fetch data from dQP to DstFn
func FetchFromDQP(key string, chunkSizeInBytes int) (time.Duration, []byte) {
	serverAddr := ":50006"
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}
	start := time.Now()

	client := downXDT.NewXDTtoFnClient(conn)
	in := &downXDT.DataRequest{Key: key, ChunkSize: int64(chunkSizeInBytes)}
	stream, err := client.XDTDataServe(context.Background(), in)
	if err != nil {
		log.Fatalf("open stream error %v", err)
	}

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

// start DstQP server
func StartDstServer(serverAddr string) {

	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	downXDT.RegisterXDTtoFnServer(server, downXDTServer{})

	log.Println("start server")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
