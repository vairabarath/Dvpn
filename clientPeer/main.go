package main

import (
	"Client_peer/client"
	"Client_peer/exitpeer"
	basepb "Client_peer/pb"
	"Client_peer/utils"
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func generateRandomID(region string) string {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return region + "000"
	}
	return fmt.Sprintf("peer-%s-%s", region, hex.EncodeToString(b))
}

func main() {
	baseIP := flag.String("base-ip", "127.0.0.1", "IP address of the Base Node")
	region := flag.String("region", "IN", "Region code for Super Node")
	exitPeerPort := flag.String("exit-port", "6000", "Port to run Exit Peer gRPC Server")
	reqRegion := flag.String("req-region", "", "Region code to request exit (optional)")
	flag.Parse()

	ip := utils.GetLocalIP()
	addr := fmt.Sprintf("%s:%s", ip, *exitPeerPort)
	id := generateRandomID(*region)

	// Setup signal handling for clean shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup
	var peer *client.ClientPeer
	var cleanupOnce sync.Once

	// Cleanup function
	cleanup := func() {
		cleanupOnce.Do(func() {
			log.Println("🛑 Shutdown signal received. Cleaning up...")
			if peer != nil {
				peer.Cleanup()
				log.Println("✅ Client peer cleanup completed")
			}
			log.Println("✅ Cleanup complete. Exiting.")
		})
	}

	// Ensure cleanup runs on normal exit too (panic, fatal errors, etc.)
	defer cleanup()

	// Start goroutine to handle shutdown signals
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-sigChan
		cleanup()
	}()

	// 🛰 Start Exit Peer gRPC server in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("❌ Failed to listen on exit peer port %s: %v", *exitPeerPort, err)
		}
		grpcServer := grpc.NewServer()
		basepb.RegisterExitPeerServiceServer(grpcServer, exitpeer.NewExitPeerServer())
		log.Printf("🚪 Exit Peer gRPC server running on port %s", *exitPeerPort)

		// Start server in a goroutine so we can stop it on signal
		serverErr := make(chan error, 1)
		go func() {
			serverErr <- grpcServer.Serve(lis)
		}()

		// Wait for shutdown signal or server error
		select {
		case <-sigChan:
			log.Println("🛑 Gracefully stopping Exit Peer server...")
			grpcServer.GracefulStop()
		case err := <-serverErr:
			if err != nil {
				log.Printf("❌ Exit Peer server failed: %v", err)
			}
		}
	}()

	basePort := map[string]int{
		"IN": 50051,
		"US": 50053,
	}[*region]
	if basePort == 0 {
		basePort = 50051
	}
	baseAddr := fmt.Sprintf("%s:%d", *baseIP, basePort)

	// 🌐 Connect to Base Node
	baseConn, err := grpc.Dial(baseAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("❌ Failed to connect to base node: %v", err)
	}
	defer baseConn.Close()

	baseClient := basepb.NewBaseNodeServiceClient(baseConn)

	// 🔍 Get Super Nodes
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

	log.Printf("🎉 Connecting to Super Node: %s at %s", chosen.NodeId, chosen.Ip)

	saddr := fmt.Sprintf("%s:%s", chosen.Ip, chosen.Port)
	superConn, err := grpc.Dial(saddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("❌ Failed to connect to super node: %v", err)
	}
	defer superConn.Close()

	peer = client.NewClientPeer(superConn, id, *region)

	if err := peer.Register(); err != nil {
		log.Fatalf("❌ Failed to register peer: %v", err)
	}

	log.Println("✅ Peer registered. Starting heartbeat...")
	go peer.StartHeartbeat()

	if *reqRegion != "" {
		log.Printf("📨 Requesting exit to region %s...", *reqRegion)
		if err := peer.RequestExitEndpoint(*reqRegion, 10.0, 100.0); err != nil {
			log.Fatalf("❌ Failed to request exit: %v", err)
		}
	} else {
		log.Println("ℹ️ No --req-region specified, skipping exit peer request.")
	}

	// Wait for shutdown signal
	wg.Wait()
}
