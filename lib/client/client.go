package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/alinn/webhook-forwarder/proto"
)

type WebhookClient struct {
	conn   *grpc.ClientConn
	client pb.WebhookServiceClient
}

func NewWebhookClient(serverAddr string) (*WebhookClient, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}

	client := pb.NewWebhookServiceClient(conn)

	return &WebhookClient{
		conn:   conn,
		client: client,
	}, nil
}

func (wc *WebhookClient) Close() error {
	return wc.conn.Close()
}

func (wc *WebhookClient) RegisterListener(localURL string, webhookID string) (*pb.RegisterListenerResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &pb.RegisterListenerRequest{
		LocalUrl:  localURL,
		WebhookId: webhookID,
	}

	resp, err := wc.client.RegisterListener(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to register listener: %v", err)
	}

	return resp, nil
}

func (wc *WebhookClient) StartWebhookStream(webhookID string, handler func(*pb.WebhookRequest) (*pb.WebhookResponse, error)) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := &pb.StreamWebhooksRequest{
		WebhookId: webhookID,
	}

	stream, err := wc.client.StreamWebhooks(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to start webhook stream: %v", err)
	}

	log.Printf("Started webhook stream for ID: %s", webhookID)

	for {
		webhookReq, err := stream.Recv()
		if err != nil {
			log.Printf("Error receiving webhook: %v", err)
			return err
		}

		// Process the webhook request
		response, err := handler(webhookReq)
		if err != nil {
			log.Printf("Error processing webhook: %v", err)
			// Send error response
			response = &pb.WebhookResponse{
				RequestId:  webhookReq.RequestId,
				StatusCode: 500,
				Body:       []byte(fmt.Sprintf("Error processing webhook: %v", err)),
			}
		}

		// Send response back to server
		ack, err := wc.client.SendWebhookResponse(ctx, response)
		if err != nil {
			log.Printf("Error sending webhook response: %v", err)
		} else if !ack.Success {
			log.Printf("Server rejected webhook response: %s", ack.Error)
		}
	}
}
