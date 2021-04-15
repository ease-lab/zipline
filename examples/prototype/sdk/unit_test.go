package sdk

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"strconv"
	"testing"
	"time"

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

func TestSDK_to_sQP_data_transfer(t *testing.T) {

	// start server at sQP
	go sqp.StartServer(":50005")

	// create random payload
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

	duration := sdk.PushData(key, payloadByteArray, chunkSizeInBytes)
	log.Printf("sent %d bytes in %s", len(payloadByteArray), duration)
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
	chunkSizeInBytes := 64 * 1024

	payloadToSend := &payload{
		FunctionName: "HelloXDT",
		Data:         payload_data,
		Key:          ""}
	payloadByteArray, _ := json.Marshal(payloadToSend)

	duration := sdk.PushData(key, payloadByteArray, chunkSizeInBytes)

	log.Printf("transferred %d bytes from SrcFn to sQP in %s", len(payloadByteArray), duration)

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
	chunkSizeInBytes := 64 * 1024

	payloadToSend := &payload{
		FunctionName: "HelloXDT",
		Data:         payload_data,
		Key:          ""}
	payloadByteArray, _ := json.Marshal(payloadToSend)

	duration := sdk.PushData(key, payloadByteArray, chunkSizeInBytes)

	log.Printf("transferred %d bytes from SrcFn to sQP in %s", len(payloadByteArray), duration)

	duration, payloadData := dqp.PullDataFromSrcQP(key, chunkSizeInBytes)

	log.Printf("transferred %d bytes from sQP to dQP in %s", len(payloadData), duration)

	duration, payloadData = gx.FetchFromDQP(key, chunkSizeInBytes)

	log.Printf("transferred %d bytes from dQP to DstFn in %s", len(payloadData), duration)

}
