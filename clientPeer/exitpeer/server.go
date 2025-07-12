package exitpeer

import (
	"Client_peer/pb"
	"context"
	"log"
)

type ExitPeerServer struct {
	pb.UnimplementedExitPeerServiceServer
}

func NewExitPeerServer() *ExitPeerServer {
	return &ExitPeerServer{}
}

func (e *ExitPeerServer) GetWireGuardInfo(ctx context.Context, req *pb.ExitPeerInfoRequest) (*pb.ExitPeerInfoResponse, error) {
	log.Printf("📡 Exit peer received request from %s", req.RequesterId)

	// Simulate validation logic or real monitoring data
	return &pb.ExitPeerInfoResponse{
		PublicKey:     "real-exit-wg-pubkey",
		EndpointIp:    "192.168.1.100",
		EndpointPort:  "51820",
		AllowedIps:    "0.0.0.0/0",
		BandwidthMbps: 85.0,
		LatencyMs:     15.0,
	}, nil
}
