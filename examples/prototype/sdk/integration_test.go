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
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"sort"
	"testing"
	"time"

	"XDTgRPC_stream/plotter"
	"XDTprototype/dqp"
	"XDTprototype/sdk"
	"XDTprototype/sqp"
	"XDTprototype/tracing"
	"XDTprototype/utils"
	log "github.com/sirupsen/logrus"
)

var sampleSize = flag.Int("sample", 10, "sampleSize")
var URL = flag.String("url", "helloworld.default.192.168.1.240.nip.io", "Function URL")
var numConcurrentFunctions = flag.Int("concurrentCalls", 5, "num of simultaneous calls")
var chunkSizeInBytes = utils.LoadConfig.ChunkSizeInBytes

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05.000000", FullTimestamp: true, ForceColors: true})
}

var handler = func(data []byte) {
	log.Infof("integ-test destination handler received data of size %d", len(data))
}

func preparePayload() utils.Payload {
	payloadData := make([]byte, 10*1024*1024) // 10MiB
	if _, err := rand.Read(payloadData); err != nil {
		log.Fatal(err)
	}

	payloadToSend := utils.Payload{
		FunctionName: "HelloXDT",
		Data:         payloadData,
		Key:          "",
	}
	return payloadToSend
}

func TestSdk_InvokeWithXDT(t *testing.T) {

	config := utils.LoadConfig
	if config.TracingEnabled {
		shutdown, err := tracing.InitTracer()
		if err != nil {
			log.Fatal(err)
		}
		defer shutdown()
	}
	// start server at sQP
	go sqp.StartServer(config)
	go dqp.StartServer(config)
	go sdk.StartDstServer(config, handler)

	time.Sleep(time.Second * 1)

	start := time.Now()
	log.Infof("starting integ test")
	url := config.LBAddr
	if err := sdk.InvokeWithXDT(url, preparePayload(), config.SQPServerAddr, chunkSizeInBytes); err != nil {
		log.Fatalf("TestSdk_InvokeWithXDT failed %v", err)
	}
	elapsed := time.Since(start)

	log.Infof("completed XDT in %s", elapsed)
}

func TestErr_DQPTimeout(t *testing.T) {

	config := utils.LoadConfig
	if config.TracingEnabled {
		shutdown, err := tracing.InitTracer()
		if err != nil {
			log.Fatal(err)
		}
		defer shutdown()
	}

	// start server at sQP
	go sqp.StartServer(config)
	time.Sleep(time.Second)
	go sdk.StartDstServer(config, handler)

	time.Sleep(time.Second * 1)

	start := time.Now()
	log.Infof("starting integ test")
	url := config.LBAddr
	if err := sdk.InvokeWithXDT(url, preparePayload(), config.SQPServerAddr, chunkSizeInBytes); err == context.DeadlineExceeded {
		log.Errorf("TestSdk_InvokeWithXDT failed predictably")
	} else {
		log.Fatalf("Unexpected Error Occured: %v", err)
	}
	elapsed := time.Since(start)

	log.Infof("completed XDT in %s", elapsed)
}

func TestErr_DSTTimeout(t *testing.T) {

	config := utils.LoadConfig
	if config.TracingEnabled {
		shutdown, err := tracing.InitTracer()
		if err != nil {
			log.Fatal(err)
		}
		defer shutdown()
	}

	// start server at sQP
	go dqp.StartServer(config)
	time.Sleep(time.Second * 1)
	go sqp.StartServer(config)

	time.Sleep(time.Second * 1)

	start := time.Now()
	log.Infof("starting integ test")
	url := config.LBAddr
	if err := sdk.InvokeWithXDT(url, preparePayload(), config.SQPServerAddr, chunkSizeInBytes); err == context.DeadlineExceeded {
		log.Errorf("TestSdk_InvokeWithXDT failed predictably")
	} else {
		log.Fatalf("Unexpected Error Occured: %v", errors.Unwrap(err))
	}
	elapsed := time.Since(start)

	log.Infof("completed XDT in %s", elapsed)
}

func TestParallel_Invoke(t *testing.T) {

	config := utils.LoadConfig
	if config.TracingEnabled {
		shutdown, err := tracing.InitTracer()
		if err != nil {
			log.Fatal(err)
		}
		defer shutdown()
	}

	// start server at sQP
	go sqp.StartServer(config)
	go dqp.StartServer(config)
	go sdk.StartDstServer(config, handler)

	time.Sleep(time.Second * 1)

	start := time.Now()
	log.Infof("starting integ test")
	url := config.LBAddr
	errChannel := make(chan error, *numConcurrentFunctions)
	for i := 0; i < *numConcurrentFunctions; i += 1 {
		go func() {
			errChannel <- sdk.InvokeWithXDT(url, preparePayload(), config.SQPServerAddr, chunkSizeInBytes)
		}()
	}
	for i := 0; i < *numConcurrentFunctions; i += 1 {
		err := <-errChannel
		if err != nil {
			log.Fatalf("InvokeWithXDT no. %d failed with: %v", i, err)
		}

	}
	elapsed := time.Since(start)

	log.Infof("completed TestFan_In in %s", elapsed)
}

func TestParallel_FanIn(t *testing.T) {

	config := utils.LoadConfig
	if config.TracingEnabled {
		shutdown, err := tracing.InitTracer()
		if err != nil {
			log.Fatal(err)
		}
		defer shutdown()
	}

	// start server at sQP
	sQPAddr := 50009
	for i := 0; i < *numConcurrentFunctions; i += 1 {
		tmpConfig := utils.LoadConfig
		tmpConfig.SQPServerAddr = ":" + fmt.Sprint(sQPAddr+i)
		log.Infof("starting sQP server no. %d", i+1)
		go sqp.StartServer(tmpConfig)
		time.Sleep(time.Second * 10)
	}
	go dqp.StartServer(config)
	time.Sleep(time.Second * 2)
	go sdk.StartDstServer(config, handler)
	time.Sleep(time.Second * 2)

	start := time.Now()
	log.Infof("starting integ test")
	url := config.LBAddr
	numberOfSources := *numConcurrentFunctions
	errChannel := make(chan error, numberOfSources)

	for i := 0; i < numberOfSources; i += 1 {
		i := i
		go func() {
			errChannel <- sdk.InvokeWithXDT(url, preparePayload(), ":"+fmt.Sprint(sQPAddr+i), chunkSizeInBytes)
		}()
	}

	for i := 0; i < numberOfSources; i += 1 {
		err := <-errChannel
		if err != nil {
			log.Fatalf("InvokeWithXDT no. %d failed with: %v", i, err)
		}

	}
	elapsed := time.Since(start)

	log.Infof("completed TestFan_In in %s", elapsed)
}

func TestParallel_FanOut(t *testing.T) {

	config := utils.LoadConfig
	if config.TracingEnabled {
		shutdown, err := tracing.InitTracer()
		if err != nil {
			log.Fatal(err)
		}
		defer shutdown()
	}

	// start server at sQP
	dQPAddr := 50009
	for i := 0; i < *numConcurrentFunctions; i += 1 {
		config.DstServerAddr = ":" + fmt.Sprint(dQPAddr+i)
		config.DQPServerAddr = ":" + fmt.Sprint(dQPAddr+i+*numConcurrentFunctions)
		tmpDstConfig := config
		log.Infof("starting Dst server no. %d", i+1)
		go sdk.StartDstServer(tmpDstConfig, handler)
		time.Sleep(time.Second * 10)
		tmpDQPConfig := config
		log.Infof("starting dQP server no. %d", i+1)
		go dqp.StartServer(tmpDQPConfig)
		time.Sleep(time.Second * 10)
	}
	time.Sleep(time.Second * 5)
	go sqp.StartServer(config)

	time.Sleep(time.Second * 1)

	start := time.Now()
	log.Infof("starting integ test")
	numberOfSources := *numConcurrentFunctions
	errChannel := make(chan error, numberOfSources)

	for i := 0; i < numberOfSources; i += 1 {
		url := ":" + fmt.Sprint(dQPAddr+i+*numConcurrentFunctions)
		go func() {
			errChannel <- sdk.InvokeWithXDT(url, preparePayload(), config.SQPServerAddr, chunkSizeInBytes)
		}()
	}

	for i := 0; i < numberOfSources; i += 1 {
		err := <-errChannel
		if err != nil {
			log.Fatalf("InvokeWithXDT no. %d failed with: %v", i, err)
		}

	}
	elapsed := time.Since(start)

	log.Infof("completed TestFan_In in %s", elapsed)
}

func TestPython_SDK(t *testing.T) {

	config := utils.LoadConfig
	// start servers
	go sqp.StartServer(config)
	dqp.StartServer(config)
	//sdk.StartDstServer(sdk.LoadedConfig.DstServerAddr, handler)

}

func TestBenchmark_XDT(t *testing.T) {

	config := utils.LoadConfig
	if *sampleSize < 10 {
		log.Fatal("invalid sample size. Acceptable input is integers >= 10")
	}

	go sqp.StartServer(config)
	go dqp.StartServer(config)
	go sdk.StartDstServer(config, handler)

	payloadSizes := []int{10, 100, 1000, 10000, 100000}

	latencyMap := make(map[int][]float64)

	payloadData := make([]byte, 101*1024*1024) // 10MiB
	if _, err := rand.Read(payloadData); err != nil {
		log.Fatal(err)
	}
	url := utils.LoadConfig.LBAddr

	benchPayload := func(payloadSize int, chunkSizeInBytes int, sampleSize int, URL string, payloadData []byte) []float64 {
		var latencies []float64
		payloadToSend := utils.Payload{
			FunctionName: "HelloXDT",
			Data:         payloadData[:payloadSize],
			Key:          "",
		}

		for i := 0; i < sampleSize; i += 1 {
			start := time.Now()
			if err := sdk.InvokeWithXDT(url, payloadToSend, config.SQPServerAddr, chunkSizeInBytes); err != nil {
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
