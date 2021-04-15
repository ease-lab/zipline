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

.PHONY: proto_gen
proto_gen:
	GO111MODULE="on" go get google.golang.org/protobuf/cmd/protoc-gen-go \
        google.golang.org/grpc/cmd/protoc-gen-go-grpc

	GO111MODULE="on" PATH="$$PATH:$$(go env GOPATH)/bin" protoc proto/CrossQP.proto \
	    --proto_path=./proto \
	    --go_out=./proto/CrossQP --go-grpc_out=./proto/CrossQP \
	    --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative

	GO111MODULE="on" PATH="$$PATH:$$(go env GOPATH)/bin" protoc proto/FnInvocation.proto \
	    --proto_path=./proto \
	    --go_out=./proto/FnInvocation --go-grpc_out=./proto/FnInvocation \
	    --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative

	GO111MODULE="on" PATH="$$PATH:$$(go env GOPATH)/bin" protoc proto/QPToDstFn.proto \
	    --proto_path=./proto \
	    --go_out=./proto/QPToDstFn --go-grpc_out=./proto/QPToDstFn \
	    --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative

	GO111MODULE="on" PATH="$$PATH:$$(go env GOPATH)/bin" protoc proto/SrcFnToQP.proto \
	    --proto_path=./proto \
	    --go_out=./proto/SrcFnToQP --go-grpc_out=./proto/SrcFnToQP \
	    --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative

local: proto_gen build_local

build_local:
	mkdir -p bins
	cd dQP && go build -o ../bins/dqp
	cd sdk && go build -o ../bins/sdk
	cd sQP && go build -o ../bins/sqp
	cd user-functions && go build -o ../bins/gx

clean:
	rm -rf bins
	
unit-test:
	cd sdk && go test unit_test.go -run TestSDK_to_sQP_data_transfer -v
	cd sdk && go test unit_test.go -run TestSQP_to_dQP_data_transfer -v
	cd sdk && go test unit_test.go -run TestDQP_to_DstFn_data_transfer -v

