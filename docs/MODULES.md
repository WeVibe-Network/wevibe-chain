# WeVibe Chain Module Reference

This document provides detailed specifications for each module including message types, query endpoints, state objects, parameters, and notable implementation details.

## Table of Contents

- [Org Module](#org-module-xorg)
- [Memory Module](#memory-module-xmemory)
- [Serve Module](#serve-module-xserve)
- [Attestation Module](#attestation-module-xattestation)
- [Bandwidth Module](#bandwidth-module-xbandwidth)
- [Reputation Module](#reputation-module-xreputation)
- [Emissions Module](#emissions-module-xemissions)

---

## Org Module (`x/org`)

Manages organization registration, membership, treasury operations, and payout tier configuration.

### State Objects

#### StoredOrg
```protobuf
message StoredOrg {
  string org_id = 1;
  string leader = 2;
  uint64 created_at = 3;
  uint64 renewal_height = 4;
  uint64 storage_quota = 5;
  uint64 retrieval_budget = 6;
  int32 status = 7;  // 0=active, 1=suspended, 2=closed
  string domain = 8;  // Organization domain (e.g. "backend-infra", "mobile-ios")
}
```

#### StoredMemberRecord
```protobuf
message StoredMemberRecord {
  string org_id = 1;
  string pubkey = 2;
  string role = 3;  // "leader", "moderator", or custom role name
}
```

#### StoredDynamicPrice
```protobuf
message StoredDynamicPrice {
  uint64 price = 1;           // Current burn price in uvibe
  uint64 last_creation = 2;    // Epoch of last org creation
  uint64 creation_count = 3;   // Orgs created in current window
}
```

#### StoredTreasury
```protobuf
message StoredTreasury {
  string org_id = 1;
  string balance = 2;  // Decimal string for arbitrary precision
}
```

#### RepTier
```protobuf
message RepTier {
  uint64 min_reputation = 1;
  uint64 max_reputation = 2;
  uint64 max_contributions_per_epoch = 3;
  string payout_per_serve = 4;  // String-encoded decimal
}
```

#### StoredOrgConfig
```protobuf
message StoredOrgConfig {
  string org_id = 1;
  bool serve_receipt_required = 2;
  uint64 decay_rate_bps = 3;      // Basis points per epoch when no serves
  uint64 contest_stake_vibe = 4;  // Stake amount for memory contests
}
```

### Parameters

```protobuf
message Params {
  uint64 min_registration_fee = 1;
  uint64 annual_renewal_fee = 2;
  uint64 default_storage_quota = 3;
  uint64 default_retrieval_budget = 4;
  uint64 grace_period_epochs = 5;
  uint64 burn_price_decay_epochs = 6;
  uint64 base_burn_price = 7;
  uint64 burn_price_increase_percent = 8;
}
```

| Parameter | Description |
|-----------|-------------|
| `min_registration_fee` | Minimum VIBE burned on registration |
| `annual_renewal_fee` | Annual fee for org renewal |
| `default_storage_quota` | Default storage allocation |
| `default_retrieval_budget` | Default retrieval budget |
| `grace_period_epochs` | Epochs before org expires without renewal |
| `burn_price_decay_epochs` | Epochs until creation count decays |
| `base_burn_price` | Starting price before compounding |
| `burn_price_increase_percent` | Percentage increase per org in window |

### Messages

#### MsgRegisterOrg
Burns VIBE at dynamic price and creates a new organization.

```protobuf
message MsgRegisterOrg {
  string signer = 1;
  string org_id = 2;
  string leader = 3;
  uint64 storage_quota = 4;
  uint64 retrieval_budget = 5;
  string domain = 6;  // Optional domain identifier
}
```

**Fields:**
- `signer`: Account burning tokens (authority)
- `org_id`: Unique identifier for the org
- `leader`: Bech32 address of org leader
- `storage_quota`: Maximum storage allocation
- `retrieval_budget`: Retrieval budget allocation

**Authorization:** Signer must have sufficient VIBE balance.

#### MsgAddMember
Adds a member to an organization with specified role.

```protobuf
message MsgAddMember {
  string signer = 1;
  string org_id = 2;
  string pubkey = 3;
  string role = 4;
}
```

#### MsgRemoveMember
Removes a member from organization's roster.

```protobuf
message MsgRemoveMember {
  string signer = 1;
  string org_id = 2;
  string pubkey = 3;
}
```

#### MsgFundTreasury
Adds funds to organization treasury.

```protobuf
message MsgFundTreasury {
  string signer = 1;
  string org_id = 2;
  string amount = 3;  // String-encoded decimal
}
```

#### MsgWithdrawTreasury
Withdraws funds from organization treasury.

```protobuf
message MsgWithdrawTreasury {
  string signer = 1;
  string org_id = 2;
  string amount = 3;
  string recipient = 4;
}
```

**Authorization:** Signer must be org leader.

#### MsgSetRepTiers
Configures payout tiers for epoch rewards.

```protobuf
message MsgSetRepTiers {
  string signer = 1;
  string org_id = 2;
  repeated RepTier tiers = 3;
}
```

**Note:** Tiers should be ordered from lowest to highest reputation. The lowest tier acts as default for contributors below all tier thresholds.

#### MsgSetOrgConfig
Updates organization configuration.

```protobuf
message MsgSetOrgConfig {
  string signer = 1;
  string org_id = 2;
  bool serve_receipt_required = 3;
  uint64 decay_rate_bps = 4;
  uint64 contest_stake_vibe = 5;
}
```

#### MsgGrantTrialAllowance
Grants fee grant for trial memory submissions.

```protobuf
message MsgGrantTrialAllowance {
  string signer = 1;
  string org_id = 2;
  string grantee = 3;
  uint64 daily_submissions = 4;
  uint64 trial_days = 5;
}
```

### Query Endpoints

| Endpoint | REST Path | Response |
|----------|-----------|----------|
| `GetOrg` | `/wevibe/org/v1/org/{org_id}` | `QueryGetOrgResponse` |
| `GetMembers` | `/wevibe/org/v1/members/{org_id}` | `QueryGetMembersResponse` (repeated `MemberInfo`) |
| `IsMember` | `/wevibe/org/v1/is_member/{org_id}/{pubkey}` | `QueryIsMemberResponse` (bool) |
| `IsModerator` | `/wevibe/org/v1/is_moderator/{org_id}/{pubkey}` | `QueryIsModeratorResponse` (bool) |
| `GetTreasury` | `/wevibe/org/v1/treasury/{org_id}` | `QueryGetTreasuryResponse` (balance string) |
| `GetRepTiers` | `/wevibe/org/v1/rep_tiers/{org_id}` | `QueryGetRepTiersResponse` (repeated `RepTier`) |
| `GetOrgConfig` | `/wevibe/org/v1/config/{org_id}` | `QueryGetOrgConfigResponse` |
| `Params` | `/wevibe/org/v1/params` | `QueryParamsResponse` |

### Notable Implementation Details

- **Dynamic Pricing**: Registration price = `base_burn_price * (1 + burn_price_increase_percent/100)^creation_count`, subject to decay after `burn_price_decay_epochs`
- **Treasury Precision**: Balances stored as strings to avoid floating-point issues
- **Leader Removal**: Leaders cannot be removed unless they transfer leadership first
- **Hook Dependencies**: Org keeper only depends on BankKeeper for burns and transfers

---

## Memory Module (`x/memory`)

Manages organizational memory commitments including lifecycle states, relationships, and dispute resolution.

### State Objects

#### MemoryState Enum
```protobuf
enum MemoryState {
  MEMORY_STATE_UNSPECIFIED = 0;
  MEMORY_STATE_PENDING = 1;
  MEMORY_STATE_APPROVED = 2;
  MEMORY_STATE_STABLE = 3;
  MEMORY_STATE_CONTESTED = 4;
  MEMORY_STATE_DEGRADED = 5;
  MEMORY_STATE_DORMANT = 6;
  MEMORY_STATE_ARCHIVED = 7;
  MEMORY_STATE_REJECTED = 8;
}
```

#### RelationType Enum
```protobuf
enum RelationType {
  RELATION_TYPE_UNSPECIFIED = 0;
  RELATION_TYPE_CONTRADICTS = 1;
  RELATION_TYPE_REPLACES = 2;
  RELATION_TYPE_DEPRECATES = 3;
  RELATION_TYPE_SUPERSEDES = 4;
}
```

#### ContestState Enum
```protobuf
enum ContestState {
  CONTEST_STATE_UNSPECIFIED = 0;
  CONTEST_STATE_PENDING = 1;
  CONTEST_STATE_UPHELD = 2;
  CONTEST_STATE_REJECTED = 3;
}
```

#### StoredPendingCommitment
```protobuf
message StoredPendingCommitment {
  string org_id = 1;
  bytes content_hash = 2;
  repeated string keywords = 3;
  string contributor_id = 4;
  uint64 epoch = 5;
  uint64 submitted_at_height = 6;
}
```

#### StoredMemoryCommitment
```protobuf
message StoredMemoryCommitment {
  string org_id = 1;
  bytes content_hash = 2;
  bytes encrypted_blob = 3;
  repeated string keywords = 4;
  string contributor_id = 5;
  uint64 epoch = 6;
  uint64 approved_at_height = 7;
  string approver = 8;
  MemoryState state = 9;
  uint64 retrieval_confidence_bps = 10;  // 0-10000 basis points
  uint64 last_decay_epoch = 12;
}
```

#### StoredMemoryRelationship
```protobuf
message StoredMemoryRelationship {
  string org_id = 1;
  string source_cid = 2;
  string target_cid = 3;
  RelationType relation_type = 4;
  string proposer = 5;
  bool approved = 6;
  uint64 epoch = 7;
}
```

#### StoredMemoryContest
```protobuf
message StoredMemoryContest {
  string org_id = 1;
  string contest_id = 2;
  string memory_cid = 3;
  string contester = 4;
  uint64 stake_amount = 5;
  string reason = 6;
  ContestState state = 7;
  uint64 epoch = 8;
}
```

#### StoredContestSettlement
```protobuf
message StoredContestSettlement {
  string org_id = 1;
  string memory_cid = 2;
  string contest_id = 3;
  ContestSettlementOutcome outcome = 4;
  string resolver = 5;
  uint64 resolved_epoch = 6;
  uint64 resolved_at_height = 7;
  int64 resolved_at_unix = 8;
  uint64 stake_refunded = 9;
  uint64 stake_burned = 10;
  uint64 treasury_reward = 11;
}
```

#### StoredValidityMetadata
```protobuf
message StoredValidityMetadata {
  string org_id = 1;
  string memory_cid = 2;
  uint64 valid_after_epoch = 3;
  uint64 valid_until_epoch = 4;
  bytes scope_tags_bz = 5;  // Serialized scope tags
}
```

#### StoredEpochMerkleRoot
```protobuf
message StoredEpochMerkleRoot {
  string org_id = 1;
  uint64 epoch = 2;
  bytes merkle_root = 3;
  uint64 memory_count = 4;
}
```

#### StoredMemoryReport
```protobuf
message StoredMemoryReport {
  string org_id = 1;
  bytes content_hash = 2;
  string reporter_id = 3;
  string reason = 4;
  uint64 epoch = 5;
  uint64 reported_at_height = 6;
}
```

### Parameters

```protobuf
message Params {
  uint64 max_pending_per_org = 1;
  uint64 pending_retention_epochs = 2;
  uint64 max_blob_size_bytes = 3;
  uint32 max_keywords_per_memory = 4;
  uint64 min_retrieval_decay_bps = 5;
  uint64 stable_threshold_bps = 6;
  uint64 degraded_threshold_bps = 7;
  uint64 dormant_threshold_bps = 8;
  uint64 initial_confidence_bps = 9;
  uint64 contest_window_epochs = 10;
}
```

| Parameter | Description |
|-----------|-------------|
| `max_pending_per_org` | Maximum pending commitments per org |
| `pending_retention_epochs` | Epochs before stale pending commitments purged |
| `max_blob_size_bytes` | Maximum encrypted blob size |
| `max_keywords_per_memory` | Maximum keywords per memory commitment |
| `min_retrieval_decay_bps` | Minimum decay rate (basis points) |
| `stable_threshold_bps` | Confidence threshold for STABLE state |
| `degraded_threshold_bps` | Confidence threshold for DEGRADED state |
| `dormant_threshold_bps` | Confidence threshold for DORMANT state |
| `initial_confidence_bps` | Initial confidence on approval (default 10000) |
| `contest_window_epochs` | Epochs before auto-rejecting stale contests |

### Messages

#### MsgSubmitCommitment
Submit a memory commitment with keywords.

```protobuf
message MsgSubmitCommitment {
  string signer = 1;
  string org_id = 2;
  bytes content_hash = 3;
  repeated string keywords = 4;
  string contributor_id = 5;
}
```

**Validation:**
- Org must exist
- Signer must be org member
- Memory bandwidth must be available
- Pending queue must not exceed limit

#### MsgApproveMemory
Approve a pending commitment with encrypted blob.

```protobuf
message MsgApproveMemory {
  string signer = 1;
  string org_id = 2;
  bytes content_hash = 3;
  bytes encrypted_blob = 4;
}
```

**Validation:**
- Signer must be org leader or moderator
- Blob size must not exceed `max_blob_size_bytes`
- Commitment must exist in pending state

**Effects:**
- Record promoted to approved
- Retrieval confidence seeded at `initial_confidence_bps`
- Approved counter incremented

#### MsgRejectMemory
Reject a pending commitment.

```protobuf
message MsgRejectMemory {
  string signer = 1;
  string org_id = 2;
  bytes content_hash = 3;
}
```

#### MsgPurgeExpired
Remove stale pending entries.

```protobuf
message MsgPurgeExpired {
  string signer = 1;
  string org_id = 2;
}
```

**Response:** `purged_count` — number of entries removed

#### MsgRelateMemories
Create a relationship between two approved memories.

```protobuf
message MsgRelateMemories {
  string sender = 1;
  string org_id = 2;
  string source_cid = 3;
  string target_cid = 4;
  RelationType relation_type = 5;
}
```

#### MsgApproveRelationship
Activate a proposed relationship.

```protobuf
message MsgApproveRelationship {
  string sender = 1;
  string org_id = 2;
  string source_cid = 3;
  string target_cid = 4;
}
```

**Effects on target memory:**
- `CONTRADICTS`: Both enter CONTESTED
- `REPLACES`: -2000 confidence bps
- `DEPRECATES`: -1000 confidence bps
- `SUPERSEDES`: Transitions to ARCHIVED

#### MsgContestMemory
Stake-backed challenge of an approved memory.

```protobuf
message MsgContestMemory {
  string sender = 1;
  string org_id = 2;
  string memory_cid = 3;
  string reason = 4;
}
```

**Validation:**
- Memory must be in APPROVED, STABLE, DEGRADED, or DORMANT state
- Sender must stake `contest_stake_vibe` (escrowed in memory module account)

**Effects:**
- Memory enters CONTESTED state
- Stake held until resolution

#### MsgResolveContest
Adjudicate a memory contest.

```protobuf
message MsgResolveContest {
  string sender = 1;
  string org_id = 2;
  string contest_id = 3;
  bool upheld = 4;  // true = memory archived, false = contest dismissed
}
```

**Effects:**
- **Upheld**: Memory archived, stake refunded + treasury reward
- **Rejected**: Stake burned, confidence re-evaluated via lifecycle

#### MsgSetValidityBounds
Set validity windows for automatic expiry.

```protobuf
message MsgSetValidityBounds {
  string sender = 1;
  string org_id = 2;
  string memory_cid = 3;
  uint64 valid_after_epoch = 4;
  uint64 valid_until_epoch = 5;
  map<string, string> scope_tags = 6;
}
```

#### MsgArchiveMemory
Manual archival without contest.

```protobuf
message MsgArchiveMemory {
  string sender = 1;
  string org_id = 2;
  string memory_cid = 3;
}
```

#### MsgReportMemory
Report a memory for violations (hub-gated, stores minimal fact on-chain).

```protobuf
message MsgReportMemory {
  string signer = 1;
  string org_id = 2;
  bytes content_hash = 3;
  string reporter_id = 4;
  string reason = 5;
  uint64 epoch = 6;
}
```

**Validation:**
- Signer must be valid Bech32 address
- Org must exist
- Memory must exist in approved state
- Reporter ID cannot be empty
- Reason cannot be empty
- No duplicate report by same reporter for same memory

**Effects:**
- StoredMemoryReport persisted with key `report/{org_id}/{content_hash_hex}/{reporter_id}`
- Duplicate reports rejected with `ErrReportExists`

### Query Endpoints

| Endpoint | REST Path | Response |
|----------|-----------|----------|
| `GetMemory` | `/wevibe/memory/v1/memory/{org_id}/{content_hash}` | `QueryGetMemoryResponse` |
| `GetPendingCommitments` | `/wevibe/memory/v1/pending/{org_id}` | `QueryGetPendingCommitmentsResponse` |
| `GetMemoryCount` | `/wevibe/memory/v1/count/{org_id}` | `QueryGetMemoryCountResponse` |
| `GetEpochMerkleRoot` | `/wevibe/memory/v1/merkle_root/{org_id}/{epoch}` | `QueryGetEpochMerkleRootResponse` |
| `ListRelationships` | `/wevibe/memory/v1/relationships/{org_id}/{cid}` | `QueryListRelationshipsResponse` |
| `GetValidity` | `/wevibe/memory/v1/validity/{org_id}/{cid}` | `QueryGetValidityResponse` |
| `ListContests` | `/wevibe/memory/v1/contests/{org_id}/{cid}` | `QueryListContestsResponse` |
| `GetContest` | `/wevibe/memory/v1/contest/{org_id}/{contest_id}` | `QueryGetContestResponse` |
| `ListContestSettlements` | `/wevibe/memory/v1/settlements/{org_id}/{memory_cid}` | `QueryListContestSettlementsResponse` |
| `GetContestSettlement` | `/wevibe/memory/v1/settlements/{org_id}/contests/{contest_id}` | `QueryGetContestSettlementResponse` |
| `Params` | `/wevibe/memory/v1/params` | `QueryParamsResponse` |

### Lifecycle State Transitions

```
                    Approval
PENDING ──────────► APPROVED
                        │
                        │ Confidence ≥ stable_threshold
                        ▼
                     STABLE
                        │
                        │ Confidence < stable_threshold
                        ▼
                     DEGRADED
                        │
                        │ Confidence < dormant_threshold
                        ▼
                     DORMANT
                        │
                        │ Expiry / Archive / Contest Upheld
                        ▼
                    ARCHIVED
```

**Contest Path:**
```
APPROVED/STABLE/DEGRADED/DORMANT ──► CONTESTED
                                         │
                          ┌──────────────┼──────────────┐
                          ▼                             ▼
                       UPHELD                        REJECTED
                          │                             │
                          ▼                             ▼
                      ARCHIVED               Confidence re-evaluated
```

### Epoch Hook Operations

The memory module's `AfterEpochEnd` hook performs:

1. **ApplyEpochDecay**: Adjusts confidence based on serves and org decay rate
   - +100 bps per serve (configurable)
   - -max(org decay_rate_bps, min_retrieval_decay_bps) when no serves

2. **CheckEpochExpiry**: Archives memories past validity window

3. **CheckContestExpiry**: Auto-rejects contests older than contest_window_epochs

4. **Merkle Root Computation**: Computes and stores epoch Merkle root

---

## Serve Module (`x/serve`)

Handles serve receipts, deduplication, and statistics.

### State Objects

#### StoredServeReceipt
```protobuf
message StoredServeReceipt {
  string org_id = 1;
  bytes memory_content_hash = 2;
  string serve_key = 3;
  string contributor_id = 4;
  uint64 epoch = 5;
  bytes nullifier = 6;
  bool is_self_serve = 7;
  string model_id = 8;     // Model that produced the session
  uint32 turn_count = 9;   // Number of turns in the session
}
```

#### StoredEpochServeStats
```protobuf
message StoredEpochServeStats {
  string org_id = 1;
  uint64 epoch = 2;
  uint64 total_serves = 3;
  uint64 unique_memories_served = 4;
  uint64 unique_serve_keys = 5;
  uint64 self_serves = 6;
  map<string, uint64> model_breakdown = 7;  // Serve count per model_id
}
```

#### StoredContributorEpochServes
```protobuf
message StoredContributorEpochServes {
  string contributor_id = 1;
  uint64 epoch = 2;
  uint64 serve_count = 3;
  uint64 self_serve_count = 4;
  repeated string org_ids = 5;
  uint64 total_turns = 6;  // Cumulative turn count across all serves
}
```

### Parameters

```protobuf
message Params {
  uint32 max_serves_per_batch = 1;
  uint32 self_serve_discount_percent = 2;
  uint32 max_serves_per_memory_per_epoch = 3;
  uint64 min_org_age_epochs = 4;
  uint64 diminishing_returns_threshold = 5;
}
```

| Parameter | Description |
|-----------|-------------|
| `max_serves_per_batch` | Maximum serves per SubmitServeBatch call |
| `self_serve_discount_percent` | Self-serve count discount percentage |
| `max_serves_per_memory_per_epoch` | Cap on serves per memory per epoch |
| `min_org_age_epochs` | Minimum org age before serving allowed |
| `diminishing_returns_threshold` | Serve count threshold for diminishing returns |

### Messages

#### ServeEntry
```protobuf
message ServeEntry {
  bytes memory_content_hash = 1;
  string serve_key = 2;
  string contributor_id = 3;
  bytes nullifier = 4;
  string model_id = 5;    // Model that produced the session
  uint32 turn_count = 6;  // Number of turns in the session
}
```

#### MsgSubmitServeBatch
Submit a batch of serve receipts.

```protobuf
message MsgSubmitServeBatch {
  string signer = 1;
  string org_id = 2;
  uint64 epoch = 3;
  repeated ServeEntry serves = 4;
}
```

**Response:**
```protobuf
message MsgSubmitServeBatchResponse {
  uint64 accepted = 1;
  uint64 rejected_duplicate = 2;
  uint64 rejected_invalid = 3;
}
```

**Validation:**
- Batch size ≤ max_serves_per_batch
- Serve bandwidth available
- Nullifiers not previously used
- Memories are approved
- Per-memory serve cap not exceeded

**Effects:**
- Attestations stored
- Epoch stats updated
- Per-memory counts updated
- Per-contributor serves tracked
- Reputation keeper notified

### Query Endpoints

| Endpoint | REST Path | Response |
|----------|-----------|----------|
| `GetEpochServeStats` | `/wevibe/serve/v1/stats/{org_id}/{epoch}` | `QueryGetEpochServeStatsResponse` |
| `GetContributorServes` | `/wevibe/serve/v1/contributor/{contributor_id}/{epoch}` | `QueryGetContributorServesResponse` |
| `GetMemoryServeCount` | `/wevibe/serve/v1/memory/{org_id}/{content_hash}/{epoch}` | `QueryGetMemoryServeCountResponse` |
| `Params` | `/wevibe/serve/v1/params` | `QueryParamsResponse` |

### Self-Serve Detection

A serve is marked `is_self_serve` when `serve_key == contributor_id`. Self-serves:
- Accumulate in separate counters
- Receive reduced XP from reputation module
- Are tracked separately in epoch stats

---

## Attestation Module (`x/attestation`)

On-chain anchor for session-level attestation data. Stores session attestation records submitted by the hub after a coding session completes. Verification logic is currently stubbed — the module stores receipt and signature hashes but defers verification until CommitLLM receipt formats freeze and cloud providers offer receipt APIs.

### State Objects

#### StoredSessionAttestation
```protobuf
message StoredSessionAttestation {
  string org_id = 1;
  bytes session_hash = 2;              // SHA-256 of session transcript (32 bytes)
  string model_id = 3;                 // e.g. "qwen3:4b", "claude-sonnet-4-20250514"
  uint32 turn_count = 4;               // conversation turns in session
  uint32 token_count = 5;              // total tokens in session
  ProviderType provider_type = 6;      // LOCAL or CLOUD
  bytes commitllm_receipt_hash = 7;    // optional: SHA-256 of CommitLLM receipt
  bytes provider_signature_hash = 8;   // optional: SHA-256 of cloud provider signature
  string contributor_id = 9;           // who ran the session
  uint64 epoch = 10;                   // WeVibe epoch
  uint64 submitted_at_height = 11;     // block height when submitted
}
```

#### ProviderType
```protobuf
enum ProviderType {
  PROVIDER_TYPE_UNSPECIFIED = 0;
  PROVIDER_TYPE_LOCAL = 1;    // Open-weight, self-hosted (CommitLLM receipt)
  PROVIDER_TYPE_CLOUD = 2;    // Closed-weight API (cloud provider signature)
}
```

### Parameters

```protobuf
message Params {
  uint64 max_attestations_per_epoch = 1;    // per-org cap per epoch
  bool require_attestation_for_serve = 2;   // future flag, defaults false
}
```

| Parameter | Description |
|-----------|-------------|
| `max_attestations_per_epoch` | Maximum session attestations per org per epoch |
| `require_attestation_for_serve` | Reserved flag for future serve-gating |

### Messages

#### MsgSubmitSessionAttestation
Submit a single session attestation for an organization.

```protobuf
message MsgSubmitSessionAttestation {
  string signer = 1;
  string org_id = 2;
  bytes session_hash = 3;
  string model_id = 4;
  uint32 turn_count = 5;
  uint32 token_count = 6;
  ProviderType provider_type = 7;
  bytes commitllm_receipt_hash = 8;
  bytes provider_signature_hash = 9;
  string contributor_id = 10;
  uint64 epoch = 11;
}
```

**Validation:**
- `signer` must be a valid Bech32 address
- `org_id` must refer to an existing org
- `session_hash` must be exactly 32 bytes (SHA-256)
- `model_id` and `contributor_id` must be non-empty
- `provider_type` must not be `UNSPECIFIED`
- Session hash must not duplicate an existing attestation for the org

**Effects:**
- Primary key `attestation/{org_id}/{session_hash_hex}` persists the full attestation
- Secondary index `session_epoch/{org_id}/{epoch}/{session_hash_hex}` enables epoch-scoped listing
- Verification stubs log status but do not reject transactions (verification deferred)

**Response:**
```protobuf
message MsgSubmitSessionAttestationResponse {
  bool accepted = 1;
  string verification_status = 2;  // e.g. "unverified: commitllm integration pending"
}
```

#### MsgUpdateParams
Update attestation module parameters. Authority-only.

```protobuf
message MsgUpdateParams {
  string authority = 1;
  Params params = 2;
}
```

### Query Endpoints

| Endpoint | REST Path | Response |
|----------|-----------|----------|
| `GetSessionAttestation` | `/wevibe/attestation/v1/session/{org_id}/{session_hash}` | `QueryGetSessionAttestationResponse` |
| `ListSessionAttestations` | `/wevibe/attestation/v1/sessions/{org_id}/{epoch}` | `QueryListSessionAttestationsResponse` |
| `Params` | `/wevibe/attestation/v1/params` | `QueryParamsResponse` |

### Verification Stubs

Two keeper methods store hashes now and will verify later:

- `VerifyCommitLLMReceipt(ctx, receiptHash)` — currently returns `(false, "unverified: commitllm integration pending")`. Will call `wevibe-commitllm-bridge` when receipt formats freeze.
- `VerifyCloudProviderSignature(ctx, signatureHash)` — currently returns `(false, "unverified: cloud provider attestation pending")`. Will verify signed response headers when APIs become available.

### Design Notes

- **Dual attestation worlds:** Local (open-weight, self-hosted via Ollama/vLLM/llama.cpp) and Cloud (closed-weight API: Claude, GPT, Gemini). `provider_type` lets downstream consumers filter accordingly.
- **Future compatibility:** `commitllm_receipt_hash` and `provider_signature_hash` are optional. When verification is implemented, only keeper verification stubs need replacing; the stored data model is already correct.

---

## Bandwidth Module (`x/bandwidth`)

Provides per-org rate limiting for memory submissions and serve receipts.

### State Objects

#### StoredBandwidthState
```protobuf
message StoredBandwidthState {
  string org_id = 1;
  uint64 epoch = 2;
  uint64 memory_used = 3;
  uint64 memory_cap = 4;
  uint64 serve_used = 5;
  uint64 serve_cap = 6;
}
```

#### StoredBandwidthOverride
```protobuf
message StoredBandwidthOverride {
  string org_id = 1;
  uint64 memory_cap = 2;
  uint64 serve_cap = 3;
}
```

### Parameters

```protobuf
message Params {
  uint64 default_memory_cap_per_epoch = 1;
  uint64 default_serve_cap_per_epoch = 2;
}
```

### Messages

#### MsgSetBandwidthOverride
Set org-specific bandwidth caps.

```protobuf
message MsgSetBandwidthOverride {
  string signer = 1;
  string org_id = 2;
  uint64 memory_cap = 3;
  uint64 serve_cap = 4;
}
```

**Authorization:** Org leader

### Query Endpoints

| Endpoint | REST Path | Response |
|----------|-----------|----------|
| `GetBandwidthState` | `/wevibe/bandwidth/v1/state/{org_id}/{epoch}` | `QueryGetBandwidthStateResponse` |
| `GetBandwidthOverride` | `/wevibe/bandwidth/v1/override/{org_id}` | `QueryGetBandwidthOverrideResponse` |
| `GetRemainingBandwidth` | `/wevibe/bandwidth/v1/remaining/{org_id}/{epoch}` | `QueryGetRemainingBandwidthResponse` |
| `Params` | `/wevibe/bandwidth/v1/params` | `QueryParamsResponse` |

---

## Reputation Module (`x/reputation`)

Tracks contributor reputation, XP, and cross-org activity.

### State Objects

#### StoredReputationStats
```protobuf
message StoredReputationStats {
  string developer_id = 1;
  uint64 memory_count = 2;
  uint64 xp = 3;
  // Plus difficulty histogram, domain tags, serve stats, etc.
}
```

#### StoredAttestedMemory
```protobuf
message StoredAttestedMemory {
  string developer = 1;
  string memory_cid = 2;
  uint32 difficulty = 3;
  uint32 quality = 4;
  repeated string domain_tags = 5;
  string provenance = 6;
}
```

#### StoredContributorOrgSet
```protobuf
message StoredContributorOrgSet {
  string developer = 1;
  repeated string org_ids = 2;
}
```

### Parameters

```protobuf
message Params {
  bool active = 1;
  uint32 max_difficulty = 2;
  uint32 max_quality = 3;
  uint64 serve_xp_per_serve = 4;
  uint64 self_serve_xp_per_serve = 5;
}
```

### Messages

#### MsgUpdateReputation
Update contributor reputation with memory attestation.

```protobuf
message MsgUpdateReputation {
  string signer = 1;
  bytes developer = 2;
  string memory_cid = 3;
  uint32 difficulty = 4;
  uint32 quality = 5;
  repeated string domain_tags = 6;
  string provenance = 7;
}
```

**Response:**
```protobuf
message MsgUpdateReputationResponse {
  uint64 xp = 1;  // Updated XP total
}
```

### Query Endpoints

| Endpoint | REST Path | Response |
|----------|-----------|----------|
| `GetReputation` | `/wevibe/reputation/v1/reputation/{developer}` | `QueryGetReputationResponse` |
| `GetXP` | `/wevibe/reputation/v1/xp/{developer}` | `QueryGetXPResponse` |
| `IsActive` | `/wevibe/reputation/v1/active` | `QueryIsActiveResponse` |
| `GetServeStats` | `/wevibe/reputation/v1/serve_stats/{developer}` | `QueryGetServeStatsResponse` |
| `GetContributorOrgSet` | `/wevibe/reputation/v1/org_set/{developer}` | `QueryGetContributorOrgSetResponse` |
| `GetCrossOrgProfile` | `/wevibe/reputation/v1/profile/{developer}` | `QueryGetCrossOrgProfileResponse` |
| `Params` | `/wevibe/reputation/v1/params` | `QueryParamsResponse` |

---

## Emissions Module (`x/emissions`)

Handles daily minting, epoch payout distribution, and work score tracking.

### State Objects

#### StoredEmissionPool
```protobuf
message StoredEmissionPool {
  uint64 total_supply = 1;
  uint64 daily_mint = 2;
  uint64 operator_share = 3;
  uint64 validator_share = 4;
  uint64 epoch = 5;
}
```

#### StoredDailyEmission
```protobuf
message StoredDailyEmission {
  uint64 epoch = 1;
  uint64 total_emitted = 2;
  uint64 operator_share = 3;
  uint64 validator_share = 4;
  map<string, uint64> operator_rewards = 5;
  map<string, uint64> validator_rewards = 6;
}
```

#### StoredWorkScore
```protobuf
message StoredWorkScore {
  string operator_id = 1;
  string org_id = 2;
  double rarity_multiplier = 3;
  double availability_score = 4;
  uint64 retrieval_volume = 5;
  double storage_score = 6;
  double retrieval_score = 7;
  double total_score = 8;
  uint64 epoch = 9;
}
```

#### StoredAsymmetricGate
```protobuf
message StoredAsymmetricGate {
  string operator_id = 1;
  string org_id = 2;
  bool storage_passed = 3;
  bool retrieval_allowed = 4;
  uint64 epoch = 5;
}
```

### Parameters

```protobuf
message Params {
  uint64 daily_mint_amount = 1;
  uint64 operator_share_percent = 2;
  uint64 validator_share_percent = 3;
  uint64 storage_weight_percent = 4;
  uint64 retrieval_weight_percent = 5;
  string rarity_multiplier_cap = 6;
  uint64 bootstrap_duration_epochs = 7;
}
```

### Messages

#### MsgMintDailyEmission
Mint the daily emission snapshot.

```protobuf
message MsgMintDailyEmission {
  string authority = 1;
  uint64 epoch = 2;
}
```

**Response:**
```protobuf
message MsgMintDailyEmissionResponse {
  uint64 total_emitted = 1;
  uint64 operator_share = 2;
  uint64 validator_share = 3;
}
```

#### MsgDistributeOperatorRewards
Record operator reward allocations.

```protobuf
message MsgDistributeOperatorRewards {
  string signer = 1;
  repeated OperatorRewardEntry rewards = 2;
  uint64 epoch = 3;
}
```

### Query Endpoints

| Endpoint | REST Path | Response |
|----------|-----------|----------|
| `GetEmissionPool` | `/wevibe/emissions/v1/pool` | `QueryGetEmissionPoolResponse` |
| `GetWorkScore` | `/wevibe/emissions/v1/work_score/{operator_id}/{org_id}/{epoch}` | `QueryGetWorkScoreResponse` |
| `GetOperatorReward` | `/wevibe/emissions/v1/operator_reward/{operator_id}` | `QueryGetOperatorRewardResponse` |
| `Params` | `/wevibe/emissions/v1/params` | `QueryParamsResponse` |

### Epoch Hook Operations

The emissions module's `AfterEpochEnd` hook performs:

1. **Mint Daily Emission**: Updates pool accounting, records epoch snapshot
2. **Process Org Payouts**: For each org with `serve_receipt_required=true`:
   - Skip if treasury balance is zero
   - Load serve receipts for epoch
   - Aggregate serves per contributor
   - Fetch org rep tiers
   - Compute payout per contributor based on tier
   - Debit treasury for each payout
   - Break loop if treasury insufficient

**Note:** Current implementation uses placeholder reputation (0) for tier matching. Orgs should ensure their lowest tier includes zero reputation.

---

## Anti-Gaming Measures

### Nullifier Deduplication
Each serve contains a nullifier; x/serve marks it and rejects any reuse, eliminating replay.

### Per-Memory Serve Cap
`max_serves_per_memory_per_epoch` limits repeat serving of the same content.

### Bandwidth Caps
Memory submissions and serves consume per-epoch quotas. Overrides require explicit action.

### Pending Commitment Limit
`max_pending_per_org` prevents hoarding of unreviewed commitments.

### Leader-Only Approvals
Only the recorded leader can approve or reject memory commitments.

### Self-Serve Detection
`serve_key == contributor_id` flags self-serves with reduced XP rewards.

### Treasury Sufficiency Check
Payout loops break when treasury runs short, preventing negative balances.

### Contest Stake Escrow
`MsgContestMemory` requires stake; losing burns it, winning refunds plus reward.

### First-Serve Markers
`memfirst` and `keyfirst` keys track unique serves per epoch for visibility.
