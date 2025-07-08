package main

import (
	"Super_node/client"
	"Super_node/pb"
	"Super_node/server"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	// Step 1: Start gRPC server for Client Peers
	go func() {
		lis, err := net.Listen("tcp", ":50052")
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		grpcServer := grpc.NewServer()
		superNodeServer := server.NewSupreNodeServer()

		pb.RegisterSuperNodeServiceServer(grpcServer, superNodeServer)

		log.Println("🚀 Super Node Server is live on port 50052")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Step 2: Connect to Base Node as a gRPC client
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("❌ Failed to connect to base node: %v", err)
	}
	defer conn.Close()

	node := client.NewSupreNode(conn, "super-IN-001")

	if err := node.Register(); err != nil {
		log.Fatalf("❌ Registration failed: %v", err)
	}

	log.Println("✅ Super Node registered to Base Node. Starting heartbeat...")
	go node.StartHeartbeat() // ✅ Also run this in a goroutine

	select {} // 🧠 Block forever to keep both sides alive
}
