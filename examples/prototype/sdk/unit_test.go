package sdk

import (
	"crypto/rand"
	"flag"
	"strconv"
	"testing"
	"time"

	dqp "github.com/ease-lab/vhive_stealth/examples/prototype/dqp"
	sdk "github.com/ease-lab/vhive_stealth/examples/prototype/sdk"
	sqp "github.com/ease-lab/vhive_stealth/examples/prototype/sqp"
	log "github.com/sirupsen/logrus"
)

var sample_size = flag.Int("sample", 100, "sample_size")
var URL = flag.String("URL", "", "Function URL")

var config = sdk.LoadConfig("../config.json")

func init(){
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05.000000", FullTimestamp: true})
}

func TestSDK_to_sQP_data_transfer(t *testing.T) {

	// start server at sQP
	go sqp.StartServer(config.SQPServerAddr)

	// create random payload
	now := time.Now()
	key := strconv.Itoa(int(now.UnixNano()))
	payloadData := make([]byte, 10*1024*1024) // 10MiB
	//create random blob
	rand.Read(payloadData)
	chunkSizeInBytes := config.ChunkSizeInBytes

	start := time.Now()
	sdk.PushData(key, payloadData, chunkSizeInBytes)
	duration := time.Since(start)
	log.Printf("sent %d bytes in %s", len(payloadData), duration)
}

func TestSQP_to_dQP_data_transfer(t *testing.T) {

	// start server at sQP
	go sqp.StartServer(config.SQPServerAddr)

	// create random payload
	now := time.Now()
	key := strconv.Itoa(int(now.UnixNano()))
	payloadData := make([]byte, 10*1024*1024) // 10MiB
	//create random blob
	rand.Read(payloadData)
	chunkSizeInBytes := config.ChunkSizeInBytes

	start := time.Now()
	sdk.PushData(key, payloadData, chunkSizeInBytes)
	duration := time.Since(start)
	log.Printf("sent %d bytes in %s", len(payloadData), duration)

	log.Printf("transferred %d bytes from SrcFn to sQP in %s", len(payloadData), duration)

	duration, payloadCount := dqp.PullDataFromSrcQP(key, chunkSizeInBytes)

	log.Printf("transferred %d chunks from sQP to dQP in %s", payloadCount, duration)
}

func TestDQP_to_DstFn_data_transfer(t *testing.T) {

	// start server at sQP
	go sqp.StartServer(config.SQPServerAddr)
	go dqp.StartServer(config.DQPServerAddr)

	// create random payload
	now := time.Now()
	key := strconv.Itoa(int(now.UnixNano()))
	payloadData := make([]byte, 10*1024*1024) // 10MiB
	//create random blob
	rand.Read(payloadData)
	chunkSizeInBytes := config.ChunkSizeInBytes

	start := time.Now()
	sdk.PushData(key, payloadData, chunkSizeInBytes)
	duration := time.Since(start)

	log.Printf("transferred %d bytes from SrcFn to sQP in %s", len(payloadData), duration)

	duration, payloadCount := dqp.PullDataFromSrcQP(key, chunkSizeInBytes)

	log.Printf("transferred %d chunks from sQP to dQP in %s", payloadCount, duration)

	duration, payloadCount = sdk.FetchFromDQP(key, chunkSizeInBytes)

	log.Printf("transferred %d chunks from dQP to DstFn in %s", payloadCount, duration)

}
