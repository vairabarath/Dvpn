package server

import (
	"Base_node/client"
	"context"
	"fmt"
	"log"
	"time"

	pb "Base_node/pb"

	"google.golang.org/protobuf/types/known/emptypb"
)

type SuperNodeInfo struct {
	NodeID        string
	Region        string
	IP            string
	Port          string
	PublicKey     string
	Version       string
	MaxPeers      int32
	StartupTime   string
	RegisteredAt  string
	LastHeartbeat time.Time
	BandwidthMbps float32
	AvgLatency    float32
	ExitPeers     int32
}

type BaseNodeServer struct {
	pb.UnimplementedBaseNodeServiceServer
	localRegion          string
	registeredSuperNodes map[string]*SuperNodeInfo
}

func NewBaseNodeServer(local string) *BaseNodeServer {
	return &BaseNodeServer{
		localRegion:          local,
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

	log.Printf("👤 Registered Super Node: %s [%s] IP: %s:%s", req.NodeId, req.Region, req.Ip, req.Port)

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
	node.BandwidthMbps = req.BandwidthUsageMbps
	node.AvgLatency = req.AvgLatencyMs
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

func (b *BaseNodeServer) GetFilteredSuperNodes(count int32, minBW float32, maxLatency float32) []*SuperNodeInfo {
	var filtered []*SuperNodeInfo
	for _, sn := range b.registeredSuperNodes {
		if sn.BandwidthMbps >= 0 && sn.AvgLatency <= 1000 {
			filtered = append(filtered, sn)
		}
	}

	if len(filtered) > int(count) {
		filtered = filtered[:count]
	}

	return filtered
}

func (s *BaseNodeServer) RequestExitRegion(ctx context.Context, req *pb.ExitRegionRequest) (*pb.SuperNodeList, error) {
	if req.DesiredRegion == s.localRegion {
		local := s.GetFilteredSuperNodes(req.Count, req.MinBandwidthMbps, req.MaxLatencyMs)
		var list pb.SuperNodeList
		for _, n := range local {
			list.Nodes = append(list.Nodes, &pb.SuperNode{
				NodeId:          n.NodeID,
				Region:          n.Region,
				Ip:              n.IP,
				Port:            n.Port,
				Version:         n.Version,
				LatestHeartbeat: n.LastHeartbeat.Format(time.RFC3339),
				IsAlive:         true,
			})
		}

		return &list, nil
	}

	targetAddr := lookupBaseAddress(req.DesiredRegion)
	if targetAddr == "" {
		return nil, fmt.Errorf("No base node found for region %s", req.DesiredRegion)
	}

	remoteNodes, err := client.FetchRemoteSupers(targetAddr, req.DesiredRegion, req.Count, req.MinBandwidthMbps, req.MaxLatencyMs)
	if err != nil {
		return nil, err
	}

	var list pb.SuperNodeList
	for _, sn := range remoteNodes {
		list.Nodes = append(list.Nodes, &pb.SuperNode{
			NodeId:  sn.NodeId,
			Region:  sn.Region,
			Ip:      sn.Ip,
			Port:    sn.Port,
			Version: "0.1",
			IsAlive: true,
		})
	}

	return &list, nil
}

func lookupBaseAddress(region string) string {
	// Office Network Configuration
	switch region {
	case "US":
		return "192.168.1.43:50053" // US Base Node (PC at .43)
	case "IN":
		return "192.168.1.104:50051" // India Base Node (PC at .104)
	default:
		return ""
	}
}
