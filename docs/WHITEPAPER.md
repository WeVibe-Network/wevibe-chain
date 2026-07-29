# WeVibe Chain Whitepaper

**Version: 2.0**
**Status: Production**

## Abstract

WeVibe Chain is the sovereign Cosmos SDK application anchoring WeVibe Network's encrypted organizational memory system. It stores encrypted approved memories, append-only recall-pivot events, policy-version hash anchors, semantic memory relationships, validity metadata with automatic expiry, stake-backed economic dispute resolution, and epoch-driven emissions. Recall standing is computed at the edge as `f(events, policy_version)`; consensus stores facts and anchors, never verdicts or derived scores.

---

## 1. Protocol Goals

| Goal | Mechanism | Anti-Gaming |
|------|-----------|-------------|
| **Encrypted availability** | On-chain ciphertext storage with per-epoch Merkle roots for data availability proofs | blob size limits prevent storage griefing |
| **Retrieval provenance** | Consumer-signed event log plus anchored edge-policy versions | immutable content-free events; no on-chain verdicts or derived scores |
| **Authoritative curation** | Leader/moderator approvals, relationship effects, manual archival | leader-only rejections prevent arbitrary removal |
| **Dispute resolution** | Stake-backed contests with escrow/burn mechanics | contest stake prevents frivolous challenges |
| **Composable economics** | Bandwidth throttling, reputation tracking, epoch payouts via emissions | treasury sufficiency checks prevent negative balances |

---

## 2. Network Architecture

### 2.1 Three-Layer Design

```
┌─────────────────────────────────────────────────────────────┐
│                     WeVibe Network                             │
├─────────────────────────────────────────────────────────────┤
│  Consensus Layer: WeVibe Chain (Cosmos SDK + CometBFT)       │
│  - Sovereign state machine, all critical transitions         │
│  - bech32 addresses with 'wevibe' prefix                     │
│  - Fees denominated in 'uvibe' (1 VIBE = 1,000,000 uvibe)  │
├─────────────────────────────────────────────────────────────┤
│  Indexing Layer: WeVibe Hub (off-chain Go service)           │
│  - WebSocket subscriptions to new blocks                   │
│  - Materializes module-specific views                       │
│  - Pushes normalized payloads to Dashboard/analytics       │
│  - Pure observer: no keeper logic runs outside chain       │
├─────────────────────────────────────────────────────────────┤
│  Interface Layer: WeVibe Dashboard (Next.js)                   │
│  - Consumes Hub WebSocket streams for live metrics         │
│  - Direct gRPC calls for authoritative reads               │
│  - Transaction assembly via wallet connectors              │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 Organization Registration Economics

Organizations register by burning VIBE via a **dynamic pricing curve**:

```
Price = base_burn_price × (1 + burn_price_increase_percent / 100) ^ creation_count
```

- `base_burn_price` = 1,000,000 uvibe
- `burn_price_increase_percent` = 10% per org
- Compounding resets after `burn_price_decay_epochs` (default: 30)

Example: After 5 organizations, price = 1,000,000 × 1.10⁵ = 1,610,510 uvibe

This creates economic friction against spam registrations while remaining affordable for legitimate orgs. The decay prevents permanent price inflation as the ecosystem matures.

### 2.3 Treasury Model

Each organization maintains an on-chain treasury (balance stored as decimal string for arbitrary precision). Treasuries are:

- **Funded** via `MsgFundTreasury` by administrators
- **Debited** via epoch payout distributions (atomic, prevents race conditions)
- **Withdrawable** by leaders via `MsgWithdrawTreasury`

Treasury balances are mirrored as `StoredTreasury` in the `treasury/{org}` KV prefix.

---

## 3. Memory Module: Lifecycle & Recall Provenance

### 3.1 Memory States

```
                    Approval
PENDING ──────────► APPROVED
                       │
                       │ Confidence ≥ stable_threshold (8000 bps)
                       ▼
                    STABLE
                       │
                       │ Confidence < stable_threshold
                       ▼
                    DEGRADED
                       │
                       │ Confidence < dormant_threshold (2000 bps)
                       ▼
                    DORMANT
                       │
                       │ Expiry / Archive / Contest Upheld
                       ▼
                    ARCHIVED
```

**State Descriptions:**

| State | Entry Condition | Recallable | Notes |
|-------|----------------|-----------|-------|
| `PENDING` | `MsgSubmitCommitment` | No | Awaiting leader approval |
| `APPROVED` | `MsgApproveMemory` | Yes | Default post-review state |
| `STABLE` | confidence ≥ 8000 bps | Yes | Preferred ordering |
| `DEGRADED` | 5000 ≤ confidence < 8000 | Yes | Low priority |
| `DORMANT` | 2000 ≤ confidence < 5000 | Yes | Hidden from default recall |
| `CONTESTED` | `MsgContestMemory` | No | Active dispute |
| `ARCHIVED` | Manual, upheld contest, or expiry | No | Not recalled |
| `REJECTED` | `MsgRejectMemory` | No | Not stored |

### 3.2 Recall Provenance Mechanics

The recall pivot moved standing out of consensus. The chain stores immutable, append-only, content-free, consumer-signed events and governance-anchored edge-policy hashes. Edge consumers compute standing as `f(events, policy_version)`.

**Boundary rule:** events are never verdicts and must not carry weights, standing, scores, trust values, archive flags, plaintext, ciphertext, or other content.

**Policy anchoring:** `StoredPolicyAnchor` records the policy version and hash. The hash proves which published edge policy a consumer used; it does not itself rank or hide memories.

### 3.3 Per-Memory Serve Cap

`max_serves_per_memory_per_epoch` (default: 10) limits repeat serving of the same content. This prevents:
- A single memory from dominating emissions if heavily self-served
- Artificial confidence inflation from coordinated self-serving

---

## 4. Memory Relationships

### 4.1 Relationship Types

`MsgRelateMemories` + `MsgApproveRelationship` creates edges between approved memories:

| Type | Effect on Target |
|------|-----------------|
| `CONTRADICTS` | Both memories enter `CONTESTED` |
| `REPLACES` | Target loses 2,000 confidence bps |
| `DEPRECATES` | Target loses 1,000 confidence bps |
| `SUPERSEDES` | Target transitions to `ARCHIVED` |

**Why these specific effects?**
- `CONTRADICTS`: Forces org leader adjudication—bad知识 should not silently coexist
- `REPLACES`: Strong signal that new content supersedes old; big confidence penalty
- `DEPRECATES`: Weaker signal; gentle deprecation without archive
- `SUPERSEDES`: Complete replacement; old knowledge is canonically outdated

### 4.2 Validity Bounds

`MsgSetValidityBounds` sets temporal constraints:

```protobuf
valid_after_epoch: Memory becomes active
valid_until_epoch: Memory expires (auto-archived)
scope_tags: Opaque metadata for filtering (serialized bytes)
```

**Use case:** Time-sensitive knowledge (bug reports, deprecated APIs) automatically expires. Scope tags enable filtering by context (e.g., "v1-api", "deprecated-feature").

---

## 5. Contest Mechanics

### 5.1 Contest Flow

1. **Initiation** (`MsgContestMemory`):
   - Contributor stakes `contest_stake_vibe` (org-configurable)
   - Stake escrowed in memory module account (Burner permission)
   - Memory enters `CONTESTED` state

2. **Adjudication** (`MsgResolveContest`):
   - Leader-only decision
   - **Upheld**: Memory → `ARCHIVED`, stake refunded + treasury reward
   - **Rejected**: Stake burned, confidence re-evaluated via lifecycle

3. **Auto-expiry**:
   - `CheckContestExpiry` runs in memory epoch hook
   - Unresolved contests after `contest_window_epochs` (default: 14) auto-rejected

### 5.2 Economic Design

**Why stake escrow?**
- Burns require sending to `0x...dead` address; escrow keeps funds redeemable
- Upheld contests: challenger gets stake back + treasury reward (economic incentive to report bad content)
- Rejected contests: stake burned (economic cost to frivolous challenges)

**Why treasury reward on upheld?**
- Encourages community monitoring of content quality
- Offsets the risk of contesting legitimate content

---

## 6. Serve Receipts

### 6.1 Deduplication: Computed Fingerprints

`ServeEntry` does not carry a client-provided dedup key. The chain computes:

`fingerprint = SHA256(memory_content_hash ‖ serve_key_pubkey ‖ BigEndianUint64(epoch))`

from the serve's own canonical contents, then:

1. Checks `fingerprint/{hex}` prefix existence
2. If exists → reject as duplicate
3. If not exists → mark and accept

`ServeEntry` carries `serve_key_pubkey`, `serve_sig`, and `nonce`. The `nonce` is signature freshness inside the signed canonical body; it is not the dedup key.

**Why computed fingerprints?**
- The chain does not trust a client-supplied dedup value.
- Identical `(memory_content_hash, serve_key_pubkey, epoch)` tuples collide deterministically, making retries idempotent and replay-resistant.
- The dedup identity is self-asserted by the consumer's serve key (D-SERVE-CONSUMER-SIGNED).

Denial receipts follow the same model with their own computed key:
`SHA256(CanonicalDenialBody(org_id, memory_hash, epoch, serve_key_pubkey, serve_fingerprint, nonce=nil))`,
stored under `denial/{org}/{epoch}/{fingerprint}` with `denyfingerprint/{hex}` dedup markers.

### 6.2 Self-Serve Detection

```
is_self_serve = (serve_key == contributor_id)
```

**Design rationale:**
- Self-serving inflates confidence without actual retrieval value
- Reduced XP (default: 0 for self-serve vs 1 for normal) discourages the behavior
- Self-serve counts tracked separately in `StoredContributorEpochServes`

### 6.3 Per-Memory Serve Cap

`max_serves_per_memory_per_epoch` (default: 10) prevents repeat serving. Tracking:

- `memcount/{org}/{hash}/{epoch}`: counter per memory
- `memfirst/{org}/{hash}/{epoch}`: first-serve marker for uniqueness stats

### 6.4 Bandwidth Consumption

Serve receipts consume bandwidth under `state/{org}/{epoch}`:
- `serve_used`: incremented per accepted serve
- `serve_cap`: per-epoch limit (default: 10,000)
- Leaders can override via `MsgSetBandwidthOverride`

### 6.5 Stats Tracking

`StoredEpochServeStats` aggregates:
- `total_serves`: all serves
- `unique_memories_served`: via `memfirst` markers
- `unique_serve_keys`: via `keyfirst` markers
- `self_serves`: self-serve count
- `model_breakdown`: serve count per model_id (for analytics)

---

## 7. Bandwidth Module

### 7.1 Rate Limiting Design

Per-epoch caps prevent unbounded submission rates:

| Type | Default Cap | Overrideable |
|------|-------------|--------------|
| Memory submissions | 1,000/epoch | Yes (leader) |
| Serve receipts | 10,000/epoch | Yes (leader) |

### 7.2 Override Mechanism

Leaders set org-specific caps via `MsgSetBandwidthOverride`:
- Stored in `override/{org}` prefix
- Takes precedence over module `Params` defaults

### 7.3 Epoch Reset

Bandwidth state is epoch-scoped (`state/{org}/{epoch}`). At epoch boundary:
- New epoch starts with empty counters
- Automatic rollover (no explicit reset message required)

---

## 8. Reputation Module

### 8.1 XP System

```
serve_xp_per_serve: 1 (normal serves)
self_serve_xp_per_serve: 0 (self-serves)
```

XP accumulates in `StoredReputationStats.xp` and enables:
- Payout tier matching in emissions
- Contributor leadershipboards
- Cross-org quality signals

### 8.2 Cross-Org Profiles

`StoredContributorOrgSet` tracks which orgs a contributor has served:
- Enables "verified contributor across N orgs" metrics
- Hub uses this for contributor discovery

### 8.3 Domain Expertise

`MsgUpdateReputation` accepts:
- `domain_tags`: e.g., ["kubernetes", "networking"]
- `difficulty`: 1-10 scale (capped at `max_difficulty`)
- `quality`: 1-10 scale (capped at `max_quality`)

These accumulate in `StoredAttestedMemory` for provenance tracking.

### 8.4 Activation Gating

`Params.active` (default: true) gates all reputation updates:
- When false, `MsgUpdateReputation` fails with `ErrReputationNotActive`
- Allows governance to pause reputation accrual if needed

---

## 9. Emissions Module

### 9.1 Daily Minting

```
MsgMintDailyEmission (authority):
1. Mint daily_mint_amount (default: 1,000,000 uvibe)
2. Update pool: total_supply += daily_mint
3. Record StoredDailyEmission for epoch
4. Split: operator_share (10%), validator_share (5%)
```

### 9.2 Epoch Payout Processing

After minting, `AfterEpochEnd` iterates orgs with `serve_receipt_required=true`:

```
For each org with positive treasury:
  1. Load serve receipts for epoch
  2. Aggregate serves per contributor
  3. Fetch org rep tiers (ordered by reputation bounds)
  4. For each contributor:
     - Match against tiers (first tier where min <= rep <= max)
     - Calculate payout: min(contributor_serves, tier.max_contributions) × tier.payout_per_serve
     - Debit org treasury
  5. Break loop if treasury becomes insufficient
```

**Why tier-based?**
- Rewards high-quality contributors more
- `max_contributions_per_epoch` prevents any single contributor from dominating

### 9.3 Work Score Calculation

```
storage_score = availability × storage_weight_percent / 100
retrieval_score = retrieval_volume × rarity_multiplier × retrieval_weight_percent / 100
total_score = storage_score + retrieval_score
```

**Rarity multiplier:** computed based on how many orgs serve the content vs total orgs. Content served across many orgs has lower multiplier; niche content has higher multiplier.

**Asymmetric gating:** `gate/{id}/{org}/{epoch}` stores retrieval eligibility:
- `storage_passed`: org passed storage test
- `retrieval_allowed`: org cleared retrieval threshold

### 9.4 Bootstrap Credits

Early adopters receive `StoredBootstrapCredit` allocations:
- `bootstrap/{id}`: individual credit balance
- `bootstrappool/`: remaining pool total
- Redemption decrements both contributor balance and shared pool
- Expires after `bootstrap_duration_epochs` (default: 180)

---

## 10. Attestation Module

### 10.1 Session Attestations

Stores metadata about coding sessions (not the actual transcript):

```protobuf
session_hash: SHA-256 of transcript (32 bytes)
model_id: e.g., "qwen3:4b", "claude-sonnet-4-20250514"
turn_count: conversation turns
token_count: total tokens
provider_type: LOCAL (CommitLLM) or CLOUD (API)
commitllm_receipt_hash: SHA-256 of CommitLLM receipt (optional)
provider_signature_hash: SHA-256 of cloud provider signature (optional)
```

### 10.2 Verification Stubs

Currently returns `(false, "unverified: ...")`:
- `VerifyCommitLLMReceipt`: waits for CommitLLM receipt format stabilization
- `VerifyCloudProviderSignature`: waits for cloud provider receipt APIs

**Design:** Data model is already correct; only keeper verification stubs need replacement when integrations mature.

### 10.3 Provider Types

| Type | Description | Verification Path |
|------|-------------|-------------------|
| `LOCAL` | Open-weight, self-hosted (Ollama, vLLM, llama.cpp) | CommitLLM receipt |
| `CLOUD` | Closed-weight API (Claude, GPT, Gemini) | Provider signature |

---

## 11. Epoch System

### 11.1 WeVibe Epoch Configuration

```json
{
  "identifier": "wevibe_epoch",
  "duration": "60s",       // dev default
  "86400s":                 // production
}
```

### 11.2 Hook Execution Order

```
┌─────────────────────────────────────────────────────────────┐
│ AfterEpochEnd Hooks (registered with EpochsKeeper)           │
├─────────────────────────────────────────────────────────────┤
│ 1. EmissionsKeeper.AfterEpochEnd                            │
│    - MintDailyEmission                                      │
│    - ProcessOrgPayouts (debit treasuries)                   │
├─────────────────────────────────────────────────────────────┤
│ 2. MemoryKeeper.AfterEpochEnd                               │
│    - Set current epoch                                      │
│    - CheckEpochExpiry (validity-window expiry)              │
│    - ComputeEpochMerkleRoot                                 │
└─────────────────────────────────────────────────────────────┘
```

**Why emissions before memory?**
- Payout debits happen before memory epoch maintenance
- Treasury debits are atomic with minting
- Merkle root reflects the approved-memory set after epoch maintenance

### 11.3 Merkle Root Computation

`ComputeMerkleRoot` over approved memories per org per epoch:
- Collates `content_hash` from `approved/{org}/...`
- Computes Merkle tree root
- Stored in `merkle/{org}/{epoch}`

**Use case:** Data availability proofs for retrieval partners. Hub monitors these for compliance reporting.

---

## 12. Cross-Module Integration

### 12.1 Keeper Dependency Graph

```
BankKeeper
    │
    └─► OrgKeeper ─────────────────────────┐
          │                                 │
          ├─► BandwidthKeeper               │
          │     │                           │
          │     └─► MemoryKeeper            │
          │           │                     │
          │           ├─► ServeKeeper ───────┤
          │           │     │               │
          │           │     ├─► ReputationKeeper
          │           │     │
          │           │     └─► BandwidthKeeper
          │           │
          │           └─► (ServeKeeper injected via setter)
          │
          └─► EmissionsKeeper
```

**Dependency rationale:**
- OrgKeeper first ( foundational)
- BandwidthKeeper after Org (org existence)
- MemoryKeeper after Org + Bandwidth + Bank (state changes, escrow)
- ServeKeeper after Org + Memory + Bandwidth + Reputation (serves affect many systems)
- EmissionsKeeper last (needs serves + orgs)

### 12.2 Message Flow: Memory Submission

```
MsgSubmitCommitment
    │
    ├─► OrgKeeper.IsMember           (validate membership)
    ├─► BandwidthKeeper.ConsumeMemoryBandwidth
    │       (reject if cap exceeded)
    ├─► MemoryKeeper.GetPendingCount (reject if max_pending exceeded)
    │
    └─► Store pending/{org}/{hash}
```

### 12.3 Message Flow: Serve Receipt

```
MsgSubmitServeBatch
    │
    ├─► ServeKeeper.ProcessServeBatch
    │       │
    │       ├─► BandwidthKeeper.ConsumeServeBandwidth
    │       │
    │       ├─► Check fingerprint/{hex} (reject duplicates)
    │       │
    │       ├─► For each ServeEntry:
    │       │   ├─► MemoryKeeper.IsApproved    (validate memory exists)
    │       │   ├─► Check memcount cap        (reject if max exceeded)
    │       │   ├─► Determine is_self_serve
    │       │   ├─► Store attestation
    │       │   ├─► Update stats
    │       │   └─► ReputationKeeper.RecordServe
    │       │           (XP, org breadth)
    │       │
    │       └─► Return acceptance counts
```

### 12.4 Message Flow: Epoch Payout

```
AfterEpochEnd (Emissions)
    │
    ├─► MintDailyEmission
    │
    └─► For each org:
            ├─► OrgKeeper.GetOrgConfig
            │       (skip if !serve_receipt_required)
            ├─► OrgKeeper.GetTreasury
            │       (skip if balance == 0)
            ├─► ServeKeeper.GetServeAttestationsForEpoch
            │       (aggregate per contributor)
            ├─► OrgKeeper.GetRepTiers
            │       (tier matching)
            └─► OrgKeeper.DebitTreasury (atomic per payout)
                    (break if insufficient)
```

---

## 13. Anti-Gaming Measures

| Measure | Implementation | Purpose |
|---------|----------------|---------|
| **Fingerprint deduplication** | `fingerprint/{hex}` prefix check | Prevents replay attacks |
| **Per-memory serve cap** | `memcount/{org}/{hash}/{epoch}` counter | Limits repeat serving |
| **Self-serve discount** | `is_self_serve = (serve_key == contributor_id)` | Reduces self-serving incentive |
| **Bandwidth caps** | `state/{org}/{epoch}` usage tracking | Rate limits submissions/serves |
| **Leader-only approvals** | `OrgKeeper.IsLeader` check | Prevents unauthorized approvals |
| **Treasury sufficiency** | Balance check before payout debit | Prevents negative balances |
| **Contest stake escrow** | Memory module account with Burner | Economic cost for frivolous contests |
| **Pending queue limits** | `max_pending_per_org` check | Prevents hoarding |
| **Blob size limits** | `max_blob_size_bytes` validation | Prevents storage griefing |
| **Minimum org age** | `min_org_age_epochs` check | Prevents spam from new orgs |
| **XP devaluation for self-serve** | `self_serve_xp_per_serve = 0` | Removes incentive |

---

## 14. Module Summary

### 14.1 State Prefixes

| Module | Prefix | State Object |
|--------|---------|--------------|
| `x/org` | `org/` | `StoredOrg` |
| | `member/{org}/{pubkey}` | `StoredMemberRecord` |
| | `dynprice/` | `StoredDynamicPrice` |
| | `treasury/{org}` | `StoredTreasury` |
| | `reptier/{org}` | `StoredRepTierConfig` |
| | `orgconfig/{org}` | `StoredOrgConfig` |
| `x/memory` | `pending/{org}/...` | `StoredPendingCommitment` |
| | `approved/{org}/...` | `StoredMemoryCommitment` |
| | `relationship/{org}:{source}:{target}` | `StoredMemoryRelationship` |
| | `contest/{org}:{contestID}` | `StoredMemoryContest` |
| | `validity/{org}:{cid}` | `StoredValidityMetadata` |
| | `merkle/{org}/{epoch}` | `StoredEpochMerkleRoot` |
| | `count/{org}` | counter |
| `x/serve` | `fingerprint/{hash}` | marker |
| | `receipt/{org}/{epoch}/{fingerprint}` | `StoredServeReceipt` |
| | `stats/{org}/{epoch}` | `StoredEpochServeStats` |
| | `contributor/{id}/{epoch}` | `StoredContributorEpochServes` |
| | `memcount/{org}/{hash}/{epoch}` | counter |
| `x/attestation` | `attestation/{org}/{hash}` | `StoredSessionAttestation` |
| | `session_epoch/{org}/{epoch}/{hash}` | `StoredSessionAttestation` (index) |
| `x/bandwidth` | `state/{org}/{epoch}` | `StoredBandwidthState` |
| | `override/{org}` | `StoredBandwidthOverride` |
| `x/reputation` | `active/` | flag |
| | `stats/{developer_bytes}` | `StoredReputationStats` |
| | `memory/{developer_bytes}/{cid}` | `StoredAttestedMemory` |
| | `orgset/{developer_bytes}` | `StoredContributorOrgSet` |
| `x/emissions` | `pool/` | `StoredEmissionPool` |
| | `emission/{epoch}` | `StoredDailyEmission` |
| | `opreward/{id}` | ledger |
| | `valreward/{id}` | ledger |
| | `workscore/{id}/{org}/{epoch}` | `StoredWorkScore` |
| | `gate/{id}/{org}/{epoch}` | `StoredAsymmetricGate` |
| | `bootstrap/{id}` | `StoredBootstrapCredit` |
| | `bootstrappool/` | pool |

### 14.2 Key Messages

| Module | Message | Purpose |
|--------|---------|---------|
| `x/org` | `MsgRegisterOrg` | Burn VIBE, create org, auto-add leader |
| | `MsgAddMember` / `MsgRemoveMember` | Roster management |
| | `MsgFundTreasury` / `MsgWithdrawTreasury` | Treasury operations |
| | `MsgSetRepTiers` | Configure payout tiers |
| | `MsgGrantTrialAllowance` | Fee grant for trial submissions |
| `x/memory` | `MsgSubmitCommitment` | Submit memory commitment |
| | `MsgApproveMemory` | Leader/moderator approval |
| | `MsgRejectMemory` | Reject without payout |
| | `MsgRelateMemories` / `MsgApproveRelationship` | Create relationship edges |
| | `MsgContestMemory` / `MsgResolveContest` | Dispute resolution |
| | `MsgSetValidityBounds` | Set temporal constraints |
| `x/serve` | `MsgSubmitServeBatch` | Batch serve receipts |
| `x/attestation` | `MsgSubmitSessionAttestation` | Submit session metadata |
| `x/bandwidth` | `MsgSetBandwidthOverride` | Leader-configured caps |
| `x/reputation` | `MsgUpdateReputation` | Attest memory for contributor |
| `x/emissions` | `MsgMintDailyEmission` | Authority minting |
| | `MsgDistributeOperatorRewards` | Record reward allocations |

---

## 15. Technical Parameters

### 15.1 Org Module

| Parameter | Default | Description |
|-----------|---------|-------------|
| `min_registration_fee` | 1,000,000 uvibe | Minimum burn |
| `annual_renewal_fee` | 100,000 uvibe | Renewal cost |
| `default_storage_quota` | 1,000,000 | Storage units |
| `default_retrieval_budget` | 500,000 | Retrieval budget |
| `grace_period_epochs` | 365 | Inactivity grace period |
| `burn_price_decay_epochs` | 30 | Dynamic price reset |
| `base_burn_price` | 1,000,000 uvibe | Starting price |
| `burn_price_increase_percent` | 10 | Per-org compounding |

### 15.2 Memory Module

| Parameter | Default | Description |
|-----------|---------|-------------|
| `max_pending_per_org` | 1,000 | Pending queue limit |
| `pending_retention_epochs` | 30 | Auto-purge window |
| `max_blob_size_bytes` | 1,048,576 | 1 MiB blob limit |
| `max_keywords_per_memory` | 10 | Keyword cap |
| `contest_window_epochs` | 14 | Auto-reject contests |

### 15.3 Serve Module

| Parameter | Default | Description |
|-----------|---------|-------------|
| `max_serves_per_batch` | 100 | Batch size limit |
| `self_serve_discount_percent` | 50 | Self-serve penalty |
| `max_serves_per_memory_per_epoch` | 10 | Per-memory cap |
| `min_org_age_epochs` | 7 | Minimum org age |
| `diminishing_returns_threshold` | 100 | Diminishing returns start |

### 15.4 Bandwidth Module

| Parameter | Default | Description |
|-----------|---------|-------------|
| `default_memory_cap_per_epoch` | 1,000 | Memory submissions |
| `default_serve_cap_per_epoch` | 10,000 | Serve receipts |

### 15.5 Reputation Module

| Parameter | Default | Description |
|-----------|---------|-------------|
| `active` | true | Module activation |
| `max_difficulty` | 10 | Difficulty cap |
| `max_quality` | 10 | Quality cap |
| `serve_xp_per_serve` | 1 | Normal serve XP |
| `self_serve_xp_per_serve` | 0 | Self-serve XP |

### 15.6 Emissions Module

| Parameter | Default | Description |
|-----------|---------|-------------|
| `daily_mint_amount` | 1,000,000 uvibe | Daily mint |
| `operator_share_percent` | 10 | Operator allocation |
| `validator_share_percent` | 5 | Validator allocation |
| `storage_weight_percent` | 40 | Storage score weight |
| `retrieval_weight_percent` | 60 | Retrieval score weight |
| `rarity_multiplier_cap` | "10" | Rarity cap |
| `bootstrap_duration_epochs` | 180 | Bootstrap expiry |

---

## 16. Security Model

### 16.1 On-Chain Trust

- All sensitive state transitions occur on-chain
- Deterministic state machine (Cosmos SDK + CometBFT)
- Validator set secures consensus

### 16.2 Observer Services

- Hub and Dashboard are **pure observers**
- Neither signs transactions
- Only relay prepared payloads and display results

### 16.3 Governance Controls

- Critical parameters gated via `gov` module address
- `MsgUpdateParams` for each module requires governance authority
- Standard Cosmos governance (deposit, voting, quorum, threshold)

---

## 17. Protocol Compatibility

### 17.1 Protobuf Schema Evolution

- New RPCs added without breaking prior schemas
- Default values fill omitted fields
- Wire format backward compatible

### 17.2 Module Account Permissions

```protobuf
// org module: Burner for registration fee burns
// memory module: Burner for contest stake burns
```

### 17.3 Genesis Format

Extended to accept new state objects:
- Old genesis files load with defaults for missing fields
- Export iterates all prefixes faithfully

---

## 18. Design Rationale

### 18.1 Why Cosmos SDK?

- Mature battle-tested infrastructure (staking, auth, bank, gov)
- CometBFT provides deterministic finality
- Module system maps naturally to WeVibe's concerns
- Built-in epochs, feegrant, distribution simplify design

### 18.2 Why Epoch-Driven?

Time-based logic tied to `wevibe_epoch`:
- Memory validity maintenance per epoch
- Emissions mint per-epoch
- Bandwidth resets per-epoch
- Merkle roots per-epoch

This allows the entire system to advance atomically without block-level complexity.

### 18.3 Why Edge-Computed Recall Standing?

Recall quality changes faster than consensus rules. The chain therefore stores signed facts and anchors policy hashes while edge consumers compute standing from those facts. This preserves auditability without baking mutable ranking policy into consensus.

### 18.4 Why Decimal String for Treasury?

Go's `sdk.Int` uses big.Int internally but `String()` loses precision for decimal Display. Storing as string preserves arbitrary precision for financial calculations.

### 18.5 Why Separate Bandwidth Module?

Rate limiting is orthogonal to business logic:
- Memory, serve modules consume bandwidth
- Bandwidth is org-scoped, epoch-scoped
- Override mechanism lets leaders adjust without governance

### 18.6 Why Reputation as Separate Module?

Reputation concerns are distinct from memory/serving:
- Self-contained with no external deps
- Invoked by serve keeper via interface
- Can be toggled independently
- XP accumulation is independent of memory lifecycle

---

## References

- [Architecture Documentation](ARCHITECTURE.md) — System topology, data flow, keeper dependencies
- [Module Reference](MODULES.md) — Detailed protobuf message specs, state objects
- [Topology Documentation](TOPOLOGY.md) — Sprint 23/24 updates, cross-module flows
- [Parameters Reference](PARAMETERS.md) — All module parameters with defaults
- [API Reference](API.md) — gRPC and REST endpoints
- [Deployment Guide](DEPLOYMENT.md) — Production deployment procedures
- [CLI Reference](CLI.md) — Command-line interface documentation
- [PDP Documentation](PDP.md) — Protocol design proposals
