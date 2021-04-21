package gx

import (
	"log"

	sdk "github.com/ease-lab/vhive_stealth/examples/prototype/sdk"
)

type payload struct {
	FunctionName string
	Data         []byte
	Key          string
}

var handler = func(data []byte) {
	log.Printf("destination handler received data of size %d", len(data))
}

func main() {
	sdk.StartDstServer(":50007", handler)
}
