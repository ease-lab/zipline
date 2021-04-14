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
	log "github.com/sirupsen/logrus"
)

var chunk_size = flag.Int("chunk", 64, "chunk_size")
var sample_size = flag.Int("sample", 100, "sample_size")
var URL = flag.String("URL", "", "Function URL")

func TestSdk_push(t *testing.T) {

	now := time.Now()
	key := strconv.Itoa(int(now.UnixNano()))
	payload_data := make([]byte, 10*1024*1024) // 10MiB
	//create random blob
	rand.Read(payload_data)
	chunkSizeInBytes := 64 * 1024

	payloadToSend := &payload{
		FunctionName: "HelloXDT",
		Data:         payload_data,
		Key:          ""}
	payloadByteArray, _ := json.Marshal(payloadToSend)

	duration := push_data(key, payloadByteArray, chunkSizeInBytes)
	// vhive_call("bla", payload, chunk_size_in_bytes)
	log.Printf("sent %d bytes in %s", len(payloadByteArray), duration)

}

func TestSdk_VHiveCall(t *testing.T) {
	//create random blob
	payload_data := make([]byte, 10*1024*1024) // 10MiB
	rand.Read(payload_data)

	chunkSizeInBytes := 64 * 1024

	payloadToSend := &payload{
		FunctionName: "HelloXDT",
		Data:         payload_data,
		Key:          "",
	}
	payloadByteArray, _ := json.Marshal(payloadToSend)

	duration := vhive_call("", payloadByteArray, chunkSizeInBytes)

	log.Printf("completed XDT in %s", duration)
}

func TestSdk_Fninvocation(t *testing.T) {
	payloadToSend := &payload{
		FunctionName: "HelloXDT",
		Data:         []byte{},
		Key:          "123456789",
		isXDT:        true,
	}
	payloadByteArray, _ := json.Marshal(payloadToSend)

	control_path_call("", payloadByteArray)

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

	payload_sizes := []int{10, 100, 1000, 10000, 100000}

	latency_map := make(map[int][]float64)

	payload := make([]byte, 10*1024*1024) // 10MiB
	//create random payload
	rand.Read(payload)

	bench_payload := func(payload_size int, chunk_size int, sample_size int, URL string, payload []byte) []float64 {
		latencies := []float64{}
		chunk_size_in_bytes := chunk_size * 1024
		for i := 0; i < sample_size; i += 1 {
			latency_in_us := vhive_call(URL, payload[:payload_size], chunk_size_in_bytes).Microseconds()
			latencies = append(latencies, float64(latency_in_us))
		}
		sort.Float64s(latencies)
		return latencies
	}

	for _, payload_size := range payload_sizes {
		log.Printf("checking for %dKiB", payload_size)
		latencies := bench_payload(payload_size, *chunk_size, *sample_size, *URL, payload)
		plotter.PlotLatenciesCDF("./cdf_"+strconv.Itoa(payload_size)+"KiB.png", latencies, payload_size)
		latency_map[payload_size] = latencies
	}

	plotter.PlotPercentile(latency_map)
	plotter.PlotBW((latency_map))
}
