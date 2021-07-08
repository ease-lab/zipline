# MIT License
#
# Copyright (c) 2021 Dmitrii Ustiugov and EASE lab
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

all: local

.PHONY: proto_gen proto_install

PROTO_FILES:=crossXDT downXDT upXDT
ROUTING_MODES:=Store&Forward CutThrough
GO_TEST_FLAGS:=-race -v -cover
SDK_GO_FILES:=./source.go ./destination.go

.EXPORT_ALL_VARIABLES:
include user-functions/xdt.env
# network config
SQP_SERVER_HOSTNAME = localhost
SQP_SERVER_PORT = :50005
DQP_SERVER_HOSTNAME = localhost
DQP_SERVER_PORT = :50006
DST_SERVER_HOSTNAME = localhost
DST_SERVER_PORT = :50007
PROXY_HOSTNAME = localhost
PROXY_PORT = :50008

proto_install:
	pip install grpcio-tools --user
	GO111MODULE="on" go get google.golang.org/protobuf/cmd/protoc-gen-go \
            google.golang.org/grpc/cmd/protoc-gen-go-grpc
proto_gen: $(PROTO_FILES)

$(PROTO_FILES):
	python -m grpc_tools.protoc -I./proto --python_out=./proto/$@ --grpc_python_out=./proto/$@ proto/$@.proto
	GO111MODULE="on" PATH="$$PATH:$$(go env GOPATH)/bin" protoc proto/$@.proto \
	    --proto_path=./proto \
	    --go_out=./proto/$@ --go-grpc_out=./proto/$@ \
	    --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative

local: proto_install proto_gen build_local

build_local:
	mkdir -p bins
	cd user-functions/fx && go build -o ../../bins/fx
	cd user-functions/dQP && go build -o ../../bins/dQP
	cd user-functions/sQP && go build -o ../../bins/sQP
	cd user-functions/gx && go build -o ../../bins/gx

clean:
	rm -rf bins

unit-test: export ROUTING = Store&Forward
unit-test:
	cd tests && go test unit_test.go $(GO_TEST_FLAGS)

integ-test_CT: export ROUTING = CutThrough
integ-test_CT:
	cd tests && go test ./integration_test.go -run TestSdk_InvokeWithXDT $(GO_TEST_FLAGS)


integ-test_SF: export ROUTING = Store&Forward
integ-test_SF:
	cd tests && go test ./integration_test.go -run TestSdk_InvokeWithXDT $(GO_TEST_FLAGS)

integ-test: integ-test_CT integ-test_SF

timeout-test_CT: export ROUTING = CutThrough
timeout-test_CT:
	sleep 60
	cd tests && go test ./integration_test.go -run TestErr_DQPTimeout $(GO_TEST_FLAGS)

timeout-test_SF: export ROUTING = Store&Forward
timeout-test_SF:
	sleep 60
	cd tests && go test ./integration_test.go -run TestErr_DQPTimeout $(GO_TEST_FLAGS)

timeout-test: timeout-test_SF timeout-test_CT

parallel-invoke-test_CT: export ROUTING = CutThrough
parallel-invoke-test_CT:
	cd tests && go test ./integration_test.go -run TestParallel_Invoke -concurrentCalls 1 $(GO_TEST_FLAGS)
	cd tests && go test ./integration_test.go -run TestParallel_Invoke -concurrentCalls 2 $(GO_TEST_FLAGS)
	cd tests && go test ./integration_test.go -run TestParallel_Invoke -concurrentCalls 5 $(GO_TEST_FLAGS)

parallel-invoke-test_SF: export ROUTING = Store&Forward
parallel-invoke-test_SF:
	cd tests && go test ./integration_test.go -run TestParallel_Invoke -concurrentCalls 1 $(GO_TEST_FLAGS)
	cd tests && go test ./integration_test.go -run TestParallel_Invoke -concurrentCalls 2 $(GO_TEST_FLAGS)
	cd tests && go test ./integration_test.go -run TestParallel_Invoke -concurrentCalls 5 $(GO_TEST_FLAGS)

parallel-invoke-test: parallel-invoke-test_SF parallel-invoke-test_CT
fan-out-test: fan-out_SF fan-out_CT
fan-in-test: fan-in_SF fan-in_CT

fan-out_SF: export ROUTING = Store&Forward
fan-out_SF:
	cd tests && go test ./integration_test.go -run TestParallel_FanOut -concurrentCalls 1 $(GO_TEST_FLAGS)
	sleep 2
	cd tests && go test ./integration_test.go -run TestParallel_FanOut -concurrentCalls 2 $(GO_TEST_FLAGS)
	sleep 2
	cd tests && go test ./integration_test.go -run TestParallel_FanOut -concurrentCalls 5 $(GO_TEST_FLAGS)
fan-out_CT: export ROUTING = CutThrough
fan-out_CT:
	cd tests && go test ./integration_test.go -run TestParallel_FanOut -concurrentCalls 1 $(GO_TEST_FLAGS)
	sleep 2
	cd tests && go test ./integration_test.go -run TestParallel_FanOut -concurrentCalls 2 $(GO_TEST_FLAGS)
	sleep 2
	cd tests && go test ./integration_test.go -run TestParallel_FanOut -concurrentCalls 5 $(GO_TEST_FLAGS)

fan-in_SF: export ROUTING = Store&Forward
fan-in_SF:
	cd tests && go test ./integration_test.go -run TestParallel_FanIn -concurrentCalls 1 $(GO_TEST_FLAGS)
	sleep 2
	cd tests && go test ./integration_test.go -run TestParallel_FanIn -concurrentCalls 2 $(GO_TEST_FLAGS)
	sleep 2
	cd tests && go test ./integration_test.go -run TestParallel_FanIn -concurrentCalls 5 $(GO_TEST_FLAGS)
fan-in_CT: export ROUTING = CutThrough
fan-in_CT:
	cd tests && go test ./integration_test.go -run TestParallel_FanIn -concurrentCalls 1 $(GO_TEST_FLAGS)
	sleep 2
	cd tests && go test ./integration_test.go -run TestParallel_FanIn -concurrentCalls 2 $(GO_TEST_FLAGS)
	sleep 2
	cd tests && go test ./integration_test.go -run TestParallel_FanIn -concurrentCalls 5 $(GO_TEST_FLAGS)

install_python_modules:
	pip install grpcio --user
	pip install grpcio-tools --user

python-unit-test: export ROUTING = Store&Forward
python-unit-test: install_python_modules
	cd tests && go test ./integration_test.go -run TestPython_SDK $(GO_TEST_FLAGS) &
	sleep 60
	cd sdk/python && python -m unittest -v test.UnitTest
	# kill the process bound to the given port.
	-fuser -k 50005/tcp

python-integ-test: install_python_modules python-integ-test_CT python-integ-test_SF

python-integ-test_CT: export ROUTING = CutThrough
python-integ-test_CT:
	cd tests && go test ./integration_test.go -run TestPython_SDK $(GO_TEST_FLAGS) &
	sleep 60
	cd sdk/python && python destination.py &
	sleep 5
	cd sdk/python && python -m unittest -v test.IntegTest.test_Invoke_XDT
	-fuser -k 50005/tcp
	-fuser -k 50007/tcp

python-integ-test_SF: export ROUTING = Store&Forward
python-integ-test_SF:
	cd tests && go test ./integration_test.go -run TestPython_SDK $(GO_TEST_FLAGS) &
	sleep 60
	cd sdk/python && python destination.py &
	sleep 5
	cd sdk/python && python -m unittest -v test.IntegTest.test_Invoke_XDT
	-fuser -k 50005/tcp
	-fuser -k 50007/tcp

python-timeout-test: install_python_modules python-timeout-test_CT python-timeout-test_SF

python-timeout-test_CT: export ROUTING = CutThrough
python-timeout-test_CT:
	cd tests && go test ./integration_test.go -run TestPython_SDKTimeout $(GO_TEST_FLAGS) &
	sleep 60
	cd sdk/python && python -m unittest -v test.IntegTest.test_Timeout
	-fuser -k 50005/tcp

python-timeout-test_SF: export ROUTING = Store&Forward
python-timeout-test_SF:
	cd tests && go test ./integration_test.go -run TestPython_SDKTimeout $(GO_TEST_FLAGS) &
	sleep 60
	cd sdk/python && python -m unittest -v test.IntegTest.test_Timeout
	-fuser -k 50005/tcp

docker-images-push:
	cd user-functions/fx && ko publish ./ -B
	cd user-functions/dQP && ko publish ./ -B
	cd user-functions/sQP && ko publish ./ -B
	cd user-functions/gx && ko publish ./ -B

benchmark-XDT: export ROUTING = CutThrough
benchmark-XDT:
	cd tests && go test ./integration_test.go -run TestBenchmark_XDT -v