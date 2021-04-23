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
	go sqp.StartServer(":50005")

	// create random payload
	now := time.Now()
	key := strconv.Itoa(int(now.UnixNano()))
	payload_data := make([]byte, 10*1024*1024) // 10MiB
	//create random blob
	rand.Read(payload_data)
	chunkSizeInBytes := config.ChunkSizeInBytes

	duration := sdk.PushData(key, payload_data, chunkSizeInBytes)
	log.Printf("sent %d bytes in %s", len(payload_data), duration)
}

func TestSQP_to_dQP_data_transfer(t *testing.T) {

	// start server at sQP
	go sqp.StartServer(":50005")

	// create random payload
	now := time.Now()
	key := strconv.Itoa(int(now.UnixNano()))
	payload_data := make([]byte, 10*1024*1024) // 10MiB
	//create random blob
	rand.Read(payload_data)
	chunkSizeInBytes := config.ChunkSizeInBytes

	duration := sdk.PushData(key, payload_data, chunkSizeInBytes)

	log.Printf("transferred %d bytes from SrcFn to sQP in %s", len(payload_data), duration)

	duration, payloadData := dqp.PullDataFromSrcQP(key, chunkSizeInBytes)

	log.Printf("transferred %d bytes from sQP to dQP in %s", len(payloadData), duration)
}

func TestDQP_to_DstFn_data_transfer(t *testing.T) {

	// start server at sQP
	go sqp.StartServer(":50005")
	go dqp.StartServer(":50006")

	// create random payload
	now := time.Now()
	key := strconv.Itoa(int(now.UnixNano()))
	payload_data := make([]byte, 10*1024*1024) // 10MiB
	//create random blob
	rand.Read(payload_data)
	chunkSizeInBytes := config.ChunkSizeInBytes

	duration := sdk.PushData(key, payload_data, chunkSizeInBytes)

	log.Printf("transferred %d bytes from SrcFn to sQP in %s", len(payload_data), duration)

	duration, payloadData := dqp.PullDataFromSrcQP(key, chunkSizeInBytes)

	log.Printf("transferred %d bytes from sQP to dQP in %s", len(payloadData), duration)

	duration, payloadData = sdk.FetchFromDQP(key, chunkSizeInBytes)

	log.Printf("transferred %d bytes from dQP to DstFn in %s", len(payloadData), duration)

}
