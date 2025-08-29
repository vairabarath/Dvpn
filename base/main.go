package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	pb "Base_node/pb"
	"Base_node/server"
	"Base_node/utils"

	"google.golang.org/grpc"
)

func main() {
	region := flag.String("region", "", "Region of the base node")
	port := flag.String("port", "50051", "Port for Base Node Server")
	// peerBaseIP := flag.String("base-ip", "", "Optional IP:port of another base node to ping")
	flag.Parse()

	ip := utils.GetLocalIP()
	addr := fmt.Sprintf("%s:%s", ip, *port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	baseNodeServer := server.NewBaseNodeServer(*region)
	baseNodeServer.StartSuperNodeMonitoring()
	federationServer := server.NewFederationServer(*region, baseNodeServer)

	grpcServer := grpc.NewServer()
	pb.RegisterBaseNodeServiceServer(grpcServer, baseNodeServer)
	pb.RegisterBaseFederationServiceServer(grpcServer, federationServer)

	log.Printf("%s Base Node Server is listening on %s", *region, addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
