package client

import (
	"Client_peer/pb"
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

type ClientPeer struct {
	client pb.SuperNodeServiceClient
	id string
}


func NewClientPeer(conn *grpc.ClientConn, id string) *ClientPeer {
	return &ClientPeer{
		client: pb.NewSuperNodeServiceClient(conn),
		id: id,
	}
}

func (cp *ClientPeer) Register() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()	

	req := &pb.PeerRegistrationRequest{
		PeerId: cp.id,
		PublicKey: "dfkakjhrwjbfkyqgf",
		Version: "0.1",
		Os: "Linux",
		Region: "IN",
		NatType: "symmetric",
	}

	res, err := cp.client.RegisterClientPeer(ctx, req)
	if err != nil {
		return err
	}

	log.Printf("✅ Registered to Super Node: %s | ID: %s", res.Message, res.AssignedId)
	return nil
}

func (cp *ClientPeer) StartHeartbeat() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		req := &pb.PeerSessionHeartbeatRequest{
			PeerId: cp.id,
			ExitPeerId: "1535748",
			LatencyMs: 57,
			PacketLoss: 0.2,
			ThroughputMbps: 12.3,
			SessionUptimeSecs: 300,
		}

		res, err := cp.client.PeerSessionHeartbeat(ctx, req)
		if err != nil {
			log.Printf("Heartbeat failed: %v", err)
			continue
		}

		log.Printf("Heartbeat sent: %s", res.Message)
	}
}