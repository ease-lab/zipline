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

package tests

import (
	"context"
	"crypto/rand"
	"strconv"
	"testing"
	"time"

	"github.com/ease-lab/vhive-xdt/proto/downXDT"

	"google.golang.org/grpc/metadata"

	sdk "github.com/ease-lab/vhive-xdt/sdk/golang"

	ctrdlog "github.com/containerd/containerd/log"

	"github.com/ease-lab/vhive-xdt/queue-proxy/dQP"
	"github.com/ease-lab/vhive-xdt/queue-proxy/sQP"
	"github.com/ease-lab/vhive-xdt/utils"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: ctrdlog.RFC3339NanoFixed,
		FullTimestamp:   true,
		ForceColors:     true})

	go sQP.StartServer(utils.ReadConfig())
	go dQP.StartServer(utils.ReadConfig())
}

func TestSDK_to_sQP_data_transfer(t *testing.T) {

	now := time.Now()
	key := strconv.Itoa(int(now.UnixNano()))
	payloadData := make([]byte, 10*1024*1024) // 10MiB

	if _, err := rand.Read(payloadData); err != nil {
		log.Fatal(err)
	}
	config := utils.ReadConfig()

	start := time.Now()
	httpMetadata := map[string]string{
		"is_xdt":   "true",
		"key":      key,
		"sqp_addr": config.SQPServerHostname + config.SQPServerPort,
		"routing":  config.Routing,
	}

	xdtClient, err := sdk.NewXDTclient(config)
	if err != nil {
		log.Fatalf("InitXDT failed %v", err)
	}

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(httpMetadata))
	if err := xdtClient.PushData(ctx, key, payloadData); err != nil {
		log.Fatalf("TestSDK_to_sQP_data_transfer failed %v", err)
	}
	duration := time.Since(start)
	log.Infof("sent %d bytes in %s", len(payloadData), duration)
}

func TestSQP_to_dQP_data_transfer(t *testing.T) {

	now := time.Now()
	key := strconv.Itoa(int(now.UnixNano()))
	payloadData := make([]byte, 10*1024*1024) // 10MiB
	//create random blob
	if _, err := rand.Read(payloadData); err != nil {
		log.Fatal(err)
	}
	config := utils.ReadConfig()

	start := time.Now()
	httpMetadata := map[string]string{
		"is_xdt":   "true",
		"key":      key,
		"sqp_addr": config.SQPServerHostname + config.SQPServerPort,
		"routing":  config.Routing,
	}
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(httpMetadata))

	xdtClient, err := sdk.NewXDTclient(config)
	if err != nil {
		log.Fatalf("InitXDT failed %v", err)
	}

	if err := xdtClient.PushData(ctx, key, payloadData); err != nil {
		log.Fatalf("TestSDK_to_sQP_data_transfer failed %v", err)
	}
	duration := time.Since(start)
	log.Infof("sent %d bytes in %s", len(payloadData), duration)

	log.Infof("transferred %d bytes from SrcFn to sQP in %s", len(payloadData), duration)

	err = dQP.PullDataFromSrcQP(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("transferred packet from sQP to dQP in %s", duration)
}

func TestDQP_to_DstFn_data_transfer(t *testing.T) {

	now := time.Now()
	key := strconv.Itoa(int(now.UnixNano()))
	payloadData := make([]byte, 10*1024*1024) // 10MiB
	//create random blob
	if _, err := rand.Read(payloadData); err != nil {
		log.Fatal(err)
	}
	config := utils.ReadConfig()

	start := time.Now()
	httpMetadata := map[string]string{
		"is_xdt":   "true",
		"key":      key,
		"sqp_addr": config.SQPServerHostname + config.SQPServerPort,
		"routing":  config.Routing,
	}
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(httpMetadata))
	xdtClient, err := sdk.NewXDTclient(config)
	if err != nil {
		log.Fatalf("InitXDT failed %v", err)
	}
	if err := xdtClient.PushData(ctx, key, payloadData); err != nil {
		log.Fatalf("TestDQP_to_DstFn_data_transfer failed %v", err)
	}
	duration := time.Since(start)

	log.Infof("transferred %d bytes from SrcFn to sQP in %s", len(payloadData), duration)

	start = time.Now()
	err = dQP.PullDataFromSrcQP(ctx)
	if err != nil {
		log.Fatal(err)
	}
	duration = time.Since(start)

	log.Infof("transferred packet from sQP to dQP in %s", duration)

	// connect to dQP
	conn, err := utils.GetGRPCConn(context.Background(), config.DQPServerHostname+config.DQPServerPort, false)
	if err != nil {
		log.Fatalf("DST: can not connect with dQP server %v", err)
	}
	dstClient := downXDT.NewXDTtoFnClient(conn)

	start = time.Now()
	payloadBytes, err := sdk.FetchFromDQP(context.Background(), key, dstClient, config)
	if err != nil {
		log.Fatalf("FetchFromDQP failed %v", err)
	}
	duration = time.Since(start)

	log.Infof("transferred %d bytes from dQP to DstFn in %s", len(payloadBytes), duration)

}
