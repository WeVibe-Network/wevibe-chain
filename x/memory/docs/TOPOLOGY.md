# Memory Module Topology (CO-011b)

## Messages Removed (CO-011b)

The following messages were removed as dead code. Their corresponding state types and helper functions were preserved where used by live code.

| Removed Message | Reason | State Types Affected |
|---|---|---|
| MsgRejectMemory | Denial is hub-side (pending → denied via moderator) | None |
| MsgPurgeExpired | Archival is automatic via decay when all keywords hit 0 (D-4.4) | None |
| MsgRelateMemories | GAP-N5: no UX surface | None (propose relationship handler removed) |
| MsgApproveRelationship | GAP-N5: no UX surface | None (approve relationship handler removed) |
| MsgSetValidityBounds | GAP-N5: no UX surface | None (set validity bounds handler removed) |
| MsgArchiveMemory | Archival is decay-driven (D-4.4), not message-driven | None (archive memory handler removed) |

## State Types Preserved (Used by Live Code)

The following state types and their helper functions are retained because they are used by live handlers (MsgReportMemory, epoch hooks, InitGenesis/ExportGenesis).

### StoredMemoryRelationship
- **Key format:** `relationship/{org}:{source}:{target}`
- **Helpers preserved:** `relationshipKey`, `relationshipPrefix`, `loadRelationship`, `saveRelationship`, `GetRelationship`, `ListRelationshipsForMemory`, `DeleteRelationship`, `relationshipToStored`, `storedToRelationship`
- **Used by:** MsgReportMemory (via ListRelationshipsForMemory in queries), InitGenesis/ExportGenesis

### StoredValidityMetadata
- **Key format:** `validity/{org}:{cid}`
- **Helpers preserved:** `validityKey`, `GetValidityMetadata`, `CheckEpochExpiry`, `IsValidInEpoch`, `decodeValidityMetadata`, `isValidityEligibleState`
- **Used by:** epoch hooks (CheckEpochExpiry called at epoch end), InitGenesis/ExportGenesis

## Active Message Handlers

- **MsgSubmitCommitment** — Submit a memory commitment for review; emits `commitment_submitted` event with `{org_id, contributor_id, block_height}` (CO-016)
- **MsgApproveMemory** — Leader approves a commitment, promoting it to approved state
- **MsgUpdateParams** — Authority-gated parameter updates
- **MsgReportMemory** — Report a memory as violating community standards

## Query Endpoints

- `memory/{org_id}/{content_hash}` — Get approved memory
- `pending/{org_id}` — List pending commitments
- `relationships/{org_id}/{cid}` — List relationships for a memory
- `validity/{org_id}/{cid}` — Get validity metadata
- `count/{org_id}` — Get approved memory count
- `merkle_root/{org_id}/{epoch}` — Get epoch Merkle root