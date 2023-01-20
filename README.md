### Build

```bash
make local
```

### Run Tests

```bash
make unit-test
make integ-test
```

### Config options

- ChunkSizeInBytes: Size of one chunk in bytes
- SrcServerHostname: Source function hostname
- SrcServerPort: Source function port
- DrcServerHostname: Destination queue proxy hostname
- DQPServerPort: Destination queue proxy port
- DstServerHostname: Destination function hostname
- DstServerPort: Destination function port
- ProxyHostname: Proxy server hostname
- ProxyPort: Proxy server port
- TracingEnabled: Enable tracing using open-telemetry [true,false]
- RPCTimeoutMaxBackoff: Max backoff duration in milliseconds for connection re-dialing
- RPCTimeoutDuration: RPC timeout duration in milliseconds
- RPCRetryDelay: RPC retry delay in miliseconds.
- MaxDstServerThreadsPython: Maximum GRPC destination server threads (python) 
- ZipkinEndpoint: Zipkin endpoint to transfer tracing data
