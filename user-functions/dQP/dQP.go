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
	"net/http"
	"net/http/httputil"
	"os"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	ctrdlog "github.com/containerd/containerd/log"
	tracing "github.com/ease-lab/vhive/utils/tracing/go"
	"github.com/ease-lab/xdt/dQP"
	"github.com/ease-lab/xdt/utils"
	log "github.com/sirupsen/logrus"
	pkgnet "knative.dev/pkg/network"
	"knative.dev/serving/pkg/queue"
)

func main() {
	zipkinURL := flag.String("zipkin", "http://zipkin.istio-system.svc.cluster.local:9411/api/v2/spans", "zipkin url")
	flag.Parse()
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: ctrdlog.RFC3339NanoFixed,
		FullTimestamp:   true,
		ForceColors:     true})

	config := utils.ReadConfig(os.Getenv("KO_DATA_PATH") + "/config.json")
	if config.TracingEnabled {
		shutdown, err := tracing.InitBasicTracer(*zipkinURL, "dQP")
		if err != nil {
			log.Warn(err)
		}
		defer shutdown()
	}
	go dQP.StartServer(config)

	httpProxy := func(target string) *httputil.ReverseProxy {
		return &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = "http"
				req.URL.Host = target

				// Copied from httputil.NewSingleHostReverseProxy.
				if _, ok := req.Header["User-Agent"]; !ok {
					// explicitly disable User-Agent so it's not set to default value
					req.Header.Set("User-Agent", "")
				}
			},
		}
	}(config.DstServerHostname + config.DstServerPort)
	maxIdleConns := 1000 // TODO: somewhat arbitrary value for CC=0, needs experimental validation.
	httpProxy.Transport = pkgnet.NewProxyAutoTransport(maxIdleConns /* max-idle */, maxIdleConns /* max-idle-per-host */)

	var composedHandler http.Handler = httpProxy

	composedHandler = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer h.ServeHTTP(w, r)

			isXDT := r.Header.Get("isXDT")
			if isXDT == "true" {
				log.Infof("pulling from sQP using key %s addr %s", r.Header.Get("key"), r.Header.Get("sQPAddr"))
				err := dQP.PullDataFromSrcQP(r.Context(), r.Header.Get("key"), r.Header.Get("sQPAddr"), config.ChunkSizeInBytes)
				if err != nil {
					log.Errorf("Proxy: unable to pull data from sQP: %v", err)
				}
			}

		})
	}(composedHandler)

	composedHandler = queue.ForwardedShimHandler(composedHandler)

	h2s := &http2.Server{}
	// start server
	server := &http.Server{
		Addr:    config.ProxyPort,
		Handler: h2c.NewHandler(composedHandler, h2s),
	}
	log.Infof("Listening [:50005]...\n")
	err := server.ListenAndServe()
	if err != nil {
		log.Errorf("failed to start proxy server")
	}
}
