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
	"encoding/json"
	"flag"
	"go.opentelemetry.io/otel/propagation"
	"io/ioutil"
	stockLogger "log"
	"muzzammil.xyz/jsonc"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/ease-lab/vhive_stealth/examples/prototype/proto/downXDT"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/zipkin"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Payload struct {
	FunctionName string
	Data         []byte
	Key          string
	IsXDT        bool
}

type Config struct {
	ChunkSizeInBytes int
	DQPServerAddr string
	LBAddr string
	DstServerAddr string
	SQPServerAddr string
	BufferSize int
	NumberOfBuffers int
	StAndFwBufferSize int
	Routing string
	TracingEnabled bool
}

type downXDTServer struct {
	downXDT.UnimplementedXDTtoFnServer
}

var LoadedConfig = LoadConfig("../config.json")

func LoadConfig(file string) Config {
	log.Debugf("Opening JSON file with config: %s\n", file)
	jsonFile, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer func (){
		err = jsonFile.Close()
		if err != nil {
			log.Errorf("SDK: Error closing the config file")
		}
	}()

	jsoncByteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatal(err)
	}

	var config Config
	jsonByteValue := jsonc.ToJSON(jsoncByteValue)

	if err = json.Unmarshal(jsonByteValue, &config); err != nil {
		log.Fatal(err)
	}

	return config
}

// InitTracer creates a new trace provider instance and registers it as global trace provider.
func InitTracer() func() {

	url := flag.String("zipkin", "http://localhost:9411/api/v2/spans", "zipkin url")
	flag.Parse()

	var logger = stockLogger.New(os.Stderr, "zipkin-example", stockLogger.Ldate|stockLogger.Ltime|stockLogger.Llongfile)
	// Create Zipkin Exporter and install it as a global tracer.
	//
	// For demoing purposes, always sample. In a production application, you should
	// configure the sampler to a trace.ParentBased(trace.TraceIDRatioBased) set at the desired
	// ratio.

	exporter, err := zipkin.NewRawExporter(
		*url,
		zipkin.WithLogger(logger),
		zipkin.WithSDKOptions(sdktrace.WithSampler(sdktrace.AlwaysSample())),
	)
	if err != nil {
		log.Fatal(err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSyncer(exporter),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return func() {
		_ = tp.Shutdown(context.Background())
	}
}
