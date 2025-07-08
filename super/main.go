package main

import (
	"Super_node/client"
	"log"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())	
	if err != nil {
		log.Fatalf("Failed to connect to base node: %v", err)
	}
	defer conn.Close()

	node := client.NewSupreNode(conn, "super-IN-001")

	if err := node.Register(); err != nil {
		log.Fatalf("Registration failed: %v", err)
	}

	log.Println("Super Node registered. Starting heartbeat.....")
	node.StartHeartbeat()

	select {}
}
