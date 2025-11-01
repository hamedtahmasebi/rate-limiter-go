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

3. **Create a configuration file** (JSON format). The config defines rate limiting rules and persistence settings. See example below.

4. **Start the server** with the path to your configuration file:

   ```sh
   go run main.go /path/to/config.json
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
    string userID = 3;
    uint64 usageAmountReq = 4;
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
- `userID`: The identifier of the user making the request (string)
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
  "userID": "user123",
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

### Configuration File

The service requires a JSON configuration file that defines rate limiting rules. Example configuration:

```json
{
  "rules": [
    {
      "id": "rule1",
      "client_id": "main_client",
      "service_id": "test_service",
      "usage_price": 1,
      "refill_rate_per_second": 1,
      "initial_tokens": 100,
      "max_tokens": 100
    }
  ],
  "persistence_settings": {
    "disabled": false,
    "interval_seconds": 10
  }
}
```

**Configuration Fields:**
- `rules`: Array of rate limiting rules. Each rule defines:
  - `id`: Unique identifier for the rule
  - `client_id`: Client identifier this rule applies to
  - `service_id`: Service identifier this rule applies to
  - `usage_price`: Number of tokens consumed per usage unit
  - `refill_rate_per_second`: Tokens added per second
  - `initial_tokens`: Starting token count for new buckets
  - `max_tokens`: Maximum tokens a bucket can hold (must be > 0)
- `persistence_settings`: Settings for bucket persistence
  - `disabled`: If true, buckets are not persisted to disk
  - `interval_seconds`: How often to save buckets to disk (in seconds)

### Notes

- This project is for personal learning and experimentation.
- Buckets are automatically created when first accessed for a given client, service, and user combination.
- If persistence is enabled, buckets are saved to the `./persistence_files` directory.

---

### License

This project is provided as-is for educational purposes.
