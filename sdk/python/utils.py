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


class Payload:
    FunctionName: str
    Data: bytes
    Key: str
    IsXDT: bool

    def __init__(self, FunctionName, Data, Key, IsXDT):
        self.FunctionName = FunctionName
        self.Data = Data
        self.Key = Key
        self.IsXDT = IsXDT

    def tobytes(self):
        objectDict = dict()
        objectDict['FunctionName'] = self.FunctionName
        objectDict['Data'] = ''
        objectDict['Key'] = self.Key
        objectDict['IsXDT'] = self.IsXDT
        return bytes(json.dumps(objectDict), 'utf-8')

    @classmethod
    def loadFromBytes(self, jsonBytes):
        jsonDump = jsonBytes.decode('utf-8')
        objectDict = json.loads(jsonDump)
        return Payload(objectDict['FunctionName'], bytes(objectDict['Data'], 'utf-8'), objectDict['Key'],
                       objectDict['IsXDT'])


def loadConfig():
    env = Env()
    config = dict()
    config["ChunkSizeInBytes"] = env.int("CHUNKSIZEINBYTES", 65536)
    config["SQPServerHostname"] = env.str("SQPSERVERHOSTNAME", "localhost")
    config["SQPServerPort"] = env.str("SQPSERVERPORT", ":50005")
    config["DQPServerHostname"] = env.str("DQPSERVERHOSTNAME", "localhost")
    config["DQPServerPort"] = env.str("DQPSERVERPORT", ":50006")
    config["DstServerHostname"] = env.str("DSTSERVERHOSTNAME", "localhost")
    config["DstServerPort"] = env.str("DSTSERVERPORT", ":50007")
    config["ProxyHostname"] = env.str("PROXYHOSTNAME", "localhost")
    config["ProxyPort"] = env.str("PROXYPORT", ":50008")
    config["CTBufferSize"] = env.int("CTBUFFERSIZE", 25)
    config["NumberOfBuffers"] = env.int("NUMBEROFBUFFERS", 2)
    config["StAndFwBufferSize"] = env.int("STANDFWBUFFERSIZE", 1600)
    config["Routing"] = env.str("ROUTING", "Store&Forward")
    config["TracingEnabled"] = env.bool("TRACINGENABLED", False)
    config["RPCTimeoutMaxBackoff"] = env.int("RPCTIMEOUTMAXBACKOFF", 1000)
    config["RPCTimeoutDuration"] = env.int("RPCTIMEOUTDURATION", 60000)
    config["RPCRetryDelay"] = env.int("RPCRETRYDELAY", 1)
    config["MaxDstServerThreadsPython"] = env.int("MAXDSTSERVERTHREADSPYTHON", 10)
    return config


STORE_FORWARD = "Store&Forward"
CUT_THROUGH = "CutThrough"
