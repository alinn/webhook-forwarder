package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type HttpServer struct {
	grpcServer *GrpcServer
}

func NewHttpServer(grpcServer *GrpcServer) *HttpServer {
	return &HttpServer{
		grpcServer: grpcServer,
	}
}

func (s *HttpServer) ListenAndServe(addr string) error {
	http.HandleFunc("POST /hook/{hookId}", s.Webhook)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	return http.ListenAndServe(addr, nil)
}

type Challenge struct {
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
}

func (s *HttpServer) Webhook(w http.ResponseWriter, r *http.Request) {
	hookID := r.PathValue("hookId")

	// Check if we have a gRPC server and if the client exists
	if s.grpcServer == nil {
		http.Error(w, "gRPC server not available", http.StatusInternalServerError)
		return
	}

	if !(*s.grpcServer).HasClient(hookID) {
		log.Println("could not find client for hookId:", hookID)
		http.NotFound(w, r)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	challenge := Challenge{}
	err = json.Unmarshal(body, &challenge)
	if err == nil && challenge.Type == "url_verification" {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(challenge.Challenge))
		return
	}

	// Convert headers to map
	headers := make(map[string]string)
	for key, values := range r.Header {
		if key != "Host" && key != "Content-Length" {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
	}

	// Forward the webhook via gRPC
	err = (*s.grpcServer).ForwardWebhook(hookID, r.Method, r.URL.Path, headers, body)
	if err != nil {
		http.Error(w, "Failed to forward webhook via gRPC: "+err.Error(), http.StatusBadGateway)
		return
	}

	// Return a simple response indicating the webhook was forwarded
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "forwarded", "message": "Webhook forwarded via gRPC"}`))
}
