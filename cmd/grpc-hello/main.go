package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/elitegate/elitegate-examples/api/proto"
	"google.golang.org/grpc"
)

// GreeterServer implements the Greeter service
type GreeterServer struct {
	proto.UnimplementedGreeterServer
}

// SayHello implements Greeter.SayHello
func (s *GreeterServer) SayHello(ctx context.Context, req *proto.HelloRequest) (*proto.HelloResponse, error) {
	log.Printf("Received SayHello request: name=%s", req.Name)
	return &proto.HelloResponse{
		Message:   fmt.Sprintf("Hello, %s! Welcome to the gRPC Greeting Service.", req.Name),
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// SayGoodbye implements Greeter.SayGoodbye
func (s *GreeterServer) SayGoodbye(ctx context.Context, req *proto.GoodbyeRequest) (*proto.GoodbyeResponse, error) {
	log.Printf("Received SayGoodbye request: name=%s", req.Name)
	return &proto.GoodbyeResponse{
		Message:   fmt.Sprintf("Goodbye, %s! Thank you for using the gRPC Greeting Service.", req.Name),
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// NotificationServer implements the Notification service
type NotificationServer struct {
	proto.UnimplementedNotificationServer
	mu         sync.RWMutex
	alertIdGen int
}

// SendAlert implements Notification.SendAlert
func (s *NotificationServer) SendAlert(ctx context.Context, req *proto.AlertRequest) (*proto.AlertResponse, error) {
	s.mu.Lock()
	s.alertIdGen++
	alertID := fmt.Sprintf("ALERT-%d", s.alertIdGen)
	s.mu.Unlock()

	log.Printf("Received SendAlert request: user_id=%s, alert_type=%s, message=%s", req.UserId, req.AlertType, req.Message)

	return &proto.AlertResponse{
		Success: true,
		AlertId: alertID,
		Status:  "sent",
	}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen on port 50052: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Register services
	proto.RegisterGreeterServer(grpcServer, &GreeterServer{})
	proto.RegisterNotificationServer(grpcServer, &NotificationServer{alertIdGen: 1000})

	log.Println("gRPC Greeting & Notification Service starting on port 50052")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
