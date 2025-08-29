# Final Deployment Process Guide

## 📋 **Pre-Deployment Checklist**

### **Essential Tools Installation (Run on ALL 6 PCs)**

#### **Step 1: Install WireGuard**
```bash
# Update package list
sudo apt update

# Install WireGuard
sudo apt install -y wireguard wireguard-tools

# Verify installation
wg --version
```

#### **Step 2: Install Network Tools**
```bash
# Install essential networking tools
sudo apt install -y net-tools netcat-openbsd curl

# Verify installations
netstat --version
nc -h
curl --version
```

#### **Step 3: Verify System Requirements**
```bash
# Check if WireGuard kernel module loads
sudo modprobe wireguard
lsmod | grep wireguard

# Check if we can create interfaces (test)
sudo ip link add dev test-check type wireguard
sudo ip link delete test-check

echo "✅ System requirements check complete"
```

---

## 🚀 **Deployment Process**

### **Phase 1: Copy Deployment Package**

1. **From development machine, copy to all 6 PCs:**
```bash
# Replace USER and PC_IP with actual values
scp -r ./deploy/ user@192.168.1.104:~/dvpn/  # IN Base
scp -r ./deploy/ user@192.168.1.103:~/dvpn/  # IN Super  
scp -r ./deploy/ user@192.168.1.101:~/dvpn/  # IN Client
scp -r ./deploy/ user@192.168.1.43:~/dvpn/   # US Base
scp -r ./deploy/ user@192.168.1.39:~/dvpn/   # US Super
scp -r ./deploy/ user@192.168.1.40:~/dvpn/   # US Client
```

2. **On each PC, verify deployment package:**
```bash
cd ~/dvpn
ls -la
# Should see: bin/, scripts/, deployment-plan-2region.md

# Make scripts executable
chmod +x scripts/**/*.sh
```

---

### **Phase 2: Start IN Region (India)**

**⚠️ IMPORTANT: Start components in this exact order with 30-second delays**

#### **Step 1: Start IN Base Node**
**On PC 192.168.1.104:**
```bash
cd ~/dvpn
echo "Starting IN Base Node..."
sudo ./scripts/IN/start-base.sh
```
**✅ Wait for message: "IN Base Node Server is listening on 192.168.1.104:50051"**

#### **Step 2: Start IN Super Node** 
**On PC 192.168.1.103 (wait 30 seconds after base node):**
```bash
cd ~/dvpn
echo "Starting IN Super Node..."
./scripts/IN/start-super.sh
```
**✅ Wait for message: "Super Node Server is live on port 50052"**
**✅ Wait for message: "Super Node registered to Base Node"**

#### **Step 3: Start IN Client/Exit Peer**
**On PC 192.168.1.101 (wait 30 seconds after super node):**
```bash
cd ~/dvpn
echo "Starting IN Client/Exit Peer..."
sudo ./scripts/IN/start-client.sh
```
**✅ Wait for message: "Exit Peer ready — PublicKey: ..."**
**✅ Wait for message: "Peer registered. Starting heartbeat..."**

---

### **Phase 3: Start US Region (United States)**

**⚠️ IMPORTANT: Start components in this exact order with 30-second delays**

#### **Step 1: Start US Base Node**
**On PC 192.168.1.43:**
```bash
cd ~/dvpn
echo "Starting US Base Node..."
sudo ./scripts/US/start-base.sh
```
**✅ Wait for message: "US Base Node Server is listening on 192.168.1.43:50053"**

#### **Step 2: Start US Super Node**
**On PC 192.168.1.39 (wait 30 seconds after base node):**
```bash
cd ~/dvpn
echo "Starting US Super Node..."
./scripts/US/start-super.sh
```
**✅ Wait for message: "Super Node Server is live on port 50054"**
**✅ Wait for message: "Super Node registered to Base Node"**

#### **Step 3: Start US Client/Exit Peer**
**On PC 192.168.1.40 (wait 30 seconds after super node):**
```bash
cd ~/dvpn
echo "Starting US Client/Exit Peer..."
sudo ./scripts/US/start-client.sh
```
**✅ Wait for message: "Exit Peer ready — PublicKey: ..."**
**✅ Wait for message: "Peer registered. Starting heartbeat..."**

---

### **Phase 4: System Verification**

#### **Step 1: Check All Services are Running**
**Run on any PC:**
```bash
cd ~/dvpn
./scripts/test-cross-region.sh
```

Expected output:
```
✅ IN Base Node is up
✅ US Base Node is up  
✅ IN Super Node is up
✅ US Super Node is up
✅ US Exit Peer is up
```

#### **Step 2: Verify Network Connectivity**
**Check ports are listening (run on respective PCs):**
```bash
# Check IN region ports
netstat -tulpn | grep :50051  # Base (104)
netstat -tulpn | grep :50052  # Super (103)  
netstat -tulpn | grep :6000   # Client (101)

# Check US region ports
netstat -tulpn | grep :50053  # Base (43)
netstat -tulpn | grep :50054  # Super (39)
netstat -tulpn | grep :6001   # Client (40)
```

#### **Step 3: Verify WireGuard Readiness**
**On client PCs (101 and 40):**
```bash
# Check WireGuard interface exists
sudo ip addr show | grep wg-exit

# Check WireGuard configuration
sudo wg show
```

---

### **Phase 5: Test Cross-Region VPN**

#### **Step 1: Request Cross-Region Exit**
**On IN Client (PC 192.168.1.101):**
```bash
cd ~/dvpn
echo "Requesting exit through US region..."
timeout 30 ./bin/clientPeer -base-ip=192.168.1.104 -region=IN -exit-port=6000 -req-region=US
```

#### **Step 2: Monitor WireGuard Tunnel Creation**
**On US Exit Peer (PC 192.168.1.40):**
```bash
# Watch for new peer connections
watch -n 2 'sudo wg show'
```

**Expected to see:**
```
interface: wg-exit
  public key: <US_EXIT_PUBLIC_KEY>
  private key: (hidden)
  listening port: 51820

peer: <IN_CLIENT_PUBLIC_KEY>
  allowed ips: 10.100.0.2/32
  latest handshake: X seconds ago
  transfer: X.XX KiB received, X.XX KiB sent
```

#### **Step 3: Test Traffic Through Tunnel**
**On IN Client (PC 192.168.1.101) - if tunnel established:**
```bash
# Test connectivity through WG interface
ping -I wg-exit 8.8.8.8

# Test web traffic (if internet available)  
curl --interface wg-exit http://httpbin.org/ip
```

---

## 🔍 **Troubleshooting Commands**

### **Check Component Status**
```bash
# Check if components are running
ps aux | grep -E '(base|super|clientPeer)'

# Check component logs (if running in background)
journalctl -f | grep -E '(base|super|clientPeer)'
```

### **Network Diagnostics**
```bash
# Check network connectivity between regions
nc -zv 192.168.1.104 50051  # Test IN Base
nc -zv 192.168.1.43 50053   # Test US Base

# Check local interfaces
ip addr show

# Check routing tables
ip route show
```

### **WireGuard Diagnostics**
```bash
# Check WireGuard interfaces
sudo wg show all

# Check WireGuard interface details
sudo ip addr show wg-exit

# Check if WireGuard module loaded
lsmod | grep wireguard
```

### **Firewall Check**
```bash
# Check if ports are blocked
sudo ufw status

# If UFW is active, allow required ports:
sudo ufw allow 50051
sudo ufw allow 50052  
sudo ufw allow 50053
sudo ufw allow 50054
sudo ufw allow 6000
sudo ufw allow 6001
sudo ufw allow 51820
```

---

## 🚨 **Common Issues & Solutions**

### **Issue: "Connection refused" errors**
**Solution:**
1. Check if base node is running first
2. Verify correct IP addresses in commands
3. Check firewall settings
4. Ensure 30-second delays between component starts

### **Issue: WireGuard interface not created**
**Solution:**
```bash
# Check WireGuard installation
sudo apt install wireguard wireguard-tools

# Load WireGuard kernel module
sudo modprobe wireguard

# Check permissions
sudo chmod +x ./scripts/**/*.sh
```

### **Issue: "Permission denied" for network operations**
**Solution:**
- Client peers MUST run with sudo
- Base nodes may need sudo for some operations
- Check sudo permissions for network commands

### **Issue: Components can't find each other**
**Solution:**
1. Verify IP addresses in deployment plan
2. Check network connectivity: `ping 192.168.1.104`
3. Ensure all PCs are on same subnet
4. Check DNS resolution if using hostnames

---

## 📊 **Success Indicators**

### **✅ IN Region Success:**
- Base Node: "IN Base Node Server is listening on 192.168.1.104:50051"  
- Super Node: "Super Node registered to Base Node"
- Client: "Exit Peer ready — PublicKey: ..."

### **✅ US Region Success:**
- Base Node: "US Base Node Server is listening on 192.168.1.43:50053"
- Super Node: "Super Node registered to Base Node"  
- Client: "Exit Peer ready — PublicKey: ..."

### **✅ Cross-Region VPN Success:**
- WireGuard interface `wg-exit` created
- Peer handshake established
- Traffic flows through tunnel
- IP address changes when using VPN

---

## 🎯 **Final Verification Checklist**

- [ ] All 6 PCs have WireGuard installed
- [ ] All 6 PCs have deploy package copied
- [ ] IN region fully started and registered  
- [ ] US region fully started and registered
- [ ] Cross-region connectivity test passed
- [ ] WireGuard tunnels can be established
- [ ] Traffic flows through VPN tunnels

**🎉 Congratulations! Your decentralized VPN system is now operational!**
