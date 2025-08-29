package client

import (
	"Client_peer/crypto"
	"Client_peer/pb"
	"Client_peer/utils"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"google.golang.org/grpc"
)

type ClientPeer struct {
	client          pb.SuperNodeServiceClient
	id              string
	region          string
	originalDNS     string
	originalGateway string
	ifaceName       string
	mu              sync.Mutex
}

func NewClientPeer(conn *grpc.ClientConn, id string, region string) *ClientPeer {
	return &ClientPeer{
		client: pb.NewSuperNodeServiceClient(conn),
		id:     id,
		region: region,
	}
}

func (cp *ClientPeer) Register() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	priv, pub, _ := crypto.LoadOrCreateKeypair()
	nonce := crypto.GenerateNonce()
	signature := crypto.SignPeerPayload(priv, cp.id, cp.region, "Linux", "symmetric", nonce)

	req := &pb.PeerRegistrationRequest{
		PeerId:    cp.id,
		PublicKey: base64.StdEncoding.EncodeToString(pub),
		Version:   "0.1",
		Os:        "Linux",
		Region:    cp.region,
		NatType:   "symmetric",
		Signature: signature,
		Nonce:     nonce,
		Ip:        utils.GetLocalIP(),
		GrpcPort:  "6000",
	}

	res, err := cp.client.RegisterClientPeer(ctx, req)
	if err != nil {
		return err
	}

	if !res.Success {
		return fmt.Errorf("registration failed: %s", res.Message)
	}

	log.Printf("✅ Registered to Super Node: %s | ID: %s", res.Message, res.AssignedId)
	return nil
}

func (cp *ClientPeer) StartHeartbeat() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

		req := &pb.PeerSessionHeartbeatRequest{
			PeerId:            cp.id,
			ExitPeerId:        "1535748",
			LatencyMs:         57,
			PacketLoss:        0.2,
			ThroughputMbps:    12.3,
			SessionUptimeSecs: 300,
		}

		res, err := cp.client.PeerSessionHeartbeat(ctx, req)
		cancel()
		if err != nil {
			log.Printf("Heartbeat failed: %v", err)
			continue
		}

		log.Printf("Heartbeat sent: %s", res.Message)
	}
}

func (cp *ClientPeer) RequestExitEndpoint(region string, minBW float32, maxLatency float32) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	wgPriv, wgPub, _ := utils.LoadOrCreateWGKeypair()
	pubB64 := utils.PublicKeyBase64(wgPub)

	// DEBUG: Log the keys to see what's being sent
	log.Printf("🔑 Local private key: %s", wgPriv.String())
	log.Printf("🔑 Local public key: %s", wgPub.String())
	log.Printf("🔑 Base64 public key being sent: %s", pubB64)

	req := &pb.ExitRequest{
		PeerId:           cp.id,
		ClientPublicKey:  pubB64,
		RequestedRegion:  region,
		MinBandwidthMbps: minBW,
		MaxLatencyMs:     maxLatency,
	}

	wgCfg, err := cp.client.RequestExit(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to request exit: %w", err)
	}

	log.Printf("✅ Received WG config from SuperNode. Setting up interface...")

	log.Println("🎯 Received WireGuard Config:")
	log.Printf("Interface Private Key: %s", wgCfg.InterfacePrivateKey)
	log.Printf("Interface Address:    %s", wgCfg.InterfaceAddress)
	log.Printf("DNS:                  %s", wgCfg.Dns)
	log.Printf("Peer Public Key:      %s", wgCfg.PeerPublicKey)
	log.Printf("Peer Endpoint:        %s", wgCfg.PeerEndpoint)
	log.Printf("Allowed IPs:          %s", wgCfg.AllowedIps)
	log.Printf("Keepalive:            %d", wgCfg.Keepalive)

	originalDNS, originalGateway, err := utils.StoreOriginalSettings()
	if err != nil {
		log.Printf("Warning: Failed to store original settings: %v", err)
	}

	const ifaceName = "wg-exit"
	// Use the ALLOWED_IPS as the interface address, not a generated one
	interfaceAddress := wgCfg.InterfaceAddress
	if interfaceAddress == "" || interfaceAddress == "0.0.0.0/0" {
		interfaceAddress = "10.100.0.2/32" // Fallback for one-to-one setup
		log.Printf("⚠️ Invalid interface address from config (%s), using fallback: %s", wgCfg.InterfaceAddress, interfaceAddress)
	}
	// Validate CIDR format
	if _, _, err := net.ParseCIDR(interfaceAddress); err != nil {
		return fmt.Errorf("invalid interface address %s: %v", interfaceAddress, err)
	}

	// Remove any existing IPs on wg-exit
	_ = utils.RunCmd("ip", "addr", "flush", "dev", ifaceName)
	if err := utils.SetInterfaceAddress(ifaceName, interfaceAddress); err != nil {
		return fmt.Errorf("failed to assign IP %s: %v", interfaceAddress, err)
	}

	// Use local private key if not provided by super node
	ifacePrivKey := wgPriv
	if wgCfg.InterfacePrivateKey != "" {
		// If super node provided a private key, use it
		providedKey, err := wgtypes.ParseKey(wgCfg.InterfacePrivateKey)
		if err != nil {
			log.Printf("Warning: Invalid provided private key, using local: %v", err)
		} else {
			ifacePrivKey = providedKey
		}
	}

	// Parse peer public key
	peerPubKey, err := wgtypes.ParseKey(wgCfg.PeerPublicKey)
	if err != nil {
		return fmt.Errorf("invalid peer public key: %v", err)
	}

	// Parse endpoint
	host, portStr, err := net.SplitHostPort(wgCfg.PeerEndpoint)
	if err != nil {
		return fmt.Errorf("invalid peer endpoint format: %v", err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid peer endpoint port: %v", err)
	}

	// Parse allowed ips
	_, allowedNet, err := net.ParseCIDR(wgCfg.AllowedIps)
	if err != nil {
		return fmt.Errorf("invalid allowed IPs: %v", err)
	}

	// Set up Wireguard interface
	if err := utils.EnsureInterface(ifaceName); err != nil {
		return fmt.Errorf("failed to create interface: %v", err)
	}

	// Use the interface address
	if err := utils.SetInterfaceAddress(ifaceName, interfaceAddress); err != nil {
		return fmt.Errorf("failed to assign IP: %v", err)
	}

	keepalive := time.Duration(wgCfg.Keepalive) * time.Second

	peer := wgtypes.PeerConfig{
		PublicKey: peerPubKey,
		Endpoint: &net.UDPAddr{
			IP:   net.ParseIP(host),
			Port: port,
		},
		AllowedIPs:                  []net.IPNet{*allowedNet},
		PersistentKeepaliveInterval: &keepalive,
	}

	// TEST: Check if we can reach the exit peer first
	if err := utils.TestConnectivity(host, port); err != nil {
		log.Printf("⚠️  Connectivity warning: %v", err)
	} else {
		log.Printf("✅ Connectivity to exit peer %s:%d is good", host, port)
	}

	// DEBUG: Detailed WireGuard configuration logging
	log.Printf("🔧 CLIENT WireGuard Configuration:")
	log.Printf("   Interface Private Key: %s", ifacePrivKey.String())
	log.Printf("   Interface Address: %s", interfaceAddress)
	log.Printf("   Peer Public Key: %s", peerPubKey.String())
	log.Printf("   Peer Endpoint: %s:%d", host, port)
	log.Printf("   Allowed IPs: %s", allowedNet.String())
	log.Printf("   Keepalive: %v", keepalive)

	if err := utils.ConfigureWG(ifaceName, ifacePrivKey, 51820, []wgtypes.PeerConfig{peer}); err != nil {
		log.Printf("❌ Failed to configure WireGuard: %v", err)
		return fmt.Errorf("failed to configure WireGuard interface: %v", err)
	}

	// Verify the configuration was applied
	if err := utils.DebugWGStatus(ifaceName); err != nil {
		log.Printf("❌ WireGuard status check failed: %v", err)
	}

	// Configure DNS
	dnsServer := wgCfg.Dns
	if dnsServer == "" {
		dnsServer = "1.1.1.1" // Default fallback DNS
	}
	if err := utils.ConfigureDNS(ifaceName, dnsServer); err != nil {
		log.Printf("Warning: Failed to configure DNS: %v", err)
	}

	// Setup default route
	if err := utils.SetupDefaultRoute(ifaceName); err != nil {
		return fmt.Errorf("failed to setup default route: %v", err)
	}

	cp.originalDNS = originalDNS
	cp.originalGateway = originalGateway
	cp.ifaceName = ifaceName

	log.Println("🎉 WireGuard tunnel is up! You should now be able to route traffic through the Exit Peer.")
	return nil
}

func (cp *ClientPeer) Cleanup() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if cp.ifaceName != "" {
		utils.CleanupInterface(cp.ifaceName)
	}
	if cp.originalGateway != "" {
		utils.RestoreOriginalRoute(cp.originalGateway)
	}
	if cp.originalDNS != "" {
		utils.RestoreOriginalDNS(cp.originalDNS)
	}
}
