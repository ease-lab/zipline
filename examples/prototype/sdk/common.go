package sdk

import (
	"encoding/json"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"

	downXDT "github.com/ease-lab/vhive_stealth/examples/prototype/proto/downXDT"
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
