package client

import (
	"Super_node/pb"
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)


type SuperNode struct {
	client pb.BaseNodeServiceClient
	id string
}

func NewSupreNode(conn *grpc.ClientConn, id string) *SuperNode {
	return &SuperNode{
		client: pb.NewBaseNodeServiceClient(conn),
		id: id,
	}
}

func (s *SuperNode) Register() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	req := &pb.RegisterRequest{
		NodeId: s.id,
		Region: "IN",
		Ip: "127.0.0.1",
		PublicKey: "dsfgakjhdcbaykgfakhjsg",
		MaxPeers: 100,
		Version: "0.1",
		StartupTime: time.Now().Format(time.RFC3339),
	}

	res, err := s.client.RegisterSuperNode(ctx, req)
	if err != nil{
		return err
	}

	log.Printf("Registered with base node: %s", res.Message)
	return nil
}

func (s *SuperNode) StartHeartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		req := &pb.HeartbeatRequest{
			NodeId: s.id,
			ActivePeers: 80,
			AvgLatencyMs: 57,
			ExitPeersAvailable: 10,
			BandwidthUsageMbps: 72.6,
			Timestamp: time.Now().Format(time.RFC3339),
		}

		res, err := s.client.SuperNodeHeartbeat(ctx, req)
		if err != nil {
			log.Printf("Heartbeat failed: %v", err)
			continue
		}

		log.Printf("Heartbeat sent: %s", res.Message)
	}
}