# Org Module Topology (CO-011b)

## Messages (as of CO-011b)

### MsgSetMemberCapabilities
Leader sets per-member capability flags (`can_contribute`, `can_moderate`) while member roles remain `leader`/`member` only.

- **Signer authority:** Leader wallet of the org (signer must match org.LeaderWalletAddress)
- **ValidateBasic:** signer not empty, org_id not empty, pubkey not empty
- **State mutations:**
  1. Read StoredOrg from `org/{org_id}`, verify signer == leader wallet
  2. Read StoredMemberRecord from `member/{org_id}/{pubkey}`, verify exists
  3. Update member.CanContribute / member.CanModerate, marshal and store

### MsgTransferLeadership
Leader transfers leadership to another existing member. Old leader's role changes from "leader" to "member".

- **Signer authority:** Leader of the org (wallet-signed per D-1.3)
- **ValidateBasic:** signer not empty, org_id not empty, new_leader not empty, signer != new_leader
- **State mutations:**
  1. Read StoredOrg from `org/{org_id}`, verify signer == leader
  2. Verify org.Status == OrgStatus_ACTIVE
  3. Verify new_leader is an existing member (read memberKey(org_id, newLeader))
  4. Read old leader's StoredMemberRecord, set role = "member", write back
  5. Read new leader's StoredMemberRecord, set role = "leader", write back
  6. Update org.Leader = newLeader, marshal and store

### MsgCloseOrg
Leader permanently closes the org. Hub-side cleanup (kfrag deletion, Qdrant collection removal) is handled by ChainWatcher's processCloseOrgEvent in a future CO.

- **Signer authority:** Leader of the org (wallet-signed per D-1.3)
- **ValidateBasic:** signer not empty, org_id not empty
- **State mutations:**
  1. Read StoredOrg from `org/{org_id}`, verify signer == leader
  2. Verify org.Status == OrgStatus_ACTIVE (cannot close already-closed orgs)
  3. Set org.Status = int32(OrgStatus_CLOSED) (= 3)
  4. Marshal and store updated org

## OrgStatus Constants

| Constant | Value | Description |
|---|---|---|
| OrgStatus_ACTIVE | 0 | Org is active and accepting transactions |
| OrgStatus_DORMANT | 1 | Org is dormant (future use) |
| OrgStatus_SUSPENDED | 2 | Org is suspended (future use) |
| OrgStatus_CLOSED | 3 | Org is permanently closed (CO-011b) |

## State Types Retained

- `StoredOrg` — stored at `org/{org_id}`
- `StoredMemberRecord` — stored at `member/{org_id}/{pubkey}`
- `StoredDynamicPrice` — stored at `dynprice/`
- `StoredTreasury` — stored at `treasury/{org_id}`
- `StoredRepTierConfig` — stored at `reptier/{org_id}`
- `StoredOrgConfig` — stored at `orgconfig/{org_id}`
