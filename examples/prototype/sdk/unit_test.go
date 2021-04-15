package sdk

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"strconv"
	"testing"
	"time"

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
