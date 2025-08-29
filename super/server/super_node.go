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
	Ip             string
	GrpcPort       string
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

func NewSupreNodeServer(baseClient pb.BaseNodeServiceClient, region string) *SuperNodeServer {
	s := &SuperNodeServer{
		registeredPeers: make(map[string]*ClientPeerInfo),
		exitPeers:       make(map[string]*ExitPeerInfo),
		baseClient:      baseClient,
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
		Ip:            req.Ip,
		GrpcPort:      req.GrpcPort,
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

	peer, ok := s.registeredPeers[req.PeerId]
	if !ok {
		log.Printf("Heartbeat from unknown peer: %s", req.PeerId)
		return &pb.Ack{
			Received: false,
			Message:  "Peer not found",
		}, nil
	}

	peer.LatencyMs = req.LatencyMs
	peer.PacketLoss = req.PacketLoss
	peer.ThroughputMbps = req.ThroughputMbps
	peer.SessionUptime = req.SessionUptimeSecs
	peer.LastHeartbeat = time.Now()

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

// super to super for exit peer
func (s *SuperNodeServer) RequestExitPeer(ctx context.Context, req *pb.ExitPeerRequest) (*pb.ExitPeerResponse, error) {
	log.Printf("📞 Dynamically searching for exit peer in region: %s", req.RequestedRegion)

	var chosen *ClientPeerInfo
	for _, peer := range s.registeredPeers {
		if peer.Region == req.RequestedRegion &&
			peer.ThroughputMbps >= req.MinBandwidthMbps &&
			float32(peer.LatencyMs) <= req.MaxLatencyMs {

			log.Printf("✅ Candidate: %s | IP: %s:%s | Latency: %dms | BW: %.2f Mbps",
				peer.PeerID, peer.Ip, peer.GrpcPort, peer.LatencyMs, peer.ThroughputMbps)

			chosen = peer
			break // for now, pick the first match — later apply ranking logic
		}
	}

	if chosen == nil {
		log.Printf("❌ No suitable exit peer found in registered peers")
		return nil, fmt.Errorf("no suitable exit peer found in registered peers")
	}

	exitPeerAddr := fmt.Sprintf("%s:%s", chosen.Ip, chosen.GrpcPort)
	log.Printf("🔁 Connecting to exit peer %s at %s", chosen.PeerID, exitPeerAddr)

	conn, err := grpc.Dial(exitPeerAddr, grpc.WithInsecure())
	if err != nil {
		log.Printf("❌ Failed to connect to Exit Peer at %s: %v", exitPeerAddr, err)
		return nil, err
	}
	defer conn.Close()

	exitClient := pb.NewExitPeerServiceClient(conn)

	infoRes, err := exitClient.GetWireGuardInfo(ctx, &pb.ExitPeerInfoRequest{
		RequesterId:      req.RequesterId,
		ClientPublicKey:  req.ClientPublicKey,
		Region:           req.RequestedRegion,
		MinBandwidthMbps: req.MinBandwidthMbps,
		MaxLatencyMs:     req.MaxLatencyMs,
	})
	if err != nil {
		log.Printf("❌ Failed to fetch WireGuard info from Exit Peer %s: %v", chosen.PeerID, err)
		return nil, err
	}

	log.Printf("✅ WireGuard info received from exit peer %s: %s:%s",
		chosen.PeerID, infoRes.EndpointIp, infoRes.EndpointPort)

	return &pb.ExitPeerResponse{
		PublicKey:    infoRes.PublicKey,
		EndpointIp:   infoRes.EndpointIp,
		EndpointPort: infoRes.EndpointPort,
		AllowedIps:   infoRes.AllowedIps,
		PeerId:       chosen.PeerID,
		Region:       req.RequestedRegion,
		ClientIp:     infoRes.ClientIp,
	}, nil
}

// client and super
func (s *SuperNodeServer) RequestExit(ctx context.Context, req *pb.ExitRequest) (*pb.WireguardConfig, error) {
	log.Printf("📨 Exit request from Peer %s for region %s", req.PeerId, req.RequestedRegion)

	// HACK: find the client public key locally
	_, ok := s.registeredPeers[req.PeerId]
	if !ok {
		return nil, fmt.Errorf("unknown requesting peer %s", req.PeerId)
	}

	exitReq := &pb.ExitRegionRequest{
		DesiredRegion:    req.RequestedRegion,
		MinBandwidthMbps: req.MinBandwidthMbps,
		MaxLatencyMs:     req.MaxLatencyMs,
		Count:            1,
	}

	superList, err := s.baseClient.RequestExitRegion(ctx, exitReq) //requesting to the local basenode for remote supernodes
	if err != nil {
		log.Printf("❌ Failed to request remote SuperNodes: %v", err)
		return nil, err
	}
	if len(superList.Nodes) == 0 {
		log.Printf("No SuperNodes returened for region %s", req.RequestedRegion)
		return nil, fmt.Errorf("no SuperNodes available for region %s", req.RequestedRegion)
	}

	chosen := superList.Nodes[0] // TODO: implement load balancing logic for fairer distribution
	log.Printf("🛰 Chosen remote super: %s (%s:%s)", chosen.NodeId, chosen.Ip, chosen.Port)
	addr := fmt.Sprintf("%s:%s", chosen.Ip, chosen.Port)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Printf("❌ Failed to connect to remote SuperNode: %v", err)
		return nil, err
	}
	println("Connected to remote SuperNode")
	defer conn.Close()

	RemoteSuperNode := pb.NewSuperNodeServiceClient(conn)

	remoteReq := &pb.ExitPeerRequest{
		RequesterId:      req.PeerId,
		MinBandwidthMbps: req.MinBandwidthMbps,
		MaxLatencyMs:     req.MaxLatencyMs,
		RequestedRegion:  req.RequestedRegion,
		ClientPublicKey:  req.ClientPublicKey,
	}

	exitRes, err := RemoteSuperNode.RequestExitPeer(ctx, remoteReq)
	if err != nil {
		log.Printf("❌ Failed to request exit peer: %v", err)
		return nil, err
	}

	config := &pb.WireguardConfig{
		InterfacePrivateKey: "", // HACK: generate private key
		InterfaceAddress:    exitRes.ClientIp,
		Dns:                 "1.1.1.1",
		PeerPublicKey:       exitRes.PublicKey,
		PeerEndpoint:        fmt.Sprintf("%s:%s", exitRes.EndpointIp, exitRes.EndpointPort),
		AllowedIps:          "0.0.0.0/0",
		Keepalive:           25,
	}

	log.Printf("🎯 Prepared WireGuard config for peer %s to exit via %s", req.PeerId, exitRes.PeerId)
	return config, nil
}
