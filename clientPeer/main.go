package main

import (
	"Client_peer/client"
	"log"

	"google.golang.org/grpc"
)




func main() {
	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to Super Node: %v", err)
	}
	defer conn.Close()
	
	peer := client.NewClientPeer(conn, "peer-In-7788")

	if err := peer.Register(); err != nil {
		log.Fatalf("Registration failed: %v", err)
	}

	log.Println("Peer registered. Starting heartbeat.....")
	peer.StartHeartbeat()
}