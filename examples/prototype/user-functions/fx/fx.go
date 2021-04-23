package fx

import (
	"crypto/rand"
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"

	sdk "github.com/ease-lab/vhive_stealth/examples/prototype/sdk"
)

type payload struct {
	FunctionName string
	Data         []byte
	Key          string
	isXDT        bool
}

var config = sdk.LoadConfig("../config.json")

func main() {
	payload_data := make([]byte, 10*1024*1024) // 10MiB
	rand.Read(payload_data)

	chunkSizeInBytes := config.ChunkSizeInBytes

	payloadToSend := &payload{
		FunctionName: "HelloXDT",
		Data:         payload_data,
		Key:          "",
	}
	payloadByteArray, _ := json.Marshal(payloadToSend)

	start := time.Now()
	sdk.InvokeWithXDT("", payloadByteArray, chunkSizeInBytes)
	elapsed := time.Since(start)

	log.Printf("completed XDT in %s", elapsed)
}
