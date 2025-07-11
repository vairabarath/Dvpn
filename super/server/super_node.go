package server

import (
	"Super_node/pb"
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
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

type ExitPeerInfo struct {
	PeerId        string
	PublicKey     string
	EndpointIp    string
	EndpointPort  string
	AllowedIps    string
	Region        string
	BandwidthMbps float32
	LatencyMs     float32
	LastSeen      time.Time
}

type SuperNodeServer struct {
	pb.UnimplementedSuperNodeServiceServer
	registeredPeers map[string]*ClientPeerInfo
	exitPeers       map[string]*ExitPeerInfo
	baseClient      pb.BaseNodeServiceClient
}

func NewSupreNodeServer(baseClient pb.BaseNodeServiceClient) *SuperNodeServer {
	s := &SuperNodeServer{
		registeredPeers: make(map[string]*ClientPeerInfo),
		exitPeers:       make(map[string]*ExitPeerInfo),
		baseClient:      baseClient,
	}

	s.exitPeers["exit-us-001"] = &ExitPeerInfo{
		PeerId:        "exit-us-001",
		PublicKey:     "wgpubkey123",
		EndpointIp:    "192.168.1.100",
		EndpointPort:  "51820",
		AllowedIps:    "0.0.0.0/0",
		Region:        "US",
		BandwidthMbps: 100.0,
		LatencyMs:     20.0,
		LastSeen:      time.Now(),
	}

	return s
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

// super to super
func (s *SuperNodeServer) RequestExitPeer(ctx context.Context, req *pb.ExitPeerRequest) (*pb.ExitPeerResponse, error) {
	log.Printf("🚪 Request from %s for exit peer in %s", req.RequesterId, req.RequestedRegion)

	for _, peer := range s.exitPeers {
		if peer.Region == req.RequestedRegion &&
			peer.BandwidthMbps >= req.MinBandwidthMbps &&
			peer.LatencyMs <= req.MaxLatencyMs {

			log.Printf("✅ Exit peer selected: %s", peer.PeerId)
			return &pb.ExitPeerResponse{
				PublicKey:    peer.PublicKey,
				EndpointIp:   peer.EndpointIp,
				EndpointPort: peer.EndpointPort,
				AllowedIps:   peer.AllowedIps,
				PeerId:       peer.PeerId,
				Region:       peer.Region,
			}, nil
		}
	}

	log.Println("❌ No matching exit peer found")
	return nil, fmt.Errorf("no suitable exit peer found")
}

// client and super
func (s *SuperNodeServer) RequestExit(ctx context.Context, req *pb.ExitRequest) (*pb.WireguardConfig, error) {
	log.Printf("📨 Exit request from Peer %s for region %s", req.PeerId, req.RequestedRegion)

	exitReq := &pb.ExitRegionRequest{
		DesiredRegion:    req.RequestedRegion,
		MinBandwidthMbps: req.MinBandwidthMbps,
		MaxLatencyMs:     req.MaxLatencyMs,
		Count:            1,
	}

	superList, err := s.baseClient.RequestExitRegion(ctx, exitReq)
	if err != nil {
		log.Printf("❌ Failed to request remote SuperNodes: %v", err)
		return nil, err
	}
	if len(superList.Nodes) == 0 {
		log.Printf("No SuperNodes returened for region %s", req.RequestedRegion)
		return nil, fmt.Errorf("no SuperNodes available for region %s", req.RequestedRegion)
	}

	chosen := superList.Nodes[0]
	log.Printf("🛰 Chosen remote super: %s (%s:%s)", chosen.NodeId, chosen.Ip, chosen.Port)

	addr := fmt.Sprintf("%s:%s", chosen.Ip, chosen.Port)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Printf("❌ Failed to connect to remote SuperNode: %v", err)
		return nil, err
	}
	defer conn.Close()

	client := pb.NewSuperNodeServiceClient(conn)

	exitRes, err := client.RequestExitPeer(ctx, &pb.ExitPeerRequest{
		RequesterId:      req.PeerId,
		MinBandwidthMbps: req.MinBandwidthMbps,
		MaxLatencyMs:     req.MaxLatencyMs,
		RequestedRegion:  req.RequestedRegion,
	})

	if err != nil {
		log.Printf("❌ Failed to request exit peer: %v", err)
		return nil, err
	}

	config := &pb.WireguardConfig{
		InterfacePrivateKey: "", // TODO
		InterfaceAddress:    "10.0.0.2/32",
		Dns:                 "1.1.1.1",
		PeerPublicKey:       exitRes.PublicKey,
		PeerEndpoint:        fmt.Sprintf("%s:%s", exitRes.EndpointIp, exitRes.EndpointPort),
		AllowedIps:          exitRes.AllowedIps,
		Keepalive:           25,
	}

	log.Printf("🎯 Prepared WireGuard config for peer %s to exit via %s", req.PeerId, exitRes.PeerId)

	return config, nil
}
