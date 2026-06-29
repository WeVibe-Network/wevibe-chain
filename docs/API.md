# WeVibe Chain API Reference

This document provides a comprehensive reference for the WeVibe Chain gRPC and REST APIs.

## Table of Contents

- [Base URLs](#base-urls)
- [Authentication](#authentication)
- [Rate Limiting](#rate-limiting)
- [Response Format](#response-format)
- [Org Module](#org-module)
- [Memory Module](#memory-module)
- [Serve Module](#serve-module)
- [Attestation Module](#attestation-module)
- [Bandwidth Module](#bandwidth-module)
- [Reputation Module](#reputation-module)
- [Emissions Module](#emissions-module)

---

## Base URLs

| Environment | gRPC | REST (gRPC-gateway) |
|-------------|------|-------------------|
| Local | `localhost:9090` | `localhost:1317` |
| Production | `https://grpc.wevibe.network` | `https://api.wevibe.network` |

## Authentication

Transactions require a valid signature from a registered account. Clients use bech32-encoded addresses with the `wevibe` prefix:

```
wevibe1abc123def456...
```

For transaction submission, authenticate using:
- **Keplr Wallet** (browser)
- **Ledger Device** (hardware)
- **Local keyring** (`--keyring-backend test` for development)

## Rate Limiting

Bandwidth limits are enforced per-organization per-epoch:

| Limit Type | Default | Override Available |
|------------|---------|-------------------|
| Memory submissions | 1000/epoch | Yes (leader) |
| Serve receipts | 10000/epoch | Yes (leader) |

Query endpoints are not rate-limited but abuse may result in service suspension.

## Response Format

### gRPC Response

Standard protobuf response messages. See individual endpoint descriptions for response types.

### REST (JSON) Response

All REST endpoints return JSON. Timestamps use RFC3339 format. Bytes are base64-encoded.

Example successful response:
```json
{
  "org_id": "my-org",
  "leader": "wevibe1abc...",
  "created_at": "12345678"
}
```

Example error response:
```json
{
  "code": 3,
  "message": "org not found",
  "details": []
}
```

---

## Org Module

### Transactions

#### MsgRegisterOrg

Register a new organization.

**REST:** `POST /wevibe/org/v1/tx/register_org`

**Request Body:**
```json
{
  "signer": "wevibe1abc...",
  "org_id": "my-org",
  "leader": "wevibe1abc...",
  "storage_quota": "1000000",
  "retrieval_budget": "500000"
}
```

**Response:** `{}`

---

#### MsgAddMember

Add a member to an organization.

**REST:** `POST /wevibe/org/v1/tx/add_member`

**Request Body:**
```json
{
  "signer": "wevibe1leader...",
  "org_id": "my-org",
  "pubkey": "wevibe1xyz...",
  "role": "moderator"
}
```

**Response:** `{}`

---

#### MsgRemoveMember

Remove a member from an organization.

**REST:** `POST /wevibe/org/v1/tx/remove_member`

**Request Body:**
```json
{
  "signer": "wevibe1leader...",
  "org_id": "my-org",
  "pubkey": "wevibe1xyz..."
}
```

**Response:** `{}`

---

#### MsgFundTreasury

Add funds to the organization treasury.

**REST:** `POST /wevibe/org/v1/tx/fund_treasury`

**Request Body:**
```json
{
  "signer": "wevibe1abc...",
  "org_id": "my-org",
  "amount": "1000000"
}
```

**Response:** `{}`

---

#### MsgWithdrawTreasury

Withdraw funds from the organization treasury.

**REST:** `POST /wevibe/org/v1/tx/withdraw_treasury`

**Request Body:**
```json
{
  "signer": "wevibe1leader...",
  "org_id": "my-org",
  "amount": "500000",
  "recipient": "wevibe1recipient..."
}
```

**Response:** `{}`

---

#### MsgSetRepTiers

Configure payout tiers for epoch rewards.

**REST:** `POST /wevibe/org/v1/tx/set_rep_tiers`

**Request Body:**
```json
{
  "signer": "wevibe1leader...",
  "org_id": "my-org",
  "tiers": [
    {
      "min_reputation": "0",
      "max_reputation": "1000",
      "max_contributions_per_epoch": "100",
      "payout_per_serve": "1"
    },
    {
      "min_reputation": "1001",
      "max_reputation": "10000",
      "max_contributions_per_epoch": "500",
      "payout_per_serve": "2"
    }
  ]
}
```

**Response:** `{}`

---

#### MsgSetOrgConfig

Update organization configuration.

**REST:** `POST /wevibe/org/v1/tx/set_org_config`

**Request Body:**
```json
{
  "signer": "wevibe1leader...",
  "org_id": "my-org",
  "serve_receipt_required": true,
  "decay_rate_bps": 100,
  "contest_stake_vibe": "100000"
}
```

**Response:** `{}`

---

#### MsgGrantTrialAllowance

Grant fee grant for trial submissions.

**REST:** `POST /wevibe/org/v1/tx/grant_trial_allowance`

**Request Body:**
```json
{
  "signer": "wevibe1leader...",
  "org_id": "my-org",
  "grantee": "wevibe1trial...",
  "daily_submissions": 10,
  "trial_days": 30
}
```

**Response:** `{}`

---

### Queries

#### GetOrg

Get organization details.

**REST:** `GET /wevibe/org/v1/org/{org_id}`

**Response:**
```json
{
  "org_id": "my-org",
  "leader": "wevibe1abc...",
  "created_at": "12345678",
  "renewal_height": "13000000",
  "storage_quota": "1000000",
  "retrieval_budget": "500000",
  "status": 0
}
```

| Status Values | Description |
|---------------|-------------|
| 0 | Active |
| 1 | Suspended |
| 2 | Closed |

---

#### GetMembers

List organization members.

**REST:** `GET /wevibe/org/v1/members/{org_id}`

**Response:**
```json
{
  "members": [
    {
      "org_id": "my-org",
      "pubkey": "wevibe1leader...",
      "role": "leader"
    },
    {
      "org_id": "my-org",
      "pubkey": "wevibe1mod...",
      "role": "moderator"
    }
  ]
}
```

---

#### IsMember

Check if an address is a member.

**REST:** `GET /wevibe/org/v1/is_member/{org_id}/{pubkey}`

**Response:**
```json
{
  "is_member": true
}
```

---

#### GetTreasury

Get organization treasury balance.

**REST:** `GET /wevibe/org/v1/treasury/{org_id}`

**Response:**
```json
{
  "balance": "1000000"
}
```

---

#### GetRepTiers

Get organization payout tiers.

**REST:** `GET /wevibe/org/v1/rep_tiers/{org_id}`

**Response:**
```json
{
  "tiers": [
    {
      "min_reputation": "0",
      "max_reputation": "1000",
      "max_contributions_per_epoch": "100",
      "payout_per_serve": "1"
    }
  ]
}
```

---

#### GetOrgConfig

Get organization configuration.

**REST:** `GET /wevibe/org/v1/config/{org_id}`

**Response:**
```json
{
  "serve_receipt_required": true,
  "decay_rate_bps": "100",
  "contest_stake_vibe": "100000"
}
```

---

#### Params

Get org module parameters.

**REST:** `GET /wevibe/org/v1/params`

**Response:**
```json
{
  "params": {
    "min_registration_fee": "1000000",
    "annual_renewal_fee": "100000",
    "default_storage_quota": "1000000",
    "default_retrieval_budget": "500000",
    "grace_period_epochs": "365",
    "burn_price_decay_epochs": "30",
    "base_burn_price": "1000000",
    "burn_price_increase_percent": "10"
  }
}
```

---

## Memory Module

### Transactions

#### MsgSubmitCommitment

Submit a memory commitment.

**REST:** `POST /wevibe/memory/v1/tx/submit_commitment`

**Request Body:**
```json
{
  "signer": "wevibe1contributor...",
  "org_id": "my-org",
  "content_hash": "base64-encoded-hash",
  "keywords": ["knowledge", "guide"],
  "contributor_id": "wevibe1contributor..."
}
```

**Response:** `{}`

---

#### MsgApproveMemory

Approve a pending memory commitment.

**REST:** `POST /wevibe/memory/v1/tx/approve_memory`

**Request Body:**
```json
{
  "signer": "wevibe1leader...",
  "org_id": "my-org",
  "content_hash": "base64-encoded-hash",
  "encrypted_blob": "base64-encoded-blob"
}
```

**Response:** `{}`

---

#### MsgRejectMemory

Reject a pending memory commitment.

**REST:** `POST /wevibe/memory/v1/tx/reject_memory`

**Request Body:**
```json
{
  "signer": "wevibe1leader...",
  "org_id": "my-org",
  "content_hash": "base64-encoded-hash"
}
```

**Response:** `{}`

---

#### MsgPurgeExpired

Purge expired pending commitments.

**REST:** `POST /wevibe/memory/v1/tx/purge_expired`

**Request Body:**
```json
{
  "signer": "wevibe1leader...",
  "org_id": "my-org"
}
```

**Response:**
```json
{
  "purged_count": "5"
}
```

---

#### MsgRelateMemories

Create a relationship between memories.

**REST:** `POST /wevibe/memory/v1/tx/relate_memories`

**Request Body:**
```json
{
  "sender": "wevibe1leader...",
  "org_id": "my-org",
  "source_cid": "cid-source",
  "target_cid": "cid-target",
  "relation_type": "RELATION_TYPE_DEPRECATES"
}
```

**Relation Types:**
| Value | Description |
|-------|-------------|
| `RELATION_TYPE_CONTRADICTS` | Contradicts target memory |
| `RELATION_TYPE_REPLACES` | Replaces target memory |
| `RELATION_TYPE_DEPRECATES` | Deprecates target memory |
| `RELATION_TYPE_SUPERSEDES` | Supersedes target memory |

**Response:** `{}`

---

#### MsgContestMemory

Contest an approved memory.

**REST:** `POST /wevibe/memory/v1/tx/contest_memory`

**Request Body:**
```json
{
  "sender": "wevibe1contester...",
  "org_id": "my-org",
  "memory_cid": "cid-to-contest",
  "reason": "Incorrect information"
}
```

**Response:** `{}`

---

#### MsgResolveContest

Resolve a memory contest.

**REST:** `POST /wevibe/memory/v1/tx/resolve_contest`

**Request Body:**
```json
{
  "sender": "wevibe1leader...",
  "org_id": "my-org",
  "contest_id": "contest-123",
  "upheld": true
}
```

**Response:** `{}`

---

#### MsgSetValidityBounds

Set validity bounds for a memory.

**REST:** `POST /wevibe/memory/v1/tx/set_validity_bounds`

**Request Body:**
```json
{
  "sender": "wevibe1leader...",
  "org_id": "my-org",
  "memory_cid": "cid-to-bound",
  "valid_after_epoch": "100",
  "valid_until_epoch": "200",
  "scope_tags": {
    "region": "us-west",
    "language": "en"
  }
}
```

**Response:** `{}`

---

#### MsgArchiveMemory

Archive a memory.

**REST:** `POST /wevibe/memory/v1/tx/archive_memory`

**Request Body:**
```json
{
  "sender": "wevibe1leader...",
  "org_id": "my-org",
  "memory_cid": "cid-to-archive"
}
```

**Response:** `{}`

---

#### MsgReportMemory

Report a memory for violations.

**REST:** `POST /wevibe/memory/v1/tx/report_memory`

**Request Body:**
```json
{
  "signer": "wevibe1reporter...",
  "org_id": "my-org",
  "content_hash": "base64-encoded-hash",
  "reporter_id": "wevibe1reporter...",
  "reason": "Violates content policy",
  "epoch": 100
}
```

**Response:** `{}`

---

### Queries

#### GetMemory

Get an approved memory.

**REST:** `GET /wevibe/memory/v1/memory/{org_id}/{content_hash}`

**Response:**
```json
{
  "memory": {
    "org_id": "my-org",
    "content_hash": "base64-encoded-hash",
    "encrypted_blob": "base64-encoded-blob",
    "keywords": ["knowledge", "guide"],
    "contributor_id": "wevibe1contributor...",
    "epoch": "100",
    "approved_at_height": "123456",
    "approver": "wevibe1leader...",
    "state": "MEMORY_STATE_APPROVED",
    "retrieval_confidence_bps": "10000",
    "last_decay_epoch": "100"
  }
}
```

| State Values | Description |
|-------------|-------------|
| 0 | Unspecified |
| 1 | Pending |
| 2 | Approved |
| 3 | Stable |
| 4 | Contested |
| 5 | Degraded |
| 6 | Dormant |
| 7 | Archived |
| 8 | Rejected |

---

#### GetPendingCommitments

List pending commitments.

**REST:** `GET /wevibe/memory/v1/pending/{org_id}`

**Response:**
```json
{
  "commitments": [
    {
      "org_id": "my-org",
      "content_hash": "base64-encoded-hash",
      "keywords": ["draft"],
      "contributor_id": "wevibe1contributor...",
      "epoch": "100",
      "submitted_at_height": "123400"
    }
  ]
}
```

---

#### GetMemoryCount

Get approved memory count.

**REST:** `GET /wevibe/memory/v1/count/{org_id}`

**Response:**
```json
{
  "count": "42"
}
```

---

#### GetEpochMerkleRoot

Get epoch Merkle root.

**REST:** `GET /wevibe/memory/v1/merkle_root/{org_id}/{epoch}`

**Response:**
```json
{
  "merkle_root": "base64-encoded-root",
  "memory_count": "42"
}
```

---

#### ListRelationships

List memory relationships.

**REST:** `GET /wevibe/memory/v1/relationships/{org_id}/{cid}`

**Response:**
```json
{
  "relationships": [
    {
      "org_id": "my-org",
      "source_cid": "cid-source",
      "target_cid": "cid-target",
      "relation_type": "RELATION_TYPE_DEPRECATES",
      "proposer": "wevibe1leader...",
      "approved": true,
      "epoch": "100"
    }
  ]
}
```

---

#### GetValidity

Get memory validity metadata.

**REST:** `GET /wevibe/memory/v1/validity/{org_id}/{cid}`

**Response:**
```json
{
  "metadata": {
    "org_id": "my-org",
    "memory_cid": "cid",
    "valid_after_epoch": "100",
    "valid_until_epoch": "200",
    "scope_tags_bz": "base64-encoded-tags"
  },
  "found": true
}
```

---

#### ListContests

List contests for a memory.

**REST:** `GET /wevibe/memory/v1/contests/{org_id}/{cid}`

**Response:**
```json
{
  "contests": [
    {
      "org_id": "my-org",
      "contest_id": "contest-123",
      "memory_cid": "cid-contested",
      "contester": "wevibe1contester...",
      "stake_amount": "100000",
      "reason": "Incorrect info",
      "state": "CONTEST_STATE_PENDING",
      "epoch": "100"
    }
  ]
}
```

| Contest State Values | Description |
|---------------------|-------------|
| 0 | Unspecified |
| 1 | Pending |
| 2 | Upheld |
| 3 | Rejected |

---

#### GetContest

Get a specific contest.

**REST:** `GET /wevibe/memory/v1/contest/{org_id}/{contest_id}`

**Response:**
```json
{
  "contest": {
    "org_id": "my-org",
    "contest_id": "contest-123",
    "memory_cid": "cid",
    "contester": "wevibe1contester...",
    "stake_amount": "100000",
    "reason": "Incorrect info",
    "state": "CONTEST_STATE_PENDING",
    "epoch": "100"
  },
  "found": true
}
```

---

#### Params

Get memory module parameters.

**REST:** `GET /wevibe/memory/v1/params`

**Response:**
```json
{
  "params": {
    "max_pending_per_org": "1000",
    "pending_retention_epochs": "30",
    "max_blob_size_bytes": "1048576",
    "max_keywords_per_memory": "10",
    "min_retrieval_decay_bps": "10",
    "stable_threshold_bps": "8000",
    "degraded_threshold_bps": "5000",
    "dormant_threshold_bps": "2000",
    "initial_confidence_bps": "10000",
    "contest_window_epochs": "14"
  }
}
```

---

## Serve Module

### Transactions

#### MsgSubmitServeBatch

Submit a batch of serve receipts.

**REST:** `POST /wevibe/serve/v1/tx/submit_serve_batch`

**Request Body:**
```json
{
  "signer": "wevibe1servicer...",
  "org_id": "my-org",
  "epoch": "100",
  "serves": [
    {
      "memory_content_hash": "base64-encoded-hash",
      "contributor_id": "wevibe1contributor...",
      "matched_keywords": ["keyword1", "keyword2"],
      "serve_key_pubkey": "base64-encoded-ed25519-pubkey",
      "serve_sig": "base64-encoded-ed25519-signature",
      "nonce": "base64-encoded-freshness-nonce"
    }
  ]
}
```

**Response:**
```json
{
  "accepted": "10",
  "rejected_duplicate": "2",
  "rejected_invalid": "0"
}
```

---

### Queries

#### GetEpochServeStats

Get epoch serve statistics.

**REST:** `GET /wevibe/serve/v1/stats/{org_id}/{epoch}`

**Response:**
```json
{
  "stats": {
    "org_id": "my-org",
    "epoch": "100",
    "total_serves": "1500",
    "unique_memories_served": "500",
    "unique_serve_keys": "50",
    "self_serves": "100"
  }
}
```

---

#### GetContributorServes

Get contributor serves for an epoch.

**REST:** `GET /wevibe/serve/v1/contributor/{contributor_id}/{epoch}`

**Response:**
```json
{
  "serves": {
    "contributor_id": "wevibe1contributor...",
    "epoch": "100",
    "serve_count": "25",
    "self_serve_count": "0",
    "org_ids": ["my-org", "other-org"]
  }
}
```

---

#### GetMemoryServeCount

Get serve count for a specific memory.

**REST:** `GET /wevibe/serve/v1/memory/{org_id}/{content_hash}/{epoch}`

**Response:**
```json
{
  "count": "15"
}
```

---

#### Params

Get serve module parameters.

**REST:** `GET /wevibe/serve/v1/params`

**Response:**
```json
{
  "params": {
    "max_serves_per_batch": "100",
    "self_serve_discount_percent": "50",
    "max_serves_per_memory_per_epoch": "10",
    "min_org_age_epochs": "7",
    "diminishing_returns_threshold": "100"
  }
}
```

---

---

## Attestation Module

### Transactions

#### MsgSubmitSessionAttestation

Submit a session attestation for a coding session.

**REST:** `POST /wevibe/attestation/v1/tx/submit_session_attestation`

**Request Body:**
```json
{
  "signer": "wevibe1contributor...",
  "org_id": "my-org",
  "session_hash": "base64-encoded-32bytes",
  "model_id": "qwen3:4b",
  "turn_count": 15,
  "token_count": 3200,
  "provider_type": "PROVIDER_TYPE_LOCAL",
  "commitllm_receipt_hash": "base64-encoded-32bytes",
  "provider_signature_hash": "",
  "contributor_id": "wevibe1contributor...",
  "epoch": 100
}
```

**Response:**
```json
{
  "accepted": true,
  "verification_status": "unverified: commitllm integration pending"
}
```

---

### Queries

#### GetSessionAttestation

Get a session attestation by org and session hash.

**REST:** `GET /wevibe/attestation/v1/session/{org_id}/{session_hash}`

**Response:**
```json
{
  "attestation": {
    "org_id": "my-org",
    "session_hash": "base64-encoded-32bytes",
    "model_id": "qwen3:4b",
    "turn_count": 15,
    "token_count": 3200,
    "provider_type": "PROVIDER_TYPE_LOCAL",
    "commitllm_receipt_hash": "base64-encoded-32bytes",
    "provider_signature_hash": "",
    "contributor_id": "wevibe1contributor...",
    "epoch": 100,
    "submitted_at_height": 500000
  }
}
```

---

#### ListSessionAttestations

List session attestations for an org in a specific epoch.

**REST:** `GET /wevibe/attestation/v1/sessions/{org_id}/{epoch}`

**Response:**
```json
{
  "attestations": [
    {
      "org_id": "my-org",
      "session_hash": "base64-encoded-32bytes",
      "model_id": "qwen3:4b",
      "turn_count": 15,
      "token_count": 3200,
      "provider_type": "PROVIDER_TYPE_LOCAL",
      "commitllm_receipt_hash": "base64-encoded-32bytes",
      "provider_signature_hash": "",
      "contributor_id": "wevibe1contributor...",
      "epoch": 100,
      "submitted_at_height": 500000
    }
  ]
}
```

---

#### Params

Get attestation module parameters.

**REST:** `GET /wevibe/attestation/v1/params`

**Response:**
```json
{
  "params": {
    "max_attestations_per_epoch": "10000",
    "require_attestation_for_serve": false
  }
}
```

---

## Bandwidth Module

### Queries

#### GetBandwidthState

Get current bandwidth state.

**REST:** `GET /wevibe/bandwidth/v1/state/{org_id}/{epoch}`

**Response:**
```json
{
  "state": {
    "org_id": "my-org",
    "epoch": "100",
    "memory_used": "500",
    "memory_cap": "1000",
    "serve_used": "5000",
    "serve_cap": "10000"
  }
}
```

---

#### GetBandwidthOverride

Get org bandwidth override.

**REST:** `GET /wevibe/bandwidth/v1/override/{org_id}`

**Response:**
```json
{
  "override": {
    "org_id": "my-org",
    "memory_cap": "2000",
    "serve_cap": "20000"
  },
  "has_override": true
}
```

---

#### GetRemainingBandwidth

Get remaining bandwidth.

**REST:** `GET /wevibe/bandwidth/v1/remaining/{org_id}/{epoch}`

**Response:**
```json
{
  "memory_remaining": "500",
  "serve_remaining": "5000"
}
```

---

#### Params

Get bandwidth module parameters.

**REST:** `GET /wevibe/bandwidth/v1/params`

**Response:**
```json
{
  "params": {
    "default_memory_cap_per_epoch": "1000",
    "default_serve_cap_per_epoch": "10000"
  }
}
```

---

## Reputation Module

### Transactions

#### MsgUpdateReputation

Update contributor reputation.

**REST:** `POST /wevibe/reputation/v1/tx/update_reputation`

**Request Body:**
```json
{
  "signer": "wevibe1attester...",
  "developer": "base64-encoded-developer-id",
  "memory_cid": "cid-verified",
  "difficulty": 5,
  "quality": 8,
  "domain_tags": ["rust", "cosmos"],
  "provenance": "verified"
}
```

**Response:**
```json
{
  "xp": "1500"
}
```

---

### Queries

#### GetReputation

Get contributor reputation summary.

**REST:** `GET /wevibe/reputation/v1/reputation/{developer}`

**Response:**
```json
{
  "developer_id": "wevibe1developer...",
  "memory_count": "25",
  "xp": "1500"
}
```

---

#### GetXP

Get contributor XP.

**REST:** `GET /wevibe/reputation/v1/xp/{developer}`

**Response:**
```json
{
  "xp": "1500"
}
```

---

#### IsActive

Check if reputation module is active.

**REST:** `GET /wevibe/reputation/v1/active`

**Response:**
```json
{
  "active": true
}
```

---

#### GetServeStats

Get contributor serve statistics.

**REST:** `GET /wevibe/reputation/v1/serve_stats/{developer}`

**Response:**
```json
{
  "serve_count": "500",
  "self_serve_count": "50",
  "org_breadth": "3",
  "serve_xp": "450",
  "first_seen_epoch": "50"
}
```

---

#### GetContributorOrgSet

Get contributor org set.

**REST:** `GET /wevibe/reputation/v1/org_set/{developer}`

**Response:**
```json
{
  "org_ids": ["my-org", "other-org", "third-org"]
}
```

---

#### GetCrossOrgProfile

Get cross-organization profile.

**REST:** `GET /wevibe/reputation/v1/profile/{developer}`

**Response:**
```json
{
  "developer_id": "wevibe1developer...",
  "memory_count": "25",
  "xp": "1500",
  "serve_count": "500",
  "self_serve_count": "50",
  "org_breadth": "3",
  "serve_xp": "450",
  "first_seen_epoch": "50",
  "org_ids": ["my-org", "other-org"],
  "domain_tags": {
    "rust": "10",
    "cosmos": "15"
  }
}
```

---

#### Params

Get reputation module parameters.

**REST:** `GET /wevibe/reputation/v1/params`

**Response:**
```json
{
  "params": {
    "active": true,
    "max_difficulty": 10,
    "max_quality": 10,
    "serve_xp_per_serve": "1",
    "self_serve_xp_per_serve": "0"
  }
}
```

---

## Emissions Module

### Transactions

#### MsgMintDailyEmission

Mint daily emission (authority only).

**REST:** `POST /wevibe/emissions/v1/tx/mint_daily_emission`

**Request Body:**
```json
{
  "authority": "wevibe1gov...",
  "epoch": "101"
}
```

**Response:**
```json
{
  "total_emitted": "1000000",
  "operator_share": "100000",
  "validator_share": "50000"
}
```

---

#### MsgDistributeOperatorRewards

Distribute operator rewards.

**REST:** `POST /wevibe/emissions/v1/tx/distribute_operator_rewards`

**Request Body:**
```json
{
  "signer": "wevibe1operator...",
  "rewards": [
    {
      "operator_id": "op-1",
      "amount": "50000"
    }
  ],
  "epoch": "100"
}
```

**Response:** `{}`

---

### Queries

#### GetEmissionPool

Get emission pool status.

**REST:** `GET /wevibe/emissions/v1/pool`

**Response:**
```json
{
  "total_supply": "1000000000",
  "daily_mint": "1000000",
  "operator_share": "100000",
  "validator_share": "50000",
  "epoch": "100"
}
```

---

#### GetWorkScore

Get work score for an operator.

**REST:** `GET /wevibe/emissions/v1/work_score/{operator_id}/{org_id}/{epoch}`

**Response:**
```json
{
  "operator_id": "op-1",
  "org_id": "my-org",
  "rarity_multiplier": "1.5",
  "availability_score": "0.95",
  "retrieval_volume": "1000",
  "total_score": "142.5",
  "epoch": "100"
}
```

---

#### GetOperatorReward

Get operator reward amount.

**REST:** `GET /wevibe/emissions/v1/operator_reward/{operator_id}`

**Response:**
```json
{
  "amount": "50000"
}
```

---

#### Params

Get emissions module parameters.

**REST:** `GET /wevibe/emissions/v1/params`

**Response:**
```json
{
  "params": {
    "daily_mint_amount": "1000000",
    "operator_share_percent": "10",
    "validator_share_percent": "5",
    "storage_weight_percent": "40",
    "retrieval_weight_percent": "60",
    "rarity_multiplier_cap": "10",
    "bootstrap_duration_epochs": "180"
  }
}
```

---

## Error Codes

| Code | gRPC Status | Description |
|------|-------------|-------------|
| 0 | OK | Success |
| 1 | Canceled | Operation canceled |
| 2 | Unknown | Unknown error |
| 3 | InvalidArgument | Invalid parameter |
| 4 | DeadlineExceeded | Deadline exceeded |
| 5 | NotFound | Resource not found |
| 6 | AlreadyExists | Resource already exists |
| 7 | PermissionDenied | Permission denied |
| 8 | ResourceExhausted | Resource exhausted |
| 10 | Aborted | Operation aborted |
| 16 | Unauthenticated | Not authenticated |
| 17 | Unauthorized | Not authorized |
| 2 | Internal | Internal error |

---

## Pagination

List endpoints support pagination via Cosmos SDK's pagination proto:

**Request:**
```json
{
  "pagination": {
    "key": "base64-encoded-key",
    "offset": "100",
    "limit": "50",
    "count_total": true
  }
}
```

**Response:**
```json
{
  "results": [...],
  "pagination": {
    "next_key": "base64-encoded-next-key",
    "total": "150"
  }
}
```
