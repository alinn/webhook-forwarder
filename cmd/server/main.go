package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alinn/webhook-forwarder/lib/server"
)

func main() {
	var httpPort = os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	var grpcPort = os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9090"
	}
	grpcServer := server.NewGrpcServer()

	httpServer := server.NewHttpServer(grpcServer)

	go func() {
		log.Printf("Starting gRPC server on port %s", grpcPort)
		if err := grpcServer.StartGRPCServer(grpcPort); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	go func() {
		log.Printf("Starting HTTP server on port %s", httpPort)
		if err := httpServer.ListenAndServe(":" + httpPort); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Printf("Webhook Forwarder Server started!\n")
	fmt.Printf("HTTP server: http://localhost:%s\n", httpPort)
	fmt.Printf("gRPC server: localhost:%s\n", grpcPort)
	fmt.Printf("\nPress Ctrl+C to stop...\n")

	<-sigChan
	fmt.Println("\nShutting down...")
}
