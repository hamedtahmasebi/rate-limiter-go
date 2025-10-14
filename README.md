# Rate Limiter (gRPC, Go) [AI GENERATED README]

This project is a simple gRPC-based rate limiter written in Go. It is a personal practice project designed to help the author learn gRPC and Go more deeply. The service exposes a gRPC API for rate limiting, which can be integrated into distributed systems or used as a standalone service for experimentation.

---

## User Documentation

### Prerequisites

- Go 1.24 or later
- `protoc` (Protocol Buffers compiler) with Go plugins
- `grpcurl` or any gRPC client for testing (optional)

### Running the Service

1. **Clone the repository** and navigate to the project directory.

2. **Compile the proto file** (if needed):

   ```sh
   make compile_proto
   ```

3. **Start the server**:

   ```sh
   go run main.go
   ```

   The service will listen on `localhost:50051` by default.

---

### gRPC API Overview

The service exposes a single gRPC service: `RateLimiter`.

#### Service Definition

```proto
service RateLimiter {
    rpc GetAccessStatus(GetAccessStatusRequest) returns (GetAccessStatusResponse) {}
}

message GetAccessStatusRequest {
    string clientID = 1;
    string serviceID = 2;
    uint64 usageAmountReq = 3;
}

message GetAccessStatusResponse {
    bool   isAllowed = 1;
    uint64 retryAfterSeconds = 2;
}
```

---

### How to Use

#### 1. Connect to the Service

- Host: `localhost`
- Port: `50051`
- Protocol: gRPC

#### 2. Make a Request

Call the `GetAccessStatus` RPC with the following fields:

- `clientID`: Your unique client identifier (string)
- `serviceID`: The identifier of the service you want to access (string)
- `usageAmountReq`: The number of usage units you want to consume (uint64)

#### 3. Interpret the Response

- `isAllowed`: `true` if your request is permitted, `false` otherwise.
- `retryAfterSeconds`: If not allowed, this tells you how many seconds to wait before retrying.

---

### Example Usage

#### Using `grpcurl`

```sh
grpcurl -plaintext -d '{
  "clientID": "main_client",
  "serviceID": "test_service",
  "usageAmountReq": 1
}' localhost:50051 RateLimiter/GetAccessStatus
```

#### Example Response

```json
{
  "isAllowed": true,
  "retryAfterSeconds": 0
}
```

If the request is not allowed:

```json
{
  "isAllowed": false,
  "retryAfterSeconds": 12
}
```

---

### Notes

- This project is for personal learning and experimentation.
- The service and bucket configuration are hardcoded for demonstration.
- For custom setups, you may need to modify the source code.

---

### License

This project is provided as-is for educational purposes.
