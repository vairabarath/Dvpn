package main

import (
	"Client_peer/client"
	basepb "Client_peer/pb"
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	baseConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("❌ Failed to connect to base node: %v", err)
	}
	defer baseConn.Close()

	baseClient := basepb.NewBaseNodeServiceClient(baseConn)

	res, err := baseClient.GetActiveSuperNodes(context.Background(), &emptypb.Empty{})
	if err != nil {
		log.Fatalf("❌ Failed to get active super nodes: %v", err)
	}
	if len(res.Nodes) == 0 {
		log.Fatalf("❌ No active super nodes found")
	}

	var chosen *basepb.SuperNode
	for _, node := range res.Nodes {
		if node.IsAlive {
			chosen = node
			break
		}
	}
	if chosen == nil {
		log.Fatalf("❌ No alive super nodes found")
	}

	log.Printf("🎉 Connecting to super Node: %s at %s", chosen.NodeId, chosen.Ip)

	addr := fmt.Sprintf("%s:%s", chosen.Ip, chosen.Port)
	superConn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("❌ Failed to connect to super node: %v", err)
	}
	defer superConn.Close()

	peer := client.NewClientPeer(superConn, "peer-IN-7788")

	if err := peer.Register(); err != nil {
		log.Fatalf("❌ Failed to register peer: %v", err)
	}

	log.Println("✅ Peer registered. Starting heartbeat...")
	peer.StartHeartbeat()

	select {}
}
