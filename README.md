# Webhook Forwarder

A webhook forwarding service that allows you to receive webhooks from third-party services on your local development machine. The project consists of two parts: a server (deployed in production) and a client (runs locally).

## Architecture

- **Server**: Deployed in production, receives webhooks from third-party services via HTTP and forwards them to connected clients via gRPC streaming
- **Client**: Runs locally, registers webhook listeners via gRPC, establishes streaming connection to the server, and forwards received webhooks to your local application

### Flow:
1. **Client registration**: Client registers a webhook listener via gRPC and gets a webhook URL
2. **gRPC streaming**: Client establishes gRPC streaming connection to server
3. **3rd party webhook**: Third-party service (GitHub, Stripe, etc.) sends webhook to server's HTTP endpoint (`/hook/{webhookId}`)
4. **Server forwarding**: Server forwards webhook via gRPC stream to connected client
5. **Local forwarding**: Client forwards webhook to local application

## Features

- gRPC communication between server and client
- Automatic webhook URL generation
- Request/response forwarding with headers
- Concurrent webhook handling
- Easy deployment and configuration

## Prerequisites

- Go 1.21 or later
- Protocol Buffers compiler (protoc)
- protoc-gen-go and protoc-gen-go-grpc plugins

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd webhook-forwarder
```

2. Install dependencies:
```bash
make deps
```

3. Generate protobuf code:
```bash
make proto
```

4. Build the project:
```bash
make build-all
```

## Usage

### Running the Server

The server should be deployed in your production environment or any publicly accessible infrastructure.

```bash
# Build and run the server
make run-server

# Or run with custom ports
./bin/webhook-server -http=8080 -grpc=9090
```

The server will start:
- HTTP server on port 8080 (for receiving webhooks)
- gRPC server on port 9090 (for client communication)

### Running the Client

The client runs on your local machine and connects to the server.

```bash
# Build and run the client
make run-client

# Or run with custom settings
./bin/webhook-client -server=your-server.com:9090 -local=http://localhost:3000
```

### API Endpoints

#### Webhook Endpoint (3rd Party → Server)

**This endpoint is for third-party services only** (GitHub, Stripe, etc.):

```
POST http://your-server.com:8080/hook/{webhook_id}
```

The client does not interact with this endpoint directly - it only receives webhooks via gRPC streaming.

## Development

### Project Structure

```
webhook-forwarder/
├── cmd/
│   ├── client/          # Client CLI application
│   └── server/          # Server CLI application
├── lib/
│   ├── client/          # gRPC client library
│   ├── router/          # HTTP router for webhook handling
│   └── server/          # gRPC server implementation
├── proto/               # Protocol Buffer definitions
├── bin/                 # Built binaries
└── Makefile            # Build scripts
```

### Building

```bash
# Build everything
make build-all

# Build only server
make build-server

# Build only client
make build-client
```

### Running Tests

```bash
make test
```

### Code Formatting

```bash
make fmt
```

## Configuration

### Server Configuration

- `-http`: HTTP server port (default: 8080)
- `-grpc`: gRPC server port (default: 9090)

### Client Configuration

- `-server`: gRPC server address (default: localhost:9090)
- `-local`: Local URL to forward webhooks to (default: http://localhost:3000)
- `-webhook-id`: Optional webhook ID to use (generated if not provided)

## Example Workflow

1. Deploy the server to your production environment
2. Start the client on your local machine:
   ```bash
   # With auto-generated webhook ID
   ./bin/webhook-client -server=your-server.com:9090 -local=http://localhost:3000
   
   # With specific webhook ID
   ./bin/webhook-client -server=your-server.com:9090 -local=http://localhost:3000 -webhook-id=my-webhook-123
   ```
3. The client will register with the server and receive a webhook URL
4. Configure your third-party service to send webhooks to the provided URL
5. Webhooks will be automatically forwarded to your local application

## Security Considerations

- The current implementation uses insecure gRPC connections
- For production use, implement TLS/SSL for both HTTP and gRPC connections
- Add authentication and authorization mechanisms
- Implement rate limiting and request validation

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and formatting
5. Submit a pull request

## License

[Add your license information here]
