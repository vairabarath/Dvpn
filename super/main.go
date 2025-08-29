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
	// 🏁 CLI flags
	peerPort := flag.String("peer-port", "50052", "Port for Super Node Server")
	nodeID := flag.String("id", "", "Node ID (optional)")
	region := flag.String("region", "IN", "Region code for Super Node")
	baseIP := flag.String("base-ip", "127.0.0.1", "Base Node IP address")
	flag.Parse()

	finalID := *nodeID
	if finalID == "" {
		finalID = generateRandomID(*region)
	}

	// 🌐 Choose Base Node port based on region
	basePort := map[string]int{
		"IN": 50051,
		"US": 50053,
	}[*region]
	if basePort == 0 {
		basePort = 50051
	}
	baseAddr := fmt.Sprintf("%s:%d", *baseIP, basePort)

	// 🎯 Super Node's listen address
	localIP := utils.GetLocalIP()
	superNodeAddr := fmt.Sprintf("%s:%s", localIP, *peerPort)

	// 🌐 Dial base node
	conn, err := grpc.Dial(baseAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("❌ Failed to connect to base node: %v", err)
	}
	defer conn.Close()

	// Create reusable gRPC client for base
	baseClient := pb.NewBaseNodeServiceClient(conn)

	// 👂 Start gRPC server for client peers
	go func() {
		lis, err := net.Listen("tcp", superNodeAddr)
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		grpcServer := grpc.NewServer()

		// ⬇️ Pass baseClient into server handler
		superNodeServer := server.NewSupreNodeServer(baseClient, *region)
		superNodeServer.StartPeerMonitoring()

		pb.RegisterSuperNodeServiceServer(grpcServer, superNodeServer)

		log.Printf("🚀 Super Node Server is live on port %s", *peerPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// 🔐 Register this Super Node to base
	node := client.NewSupreNode(conn, finalID, *peerPort, *region)
	if err := node.Register(); err != nil {
		log.Fatalf("❌ Registration failed: %v", err)
	}

	log.Println("✅ Super Node registered to Base Node. Starting heartbeat...")
	go node.StartHeartbeat()

	select {}
}
