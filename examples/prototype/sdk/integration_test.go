// MIT License
//
// Copyright (c) 2021 Shyam Jesalpura and EASE lab
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package sdk

import (
	"crypto/rand"
	"flag"
	"sort"
	"testing"
	"time"

	"XDTgRPC_stream/plotter"
	"XDTprototype/dqp"
	"XDTprototype/sdk"
	"XDTprototype/sqp"
	log "github.com/sirupsen/logrus"
)

var sampleSize = flag.Int("sample", 10, "sampleSize")
var URL = flag.String("url", "helloworld.default.192.168.1.240.nip.io", "Function URL")

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05.000000", FullTimestamp: true})
}

var handler = func(data []byte) {
	log.Infof("integ-test destination handler received data of size %d", len(data))
}

func TestSdk_InvokeWithXDT(t *testing.T) {

	payloadData := make([]byte, 10*1024*1024) // 10MiB
	if _, err := rand.Read(payloadData); err != nil {
		log.Fatal(err)
	}

	// start server at sQP
	go sqp.StartServer(sdk.LoadedConfig.SQPServerAddr)
	go dqp.StartServer(sdk.LoadedConfig.DQPServerAddr)
	go sdk.StartDstServer(sdk.LoadedConfig.DstServerAddr, handler)

	time.Sleep(time.Second * 2)

	chunkSizeInBytes := sdk.LoadedConfig.ChunkSizeInBytes

	payloadToSend := sdk.Payload{
		FunctionName: "HelloXDT",
		Data:         payloadData,
		Key:          "",
	}

	start := time.Now()
	log.Infof("starting integ test")
	url := sdk.LoadedConfig.LBAddr
	if err := sdk.InvokeWithXDT(url, payloadToSend, chunkSizeInBytes); err != nil {
		log.Fatalf("TestSdk_InvokeWithXDT failed %v", err)
	}
	elapsed := time.Since(start)

	log.Infof("completed XDT in %s", elapsed)
}

func TestBenchmark_XDT(t *testing.T) {

	if *sampleSize < 10 {
		log.Fatal("invalid sample size. Acceptable input is integers >= 10")
	}

	go sqp.StartServer(sdk.LoadedConfig.SQPServerAddr)
	go dqp.StartServer(sdk.LoadedConfig.DQPServerAddr)
	go sdk.StartDstServer(sdk.LoadedConfig.DstServerAddr, handler)

	payloadSizes := []int{10, 100, 1000, 10000, 100000}

	latencyMap := make(map[int][]float64)

	payloadData := make([]byte, 101*1024*1024) // 10MiB
	if _, err := rand.Read(payloadData); err != nil {
		log.Fatal(err)
	}
	chunkSizeInBytes := sdk.LoadedConfig.ChunkSizeInBytes
	url := sdk.LoadedConfig.LBAddr

	benchPayload := func(payloadSize int, chunkSizeInBytes int, sampleSize int, URL string, payloadData []byte) []float64 {
		var latencies []float64
		payloadToSend := sdk.Payload{
			FunctionName: "HelloXDT",
			Data:         payloadData[:payloadSize],
			Key:          "",
		}

		for i := 0; i < sampleSize; i += 1 {
			start := time.Now()
			if err := sdk.InvokeWithXDT(url, payloadToSend, chunkSizeInBytes); err != nil {
				log.Fatalf("TestBenchmark_XDT failed %v", err)
			}
			latencyInUs := time.Since(start).Microseconds()
			latencies = append(latencies, float64(latencyInUs))
		}
		sort.Float64s(latencies)
		return latencies
	}

	for _, payloadSize := range payloadSizes {
		payloadSizeInBytes := payloadSize * 1024
		log.Infof("checking for %dKiB", payloadSize)
		latencies := benchPayload(payloadSizeInBytes, chunkSizeInBytes, *sampleSize, *URL, payloadData)
		plotter.PlotLatenciesCDF(latencies, payloadSize)
		latencyMap[payloadSize] = latencies
	}

	plotter.PlotPercentile(latencyMap)
	plotter.PlotBW(latencyMap)
}
