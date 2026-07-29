# Memory Module Topology (CO-031 Rev 2)

## Scope

This module owns commitment lifecycle state for pending/approved memories.
Standing and visibility outcomes are computed at the edge from event logs and
anchored policy, not by consensus memory state.

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

Stores encrypted memory payload + lifecycle state:

- Keyword labels: flat `repeated string keywords`.
- Lifecycle state: `state`, `last_active_epoch`, `approved_at_epoch`.
- Canonical-body verification fields: `plaintext_hash`, `salt`,
  `ciphertext_hash`, `wrapped_dek_hash`, `contributor_sig`.

### Other live state

- `pending/{org}/{hash}`: `StoredPendingCommitment` review queue.
- `relationship/{org}:{source}:{target}`: `StoredMemoryRelationship`.
- `validity/{org}:{cid}`: `StoredValidityMetadata`.
- `merkle/{org}/{epoch}`: `StoredEpochMerkleRoot`.
- `count/{org}`: approved-memory counter.

## Params

`x/memory/types/params.go` owns only storage/review bounds:

- `max_pending_per_org`
- `pending_retention_epochs`
- `max_blob_size_bytes`
- `max_keywords_per_memory`
- `contest_window_epochs`

## Query Endpoints

- `memory/{org_id}/{content_hash}` — get approved memory.
- `pending/{org_id}` — list pending commitments.
- `relationships/{org_id}/{cid}` — list relationship edges.
- `validity/{org_id}/{cid}` — get validity metadata.
- `count/{org_id}` — approved memory count.
- `merkle_root/{org_id}/{epoch}` — per-epoch Merkle root.
