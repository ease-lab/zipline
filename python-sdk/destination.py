# MIT License

# Copyright (c) 2021 Shyam Jesalpura and EASE lab

# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:

# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.

# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

import sys
import os

# adding gRPC sources to the system path
sys.path.insert(0, os.getcwd() + '/../proto/downXDT')

from concurrent import futures
import grpc
import logging as log
import downXDT_pb2_grpc
import downXDT_pb2
from utils import Payload, loadConfig


# XDTtoFnServicer is to be called by dQP to invoke DstFn
class XDTtoFnServicer(downXDT_pb2_grpc.XDTtoFnServicer):
    def XDTFnCall(self, request, context):
        log.info("DST: received invocation call %s", request.XDTJSON)
        xdtPayload = Payload.loadFromBytes(request.XDTJSON)
        key = xdtPayload.Key

        global config
        chunkSizeInBytes = config['ChunkSizeInBytes']

        # fetch data from dQP
        payloadBytes = FetchFromDQP(key, chunkSizeInBytes)

        global dstHandler
        # call destination function
        dstHandler(payloadBytes)
        return downXDT_pb2.Empty()


# FetchFromDQP fetches data from dQP to DstFn
def FetchFromDQP(key, chunkSizeInBytes):
    global config
    serverAddr = config['DQPServerHostname']+config['DQPServerPort']

    request = downXDT_pb2.DataRequest(key=key, ChunkSize=chunkSizeInBytes)
    with grpc.insecure_channel(serverAddr) as channel:
        stub = downXDT_pb2_grpc.XDTtoFnStub(channel)
        chunks = stub.XDTDataServe(request)

        payloadBytes = bytearray()
        for chunk in chunks:
            payloadBytes += chunk.chunk
        log.info("DST: payload of length %d received", len(payloadBytes))
        return payloadBytes


# StartDstServer starts DstQP server
def StartDstServer(config, handler):
    global dstHandler
    dstHandler = handler
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=config['MaxDstServerThreadsPython']))
    downXDT_pb2_grpc.add_XDTtoFnServicer_to_server(
        XDTtoFnServicer(), server)
    server.add_insecure_port("[::]"+config['DstServerPort'])
    server.start()
    server.wait_for_termination()


if __name__ == '__main__':
    log.basicConfig(level=log.INFO)
    global config
    config = loadConfig()


    def handler(payload):
        log.info("destination received payload of length %d", len(payload))


    StartDstServer(config, handler)
