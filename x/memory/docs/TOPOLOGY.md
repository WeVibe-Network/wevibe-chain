# Memory Module Topology (CO-031 Rev 2)

## Scope

This module owns commitment lifecycle state for pending/approved memories and
applies Earned Trust decay to per-keyword weights.

## Active Message Handlers

- **MsgSubmitCommitment** — stores pending commitments under
  `pending/{org}/{hash}`. Submit authorization requires
  `can_contribute` capability (`Role == "leader"` is always allowed);
  org roles are `{leader, member}` only.
- **MsgApproveMemory** — promotes pending commitment to approved state under
  `approved/{org}/{hash}` after canonical-body verification.
- **MsgUpdateParams** — governance-authority parameter updates.
- **MsgReportMemory** — stores upheld-report evidence and relationship context.

## Primary State Objects

### `StoredMemoryCommitment` (`approved/{org}/{hash}`)

Stores encrypted memory payload + decay state:

- Per-keyword state: `repeated KeywordWeight keywords` (`keyword`, `weight`,
  `serve_count`, `denial_count`).
- Lifecycle state: `state`, `last_active_epoch`, `approved_at_epoch`.
- Canonical-body verification fields: `plaintext_hash`, `salt`,
  `ciphertext_hash`, `wrapped_dek_hash`, `contributor_sig`.
- **CO-031 Rev 2 additions:**
  - `serve_count_total` (proto tag 20)
  - `denial_count_total` (proto tag 21)
  - `archived_epoch` (proto tag 22)

### Other live state

- `pending/{org}/{hash}`: `StoredPendingCommitment` review queue.
- `relationship/{org}:{source}:{target}`: `StoredMemoryRelationship`.
- `validity/{org}:{cid}`: `StoredValidityMetadata`.
- `merkle/{org}/{epoch}`: `StoredEpochMerkleRoot`.
- `count/{org}`: approved-memory counter.

## Params (Earned Trust)

`x/memory/types/params.go` defaults now use the D-4.2 Earned Trust parameter
set:

- `serve_d_bps`, `denial_d_bps`, `idle_d_bps`
- `serve_floor_bps`, `denial_floor_bps`
- `idle_protect_bps`, `idle_untrusted_bps`
- `trust_min_serves`, `trust_max_rate_bps`
- `grace_epochs`, `retrieval_threshold_bps`

Legacy flat-count params (`serve_boost_bps`, `denial_decay_bps`,
`idle_decay_rate_bps`, `max_serve_boost_per_epoch`, `bootstrap_grace_epochs`)
were removed from business logic.

## Decay Execution Path

Canonical decay function: `x/memory/keeper/lifecycle.go:applyDecay`.

Wrappers:

- `ApplyServeBoost` (event-time serve path)
- `ApplyDenialDecay` (event-time denial path)
- `ApplyIdleDecay` (epoch-end idle sweep)

Cross-module inputs from `ServeKeeper`:

- `GetMemoryServeCountForEpoch`
- `GetMemoryDenialCountForEpoch`
- `GetMatchedKeywordsForEpoch`

Behavior:

- Serve and denial operations apply only to keywords present in
  `matched_keywords` for the memory+epoch.
- Unmatched keywords on active memories still receive idle decay.
- `ApplyServeBoost` increments `KeywordWeight.serve_count` for matched keywords
  and increments `memory.serve_count_total` once per handler invocation.
- `ApplyDenialDecay` increments `KeywordWeight.denial_count` for matched
  keywords and increments `memory.denial_count_total` once per invocation.
- Archive transition uses D-4.4 `.every()` semantics: memory archives when all
  keyword weights are `<= retrieval_threshold_bps`; `archived_epoch` is set.

## Query Endpoints

- `memory/{org_id}/{content_hash}` — get approved memory.
- `pending/{org_id}` — list pending commitments.
- `relationships/{org_id}/{cid}` — list relationship edges.
- `validity/{org_id}/{cid}` — get validity metadata.
- `count/{org_id}` — approved memory count.
- `merkle_root/{org_id}/{epoch}` — per-epoch Merkle root.
