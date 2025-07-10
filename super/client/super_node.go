package client

import (
	super "Super_node/crypto"
	"Super_node/pb"
	"Super_node/utils"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
)

type SuperNode struct {
	client pb.BaseNodeServiceClient
	id     string
	port   string
}

func NewSupreNode(conn *grpc.ClientConn, id string, port string) *SuperNode {
	return &SuperNode{
		client: pb.NewBaseNodeServiceClient(conn),
		id:     id,
		port:   port,
	}
}

func (s *SuperNode) Register() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	priv, pub, err := super.LoadOrCreateKeypair()
	if err != nil {
		log.Fatalf("❌ Failed to load/create keypair: %v", err)
	}

	nonce := super.GenerateNonce()
	ip := utils.GetLocalIP()
	fmt.Println(nonce)
	sign := super.SignPayload(priv, s.id, "IN", ip, nonce)

	req := &pb.RegisterRequest{
		NodeId:      s.id,
		Region:      "IN",
		Ip:          ip,
		Port:        s.port,
		PublicKey:   base64.StdEncoding.EncodeToString(pub),
		Signature:   sign,
		Nonce:       nonce,
		MaxPeers:    100,
		Version:     "0.1",
		StartupTime: time.Now().Format(time.RFC3339),
	}

	res, err := s.client.RegisterSuperNode(ctx, req)
	if err != nil {
		return err
	}

	if !res.Success {
		log.Printf("❌ Registration rejected by base node: %s", res.Message)
		return fmt.Errorf("registration failed: %s", res.Message)
	}

	log.Printf("✅ Registered with base node: %s", res.Message)
	return nil
}

func (s *SuperNode) StartHeartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

		req := &pb.HeartbeatRequest{
			NodeId:             s.id,
			ActivePeers:        80,
			AvgLatencyMs:       57,
			ExitPeersAvailable: 10,
			BandwidthUsageMbps: 72.6,
			Timestamp:          time.Now().Format(time.RFC3339),
		}

		res, err := s.client.SuperNodeHeartbeat(ctx, req)
		cancel()
		if err != nil {
			log.Printf("The base node went down. Heartbeat failed: %v", err)
			continue
		}

		log.Printf("Heartbeat sent: %s", res.Message)
	}
}
