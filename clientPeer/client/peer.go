package client

import (
	"Client_peer/crypto"
	"Client_peer/pb"
	"Client_peer/utils"
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
	region string
}

func NewClientPeer(conn *grpc.ClientConn, id string, region string) *ClientPeer {
	return &ClientPeer{
		client: pb.NewSuperNodeServiceClient(conn),
		id:     id,
		region: region,
	}
}

func (cp *ClientPeer) Register() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	priv, pub, _ := crypto.LoadOrCreateKeypair()
	nonce := crypto.GenerateNonce()
	signature := crypto.SignPeerPayload(priv, cp.id, cp.region, "Linux", "symmetric", nonce)

	req := &pb.PeerRegistrationRequest{
		PeerId:    cp.id,
		PublicKey: base64.StdEncoding.EncodeToString(pub),
		Version:   "0.1",
		Os:        "Linux",
		Region:    cp.region,
		NatType:   "symmetric",
		Signature: signature,
		Nonce:     nonce,
		Ip:        utils.GetLocalIP(),
		GrpcPort:  "6000",
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

func (cp *ClientPeer) RequestExitEndpoint(region string, minBW float32, maxLatency float32) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	req := &pb.ExitRequest{
		PeerId:           cp.id,
		RequestedRegion:  region,
		MinBandwidthMbps: minBW,
		MaxLatencyMs:     maxLatency,
	}

	res, err := cp.client.RequestExit(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to request exit: %w", err)
	}

	log.Println("🎯 Received WireGuard Config:")
	log.Printf("Interface Private Key: %s", res.InterfacePrivateKey)
	log.Printf("Interface Address:    %s", res.InterfaceAddress)
	log.Printf("DNS:                  %s", res.Dns)
	log.Printf("Peer Public Key:      %s", res.PeerPublicKey)
	log.Printf("Peer Endpoint:        %s", res.PeerEndpoint)
	log.Printf("Allowed IPs:          %s", res.AllowedIps)
	log.Printf("Keepalive:            %d", res.Keepalive)

	return nil
}
