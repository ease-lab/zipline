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

package integration_tests

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"sort"
	"testing"
	"time"

	"github.com/ease-lab/xdt/sdk"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	pkgnet "knative.dev/pkg/network"
	"knative.dev/serving/pkg/queue"

	"XDTgRPC_stream/plotter"

	ctrdlog "github.com/containerd/containerd/log"
	tracing "github.com/ease-lab/vhive/utils/tracing/go"
	"github.com/ease-lab/xdt/dQP"
	"github.com/ease-lab/xdt/sQP"
	"github.com/ease-lab/xdt/utils"

	log "github.com/sirupsen/logrus"
)

var sampleSize = flag.Int("sample", 10, "sampleSize")
var URL = flag.String("url", "helloworld.default.192.168.1.240.nip.io", "Function URL")
var zipkinURL = flag.String("zipkin", "http://localhost:9411/api/v2/spans", "zipkin url")
var numConcurrentFunctions = flag.Int("concurrentCalls", 5, "num of simultaneous calls")
var chunkSizeInBytes = utils.LoadConfig.ChunkSizeInBytes

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: ctrdlog.RFC3339NanoFixed,
		FullTimestamp:   true,
		ForceColors:     true})
}

var handler = func(data []byte) {
	log.Infof("integ-test destination handler received data of size %d", len(data))
}

func knativeQP(config utils.Config) {
	go dQP.StartServer(config)

	httpProxy := func(target string) *httputil.ReverseProxy {
		return &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = "http"
				req.URL.Host = target

				// Copied from httputil.NewSingleHostReverseProxy.
				if _, ok := req.Header["User-Agent"]; !ok {
					// explicitly disable User-Agent so it's not set to default value
					req.Header.Set("User-Agent", "")
				}
			},
		}
	}(config.DstServerHostname + config.DstServerPort)
	httpProxy.Transport = pkgnet.NewProxyAutoTransport(10 /* max-idle */, 5 /* max-idle-per-host */)

	var composedHandler http.Handler = httpProxy

	composedHandler = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer h.ServeHTTP(w, r)

			isXDT := r.Header.Get("is_xdt")
			if isXDT == "true" {
				log.Infof("pulling from sQP using key %s addr %s", r.Header.Get("key"), r.Header.Get("sqp_addr"))
				go func() {
					err := dQP.PullDataFromSrcQP(r.Context(), r.Header.Get("key"), r.Header.Get("sqp_addr"), config.ChunkSizeInBytes)
					if err != nil {
						log.Fatalf("Proxy: Failed to pull data from sQP: %v", err)
					}
				}()
			}

		})
	}(composedHandler)

	composedHandler = queue.ForwardedShimHandler(composedHandler)

	h2s := &http2.Server{}
	// start server
	server := &http.Server{
		Addr:    config.ProxyPort,
		Handler: h2c.NewHandler(composedHandler, h2s),
	}
	log.Infof("Listening to %s...\n", config.ProxyPort)
	err := server.ListenAndServe()
	if err != nil {
		log.Errorf("failed to start proxy server")
	}
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
		shutdown, err := tracing.InitBasicTracer(*zipkinURL, "xdt")
		if err != nil {
			log.Warn(err)
		}
		defer shutdown()
	}
	// start server at sQP
	go sQP.StartServer(config)
	go knativeQP(config)
	go sdk.StartDstServer(config, handler)

	time.Sleep(time.Second * 1)

	start := time.Now()
	log.Infof("starting integ test")
	url := config.ProxyHostname + config.ProxyPort
	if err := sdk.InvokeWithXDT(url, preparePayload(), config.SQPServerHostname+config.SQPServerPort, chunkSizeInBytes); err != nil {
		log.Fatalf("TestSdk_InvokeWithXDT failed %v", err)
	}
	elapsed := time.Since(start)

	log.Infof("completed XDT in %s", elapsed)
}

func TestErr_DQPTimeout(t *testing.T) {

	config := utils.LoadConfig
	if config.TracingEnabled {
		shutdown, err := tracing.InitBasicTracer(*zipkinURL, "xdt")
		if err != nil {
			log.Warn(err)
		}
		defer shutdown()
	}

	// start server at sQP
	go sQP.StartServer(config)
	time.Sleep(time.Second)
	go sdk.StartDstServer(config, handler)

	time.Sleep(time.Second * 1)

	start := time.Now()
	log.Infof("starting integ test")
	url := config.ProxyHostname + config.ProxyPort
	if err := sdk.InvokeWithXDT(url, preparePayload(), config.SQPServerHostname+config.SQPServerPort, chunkSizeInBytes); err == context.DeadlineExceeded {
		log.Errorf("TestSdk_InvokeWithXDT failed predictably")
	} else {
		log.Fatalf("Unexpected Error Occured: %v", err)
	}
	elapsed := time.Since(start)

	log.Infof("completed XDT in %s", elapsed)
}

func TestParallel_Invoke(t *testing.T) {

	config := utils.LoadConfig
	if config.TracingEnabled {
		shutdown, err := tracing.InitBasicTracer(*zipkinURL, "xdt")
		if err != nil {
			log.Warn(err)
		}
		defer shutdown()
	}

	// start server at sQP
	go sQP.StartServer(config)
	go knativeQP(config)
	go sdk.StartDstServer(config, handler)

	time.Sleep(time.Second * 1)

	start := time.Now()
	log.Infof("starting integ test")
	url := config.ProxyHostname + config.ProxyPort
	errChannel := make(chan error, *numConcurrentFunctions)
	for i := 0; i < *numConcurrentFunctions; i += 1 {
		go func() {
			errChannel <- sdk.InvokeWithXDT(url, preparePayload(), config.SQPServerHostname+config.SQPServerPort, chunkSizeInBytes)
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
		shutdown, err := tracing.InitBasicTracer(*zipkinURL, "xdt")
		if err != nil {
			log.Warn(err)
		}
		defer shutdown()
	}

	// start server at sQP
	sQPPort := 50009
	for i := 0; i < *numConcurrentFunctions; i += 1 {
		tmpConfig := utils.LoadConfig
		tmpConfig.SQPServerPort = ":" + fmt.Sprint(sQPPort+i)
		log.Infof("starting sQP server no. %d", i+1)
		go sQP.StartServer(tmpConfig)
		time.Sleep(time.Second * 10)
	}
	go knativeQP(config)
	time.Sleep(time.Second * 2)
	go sdk.StartDstServer(config, handler)
	time.Sleep(time.Second * 2)

	start := time.Now()
	log.Infof("starting integ test")
	url := config.ProxyHostname + config.ProxyPort
	numberOfSources := *numConcurrentFunctions
	errChannel := make(chan error, numberOfSources)

	for i := 0; i < numberOfSources; i += 1 {
		i := i
		go func() {
			errChannel <- sdk.InvokeWithXDT(url, preparePayload(), ":"+fmt.Sprint(sQPPort+i), chunkSizeInBytes)
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
		shutdown, err := tracing.InitBasicTracer(*zipkinURL, "xdt")
		if err != nil {
			log.Warn(err)
		}
		defer shutdown()
	}

	// start server at sQP
	dQPPort := 50009
	for i := 0; i < *numConcurrentFunctions; i += 1 {
		config.DstServerPort = ":" + fmt.Sprint(dQPPort+i)
		config.DQPServerPort = ":" + fmt.Sprint(dQPPort+i+*numConcurrentFunctions)
		config.ProxyPort = ":" + fmt.Sprint(dQPPort+i+*numConcurrentFunctions+*numConcurrentFunctions)
		tmpDstConfig := config
		log.Infof("starting Dst server no. %d", i+1)
		go sdk.StartDstServer(tmpDstConfig, handler)
		time.Sleep(time.Second * 10)
		tmpDQPConfig := config
		log.Infof("starting dQP server no. %d", i+1)
		go knativeQP(tmpDQPConfig)
		time.Sleep(time.Second * 10)
	}
	time.Sleep(time.Second * 5)
	go sQP.StartServer(config)

	time.Sleep(time.Second * 1)

	start := time.Now()
	log.Infof("starting integ test")
	numberOfSources := *numConcurrentFunctions
	errChannel := make(chan error, numberOfSources)

	for i := 0; i < numberOfSources; i += 1 {
		url := ":" + fmt.Sprint(dQPPort+i+*numConcurrentFunctions+*numConcurrentFunctions)
		go func() {
			errChannel <- sdk.InvokeWithXDT(url, preparePayload(), config.SQPServerHostname+config.SQPServerPort, chunkSizeInBytes)
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
	go sQP.StartServer(config)
	knativeQP(config)
}

func TestPython_SDKTimeout(t *testing.T) {

	config := utils.LoadConfig
	// start servers
	sQP.StartServer(config)
}

func TestBenchmark_XDT(t *testing.T) {

	config := utils.LoadConfig
	if *sampleSize < 10 {
		log.Fatal("invalid sample size. Acceptable input is integers >= 10")
	}

	go sQP.StartServer(config)
	go knativeQP(config)
	go sdk.StartDstServer(config, handler)

	payloadSizes := []int{10, 100, 1000, 10000, 100000}

	latencyMap := make(map[int][]float64)

	payloadData := make([]byte, 101*1024*1024) // 10MiB
	if _, err := rand.Read(payloadData); err != nil {
		log.Fatal(err)
	}
	url := config.ProxyHostname + config.ProxyPort

	benchPayload := func(payloadSize int, chunkSizeInBytes int, sampleSize int, URL string, payloadData []byte) []float64 {
		var latencies []float64
		payloadToSend := utils.Payload{
			FunctionName: "HelloXDT",
			Data:         payloadData[:payloadSize],
			Key:          "",
		}

		for i := 0; i < sampleSize; i += 1 {
			start := time.Now()
			if err := sdk.InvokeWithXDT(url, payloadToSend, config.SQPServerHostname+config.SQPServerPort, chunkSizeInBytes); err != nil {
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
