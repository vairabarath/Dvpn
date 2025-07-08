package server

import (
	"Super_node/pb"
	"context"
	"log"
	"time"
)


type SuperNodeServer struct {
	pb.UnimplementedSuperNodeServiceServer
	registeredPeers map[string]*pb.PeerRegistrationRequest
}

func NewSupreNodeServer() *SuperNodeServer {
	return &SuperNodeServer{
		registeredPeers: make(map[string]*pb.PeerRegistrationRequest),
	}
}

func (s *SuperNodeServer) RegisterClientPeer(ctx context.Context, req *pb.PeerRegistrationRequest) (*pb.RegisterResponse, error) {
	s.registeredPeers[req.PeerId] = req

	log.Printf("👤 Registered Peer: %s [%s] OS: %s NAT: %s", req.PeerId, req.Region, req.Os, req.NatType)

	return &pb.RegisterResponse{
		Success: true,
		Message: "Client peer registered successfully",
		AssignedId: req.PeerId,
		RegisteredAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *SuperNodeServer) PeerSessionHeartbeat(ctx context.Context, req *pb.PeerSessionHeartbeatRequest) (*pb.Ack, error) {
	log.Printf("💓 Heartbeat from %s (exit: %s) — latency: %dms, loss: %.1f%%, throughput: %.2f Mbps, uptime: %ds",
		req.PeerId, req.ExitPeerId, req.LatencyMs, req.PacketLoss, req.ThroughputMbps, req.SessionUptimeSecs)

	return &pb.Ack{
		Received: true,
		Message: "Heartbeat received",
	}, nil
}