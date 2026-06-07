# WeVibe Chain

WeVibe Chain is WeVibe Network's sovereign Cosmos SDK + CometBFT appchain. It is the source of truth for encrypted organizational memory, membership and role state, serve attestations, contributor reputation aggregates, and VIBE economic state.

## Overview

WeVibe Chain couples the staking, governance, bank, epochs, and distribution foundations of Cosmos SDK with purpose-built modules that orchestrate:

- **Organization Slots & Membership** — Registration, role management, and treasuries
- **Memory Curation** — Encrypted commitments with lifecycle and epoch hooks
- **Serve & Denial Attestations** — Delivery/denial evidence, deduplication, and per-epoch stats
- **Bandwidth Throttling** — Per-org rate limiting for submission and serve traffic
- **Reputation Tracking** — Contributor XP, serve stats, and cross-org profiles
- **Emission Accounting** — Fixed-schedule pool math and contributor reward accrual
- **Session Attestation Surface** — Session attestation schema and APIs (currently disabled)
- **Identity Linking** — Passkey-to-wallet aliasing used for migration and reward claims

### Status

WeVibe Chain is in active alpha/testnet development. Core modules are wired and running, while selected economic and attestation paths remain under active rollout (see [ROADMAP.md](ROADMAP.md)).

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
| `x/org` | Slot-based organization registry, membership roles, treasury balances, and serving configuration |
| `x/memory` | Encrypted memory commitments, lifecycle transitions, Earned Trust decay, and epoch Merkle roots |
| `x/serve` | Serve and denial attestation ingestion, nullifier deduplication, matched-keyword indexing, and epoch stats |
| `x/emissions` | Emission-pool schedule accounting, per-epoch attribution, and contributor reward ledgers |
| `x/bandwidth` | Per-org, per-epoch submission/serve rate limits and overrides |
| `x/reputation` | Contributor XP, serve metrics, and cross-org profile aggregates |
| `x/attestation` | Session-attestation storage/query surface (message path currently disabled) |
| `x/identity` | Passkey-to-wallet identity aliasing and migration records |

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

### Epoch Processing

At the end of each `wevibe_epoch`:
1. `x/emissions` advances fixed-schedule emission accounting and contributor attribution.
2. `x/memory` runs lifecycle maintenance (including decay/expiry handling) and stores epoch Merkle roots.
3. Ongoing economic rollout details are tracked in [ROADMAP.md](ROADMAP.md).

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

## Roadmap

For rollout status and planned milestones, see [ROADMAP.md](ROADMAP.md).

## License

Apache 2.0 - see LICENSE file for details.

## Links

- Canonical docs: https://github.com/WeVibe-Network/wevibe-docs
- WeVibe Network org: https://github.com/WeVibe-Network
- X/Twitter: https://x.com/WeVibe_Network
