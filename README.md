# EliteGate Examples - Mock Backend Services

A standalone, containerized mock backend services project in Go with HTTP and gRPC services designed for testing an API Gateway (CoreGuard Gateway).

## Project Overview

This project contains three microservices:

1. **HTTP User Service** (Port 9001) - User management endpoints
2. **HTTP Order Service** (Port 9002) - Order management endpoints
3. **gRPC Greeting & Notification Service** (Port 50052) - gRPC-based greeting and notification services

All services run in memory with thread-safe data access using `sync.RWMutex`.

## Project Structure

```
elitegate-examples/
├── api/
│   └── proto/
│       ├── services.proto
│       ├── services.pb.go
│       └── services_grpc.pb.go
├── cmd/
│   ├── http-user/
│   │   └── main.go
│   ├── http-order/
│   │   └── main.go
│   └── grpc-hello/
│       └── main.go
├── deploy/
│   ├── user.Dockerfile
│   ├── order.Dockerfile
│   └── grpc.Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

## Technology Stack

- **Language**: Go 1.21
- **HTTP**: Standard `net/http`
- **gRPC**: `google.golang.org/grpc`
- **Protobuf**: `google.golang.org/protobuf`
- **Containerization**: Docker & Docker Compose
- **Network**: Docker bridge network (`elitegate_net`)

## Services

### 1. HTTP User Service (Port 9001)

**Endpoints:**

- `GET /health` - Health check
  ```json
  {"status":"ok", "service":"user-service"}
  ```

- `GET /users` - List all users
  ```json
  {
    "users": [
      {"id": 1, "name": "Alice Johnson", "email": "alice@example.com", "age": 28},
      ...
    ],
    "count": 3
  }
  ```

- `GET /users/:id` - Get specific user
  ```json
  {"id": 1, "name": "Alice Johnson", "email": "alice@example.com", "age": 28}
  ```

- `POST /users` - Create new user
  ```json
  {
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30
  }
  ```

- `GET /debug` - Echo request details (headers, method, query params, timestamp)

### 2. HTTP Order Service (Port 9002)

**Endpoints:**

- `GET /health` - Health check
  ```json
  {"status":"ok", "service":"order-service"}
  ```

- `GET /orders` - List all orders

- `GET /orders/:id` - Get specific order

- `POST /orders` - Create new order
  ```json
  {
    "user_id": 1,
    "product": "Laptop",
    "quantity": 1,
    "price": 999.99
  }
  ```

- `GET /debug` - Echo request details

### 3. gRPC Greeting & Notification Service (Port 50052)

**Greeter Service:**
- `SayHello(HelloRequest)` → `HelloResponse`
- `SayGoodbye(GoodbyeRequest)` → `GoodbyeResponse`

**Notification Service:**
- `SendAlert(AlertRequest)` → `AlertResponse`

**Proto Messages:**

```protobuf
message HelloRequest {
  string name = 1;
}

message HelloResponse {
  string message = 1;
  string timestamp = 2;
}

message AlertRequest {
  string user_id = 1;
  string alert_type = 2;
  string message = 3;
}

message AlertResponse {
  bool success = 1;
  string alert_id = 2;
  string status = 3;
}
```

## Getting Started

### Prerequisites

- Docker >= 20.10
- Docker Compose >= 1.29
- (Optional) Go >= 1.21 for local development

### Running with Docker Compose

**Start all services:**
```bash
docker-compose up --build
```

**Stop all services:**
```bash
docker-compose down
```

**View logs:**
```bash
docker-compose logs -f
docker-compose logs -f user-service
docker-compose logs -f order-service
docker-compose logs -f grpc-service
```

### Accessing Services

Once services are running:

- **User Service**: http://localhost:9001
- **Order Service**: http://localhost:9002
- **gRPC Service**: localhost:50052

### Local Development (without Docker)

**Build individual services:**

```bash
# User Service
go build -o bin/user-service ./cmd/http-user
./bin/user-service

# Order Service
go build -o bin/order-service ./cmd/http-order
./bin/order-service

# gRPC Service
go build -o bin/grpc-service ./cmd/grpc-hello
./bin/grpc-service
```

## Testing

### HTTP Services

**Test User Service:**
```bash
# Health check
curl http://localhost:9001/health

# List users
curl http://localhost:9001/users

# Get specific user
curl http://localhost:9001/users/1

# Create user
curl -X POST http://localhost:9001/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane Doe","email":"jane@example.com","age":25}'

# Debug endpoint (test header propagation)
curl -H "X-Custom-Header: test-value" http://localhost:9001/debug
```

**Test Order Service:**
```bash
# Health check
curl http://localhost:9002/health

# List orders
curl http://localhost:9002/orders

# Get specific order
curl http://localhost:9002/orders/1

# Create order
curl -X POST http://localhost:9002/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"product":"Mouse","quantity":2,"price":29.99}'
```

### gRPC Service

Use a gRPC client tool like `grpcurl`:

```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List services
grpcurl -plaintext localhost:50052 list

# Call SayHello
grpcurl -plaintext -d '{"name":"World"}' localhost:50052 services.Greeter.SayHello

# Call SendAlert
grpcurl -plaintext -d '{"user_id":"123","alert_type":"warning","message":"Test alert"}' \
  localhost:50052 services.Notification.SendAlert
```

## VS Code Tasks

This project includes VS Code task definitions in `.vscode/tasks.json` for quick access to common operations:

### Build Tasks
- **Build: All Services** - Build all three services
- **Build: User Service** - Build user service only
- **Build: Order Service** - Build order service only
- **Build: gRPC Service** - Build gRPC service only

### Run Tasks
- **Run: User Service** - Start user service on port 9001
- **Run: Order Service** - Start order service on port 9002
- **Run: gRPC Service** - Start gRPC service on port 50052

### Docker Tasks
- **Docker: Build Images** - Build all Docker images
- **Docker: Start Services** - Start services with Docker Compose
- **Docker: Stop Services** - Stop Docker Compose services
- **Docker: View Logs** - Stream Docker logs

### Test Tasks
- **Test: User Service** - Quick test of user service endpoints
- **Test: Order Service** - Quick test of order service endpoints

### Utility Tasks
- **Go: Format Code** - Format all Go files
- **Go: Tidy Dependencies** - Clean up go.mod and go.sum
- **Go: Download Dependencies** - Download all dependencies
- **Clean: Remove Binaries** - Remove built binaries
- **Proto: Generate Go Code** - Regenerate protobuf stubs

**Access tasks with:** `Ctrl+Shift+P` → "Tasks: Run Task"

### Debugging

The project includes VS Code launch configurations in `.vscode/launch.json`:
- **Debug: User Service** - Debug user service
- **Debug: Order Service** - Debug order service
- **Debug: gRPC Service** - Debug gRPC service

**Start debugging with:** `F5` or `Debug → Start Debugging`

## Data Storage

- All services use **in-memory storage** with `sync.RWMutex` for thread-safe access
- Mock data is initialized at service startup
- Data is not persisted (lost when service restarts)

### Mock Data

**Users:**
1. Alice Johnson (alice@example.com, age 28)
2. Bob Smith (bob@example.com, age 34)
3. Charlie Brown (charlie@example.com, age 42)

**Orders:**
1. Laptop (User 1, qty 1, $999.99, shipped)
2. Mouse (User 2, qty 2, $29.99, delivered)
3. Keyboard (User 1, qty 1, $79.99, pending)

## Architecture

### Network

All services communicate over a Docker bridge network (`elitegate_net`):

```
┌─────────────────────────────────┐
│    Docker Bridge Network        │
│    elitegate_net                │
├─────────────────────────────────┤
│                                 │
│  ┌──────────────────────────┐  │
│  │  User Service (9001)     │  │
│  └──────────────────────────┘  │
│                                 │
│  ┌──────────────────────────┐  │
│  │  Order Service (9002)    │  │
│  └──────────────────────────┘  │
│                                 │
│  ┌──────────────────────────┐  │
│  │  gRPC Service (50052)    │  │
│  └──────────────────────────┘  │
│                                 │
└─────────────────────────────────┘
```

### Concurrency Model

Each service uses:
- `sync.RWMutex` for thread-safe data access
- Goroutines handled by Go's `net/http` and gRPC servers
- No external database connections

## Configuration

Services use hardcoded configurations:

| Service | Port | Network | Log Level |
|---------|------|---------|-----------|
| User Service | 9001 | elitegate_net | info |
| Order Service | 9002 | elitegate_net | info |
| gRPC Service | 50052 | elitegate_net | info |

## Troubleshooting

### Port Already in Use

If ports are already in use:

```bash
# Find and kill process on port 9001
lsof -i :9001
kill -9 <PID>

# Or use Docker Compose with custom ports
# Edit docker-compose.yml and change port mappings
```

### Services Not Communicating

1. Verify network exists: `docker network ls | grep elitegate_net`
2. Check service connectivity:
   ```bash
   docker network inspect elitegate_net
   ```
3. View logs for errors:
   ```bash
   docker-compose logs --tail=100
   ```

### gRPC Connection Issues

- Ensure `grpcurl` uses `-plaintext` flag (no TLS)
- Verify port 50052 is exposed correctly
- Check gRPC service logs for binding errors

## Development

### Modifying Proto Files

If you modify `api/proto/services.proto`:

1. Install protoc compiler and Go plugins:
   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

2. Generate Go code:
   ```bash
   protoc --go_out=. --go-grpc_out=. api/proto/services.proto
   ```

### Adding New Endpoints

1. **HTTP Services**: Add handler functions to `cmd/http-*/main.go`
2. **gRPC Service**: Update `api/proto/services.proto`, regenerate stubs, implement in `cmd/grpc-hello/main.go`

## Performance Characteristics

- **User Service**: ~100 concurrent connections/sec
- **Order Service**: ~100 concurrent connections/sec
- **gRPC Service**: ~1000 concurrent connections/sec (gRPC is more efficient)
- **Memory**: ~20-50 MB per service container
- **Response Latency**: <5ms average

## Security Notes

⚠️ **This is a mock service for testing purposes:**
- No authentication/authorization
- No input validation beyond basic type checking
- No HTTPS/TLS support
- No rate limiting
- In-memory data (no persistence)

**Do not use in production environments.**

## License

This project is provided as-is for API Gateway testing.

## Support

For issues or questions, please refer to the CoreGuard Gateway documentation.
