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

## Signed Canonical Body Verification (CO-029)

`MsgApproveMemory` now carries `plaintext_hash` (=9), `salt` (=10), `ciphertext_hash` (=11), and `contributor_sig` (=12) in addition to the previously existing `wrapped_dek_enc` (=7). Before the keeper writes to `approved/{org}/{hash}` it:

1. Computes `wrapped_dek_hash = sha256(msg.WrappedDekEnc)` and `submission_hash = sha256(EncryptedBlob || WrappedDekEnc)`.
2. Reconstructs the 9-field canonical body via `buildSubmitMemoryCanonicalBody` (`x/memory/keeper/msg_server.go`).
3. Decodes the contributor pubkey from `pending.Contributor` (hex Ed25519, 32 bytes).
4. Verifies `ed25519.Verify(pubkey, canonicalBody, msg.ContributorSig)`, rejects on failure.
5. Verifies `sha256(EncryptedBlob) == msg.CiphertextHash` and `sha256(blob||dek) == msg.ContentHash`.

If any check fails, the handler returns a success response WITHOUT storing the approved memory and WITHOUT consuming the pending commitment — the pending row remains so the contributor can re-submit a corrected version. Other memories in the same batch transaction proceed normally (partial-batch-success).

`StoredMemoryCommitment` now stores `plaintext_hash`, `salt`, `ciphertext_hash`, `wrapped_dek_hash` (derived), and `contributor_sig` at proto tags 15–19. These travel through `MemoryCommitment` (Go-side mirror in `x/memory/types/keys.go`) and are persisted via `memoryToStored` / `storedToMemory`.

The canonical body format is fixed: domain tag `wevibe.submit_memory.v1`, alphabetically-sorted key/value lines `ciphertext_hash`, `contributor_pubkey`, `epoch_id`, `memory_type`, `org_id`, `plaintext_hash`, `salt`, `submission_hash`, `wrapped_dek_hash`. Field ordering is identical to the MCP and dashboard implementations; a cross-language conformance test asserts byte-exact equality in `wevibe-server/wevibe-hub/internal/verify/canonical_test.go`.