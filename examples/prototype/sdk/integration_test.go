package sdk

import (
	"crypto/rand"
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

var sample_size = flag.Int("sample", 10, "sample_size")
var URL = flag.String("URL", "bla", "Function URL")

func init(){
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05.000000", FullTimestamp: true})
}

func TestSdk_InvokeWithXDT(t *testing.T) {
	//create random blob
	payloadData := make([]byte, 10*1024*1024) // 10MiB
	rand.Read(payloadData)

	// start server at sQP
	go sqp.StartServer(sdk.LoadedConfig.SQPServerAddr)
	go dqp.StartServer(sdk.LoadedConfig.DQPServerAddr)
	go sdk.StartDstServer(sdk.LoadedConfig.DstServerAddr)

	chunkSizeInBytes := sdk.LoadedConfig.ChunkSizeInBytes

	payloadToSend := sdk.Payload{
		FunctionName: "HelloXDT",
		Data:         payloadData,
		Key:          "",
	}

	start := time.Now()
	log.Infof("starting integ test")
	sdk.InvokeWithXDT("", payloadToSend, chunkSizeInBytes)
	elapsed := time.Since(start)

	log.Printf("completed XDT in %s", elapsed)
}

func TestBenchmark_gRPC(t *testing.T) {

	if *sample_size < 10 {
		log.Fatal("invalid sample size. Acceptable input is integers >= 10")
	}

	// if *URL == "" {
	// 	log.Fatal("please enter destination url")
	// }

	go sqp.StartServer(sdk.LoadedConfig.SQPServerAddr)
	go dqp.StartServer(sdk.LoadedConfig.DQPServerAddr)
	go sdk.StartDstServer(sdk.LoadedConfig.DstServerAddr)

	payloadSizes := []int{10, 100, 1000, 10000, 100000}

	latencyMap := make(map[int][]float64)

	payloadData := make([]byte, 101*1024*1024) // 10MiB
	//create random payload
	rand.Read(payloadData)

	chunkSizeInBytes := sdk.LoadedConfig.ChunkSizeInBytes

	bench_payload := func(payloadSize int, chunkSizeInBytes int, sample_size int, URL string, payloadData []byte) []float64 {
		var latencies []float64
		payloadToSend := sdk.Payload{
			FunctionName: "HelloXDT",
			Data:         payloadData[:payloadSize],
			Key:          "",
		}

		for i := 0; i < sample_size; i += 1 {
			start := time.Now()
			sdk.InvokeWithXDT(URL, payloadToSend, chunkSizeInBytes)
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
		latencyMap[payloadSize] = latencies
	}

	plotter.PlotPercentile(latencyMap)
	plotter.PlotBW((latencyMap))
}
