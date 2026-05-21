# WeVibe Chain Deployment Guide

This guide covers deploying WeVibe Chain for production environments.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Single Node Setup](#single-node-setup)
- [Multi-Node Production Setup](#multi-node-production-setup)
- [Genesis Configuration](#genesis-configuration)
- [Service Configuration](#service-configuration)
- [Security Hardening](#security-hardening)
- [Monitoring](#monitoring)
- [Backup and Recovery](#backup-and-recovery)
- [Upgrade Procedures](#upgrade-procedures)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Hardware Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | 4 cores | 8 cores |
| RAM | 8 GB | 32 GB |
| Storage | 100 GB SSD | 500 GB NVMe SSD |
| Network | 100 Mbps | 1 Gbps |

### Software Requirements

- Go 1.22+
- Docker 20.10+ (for containerized deployment)
- Cosmos SDK compatible tools
- TLS certificates (for production)

### System Setup

```bash
# Create dedicated user
sudo useradd -m -s /bin/bash wevibed

# Create necessary directories
sudo mkdir -p /var/lib/wevibed
sudo mkdir -p /etc/wevibed
sudo mkdir -p /var/log/wevibed

# Set ownership
sudo chown -R wevibed:wevibed /var/lib/wevibed
sudo chown -R wevibed:wevibed /var/log/wevibed
```

---

## Single Node Setup

### Build the Binary

```bash
cd /opt/wevibe-chain
make build
sudo cp build/wevibed /usr/local/bin/
sudo chmod +x /usr/local/bin/wevibed
```

### Initialize the Node

```bash
# Initialize
wevibed init my-node --chain-id wevibe-local-1 --home /var/lib/wevibed

# Create operator key
wevibed keys add operator --keyring-backend file --home /var/lib/wevibed

# Add genesis account
wevibed genesis add-genesis-account $(wevibed keys show operator -a --keyring-backend file --home /var/lib/wevibed) 1000000000uvibe --home /var/lib/wevibed
```

### Configure Epoch Duration

For production (daily epochs):

```bash
EPOCH_DURATION=86400s ./scripts/init-chain.sh --start
```

Or manually edit genesis:

```bash
jq '.app_state.epochs.epochs += [{
  "identifier": "wevibe_epoch",
  "duration": "86400s",
  "current_epoch": "0",
  "current_epoch_start_height": "0",
  "current_epoch_start_time": "0001-01-01T00:00:00Z",
  "epoch_counting_started": false
}]' /var/lib/wevibed/config/genesis.json > /tmp/genesis.json
mv /tmp/genesis.json /var/lib/wevibed/config/genesis.json
```

### Create Systemd Service

```ini
[Unit]
Description=WeVibe Chain Node
After=network-online.target
Wants=network-online.target

[Service]
User=wevibed
Group=wevibed
ExecStart=/usr/local/bin/wevibed start --home /var/lib/wevibed
Restart=always
RestartSec=3
LimitNOFILE=65535

# Storage
WorkingDirectory=/var/lib/wevibed
Environment=HOME=/var/lib/wevibed

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=wevibed

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable wevibed
sudo systemctl start wevibed
```

### Verify Operation

```bash
# Check status
sudo systemctl status wevibed

# View logs
sudo journalctl -u wevibed -f

# Check sync status
curl -s http://localhost:26657/status | jq '.result.sync_info'
```

---

## Multi-Node Production Setup

### Network Architecture

```
                    ┌─────────────────┐
                    │   Load Balancer  │
                    │   (gRPC/REST)    │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
   ┌────┴────┐          ┌────┴────┐          ┌────┴────┐
   │ Validator│          │ Validator│          │ Validator│
   │   Node   │          │   Node   │          │   Node   │
   └─────────┘          └─────────┘          └─────────┘
        │                    │                    │
        └────────────────────┼────────────────────┘
                             │
                    ┌────────┴────────┐
                    │  Seed/Peers   │
                    │   (P2P Net)   │
                    └───────────────┘
```

### Validator Setup

#### 1. Initialize Validator Node

```bash
# Initialize
wevibed init validator-1 --chain-id wevibe-1 --home /var/lib/wevibed

# Create validator key
wevibed keys add validator --keyring-backend file --home /var/lib/wevibed

# Add to genesis (pre-launch only)
wevibed genesis add-genesis-account $(wevibed keys show validator -a --keyring-backend file --home /var/lib/wevibed) 1000000000uvibe --home /var/lib/wevibed
```

#### 2. Create Validator Metadata

```bash
# Create validator.json
cat > /var/lib/wevibed/config/validator.json << EOF
{
  "name": "WeVibe Validator 1",
  "identity": "example.com/validator1",
  "website": "https://example.com",
  "security_contact": "security@example.com",
  "details": "Professional validator operator"
}
EOF
```

#### 3. Create Tendermint Validator

```bash
# Create gentx (pre-launch)
wevibed genesis gentx validator 500000000uvibe \
  --chain-id wevibe-1 \
  --home /var/lib/wevibed \
  --keyring-backend file \
  --moniker validator-1

# Collect gentxs (launch coordinator only)
wevibed genesis collect-gentxs --home /var/lib/wevibed
```

#### 4. Configure P2P Networking

Edit `/var/lib/wevibed/config/config.toml`:

```toml
[p2p]
pex = true
seed_mode = false

# Persistent peers
persistent_peers = "peer1@node1:26656,peer2@node2:26656"

# Seeds
seeds = "seed@seed.example.com:26656"

[laddr]
laddr = "tcp://0.0.0.0:26656"

# Private peer IDs (for sentry nodes)
private_peer_ids = "validator_id"
```

#### 5. State Sync (for快速同步)

```toml
[statesync]
enable = true
rpc_servers = "node1:26657,node2:26657"
trust_height = 1000000
trust_hash = "abc123..."
trust_period = "168h"
```

### Sentry Node Architecture

For DDoS protection:

```
                    ┌─────────────┐
                    │   Internet   │
                    └──────┬──────┘
                           │
               ┌───────────┴───────────┐
               │                       │
          ┌────┴────┐            ┌────┴────┐
          │  Sentry │            │  Sentry │
          │   Node  │            │   Node  │
          └────┬────┘            └────┬────┘
               │                       │
               └───────────┬───────────┘
                           │
                    ┌──────┴──────┐
                    │  Validator  │
                    │    Node     │
                    └─────────────┘
```

#### Sentry Configuration

```toml
[p2p]
pex = true
persistent_peers = "validator_id@sentry1:26656,validator_id@sentry2:26656"
unconditional_peer_ids = "validator_id"
```

#### Validator Configuration

```toml
[p2p]
pex = false
persistent_peers = "sentry1_id@sentry1:26656,sentry2_id@sentry2:26656"
private_peer_ids = "validator_id"
```

---

## Genesis Configuration

### Production Genesis Template

```json
{
  "app_state": {
    "epochs": {
      "epochs": [
        {
          "identifier": "wevibe_epoch",
          "duration": "86400s",
          "current_epoch": "0",
          "current_epoch_start_height": "0",
          "current_epoch_start_time": "0001-01-01T00:00:00Z",
          "epoch_counting_started": true
        }
      ]
    },
    "mint": {
      "minter": {
        "inflation": "0.000000000000000000",
        "annual_provisions": "0.000000000000000000"
      },
      "params": {
        "mint_denom": "uvibe",
        "inflation_rate_change": "0.000000000000000000",
        "inflation_max": "0.000000000000000000",
        "inflation_min": "0.000000000000000000",
        "goal_bonded": "0.670000000000000000",
        "blocks_per_year": "4360000"
      }
    },
    "gov": {
      "params": {
        "min_deposit": [{"denom": "uvibe", "amount": "10000000"}],
        "max_deposit_period": "172800s",
        "voting_period": "172800s",
        "quorum": "0.334000000000000000",
        "threshold": "0.500000000000000000",
        "veto_threshold": "0.334000000000000000"
      }
    },
    "emissions": {
      "params": {
        "daily_mint_amount": "1000000",
        "operator_share_percent": "10",
        "validator_share_percent": "5",
        "storage_weight_percent": "40",
        "retrieval_weight_percent": "60",
        "rarity_multiplier_cap": "10",
        "bootstrap_duration_epochs": "180"
      }
    },
    "bandwidth": {
      "params": {
        "default_memory_cap_per_epoch": "1000",
        "default_serve_cap_per_epoch": "10000"
      }
    },
    "reputation": {
      "params": {
        "active": true,
        "max_difficulty": 10,
        "max_quality": 10,
        "serve_xp_per_serve": "1",
        "self_serve_xp_per_serve": "0"
      }
    }
  }
}
```

### Validate Genesis

```bash
wevibed genesis validate-genesis /var/lib/wevibed/config/genesis.json
```

---

## Service Configuration

### Environment Variables

```bash
# /etc/wevibed/environment
WEVIBE_HOME=/var/lib/wevibed
WEVIBE_LOG_LEVEL=info
WEVIBE_PRUNING=syncable
WEVIBE_MIN_GAS_PRICE=0.0001uvibe
WEVIBE_HALT_HEIGHT=0
```

### Application Configuration

```bash
cat > /var/lib/wevibed/config/app.toml << EOF
# Toml config for wevibed

[api]
enable = true
address = "tcp://0.0.0.0:1317"
swagger = true

[api.grpc]
enable = true
address = "0.0.0.0:9090"

[grpc-web]
enable = true
address = "0.0.0.0:9091"

[rosetta]
enable = false

[state-sync]
snapshot-interval = 1000
snapshot-keep-recent = 10

[store]
streaming.enabled = true
EOF
```

### gRPC Gateway CORS

```bash
cat > /var/lib/wevibed/config/cors.toml << EOF
[ws]
allowed-origins = ["https://dashboard.wevibe.network"]
allowed-credentials = true
EOF
```

---

## Security Hardening

### Firewall Configuration

```bash
# UFW example
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 26656/tcp   # P2P
sudo ufw allow 26657/tcp   # RPC
sudo ufw allow 1317/tcp    # REST API
sudo ufw allow 9090/tcp     # gRPC
sudo ufw enable
```

### TLS Configuration for RPC

```bash
# Generate certificates
sudo openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes \
  -subj "/CN=wevibed.example.com"

# Configure nginx reverse proxy
cat > /etc/nginx/sites-available/wevibed << EOF
server {
    listen 443 ssl;
    server_name wevibed.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://127.0.0.1:26657;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
EOF
```

### Key Security

1. **Use Hardware Wallets**: Store validator keys on Ledger/Trezor
2. **Use File Keyring**: For non-validator keys, use file keyring with secure permissions
3. **Rotate Keys**: Regularly rotate operational keys

```bash
# Secure keyring permissions
chmod 700 /var/lib/wevibed/keyring-file
chown wevibed:wevibed /var/lib/wevibed/keyring-file
```

### Sentry Node Isolation

- Run sentry nodes in separate security groups
- Use VPN or private networking between sentries and validators
- Never expose validator RPC directly to internet

---

## Monitoring

### Prometheus Metrics

Enable Prometheus in `config.toml`:

```toml
[prometheus]
enable = true
address = "0.0.0.0:26660"
```

### Key Metrics to Monitor

| Metric | Alert Threshold | Description |
|--------|---------------|-------------|
| `wevibed_peers` | < 3 | Low peer count |
| `wevibed_mempool_size` | > 1000 | Mempool congestion |
| `wevibed_consensus_height` | stuck | Consensus stall |
| `wevibed_processor_cpu` | > 80% | High CPU |
| `wevibed_processor_mem` | > 90% | High memory |

### Alerting Rules

```yaml
groups:
  - name: wevibed
    rules:
      - alert: WevibedPeersLow
        expr: wevibed_peers < 3
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Low peer count on {{ $labels.instance }}"
      
      - alert: WevibedConsensusStalled
        expr: rate(wevibed_consensus_height[5m]) == 0
        for: 10m
        labels:
          severity: critical
        annotations:
          summary: "Consensus stalled on {{ $labels.instance }}"
```

### Dashboard

Import Grafana dashboard JSON from Cosmos SDK templates.

---

## Backup and Recovery

### Backup Commands

```bash
# Backup genesis
cp /var/lib/wevibed/config/genesis.json /backup/genesis-$(date +%Y%m%d).json

# Backup validator state
tar -czf /backup/priv_validator_state-$(date +%Y%m%d).tar.gz \
  -C /var/lib/wevibed/data .
```

### Automated Backup Script

```bash
#!/bin/bash
# /usr/local/bin/backup-wevibed.sh

DATE=$(date +%Y%m%d-%H%M)
BACKUP_DIR=/backup/wevibed
KEEP_DAYS=7

mkdir -p $BACKUP_DIR

# Stop node
sudo systemctl stop wevibed

# Backup data
tar -czf $BACKUP_DIR/wevibed-data-$DATE.tar.gz /var/lib/wevibed/data

# Backup genesis
cp /var/lib/wevibed/config/genesis.json $BACKUP_DIR/genesis-$DATE.json

# Start node
sudo systemctl start wevibed

# Cleanup old backups
find $BACKUP_DIR -name "*.tar.gz" -mtime +$KEEP_DAYS -delete
find $BACKUP_DIR -name "genesis-*.json" -mtime +$KEEP_DAYS -delete
```

### Recovery Procedure

```bash
# 1. Stop node
sudo systemctl stop wevibed

# 2. Restore data
tar -xzf /backup/wevibed-data-YYYYMMDD.tar.gz -C /var/lib/wevibed/data

# 3. Verify
wevibed genesis validate-genesis /var/lib/wevibed/config/genesis.json

# 4. Start node
sudo systemctl start wevibed
```

---

## Upgrade Procedures

### Preparations

1. **Monitor chain height** before upgrade window
2. **Backup state** before applying upgrade
3. **Test upgrade** on testnet first

### Manual Upgrade

```bash
# 1. Download new binary
wget https://github.com/wevibe-network/wevibe-chain/releases/v1.1.0/wevibed
chmod +x wevibed
sudo mv wevibed /usr/local/bin/

# 2. Verify version
wevibed version

# 3. Stop node
sudo systemctl stop wevibed

# 4. Apply upgrade (if usingCosmovisor)
export DAEMON_NAME=wevibed
export DAEMON_HOME=/var/lib/wevibed
cosmovisor run start --home $DAEMON_HOME

# Or manually restart
sudo systemctl start wevibed
```

### Cosmovisor Setup

```bash
# Install Cosmovisor
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@latest

# Setup directories
mkdir -p ~/.cosmovisor/genesis/bin
mkdir -p ~/.cosmovisor/upgrades/v1.1.0/bin

# Link current binary
ln -s /usr/local/bin/wevibed ~/.cosmovisor/genesis/bin/wevibed

# Enable auto-download
export DAEMON_ALLOW_DOWNLOAD_BINARIES=true
```

---

## Troubleshooting

### Common Issues

#### Node Won't Start

```bash
# Check logs
sudo journalctl -u wevibed -n 100

# Common causes:
# - Port already in use
# - Corrupted genesis
# - Wrong chain ID
```

#### Peers Not Connecting

```bash
# Check firewall
sudo ufw status

# Verify addresses
wevibed tendermint show-node-id --home /var/lib/wevibed

# Test connectivity
nc -zv peer.example.com 26656
```

#### State Sync Fails

```bash
# Check trust store
ls -la /var/lib/wevibed/data/

# Verify RPC servers are synced
curl -s http://node1:26657/status | jq '.result.sync_info'

# Manual state sync
wevibed export > /tmp/state.json
```

#### Memory Issues

```bash
# Check memory usage
free -h

# Increase limits in systemd
sudoedit /etc/systemd/system/wevibed.service
# Add: MemoryMax=32G
sudo systemctl daemon-reload
sudo systemctl restart wevibed
```

### Recovery from Snapshot

```bash
# 1. Stop node
sudo systemctl stop wevibed

# 2. Backup current state
mv /var/lib/wevibed/data /var/lib/wevibed/data.old

# 3. Download snapshot
wget https://snapshots.wevibe.network/latest.tar.gz
tar -xzf latest.tar.gz -C /var/lib/wevibed

# 4. Fix permissions
chown -R wevibed:wevibed /var/lib/wevibed/data

# 5. Start node
sudo systemctl start wevibed
```

### Governance Issues

```bash
# Check proposal status
wevibed query gov proposals --status voting_period

# Vote manually
wevibed tx gov vote 1 yes --from validator --chain-id wevibe-1
```

---

## Quick Reference

### Essential Commands

```bash
# Start node
sudo systemctl start wevibed

# Stop node
sudo systemctl stop wevibed

# Check status
sudo systemctl status wevibed

# View logs
sudo journalctl -u wevibed -f

# Check sync
curl -s http://localhost:26657/status | jq '.result.sync_info.latest_block_height'

# Check peers
curl -s http://localhost:26657/net_info | jq '.result.peers | length'

# Validate genesis
wevibed genesis validate-genesis

# Check module params
wevibed query <module> params
```
