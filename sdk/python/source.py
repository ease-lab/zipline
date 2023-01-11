# MIT License
import multiprocessing
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
import time
from socket import socket

sys.path.insert(0, os.getcwd()+'/../../proto/downXDT')
sys.path.insert(0, os.getcwd()+'/../../proto/upXDT')
sys.path.insert(0, os.getcwd()+'/../../proto/crossXDT')

from concurrent import futures
import grpc
import logging as log
import upXDT_pb2_grpc
import upXDT_pb2
import downXDT_pb2_grpc
import downXDT_pb2
import crossXDT_pb2_grpc
import crossXDT_pb2
import multiprocessing as mp
import utils
import queue
import threading
import capnp
crossXDT_capnp = capnp.load('crossXDT.py.capnp')
capnp.remove_event_loop()
capnp.create_event_loop(threaded=True)


class StreamDataCapnpImpl(crossXDT_capnp.StreamData.Server):
    def __init__(self, config, payloadDataMap):
        self.config = config
        self.payloadDataMap = payloadDataMap

    def serveData(self, key, _context, **kwargs):
        log.info("SRC: received noCopy get request for key %s", key)
        payloadBytes = self.payloadDataMap[key]
        del self.payloadDataMap[key]
        return payloadBytes

    def serveBroadcastData(self, key, _context, **kwargs):
        log.info("SRC: received noCopy broadcast get request for key %s", key)
        payloadBytes = self.payloadDataMap[key]
        return payloadBytes


def handleConnection(conn, config, payloadDataMap):
    with conn:
        server = capnp.TwoPartyServer(conn, bootstrap=StreamDataCapnpImpl(config=config, payloadDataMap=payloadDataMap))
        server.on_disconnect().wait()


def startCapnpServer(config, payloadDataMap, event):
    # server = capnp.TwoPartyServer("[::]"+config['SrcServerPort'], bootstrap=StreamDataCapnpImpl(config=config, payloadDataMap=payloadDataMap))
    # # mark server as started
    # log.info("[src]: capnp server started")
    # event.set()
    # while True:
    #     server.poll_once()
    #     time.sleep(0.001)
    log.info("init socket")
    with socket() as s:
        log.info("binding socket to %s %d", '0.0.0.0', int(config['SrcServerPort'][1:]))
        s.bind(('0.0.0.0', int(config['SrcServerPort'][1:])))
        s.listen()
        event.set()
        while True:
            conn, addr = s.accept()
            t = threading.Thread(target=handleConnection, args=(conn, config, payloadDataMap,))
            t.daemon = True
            t.start()
            log.info("[src] Thread started, looping back")


class XDTclient:
    def __init__(self, config):
        self.config = config
        self.ip = utils.get_self_ip() + config["SQPServerPort"]
        self.atom = 0
        self._lock = threading.Lock()
        self.payloadDataMap = dict()
        if config["NoCopy"]:
            self.ip = utils.get_self_ip() + config["SrcServerPort"]
            log.info("[src] starting the host server")
            event = threading.Event()
            thread = threading.Thread(target=startCapnpServer, args=(config, self.payloadDataMap, event,))
            thread.daemon = True
            thread.start()
            # wait for the GRPC server to start
            event.wait()

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

        sQPAddr = self.ip
        log.info("Src: Sourcing the payload at at %s", sQPAddr)
        key, payloadData, xdtPayload = self.splitPayload(xdtPayload)
        serialisedPayload = xdtPayload.tobytes()

        metadata = (
            ('is_xdt', 'true'),
            ('key', key),
            ('sqp_addr', sQPAddr),
            ('routing', self.config['Routing']),
        )

        if not self.config["NoCopy"]:
            mpQueue = mp.Queue()
            thread = threading.Thread(target=PushData, args=(metadata, key, payloadData, sQPAddr, self.config["ChunkSizeInBytes"], mpQueue,))

            if self.config['Routing'] == utils.CUT_THROUGH:
                log.info("SDK: using CutThrough routing")
                thread.start()
            elif self.config['Routing'] == utils.STORE_FORWARD:
                log.info("SDK: using store & forward routing")
                PushData(metadata, key, payloadData, sQPAddr, self.config["ChunkSizeInBytes"])
            else:
                log.fatal("SDK: invalid routing specified in config")
        else:
            self.payloadDataMap[key] = payloadData

        response = fnInvocationCall(URL, serialisedPayload, metadata, self.config)
        if not self.config["NoCopy"]:
            if self.config['Routing'] == utils.CUT_THROUGH:
                try:
                    err = mpQueue.get(block=True, timeout=self.config['RPCTimeoutDuration']/1000)
                    if err is not None:
                        raise err
                except queue.Empty:
                    raise grpc.RpcError
                thread.join()
        return response.message, response.ok

    # Put uploads the data to sQP and returns key and sQP address
    def Put(self, payload):
        payloadLocation = self.config["SQPServerHostname"]+self.config["SQPServerPort"]
        key, payloadData, _ = self.splitPayload(utils.Payload(FunctionName="foo", Data=payload))
        if self.config["NoCopy"]:
            # payloadLocation = self.config["SrcServerHostname"]+self.config["SrcServerPort"]
            self.payloadDataMap[key] = payloadData
            return key

        metadata = (
            ('is_xdt', 'true'),
            ('key', key),
            ('sqp_addr', payloadLocation),
            ('routing', utils.STORE_FORWARD),
        )
        if not self.config["NoCopy"]:
            PushData(metadata, key, payloadData, payloadLocation, self.config["ChunkSizeInBytes"])
        return key

    # Put uploads the data to sQP and returns key and sQP address
    def BroadcastPut(self, payload):
        payloadLocation = self.config["SQPServerHostname"]+self.config["SQPServerPort"]
        key, payloadData, _ = self.splitPayload(utils.Payload(FunctionName="foo", Data=payload))
        if self.config["NoCopy"]:
            # payloadLocation = self.config["SrcServerHostname"]+self.config["SrcServerPort"]
            self.payloadDataMap[key] = payloadData
            return key

        metadata = (
            ('is_xdt', 'true'),
            ('key', key),
            ('sqp_addr', payloadLocation),
            ('routing', utils.STORE_FORWARD),
        )
        if not self.config["NoCopy"]:
            PushBroadcastData(metadata, key, payloadData, payloadLocation, self.config["ChunkSizeInBytes"])
        return key


# fnInvocationCall makes fn invocation call to dQP with xdt payload
def fnInvocationCall(URL, serialisedPayload, metadata, config):

    channel = grpc.insecure_channel(URL)
    channel_ready_future = grpc.channel_ready_future(channel)
    try:
        channel_ready_future.result(timeout=config['RPCTimeoutDuration']/1000)
    except grpc.FutureTimeoutError as e:
        log.error("Src: connection to LB/DQP timed out")
        raise e
    else:
        stub = downXDT_pb2_grpc.XDTtoFnStub(channel)
        response = stub.XDTFnCall(downXDT_pb2.InvocationRequest(
            XDTJSON=serialisedPayload), metadata=metadata, timeout=config['RPCTimeoutDuration']/1000)
        return response


# generate_chunks is a generator for payload stream
def generate_chunks(payload, key, chunkSizeInBytes, noCopy=False):
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
        if noCopy:
            req = crossXDT_pb2.Response(chunk=chunk, TotalChunks=chunkTotal)
        else:
            req = upXDT_pb2.Request(chunk=chunk, key=key, TotalChunks=chunkTotal)
        log.debug("Src: pushed %d bytes to sQP", currentByte)
        currentByte += chunkSizeInBytes
        yield req


# PushData to sQP
def PushData(metadata, key, payload, sQPAddr, chunkSizeInBytes, mpQueue=None):

    try:
        with grpc.insecure_channel(sQPAddr) as channel:
            stub = upXDT_pb2_grpc.StreamDataStub(channel)
            payload_iterator = generate_chunks(payload, key, chunkSizeInBytes)
            log.info("Src: iterator defined")
            route_summary = stub.SendData(payload_iterator, metadata=metadata)
            log.info("Src: data pushed to sQP successfully")
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


# PushBroadcastData to sQP
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
