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
- SQPServerHostname: Source queue proxy hostname
- SQPServerPort: Source queue proxy port
- DQPServerHostname: Destination queue proxy hostname
- DQPServerPort: Destination queue proxy port
- DstServerHostname: Destination function hostname
- DstServerPort: Destination function port
- ProxyHostname: Proxy server hostname
- ProxyPort: Proxy server port
- CTBufferSize: Number of chunks to buffer inside sQP and dQP for cut-through routing
- NumberOfBuffers: Number of buffer channels to create
- StAndFwBufferSize: Number of chunks to buffer inside sQP and dQP for store and forward routing
- Routing: Routing type [Store&Forward, CutThrough]
- TracingEnabled: Enable tracing using open-telemetry [true,false]
- RPCTimeoutMaxBackoff: Max backoff duration in milliseconds for connection re-dialing
- RPCTimeoutDuration: RPC timeout duration in milliseconds
- RPCRetryDelay: RPC retry delay in miliseconds.
- MaxDstServerThreadsPython: Maximum GRPC destination server threads (python) 
- ZipkinEndpoint: Zipkin endpoint to transfer tracing data
