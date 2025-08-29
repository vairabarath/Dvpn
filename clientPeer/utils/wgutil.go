package utils

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func RunCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v: %v - %s", name, args, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func EnsureInterface(iface string) error {

	_ = RunCmd("ip", "link", "show", iface)

	if err := RunCmd("ip", "link", "add", "dev", iface, "type", "wireguard"); err != nil {
		if !strings.Contains(err.Error(), "File exists") {
			return fmt.Errorf("failed to create WireGuard interface %s: %v", iface, err)
		}
	}

	time.Sleep(100 * time.Millisecond)

	if err := RunCmd("ip", "link", "set", "up", "dev", iface); err != nil {
		return err
	}
	return nil
}

func SetInterfaceAddress(iface, cidr string) error {
	// Remove any existing address first
	_ = RunCmd("ip", "addr", "del", cidr, "dev", iface)
	// Add the new address
	return RunCmd("ip", "addr", "add", cidr, "dev", iface)
}

func ConfigureWG(iface string, privateKey wgtypes.Key, listenPort int, peers []wgtypes.PeerConfig) error {
	client, err := wgctrl.New()
	if err != nil {
		return err
	}
	defer client.Close()

	cfg := wgtypes.Config{
		PrivateKey: &privateKey,
		ListenPort: &listenPort,
		Peers:      peers,
	}

	return client.ConfigureDevice(iface, cfg)
}

func GenerateKeypair() (wgtypes.Key, wgtypes.Key, error) {
	priv, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return wgtypes.Key{}, wgtypes.Key{}, err
	}
	pub := priv.PublicKey()
	return priv, pub, nil
}

func EnableIPForwarding() error {
	if err := RunCmd("sysctl", "-w", "net.ipv4.ip_forward=1"); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %v", err)
	}
	return nil
}

func SetupMasquerade(outIf string) error {
	_ = RunCmd("iptables", "-t", "nat", "-C", "POSTROUTING", "-o", outIf, "-j", "MASQUERADE")
	if err := RunCmd("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", outIf, "-j", "MASQUERADE"); err != nil {
		return err
	}

	return nil
}

func waitForInterface(iface string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := RunCmd("ip", "link", "show", iface); err == nil {
			return nil
		}

		time.Sleep(200 * time.Millisecond)
	}

	return fmt.Errorf("interface %s not found after timeout", iface)
}

func StoreOriginalSettings() (string, string, error) {
	originalDNS := getOriginalDNS()
	originalGateway := getDefaultGateway()
	return originalDNS, originalGateway, nil
}

func getOriginalDNS() string {
	out, err := exec.Command("grep", "^nameserver", "/etc/resolv.conf").Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) > 0 {
		parts := strings.Fields(lines[0])
		if len(parts) >= 2 {
			return parts[1]
		}
	}
	return ""
}

func getDefaultGateway() string {
	out, err := exec.Command("ip", "route", "show", "default").Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 3 && parts[0] == "default" && parts[1] == "via" {
			return parts[2]
		}
	}
	return ""
}

// ConfigureDNS sets DNS for the interface
func ConfigureDNS(ifaceName, dnsServer string) error {
	// Try resolvectl first
	if err := RunCmd("resolvectl", "dns", ifaceName, dnsServer); err != nil {
		// Fallback to resolv.conf
		return RunCmd("sh", "-c", fmt.Sprintf("echo 'nameserver %s' > /etc/resolv.conf", dnsServer))
	}
	return nil
}

// SetupDefaultRoute sets default route through the WireGuard interface
func SetupDefaultRoute(ifaceName string) error {
	// Remove existing default routes
	_ = RunCmd("ip", "route", "del", "default")
	if err := RunCmd("ip", "route", "add", "default", "dev", ifaceName); err != nil {
		return fmt.Errorf("failed to setup default route via %s: %v", ifaceName, err)
	}
	return nil
}

// RestoreOriginalRoute restores original default route
func RestoreOriginalRoute(originalGateway string) error {
	if originalGateway != "" {
		return RunCmd("ip", "route", "add", "default", "via", originalGateway)
	}
	return nil
}

// RestoreOriginalDNS restores original DNS settings
func RestoreOriginalDNS(originalDNS string) error {
	if originalDNS != "" {
		return RunCmd("sh", "-c", fmt.Sprintf("echo 'nameserver %s' > /etc/resolv.conf", originalDNS))
	}
	return nil
}

// CleanupInterface removes the WireGuard interface
func CleanupInterface(ifaceName string) error {
	return RunCmd("ip", "link", "del", "dev", ifaceName)
}

// Add these functions to your existing utils package

func TestConnectivity(host string, port int) error {
	timeout := time.Second * 3
	conn, err := net.DialTimeout("udp", fmt.Sprintf("%s:%d", host, port), timeout)
	if err != nil {
		return fmt.Errorf("UDP connectivity to %s:%d failed: %v", host, port, err)
	}
	defer conn.Close()

	log.Printf("✅ UDP connectivity to %s:%d successful", host, port)
	return nil
}

func DebugWGStatus(iface string) error {
	client, err := wgctrl.New()
	if err != nil {
		return err
	}
	defer client.Close()

	device, err := client.Device(iface)
	if err != nil {
		return fmt.Errorf("failed to get device %s: %v", iface, err)
	}

	log.Printf("📊 WireGuard Status for %s:", iface)
	log.Printf("   Public Key: %s", device.PublicKey.String())
	log.Printf("   Listen Port: %d", device.ListenPort)
	log.Printf("   Number of Peers: %d", len(device.Peers))

	for i, peer := range device.Peers {
		log.Printf("   Peer %d:", i+1)
		log.Printf("     Public Key: %s", peer.PublicKey.String())
		log.Printf("     Allowed IPs: %v", peer.AllowedIPs)
		if peer.Endpoint != nil {
			log.Printf("     Endpoint: %s", peer.Endpoint.String())
		}
		log.Printf("     Last Handshake: %v", peer.LastHandshakeTime)
		log.Printf("     Transfer: RX: %d, TX: %d", peer.ReceiveBytes, peer.TransmitBytes)
	}

	return nil
}

// SetupForwardRules adds iptables FORWARD rules for WireGuard traffic
func SetupForwardRules(wgInterface, outInterface string) error {
	// Allow traffic from WireGuard interface to outbound interface
	_ = RunCmd("iptables", "-C", "FORWARD", "-i", wgInterface, "-o", outInterface, "-j", "ACCEPT")
	if err := RunCmd("iptables", "-A", "FORWARD", "-i", wgInterface, "-o", outInterface, "-j", "ACCEPT"); err != nil {
		return fmt.Errorf("failed to add forward rule %s->%s: %v", wgInterface, outInterface, err)
	}

	// Allow established and related connections back
	_ = RunCmd("iptables", "-C", "FORWARD", "-i", outInterface, "-o", wgInterface, "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT")
	if err := RunCmd("iptables", "-A", "FORWARD", "-i", outInterface, "-o", wgInterface, "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT"); err != nil {
		return fmt.Errorf("failed to add return forward rule %s->%s: %v", outInterface, wgInterface, err)
	}

	// Allow ICMP echo-request (for outgoing pings)
	_ = RunCmd("iptables", "-C", "FORWARD", "-p", "icmp", "--icmp-type", "echo-request", "-j", "ACCEPT")
	if err := RunCmd("iptables", "-A", "FORWARD", "-p", "icmp", "--icmp-type", "echo-request", "-j", "ACCEPT"); err != nil {
		return fmt.Errorf("failed to add ICMP echo-request rule: %v", err)
	}

	// Allow ICMP echo-reply (for ping responses)
	_ = RunCmd("iptables", "-C", "FORWARD", "-p", "icmp", "--icmp-type", "echo-reply", "-j", "ACCEPT")
	if err := RunCmd("iptables", "-A", "FORWARD", "-p", "icmp", "--icmp-type", "echo-reply", "-j", "ACCEPT"); err != nil {
		return fmt.Errorf("failed to add ICMP echo-reply rule: %v", err)
	}

	return nil
}

// GetOutboundInterface detects the interface used for outbound internet traffic
func GetOutboundInterface() (string, error) {
	// Use ip route to find the default gateway interface
	out, err := exec.Command("ip", "route", "get", "8.8.8.8").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get route to 8.8.8.8: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		for i, part := range parts {
			if part == "dev" && i+1 < len(parts) {
				return parts[i+1], nil
			}
		}
	}

	// Fallback: try to detect active non-loopback interface
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to list interfaces: %v", err)
	}

	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Get addresses for this interface
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		// Check if it has a valid IP address
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return iface.Name, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no suitable outbound interface found")
}
