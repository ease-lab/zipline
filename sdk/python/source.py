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

import os
import sys
# adding gRPC sources to the system path
sys.path.insert(0, os.getcwd()+'/../../proto/downXDT')
sys.path.insert(0, os.getcwd()+'/../../proto/upXDT')

import grpc
import logging as log
import upXDT_pb2_grpc
import upXDT_pb2
import downXDT_pb2_grpc
import downXDT_pb2
import multiprocessing as mp
import utils
import queue
import threading


class XDTclient:
    def __init__(self, config):
        self.config = config
        self.ip = utils.get_self_ip() + config["SQPServerPort"]
        self.atom = 0
        self._lock = threading.Lock()

    def splitPayload(self, xdtPayload):
        key = ""
        with self._lock:
            key = str(self.atom) + "|" + self.ip
            log.info("XDT invoke called with payload size %d", len(xdtPayload.Data))
            self.atom += 1

        payloadData = xdtPayload.Data
        log.info("%s %s", [b for b in payloadData[0:9]], [
                 b for b in payloadData[len(payloadData)-9:]])
        xdtPayload.Data = b''
        return key, payloadData, xdtPayload

    # InvokeWithXDT invokes the RPC call with XDT
    def Invoke(self, URL, xdtPayload):

        sQPAddr = self.config["SQPServerHostname"]+self.config["SQPServerPort"]
        key, payloadData, xdtPayload = self.splitPayload(xdtPayload)
        serialisedPayload = xdtPayload.tobytes()

        metadata = (
            ('is_xdt', 'true'),
            ('key', key),
            ('sqp_addr', sQPAddr),
            ('routing', self.config['Routing']),
        )

        mpQueue = mp.Queue()
        p = mp.Process(target=PushData, args=(metadata, key, payloadData, sQPAddr, self.config["ChunkSizeInBytes"], mpQueue,))

        if self.config['Routing'] == utils.CUT_THROUGH:
            log.info("SDK: using CutThrough routing")
            p.start()
        elif self.config['Routing'] == utils.STORE_FORWARD:
            log.info("SDK: using store & forward routing")
            PushData(metadata, key, payloadData, sQPAddr, self.config["ChunkSizeInBytes"])
        else:
            log.fatal("SDK: invalid routing specified in config")

        response = fnInvocationCall(URL, serialisedPayload, metadata, self.config)
        if self.config['Routing'] == utils.CUT_THROUGH:
            try:
                err = mpQueue.get(block=True, timeout=self.config['RPCTimeoutDuration']/1000)
                if err is not None:
                    raise err
            except queue.Empty:
                raise grpc.RpcError
            p.join()
        return response.message, response.ok

    # Put uploads the data to sQP and returns key and sQP address
    def Put(self, payload):
        sQPAddr = self.config["SQPServerHostname"]+self.config["SQPServerPort"]
        key, payloadData, _ = self.splitPayload(utils.Payload(FunctionName="foo", Data=payload))

        metadata = (
            ('is_xdt', 'true'),
            ('key', key),
            ('sqp_addr', sQPAddr),
            ('routing', utils.STORE_FORWARD),
        )
        PushData(metadata, key, payloadData, sQPAddr, self.config["ChunkSizeInBytes"])
        return key

    # Put uploads the data to sQP and returns key and sQP address
    def BroadcastPut(self, payload):
        sQPAddr = self.config["SQPServerHostname"]+self.config["SQPServerPort"]
        key, payloadData, _ = self.splitPayload(utils.Payload(FunctionName="foo", Data=payload))

        metadata = (
            ('is_xdt', 'true'),
            ('key', key),
            ('sqp_addr', sQPAddr),
            ('routing', utils.STORE_FORWARD),
        )
        PushBroadcastData(metadata, key, payloadData, sQPAddr, self.config["ChunkSizeInBytes"])
        return key


# fnInvocationCall makes fn invocation call to dQP with xdt payload
def fnInvocationCall(URL, serialisedPayload, metadata, config):

    channel = grpc.insecure_channel(URL)
    channel_ready_future = grpc.channel_ready_future(channel)
    try:
        channel_ready_future.result(timeout=config['RPCTimeoutDuration']/1000)
    except grpc.FutureTimeoutError as e:
        log.error("SRC: connection to LB/DQP timed out")
        raise e
    else:
        stub = downXDT_pb2_grpc.XDTtoFnStub(channel)
        response = stub.XDTFnCall(downXDT_pb2.InvocationRequest(
            XDTJSON=serialisedPayload), metadata=metadata, timeout=config['RPCTimeoutDuration']/1000)
        return response


# generate_chunks is a generator for payload stream
def generate_chunks(payload, key, chunkSizeInBytes):
    chunkTotal = int(len(payload) / chunkSizeInBytes)
    if len(payload) % chunkSizeInBytes != 0:
        chunkTotal += 1

    payloadSize = len(payload)
    currentByte = 0
    while currentByte < payloadSize:
        if currentByte+chunkSizeInBytes > payloadSize:
            chunk = payload[currentByte:payloadSize]
        else:
            chunk = payload[currentByte: currentByte+chunkSizeInBytes]
        req = upXDT_pb2.Request(chunk=chunk, key=key, TotalChunks=chunkTotal)
        currentByte += chunkSizeInBytes
        yield req


# PushData to sQP
def PushData(metadata, key, payload, sQPAddr, chunkSizeInBytes, mpQueue=None):

    try:
        with grpc.insecure_channel(sQPAddr) as channel:
            stub = upXDT_pb2_grpc.StreamDataStub(channel)
            payload_iterator = generate_chunks(payload, key, chunkSizeInBytes)
            route_summary = stub.SendData(payload_iterator, metadata=metadata)
            if route_summary == upXDT_pb2.Empty():
                log.info("Src: payload pushed successfully")
            else:
                log.info("Src: err while pushing the data")
                log.info(route_summary)
    except grpc.RpcError as e:
        log.info("Push data timed out")
        if mpQueue is not None:
            mpQueue.put(e)
            return
        else:
            raise e
    else:
        log.info("Push data successful")

        if mpQueue is not None:
            mpQueue.put(None)
        return


# PushData to sQP
def PushBroadcastData(metadata, key, payload, sQPAddr, chunkSizeInBytes, mpQueue=None):

    try:
        with grpc.insecure_channel(sQPAddr) as channel:
            stub = upXDT_pb2_grpc.StreamDataStub(channel)
            payload_iterator = generate_chunks(payload, key, chunkSizeInBytes)
            route_summary = stub.BroadcastUpload(payload_iterator, metadata=metadata)
            if route_summary == upXDT_pb2.Empty():
                log.info("Src: payload pushed successfully")
            else:
                log.info("Src: err while pushing the data")
                log.info(route_summary)
    except grpc.RpcError as e:
        log.info("Push data timed out")
        if mpQueue is not None:
            mpQueue.put(e)
            return
        else:
            raise e
    else:
        log.info("Push data successful")

        if mpQueue is not None:
            mpQueue.put(None)
        return
