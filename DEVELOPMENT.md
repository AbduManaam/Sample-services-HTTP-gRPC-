# Development Guide

This document provides guidance for developers extending and working with the EliteGate Examples mock services.

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Git
- (Optional) grpcurl for testing gRPC services
- (Optional) Protocol Buffer compiler (protoc) for modifying proto definitions

## Project Structure Deep Dive

### Core Components

```
elitegate-examples/
├── api/                          # API definitions and generated code
│   ├── proto/
│   │   ├── services.proto        # Protocol Buffer definitions
│   │   ├── services.pb.go        # Generated message stubs
│   │   └── services_grpc.pb.go   # Generated gRPC stubs
│   └── openapi.yaml              # OpenAPI 3.0 specification
│
├── cmd/                          # Service implementations
│   ├── http-user/main.go         # User service
│   ├── http-order/main.go        # Order service
│   └── grpc-hello/main.go        # gRPC service
│
├── deploy/                       # Container definitions
│   ├── user.Dockerfile
│   ├── order.Dockerfile
│   └── grpc.Dockerfile
│
├── .vscode/                      # VS Code configuration
│   ├── tasks.json                # Build and run tasks
│   ├── launch.json               # Debug configurations
│   ├── settings.json             # Workspace settings
│   └── extensions.json           # Recommended extensions
│
├── docker-compose.yml            # Orchestration
├── go.mod                        # Module definition
├── go.sum                        # Dependency checksums
└── README.md                     # User documentation
```

## Using VS Code Tasks

This project includes comprehensive VS Code task definitions for building, running, and testing:

1. **Open VS Code Command Palette:** `Ctrl+Shift+P`
2. **Type:** "Tasks: Run Task"
3. **Select:** The task you want to run

### Common Tasks

| Task | Purpose |
|------|---------|
| Build: All Services | Compile all services |
| Run: User Service | Start user service locally |
| Run: Order Service | Start order service locally |
| Run: gRPC Service | Start gRPC service locally |
| Docker: Start Services | Start all services with Docker Compose |
| Docker: Stop Services | Stop Docker Compose services |
| Test: User Service | Quick test of endpoints |
| Test: Order Service | Quick test of endpoints |
| Go: Format Code | Format all Go files |
| Go: Tidy Dependencies | Clean up dependencies |

### Debugging in VS Code

1. **Press:** `F5` or `Ctrl+Shift+D`
2. **Select:** A debug configuration (User Service, Order Service, gRPC Service)
3. **Set breakpoints** by clicking on line numbers
4. **Step through code** using VS Code debug controls

## Development Workflow

### Setting Up Your Development Environment

1. **Clone the repository:**
   ```bash
   git clone <repo-url>
   cd elitegate-examples
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   go mod tidy
   ```

3. **Verify setup:**
   ```bash
   make build
   ```

### Making Changes to HTTP Services

#### Adding a New Endpoint to User Service

1. **Edit [cmd/http-user/main.go](cmd/http-user/main.go):**

   ```go
   // Add a new handler function
   func getUsersByAgeHandler(w http.ResponseWriter, r *http.Request) {
       ageStr := r.URL.Query().Get("age")
       age, err := strconv.Atoi(ageStr)
       if err != nil {
           w.WriteHeader(http.StatusBadRequest)
           json.NewEncoder(w).Encode(map[string]string{"error": "invalid age"})
           return
       }

       userStore.mu.RLock()
       var filteredUsers []User
       for _, user := range userStore.users {
           if user.Age == age {
               filteredUsers = append(filteredUsers, user)
           }
       }
       userStore.mu.RUnlock()

       w.Header().Set("Content-Type", "application/json")
       json.NewEncoder(w).Encode(filteredUsers)
   }

   // Add to the router switch statement
   case path == "/users/by-age" && r.Method == http.MethodGet:
       getUsersByAgeHandler(w, r)
   ```

2. **Test locally:**
   ```bash
   make run-user
   # In another terminal:
   curl http://localhost:9001/users/by-age?age=28
   ```

3. **Test with Docker:**
   ```bash
   docker-compose up --build user-service
   curl http://localhost:9001/users/by-age?age=28
   ```

### Making Changes to gRPC Service

#### Adding a New gRPC Method

1. **Update [api/proto/services.proto](api/proto/services.proto):**

   ```protobuf
   service Greeter {
       rpc SayHello (HelloRequest) returns (HelloResponse);
       rpc SayGoodbye (GoodbyeRequest) returns (GoodbyeResponse);
       rpc GetTime (EmptyRequest) returns (TimeResponse);  // New method
   }

   message EmptyRequest {}

   message TimeResponse {
       string current_time = 1;
   }
   ```

2. **Generate Go stubs:**
   ```bash
   make proto-gen
   ```

3. **Implement the method in [cmd/grpc-hello/main.go](cmd/grpc-hello/main.go):**

   ```go
   func (s *GreeterServer) GetTime(ctx context.Context, req *proto.EmptyRequest) (*proto.TimeResponse, error) {
       return &proto.TimeResponse{
           CurrentTime: time.Now().Format(time.RFC3339),
       }, nil
   }
   ```

4. **Test the new method:**
   ```bash
   make run-grpc
   # In another terminal:
   grpcurl -plaintext -d '{}' localhost:50052 services.Greeter.GetTime
   ```

## Testing Strategy

### Unit Testing

Add tests to each service main.go:

```go
// In cmd/http-user/main.go or cmd/http-order/main.go
func TestListUsers(t *testing.T) {
    req, _ := http.NewRequest("GET", "/users", nil)
    w := httptest.NewRecorder()
    router(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
}
```

Run tests:
```bash
go test ./cmd/http-user
go test ./cmd/http-order
go test ./cmd/grpc-hello
```

### Integration Testing

#### Testing All Services

```bash
# Start all services
make docker-up

# Test User Service
make test-user

# Test Order Service
make test-order

# Test gRPC Service
make test-grpc

# Stop services
make docker-down
```

#### Load Testing

Using Apache Bench:

```bash
# User Service endpoints
ab -n 1000 -c 10 http://localhost:9001/users

# Order Service endpoints
ab -n 1000 -c 10 http://localhost:9002/orders
```

### Testing with API Gateway

When testing with an API Gateway (CoreGuard Gateway):

1. **Test header propagation** using the `/debug` endpoint:
   ```bash
   curl -H "X-Gateway-ID: test-gw" \
        -H "X-Request-ID: abc123" \
        http://localhost:9001/debug
   ```

2. **Test request routing** by accessing different endpoints:
   ```bash
   curl http://localhost:9001/users
   curl http://localhost:9002/orders
   grpcurl -plaintext localhost:50052 list
   ```

3. **Monitor response times:**
   ```bash
   curl -w "\nTotal: %{time_total}s\n" http://localhost:9001/users
   ```

## Data Management

### Mock Data Initialization

All services initialize with mock data in their `init()` functions:

```go
func init() {
    userStore.users[1] = User{ID: 1, Name: "Alice Johnson", ...}
    // More mock data...
}
```

### Adding More Mock Data

Edit the service's `init()` function:

```go
func init() {
    userStore.users[1] = User{ID: 1, Name: "Alice Johnson", Email: "alice@example.com", Age: 28}
    userStore.users[2] = User{ID: 2, Name: "Your New User", Email: "new@example.com", Age: 25}
    userStore.idGen = 3
}
```

### Thread-Safe Access Pattern

All services use the same pattern for thread-safe data access:

```go
// Reading
userStore.mu.RLock()
user, exists := userStore.users[id]
userStore.mu.RUnlock()

// Writing
userStore.mu.Lock()
userStore.users[newID] = newUser
userStore.mu.Unlock()
```

## Container Management

### Building Individual Containers

```bash
# User Service
docker build -t user-service:latest -f deploy/user.Dockerfile .

# Order Service
docker build -t order-service:latest -f deploy/order.Dockerfile .

# gRPC Service
docker build -t grpc-service:latest -f deploy/grpc.Dockerfile .
```

### Pushing to Registry

```bash
# Tag images
docker tag user-service:latest your-registry/user-service:latest
docker tag order-service:latest your-registry/order-service:latest
docker tag grpc-service:latest your-registry/grpc-service:latest

# Push to registry
docker push your-registry/user-service:latest
docker push your-registry/order-service:latest
docker push your-registry/grpc-service:latest
```

### Container Networking

All containers run on the `elitegate_net` bridge network. To test inter-service communication:

```bash
# Start services
docker-compose up -d

# Test connectivity from user-service container
docker exec user-service curl http://order-service:9002/health

# Test gRPC connectivity
docker exec user-service grpcurl -plaintext grpc-service:50052 list
```

## Debugging

### Local Debugging

#### Using Delve Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Start debugger on User Service
dlv debug ./cmd/http-user/
(dlv) break main.main
(dlv) continue
```

#### Using Printf Debugging

Add logging to services:

```go
import "log"

log.Printf("Received request: method=%s, path=%s", r.Method, r.URL.Path)
```

### Docker Debugging

```bash
# View service logs
docker logs user-service
docker logs -f user-service  # Follow logs

# Execute commands in container
docker exec user-service curl http://localhost:9001/health

# Interactive shell
docker exec -it user-service sh
```

### Verbose Logging

Add debug logging to HTTP services:

```go
func debugLogging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)
        next.ServeHTTP(w, r)
    })
}

// Wrap router
http.HandleFunc("/", debugLogging(http.HandlerFunc(router)))
```

## Performance Optimization

### Benchmarking

```bash
# Create benchmark file: cmd/http-user/main_test.go
func BenchmarkListUsers(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // Call listUsersHandler
    }
}

# Run benchmark
go test -bench=. ./cmd/http-user -benchmem
```

### Memory Profiling

```bash
# Add profiling endpoint to service
import _ "net/http/pprof"

# Access profiling data
curl http://localhost:6060/debug/pprof/heap
```

### CPU Profiling

```bash
# Generate CPU profile
go test -cpuprofile=cpu.prof ./cmd/http-user

# Analyze profile
go tool pprof cpu.prof
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build and Test
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - run: go mod download
      - run: go build -o bin/user-service ./cmd/http-user
      - run: go build -o bin/order-service ./cmd/http-order
      - run: go build -o bin/grpc-service ./cmd/grpc-hello
      - run: docker-compose up -d
      - run: make test-user
      - run: docker-compose down
```

## Extending the Services

### Adding a New HTTP Service

1. Create directory: `mkdir -p cmd/http-new-service`
2. Create [cmd/http-new-service/main.go](cmd/http-new-service/main.go)
3. Create [deploy/new-service.Dockerfile](deploy/new-service.Dockerfile)
4. Update `docker-compose.yml`
5. Update `Makefile`

### Adding a New gRPC Service

1. Add service definition to `api/proto/services.proto`
2. Run `make proto-gen`
3. Implement service in `cmd/grpc-hello/main.go`
4. Register service in `main()`

## Common Issues and Solutions

### Port Already in Use

```bash
# Find process using port
lsof -i :9001

# Kill process
kill -9 <PID>

# Or use Docker Compose with port override
PORT=9010 docker-compose up user-service
```

### gRPC Connection Refused

- Ensure service is running: `docker logs grpc-service`
- Verify port 50052 is exposed: `docker ps`
- Use plaintext flag: `grpcurl -plaintext localhost:50052 list`

### Module Import Errors

```bash
# Update go.mod and go.sum
go get -u ./...
go mod tidy
go mod verify
```

### Docker Build Failures

```bash
# Check Dockerfile syntax
docker build --no-cache -f deploy/user.Dockerfile .

# View detailed build logs
docker-compose build --no-cache user-service --verbose
```

## Best Practices

1. **Always use `defer` for resource cleanup**
   ```go
   resp, err := http.Get(url)
   if err != nil {
       return err
   }
   defer resp.Body.Close()
   ```

2. **Use context for cancellation**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()
   ```

3. **Log significant operations**
   ```go
   log.Printf("Creating user: %s (%s)", user.Name, user.Email)
   ```

4. **Test edge cases**
   - Empty requests
   - Invalid IDs
   - Concurrent access
   - Large payloads

5. **Keep mock data realistic**
   - Use believable names and emails
   - Add variety to test data
   - Include edge cases (very old/young users, etc.)

## Resources

- [Go Documentation](https://golang.org/doc)
- [gRPC Go Tutorial](https://grpc.io/docs/languages/go/)
- [Protocol Buffers](https://developers.google.com/protocol-buffers)
- [Docker Documentation](https://docs.docker.com/)
- [Effective Go](https://golang.org/doc/effective_go)

## Getting Help

- Check the main [README.md](README.md)
- Review service source code for implementation examples
- Check Docker logs for runtime errors
- Use `make help` to see available build targets
