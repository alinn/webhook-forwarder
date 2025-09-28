package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alinn/webhook-forwarder/lib/client"
	pb "github.com/alinn/webhook-forwarder/proto"
)

func main() {
	var serverAddr = os.Getenv("GRPC_SERVER_URL") // localhost:9090 - gRPC server address
	if serverAddr == "" {
		panic("GRPC_SERVER_URL environment variable not set")
	}
	var localURL = os.Getenv("LOCAL_FORWARD_URL") // http://localhost:3000 - Local URL to forward webhooks to
	if localURL == "" {
		panic("LOCAL_FORWARD_URL environment variable not set")
	}
	var webhookID = os.Getenv("WEBHOOK_ID") // Webhook ID to use (generated if not provided)

	// Create webhook client
	wc, err := client.NewWebhookClient(serverAddr)
	if err != nil {
		log.Fatalf("Failed to create webhook client: %v", err)
	}
	defer wc.Close()

	// Register listener and get webhook details via gRPC
	resp, err := wc.RegisterListener(localURL, webhookID)
	if err != nil {
		log.Fatalf("Failed to register listener: %v", err)
	}

	if !resp.Success {
		log.Fatalf("Failed to register listener: %s", resp.Error)
	}

	fmt.Printf("Webhook registered successfully!\n")
	fmt.Printf("Webhook ID: %s\n", resp.WebhookId)
	fmt.Printf("Webhook URL: %s\n", resp.WebhookUrl)
	if webhookID != "" {
		fmt.Printf("Using provided webhook ID: %s\n", webhookID)
	} else {
		fmt.Printf("Generated new webhook ID: %s\n", resp.WebhookId)
	}
	fmt.Printf("Configure this URL with your 3rd party service\n")
	fmt.Printf("Forwarding webhooks to: %s\n", localURL)
	fmt.Printf("\nPress Ctrl+C to stop...\n")

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start webhook stream in a goroutine
	go func() {
		err := wc.StartWebhookStream(resp.WebhookId, handleWebhook(localURL))
		if err != nil {
			log.Printf("Webhook stream error: %v", err)
		}
	}()

	// Wait for signal
	<-sigChan
	fmt.Println("\nShutting down...")
}

func handleWebhook(localURL string) func(req *pb.WebhookRequest) (*pb.WebhookResponse, error) {
	return func(req *pb.WebhookRequest) (*pb.WebhookResponse, error) {
		log.Printf("Received webhook: %s %s", req.Method, req.Path)

		// Use the local URL from the registration

		httpReq, err := http.NewRequest(req.Method, localURL, bytes.NewReader(req.Body))
		if err != nil {
			return &pb.WebhookResponse{
				RequestId:  req.RequestId,
				StatusCode: 500,
				Body:       []byte(fmt.Sprintf("Error creating request: %v", err)),
			}, nil
		}

		// Set headers
		for key, value := range req.Headers {
			httpReq.Header.Set(key, value)
		}

		// Send request to local server
		httpClient := &http.Client{Timeout: 30 * time.Second}
		resp, err := httpClient.Do(httpReq)
		if err != nil {
			return &pb.WebhookResponse{
				RequestId:  req.RequestId,
				StatusCode: 502,
				Body:       []byte(fmt.Sprintf("Error forwarding request: %v", err)),
			}, nil
		}
		defer resp.Body.Close()

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return &pb.WebhookResponse{
				RequestId:  req.RequestId,
				StatusCode: 500,
				Body:       []byte(fmt.Sprintf("Error reading response: %v", err)),
			}, nil
		}

		// Convert headers
		headers := make(map[string]string)
		for key, values := range resp.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}

		return &pb.WebhookResponse{
			RequestId:  req.RequestId,
			StatusCode: int32(resp.StatusCode),
			Headers:    headers,
			Body:       body,
		}, nil
	}
}
