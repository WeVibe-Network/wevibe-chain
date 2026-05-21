# WeVibe Chain Architecture

This document describes the system architecture, module topology, data flow, and keeper dependencies for WeVibe Chain.

## System Components

### WeVibe Chain

WeVibe Chain is the Cosmos SDK application anchoring WeVibe Network's encrypted organizational memory. It exposes two public APIs:

- **Tendermint RPC** — Block and transaction gossip
- **gRPC-gateway REST** — Derived from protobuf services in `proto/wevibe/**`

Client authentication uses bech32 addresses with the `wevibe` prefix. Fees are denominated in `vibe`. Governance retains authority over critical parameters via the `gov` module address.

### WeVibe Hub

WeVibe Hub is the stateless service tier that:
- Watches new blocks via WebSocket subscriptions
- Materializes module-specific views (live memories per org, serve statistics)
- Pushes normalized payloads over WebSockets and gRPC streams
- Relays event notifications to Dashboard and analytics pipelines

Hub relies on the same query services documented below and remains a pure observer— no keeper logic runs outside the chain.

### WeVibe Dashboard

WeVibe Dashboard is the interactive client for organization leaders, contributors, and analysts:
- Consumes Hub streams for live metrics
- Calls chain gRPC endpoints directly for authoritative reads
- Assembles transaction payloads and relays them through wallet connectors

## Module Topology

### Org Module (`x/org`)

The org module manages organization registration, membership, and treasuries.

#### State Objects

| Prefix | Type | Description |
|--------|------|-------------|
| `org/` | `StoredOrg` | Organization metadata (leader, creation height, renewal height, storage quota, retrieval budget, status, domain) |
| `member/{org}/{pubkey}` | `StoredMemberRecord` | Membership roles (leader plus custom roles) |
| `dynprice/` | `StoredDynamicPrice` | Dynamic burn price inputs for new registrations |
| `treasury/{org}` | `StoredTreasury` | Organization treasury balance in string-encoded `vibe` |
| `reptier/{org}` | `StoredRepTierConfig` | Ordered payout tiers keyed by reputation bounds |
| `orgconfig/{org}` | `StoredOrgConfig` | Runtime toggles (serve_attestation_required, decay_rate_bps, contest_stake_vibe) |
| `params` | `Params` | Module-level fees, quota defaults, burn behavior |

#### Messages

| Message | Description |
|---------|-------------|
| `MsgRegisterOrg` | Burn WEVIBE at dynamic price, create org with designated leader |
| `MsgAddMember` | Add member to organization with assigned role |
| `MsgRemoveMember` | Remove member from organization's roster |
| `MsgFundTreasury` | Add funds from signer to org treasury |
| `MsgWithdrawTreasury` | Remove funds from org treasury to recipient |
| `MsgSetRepTiers` | Configure payout tiers for org |
| `MsgSetOrgConfig` | Set org configuration (serve attestation required, decay rate, contest stake) |
| `MsgGrantTrialAllowance` | Grant fee grant for trial submissions |
| `MsgUpdateParams` | Governance parameter update |

#### Query Endpoints

| Endpoint | Path |
|----------|------|
| `GetOrg` | `/wevibe/org/v1/org/{org_id}` |
| `GetMembers` | `/wevibe/org/v1/members/{org_id}` |
| `IsMember` | `/wevibe/org/v1/is_member/{org_id}/{pubkey}` |
| `IsModerator` | `/wevibe/org/v1/is_moderator/{org_id}/{pubkey}` |
| `GetTreasury` | `/wevibe/org/v1/treasury/{org_id}` |
| `GetRepTiers` | `/wevibe/org/v1/rep_tiers/{org_id}` |
| `GetOrgConfig` | `/wevibe/org/v1/config/{org_id}` |
| `Params` | `/wevibe/org/v1/params` |

#### Keeper Dependencies

- `BankKeeper` — Burns and treasury transfers

---

### Memory Module (`x/memory`)

The memory module manages organizational memory commitments with lifecycle states, relationships, and dispute resolution.

#### State Objects

| Prefix | Type | Description |
|--------|------|-------------|
| `pending/{org}/...` | `StoredPendingCommitment` | Content hash, keywords, contributor, epoch, submission height |
| `approved/{org}/...` | `StoredMemoryCommitment` | Encrypted blob, lifecycle state, retrieval confidence, approver |
| `relationship/{org}:{source}:{target}` | `StoredMemoryRelationship` | CONTRADICTS/REPLACES/DEPRECATES/SUPERSEDES links |
| `contest/{org}:{contestID}` | `StoredMemoryContest` | Challenger, stake, epoch, resolution state |
| `validity/{org}:{cid}` | `StoredValidityMetadata` | Validity windows and scope tags |
| `merkle/{org}/{epoch}` | `StoredEpochMerkleRoot` | Cached Merkle root plus memory count |
| `count/{org}` | counter | Approved memory count per org |
| `params` | `Params` | Queue sizes, blob size, confidence thresholds, decay floor, contest window |

#### Memory States

| State | Description |
|-------|-------------|
| `PENDING` | Submitted but not yet approved |
| `APPROVED` | Default post-review state, eligible for recall |
| `STABLE` | Confidence above stable threshold, preferred ordering |
| `DEGRADED` | Confidence below degraded threshold, low priority |
| `DORMANT` | Confidence below dormant threshold, hidden from default recall |
| `CONTESTED` | Active dispute, excluded until resolved |
| `ARCHIVED` | Manual archive, upheld contest, or expired validity |
| `REJECTED` | Declined commitment, not stored |

#### Relationship Types

| Type | Effect on Approval |
|------|-------------------|
| `CONTRADICTS` | Both memories enter CONTESTED |
| `REPLACES` | Target loses 2,000 confidence bps |
| `DEPRECATES` | Target loses 1,000 confidence bps |
| `SUPERSEDES` | Target transitions to ARCHIVED |

#### Contest States

| State | Description |
|-------|-------------|
| `PENDING` | Active dispute, awaiting resolution |
| `UPHELD` | Memory archived, stake refunded plus treasury reward |
| `REJECTED` | Contest dismissed, stake burned |

#### Messages

| Message | Description |
|---------|-------------|
| `MsgSubmitCommitment` | Submit memory commitment with keywords, consumes bandwidth |
| `MsgApproveMemory` | Leader/moderator approval with encrypted blob |
| `MsgRejectMemory` | Reject pending commitment without payout |
| `MsgPurgeExpired` | Remove stale pending entries beyond retention window |
| `MsgRelateMemories` | Create relationship between approved memories |
| `MsgApproveRelationship` | Activate relationship effect |
| `MsgContestMemory` | Stake-backed challenge moving memory to CONTESTED |
| `MsgResolveContest` | Leader adjudication (uphold or reject) |
| `MsgSetValidityBounds` | Set validity windows and scope tags |
| `MsgArchiveMemory` | Manual archival of memory |
| `MsgUpdateParams` | Governance parameter update |

#### Query Endpoints

| Endpoint | Path |
|----------|------|
| `GetMemory` | `/wevibe/memory/v1/memory/{org_id}/{content_hash}` |
| `GetPendingCommitments` | `/wevibe/memory/v1/pending/{org_id}` |
| `GetMemoryCount` | `/wevibe/memory/v1/count/{org_id}` |
| `GetEpochMerkleRoot` | `/wevibe/memory/v1/merkle_root/{org_id}/{epoch}` |
| `ListRelationships` | `/wevibe/memory/v1/relationships/{org_id}/{cid}` |
| `GetValidity` | `/wevibe/memory/v1/validity/{org_id}/{cid}` |
| `ListContests` | `/wevibe/memory/v1/contests/{org_id}/{cid}` |
| `GetContest` | `/wevibe/memory/v1/contest/{org_id}/{contest_id}` |
| `ListContestSettlements` | `/wevibe/memory/v1/settlements/{org_id}/{memory_cid}` |
| `GetContestSettlement` | `/wevibe/memory/v1/settlements/{org_id}/contests/{contest_id}` |
| `Params` | `/wevibe/memory/v1/params` |

#### Keeper Dependencies

- `OrgKeeper` — Membership, leadership, config
- `BandwidthKeeper` — Consumption accounting
- `BankKeeper` — Contest escrow
- `ServeKeeper` — Serve counts for confidence decay (injected via setter)

---

### Serve Module (`x/serve`)

The serve module handles serve attestations and deduplication.

#### State Objects

| Prefix | Type | Description |
|--------|------|-------------|
| `nullifier/{hash}` | marker | Set membership guard for deduplication |
| `attestation/{org}/{epoch}/{nullifier}` | `StoredServeAttestation` | Memory hash, serve key, contributor, epoch, nullifier, self-serve flag, model_id, turn_count |
| `stats/{org}/{epoch}` | `StoredEpochServeStats` | Totals, unique memories, unique serve keys, self-serve counts, model_breakdown |
| `contributor/{id}/{epoch}` | `StoredContributorEpochServes` | Per-contributor serve counts, self-serve counts, org coverage, total_turns |
| `memcount/{org}/{hash}/{epoch}` | counter | Per-memory serve count per epoch |
| `memfirst/{org}/{hash}/{epoch}` | marker | First-serve marker for uniqueness |
| `keyfirst/{org}/{key}/{epoch}` | marker | First-serve-key marker for uniqueness |
| `params` | `Params` | Batch size, self-serve discount, per-memory caps |

#### Messages

| Message | Description |
|---------|-------------|
| `MsgSubmitServeBatch` | Submit batch of serve attestations for org and epoch |
| `MsgUpdateParams` | Governance parameter update |

#### Query Endpoints

| Endpoint | Path |
|----------|------|
| `GetEpochServeStats` | `/wevibe/serve/v1/stats/{org_id}/{epoch}` |
| `GetContributorServes` | `/wevibe/serve/v1/contributor/{contributor_id}/{epoch}` |
| `GetMemoryServeCount` | `/wevibe/serve/v1/memory/{org_id}/{content_hash}/{epoch}` |
| `Params` | `/wevibe/serve/v1/params` |

#### Keeper Dependencies

- `OrgKeeper` — Org existence validation
- `MemoryKeeper` — Approved memory existence
- `BandwidthKeeper` — Serve throttling
- `ReputationKeeper` — Serve XP credit

---

### Attestation Module (`x/attestation`)

On-chain anchor for session-level attestation data. Stores session attestation records after a coding session completes. Verification logic is currently stubbed.

#### State Objects

| Prefix | Type | Description |
|--------|------|-------------|
| `attestation/{org}/{hash}` | `StoredSessionAttestation` | Primary key: org + session hash |
| `session_epoch/{org}/{epoch}/{hash}` | `StoredSessionAttestation` | Secondary index for epoch-scoped listing |
| `params` | `Params` | max_attestations_per_epoch, require_attestation_for_serve |

#### Messages

| Message | Description |
|---------|-------------|
| `MsgSubmitSessionAttestation` | Submit a single session attestation |
| `MsgUpdateParams` | Governance parameter update |

#### Query Endpoints

| Endpoint | Path |
|----------|------|
| `GetSessionAttestation` | `/wevibe/attestation/v1/session/{org_id}/{session_hash}` |
| `ListSessionAttestations` | `/wevibe/attestation/v1/sessions/{org_id}/{epoch}` |
| `Params` | `/wevibe/attestation/v1/params` |

#### Keeper Dependencies

- `OrgKeeper` — Org existence validation

#### Notable Implementation

`VerifyCommitLLMReceipt` and `VerifyCloudProviderSignature` return `(false, "unverified: ...")` until CommitLLM receipt formats freeze and cloud providers offer receipt APIs. The stored data model already holds both hashes, so only keeper stubs need replacing later.

---

### Bandwidth Module (`x/bandwidth`)

The bandwidth module provides per-org rate limiting.

#### State Objects

| Prefix | Type | Description |
|--------|------|-------------|
| `state/{org}/{epoch}` | `StoredBandwidthState` | Usage and caps for memory submissions and serves |
| `override/{org}` | `StoredBandwidthOverride` | Org-specific custom caps |
| `params` | `Params` | Default per-epoch caps |

#### Messages

| Message | Description |
|---------|-------------|
| `MsgSetBandwidthOverride` | Set org-specific memory and serve caps |
| `MsgUpdateParams` | Governance parameter update |

#### Query Endpoints

| Endpoint | Path |
|----------|------|
| `GetBandwidthState` | `/wevibe/bandwidth/v1/state/{org_id}/{epoch}` |
| `GetBandwidthOverride` | `/wevibe/bandwidth/v1/override/{org_id}` |
| `GetRemainingBandwidth` | `/wevibe/bandwidth/v1/remaining/{org_id}/{epoch}` |
| `Params` | `/wevibe/bandwidth/v1/params` |

#### Keeper Dependencies

- `OrgKeeper` — Org existence and leadership validation

---

### Reputation Module (`x/reputation`)

The reputation module tracks contributor stats and XP.

#### State Objects

| Prefix | Type | Description |
|--------|------|-------------|
| `active/` | flag | Module activation state |
| `stats/{developer_bytes}` | `StoredReputationStats` | Memory counts, difficulty, domain tags, XP, serve counts, org breadth |
| `memory/{developer_bytes}/{cid}` | `StoredAttestedMemory` | Per-memory provenance |
| `orgset/{developer_bytes}` | `StoredContributorOrgSet` | Org coverage |
| `params` | `Params` | Activation and XP weights |

#### Messages

| Message | Description |
|---------|-------------|
| `MsgUpdateReputation` | Ingest attested memory for contributor |
| `MsgUpdateParams` | Governance parameter update |

#### Serve Integration

`RecordServe` is invoked by the serve keeper to:
- Increment serve counts and self-serve counts
- Award serve XP
- Set first seen epoch if unset
- Expand org breadth tracking

#### Query Endpoints

| Endpoint | Path |
|----------|------|
| `GetReputation` | `/wevibe/reputation/v1/reputation/{developer}` |
| `GetXP` | `/wevibe/reputation/v1/xp/{developer}` |
| `IsActive` | `/wevibe/reputation/v1/active` |
| `GetServeStats` | `/wevibe/reputation/v1/serve_stats/{developer}` |
| `GetContributorOrgSet` | `/wevibe/reputation/v1/org_set/{developer}` |
| `GetCrossOrgProfile` | `/wevibe/reputation/v1/profile/{developer}` |
| `Params` | `/wevibe/reputation/v1/params` |

#### Keeper Dependencies

None (self-contained; invoked by x/serve)

---

### Emissions Module (`x/emissions`)

The emissions module handles daily minting and epoch payout distribution.

#### State Objects

| Prefix | Type | Description |
|--------|------|-------------|
| `pool/` | `StoredEmissionPool` | Total supply, daily mint, operator/validator shares, epoch |
| `emission/{epoch}` | `StoredDailyEmission` | Minted totals plus reward maps |
| `opreward/{id}` | ledger | Service reward recipients |
| `valreward/{id}` | ledger | Validator rewards |
| `workscore/{id}/{org}/{epoch}` | `StoredWorkScore` | Rarity, availability, retrieval volume, scores |
| `gate/{id}/{org}/{epoch}` | `StoredAsymmetricGate` | Retrieval gating based on storage tests |
| `bootstrap/{id}` | credit | Early adopter allocations |
| `bootstrappool/` | pool | Remaining bootstrap credits |
| `params` | `Params` | Mint rate, share percentages, weights, bootstrap duration |

#### Messages

| Message | Description |
|---------|-------------|
| `MsgMintDailyEmission` | Authority minting of daily emission snapshot |
| `MsgDistributeOperatorRewards` | Record operator reward allocations |
| `MsgUpdateParams` | Governance parameter update |

#### Query Endpoints

| Endpoint | Path |
|----------|------|
| `GetEmissionPool` | `/wevibe/emissions/v1/pool` |
| `GetWorkScore` | `/wevibe/emissions/v1/work_score/{operator_id}/{org_id}/{epoch}` |
| `GetOperatorReward` | `/wevibe/emissions/v1/operator_reward/{operator_id}` |
| `Params` | `/wevibe/emissions/v1/params` |

#### Keeper Dependencies

- `ServeKeeper` — Serve attestation lookup
- `OrgKeeper` — Treasury balance, config, rep tiers

---

## Keeper Dependency Graph

Instantiation order (from `app/app.go`):

1. `OrgKeeper` (depends on `BankKeeper`)
2. `BandwidthKeeper` (depends on `OrgKeeper`)
3. `ReputationKeeper` (no external deps)
4. `MemoryKeeper` (depends on `OrgKeeper`, `BankKeeper`, `BandwidthKeeper`; `ServeKeeper` injected via setter)
5. `ServeKeeper` (depends on `OrgKeeper`, `MemoryKeeper`, `BandwidthKeeper`, `ReputationKeeper`)
6. `EmissionsKeeper` (depends on `ServeKeeper`, `OrgKeeper`)

This order ensures each keeper's dependencies are constructed before injection.

---

## Cross-Module Data Flow

### Organization Registration

1. Leader submits `MsgRegisterOrg` via Dashboard
2. Org keeper validates, checks dynamic price, burns `vibe` via `BankKeeper`
3. Leader inserted as member with `leader` role
4. Dynamic price counter increments
5. Hub queries `/wevibe/org/v1/org/{org_id}` and broadcasts to clients

### Memory Submission and Approval

1. Contributor calls `MsgSubmitCommitment` with content hash, keywords, contributor ID
2. Memory keeper checks org exists, consumes bandwidth, enforces pending queue limits
3. Reviewer calls `MsgApproveMemory`, verifying leadership via `OrgKeeper.IsLeader`
4. Blob size validated, record promoted to approved, counter incremented
5. Hub updates dashboards and triggers proof packaging

### Serve Attestation Ingestion

1. Serve nodes batch attestations in `MsgSubmitServeBatch`
2. Serve keeper enforces batch size, consumes bandwidth, rejects nullifier duplicates
3. Validates memories approved, determines self-serve flag
4. For each accepted entry: writes attestation, updates stats, increments counts
5. Forwards to `ReputationKeeper.RecordServe` for XP and org breadth
6. Hub subscribes to stats endpoints for real-time metrics

### Epoch Payout Processing

1. `EmissionsKeeper.AfterEpochEnd` triggers after `wevibe_epoch` ends
2. Mints daily emission, updates pool
3. Iterates orgs with `serve_attestation_required=true` and positive treasury
4. Aggregates contributor serves, fetches rep tiers
5. Debits treasury per contributor based on tier payout rates
6. Loop breaks when treasury insufficient

### Epoch Merkle Root Computation

1. `MemoryKeeper.AfterEpochEnd` runs after emissions hook
2. Enumerates approved memories for the epoch
3. Computes Merkle root via `types.ComputeMerkleRoot`
4. Stores result under `merkle/{org}/{epoch}`
5. Hub monitors for data availability proofs

---

## Module Account Permissions

| Module | Permissions | Purpose |
|--------|-------------|---------|
| `org` | `Burner` | Burn WEVIBE on org registration |
| `memory` | `Burner` | Burn contest escrow on rejected contests |

Standard Cosmos module permissions (staking, mint, distribution, gov, fee_collector) follow Cosmos SDK defaults.

---

## State Machine Transitions

### Organizations

1. **Registration**: Inserts `StoredOrg`, initializes treasury (zero), dynamic price counters, default config
2. **Membership**: `MsgAddMember`/`MsgRemoveMember` update roster
3. **Treasury**: Grows via `MsgFundTreasury`, shrinks via withdrawals or epoch payouts
4. **Config**: Leaders adjust storage quotas, retrieval budgets, payout tiers

### Memories

1. **Pending**: Commitments enter after bandwidth throttling
2. **Approved**: Approvals promote to approved, capturing blobs and approver
3. **Decay**: Epoch hook adjusts confidence based on serves
4. **Expiry**: Validity windows trigger automatic archival
5. **Contest**: Stake-backed disputes move to CONTESTED

### Serves

1. **Batch ingestion**: Accepted or rejected (duplicate, not found, cap exceeded)
2. **Statistics**: Counts, uniqueness trackers, per-contributor tallies
3. **Reputation**: Cascades to reputation keeper for XP and org breadth
4. **Nullifiers**: Guarantees idempotency

### Contributor Reputation

1. **Activation**: Governance toggles via `Params.active`
2. **Memory attestation**: `MsgUpdateReputation` inserts provenance data
3. **Serve XP**: `RecordServe` adds XP per params
4. **Domain expertise**: Tags and difficulty buckets accumulate

### Emission Pool

1. **Initialization**: Seeds total supply, daily mint, share ratios
2. **Daily minting**: Increments supply, records epoch entry
3. **Reward assignment**: Compatibility ledgers for future distribution
4. **Work scores**: Captures rarity, availability, retrieval metrics

---

## Network Topology

### Validators and Full Nodes

Cosmos validators run WeVibe Chain binaries, expose P2P ports, optionally enable gRPC. Full nodes mirror without signing. Both run gRPC gateway for REST access.

### Hub Instances

Hub processes subscribe to Tendermint WebSocket events, call module query endpoints, project to durable caches and ephemeral push streams. Hub brokers notifications to Dashboard.

### Dashboard Clients

Browser or native clients connect to Hub WebSockets for streaming metrics while issuing direct gRPC calls for strongly consistent reads. Transaction submission flows through wallet connectors.

### Data Propagation

Keepers emit structured logs. Hub enriches and forwards to analytics. Ecosystem partners can replay via gRPC/REST backed by protobuf definitions.

### Security Boundaries

- All sensitive transitions on-chain
- Hub and Dashboard never sign transactions
- Governance controls critical parameters via `gov` module address
