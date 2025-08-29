package exitpeer

import (
	"Client_peer/pb"
	"Client_peer/utils"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"sync"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	ifaceName  = "wg-exit"
	listenPort = 51820
	exitPeerIP = "10.100.0.1/24"
	clientCIDR = "10.100.0.%d/32"
)

type ExitPeerServer struct {
	pb.UnimplementedExitPeerServiceServer
	privKey   wgtypes.Key
	pubKey    wgtypes.Key
	ipAlloc   int
	ipAllocMu sync.Mutex
	allocMap  map[string]string
}

func NewExitPeerServer() *ExitPeerServer {
	priv, pub, err := utils.LoadOrCreateWGKeypair()
	if err != nil {
		log.Fatalf("❌ Failed to generate exit peer key: %v", err)
	}

	// DEBUG: Log the exit peer's keys
	log.Printf("🔑 Exit peer private key: %s", priv.String())
	log.Printf("🔑 Exit peer public key: %s", pub.String())

	if err := utils.EnsureInterface(ifaceName); err != nil {
		log.Fatalf("❌ Failed to ensure Exit WG interface: %v", err)
	}

	if err := utils.SetInterfaceAddress(ifaceName, exitPeerIP); err != nil {
		log.Fatalf("❌ Failed to set Exit peer IP: %v", err)
	}

	if err := utils.ConfigureWG(ifaceName, priv, listenPort, nil); err != nil {
		log.Fatalf("❌ Failed to configure Exit WG interface: %v", err)
	}

	if err := utils.EnableIPForwarding(); err != nil {
		log.Fatalf("❌ Failed to enable IP forwarding: %v", err)
	}

	// Auto-detect the correct outbound interface
	publicIface, err := utils.GetOutboundInterface()
	if err != nil {
		log.Fatalf("❌ Failed to detect outbound interface: %v", err)
	}
	log.Printf("🔍 Detected outbound interface: %s", publicIface)

	if err := utils.SetupMasquerade(publicIface); err != nil {
		log.Fatalf("❌ Failed to setup masquerade: %v", err)
	}

	if err := utils.SetupForwardRules(ifaceName, publicIface); err != nil {
		log.Fatalf("❌ Failed to setup forward rules: %v", err)
	}

	log.Printf("🚀 Exit Peer ready — PublicKey: %s | Listening on %d | LAN NAT via %s",
		pub.String(), listenPort, publicIface)

	return &ExitPeerServer{
		privKey:  priv,
		pubKey:   pub,
		ipAlloc:  2,
		allocMap: make(map[string]string),
	}
}

func (e *ExitPeerServer) allocateIPForPeer(peerID string) string {
	e.ipAllocMu.Lock()
	defer e.ipAllocMu.Unlock()

	// If already allocated, return existing
	if ip, ok := e.allocMap[peerID]; ok {
		return ip
	}

	ip := fmt.Sprintf(clientCIDR, e.ipAlloc)
	e.ipAlloc++
	e.allocMap[peerID] = ip

	return ip
}

func (e *ExitPeerServer) GetWireGuardInfo(ctx context.Context, req *pb.ExitPeerInfoRequest) (*pb.ExitPeerInfoResponse, error) {
	log.Printf("📡 Exit peer received request from %s", req.RequesterId)

	// Must receive client public key
	if req.ClientPublicKey == "" {
		return nil, fmt.Errorf("client public key missing in request")
	}

	// Proper base64 decoding
	clientKeyBytes, err := base64.StdEncoding.DecodeString(req.ClientPublicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid client public key base64: %v", err)
	}

	// Convert bytes to string properly
	clientKeyStr := string(clientKeyBytes)
	clientPubKey, err := wgtypes.ParseKey(clientKeyStr)
	if err != nil {
		log.Printf("❌ Failed to parse client public key: %v", err)
		log.Printf("❌ Received key bytes: %v", clientKeyBytes)
		log.Printf("❌ Received key string: %s", clientKeyStr)
		return nil, fmt.Errorf("invalid client public key: %v", err)
	}

	log.Printf("✅ Client public key parsed successfully: %s", clientPubKey.String())

	clientIP := e.allocateIPForPeer(req.RequesterId)

	// Allow all traffic (0.0.0.0/0) through the tunnel for VPN functionality
	_, allowAllNet, _ := net.ParseCIDR("0.0.0.0/0")

	// DEBUG: Log the peer configuration
	log.Printf("🔧 SERVER Adding Peer:")
	log.Printf("   Client Public Key: %s", clientPubKey.String())
	log.Printf("   Client IP: %s", clientIP)
	log.Printf("   Allowed IPs: %s (routing all traffic)", allowAllNet.String())

	peerCfg := wgtypes.PeerConfig{
		PublicKey:  clientPubKey,
		AllowedIPs: []net.IPNet{*allowAllNet},
	}

	if err := utils.ConfigureWG(ifaceName, e.privKey, listenPort, []wgtypes.PeerConfig{peerCfg}); err != nil {
		log.Printf("❌ Failed to configure WireGuard on server: %v", err)
		return nil, fmt.Errorf("failed to add peer to WG interface: %v", err)
	}

	// Verify server configuration
	if err := utils.DebugWGStatus(ifaceName); err != nil {
		log.Printf("❌ Server WireGuard status check failed: %v", err)
	}

	return &pb.ExitPeerInfoResponse{
		PublicKey:     e.pubKey.String(),
		EndpointIp:    utils.GetLocalIP(),
		EndpointPort:  fmt.Sprintf("%d", listenPort),
		AllowedIps:    "0.0.0.0/0",
		BandwidthMbps: 85.0,
		LatencyMs:     15.0,
		ClientIp:      clientIP,
	}, nil
}
