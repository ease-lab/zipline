package sdk

import (
	"context"
	"encoding/json"
	"flag"
	"go.opentelemetry.io/otel/propagation"
	"io/ioutil"
	stockLogger "log"
	"os"

	log "github.com/sirupsen/logrus"

	downXDT "github.com/ease-lab/vhive_stealth/examples/prototype/proto/downXDT"
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
	DstServerAddr string
	SQPServerAddr string
	BufferSize int
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
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatal(err)
	}

	var config Config

	json.Unmarshal(byteValue, &config)

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