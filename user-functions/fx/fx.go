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
	"crypto/rand"
	"time"

	"github.com/ease-lab/xdt/utils"

	log "github.com/sirupsen/logrus"

	"github.com/ease-lab/xdt/sdk"
)

func main() {
	payloadData := make([]byte, 10*1024*1024) // 10MiB
	if _, err := rand.Read(payloadData); err != nil {
		log.Fatal(err)
	}

	chunkSizeInBytes := utils.LoadConfig.ChunkSizeInBytes

	payloadToSend := utils.Payload{
		FunctionName: "HelloXDT",
		Data:         payloadData,
		Key:          "",
	}

	start := time.Now()
	log.Infof("starting XDT call")
	url := utils.LoadConfig.LBAddr
	if err := sdk.InvokeWithXDT(url, payloadToSend, utils.LoadConfig.SQPServerAddr, chunkSizeInBytes); err != nil {
		log.Fatalf("TestSQP_to_dQP_data_transfer failed %v", err)
	}
	elapsed := time.Since(start)

	log.Infof("completed XDT in %s", elapsed)
}
