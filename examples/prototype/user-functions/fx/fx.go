package fx

import (
	"crypto/rand"
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
	payloadData := make([]byte, 10*1024*1024) // 10MiB
	rand.Read(payloadData)

	chunkSizeInBytes := config.ChunkSizeInBytes

	payloadToSend := &sdk.Payload{
		FunctionName: "HelloXDT",
		Data:         payloadData,
		Key:          "",
	}
	//payloadByteArray, _ := json.Marshal(payloadToSend)

	start := time.Now()
	sdk.InvokeWithXDT("", *payloadToSend, chunkSizeInBytes)
	elapsed := time.Since(start)

	log.Printf("completed XDT in %s", elapsed)
}
