# Webhook Forwarder - Example Usage

This document provides step-by-step examples of how to use the webhook forwarder.

## Prerequisites

1. Build the project:
```bash
make build-all
```

2. Have Python 3 installed (for the test server)

## Example 1: Complete gRPC Flow (Recommended)

This example shows the complete gRPC-based webhook forwarding flow.

### Step 1: Start the Webhook Server

```bash
# Terminal 1
./bin/webhook-server
```

The server will start on:
- HTTP: http://localhost:8080 (for webhook registration and receiving)
- gRPC: localhost:9090 (for client communication)

### Step 2: Start a Local Test Server

```bash
# Terminal 2
python3 test_local_server.py
```

This starts a simple HTTP server on http://localhost:3000 that will receive forwarded webhooks.

### Step 3: Start the gRPC Client

```bash
# Terminal 3 - Auto-generate webhook ID
./bin/webhook-client -server=localhost:9090 -local=http://localhost:3000

# Or with specific webhook ID
./bin/webhook-client -server=localhost:9090 -local=http://localhost:3000 -webhook-id=my-webhook-123
```

The client will:
1. Register a webhook listener via gRPC
2. Establish a gRPC streaming connection to the server
3. Start listening for webhook streams

### Step 4: Configure 3rd Party Service

Use the webhook URL from the client output to configure your 3rd party service (e.g., GitHub, Stripe, etc.) to send webhooks to:

```
http://your-server.com:8080/hook/{webhook_id}
```

### Step 5: Test Webhook Forwarding

When your 3rd party service sends a webhook, it will:
1. Be received by the HTTP server at `/hook/{webhook_id}`
2. Be forwarded via gRPC to the connected client
3. Be forwarded by the client to your local application

For testing purposes, you can simulate a 3rd party webhook:

```bash
# Terminal 4 - Simulate 3rd party webhook
curl -X POST http://localhost:8080/hook/{webhook_id} \
  -H "Content-Type: application/json" \
  -H "X-Test-Header: test-value" \
  -d '{"test": "webhook data", "timestamp": "2024-01-01T12:00:00Z"}'
```

## Example 2: Automated Testing

Run the complete gRPC flow test:

```bash
./test_grpc_flow.sh
```

This script will:
1. Start the test local server
2. Start the webhook server
3. Start the gRPC client
4. Send test webhooks
5. Show the complete flow in action

## Example 2: Using the Test Script

Run the automated test:

```bash
./test_webhook.sh
```

This script will:
1. Start the webhook server
2. Register a webhook listener
3. Send a test webhook
4. Clean up

## Example 3: gRPC Client (Future Implementation)

The gRPC client is implemented but not yet integrated with the HTTP server. To use it:

```bash
# Start the server (Terminal 1)
./bin/webhook-server

# Start the gRPC client (Terminal 2)
./bin/webhook-client -server=localhost:9090 -local=http://localhost:3000
```

## Architecture Overview

```
Third-party Service
        ↓ (HTTP POST)
Webhook Server (HTTP:8080)
        ↓ (gRPC - Future)
gRPC Client
        ↓ (HTTP POST)
Local Application (HTTP:3000)
```

## Current Status

✅ **Completed:**
- HTTP webhook server
- Webhook registration endpoint
- Request forwarding to local URLs
- gRPC client implementation
- gRPC server implementation
- Build system and documentation

🔄 **In Progress:**
- Integration between HTTP server and gRPC server
- Complete gRPC client-server communication

## Next Steps

1. Integrate the gRPC server with the HTTP router
2. Implement proper client-server communication
3. Add authentication and security features
4. Add configuration management
5. Add logging and monitoring

## Troubleshooting

### Build Issues
```bash
# Clean and rebuild
make clean
make build-all
```

### Port Conflicts
If ports 8080 or 9090 are in use, specify different ports:
```bash
./bin/webhook-server -http=8081 -grpc=9091
```

### Connection Issues
Make sure the local test server is running on port 3000 before testing webhook forwarding.
