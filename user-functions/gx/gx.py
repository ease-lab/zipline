import sys

from utils import loadConfig
import destination as XDTdst

import os
import logging as log
import argparse
config = loadConfig()
log.basicConfig(stream=sys.stdout, level=log.INFO)

def main():
    def handler(payload):
        log.info("[gx]: received %d bytes", len(payload))
        log.info("%s %s", [b for b in payload[0:9]], [
            b for b in payload[len(payload)-9:]])
        return b'yeah', True
    XDTdst.StartDstServer(config, handler)


if __name__ == '__main__':
    main()
