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
