package gx

import (
	log "github.com/sirupsen/logrus"

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

var config = sdk.LoadConfig("../../config.json")

func main() {
	sdk.StartDstServer(config.DstServerAddr)
}
