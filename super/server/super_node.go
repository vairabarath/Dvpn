package server

import (
	"Super_node/pb"
	"context"
	"log"
	"time"
)

type ClientPeerInfo struct {
	PeerID         string
	PublicKey      string
	Version        string
	Os             string
	Region         string
	NatType        string
	RegisteredAt   string
	LastHeartbeat  time.Time
	SessionUptime  int32
	LatencyMs      int32
	PacketLoss     float32
	ThroughputMbps float32
}

type SuperNodeServer struct {
	pb.UnimplementedSuperNodeServiceServer
	registeredPeers map[string]*ClientPeerInfo
}

func NewSupreNodeServer() *SuperNodeServer {
	return &SuperNodeServer{
		registeredPeers: make(map[string]*ClientPeerInfo),
	}
}

func (s *SuperNodeServer) RegisterClientPeer(ctx context.Context, req *pb.PeerRegistrationRequest) (*pb.RegisterResponse, error) {
	if !verifyClientPeer(
		req.PeerId,
		req.Region,
		req.Os,
		req.NatType,
		req.Nonce,
		req.PublicKey,
		req.Signature,
	) {
		log.Printf("❌ Signature verification failed for peer %s", req.PeerId)
		return &pb.RegisterResponse{
			Success: false,
			Message: "Invalid signature",
		}, nil
	}

	s.registeredPeers[req.PeerId] = &ClientPeerInfo{
		PeerID:        req.PeerId,
		PublicKey:     req.PublicKey,
		Version:       req.Version,
		Os:            req.Os,
		Region:        req.Region,
		NatType:       req.NatType,
		RegisteredAt:  time.Now().Format(time.RFC3339),
		LastHeartbeat: time.Now(),
	}

	log.Printf("👤 Registered Peer: %s [%s] OS: %s NAT: %s", req.PeerId, req.Region, req.Os, req.NatType)

	return &pb.RegisterResponse{
		Success:      true,
		Message:      "Client peer registered successfully",
		AssignedId:   req.PeerId,
		RegisteredAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *SuperNodeServer) PeerSessionHeartbeat(ctx context.Context, req *pb.PeerSessionHeartbeatRequest) (*pb.Ack, error) {
	log.Printf("💓 Heartbeat from %s (exit: %s) — latency: %dms, loss: %.1f%%, throughput: %.2f Mbps, uptime: %ds",
		req.PeerId, req.ExitPeerId, req.LatencyMs, req.PacketLoss, req.ThroughputMbps, req.SessionUptimeSecs)

	return &pb.Ack{
		Received: true,
		Message:  "Heartbeat received",
	}, nil
}

func (s *SuperNodeServer) StartPeerMonitoring() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()
			for peerID, info := range s.registeredPeers {
				age := now.Sub(info.LastHeartbeat)
				if age > 2*time.Minute {
					log.Printf("❌ Peer %s is stale, last heartbeat: %s", peerID, info.LastHeartbeat.Format(time.RFC3339))
				}
			}
		}
	}()
}
