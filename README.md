<div align="center">

<img src="https://capsule-render.vercel.app/api?type=waving&color=0:02100a,100:2fe07a&height=160&section=header&text=WeVibe%20Chain&fontColor=54f59a&fontSize=42&fontAlignY=40&desc=Sovereign%20appchain%20for%20encrypted%20memory%20and%20reputation&descAlignY=64&descSize=16" alt="WeVibe Chain" width="100%" />

![Go](https://img.shields.io/badge/Go-00ADD8?style=flat-square&logo=go&logoColor=white)
![Cosmos SDK](https://img.shields.io/badge/Cosmos%20SDK-1B1F3B?style=flat-square)
[![status-alpha](https://img.shields.io/badge/status-alpha-ffc266?style=flat-square)](https://github.com/WeVibe-Network)
[![license-Apache--2.0](https://img.shields.io/badge/license-Apache--2.0-82aaff?style=flat-square)](LICENSE)
[![docs-wevibe-docs](https://img.shields.io/badge/docs-wevibe--docs-54f59a?style=flat-square)](https://github.com/WeVibe-Network/wevibe-docs)
[![%40WeVibe__Network](https://img.shields.io/badge/%40WeVibe__Network-0a0a0a?style=flat-square&logo=x&logoColor=white)](https://x.com/WeVibe_Network)

</div>

---

**WeVibe** is a memory layer for AI coding agents: attributed, encrypted, human-gated memory that crosses trust boundaries. Contributed memories carry provenance, and recalled items are decrypted locally under member-held keys — nothing reaches an agent's context without passing human review.

**wevibe-chain** (binary `wevibed`) is the sole durable authority in that system. It is a Cosmos SDK + CometBFT appchain that records organization membership, memory commitments, and consumer-signed evidence events. Everything else — hubs, dashboards, caches, indexes — is disposable and rebuildable from chain state plus the keys members hold.

## Evidence, Not Verdicts

WeVibe Chain deliberately refuses to judge. It is not a trust-score-in-consensus chain: no standing, weight, score, trust value, or memory content is ever stored on-chain. It is an append-only, content-free, consumer-signed event log.

Standing is computed at the edge as a deterministic pure function of the evidence and an anchored policy version:

```
standing = f(events, policy_version)
```

The policy is versioned code that lives outside consensus; only its hash is anchored on-chain. Because the record is immutable and the interpretation is revisable, rewriting history has nothing to alter. Because nothing is ever deleted, killing knowledge has nothing to grip. Because anyone can recompute standing from public inputs, no operator can quietly decide what a stranger's knowledge is worth.

The chain carries evidence; the edge carries judgment; the bench carries measurement.

## Four Exit Guarantees

The sovereignty contract: no single party — including WeVibe, any hub operator, or any org leader — holds unilateral ability to

1. **READ** a member's memory plaintext from outside;
2. **WITHHOLD** the network's function from a principal acting within their rights;
3. **REWRITE** the historical record; or
4. **KILL** an organization's knowledge or a contributor's standing by withdrawing infrastructure.

The append-only evidence log plus edge-computed standing is what makes (3) and (4) nothing to aim at.

## Status

Alpha / testnet. Core modules are wired and running; selected economics and attestation paths are design-intent or partially wired, not production-complete. Event types E4 (contest) and E5 (sponsorship) are parked — enum slots reserved only. Rollout detail lives in [ROADMAP.md](ROADMAP.md).

## The Event Log

Evidence events are consumer-signed and content-free: hashes and references rather than payloads. They enter the log via the hub relay — the only submission path — where the chain enforces a serving-key signer gate. Serve and denial evidence land as receipts; outcome, validity, cost, and convergence evidence land in the event log:

| Code | Event | Stored as |
|------|-------|-----------|
| E1 | Serve — proof of content delivery | `StoredServeReceipt` |
| E2 | Block — proof of delivery denial | `StoredDenialReceipt` (`neg_anchor` slot reserved, inert) |
| E3 | Outcome — use-leg: whether a served memory worked in the consuming task (`worked` \| `didnt_work` \| `unobserved`) | `StoredEvent` |
| E4 | Contest | parked — enum slot only |
| E5 | Sponsorship | parked — enum slot only |
| E6 | Validity-predicate result | `StoredEvent` |
| E7 | Cost-to-discover evidence | `StoredEvent` |
| E8 | Convergence of independent discoveries | `StoredEvent` |

E3 is the load-bearing event: an unobserved use is recorded as unobserved — silence is not a vote.

## Policy Anchor

Standing is never written on-chain; the policy-version hash is anchored instead. `StoredPolicyAnchor` records `(policy_version, policy_hash, anchored_at_epoch, anchored_at_height)`. The initial anchor is seeded at genesis (env-gated on `WEVIBE_EDGE_POLICY_FILE`); re-anchoring is authority-gated via `MsgAnchorPolicyVersion`.

Current testnet anchor: `edge-policy-v1`, sha256 `2d2faa14461aa51bb72735b05debf30defff039750e5f90c1922ae813c87899e`, anchored at height 45.

## Modules

21 modules total: 8 custom, plus 13 from Cosmos SDK (auth, bank, staking, distribution, slashing, mint, consensus, genutil, epochs, gov, feegrant, authz, upgrade).

| Module | Role |
|--------|------|
| `x/org` | Organization registry and slot registry. A sequential ascending slot-price schedule with 50% of each slot fee burned (module holds the Burner permission). Single fixed epoch. |
| `x/memory` | Commitment lifecycle (`PENDING` → `PENDING_KEYWORD` / `PENDING_CHAIN` → `COMMITTED`, with `DENIED`, `ARCHIVED`, `REPORTED_DELETED`), provenance fields, validity metadata, per-epoch Merkle roots, relationships, upheld reports. Keywords are inert labels only — no time-based degradation, no trust scores, no weighting. |
| `x/serve` | The append-only evidence log (E1–E8) and the `StoredPolicyAnchor` registry. |
| `x/emissions` | Flat daily mint plus a flat even-split contributor reward over approval-qualified contributors (approval count ≥ threshold). The validator share is accounted, not minted here. |
| `x/bandwidth` | Per-org, per-epoch caps on memory serves and serve events. |
| `x/reputation` | Contributor / moderator / leader profile counters: XP, contributions, serve records, bans. |
| `x/attestation` | Session-attestation receipt store — producer provenance evidence. |
| `x/identity` | Passkey-first alias registry; wallet binding is optional (bound via `migrate-identity`). |

## Framework and Ports

- **Stack:** Go 1.25, Cosmos SDK v0.53.5, CometBFT v0.38.20.
- **Denom / prefix:** `uvibe` (1 VIBE = 1,000,000 uvibe); address prefix `wevibe`; binary `wevibed`.
- **Time:** the `wevibe_epoch` identifier drives time-based processing. Local development defaults to 60 seconds, configurable through `WEVIBE_EPOCH_DURATION_SECONDS` in `scripts/init-chain.sh`.

SDK defaults bind all client interfaces to loopback. The repo's `scripts/init-chain.sh` opens them up for **local development only**:

| Endpoint | Port | SDK default | `init-chain.sh` (local dev) |
|----------|------|-------------|------------------------------|
| RPC | 26657 | loopback | `0.0.0.0` (wildcard CORS) |
| gRPC | 9090 | loopback | `0.0.0.0` |
| REST | 1317 | loopback | `0.0.0.0` with `enabled-unsafe-cors` |

## Getting Started

Prerequisites: Go 1.25+, Docker (for the compose-based localnet).

```bash
# Dependencies and binary
make deps
make build
```

Run the single-validator local network (Docker Compose):

```bash
make localnet-start
make localnet-logs
make localnet-stop
make localnet-reset   # destroys chain data
```

Tests and lint:

```bash
make test
make test-integration
make test-verbose
make lint
```

Basic CLI:

```bash
wevibed init <moniker> --chain-id wevibe-local-1
wevibed start
wevibed query wevibe org v1 org <org_id>
```

Module query and transaction namespaces are discoverable via `wevibed query wevibe --help` and `wevibed tx wevibe --help`.

## Documentation

- [Architecture](docs/ARCHITECTURE.md) — system topology and data flow
- [Topology](docs/TOPOLOGY.md) — network component layout
- [Module Reference](docs/MODULES.md) — detailed module specifications
- [API Reference](docs/API.md) — gRPC and REST endpoints
- [CLI Reference](docs/CLI.md) — daemon and client commands
- [Parameters](docs/PARAMETERS.md) — module parameter catalogue
- [Deployment](docs/DEPLOYMENT.md) — production deployment guide
- [Contributing](CONTRIBUTING.md) — development guidelines
- [Roadmap](ROADMAP.md) — rollout status

## License

Apache 2.0 — see [LICENSE](LICENSE) for details.

## Links

- Canonical docs: https://github.com/WeVibe-Network/wevibe-docs
- WeVibe Network org: https://github.com/WeVibe-Network
- X/Twitter: https://x.com/WeVibe_Network

