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
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	ctrdlog "github.com/containerd/containerd/log"
	pb "github.com/ease-lab/vhive/examples/protobuf/helloworld"
	"github.com/ease-lab/xdt/utils"
	log "github.com/sirupsen/logrus"

	sdk "github.com/ease-lab/xdt/sdk/go_sdk"
)

type producerServer struct {
	config       utils.Config
	url          string
	transferSize int
	pb.UnimplementedGreeterServer
}

func (ps producerServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	// establish a connection
	sQPAddr := fetchSelfIP() + ps.config.SQPServerPort
	duration := transferPayload(ps.config, ps.url, sQPAddr, ps.transferSize)
	return &pb.HelloReply{Message: fmt.Sprintf("Transferred %d KB in %s", ps.transferSize, duration)}, nil
}

func fetchSelfIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Errorf("Oops: " + err.Error())
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

func transferPayload(config utils.Config, url string, sQPAddr string, transferSize int) time.Duration {
	payloadData := make([]byte, transferSize*1024) // 10MiB
	if _, err := rand.Read(payloadData); err != nil {
		log.Fatal(err)
	}

	chunkSizeInBytes := config.ChunkSizeInBytes
	payloadToSend := utils.Payload{
		FunctionName: "HelloXDT",
		Data:         payloadData,
		Key:          "",
	}

	start := time.Now()
	log.Infof("starting XDT call")
	log.Infof("using %s as the SQP addr", sQPAddr)
	if err := sdk.InvokeWithXDT(url, payloadToSend, sQPAddr, chunkSizeInBytes); err != nil {
		log.Fatalf("SQP_to_dQP_data_transfer failed %v", err)
	}
	return time.Since(start)
}

func main() {
	dockerCompose := flag.Bool("docker-compose", false, "Set to true when used with docker compose")
	url := flag.String("url", "gx.default.192.168.1.240.sslip.io", "Destination function url")
	transferSize := flag.Int("transferSize", 10000, "Number of KB's to transfer")
	flag.Parse()

	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: ctrdlog.RFC3339NanoFixed,
		FullTimestamp:   true,
		ForceColors:     true})

	if *dockerCompose {
		config := utils.ReadConfig(os.Getenv("KO_DATA_PATH") + "/config.json")
		transferPayload(config, "dQP:50006", "sQP:50005", *transferSize)
	} else {
		// load the default config
		config := utils.ReadConfig("")

		var grpcServer *grpc.Server
		if config.TracingEnabled {
			grpcServer = grpc.NewServer(grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()))
		} else {
			grpcServer = grpc.NewServer()
		}
		reflection.Register(grpcServer)

		s := producerServer{}
		s.config = config
		s.url = *url + ":80"
		s.transferSize = *transferSize
		pb.RegisterGreeterServer(grpcServer, &s)

		//server setup
		lis, err := net.Listen("tcp", ":50010")
		if err != nil {
			log.Fatalf("[producer] failed to listen: %v", err)
		}

		log.Println("[producer] Server Started")

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("[producer] failed to serve: %s", err)
		}
	}
}
