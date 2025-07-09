package main

import (
	"log"
	"net"

	pb "Base_node/pb"
	"Base_node/server"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	baseNodeServer := server.NewBaseNodeServer()
	baseNodeServer.StartSuperNodeMonitoring()

	grpcServer := grpc.NewServer()
	pb.RegisterBaseNodeServiceServer(grpcServer, baseNodeServer)

	log.Println("Base Node Server is live on port 50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
