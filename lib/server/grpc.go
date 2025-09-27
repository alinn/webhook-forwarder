package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"google.golang.org/grpc"

	pb "github.com/alinn/webhook-forwarder/proto"
)

type GrpcServer struct {
	pb.UnimplementedWebhookServiceServer
	clients     map[string]*ClientInfo
	clientsLock sync.RWMutex
	httpClient  *http.Client
}

type ClientInfo struct {
	LocalURL     string
	StreamClient pb.WebhookService_StreamWebhooksServer
	Context      context.Context
	Cancel       context.CancelFunc
}

func NewGrpcServer() *GrpcServer {
	return &GrpcServer{
		clients:    make(map[string]*ClientInfo),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (ws *GrpcServer) RegisterListener(ctx context.Context, req *pb.RegisterListenerRequest) (*pb.RegisterListenerResponse, error) {
	var webhookID string

	// Use provided webhook ID or generate a new one
	if req.WebhookId != "" {
		webhookID = req.WebhookId
		log.Printf("Using provided webhook ID: %s", webhookID)
	} else {
		webhookID = generateWebhookID()
		log.Printf("Generated new webhook ID: %s", webhookID)
	}

	// Check if webhook ID already exists
	ws.clientsLock.RLock()
	_, exists := ws.clients[webhookID]
	ws.clientsLock.RUnlock()

	if exists {
		log.Printf("Webhook ID %s already exists, updating local URL", webhookID)
	}

	// Create webhook URL - in production this would use the actual server domain
	webhookURL := fmt.Sprintf("http://localhost:8080/hook/%s", webhookID)

	// Store client info (will be updated when client connects via stream)
	ws.clientsLock.Lock()
	ws.clients[webhookID] = &ClientInfo{
		LocalURL: req.LocalUrl,
	}
	ws.clientsLock.Unlock()

	log.Printf("Registered webhook listener: %s -> %s", webhookID, req.LocalUrl)

	return &pb.RegisterListenerResponse{
		WebhookId:  webhookID,
		WebhookUrl: webhookURL,
		Success:    true,
	}, nil
}

func (ws *GrpcServer) StreamWebhooks(req *pb.StreamWebhooksRequest, stream pb.WebhookService_StreamWebhooksServer) error {
	webhookID := req.WebhookId

	ws.clientsLock.RLock()
	_, exists := ws.clients[webhookID]
	ws.clientsLock.RUnlock()

	if !exists {
		return fmt.Errorf("webhook ID not found: %s", webhookID)
	}

	// Update client info with stream
	ctx, cancel := context.WithCancel(stream.Context())
	ws.clientsLock.Lock()
	ws.clients[webhookID].StreamClient = stream
	ws.clients[webhookID].Context = ctx
	ws.clients[webhookID].Cancel = cancel
	ws.clientsLock.Unlock()

	log.Printf("Client connected for webhook ID: %s", webhookID)

	// Wait for context to be cancelled (client disconnect)
	<-ctx.Done()

	// Clean up
	ws.clientsLock.Lock()
	delete(ws.clients, webhookID)
	ws.clientsLock.Unlock()

	log.Printf("Client disconnected for webhook ID: %s", webhookID)
	return nil
}

func (ws *GrpcServer) SendWebhookResponse(ctx context.Context, resp *pb.WebhookResponse) (*pb.WebhookResponseAck, error) {
	// This would typically be handled by the HTTP server
	// For now, just acknowledge receipt
	return &pb.WebhookResponseAck{
		Success: true,
	}, nil
}

func (ws *GrpcServer) HasClient(webhookID string) bool {
	ws.clientsLock.RLock()
	defer ws.clientsLock.RUnlock()

	clientInfo, exists := ws.clients[webhookID]
	return exists && clientInfo.StreamClient != nil
}

func (ws *GrpcServer) ForwardWebhook(webhookID string, method, path string, headers map[string]string, body []byte) error {
	ws.clientsLock.RLock()
	clientInfo, exists := ws.clients[webhookID]
	ws.clientsLock.RUnlock()

	if !exists {
		return fmt.Errorf("no client connected for webhook ID: %s", webhookID)
	}

	if clientInfo.StreamClient == nil {
		return fmt.Errorf("client stream not established for webhook ID: %s", webhookID)
	}

	// Create webhook request
	webhookReq := &pb.WebhookRequest{
		WebhookId: webhookID,
		Method:    method,
		Path:      path,
		Headers:   headers,
		Body:      body,
		RequestId: generateRequestID(),
	}

	// Send to client
	err := clientInfo.StreamClient.Send(webhookReq)
	if err != nil {
		return fmt.Errorf("failed to send webhook to client: %v", err)
	}

	log.Printf("Forwarded webhook %s to client for webhook ID: %s", webhookReq.RequestId, webhookID)
	return nil
}

func (ws *GrpcServer) StartGRPCServer(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWebhookServiceServer(grpcServer, ws)

	log.Printf("gRPC server starting on port %s", port)
	return grpcServer.Serve(lis)
}

// Helper functions
func generateWebhookID() string {
	return fmt.Sprintf("webhook_%d", time.Now().UnixNano())
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
