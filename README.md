# WeVibe Chain

WeVibe Chain is a Cosmos SDK application that serves as the anchor for WeVibe Network's encrypted organizational memory system. It provides organizations with on-chain capabilities for registering identities, curating encrypted knowledge, attesting to content retrieval, tracking contributor reputation, and distributing incentive payouts.

## Overview

WeVibe Chain couples the staking, governance, bank, epochs, and distribution foundations of Cosmos SDK with purpose-built modules that orchestrate:

- **Organization Management** — Registration, membership, and treasuries
- **Memory Curation** — Encrypted knowledge commitments with lifecycle states
- **Serve Attestations** — Proof of content delivery by serving agents
- **Bandwidth Throttling** — Per-org rate limiting for spam prevention
- **Reputation Tracking** — Contributor XP, stats, and cross-org profiles
- **Emission Incentives** — Epoch-based payouts funded by org treasuries

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        WeVibe Chain                                 │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐  │
│  │   org   │ │ memory  │ │  serve  │ │bandwidth│ │   ...   │  │
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └─────────┘  │
│       │           │           │           │                     │
│       └───────────┴─────┬─────┴───────────┘                     │
│                         │                                       │
│                    Epoch Hooks                                  │
│                         │                                       │
│              ┌─────────┴─────────┐                             │
│              │  Emissions + Memory  │                            │
│              └────────────────────┘                             │
└─────────────────────────────────────────────────────────────────┘
                          │
           ┌──────────────┼──────────────┐
           │              │              │
           ▼              ▼              ▼
    ┌──────────┐  ┌──────────┐  ┌──────────┐
    │WeVibe Hub  │  │WeVibe Dash │  │  Agents  │
    └──────────┘  └──────────┘  └──────────┘
```

## Modules

| Module | Purpose |
|--------|---------|
| `x/org` | Organization registry, membership, and treasury management |
| `x/memory` | Memory commitment lifecycle, relationships, and contests |
| `x/serve` | Serve attestation ingestion and deduplication |
| `x/bandwidth` | Per-org rate limiting for submissions and serves |
| `x/reputation` | Contributor reputation tracking and XP |
| `x/emissions` | Daily minting, epoch payouts, and work scores |

## Key Concepts

### Organizations

Organizations register on WeVibe Chain by burning VIBE tokens via a dynamic pricing curve. Each org maintains:
- A **leader** who controls approvals and configuration
- **Members** with roles for collaborative management
- A **treasury** funded by org administrators for incentive payouts

### Memory Lifecycle

Memories progress through defined states based on approval, serve activity, and disputes:

```
PENDING → APPROVED → STABLE ←→ DEGRADED ←→ DORMANT
                ↓         ↓
            CONTESTED   ARCHIVED
                ↓
            REJECTED (upheld contest)
```

### Serve Attestations

Serving agents batch proofs of content delivery. Attestations are:
- Deduplicated via nullifiers to prevent replay
- Rate-limited per org per epoch
- Tracked per contributor for reputation and payouts

### Epoch Payouts

At the end of each `wevibe_epoch`:
1. Daily emission is minted to the pool
2. Orgs with `serve_attestation_required=true` have their treasuries debited
3. Contributors receive payouts based on serve counts and org-configured tiers

## Getting Started

### Prerequisites

- Go 1.22+
- Docker (for local development)
- Cosmos SDK dependencies (handled via `go mod`)

### Build

```bash
# Clone and navigate
cd wevibe-chain

# Download dependencies
make deps

# Build the chain binary
make build
```

### Run Local Network

```bash
# Start a single-validator local network
make localnet-start

# View logs
make localnet-logs

# Stop the network
make localnet-stop

# Reset the network (destroys data)
make localnet-reset
```

### Run Tests

```bash
# All tests
make test

# Integration tests only
make test-integration

# With verbose output
make test-verbose

# Linting
make lint
```

### CLI Usage

```bash
# Initialize a new node
wevibed init <moniker> --chain-id wevibe-local-1

# Start the node
wevibed start

# Query org details
wevibed query wevibe org v1 org <org_id>

# Submit a transaction
wevibed tx wevibe org register-org [args...]
```

## Chain Configuration

### Denomination

- **Token**: VIBE
- **Smallest unit**: uvibe (1 VIBE = 1,000,000 uvibe)
- **Address prefix**: `wevibe`

### Epoch Configuration

The `wevibe_epoch` identifier drives all time-based logic:
- **Local development**: 60 seconds
- **Production**: 86,400 seconds (daily)

Configure via `scripts/init-chain.sh` using the `EPOCH_DURATION` environment variable.

### Governance

Governance controls critical module parameters. Default periods for local development:
- **Voting period**: 172,800 seconds (2 days)
- **Deposit period**: 172,800 seconds (2 days)
- **Minimum deposit**: 10,000,000 uvibe

## Network Topology

### Validators and Full Nodes

Validators run WeVibe Chain binaries, expose P2P ports, and optionally enable gRPC for client queries. Full nodes mirror this setup without signing blocks. Both expose the gRPC gateway for REST access to module query services.

### WeVibe Hub

Hub is a stateless service that:
- Subscribes to Tendermint WebSocket events
- Calls module query endpoints over gRPC
- Projects results into durable caches and push streams
- Brokers notifications to Dashboard

### WeVibe Dashboard

Dashboard provides an interactive interface for:
- Organization leaders to manage membership and treasuries
- Contributors to submit and track memories
- Analysts to view metrics and payouts

## Documentation

- [Architecture](docs/ARCHITECTURE.md) — System topology and data flow
- [Module Reference](docs/MODULES.md) — Detailed module specifications
- [API Reference](docs/API.md) — gRPC and REST API endpoints
- [CLI Reference](docs/CLI.md) — Daemon and client commands
- [Parameters](docs/PARAMETERS.md) — Module parameter catalogue
- [Deployment](docs/DEPLOYMENT.md) — Production deployment guide
- [Contributing](CONTRIBUTING.md) — Development guidelines

## License

Apache 2.0 - see LICENSE file for details.
