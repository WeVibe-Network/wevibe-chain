# WeVibe Chain Topology (Sprint 24 / CO-240, Sprint 25 / CO-258, CO-259, Sprint 27 / CO-265)

## Development Settings (CO-259)

**init-chain.sh dev-mode flags (GAP-T4):**

`scripts/init-chain.sh` enforces the following for local dev:
- Enables gRPC and binds it to `0.0.0.0:9090` so hub + dashboard can reach the node from Docker.
- Sets `pruning = "custom"` with `pruning-keep-recent = 100` and `pruning-interval = 10`, balancing stability with disk usage.
- Forces `minimum-gas-prices = "0.025uvibe"`, matching the hub's fee calculator and preventing `insufficient fee` errors during dogfood.
- Disables IAVL fast-node optimization (`iavl-disable-fastnode = true`) to avoid transient `version does not exist` errors while iterating on state.

**Why:** These settings keep the local node reachable, deterministic, and in sync with fee expectations used by the dogfood pipeline and `wevibe-hub` broadcast client.

**Status:** Pre-MVP only. These settings are unacceptable for testnet or production (D-13.13).

## Smoke Test RPC Endpoint (CO-265)

`scripts/smoke-test.sh` uses:

- `RPC_URL="${RPC_URL:-http://localhost:26657}"`

Behavior:

- Default works with local Docker Compose because `wevibed` maps `26657:26657` to host.
- Override is supported for non-default topologies: `RPC_URL=http://<host>:26657 scripts/smoke-test.sh`.

## System Components

### WeVibe Chain
WeVibe Chain is the Cosmos SDK application that anchors Sprint 23 functionality. It couples the staking, governance, bank, epochs, and distribution foundations with purpose-built modules that orchestrate organization onboarding, memory curation, serving attestations, contributor reputation, bandwidth throttling, and emission incentives. The app wiring in `app/app.go` binds Cosmos services such as the consensus parameter keeper, staking hooks, and message routers while mounting the custom store keys `attestation`, `bandwidth`, `emissions`, `memory`, `org`, `reputation`, and `serve`. Each block processes ABCI messages through `BaseApp`, so all module state transitions participate in the same deterministic state machine.

WeVibe Chain exposes two public APIs: Tendermint RPC for block and transaction gossip, and gRPC-gateway REST routes derived from the protobuf services in `proto/wevibe/**`. Clients authenticate through bech32 addresses with the `wevibe` prefix, and fees are denominated in `uvibe`. Governance retains authority over critical parameters by injecting the `gov` module address into keepers that enforce privileged updates.

## Sprint 24 Updates (CO-240)

- **Keyword weight decay model**: `x/memory` applies decay to per-keyword weights, not memory-level confidence.
  - `KeywordWeight` message on-chain: `{ keyword, weight, serve_count, denial_count }`
  - All keywords on a memory boost/decay equally per event
  - Serve boost: +100 bps per serve event, cap 5 serves/epoch
  - Denial decay: -500 bps per denial event
  - Idle decay: -50 bps per epoch with zero activity (checked via `last_active_epoch`)
  - Bootstrap grace: 14 epochs, no decay during grace period
  - Archived when ALL keyword weights = 0.0 (not confidence == 0)
- **Confidence field KILLED**: `StoredMemoryCommitment.confidence` removed. Keyword weights ARE the health metric.
- **`last_active_epoch`**: Tracks when memory last had activity (serve/denial), used for idle decay eligibility.
- **Decay parameters in params.proto**: `denial_decay_bps` (500), `serve_boost_bps` (100), `bootstrap_grace_epochs` (14), `idle_decay_bps` (50).

### WeVibe Hub
WeVibe Hub is the stateless service tier that indexes chain data and republishes it to product surfaces. It watches new blocks via WebSocket subscriptions, materializes module-specific views (for example, the set of live memory commitments per organization and the latest serve statistics), and pushes normalized payloads over WebSockets and gRPC streams to downstream consumers. Hub workers rely on the same query services documented below, so the hub remains a pure observer: no keeper logic runs outside the chain. The hub additionally relays event notifications (such as `x/serve` batch acceptance or `x/emissions` payout summaries) to the Dashboard and any analytics pipelines.

### WeVibe Dashboard
WeVibe Dashboard is the interactive client for organization leaders, contributors, and ecosystem analysts. It consumes Hub streams to render live metrics, but it also calls chain gRPC endpoints directly for authoritative reads, especially when confirming membership, treasury balances, or payout receipts. The Dashboard assembles transaction payloads (for example `MsgSubmitServeBatch`, `MsgApproveMemory`, or `MsgSetRepTiers`) and relays them through wallet connectors to broadcast against WeVibe Chain. Because all module parameters are queryable, the Dashboard can surface enforcement thresholds (pending memory caps, serve batch limits, reputation XP weights) without hardcoding business logic.

## Chain Module Topology

### Org Module (`x/org`)
- **Owned KV prefixes and state objects**
  - `org/`: `StoredOrg` records authoritative metadata (leader, creation height, renewal height, storage quota, retrieval budget, status, domain).
  - `member/{org}/{pubkey}`: `StoredMemberRecord` snapshots membership roles (leader plus custom roles).
  - `dynprice/`: `StoredDynamicPrice` tracks the adaptive burn price inputs for new registrations.
  - `treasury/{org}`: `StoredTreasury` mirrors each organization treasury balance in string-encoded `uvibe`.
  - `reptier/{org}`: `StoredRepTierConfig` stores ordered payout tiers keyed by reputation bounds.
  - `orgconfig/{org}`: `StoredOrgConfig` persists runtime toggles (currently `serve_attestation_required`).
  - `params`: module-level `Params` (fees, quota defaults, burn behavior).
- **Message handlers**
  - `MsgRegisterOrg` burns the dynamic price through `BankKeeper`, writes `StoredOrg`, auto-adds the leader as a member, and refreshes the dynamic price counter via `updateDynamicPriceOnCreation`.
  - `MsgAddMember` and `MsgRemoveMember` gate membership mutations.
  - `MsgUpdateMemberRole` (CO-011b) lets the leader promote or demote a member's role (member ↔ moderator). Leader role itself cannot be changed via this message.
  - `MsgRotateEpoch` (CO-011b) increments the epoch rotation counter after member removal. Chain-side record only; cryptographic rotation happens hub-side via Umbral sidecar.
  - `MsgTransferLeadership` (CO-011b) transfers leadership to another existing member; old leader becomes "member".
  - `MsgCloseOrg` (CO-011b) permanently closes an org by setting status to `OrgStatus_CLOSED` (3).
  - `MsgUpdateParams` is authority gated and rewrites `Params`.
  - `MsgFundTreasury` and `MsgWithdrawTreasury` move coins between external accounts and the module account, mirroring balances in `StoredTreasury`.
  - `MsgSetRepTiers` and `MsgSetOrgConfig` let leaders update payout configuration (via `payout_per_memory`) and serve expectations.
- **Query endpoints** (`/wevibe/org/v1/...`)
  - `org/{org_id}` returns flattened `QueryGetOrgResponse`.
  - `members/{org_id}` enumerates `MemberInfo` records.
  - `is_member/{org_id}/{pubkey}` resolves boolean membership.
  - `treasury/{org_id}` fetches the current balance string, while `rep_tiers/{org_id}` and `config/{org_id}` expose payout tiers and org config.
  - `params` yields module parameters.
- **Keeper dependencies**: Requires `BankKeeper` for burns and treasury movements; all other logic is local state management.
- **Genesis contents** (`InitGenesis` / `ExportGenesis`)
  - Pre-seeded orgs, members, dynamic price snapshot, treasuries, reputation tiers, and org configs. Initialization validates every entry and rewrites them into the prefixed stores; exports iterate each prefix for state machine dumps.

### Memory Module (`x/memory`)
- **Owned KV prefixes and state objects**
  - `pending/{org}/...`: `StoredPendingCommitment` keyed by content hash, with `repeated KeywordWeight keywords`, contributor id, epoch, and submission height.
  - `approved/{org}/...`: `StoredMemoryCommitment` storing the encrypted blob, `repeated KeywordWeight keywords` (no confidence field), contributor attribution, lifecycle state, `last_active_epoch`, wrapped DEK, epoch, approval height, and approver id.
  - `relationship/{org}:{source}:{target}`: `StoredMemoryRelationship` edges encoding CONTRADICTS / REPLACES / DEPRECATES / SUPERSEDES links.
  - `contest/{org}:{contestID}`: `StoredMemoryContest` capturing challenger address, stake amount, epoch, and resolution state.
  - `validity/{org}:{cid}`: `StoredValidityMetadata` persisting validity windows and opaque scope tags.
  - `merkle/{org}/{epoch}`: `StoredEpochMerkleRoot` caches the per-epoch Merkle root plus memory count.
  - `count/{org}`: big-endian counter of approved memories per org.
  - `params`: `Params` enforcing queue sizes, blob size ceiling, decay parameters (denial_decay_bps, serve_boost_bps, idle_decay_bps, bootstrap_grace_epochs).
- **Keyword weight decay model** (CO-240 — confidence KILLED)
  - At each `wevibe_epoch` close, `ApplyIdleDecay` applies idle decay (50 bps) to all keywords on memories with `last_active_epoch < current_epoch - grace_period`.
  - `ApplyDenialDecay` is called on denial TX: -500 bps per denial event to all keyword weights.
  - `ApplyServeBoost` is called on serve TX: +100 bps per serve, cap 5 serves/epoch per memory.
  - **Serve cap**: Tracked per-memory via `ServeKeeper.GetMemoryServeCountForEpoch`. If serves >= 5 for this epoch, skip boost.
  - **Bootstrap grace** (`BootstrapGraceEpochs`, default 14): memories within grace period skip decay (not boost).
  - **Archive condition**: all keyword weights = 0.0 (not confidence == 0).
  - All keywords on a memory boost/decay equally per event.
  - `last_active_epoch` updated on every serve/denial TX.
- **Message handlers**
  - `MsgSubmitCommitment` validates membership indirectly via the org keeper, consumes memory bandwidth via `BandwidthKeeper.ConsumeMemoryBandwidth`, ensures the pending queue is below `max_pending_per_org`, and writes to `pending/{org}/{hash}`.
  - `MsgApproveMemory` checks leadership, blob size, migrates the record to `approved/{org}/{hash}`, seeds retrieval confidence, increments the `count/{org}` counter, and deletes the pending entry.
  - `MsgReportMemory` lets authorized reporters submit evidence of memory violations; maintains relationship and validity metadata through epoch hooks.
  - `MsgUpdateParams` is authority guarded and rewrites parameter state (including decay floors and thresholds).
- **Query endpoints** (`/wevibe/memory/v1/...`)
  - `memory/{org_id}/{content_hash}` returns a stored approved memory with lifecycle metadata.
  - `pending/{org_id}` lists outstanding commitments, supporting review queues.
  - `relationships/{org_id}/{cid}` enumerates active relationships for a memory.
  - `validity/{org_id}/{cid}` returns validity windows and scope tags.
  - `contest/{org_id}/{contest_id}` / `contests/{org_id}/{cid}` expose contest status.
  - `count/{org_id}` reports the approved memory counter.
  - `merkle_root/{org_id}/{epoch}` retrieves the cached Merkle root and memory count.
  - `params` exposes queue, blob, confidence, decay, and contest settings.
- **Keeper dependencies**: Requires `OrgKeeper` (membership, leadership, config), `BandwidthKeeper` (consumption accounting), `BankKeeper` (contest escrow), and `ServeKeeper` (serve counts for confidence decay).
- **Genesis contents**: Pending commitments, approved memories with confidence metadata, relationships, validity records, contests, and cached Merkle roots can be bootstrapped and are faithfully exported.

### Serve Module (`x/serve`)
- **Owned KV prefixes and state objects**
  - `nullifier/{hash}`: set membership guard to block duplicate serve attestations.
  - `attestation/{org}/{epoch}/{nullifier}`: `StoredServeAttestation` capturing memory hash, serve key, contributor id, wallet address, epoch, nullifier, self-serve flag, model_id, and turn_count.
  - `stats/{org}/{epoch}`: `StoredEpochServeStats` summarizing totals, unique memories, unique serve keys, self-serve counts, and model_breakdown.
  - `contributor/{id}/{epoch}`: `StoredContributorEpochServes` storing per-contributor serve counts, self-serve counts, org coverage, and total_turns.
  - `memcount/{org}/{hash}/{epoch}`, `memfirst/...`, and `keyfirst/...` track per-epoch uniqueness accounting for stats.
  - `denial/{org}/{epoch}/{memory_hash}`: `StoredDenialAttestation` tracking denial count per memory per epoch (CO-225).
  - `params`: `Params` bounding batch size, self-serve discount percent, per-memory serve caps, minimum org age, and diminishing return threshold.
- **Denial batch submission** (CO-225)
  - `MsgSubmitDenialBatch`: org leaders submit denial attestations for memories that produced incorrect or harmful outputs.
  - Handler validates: org exists, memories are approved, nullifiers are unique per batch, batch size ≤ `max_serves_per_batch`.
  - Each entry persists a `StoredDenialAttestation`; `MemoryKeeper` queries denial count via `GetMemoryDenialCountForEpoch`.
  - Denials flow into the memory module's dual-vector decay model as the negative-signal vector.
  - **Event emission** (CO-016): Emits `denial_batch_submitted` event with attributes `{org_id, submitter, epoch, accepted_count, rejected_count, block_height}`. Queryable via CometBFT `tx_search` as `denial_batch_submitted.org_id='<org_id>' AND denial_batch_submitted.submitter='<signer>'`.
- **Message handlers**
  - `MsgSubmitServeBatch` enforces batch size, consumes serve bandwidth, rejects repeated nullifiers, validates approved memories through the memory keeper, and determines the `is_self_serve` flag. Each `ServeEntry` includes `contributor_wallet` for reputation keying.
  - `MsgUpdateParams` updates configuration under governance authority.
- **Query endpoints** (`/wevibe/serve/v1/...`)
  - `stats/{org_id}/{epoch}` exposes the stored epoch stats.
  - `contributor/{contributor_id}/{epoch}` aggregates serve counts across orgs.
  - `memory/{org_id}/{content_hash}/{epoch}` returns the serve count for a specific memory and epoch.
  - `params` discloses operational limits.
- **Keeper dependencies**: Calls into `OrgKeeper` (org existence), `MemoryKeeper` (approved memory existence and denial count queries), `BandwidthKeeper` (serve throttling), and `ReputationKeeper` (serve XP credit keyed by wallet address).
- **Genesis contents**: Attestations, epoch stats, and contributor serve tallies; exported iteratively by prefix.

### Bandwidth Module (`x/bandwidth`)
- **Owned KV prefixes and state objects**
  - `state/{org}/{epoch}`: `StoredBandwidthState` capturing usage and caps for memory submissions and serves.
  - `override/{org}`: `StoredBandwidthOverride` letting leaders customize caps.
  - `params`: `Params` specifying default per-epoch caps.
- **Message handlers**
  - `MsgSetBandwidthOverride` (leader-only) writes overrides for both memory and serve caps.
  - `MsgUpdateParams` (authority) adjusts global caps.
- **Query endpoints** (`/wevibe/bandwidth/v1/...`)
  - `state/{org_id}/{epoch}` retrieves usage and caps.
  - `override/{org_id}` reports any override with a boolean flag.
  - `remaining/{org_id}/{epoch}` computes remaining capacity.
  - `params` exposes defaults.
- **Keeper dependencies**: Uses `OrgKeeper` to validate org existence and leadership; otherwise self-contained.
- **Genesis contents**: Pre-seeded bandwidth states and overrides with deduplication of duplicate overrides.

### Reputation Module (`x/reputation`)
- **Owned KV prefixes and state objects**
  - `active/`: module activation flag.
  - `stats/{developer_bytes}`: `StoredReputationStats` capturing memory counts, difficulty histogram, domain tags, provenance breakdown, XP, serve counts, self-serve counts, org breadth, first seen epoch, and serve XP.
  - `memory/{developer_bytes}/{cid}`: `StoredAttestedMemory` for per-memory provenance.
  - `orgset/{developer_bytes}`: `StoredContributorOrgSet` recording org coverage.
  - `params`: `Params` toggling activity and XP weights.
- **Message handlers**
  - `MsgUpdateReputation` validates payloads, updates stats, persists attested memory metadata, and returns the updated XP.
  - `MsgUpdateParams` is authority gated and reconfigures XP weights or activation.
- **Serve integration**
  - `RecordServe` is invoked by the serve keeper to record serve counts, self-serve counts, XP grants, first seen epoch, and org breadth. It also maintains the contributor org set.
- **Query endpoints** (`/wevibe/reputation/v1/...`)
  - `reputation/{developer}` returns aggregated stats.
  - `xp/{developer}` returns XP.
  - `active` reports module activation state.
  - `serve_stats/{developer}` exposes serve counts and XP contributions.
  - `org_set/{developer}` returns the contributor org set.
  - `profile/{developer}` surfaces a cross-org profile with tags and counts.
  - `params` discloses activation and XP configuration.
- **Keeper dependencies**: No external module keepers beyond the authority address; the serve module calls into it.
- **Genesis contents**: Activation flag, stats, and org sets.

### Emissions Module (`x/emissions`)
- **Owned KV prefixes and state objects**
  - `pool/`: `StoredEmissionPool` retaining total supply, `daily_mint`, `operator_share`, `validator_share`, and last epoch.
  - `emission/{epoch}`: `StoredDailyEmission` capturing minted totals plus maps of reward assignments.
  - `opreward/{id}` and `valreward/{id}`: reward ledgers for service reward recipients and validators.
  - `workscore/{id}/{org}/{epoch}`: `StoredWorkScore` measuring rarity, availability, retrieval volume, and computed scores.
  - `gate/{id}/{org}/{epoch}`: `StoredAsymmetricGate` gating retrieval based on storage tests.
  - `bootstrap/{id}` and `bootstrappool/`: `StoredBootstrapCredit` allocations and remaining credit pools.
  - `params`: `Params` covering mint rate, share percents, weighting, rarity cap, and bootstrap duration.
- **Message handlers**
  - `MsgMintDailyEmission` (authority) mints the daily amount, rolls the emission pool epoch forward, and persists a `StoredDailyEmission`.
  - `MsgDistributeOperatorRewards` accepts compatibility payloads with `OperatorRewardEntry` items and writes reward balances alongside emission record updates.
  - `MsgUpdateParams` rewrites module parameters under governance.
- **Query endpoints** (`/wevibe/emissions/v1/...`)
  - `pool` fetches the emission pool summary.
  - `work_score/{operator_id}/{org_id}/{epoch}` returns the stored work score.
  - `operator_reward/{operator_id}` returns the outstanding reward amount for the compatibility field.
  - `params` exposes mint weights and bootstrap configuration.
- **Keeper dependencies**: Depends on `ServeKeeper` (for attestation pulls and denial counting), `OrgKeeper` (for org enumeration, config, and treasury debits), and `MemoryKeeper` (for approved memory counts per contributor). It also touches `math` utilities for payout arithmetic.
- **Genesis contents**: Emission pool snapshot, daily emissions, reward ledgers, bootstrap credits, work scores, and asymmetric gates.
- **Epoch hooks**
  - Registered with `EpochsKeeper.AfterEpochEnd`, it mints daily emissions, iterates orgs that require serve attestations, aggregates contributor serves per org, debits treasuries via `OrgKeeper.DebitTreasury`, and logs aggregate payout metrics.

## Cross-Module Data Flow

1. **Organization registration**
   1. A leader submits `MsgRegisterOrg` through the dashboard.
   2. The org keeper validates the payload, checks the dynamic price, burns the `uvibe` fee using `BankKeeper`, and stores the org metadata under `org/{org_id}`.
   3. During registration the leader is inserted into `member/{org}/{pubkey}` with role `leader`, and the dynamic price counter increments.
   4. Genesis or runtime watchers (Hub) query `/wevibe/org/v1/org/{org_id}` and `/wevibe/org/v1/config/{org_id}` to broadcast the new organization to clients.

2. **Memory submission and approval**
   1. Contributors call `MsgSubmitCommitment` with a content hash, keywords, and contributor id.
   2. `MemoryKeeper.SubmitCommitment` checks the org exists, consumes memory bandwidth via `BandwidthKeeper.ConsumeMemoryBandwidth`, ensures the pending queue is below `max_pending_per_org`, and writes to `pending/{org}/{hash}`.
   3. An authorized reviewer (typically the leader) calls `MsgApproveMemory`, which verifies leadership with `OrgKeeper.IsLeader`, checks blob size against `max_blob_size_bytes`, migrates the record to `approved/{org}/{hash}`, increments the `count/{org}` counter, and deletes the pending entry.
   4. The approval path logs to the hub, which updates dashboards and triggers downstream proof packaging.

3. **Serve attestation ingestion**
   1. Serve nodes batch attestations inside `MsgSubmitServeBatch` along with the org id and epoch.
   2. `ServeKeeper.ProcessServeBatch` enforces `max_serves_per_batch`, consumes serve bandwidth, rejects repeated nullifiers, validates that each memory is approved, and determines the `is_self_serve` flag.
   3. For each accepted entry the keeper writes an attestation, updates stats, increments per-memory counts, tracks unique serve keys, and accumulates per-contributor serve totals.
   4. The keeper forwards each accepted serve to `ReputationKeeper.RecordServe`, which increments serve XP, org breadth, and self-serve tallies when active.
   5. Hub listeners subscribe to `/wevibe/serve/v1/stats/{org}/{epoch}` and contributor endpoints to render real-time service metrics.

4. **Epoch payout processing** (CO-225 — pay-per-memory)
   1. At the close of each `wevibe_epoch`, `EmissionsKeeper.AfterEpochEnd` triggers.
   2. It mints the daily emission via `MintDailyEmission`, updating `pool/` and `emission/{epoch}`.
   3. The keeper iterates all orgs via `OrgKeeper.GetAllOrgs`, reading configs to ensure `serve_attestation_required` is true and treasuries are positive.
   4. For each org, it counts **approved memories** per contributor via `MemoryKeeper.GetApprovedCountByContributor` (not serve counts).
   5. Qualification gate: contributor must have `min_contributions_per_epoch` (from org config) or no leader gate is set (`== 0`).
   6. Tier cap: `MaxContributionsPerEpoch` bounds per-contributor memory count; payouts beyond the cap roll to the next tier.
   7. Each payout debits the org treasury with `OrgKeeper.DebitTreasury`; totals are logged alongside emission data.
   8. The `payout_per_memory` field (renamed from `payout_per_serve`) is used instead of serve-count-based payout.

5. **Epoch Merkle root computation**
   1. `MemoryKeeper.AfterEpochEnd` listens to the same epoch hook.
   2. It enumerates `approved/{org}/...` entries for the given epoch, collates content hashes, computes a Merkle root via `types.ComputeMerkleRoot`, and stores the result under `merkle/{org}/{epoch}`.
   3. The hub monitors these updates to publish verifiable data availability proofs for retrieval partners.

## State Machine Transitions

### Organizations
1. **Registration**: Inserting a `StoredOrg` initializes membership, treasury (zero), dynamic price counters, and default config.
2. **Lifecycle flags**: The status integer in `StoredOrg.Status` maps to `types.OrgStatus` values (for example active, suspended, or closed). Status updates flow through `UpdateOrgStatus` as governance or automation dictates.
3. **Quotas and budgets**: Leaders can adjust storage quotas and retrieval budgets via keeper helpers (`UpdateStorageQuota`, `UpdateRetrievalBudget`), supporting renewal cycles and usage-based upsizing.
4. **Treasury evolution**: Treasury balances grow via `MsgFundTreasury` and shrink through withdrawals or epoch payouts debited by the emissions keeper. Balances are always mirrored in `treasury/{org}` as decimal strings.
5. **Reputation tiers and config**: Tier adjustments and config toggles update payout logic and serve requirements in place without migrations.

### Memories
1. **Pending**: Commitments enter the `pending` prefix after passing bandwidth throttling.
2. **Approved**: Approvals promote records to `approved`, capturing encrypted blobs and approver ids; the commitment record is deleted.
3. **Retention**: `MsgPurgeExpired` or scheduled purges remove stale pending entries beyond the retention window.
4. **Counting and analytics**: Approved counters and Merkle roots produce deterministic supply metrics widely used by the hub and emissions logic.

### Serves
1. **Batch ingestion**: Serve batches are either accepted (with attestations and stats persisted) or rejected (duplicate nullifier, memory not found, per-memory cap exceeded).
2. **Statistics**: Each accepted serve increments total counts, uniqueness trackers, and per-contributor tallies stored under their own prefix.
3. **Reputation linkage**: Accepted serves cascade into the reputation keeper, awarding XP and broadening org coverage; these stats feed payout logic through reputation tiers.
4. **Nullifiers**: Stored nullifiers guarantee idempotency; replayed proofs fail before affecting counts.

### Contributor Reputation
1. **Activation**: Governance toggles the module via `Params.active`; inactive mode causes updates to fail with `ErrReputationNotActive`.
2. **Memory attestation**: `MsgUpdateReputation` and keeper helper `AddMemory` insert provenance-rich data for each contributor memory.
3. **Serve XP**: `RecordServe` adds serve XP according to `serve_xp_per_serve` and `self_serve_xp_per_serve`, updating histograms and org breadth.
4. **Domain expertise**: Domain tags, provenance, and difficulty buckets accumulate, enabling analytical queries and ranking functions.
5. **Cross-org profile**: Contributor org sets maintain normalized org coverage, with getters returning merged sets for UI highlights.

### Emission Pool and Incentives
1. **Initialization**: `SetEmissionPool` seeds total supply, daily mint, share ratios, and the starting epoch.
2. **Daily minting**: Each epoch increments total supply and records a `DailyEmission` entry.
3. **Reward assignment**: Compatibility fields (`OperatorRewardEntry`, `operator_reward/{id}`) persist minted reward shares so administrators can settle off-chain without confusing new terminology.
4. **Work scores**: `ComputeWorkScore` captures rarity multipliers, availability, retrieval volume, and computed scores per org-id pair, enabling future weighted distributions.
5. **Asymmetric gatekeeping**: `SetAsymmetricGate` stores per-org gating outcomes, letting retrieval policies reference historical compliance.
6. **Bootstrap credits**: `bootstrap/{id}` and `bootstrappool/` track early adopter credits; redemption decrements both contributor balance and the shared pool until expiry.

## Network Topology

1. **Validators and full nodes**: Cosmos validators run WeVibe Chain binaries, expose peer-to-peer ports, and optionally enable the gRPC server for direct client queries. Full nodes mirror this setup without signing blocks. Both node types configure the gRPC gateway so Hub instances and power users can reach module query services without repeating translation layers.
2. **WeVibe Hub instances**: Hub processes subscribe to Tendermint WebSocket events, call module query endpoints over gRPC, and project the results into durable caches (for pagination) and ephemeral push streams (for live UI updates). Hub also brokers notifications toward the Dashboard, such as newly approved memories, serve batch acceptance rates, and emission payout summaries.
3. **Dashboard clients**: Browser or native clients connect to Hub WebSockets for streaming metrics while issuing direct gRPC calls when they require strongly consistent reads (for example verifying treasury balances before withdrawals). Transaction submission flows through wallet connectors (Keplr, Ledger, etc.) that broadcast to WeVibe Chain RPC endpoints. Dashboard users rely on hub-curated indexes to navigate large datasets but can always drill back to canonical state thanks to the documented query endpoints.
4. **Data propagation**: Every keeper emits structured logs. Hub enriches them and forwards to analytics sinks. Ecosystem partners can replay the same flows because the wire protocol is gRPC/REST backed by protobuf definitions, ensuring deterministic decoding across languages.
5. **Security boundaries**: All sensitive transitions remain on-chain. Hub and Dashboard never sign transactions on behalf of users; they only relay prepared payloads and display results. Governance-controlled authorities (`gov` module address) remain the only actors that can mutate critical parameters across modules.

## Signed Canonical Body — Verifiable Plaintext Pathway (CO-029)

The contributor's submit-time Ed25519 signature now covers a 9-field canonical body, joined by `\n` after the domain tag `wevibe.submit_memory.v1`. Field order is alphabetical: `ciphertext_hash`, `contributor_pubkey`, `epoch_id`, `memory_type`, `org_id`, `plaintext_hash`, `salt`, `submission_hash`, `wrapped_dek_hash`. The chain re-derives every hash and verifies the signature at commit time. A consumer holding `(plaintext, salt)` can prove to the chain that the plaintext is the one the contributor signed without revealing it to the hub at submit time.

### Proto changes

`MsgApproveMemory` adds:

| Field | Tag | Type | Notes |
|---|---|---|---|
| `plaintext_hash` | 9 | bytes | `sha256(salt || plaintext_utf8)` |
| `salt` | 10 | bytes | 32 random bytes per submission |
| `ciphertext_hash` | 11 | bytes | `sha256(encrypted_blob)` |
| `contributor_sig` | 12 | bytes | Ed25519 signature over the 9-field canonical body |

`StoredMemoryCommitment` adds the same four fields at tags 15–18 plus `wrapped_dek_hash = 18` (derived at commit time from `sha256(wrapped_dek_enc)`) and `contributor_sig = 19`.

### Memory keeper at `MsgApproveMemory`

Before storing `StoredMemoryCommitment`, the keeper:

1. Computes `wrapped_dek_hash = sha256(msg.wrapped_dek_enc)` and `submission_hash = sha256(encrypted_blob || wrapped_dek_enc)`.
2. Rebuilds the 9-field canonical body via `buildSubmitMemoryCanonicalBody` (`x/memory/keeper/msg_server.go`).
3. Decodes `pending.Contributor` as hex Ed25519 pubkey (must be 32 bytes).
4. `ed25519.Verify(pubkey, canonicalBody, msg.ContributorSig)` — rejects on failure.
5. Verifies `sha256(encrypted_blob) == msg.CiphertextHash` and `sha256(blob||dek) == msg.ContentHash`.

**Partial-batch-success semantics:** Any failure above causes the keeper to return a successful response WITHOUT storing the commitment and WITHOUT consuming the pending entry. Other memories in the same batch transaction continue processing. A malformed-sig memory cannot grief a leader's batch.

After verification passes, the keeper persists `plaintext_hash`, `salt`, `ciphertext_hash`, `wrapped_dek_hash`, and `contributor_sig` onto the stored commitment (`x/memory/keeper/msg_server.go:142-154`).

### VerifyUpheldReport (reputation module)

`/wevibe/reputation/v1/verify_upheld_report/{org_id}/{content_hash}` returns the full Tier 2 evidence set, including data needed even when no report has been filed yet:

- `plaintext_hash`, `salt`, `ciphertext_hash`, `wrapped_dek_hash` (from `StoredMemoryCommitment`)
- `contributor_sig`, `contributor_pubkey` (Ed25519 verification material)
- `encrypted_blob`, `wrapped_dek_enc` (full bytes for off-chain re-hashing)
- `content_hash`, `org_id`, `epoch`, `memory_type`
- `canonical_body` — convenience field with the 9-field canonical body bytes pre-reconstructed by the query handler (verifiers can independently reconstruct from the other fields)
- If an upheld `StoredMemoryReport` exists, also: `plaintext`, `ciphertext`, `capsule`, `plaintext_oversized`, `approving_moderators`, `upholding_moderators`, `upheld_at_epoch`

The query keeper also validates `stored.wrapped_dek_hash == sha256(stored.wrapped_dek_enc)` and `stored.content_hash == sha256(stored.encrypted_blob || stored.wrapped_dek_enc)` before returning, returning `Internal` if either check fails.

### Tier 2 verification (executed by external verifiers)

Holding `(plaintext, salt)` and the response above:

1. `sha256(salt || plaintext_utf8)` must equal `plaintext_hash`.
2. `sha256(encrypted_blob)` must equal `ciphertext_hash`.
3. `sha256(wrapped_dek_enc)` must equal `wrapped_dek_hash`.
4. `sha256(encrypted_blob || wrapped_dek_enc)` must equal `content_hash`.
5. `Ed25519.verify(contributor_pubkey, canonical_body, contributor_sig)` must hold.

All five steps must pass for the memory to be considered cryptographically bound to the contributor's signed plaintext.

### R-NO-SP1-RESIDUE

No SP1 / zkVM / Groth16 primitive is present in chain code. The signed-canonical-body design is the deployed verification anchor; the prior ZK pathway never shipped.
