# WeVibe Chain Roadmap

This roadmap tracks the current public implementation status and planned milestones for WeVibe Chain.

## Purpose

WeVibe Chain is WeVibe's sovereign Cosmos SDK + CometBFT appchain. It is the network's source of truth for encrypted organization memory, membership and roles, serve and denial attestations, contributor reputation aggregates, and VIBE economic state.

Validators replicate encrypted memory state, but do not have plaintext visibility in the standard memory flow.

Runtime profile:

- Binary: `wevibed`
- CometBFT RPC: `26657`
- gRPC: `9090`
- REST (gRPC-gateway): `1317`

## Status — working today

- 8 custom modules are wired in-app (`x/org`, `x/memory`, `x/serve`, `x/emissions`, `x/bandwidth`, `x/reputation`, `x/attestation`, `x/identity`).
- Ascending-price org-slot acquisition is live via the slot registry and slot pricing path.
- A fixed 32-year emissions schedule is seeded with per-epoch pool math and contributor attribution.
- The Earned Trust memory decay model is active in the epoch-end hook.
- Genesis seeding and epoch-hook resilience paths are in place.

## Status — present but not fully wired

- Emissions disbursement is currently tracked as accrued state, but end-to-end coin movement for protocol disbursement is still being completed. In the interim, validators continue to earn standard staking rewards.
- `x/attestation` is present but transaction handling is intentionally disabled (no-op) pending a pluggable session-attestation framework.
- The following economics are designed but not yet built on-chain: demand-leg payment routing, Harberger self-assessed rent, Dutch resale of freed slots, and per-memory storage deposits. (The storage-deposit parameter is currently near-zero on testnet.)

## Near-term

- Wire emissions disbursement end-to-end (mint + claim path) with reentrancy and double-claim protections.
- Keep rewards keyed to passkey public keys and withdrawable after wallet-link resolution.
- Add an on-chain demand-leg burn router.
- Add chain-resolved hub endpoint state with a leader-signed setter flow.
- Add chain-recorded report lifecycle (`report` / `clear-report`) with storage-deposit clawback semantics.

## Mainnet priorities

- Activate per-memory storage deposit as the primary anti-spam mechanism.
- Remove `x/bandwidth` (testnet guard) once storage deposits are active.
- Enable Harberger rent with a forced-sale window.
- Enable Dutch resale for freed slots.
- Freeze badge rarity tier logic on-chain.

## Future (v2)

- Activate `x/attestation` as the pluggable session-attestation framework feeding two-layer difficulty scoring.

## Design references

- Canonical protocol references: [wevibe-docs / WHITEPAPER.md](https://github.com/WeVibe-Network/wevibe-docs/blob/main/WHITEPAPER.md), [wevibe-docs / TOPOLOGY.md](https://github.com/WeVibe-Network/wevibe-docs/blob/main/TOPOLOGY.md)
- In-repo technical references: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md), [docs/MODULES.md](docs/MODULES.md), [docs/PARAMETERS.md](docs/PARAMETERS.md)
