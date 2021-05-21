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

package sdk

import (
	"context"
	"crypto/rand"
	"strconv"
	"testing"
	"time"

	"github.com/ease-lab/vhive_stealth/examples/prototype/dqp"
	"github.com/ease-lab/vhive_stealth/examples/prototype/sdk"
	"github.com/ease-lab/vhive_stealth/examples/prototype/sqp"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05.000000", FullTimestamp: true})
}

func TestSDK_to_sQP_data_transfer(t *testing.T) {

	go sqp.StartServer(sdk.LoadedConfig.SQPServerAddr)

	now := time.Now()
	key := strconv.Itoa(int(now.UnixNano()))
	payloadData := make([]byte, 10*1024*1024) // 10MiB

	if _, err := rand.Read(payloadData); err != nil {
		log.Fatal(err)
	}
	chunkSizeInBytes := sdk.LoadedConfig.ChunkSizeInBytes

	start := time.Now()
	sdk.PushData(context.Background(), key, payloadData, chunkSizeInBytes)
	duration := time.Since(start)
	log.Infof("sent %d bytes in %s", len(payloadData), duration)
}

func TestSQP_to_dQP_data_transfer(t *testing.T) {

	go sqp.StartServer(sdk.LoadedConfig.SQPServerAddr)
	go dqp.StartServer(sdk.LoadedConfig.DQPServerAddr)

	now := time.Now()
	key := strconv.Itoa(int(now.UnixNano()))
	payloadData := make([]byte, 10*1024*1024) // 10MiB
	//create random blob
	if _, err := rand.Read(payloadData); err != nil {
		log.Fatal(err)
	}
	chunkSizeInBytes := sdk.LoadedConfig.ChunkSizeInBytes

	start := time.Now()
	sdk.PushData(context.Background(), key, payloadData, chunkSizeInBytes)
	duration := time.Since(start)
	log.Infof("sent %d bytes in %s", len(payloadData), duration)

	log.Infof("transferred %d bytes from SrcFn to sQP in %s", len(payloadData), duration)

	dqp.PullDataFromSrcQP(context.Background(), key, chunkSizeInBytes)

	log.Infof("transferred packet from sQP to dQP in %s", duration)
}

func TestDQP_to_DstFn_data_transfer(t *testing.T) {

	go sqp.StartServer(sdk.LoadedConfig.SQPServerAddr)
	go dqp.StartServer(sdk.LoadedConfig.DQPServerAddr)

	now := time.Now()
	key := strconv.Itoa(int(now.UnixNano()))
	payloadData := make([]byte, 10*1024*1024) // 10MiB
	//create random blob
	if _, err := rand.Read(payloadData); err != nil {
		log.Fatal(err)
	}
	chunkSizeInBytes := sdk.LoadedConfig.ChunkSizeInBytes

	start := time.Now()
	sdk.PushData(context.Background(), key, payloadData, chunkSizeInBytes)
	duration := time.Since(start)

	log.Infof("transferred %d bytes from SrcFn to sQP in %s", len(payloadData), duration)

	start = time.Now()
	dqp.PullDataFromSrcQP(context.Background(), key, chunkSizeInBytes)
	duration = time.Since(start)

	log.Infof("transferred packet from sQP to dQP in %s", duration)

	duration, payloadCount := sdk.FetchFromDQP(context.Background(), key, chunkSizeInBytes)

	log.Infof("transferred %d chunks from dQP to DstFn in %s", payloadCount, duration)

}
