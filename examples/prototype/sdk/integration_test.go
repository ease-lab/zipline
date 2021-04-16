package sdk

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"sort"
	"strconv"
	"testing"

	plotter "github.com/ease-lab/vhive_stealth/examples/gRPC_stream/plotter"
	dqp "github.com/ease-lab/vhive_stealth/examples/prototype/dqp"
	gx "github.com/ease-lab/vhive_stealth/examples/prototype/gx"
	sdk "github.com/ease-lab/vhive_stealth/examples/prototype/sdk"
	sqp "github.com/ease-lab/vhive_stealth/examples/prototype/sqp"
	log "github.com/sirupsen/logrus"
)

var chunk_size = flag.Int("chunk", 64, "chunk_size")
var sample_size = flag.Int("sample", 100, "sample_size")
var URL = flag.String("URL", "", "Function URL")

type payload struct {
	FunctionName string
	Data         []byte
	Key          string
	isXDT        bool
}

func TestSdk_InvokeWithXDT(t *testing.T) {
	//create random blob
	payload_data := make([]byte, 10*1024*1024) // 10MiB
	rand.Read(payload_data)

	// start server at sQP
	go sqp.StartServer(":50005")
	go dqp.StartServer(":50006")
	go gx.StartServer(":50007")

	chunkSizeInBytes := 64 * 1024

	payloadToSend := &payload{
		FunctionName: "HelloXDT",
		Data:         payload_data,
		Key:          "",
	}
	payloadByteArray, _ := json.Marshal(payloadToSend)

	duration := sdk.InvokeWithXDT("", payloadByteArray, chunkSizeInBytes)

	log.Printf("completed XDT in %s", duration)
}

func TestBenchmark_gRPC(t *testing.T) {

	if *chunk_size < 1 || *chunk_size > 4096 {
		log.Fatal("invalod chunk size in KiB. Acceptable input is integers between 1 to 4096")
	}

	if *sample_size < 10 {
		log.Fatal("invalod sample size. Acceptable input is integers >= 10")
	}

	if *URL == "" {
		log.Fatal("please enter destination url")
	}

	payloadSizes := []int{10, 100, 1000, 10000, 100000}

	latency_map := make(map[int][]float64)

	payload := make([]byte, 10*1024*1024) // 10MiB
	//create random payload
	rand.Read(payload)

	bench_payload := func(payloadSize int, chunk_size int, sample_size int, URL string, payload []byte) []float64 {
		latencies := []float64{}
		chunkSizeInBytes := chunk_size * 1024
		for i := 0; i < sample_size; i += 1 {
			latency_in_us := sdk.InvokeWithXDT(URL, payload[:payloadSize], chunkSizeInBytes).Microseconds()
			latencies = append(latencies, float64(latency_in_us))
		}
		sort.Float64s(latencies)
		return latencies
	}

	for _, payloadSize := range payloadSizes {
		log.Printf("checking for %dKiB", payloadSize)
		latencies := bench_payload(payloadSize, *chunk_size, *sample_size, *URL, payload)
		plotter.PlotLatenciesCDF("./cdf_"+strconv.Itoa(payloadSize)+"KiB.png", latencies, payloadSize)
		latency_map[payloadSize] = latencies
	}

	plotter.PlotPercentile(latency_map)
	plotter.PlotBW((latency_map))
}
