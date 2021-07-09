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

package main

import (
	"flag"

	ctrdlog "github.com/containerd/containerd/log"
	tracing "github.com/ease-lab/vhive/utils/tracing/go"
	log "github.com/sirupsen/logrus"

	sdk "github.com/ease-lab/vhive-xdt/sdk/golang"
	"github.com/ease-lab/vhive-xdt/utils"
)

var handler = func(data []byte) {
	log.Infof("gx: destination handler received data of size %d", len(data))
}

func main() {
	zipkinURL := flag.String("zipkin", "http://localhost:9411/api/v2/spans", "zipkin url")
	flag.Parse()
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: ctrdlog.RFC3339NanoFixed,
		FullTimestamp:   true,
		ForceColors:     true})

	config := utils.ReadConfig()
	if config.TracingEnabled {
		shutdown, err := tracing.InitBasicTracer(*zipkinURL, "dst")
		if err != nil {
			log.Warn(err)
		}
		defer shutdown()
	}
	sdk.StartDstServer(config, handler)
}
