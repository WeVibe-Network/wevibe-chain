# WeVibe Chain Parameters Reference

This document provides a comprehensive reference for all module parameters in WeVibe Chain.

## Table of Contents

- [Overview](#overview)
- [Org Module](#org-module)
- [Memory Module](#memory-module)
- [Serve Module](#serve-module)
- [Attestation Module](#attestation-module)
- [Bandwidth Module](#bandwidth-module)
- [Reputation Module](#reputation-module)
- [Emissions Module](#emissions-module)
- [Governance Parameters](#governance-parameters)

---

## Overview

Parameters control the behavior of each module. They can be updated via governance proposals using `MsgUpdateParams` for each module. All parameters have default values defined at chain genesis.

### Query Parameters

```bash
wevibed query <module> params
```

### Update Parameters

```bash
wevibed tx <module> update-params [authority] [params_json] --from [signer]
```

---

## Org Module

Organization registration, membership, treasury, and payout configuration.

### Parameter Schema

```json
{
  "min_registration_fee": "uint64",
  "annual_renewal_fee": "uint64",
  "default_storage_quota": "uint64",
  "default_retrieval_budget": "uint64",
  "grace_period_epochs": "uint64",
  "burn_price_decay_epochs": "uint64",
  "base_burn_price": "uint64",
  "burn_price_increase_percent": "uint64"
}
```

### Parameter Descriptions

#### min_registration_fee

**Type:** uint64  
**Default:** 1,000,000 uvibe

Minimum VIBE tokens burned when registering an organization. This creates economic friction to prevent spam registrations.

```bash
wevibed tx org update-params [authority] '{"min_registration_fee":"2000000"}' --from gov
```

#### annual_renewal_fee

**Type:** uint64  
**Default:** 100,000 uvibe

Annual fee for organization renewal. Collected to maintain active status.

#### default_storage_quota

**Type:** uint64  
**Default:** 1,000,000

Default storage allocation for new organizations (in application-specific units).

#### default_retrieval_budget

**Type:** uint64  
**Default:** 500,000

Default retrieval budget for new organizations.

#### grace_period_epochs

**Type:** uint64  
**Default:** 365

Number of epochs an organization can remain without renewal before becoming inactive.

#### burn_price_decay_epochs

**Type:** uint64  
**Default:** 30

Number of epochs after which the dynamic burn price creation count decays (compounding effect removed).

#### base_burn_price

**Type:** uint64  
**Default:** 1,000,000 uvibe

Starting burn price before compounding. Each registration within the decay window multiplies the price.

#### burn_price_increase_percent

**Type:** uint64  
**Default:** 10

Percentage increase in burn price per org created within the decay window.

**Example Calculation:**
```
Price = base_burn_price * (1 + burn_price_increase_percent/100) ^ creation_count
```

With default values after 5 orgs created:
```
1,000,000 * (1.10)^5 = 1,610,510 uvibe
```

---

## Memory Module

Memory commitment queue/blob limits and dispute retention. Recall-pivot standing is computed at the edge from events plus an anchored policy version; no memory parameter stores recall weights, trust, confidence, or archive policy.

### Parameter Schema

```json
{
  "max_pending_per_org": "uint64",
  "pending_retention_epochs": "uint64",
  "max_blob_size_bytes": "uint64",
  "max_keywords_per_memory": "uint32",
  "contest_window_epochs": "uint64"
}
```

### Parameter Descriptions

#### max_pending_per_org

**Type:** uint64  
**Default:** 1,000

Maximum number of pending (unapproved) memory commitments per organization. Prevents hoarding of unreviewed content.

#### pending_retention_epochs

**Type:** uint64  
**Default:** 30

Number of epochs before pending commitments are automatically purged. Older pending commitments are removed by `MsgPurgeExpired`.

#### max_blob_size_bytes

**Type:** uint64  
**Default:** 1,048,576 (1 MiB)

Maximum size of the encrypted blob stored on approval. Blobs exceeding this size will be rejected.

#### max_keywords_per_memory

**Type:** uint32  
**Default:** 10

Maximum number of keywords per memory commitment. Keywords exceeding this count will be truncated.
These are flat labels only. They are never weighted and never gate consensus behavior.

#### contest_window_epochs

**Type:** uint64  
**Default:** 14

Number of epochs a contest remains open before auto-rejection. After this period, unresolved contests are automatically rejected.

---

## Serve Module

Serve receipt handling and rate limiting.

### Parameter Schema

```json
{
  "max_serves_per_batch": "uint32",
  "self_serve_discount_percent": "uint32",
  "max_serves_per_memory_per_epoch": "uint32",
  "min_org_age_epochs": "uint64",
  "diminishing_returns_threshold": "uint64"
}
```

### Parameter Descriptions

#### max_serves_per_batch

**Type:** uint32  
**Default:** 100

Maximum number of serve receipts per `SubmitServeBatch` transaction. Batches exceeding this are rejected.

#### self_serve_discount_percent

**Type:** uint32  
**Default:** 50

Percentage discount applied to self-serve counts when calculating unique memory serve counts. Aims to reduce inflation from self-serving.

#### max_serves_per_memory_per_epoch

**Type:** uint32  
**Default:** 10

Maximum serves recorded per memory per epoch. Prevents repeat-serving inflation.

#### min_org_age_epochs

**Type:** uint64  
**Default:** 7

Minimum organization age (in epochs) before serve receipts are accepted. Prevents spam from newly created orgs.

#### diminishing_returns_threshold

**Type:** uint64  
**Default:** 100

Serve count threshold at which diminishing returns begin. Used for calculating effective serve value.

---

## Bandwidth Module

Per-organization rate limiting.

### Parameter Schema

```json
{
  "default_memory_cap_per_epoch": "uint64",
  "default_serve_cap_per_epoch": "uint64"
}
```

### Parameter Descriptions

#### default_memory_cap_per_epoch

**Type:** uint64  
**Default:** 1,000

Default maximum memory submissions per organization per epoch. Leaders can override via `MsgSetBandwidthOverride`.

#### default_serve_cap_per_epoch

**Type:** uint64  
**Default:** 10,000

Default maximum serve receipts per organization per epoch. Leaders can override via `MsgSetBandwidthOverride`.

---

## Attestation Module

Session attestation limits and future serve gating flag.

### Parameter Schema

```json
{
  "max_attestations_per_epoch": "uint64",
  "require_attestation_for_serve": "bool"
}
```

### Parameter Descriptions

#### max_attestations_per_epoch

**Type:** uint64
**Default:** 10,000

Maximum session attestations accepted per organization per epoch. Prevents unbounded attestation growth.

#### require_attestation_for_serve

**Type:** bool
**Default:** false

Reserved flag. When enabled in a future release, serve attestations will be rejected unless a corresponding session attestation exists for the same contributor/epoch.

---

## Reputation Module

Contributor reputation tracking and XP configuration.

### Parameter Schema

```json
{
  "active": "bool",
  "max_difficulty": "uint32",
  "max_quality": "uint32",
  "serve_xp_per_serve": "uint64",
  "self_serve_xp_per_serve": "uint64"
}
```

### Parameter Descriptions

#### active

**Type:** bool  
**Default:** true

Global activation flag for the reputation module. When false, all reputation updates are rejected.

#### max_difficulty

**Type:** uint32  
**Default:** 10

Maximum difficulty value accepted in `MsgUpdateReputation`. Higher values are capped.

#### max_quality

**Type:** uint32  
**Default:** 10

Maximum quality value accepted in `MsgUpdateReputation`. Higher values are capped.

#### serve_xp_per_serve

**Type:** uint64  
**Default:** 1

Experience points awarded per valid serve receipt. Accumulates in contributor XP.

#### self_serve_xp_per_serve

**Type:** uint64  
**Default:** 0

Experience points awarded per self-serve receipt. Set to 0 to discourage self-serving behavior.

---

## Emissions Module

Daily minting, payout distribution, and work score calculation.

### Parameter Schema

```json
{
  "daily_mint_amount": "uint64",
  "operator_share_percent": "uint64",
  "validator_share_percent": "uint64",
  "storage_weight_percent": "uint64",
  "retrieval_weight_percent": "uint64",
  "rarity_multiplier_cap": "string",
  "bootstrap_duration_epochs": "uint64"
}
```

### Parameter Descriptions

#### daily_mint_amount

**Type:** uint64  
**Default:** 1,000,000 uvibe

Amount of new tokens minted daily as part of the emission schedule. Added to total supply and emission pool.

#### operator_share_percent

**Type:** uint64  
**Default:** 10

Percentage of daily mint allocated to operators. Stored in emission pool for future distribution.

#### validator_share_percent

**Type:** uint64  
**Default:** 5

Percentage of daily mint allocated to validators. Stored in emission pool for future distribution.

#### storage_weight_percent

**Type:** uint64  
**Default:** 40

Weight percentage for storage component in work score calculation.

#### retrieval_weight_percent

**Type:** uint64  
**Default:** 60

Weight percentage for retrieval component in work score calculation.

**Work Score Calculation:**
```
storage_score = availability * storage_weight_percent / 100
retrieval_score = retrieval_volume * rarity_multiplier * retrieval_weight_percent / 100
total_score = storage_score + retrieval_score
```

#### rarity_multiplier_cap

**Type:** string  
**Default:** "10"

Maximum rarity multiplier cap as a decimal string. Prevents extreme multipliers from dominating work scores.

#### bootstrap_duration_epochs

**Type:** uint64  
**Default:** 180

Duration in epochs for bootstrap credit redemption. After this period, unused credits expire.

---

## Governance Parameters

Standard Cosmos SDK governance parameters.

### Deposit Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `min_deposit` | 10,000,000 uvibe | Minimum deposit for proposal |
| `max_deposit_period` | 172,800s (2 days) | Maximum deposit period |

### Voting Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `voting_period` | 172,800s (2 days) | Length of voting period |
| `quorum` | 33.4% | Required quorum for passage |
| `threshold` | 50% | Required yes votes for passage |
| `veto` | 33.4% | No-with-veto threshold |

### Governance Address

The governance address controls parameter updates and module governance:

```
wevibe1gov...
```

This address is derived from the governance module and should be used for all `MsgUpdateParams` transactions.

---

## Organization-Specific Configuration

Organizations can configure certain parameters per-org via `MsgSetOrgConfig`:

### Per-Org Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `serve_receipt_required` | bool | Whether org requires serve receipts for payouts |
| `decay_rate_bps` | uint64 | Legacy org-config field; recall standing now comes from edge policy over events |
| `contest_stake_vibe` | uint64 | Stake amount required to contest a memory |

### Reputation Tiers

Organizations define payout tiers per contributor:

```json
[
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
```

**Tier Matching:**
- Tiers are evaluated in order from lowest to highest
- First tier where `min_reputation <= contributor_reputation <= max_reputation` is used
- Current implementation uses placeholder reputation (0), so the lowest tier acts as default

---

## Epoch Configuration

The `wevibe_epoch` is configured via genesis:

```json
{
  "identifier": "wevibe_epoch",
  "duration": "60s",
  "current_epoch": "0",
  "current_epoch_start_height": "0",
  "current_epoch_start_time": "0001-01-01T00:00:00Z",
  "epoch_counting_started": false
}
```

**Configuration Options:**

| Duration | Use Case |
|----------|----------|
| 60s | Local development |
| 3600s (1 hour) | Testnet |
| 86400s (1 day) | Production |

Set via environment variable:
```bash
EPOCH_DURATION=86400s ./scripts/init-chain.sh
```

---

## Parameter Update Examples

### Update Org Module Parameters

```bash
# Update burn price parameters
wevibed tx org update-params wevibe1gov... '{
  "base_burn_price": "2000000",
  "burn_price_increase_percent": "15"
}' --from gov --chain-id wevibe-local-1
```

### Update Memory Module Parameters

```bash
# Adjust queue/blob limits; recall policy hashes are anchored through x/serve.
wevibed tx memory update-params wevibe1gov... '{
  "max_pending_per_org": "1500",
  "max_keywords_per_memory": 20,
  "contest_window_epochs": "10"
}' --from gov --chain-id wevibe-local-1
```

### Update Serve Module Parameters

```bash
# Increase batch limits
wevibed tx serve update-params wevibe1gov... '{
  "max_serves_per_batch": "200",
  "max_serves_per_memory_per_epoch": "20"
}' --from gov --chain-id wevibe-local-1
```

### Update Reputation Module

```bash
# Enable reputation tracking
wevibed tx reputation update-params wevibe1gov... '{
  "active": true,
  "serve_xp_per_serve": "2",
  "self_serve_xp_per_serve": "0"
}' --from gov --chain-id wevibe-local-1
```

### Update Emissions Module

```bash
# Adjust emission schedule
wevibed tx emissions update-params wevibe1gov... '{
  "daily_mint_amount": "2000000",
  "operator_share_percent": "15",
  "validator_share_percent": "10"
}' --from gov --chain-id wevibe-local-1
```
