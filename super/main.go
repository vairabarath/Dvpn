package main

import (
	"Super_node/client"
	"Super_node/pb"
	"Super_node/server"
	"Super_node/utils"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

func generateRandomID(region string) string {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return region + "000"
	}
	return fmt.Sprintf("super-%s-%s", region, hex.EncodeToString(b))
}
func main() {
	peerPort := flag.String("peer-port", ":50052", "Port for Super Node Server")
	nodeID := flag.String("id", "", "Node ID (optional, will auto-generate if blank)")
	region := flag.String("region", "IN", "Region code for the Super Node")
	baseIP := flag.String("base-ip", "127.0.0.1", "IP address of the Base Node")

	flag.Parse() // ✅ Parse BEFORE you access flag values

	var basePort int
	switch *region {
	case "IN":
		basePort = 50051
	case "US":
		basePort = 50053 // ✅ use correct port for Base-US
	default:
		basePort = 50051 // fallback
	}

	baseAddr := fmt.Sprintf("%s:%d", *baseIP, basePort)
	ip := utils.GetLocalIP()
	addr := fmt.Sprintf("%s:%s", ip, *peerPort)

	finalID := *nodeID
	if finalID == "" {
		finalID = generateRandomID(*region)
	}
	// Step 1: Start gRPC server for Client Peers
	go func() {
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		grpcServer := grpc.NewServer()
		superNodeServer := server.NewSupreNodeServer()
		superNodeServer.StartPeerMonitoring()

		pb.RegisterSuperNodeServiceServer(grpcServer, superNodeServer)

		log.Printf("🚀 Super Node Server is live on port %s", *peerPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Step 2: Connect to Base Node as a gRPC client
	conn, err := grpc.Dial(baseAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("❌ Failed to connect to base node: %v", err)
	}
	defer conn.Close()

	node := client.NewSupreNode(conn, finalID, *peerPort, *region)

	if err := node.Register(); err != nil {
		log.Fatalf("❌ Registration failed: %v", err)
	}

	exitNodes, err := node.RequestExitCandidates("US", 50.0, 200.0, 2)
	if err != nil {
		log.Fatalf("Failed to get exite Super Nodes: %v", err)
	}

	log.Printf("✅ Got %d exit candidates for US region:", len(exitNodes))
	for _, node := range exitNodes {
		log.Printf("🛰 %s (%s:%s)", node.NodeId, node.Ip, node.Port)
	}

	log.Println("✅ Super Node registered to Base Node. Starting heartbeat...")
	go node.StartHeartbeat() //  Also run this in a goroutine

	select {} // 🧠 Block forever to keep both sides alive
}
