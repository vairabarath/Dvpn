package server

import (
	"context"
	"log"
	"time"

	pb "Base_node/pb"

	"google.golang.org/protobuf/types/known/emptypb"
)

type SuperNodeInfo struct {
	NodeID        string
	Region        string
	IP            string
	PublicKey     string
	Version       string
	MaxPeers      int32
	StartupTime   string
	RegisteredAt  string
	LastHeartbeat time.Time
	Port          string
}

type BaseNodeServer struct {
	pb.UnimplementedBaseNodeServiceServer
	registeredSuperNodes map[string]*SuperNodeInfo
}

func NewBaseNodeServer() *BaseNodeServer {
	return &BaseNodeServer{
		registeredSuperNodes: make(map[string]*SuperNodeInfo),
	}
}

func (s *BaseNodeServer) RegisterSuperNode(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	valid := VerifySuperNodeSignature(req.NodeId, req.Region, req.Ip, req.Nonce, req.PublicKey, req.Signature)

	if !valid {
		return &pb.RegisterResponse{
			Success: false,
			Message: "Invalid signature",
		}, nil
	}

	log.Println("Valid signature")

	s.registeredSuperNodes[req.NodeId] = &SuperNodeInfo{
		NodeID:        req.NodeId,
		Region:        req.Region,
		IP:            req.Ip,
		PublicKey:     req.PublicKey,
		Version:       req.Version,
		MaxPeers:      req.MaxPeers,
		StartupTime:   req.StartupTime,
		RegisteredAt:  time.Now().Format(time.RFC3339),
		LastHeartbeat: time.Now(),
		Port:          req.Port,
	}

	log.Printf("👤 Registered Super Node: %s [%s] IP: %s", req.NodeId, req.Region, req.Ip)

	return &pb.RegisterResponse{
		Success:      true,
		Message:      "Registration successful",
		AssignedId:   req.NodeId,
		RegisteredAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *BaseNodeServer) SuperNodeHeartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.Ack, error) {
	log.Printf("💓 Heartbeat from %s | Peers: %d | Latency: %.1fms | Bandwidth: %.2fMbps",
		req.NodeId, req.ActivePeers, req.AvgLatencyMs, req.BandwidthUsageMbps)

	node := s.registeredSuperNodes[req.NodeId]
	if node == nil {
		log.Printf("❌ Super Node %s not found", req.NodeId)
		return &pb.Ack{
			Received: false,
			Message:  "Super Node not found",
		}, nil
	}

	node.LastHeartbeat = time.Now()
	log.Printf("Heartbeat from %s | Last heartbeat: %s", req.NodeId, node.LastHeartbeat.Format(time.RFC3339))
	return &pb.Ack{
		Received: true,
		Message:  "Heartbeat received",
	}, nil
}

func (s *BaseNodeServer) StartSuperNodeMonitoring() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		for range ticker.C {
			now := time.Now()
			log.Println("🔎Searching for stale Super Nodes")
			for id, node := range s.registeredSuperNodes {
				age := now.Sub(node.LastHeartbeat)

				status := "Alive"
				if age > 2*time.Minute {
					status = "Stale"
				}

				log.Printf("🛰  %s | Region: %s | IP: %s | LastHeartbeat: %s | Status: %s",
					id, node.Region, node.IP, node.LastHeartbeat.Format(time.RFC3339), status)
			}
			log.Println("End of super Node monitoring cycle")
		}
	}()
}

func (s *BaseNodeServer) GetActiveSuperNodes(ctx context.Context, _ *emptypb.Empty) (*pb.SuperNodeList, error) {
	now := time.Now()
	list := &pb.SuperNodeList{}

	for _, node := range s.registeredSuperNodes {
		isAlive := now.Sub(node.LastHeartbeat) <= 2*time.Minute

		list.Nodes = append(list.Nodes, &pb.SuperNode{
			NodeId:          node.NodeID,
			Region:          node.Region,
			Ip:              node.IP,
			Version:         node.Version,
			LatestHeartbeat: node.LastHeartbeat.Format(time.RFC3339),
			IsAlive:         isAlive,
			Port:            node.Port,
		})
	}

	log.Printf("📡 Returned %d Super Nodes to client peer", len(list.Nodes))
	return list, nil
}
