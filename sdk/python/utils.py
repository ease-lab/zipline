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

import json
from environs import Env
import socket


class Payload:
    FunctionName: str
    Data: bytes

    def __init__(self, FunctionName, Data):
        self.FunctionName = FunctionName
        self.Data = Data

    def tobytes(self):
        objectDict = dict()
        objectDict['FunctionName'] = self.FunctionName
        objectDict['Data'] = ''
        return bytes(json.dumps(objectDict), 'utf-8')

    @classmethod
    def loadFromBytes(self, jsonBytes):
        jsonDump = jsonBytes.decode('utf-8')
        objectDict = json.loads(jsonDump)
        return Payload(objectDict['FunctionName'], bytes(objectDict['Data'], 'utf-8'))


def loadConfig():
    env = Env()
    config = dict()
    config["ChunkSizeInBytes"] = env.int("CHUNK_SIZE_IN_BYTES", 65536)
    config["DQPServerHostname"] = env.str("DQP_SERVER_HOSTNAME", "localhost")
    config["DQPServerPort"] = env.str("DQP_SERVER_PORT", ":50006")
    config["DstServerHostname"] = env.str("DST_SERVER_HOSTNAME", "localhost")
    config["DstServerPort"] = env.str("DST_SERVER_PORT", ":50007")
    config["SrcServerHostname"] = env.str("SRC_SERVER_HOSTNAME", "localhost")
    config["SrcServerPort"] = env.str("SRC_SERVER_PORT", ":50004")
    config["ProxyHostname"] = env.str("PROXY_HOSTNAME", "localhost")
    config["ProxyPort"] = env.str("PROXY_PORT", ":50008")
    config["NoCopy"] = env.bool("NO_COPY", True)
    config["TracingEnabled"] = env.bool("TRACING_ENABLED", False)
    config["RPCTimeoutMaxBackoff"] = env.int("RPC_TIMEOUT_MAX_BACK_OFF", 1000)
    config["RPCTimeoutDuration"] = env.int("RPC_TIMEOUT_DURATION", 60000)
    config["RPCRetryDelay"] = env.int("RPC_RETRY_DELAY", 1)
    config["ZipkinEndpoint"] = env.str("ZIPKIN_ENDPOINT", "http://zipkin.istio-system.svc.cluster.local:9411/api/v2/spans")
    return config


STORE_FORWARD = "Store&Forward"
CUT_THROUGH = "CutThrough"


def get_self_ip():
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    try:
        # doesn't even have to be reachable
        s.connect(('10.255.255.255', 1))
        IP = s.getsockname()[0]
    except Exception:
        IP = '127.0.0.1'
    finally:
        s.close()
    return IP
