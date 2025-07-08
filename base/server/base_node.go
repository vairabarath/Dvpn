package server

import (
	"context"
	"log"
	"time"

	pb "Base_node/pb"
)

type BaseNodeServer struct {
	pb.UnimplementedBaseNodeServiceServer
}

func NewBaseNodeServer() *BaseNodeServer {
	return &BaseNodeServer{}
}

func (s *BaseNodeServer) RegisterSuperNode(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	log.Printf("Registered super Node: %s [%s] @ %s", req.NodeId, req.Region, req.Ip)

	return &pb.RegisterResponse{
		Success: true,
		Message: "Registration successful",
		AssignedId: req.NodeId,
		RegisteredAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *BaseNodeServer) SuperNodeHeartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.Ack, error) {
	log.Printf("💓 Heartbeat from %s | Peers: %d | Latency: %.1fms | Bandwidth: %.2fMbps",
		req.NodeId, req.ActivePeers, req.AvgLatencyMs, req.BandwidthUsageMbps)

	return &pb.Ack{
		Received: true,
		Message: "Heartbeat received",
	}, nil
}