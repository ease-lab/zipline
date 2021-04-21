package fx

import (
	"crypto/rand"
	"encoding/json"
	"log"

	sdk "github.com/ease-lab/vhive_stealth/examples/prototype/sdk"
)

type payload struct {
	FunctionName string
	Data         []byte
	Key          string
	isXDT        bool
}

func main() {
	payload_data := make([]byte, 10*1024*1024) // 10MiB
	rand.Read(payload_data)

	chunkSizeInBytes := 64 * 1024

	payloadToSend := &payload{
		FunctionName: "HelloXDT",
		Data:         payload_data,
		Key:          "",
	}
	payloadByteArray, _ := json.Marshal(payloadToSend)

	duration := sdk.InvokeWithXDT("", payloadByteArray, chunkSizeInBytes)

	log.Printf("completed XDT in %s", duration)
}
