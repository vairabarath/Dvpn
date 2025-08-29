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
	region string
}

func NewSupreNode(conn *grpc.ClientConn, id string, port string, region string) *SuperNode {
	return &SuperNode{
		client: pb.NewBaseNodeServiceClient(conn),
		id:     id,
		port:   port,
		region: region,
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
	sign := super.SignPayload(priv, s.id, s.region, ip, nonce)

	req := &pb.RegisterRequest{
		NodeId:      s.id,
		Region:      s.region,
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

func (s *SuperNode) RequestExitCandidates(region string, minBandwidth float32, maxLatency float32, count int32) ([]*pb.SuperNode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	req := &pb.ExitRegionRequest{
		DesiredRegion:    region,
		MinBandwidthMbps: minBandwidth,
		MaxLatencyMs:     maxLatency,
		Count:            count,
	}

	res, err := s.client.RequestExitRegion(ctx, req)
	if err != nil {
		return nil, err
	}

	return res.Nodes, nil
}

// super to super
func RequestExitPeerFromRemote(ip, port string, req *pb.ExitPeerRequest) (*pb.ExitPeerResponse, error) {
	addr := fmt.Sprintf("%s:%s", ip, port)

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to dial remote super node: %w", err)
	}
	defer conn.Close()

	client := pb.NewSuperNodeServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	res, err := client.RequestExitPeer(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to request exit peer: %w", err)
	}

	return res, nil
}
