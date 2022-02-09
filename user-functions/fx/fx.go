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

	tracing "github.com/ease-lab/vhive/utils/tracing/go"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	ctrdlog "github.com/containerd/containerd/log"
	"github.com/ease-lab/vhive-xdt/utils"
	pb "github.com/ease-lab/vhive/examples/protobuf/helloworld"
	log "github.com/sirupsen/logrus"

	sdk "github.com/ease-lab/vhive-xdt/sdk/golang"
)

type producerServer struct {
	config       utils.Config
	url          string
	transferSize int
	noCopy       bool
	pb.UnimplementedGreeterServer
}

func (ps producerServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	// establish a connection
	ps.config.SQPServerHostname = utils.FetchSelfIP()
	var duration time.Duration
	if ps.noCopy {
		duration = transferPayloadNoCopy(ps.config, ps.url, ps.transferSize)
	} else {
		duration = transferPayload(ps.config, ps.url, ps.transferSize)
	}
	return &pb.HelloReply{Message: fmt.Sprintf("Transferred %d KB in %s", ps.transferSize, duration)}, nil
}

func transferPayloadNoCopy(config utils.Config, url string, transferSize int) time.Duration {
	payloadData := make([]byte, transferSize*1024) // 10MiB
	if _, err := rand.Read(payloadData); err != nil {
		log.Fatal(err)
	}

	payloadToSend := utils.Payload{
		FunctionName: "HelloXDT",
		Data:         payloadData,
	}
	xdtClient, err := sdk.NewXDTclient(config)
	if err != nil {
		log.Fatalf("InitXDT failed %v", err)
	}
	start := time.Now()
	log.Infof("starting ServeAndInvoke XDT call")
	log.Infof("using %s as the src addr", config.SrcServerHostname+config.SrcServerPort)
	if message, _, err := xdtClient.ServeAndInvoke(context.Background(), url, payloadToSend); err != nil {
		log.Fatalf("SRC_to_dQP_data_transfer failed %v", err)
	} else {
		log.Infof("received %s from the dest", message)
	}
	return time.Since(start)
}

func transferPayload(config utils.Config, url string, transferSize int) time.Duration {
	payloadData := make([]byte, transferSize*1024) // 10MiB
	if _, err := rand.Read(payloadData); err != nil {
		log.Fatal(err)
	}

	payloadToSend := utils.Payload{
		FunctionName: "HelloXDT",
		Data:         payloadData,
	}
	xdtClient, err := sdk.NewXDTclient(config)
	if err != nil {
		log.Fatalf("InitXDT failed %v", err)
	}
	start := time.Now()
	log.Infof("starting XDT call")
	log.Infof("using %s as the SQP addr", config.SQPServerHostname+config.SQPServerPort)
	if message, _, err := xdtClient.Invoke(context.Background(), url, payloadToSend); err != nil {
		log.Fatalf("SQP_to_dQP_data_transfer failed %v", err)
	} else {
		log.Infof("received %s from the dest", message)
	}
	return time.Since(start)
}

func main() {
	dockerCompose := flag.Bool("dockerCompose", false, "Set to true when used with docker compose")
	url := flag.String("url", "gx.default.192.168.1.240.sslip.io", "Destination function url")
	transferSize := flag.Int("transferSize", 10000, "Number of KB's to transfer")
	noCopy := flag.Bool("noCopy", false, "Set to true to bypass sQP for object transfers")
	flag.Parse()

	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: ctrdlog.RFC3339NanoFixed,
		FullTimestamp:   true,
		ForceColors:     true})

	config := utils.ReadConfig()
	if *dockerCompose {
		if config.TracingEnabled {
			timeout := 1 * time.Second
			for {
				_, err := net.DialTimeout("tcp", "zipkin:9411", timeout)
				if err != nil {
					log.Infof("Site unreachable, error: %v", err)
				} else {
					log.Infof("Zipkin ready, starting the test")
					break
				}
				time.Sleep(time.Second)
			}
			shutdown, err := tracing.InitBasicTracer(config.ZipkinEndpoint, "fx")
			if err != nil {
				log.Warn(err)
			}
			defer shutdown()
		}
		if *noCopy {
			transferPayloadNoCopy(config, config.DQPServerHostname+config.ProxyPort, *transferSize)
		} else {
			transferPayload(config, config.DQPServerHostname+config.ProxyPort, *transferSize)
		}
	} else {
		var grpcServer *grpc.Server
		if config.TracingEnabled {
			shutdown, err := tracing.InitBasicTracer(config.ZipkinEndpoint, "fx")
			if err != nil {
				log.Warn(err)
			}
			defer shutdown()
			grpcServer = grpc.NewServer(grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()))
		} else {
			grpcServer = grpc.NewServer()
		}
		reflection.Register(grpcServer)

		s := producerServer{}
		s.config = config
		s.url = *url + ":80"
		s.transferSize = *transferSize
		s.noCopy = *noCopy
		pb.RegisterGreeterServer(grpcServer, &s)

		//server setup
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
		if err != nil {
			log.Fatalf("[producer] failed to listen: %v", err)
		}

		log.Println("[producer] Server Started")

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("[producer] failed to serve: %s", err)
		}
	}
}
