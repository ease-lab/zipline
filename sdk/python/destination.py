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
from datetime import datetime
from socket import socket

# adding gRPC sources to the system path
sys.path.insert(0, os.getcwd() + '/../../proto/downXDT')
sys.path.insert(0, os.getcwd() + '/../../proto/crossXDT')

from concurrent import futures
import grpc
import logging as log
import downXDT_pb2_grpc
import downXDT_pb2
import crossXDT_pb2_grpc
import crossXDT_pb2
from utils import loadConfig, STORE_FORWARD
import capnp
crossXDT_capnp = capnp.load('crossXDT.py.capnp')
capnp.remove_event_loop()
capnp.create_event_loop(threaded=True)


# XDTtoFnServicer is to be called by dQP to invoke DstFn
class XDTtoFnServicer(downXDT_pb2_grpc.XDTtoFnServicer):
    def __init__(self, dstHandler, payloadFetcher):
        self.dstHandler = dstHandler
        self.payloadFetcher = payloadFetcher

    def XDTFnCall(self, request, context):
        log.info("DST: received invocation call %s", request.XDTJSON)
        metadict = dict(context.invocation_metadata())
        if metadict['is_xdt'] == "true":
            key = metadict['key']
            # fetch data from dQP
            start = datetime.now()
            log.info("Fetching payload using key %s", key)
            payloadBytes = self.payloadFetcher.FetchFromDQP(key)
            duration = datetime.now() - start
            log.info("XDT pull keys took %f seconds", duration.total_seconds())
            # call destination function
            message, ok = self.dstHandler(payloadBytes)
            return downXDT_pb2.InvocationResponse(message=message, ok=ok)
        else:
            return downXDT_pb2.InvocationResponse(message=b'', ok=False)


class Fetcher:
    def __init__(self, config):
        self.config = config

# FetchFromDQP fetches data from dQP to DstFn
    def FetchFromDQP(self, key):
        # serverAddr = self.config['DQPServerHostname']+self.config['DQPServerPort']
        log.info("[dst] making a call to dqp @ %s using key %s", self.config['DQPServerHostname']+self.config['DQPServerPort'], key)
        # client = capnp.TwoPartyClient(self.config['DQPServerHostname']+self.config['DQPServerPort'])
        with socket() as s:
            s.connect((self.config['DQPServerHostname'].encode(), int(self.config['DQPServerPort'][1:])))
            client = capnp.TwoPartyClient(s)
            packet = client.bootstrap().cast_as(crossXDT_capnp.StreamData)

            request = packet.serveData_request()
            request.key = key

            # Send it, which returns a promise for the result (without blocking).
            get_promise = request.send()
            log.info("[dst] waiting for a response from dqp")
            response = get_promise.wait()
            return response.payload
        # with grpc.insecure_channel(serverAddr) as channel:
        #     stub = downXDT_pb2_grpc.XDTtoFnStub(channel)
        #     chunks = stub.XDTDataServe(request)
        #
        #     payloadBytes = bytearray()
        #     for chunk in chunks:
        #         payloadBytes += chunk.chunk
        #     log.info("DST: payload of length %d received", len(payloadBytes))
        #     return payloadBytes


# Get fetches data from sQP
def Get(capability, config):
    log.info("fetching payload using capability %s", capability)
    key = capability
    splitString = capability.split("|", 1)
    sQPAddr = splitString[1]
    metadata = (
        ('is_xdt', 'true'),
        ('key', key),
        ('sqp_addr', sQPAddr),
        ('routing', STORE_FORWARD),
    )
    client = capnp.TwoPartyClient(sQPAddr)
    packet = client.bootstrap().cast_as(crossXDT_capnp.StreamData)
    log.debug("Getting from server... ")

    request = packet.serveData_request()
    request.key = key

    # Send it, which returns a promise for the result (without blocking).
    get_promise = request.send()
    response = get_promise.wait()
    return response.payload
    # request = crossXDT_pb2.Request(key=key)
    # with grpc.insecure_channel(sQPAddr) as channel:
    #     stub = crossXDT_pb2_grpc.StreamDataStub(channel)
    #     chunks = stub.ServeData(request, metadata=metadata)
    #
    #     payloadBytes = bytearray()
    #     for chunk in chunks:
    #         payloadBytes += chunk.chunk
    #     log.info("DST: payload of length %d received", len(payloadBytes))
    #     return payloadBytes


# Get fetches data from sQP
def BroadcastGet(capability, config):
    log.info("fetching payload using capability %s", capability)
    key = capability
    splitString = capability.split("|", 1)
    sQPAddr = splitString[1]
    metadata = (
        ('is_xdt', 'true'),
        ('key', key),
        ('sqp_addr', sQPAddr),
        ('routing', STORE_FORWARD),
    )
    client = capnp.TwoPartyClient(sQPAddr)
    packet = client.bootstrap().cast_as(crossXDT_capnp.StreamData)
    log.debug("Getting from server... ")

    request = packet.serveBroadcastData_request()
    request.key = key

    # Send it, which returns a promise for the result (without blocking).
    get_promise = request.send()
    response = get_promise.wait()
    log.info("fetched payload of size %s", len(response.payload))
    return response.payload
    # request = crossXDT_pb2.BroadcastRequest(key=key, ChunkSizeInBytes=config["ChunkSizeInBytes"])
    # with grpc.insecure_channel(sQPAddr) as channel:
    #     stub = crossXDT_pb2_grpc.StreamDataStub(channel)
    #     chunks = stub.ServeBroadcastData(request, metadata=metadata)
    #
    #     payloadBytes = bytearray()
    #     for chunk in chunks:
    #         payloadBytes += chunk.chunk
    #     log.info("DST: payload of length %d received", len(payloadBytes))
    #     return payloadBytes


# StartDstServer starts DstQP server
def StartDstServer(config, dstHandler):

    payloadFetcher = Fetcher(config=config)
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=config['MaxDstServerThreadsPython']))
    xdtServicer = XDTtoFnServicer(dstHandler=dstHandler, payloadFetcher=payloadFetcher)
    downXDT_pb2_grpc.add_XDTtoFnServicer_to_server(
        xdtServicer, server)
    server.add_insecure_port("[::]"+config['DstServerPort'])
    server.start()


    server.wait_for_termination()


if __name__ == '__main__':
    log.basicConfig(level=log.INFO)

    def handler(payload):
        log.info("destination received payload of length %d", len(payload))
        return b"sample response", True

    StartDstServer(loadConfig(), handler)
