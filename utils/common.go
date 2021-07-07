// MIT License
//
// Copyright (c) 2021 Shyam Jesalpura and EASE lab
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package utils

import (
	"github.com/kelseyhightower/envconfig"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	ChunkSizeInBytes          int    `default:"65536"`
	SQPServerHostname         string `default:"localhost"`
	SQPServerPort             string `default:":50005"`
	DQPServerHostname         string `default:"localhost"`
	DQPServerPort             string `default:":50006"`
	DstServerHostname         string `default:"localhost"`
	DstServerPort             string `default:":50007"`
	ProxyHostname             string `default:"localhost"`
	ProxyPort                 string `default:":50008"`
	CTBufferSize              int    `default:"25"`
	NumberOfBuffers           int    `default:"2"`
	StAndFwBufferSize         int    `default:"1600"`
	Routing                   string `default:"Store&Forward"`
	TracingEnabled            bool   `default:"false"`
	RPCTimeoutMaxBackoff      int    `default:"1000"`
	RPCTimeoutDuration        int    `default:"60000"`
	RPCRetryDelay             int    `default:"1"`
	MaxDstServerThreadsPython int    `default:"10"`
}

type Payload struct {
	FunctionName string
	Data         []byte
}

const (
	STORE_FORWARD = "Store&Forward"
	CUT_THROUGH   = "CutThrough"
)

var LoadConfig = ReadConfig()

func ReadConfig() Config {
	log.Debugf("Loading config from env\n")
	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatalf("Error loding config environment variables: %v", err.Error())
	}
	return config
}
