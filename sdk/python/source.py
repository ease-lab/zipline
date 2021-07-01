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
import time
import multiprocessing as mp
import utils
import queue


def splitPayload(xdtPayload):
    now = time.time_ns()
    key = str(now)
    log.info("XDT invoke called with payload size %d", len(xdtPayload.Data))

    payloadData = xdtPayload.Data
    log.info("%s %s", [b for b in payloadData[0:9]], [
             b for b in payloadData[len(payloadData)-9:]])
    xdtPayload.Data = b''
    xdtPayload.Key = key
    xdtPayload.IsXDT = True
    return key, payloadData, xdtPayload


# InvokeWithXDT invokes the RPC call with XDT
def InvokeWithXDT(URL, xdtPayload, sQPAddr, chunkSizeInBytes):

    key, payloadData, xdtPayload = splitPayload(xdtPayload)
    serialisedPayload = xdtPayload.tobytes()

    metadata = (
        ('is_xdt', 'true'),
        ('key', key),
        ('sqp_addr', sQPAddr),
    )

    global config
    config = utils.loadConfig()
    mpQueue = mp.Queue()
    p = mp.Process(target=PushData, args=(key, payloadData, sQPAddr, chunkSizeInBytes, mpQueue,))

    if config['Routing'] == utils.CUT_THROUGH:
        log.info("SDK: using CutThrough routing")
        p.start()
    elif config['Routing'] == utils.STORE_FORWARD:
        log.info("SDK: using store & forward routing")
        PushData(key, payloadData, sQPAddr, chunkSizeInBytes)
    else:
        log.fatal("SDK: invalid routing specified in config")

    fnInvocationCall(URL, serialisedPayload, metadata)
    if config['Routing'] == utils.CUT_THROUGH:
        try:
            err = mpQueue.get(block=True, timeout=config['RPCTimeoutDuration']/1000)
            if err is not None:
                raise err
        except queue.Empty:
            raise grpc.RpcError
        p.join()
    return


# fnInvocationCall makes fn invocation call to dQP with xdt payload
def fnInvocationCall(URL, serialisedPayload, metadata):

    channel = grpc.insecure_channel(URL)
    channel_ready_future = grpc.channel_ready_future(channel)
    global config
    try:
        channel_ready_future.result(timeout=config['RPCTimeoutDuration']/1000)
    except grpc.FutureTimeoutError as e:
        log.error("SRC: connection to LB/DQP timed out")
        raise e
    else:
        stub = downXDT_pb2_grpc.XDTtoFnStub(channel)
        stub.XDTFnCall(downXDT_pb2.InvocationRequest(
            XDTJSON=serialisedPayload), metadata=metadata, timeout=config['RPCTimeoutDuration']/1000)
    return


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
def PushData(key, payload, sQPAddr, chunkSizeInBytes, mpQueue=None):

    try:
        with grpc.insecure_channel(sQPAddr) as channel:
            stub = upXDT_pb2_grpc.StreamDataStub(channel)
            payload_iterator = generate_chunks(payload, key, chunkSizeInBytes)
            route_summary = stub.SendData(payload_iterator)
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
