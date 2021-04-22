package sdk

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"sort"
	"strconv"
	"testing"
	"time"

	plotter "github.com/ease-lab/vhive_stealth/examples/gRPC_stream/plotter"
	dqp "github.com/ease-lab/vhive_stealth/examples/prototype/dqp"
	sdk "github.com/ease-lab/vhive_stealth/examples/prototype/sdk"
	sqp "github.com/ease-lab/vhive_stealth/examples/prototype/sqp"
	log "github.com/sirupsen/logrus"
)

var sample_size = flag.Int("sample", 100, "sample_size")
var URL = flag.String("URL", "bla", "Function URL")

type payload struct {
	FunctionName string
	Data         []byte
	Key          string
	isXDT        bool
}

var config = sdk.LoadConfig("../config.json")

func TestSdk_InvokeWithXDT(t *testing.T) {
	//create random blob
	payload_data := make([]byte, 10*1024*1024) // 10MiB
	rand.Read(payload_data)

	// start server at sQP
	go sqp.StartServer(":50005")
	go dqp.StartServer(":50006")
	go sdk.StartDstServer(":50007")

	chunkSizeInBytes := config.ChunkSizeInBytes

	payloadToSend := &payload{
		FunctionName: "HelloXDT",
		Data:         payload_data,
		Key:          "",
	}
	payloadByteArray, _ := json.Marshal(payloadToSend)

	start := time.Now()
	sdk.InvokeWithXDT("", payloadByteArray, chunkSizeInBytes)
	elapsed := time.Since(start)

	log.Printf("completed XDT in %s", elapsed)
}

func TestBenchmark_gRPC(t *testing.T) {

	if *sample_size < 10 {
		log.Fatal("invalod sample size. Acceptable input is integers >= 10")
	}

	// if *URL == "" {
	// 	log.Fatal("please enter destination url")
	// }

	go sqp.StartServer(":50005")
	go dqp.StartServer(":50006")
	go sdk.StartDstServer(":50007")

	payloadSizes := []int{10, 100, 1000, 10000, 100000}

	latency_map := make(map[int][]float64)

	payloadData := make([]byte, 101*1024*1024) // 10MiB
	//create random payload
	rand.Read(payloadData)

	chunkSizeInBytes := config.ChunkSizeInBytes

	bench_payload := func(payloadSize int, chunkSizeInBytes int, sample_size int, URL string, payloadData []byte) []float64 {
		latencies := []float64{}
		payloadToSend := &payload{
			FunctionName: "HelloXDT",
			Data:         payloadData[:payloadSize],
			Key:          "",
		}
		payloadByteArray, _ := json.Marshal(payloadToSend)

		for i := 0; i < sample_size; i += 1 {
			start := time.Now()
			sdk.InvokeWithXDT(URL, payloadByteArray, chunkSizeInBytes)
			latency_in_us := time.Since(start).Microseconds()
			latencies = append(latencies, float64(latency_in_us))
		}
		sort.Float64s(latencies)
		return latencies
	}

	for _, payloadSize := range payloadSizes {
		payloadSizeInBytes := payloadSize * 1024
		log.Printf("checking for %dKiB", payloadSize)
		latencies := bench_payload(payloadSizeInBytes, chunkSizeInBytes, *sample_size, *URL, payloadData)
		plotter.PlotLatenciesCDF("./cdf_"+strconv.Itoa(payloadSize)+"KiB.png", latencies, payloadSize)
		latency_map[payloadSize] = latencies
	}

	plotter.PlotPercentile(latency_map)
	plotter.PlotBW((latency_map))
}
