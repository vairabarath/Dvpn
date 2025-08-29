# Decentralized VPN (Dvpn) System

A 2-region decentralized VPN system with WireGuard-based cross-region tunneling.

## 📋 **Quick Start**

### **1. For Deployment**
```bash
# Use the complete deployment package
cd deploy/
./scripts/preflight-check.sh  # Check system requirements
```

**👉 Follow the complete deployment guide:** [`deploy/final-process.md`](deploy/final-process.md)

**👉 Quick reference:** [`deploy/QUICK-REFERENCE.md`](deploy/QUICK-REFERENCE.md)

### **2. For Development**
```bash
# Build all components
cd base && go build -o ../bin/base .
cd ../super && go build -o ../bin/super .
cd ../clientPeer && go build -o ../bin/clientPeer .
```

## 🏗️ **Architecture**

### **2-Region System:**
- **IN Region**: Base Node + Super Node + Client/Exit Peer
- **US Region**: Base Node + Super Node + Client/Exit Peer

### **Cross-Region VPN Flow:**
1. Client → Local Super → Local Base
2. Local Base ↔ Remote Base (Federation)
3. Local Super ↔ Remote Super (Direct)
4. Remote Super → Remote Exit Peer
5. **Direct WireGuard Tunnel**: Client ↔ Remote Exit Peer

## 📦 **Project Structure**

```
├── base/           # Base Node service (regional coordination)
├── super/          # Super Node service (client management)
├── clientPeer/     # Client Peer service (VPN client + exit peer)
├── bin/            # Compiled binaries
├── deploy/         # 🎯 COMPLETE DEPLOYMENT PACKAGE
│   ├── bin/        # Production binaries
│   ├── scripts/    # Deployment and test scripts
│   ├── final-process.md      # Complete deployment guide
│   ├── QUICK-REFERENCE.md    # Quick reference guide
│   └── deployment-plan-2region.md
└── WIREGUARD_TEST_RESULTS.md # Test verification results
```

## ✅ **System Requirements**

- **Linux Ubuntu/Debian** (tested on Ubuntu)
- **WireGuard** (`sudo apt install wireguard wireguard-tools`)
- **Network Tools** (`net-tools`, `netcat-openbsd`, `curl`)
- **Go 1.23+** (for development only)

## 🚀 **Deployment**

**The `deploy/` folder contains everything needed for production deployment.**

1. **Copy `deploy/` to all 6 PCs**
2. **Run preflight checks** on each PC
3. **Follow the step-by-step process** in `deploy/final-process.md`

## 🔐 **Security Notes**

- Client peers require **sudo** for WireGuard interface management
- Base nodes may require **sudo** for network operations  
- WireGuard keys are auto-generated and stored securely
- NAT/firewall rules configured automatically

## 🧪 **Testing**

```bash
# System requirements check
cd deploy/
./scripts/preflight-check.sh

# WireGuard functionality test
./scripts/test-wireguard-direct.sh

# Cross-region connectivity test
./scripts/test-cross-region.sh
```

## 📖 **Documentation**

- **[Complete Deployment Process](deploy/final-process.md)** - Step-by-step guide
- **[Quick Reference](deploy/QUICK-REFERENCE.md)** - Essential commands and troubleshooting
- **[Test Results](WIREGUARD_TEST_RESULTS.md)** - WireGuard functionality verification
- **[Architecture Plan](deploy/deployment-plan-2region.md)** - System architecture overview

## 🎯 **Production Ready**

✅ **WireGuard functionality tested and verified**  
✅ **All components built and ready**  
✅ **Complete deployment package prepared**  
✅ **Documentation and guides complete**  
✅ **Cross-region VPN tunneling functional**
