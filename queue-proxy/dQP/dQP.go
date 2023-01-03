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

package dQP

import (
	"bufio"
	"context"
	"google.golang.org/grpc/metadata"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/ease-lab/vhive-xdt/transport"
	"github.com/ease-lab/vhive-xdt/utils"

	log "github.com/sirupsen/logrus"
)

// bufferPool is responsible for managing bounded buffers of channels to store data
var bufferPool transport.BufferPool

// XDTDataServe is a gRPC server to serve data to the DstFn
func XDTDataServe(conn net.Conn) error {

	//Get the key
	reader := bufio.NewReader(conn)
	for {
		key, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		key = strings.TrimSuffix(key, "\n")
		log.Infof("dQP: data being fetched by DstFn using key : %s", key)

		chunkCount := 0
		var channel chan []byte
		var chunkTotal int64
		// Check whether the first packet has been received at dQP or not
		for {
			if channel, chunkTotal = bufferPool.GetChannel(key); channel != nil {
				log.Infof("dQP: found chunkTotal %d for key %s at dQP", chunkTotal, key)
				break
			}
		}

		chunk := <-channel
		chunkCount, err = conn.Write(chunk)
		if err != nil {
			log.Errorf("[src] error sending bytes to dQP: %v", err)
			bufferPool.FreeChannel(key)
			return nil
		}

		if chunkTotal == int64(chunkCount) {
			log.Infof("[dQP] transfer to DST complete, freeing channel %s ", key)
			bufferPool.FreeChannel(key)
		}
	}
}

// PullDataFromSrcQP pulls data from src QP to dst QP
func PullDataFromSrcQP(ctx context.Context) error {

	headers, _ := metadata.FromOutgoingContext(ctx)
	key := headers["key"][0]
	sQPAddr := headers["sqp_addr"][0]
	payloadSizeString := headers["payload_size_in_bytes"][0]
	payloadSize, err := strconv.Atoi(payloadSizeString)
	if err != nil {
		log.Fatal(err)
	}

	tcpConn, err := net.Dial("tcp", sQPAddr)
	if err != nil {
		log.Fatalf("[dqp]: error dialing to src %v", err)
	}

	defer tcpConn.Close()

	_, err = tcpConn.Write([]byte(key + "\n"))
	if err != nil {
		log.Errorf("dQP: open stream error %v", err)
		return err
	}

	var channel chan []byte
	var buffer = make([]byte, payloadSize)
	bytesRead, err := io.ReadFull(tcpConn, buffer)
	if err != nil {
		log.Errorf("dQP: receive error: %v", err)
		return err
	}
	log.Info("[dqp] received ", buffer[0:9], buffer[len(buffer)-9:])
	if bytesRead != payloadSize {
		log.Errorf("dQP: bytes read: %d, payloadSize %d", bytesRead, payloadSize)
		return err
	}
	log.Debugf("dQP: Received %d bytes", bytesRead)
	channel = bufferPool.CreateChannel()
	log.Debugf("dQP: channel allocated")
	channel <- buffer
	bufferPool.StoreChannel(key, int64(bytesRead), channel)
	log.Debugf("dQP: channel allocated")
	return nil
}

// StartServer starts DstQP server
func StartServer(config utils.Config) {

	bufferPool.Init(config)

	log.Infoln("dQP: start server")
	lis, err := net.Listen("tcp", config.DQPServerPort)
	if err != nil {
		log.Fatalf("src: failed to listen: %v", err)
	}

	conn, err := lis.Accept()
	if err != nil {
		log.Fatal(err)
	}
	go XDTDataServe(conn)
}
