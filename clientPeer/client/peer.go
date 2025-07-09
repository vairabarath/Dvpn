package client

import (
	"Client_peer/crypto"
	"Client_peer/pb"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
)

type ClientPeer struct {
	client pb.SuperNodeServiceClient
	id     string
}

func NewClientPeer(conn *grpc.ClientConn, id string) *ClientPeer {
	return &ClientPeer{
		client: pb.NewSuperNodeServiceClient(conn),
		id:     id,
	}
}

func (cp *ClientPeer) Register() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	priv, pub, _ := crypto.LoadOrCreateKeypair()
	nonce := crypto.GenerateNonce()
	signature := crypto.SignPeerPayload(priv, cp.id, "IN", "Linux", "symmetric", nonce)

	req := &pb.PeerRegistrationRequest{
		PeerId:    cp.id,
		PublicKey: base64.StdEncoding.EncodeToString(pub),
		Version:   "0.1",
		Os:        "Linux",
		Region:    "IN",
		NatType:   "symmetric",
		Signature: signature,
		Nonce:     nonce,
	}

	res, err := cp.client.RegisterClientPeer(ctx, req)
	if err != nil {
		return err
	}

	if !res.Success {
		return fmt.Errorf("registration failed: %s", res.Message)
	}

	log.Printf("✅ Registered to Super Node: %s | ID: %s", res.Message, res.AssignedId)
	return nil
}

func (cp *ClientPeer) StartHeartbeat() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

		req := &pb.PeerSessionHeartbeatRequest{
			PeerId:            cp.id,
			ExitPeerId:        "1535748",
			LatencyMs:         57,
			PacketLoss:        0.2,
			ThroughputMbps:    12.3,
			SessionUptimeSecs: 300,
		}

		res, err := cp.client.PeerSessionHeartbeat(ctx, req)
		cancel()
		if err != nil {
			log.Printf("Heartbeat failed: %v", err)
			continue
		}

		log.Printf("Heartbeat sent: %s", res.Message)
	}
}
