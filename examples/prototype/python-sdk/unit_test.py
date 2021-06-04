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

from utils import Payload, loadConfig
from source import splitPayload, generate_chunks, PushData, InvokeWithXDT
import logging as log
import random

config = loadConfig()


def test_splitPayload():
    payloadToSplit = Payload(
        FunctionName="foo", Data=b'0123456789', Key="", IsXDT=False)
    key, data = splitPayload(payloadToSplit)
    log.info("Generated Key is %s", key)
    assert data == b'0123456789'


def test_generate_chunks():
    for packet in generate_chunks(payload=b'01234567890', key='secret', chunkSizeInBytes=2):
        log.info(packet)
    # assert data == b'0123456789'


def test_Push_data():
    PushData(key='secret', payload=b'01234567890',
             sQPAddr=config['SQPServerAddr'], chunkSizeInBytes=2)


def test_Invoke_XDT():
    data = random.randbytes(1024 * 1024 *100)
    payload = Payload(FunctionName="foo", Data=data, Key="", IsXDT=False)
    InvokeWithXDT(URL=config['LBAddr'], xdtPayload=payload,
                  sQPAddr=config['SQPServerAddr'], chunkSizeInBytes=65536)


if __name__ == "__main__":
    log.basicConfig(level=log.INFO)
    test_Invoke_XDT()
    print("Everything passed")
