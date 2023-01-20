import sys

from utils import Payload, loadConfig
from source import XDTclient
import os
import logging as log
import argparse

# Create the parser
parser = argparse.ArgumentParser()  # Add an argument
parser.add_argument('--url', type=str, required=True)  # Parse the argument
args = parser.parse_args()  # Print "Hello" + the user input argument

config = loadConfig()
log.basicConfig(stream=sys.stdout, level=log.INFO)
transferSizeinKB = 10


def invoke_XDT(url, data):
    payload = Payload(FunctionName="foo", Data=data)
    xdtClient = XDTclient(config=config)
    message, ok = xdtClient.Invoke(URL=url, xdtPayload=payload)
    log.info("destination returned %s %s", message, ok)
    payload = Payload(FunctionName="foo", Data=data)
    message, ok = xdtClient.Invoke(URL=url, xdtPayload=payload)
    log.info("destination returned %s %s", message, ok)
    return


def main():
    data = bytes(os.urandom(transferSizeinKB*1024))
    invoke_XDT(config["DQPServerHostname"]+config["ProxyPort"], data)
    return


if __name__ == '__main__':
    main()
