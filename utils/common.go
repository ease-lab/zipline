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
	"net"

	"github.com/kelseyhightower/envconfig"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	ChunkSizeInBytes          int    `default:"65536" envconfig:"CHUNK_SIZE_IN_BYTES"`
	SQPServerHostname         string `default:"localhost" envconfig:"SQP_SERVER_HOSTNAME"`
	SQPServerPort             string `default:":50005" envconfig:"SQP_SERVER_PORT"`
	DQPServerHostname         string `default:"localhost" envconfig:"DQP_SERVER_HOSTNAME"`
	DQPServerPort             string `default:":50006" envconfig:"DQP_SERVER_PORT"`
	DstServerHostname         string `default:"localhost" envconfig:"DST_SERVER_HOSTNAME"`
	DstServerPort             string `default:":50007" envconfig:"DST_SERVER_PORT"`
	ProxyHostname             string `default:"localhost" envconfig:"PROXY_HOSTNAME"`
	ProxyPort                 string `default:":50008" envconfig:"PROXY_PORT"`
	CTBufferSize              int    `default:"25" envconfig:"CT_BUFFER_SIZE"`
	NumberOfBuffers           int    `default:"2" envconfig:"NUMBER_OF_BUFFERS"`
	StAndFwBufferSize         int    `default:"1600" envconfig:"ST_AND_FW_BUFFER_SIZE"`
	Routing                   string `default:"CutThrough" envconfig:"ROUTING"`
	TracingEnabled            bool   `default:"false" envconfig:"TRACING_ENABLED"`
	RPCTimeoutMaxBackoff      int    `default:"1000" envconfig:"RPC_TIMEOUT_MAX_BACK_OFF"`
	RPCTimeoutDuration        int    `default:"60000" envconfig:"RPC_TIMEOUT_DURATION"`
	RPCRetryDelay             int    `default:"1" envconfig:"RPC_RETRY_DELAY"`
	MaxDstServerThreadsPython int    `default:"10" envconfig:"MAX_DST_SERVER_THREADS_PYTHON"`
	ZipkinEndpoint            string `default:"http://zipkin.istio-system.svc.cluster.local:9411/api/v2/spans" envconfig:"ZIPKIN_ENDPOINT"`
}

type Payload struct {
	FunctionName string
	Data         []byte
}

const (
	STORE_FORWARD = "Store&Forward"
	CUT_THROUGH   = "CutThrough"
)

func ReadConfig() Config {
	log.Debugf("Loading config from env\n")
	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatalf("Error loding config environment variables: %v", err.Error())
	}
	return config
}

func FetchSelfIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Errorf("Error fetching self IP: " + err.Error())
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	log.Errorf("unable to find IP, returning empty string")
	return ""
}
