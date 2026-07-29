# WeVibe Chain Protocol Design Paper (Sprint 23)

> **⚠ HISTORICAL SNAPSHOT — SUPERSEDED (2026-07-29 recall-pivot joint amendment).** This paper predates
> the recall pivot: all decay/confidence/lifecycle-mutation mechanics described below (e.g.
> `ApplyEpochDecay`, retrieval-confidence adjustment) have LEFT consensus entirely. The chain now stores
> immutable, content-free, consumer-signed EVENTS only; standing is computed at the edge as
> `f(events, policy_version)` with the policy-version hash anchored on-chain. See `docs/TOPOLOGY.md` and
> `RECALL-PIVOT-SPEC.md` (workspace root) for current reality. Retained unedited as historical record.

Version: 3.0  
Sprint: 23  
Date: 2026-05-10

## Sprint 24 Updates

- `x/memory` authorizes moderator approvals while hub enforces quorum using the new `required_approvals` org config. Leader override remains intact.
- `x/org` exposes moderator lookups consumed by the updated memory keeper and hub API handlers.
- Integration tests gained coverage for `MsgGrantTrialAllowance`, ensuring moderators retain gas via fee grants when submitting approvals after quorum.

## Protocol Overview

WeVibe is a three-layer organizational memory network. WeVibe Chain is the Cosmos SDK appchain that anchors consensus, persistent state, and the custom modules that enforce the protocol. WeVibe Hub is the off-chain Go service that ingests encrypted knowledge from contributors, surfaces it for retrieval, and orchestrates serve receipts from serving agents. WeVibe Dashboard is the web interface administrators use to register organizations, curate memories, configure payout tiers, and monitor treasuries.

Organizations register on WeVibe Chain by burning VIBE (uvibe) via a dynamic pricing curve. Each organization maintains an on-chain treasury that they fund and withdraw through the x/org module. Memory contributors submit commitments that become payable only after an org leader approves them; approval emits ciphertext storage and indexes keywords for off-chain discovery. Serving agents batch the memories they deliver and attest to them through x/serve; the module deduplicates by chain-computed fingerprints, tracks first-serve events, and records per-epoch contributor statistics. Bandwidth limits from x/bandwidth rate-limit pending commitments and serve batches to prevent spam while still allowing overrides for trusted orgs.

At the end of every `wevibe_epoch`, the x/emissions module mints a daily emission snapshot, scans every org that requires attestations, tallies serve counts per contributor for the epoch, and debits the org’s treasury according to the configured reputation tiers. Reputation data is tracked in x/reputation: every serve updates XP, self-serve counts, and cross-org breadth, and org admins can configure tiers that gate payouts. Once the epoch hook finishes, x/memory builds Merkle roots over the epoch’s approved memories so hubs can prove inclusion to downstream verifiers. Treasury-funded payouts, serve receipt checks, and Merkle proofs are therefore all derived from the on-chain state, while Hub and Dashboard ensure real-world usability.

## Module Architecture

### x/org (Organization Registry)

**Owned state:**  
- `org/<orgID>`: stored organization metadata (`StoredOrg`) with leader, quotas, renewal height, storage quota, retrieval budget, and status.  
- `member/<orgID>/<pubkey>`: member records keyed by org and public key.  
- `treasury/<orgID>`: on-chain treasury balance stored as a decimal string.  
- `reptier/<orgID>`: serialized `StoredRepTierConfig` containing ordered reputation tiers.  
- `orgconfig/<orgID>`: organizational toggles (`ServeAttestationRequired`, `DecayRateBps`, `ContestStakeVibe`).  
- `dynprice/`: single `StoredDynamicPrice` record tracking creation count and cached price.  
- `params`: `Params` blob with burn curve and other module settings.

**Messages:**  
`MsgRegisterOrg`, `MsgAddMember`, `MsgRemoveMember`, `MsgFundTreasury`, `MsgWithdrawTreasury`, `MsgSetRepTiers`, `MsgSetOrgConfig`, `MsgUpdateParams`.

**Keeper dependencies:**  
`BankKeeper` for burns, treasury transfers, and module account balance management.

**Genesis data:**  
`InitGenesis` loads orgs, members, treasuries, dynamic burn state, rep tiers, and org configs exactly as supplied. Anything omitted falls back to defaults (e.g., no serve receipt requirement). `ExportGenesis` mirrors those collections back out.

**Notable logic:**  
Registration burns VIBE at the dynamic price computed from `BaseBurnPrice`, compounded percentage increases, and epoch-based decay. Each registration automatically grants the leader membership with the `leader` role. Treasury balances are tracked as strings to preserve arbitrary precision; helper methods parse into `math.Int` before arithmetic. The keeper can debit the treasury directly (`DebitTreasury`) for epoch payouts without routing funds through bank accounts.

### x/memory (Organizational Memory)

**Owned state:**  
- `pending/<orgID>/<contentHashHex>`: `StoredPendingCommitment` with contributor IDs, keywords, scoped epoch metadata, and submission height.  
- `approved/<orgID>/<contentHashHex>`: `StoredMemoryCommitment` containing ciphertext blobs, contributor attribution, current lifecycle state, retrieval confidence (bps), epoch, approval height, and approver id.  
- `relationship/<orgID>:<sourceCid>:<targetCid>`: `StoredMemoryRelationship` describing CONTRADICTS / REPLACES / DEPRECATES / SUPERSEDES links.  
- `contest/<orgID>:<contestID>`: `StoredMemoryContest` capturing contest metadata, stake amount, epoch, and adjudication state.  
- `validity/<orgID>:<cid>`: `StoredValidityMetadata` serializing validity windows and opaque scope tags.  
- `merkle/<orgID>/<epoch>`: `StoredEpochMerkleRoot` caching per-epoch Merkle roots and memory counts.  
- `count/<orgID>`: big-endian counter of total approved memories.  
- `params`: `Params` now covering pending limits, blob size, confidence thresholds, decay floor, initial confidence, and contest window epochs.

**Messages:**  
`MsgSubmitCommitment`, `MsgApproveMemory`, `MsgRejectMemory`, `MsgPurgeExpired`, `MsgRelateMemories`, `MsgApproveRelationship`, `MsgContestMemory`, `MsgResolveContest`, `MsgSetValidityBounds`, `MsgArchiveMemory`, `MsgUpdateParams`.

**Keeper dependencies:**  
`OrgKeeper` (org existence, leadership and configuration), `BandwidthKeeper` (submission throttling), `BankKeeper` (contest escrow via the memory module account), `ServeKeeper` (serve counts for decay calculations).

**Genesis data:**  
Pending commitments, approved memories, lifecycle confidence, relationships, validity metadata, contests, and cached Merkle roots can all be seeded at genesis. `InitGenesis` persists them verbatim; `ExportGenesis` mirrors the same collections.

**Notable logic:**  
`SubmitCommitment` enforces org existence, consumes memory bandwidth for the submission epoch, and enforces `MaxPendingPerOrg`. Approvals require the signer to be an org leader, promote entries from pending to approved, initialise retrieval confidence from params, and increment the org’s approved count. `AfterEpochEnd` now executes three lifecycle phases in order: (1) `ApplyEpochDecay` adjusts confidence using per-org decay (floored by the protocol minimum) plus serve counts supplied by x/serve, mutating lifecycle state automatically; (2) `CheckEpochExpiry` archives memories whose validity windows expired; (3) `CheckContestExpiry` auto-rejects stale contests and releases or burns escrow. Only after lifecycle maintenance does the keeper compute and store Merkle roots for the epoch. Relationship approvals apply confidence penalties or forced state transitions, and `MsgContestMemory` / `MsgResolveContest` move memories into and out of `CONTESTED` with treasury-settled economics.

### x/serve (Serve Receipts)

**Owned state:**  
- `fingerprint/<hex>`: single-byte presence marker for seen serve fingerprints.  
- `receipt/<orgID>/<epoch>/<fingerprintHex>`: serialized `StoredServeReceipt`.  
- `denial/<orgID>/<epoch>/<fingerprintHex>`: serialized `StoredDenialReceipt`.  
- `denyfingerprint/<hex>`: single-byte presence marker for seen denial fingerprints.  
- `stats/<orgID>/<epoch>`: `StoredEpochServeStats` tracking totals, unique memories, unique serve keys, and self-serve counts.  
- `contributor/<contributorID>/<epoch>`: `StoredContributorEpochServes` with serve count, self-serve count, and org set.  
- `memcount/<orgID>/<contentHashHex>/<epoch>`: Big-endian serve counts per memory per epoch.  
- `memfirst/<orgID>/<contentHashHex>/<epoch>` and `keyfirst/<orgID>/<serveKey>/<epoch>`: first-serve markers to drive uniqueness metrics.  
- `params`: `Params` controlling batch size and per-memory caps.

**Messages:**  
`MsgSubmitServeBatch`, `MsgUpdateParams`.

**Keeper dependencies:**  
`OrgKeeper` (org validation if needed), `MemoryKeeper` (approved memory lookups), `BandwidthKeeper` (serve bandwidth), `ReputationKeeper` (per-serve XP).

**Genesis data:**  
Attestations, epoch stats, and contributor serve summaries can be bootstrapped. Fingerprint and counter stores are populated implicitly when attestations are loaded.

**Notable logic:**  
`ProcessServeBatch` enforces batch limits, consumes serve bandwidth, computes the serve fingerprint as `SHA256(memory_content_hash ‖ serve_key_pubkey ‖ BigEndianUint64(epoch))`, rejects duplicate fingerprints, ensures the referenced memory is approved, and enforces `MaxServesPerMemoryPerEpoch`. It populates stats incrementally: first-serve flags update unique counts, and each accepted serve increments contributor and org stats before calling `RecordServe` in x/reputation. Self-serve detection is triggered when `serve_key == contributor_id`; these events accrue separate counters and XP deltas. Denial processing follows the same deterministic dedup model using `SHA256(CanonicalDenialBody(org_id, memory_hash, epoch, serve_key_pubkey, serve_fingerprint, nonce=nil))` and links each denial to its originating serve via `serve_fingerprint`.

### x/bandwidth (Rate Limiting)

**Owned state:**  
- `state/<orgID>/<epoch>`: `StoredBandwidthState` recording memory/serve caps and usage for the epoch.  
- `override/<orgID>`: `StoredBandwidthOverride` for org-specific caps.  
- `params`: `Params` with default caps and governance authority.

**Messages:**  
`MsgSetBandwidthOverride`, `MsgUpdateParams`.

**Keeper dependencies:**  
`OrgKeeper` (currently used for contextual org checks when needed).

**Genesis data:**  
Pre-populated bandwidth states and overrides may be supplied. Duplicate overrides are deduplicated by org ID during initialization.

**Notable logic:**  
`GetOrInitBandwidthState` lazily creates a state record, drawing capacity from overrides if present; otherwise it uses defaults from params. Memory submissions and serve batches call into `ConsumeMemoryBandwidth` and `ConsumeServeBandwidth`, returning errors if caps would be exceeded. Overrides can be deleted entirely to fall back to defaults.

### x/reputation (Contributor Reputation)

**Owned state:**  
- `active/`: single-byte flag enabling or disabling reputation tracking.  
- `stats/<developerByteString>`: `StoredReputationStats` capturing difficulty histogram, domain tags, provenance breakdown, total XP, serve counts, self-serve counts, org breadth, first seen epoch, and cumulative serve XP.  
- `memory/<developerByteString>/<memoryCID>`: `StoredAttestedMemory` for deterministic audit of memory metadata.  
- `orgset/<developerByteString>`: `StoredContributorOrgSet` enumerating org IDs a contributor has served.  
- `params`: `Params` toggling activity and XP weights.

**Messages:**  
`MsgUpdateReputation`, `MsgUpdateParams`.

**Keeper dependencies:**  
None; keeper is self-contained but invoked by x/serve.

**Genesis data:**  
Activation flag, reputation stats, and contributor org sets can be seeded. If `Active` is false, `UpdateReputation` and `RecordServe` reject until governance activates the module.

**Notable logic:**  
`RecordServe` increments serve counts, distinguishes self-serves, applies XP from params (`ServeXpPerServe` vs `SelfServeXpPerServe`), sets the first seen epoch if unset, and expands the org set to track breadth. `UpdateReputation` handles memory attestation stats but is separate from serve tallies; both feed into tier-gating by orgs.

### x/emissions (Emissions and Treasury Payouts)

**Owned state:**  
- `pool/`: `StoredEmissionPool` tracking aggregate supply minted by the module, configured daily mint, and operator/validator share ratios.  
- `emission/<epoch>`: `StoredDailyEmission` snapshot containing total emission and reward maps.  
- `opreward/<operatorID>` and `valreward/<validatorID>`: pending reward amounts keyed by ID.  
- `workscore/<operatorID>/<orgID>/<epoch>`: `StoredWorkScore` for retrieval performance metrics.  
- `gate/<operatorID>/<orgID>/<epoch>`: `StoredAsymmetricGate` gating retrieval using binary flags.  
- `bootstrap/<operatorID>` and `bootstrappool/`: bootstrap credit tracking for operators.  
- `bootstrapExpiry`: expiry epoch for bootstrap credit redemption.  
- `params`: `Params` enabling governance control.

**Messages:**  
`MsgMintDailyEmission`, `MsgDistributeOperatorRewards`, `MsgUpdateParams`.

**Keeper dependencies:**  
`ServeKeeper` (serve receipt lookup) and `OrgKeeper` (treasury balance, config, and rep tiers).

**Genesis data:**  
Emission pool, daily emission history, operator and validator rewards, bootstrap credits, work scores, asymmetric gates, and bootstrap expiry can all be set during genesis. Absent data leaves the module inert until configured via governance.

**Notable logic:**  
`MintDailyEmission` requires the emission pool to be configured and enforces strictly increasing epochs. It does not transfer newly minted coins; instead it updates the pool accounting and records the emission snapshot. `AfterEpochEnd` orchestrates treasury-funded payouts: iterate orgs, skip those not requiring attestations, gather serve receipts for the current epoch, group them by contributor, fetch the org’s rep tiers, and compute `payoutPerServe * serveCount`. The helper `getPayoutPerServeForContributor` currently passes a placeholder reputation value of zero, making tier order critical (tiers should be ordered low-to-high, and the lowest tier acts as the default). The hook breaks out once the treasury cannot cover the next payout but does not refund partial payments. Total payout statistics are logged for monitoring. The emission pool continues to record operator and validator share proportions even though the hook only debits org treasuries; this ensures compatibility with future operator reward distribution without affecting current payouts.

## Keeper Dependency Graph

Instantiation order from `app/app.go` (lines 323-329):

1. `OrgKeeper` (depends on `BankKeeper` for burns and treasury transfers).  
2. `BandwidthKeeper` (depends on `OrgKeeper`).  
3. `ReputationKeeper` (no external module dependencies).  
4. `MemoryKeeper` (depends on `OrgKeeper`, `BankKeeper`, and `BandwidthKeeper`; `ServeKeeper` is injected afterward via a setter for decay lookups).  
5. `ServeKeeper` (depends on `OrgKeeper`, `MemoryKeeper`, `BandwidthKeeper`, and `ReputationKeeper`).  
6. `EmissionsKeeper` (depends on `ServeKeeper` and `OrgKeeper`).

This order ensures that each keeper’s dependencies have been constructed before injection into the next layer, and it mirrors the data flow from org configuration through memory and serve receipt to epoch payouts.

## Epoch Lifecycle

WeVibe Chain configures a single epoch identifier `wevibe_epoch` via `scripts/init-chain.sh`. Local development defaults to 60-second epochs, and production deployments set 86,400-second epochs. The epochs module registers a multi-hook with x/emissions first and x/memory second, so `AfterEpochEnd` runs in that order every epoch.

### Emissions hook (`EmissionsKeeper.AfterEpochEnd`)

1. Mint snapshot: `MintDailyEmission` reads the emission pool, enforces that the current epoch number monotonically increases, adds `DailyMint` to `TotalSupply`, and records a `DailyEmission` entry with operator and validator shares computed from the configured percentages. As no coins move, this step updates accounting only.  
2. Enumerate organizations: `GetAllOrgs` returns all registered orgs. For each:  
   a. Fetch org config. If `ServeAttestationRequired` is false, the org is skipped.  
   b. Fetch treasury balance using `GetTreasuryBalanceInt`. Zero or negative balances trigger a skip.  
   c. Load serve receipts for the epoch via `ServeKeeper.GetServeReceipts`. If none, skip.  
   d. Load reputation tiers through `GetRepTiers`. The helper `getPayoutPerServeForContributor` iterates tiers in order, comparing the placeholder reputation (currently 0) to tier ranges. The first matching tier supplies the payout rate.  
   e. Aggregate serve counts per contributor, multiply by payout rate, and attempt to debit the org’s treasury for each contributor in turn. If the treasury cannot cover the next payout, the loop breaks and remaining contributors for that org are skipped.  
   f. Counters accumulate per-org totals, number of contributors paid, and aggregate payouts.  
3. Logging: The keeper writes a structured log with epoch, total minted emission, number of orgs scanned, how many processed, total contributors paid, and the sum of payouts. Because the minted emission pool tracks operator/validator shares, future features can distribute those amounts without rewriting epoch logic now.

### Memory hook (`MemoryKeeper.AfterEpochEnd`)

1. Discover relevant orgs: `getAllOrgsWithMemories` scans the approved memory prefix, collecting org IDs that have any approved memory.  
2. Lifecycle maintenance (runs before Merkle roots):  
   a. `ApplyEpochDecay` fetches per-org decay configuration, clamps it against `MinRetrievalDecayBps`, pulls serve counts from x/serve, adjusts `retrieval_confidence_bps`, and mutates lifecycle states (APPROVED/STABLE/DEGRADED/DORMANT) accordingly.  
   b. `CheckEpochExpiry` iterates validity metadata; any memory whose `valid_until_epoch` elapsed is archived automatically.  
   c. `CheckContestExpiry` walks pending contests, auto-rejects those outside the configured window, burns escrow for rejected contests, refunds on restoration, and transitions memories out of `CONTESTED`.  
3. Merkle computation: iterate approved memories, filter entries where `stored.Epoch == epoch`, collect content hashes, and compute a Merkle root using `types.ComputeMerkleRoot`.  
4. Persist a `StoredEpochMerkleRoot` containing org ID, epoch, Merkle root bytes, and memory count.  
5. Log structured summaries so monitoring captures lifecycle adjustments and Merkle publication counts per epoch.

These hooks leave the store prepared for hubs to fetch Merkle proofs and for dashboards to show dividend reports immediately after each epoch finalizes.

## Transaction Types

| Module | Message | Description |
| --- | --- | --- |
| org | MsgRegisterOrg | Burn VIBE at the dynamic price and create a new organization with a designated leader. |
| org | MsgAddMember | Add a member to an organization, assigning the provided role. |
| org | MsgRemoveMember | Remove a member from an organization’s roster. |
| org | MsgFundTreasury | Move funds from the signer into the organization’s treasury balance. |
| org | MsgWithdrawTreasury | Withdraw treasury funds to a recipient address signed by an org admin. |
| org | MsgSetRepTiers | Replace the organization’s reputation tiers used for payout calculations. |
| org | MsgSetOrgConfig | Update the organization’s configuration flags (currently serve receipt requirement). |
| org | MsgUpdateParams | Governance action to update module parameters such as burn curve settings. |
| memory | MsgSubmitCommitment | Submit a memory commitment with keywords and contributor ID, consuming bandwidth. |
| memory | MsgApproveMemory | Leader-only approval that promotes a commitment, seeds retrieval confidence, and records lifecycle state. |
| memory | MsgRejectMemory | Leader-only rejection that deletes a pending commitment without payout. |
| memory | MsgPurgeExpired | Remove stale pending commitments older than the configured retention window. |
| memory | MsgRelateMemories | Propose a relationship (CONTRADICTS/REPLACES/DEPRECATES/SUPERSEDES) between two approved memories. |
| memory | MsgApproveRelationship | Leader approval that activates a relationship and applies its mechanical effect (confidence penalty, contest, archive). |
| memory | MsgContestMemory | Stake-backed challenge that moves an approved memory into `CONTESTED` while escrow is held in the memory module account. |
| memory | MsgResolveContest | Leader decision that archives on uphold or releases on reject, handling refunds/burns. |
| memory | MsgSetValidityBounds | Attach validity windows and scope tags to a memory for automatic expiry management. |
| memory | MsgArchiveMemory | Leader-initiated archival of a memory without contest. |
| memory | MsgUpdateParams | Governance update for memory parameters (pending limits, blob size, confidence thresholds, decay floor, contest window). |
| serve | MsgSubmitServeBatch | Submit a batch of serve receipts for an org and epoch; returns acceptance counts. |
| serve | MsgUpdateParams | Governance update for serve parameters (batch caps, per-memory caps, etc.). |
| bandwidth | MsgSetBandwidthOverride | Set org-specific memory and serve caps for future epochs. |
| bandwidth | MsgUpdateParams | Governance update for default bandwidth caps and administrative authority. |
| reputation | MsgUpdateReputation | Ingest an attested memory for a contributor, adjusting difficulty histograms and XP. |
| reputation | MsgUpdateParams | Governance update for reputation parameters, including per-serve XP. |
| emissions | MsgMintDailyEmission | Authority-only call to mint the daily emission snapshot for a specified epoch. |
| emissions | MsgDistributeOperatorRewards | Record operator reward allocations for an epoch (no transfers performed). |
| emissions | MsgUpdateParams | Governance update for emissions parameters (governance authority, limits). |

## Economic Model

**Org registration burn:** The dynamic pricing curve starts from `BaseBurnPrice` and multiplies by `(1 + BurnPriceIncreasePercent/100)` for each org created within the tracked window. `BurnPriceDecayEpochs` decrements the creation count over time, letting demand surges cool off. Registration burns VIBE from the registrant’s account through the org module account, so chain supply decreases accordingly.

**Treasury-funded payouts:** Organizations decide when and how much to fund their treasuries using `MsgFundTreasury`. Epoch payouts never mint to contributors; instead, x/emissions debits org treasuries for each contributor based on serve counts and tiered payout rates. The hook stops when a treasury lacks sufficient funds, preventing negative balances. Because payouts derive from the org’s balance rather than the daily emission mint, orgs retain full control over their reward schedule by choosing tier rates and deposit cadence.

**Reputation tiers:** Each tier specifies `[MinReputation, MaxReputation]`, `MaxContributionsPerEpoch`, and `PayoutPerServe`. The emissions hook currently uses a placeholder reputation value of zero, effectively mapping contributors to the first tier that includes zero. Orgs should order tiers from lowest to highest reputation and ensure the lowest tier captures the base case (including zero). As x/reputation matures, orgs can leverage serve XP or other metrics to move contributors across tiers.

**Bandwidth and serve incentives:** Default per-epoch caps (`DefaultMemoryCapPerEpoch`, `DefaultServeCapPerEpoch`) throttle spam while still letting governance or org admins provide overrides for high-throughput workloads. Contributors are incentivized to diversify across orgs because x/reputation increases `OrgBreadth`, and serve XP rewards legit delivery while flagging self-serves separately. Bandwidth exhaustion prevents the same org from flooding the chain even if they have a large treasury.

**Emission pool accounting:** The emission pool tracks `DailyMint`, `OperatorShare`, and `ValidatorShare`, and increments `TotalSupply` with every `MintDailyEmission` call. Despite this minted supply, the epoch hook currently only debits org treasuries and does not distribute operator or validator rewards. Nevertheless, operator and validator share percentages remain recorded per epoch so a future upgrade can disburse those balances without rewriting historical data. Operator reward distribution can also be recorded explicitly through `MsgDistributeOperatorRewards`, updating the epoch’s emission snapshot without moving funds.

**Bootstrap and work scores:** x/emissions keeps additional state for bootstrap credits and work scores. While not yet wired into the payout hook, these records allow off-chain services or future modules to condition retrieval access or operator incentives on historical work, asymmetric gate flags, and bootstrap balances.

## Anti-Gaming

1. **Fingerprint deduplication:** x/serve computes each serve fingerprint as `SHA256(memory_content_hash ‖ serve_key_pubkey ‖ BigEndianUint64(epoch))`, marks it in `fingerprint/<hex>`, and rejects any reuse, eliminating replay and duplicate payouts. Denials are deduplicated separately under `denyfingerprint/<hex>` using the canonical denial-body hash that references `serve_fingerprint`.  
2. **Per-memory serve cap:** `MaxServesPerMemoryPerEpoch` limits repeat serving of the same content within an epoch, blocking straightforward inflation attacks.  
3. **Bandwidth caps:** Memory submissions and serve batches consume per-epoch quotas. Overrides require explicit governance or admin action, providing friction against spam.  
4. **Pending commitment limit:** `MaxPendingPerOrg` prevents an org from hoarding unreviewed commitments; contributors need leader participation to progress.  
5. **Leader-only approvals/rejections:** Only the recorded leader can approve or reject memory commitments, preventing rogue members from minting payable memories.  
6. **Self-serve detection:** `ServeKey == ContributorID` flags serves as self-serve, increments `SelfServeCount`, and applies the lower `SelfServeXpPerServe` reward, making it detectable if an org consistently serves its own content.  
7. **Reputation gating:** Although the emissions hook defaults to the lowest tier today, orgs can configure zero-payout tiers for low reputation and higher payouts for proven contributors, ready for enhanced reputation integration.  
8. **Treasury sufficiency check:** Payout loops break once treasury funds run short, ensuring contributors cannot drive the treasury negative by spamming serves.  
9. **First-serve markers:** `memfirst` and `keyfirst` keys track unique memories and serve keys per epoch, giving orgs and off-chain monitors visibility into diversify vs. repeated servicer behavior.  
10. **Contest stake escrow:** `MsgContestMemory` requires the challenger to post the org-configured stake; losing burns the escrow, winning refunds plus a reward, discouraging frivolous disputes.  
11. **Genesis validation:** All keepers validate incoming genesis state; malformed entries are skipped or rejected, preventing attackers from seeding inconsistent initial state.

## Module Account Permissions

Module account permissions defined in `app/app.go`:

- `fee_collector` (Cosmos core) and `distribution` have no special permissions.  
- `mint` retains the `Minter` role for compatibility with Cosmos defaults, though `init-chain.sh` zeros inflation.  
- `staking` bonded and not-bonded pools carry `Burner` and `Staking` permissions to support slashing and redelegation.  
- `gov` is granted `Burner` to support deposit burns where needed.  
- `memory` has the `Burner` permission so the contest keeper can escrow stake in the memory module account and burn it on rejected contests.  
- `org` retains the `Burner` permission so `MsgRegisterOrg` can burn VIBE when creating new orgs.

## Genesis State

The repository’s `genesis.json` seeds only the bare minimum for custom modules:

```
"app_state": {
  "org": {},
  "bandwidth": {},
  "memory": {},
  "serve": {},
  "emissions": {},
  "reputation": {}
}
```

This leaves all module state empty, relying on defaults at startup. Because x/emissions requires an emission pool before minting, governance must configure it post-launch or ship a non-empty genesis for production.

`scripts/init-chain.sh` orchestrates a full single-validator bootstrap for local or containerized environments:

- Initializes a validator and hub-submitter account, each funded with 1,000,000,000 uvibe.  
- Generates a staking gentx for 500,000,000 uvibe and collects it.  
- Appends the `wevibe_epoch` definition to the epochs module with a configurable `EPOCH_DURATION` environment variable (defaults to `60s`).  
- Tightens governance voting and deposit periods to 2 days for fast iteration.  
- Sets x/mint inflation parameters to zero so only x/emissions tracks supply.  
- Leaves all custom module states empty, meaning orgs, treasuries, emission pool, bandwidth overrides, and reputation data all start from scratch.

Operators targeting production should pre-populate `genesis.json` with at least one emission pool configuration (daily mint and share percentages), default org configs if desired, and any bootstrap credits before distributing the genesis file. The init script already demonstrates how to append epochs and tweak Cosmos defaults without interfering with the custom module state.
