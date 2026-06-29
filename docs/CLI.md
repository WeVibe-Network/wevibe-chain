# WeVibe Chain CLI Reference

This document provides a comprehensive reference for the `wevibed` CLI commands.

## Table of Contents

- [Installation](#installation)
- [Daemon Commands](#daemon-commands)
- [Query Commands](#query-commands)
- [Transaction Commands](#transaction-commands)
- [Key Management](#key-management)
- [Genesis Commands](#genesis-commands)
- [Governance Commands](#governance-commands)

---

## Installation

### From Source

```bash
git clone https://github.com/wevibe-network/wevibe-chain.git
cd wevibe-chain
make build
```

The binary will be at `build/wevibed`.

### From Release

Download the appropriate release for your platform from the releases page and ensure it's executable:

```bash
chmod +x wevibed
sudo mv wevibed /usr/local/bin/
```

---

## Daemon Commands

### start

Start the chain node.

```bash
wevibed start [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--home` | string | `~/.wevibed` | Node home directory |
| `--pruning` | string | `syncable` | Pruning strategy (syncable, nothing, everything) |
| `--grpc-address` | string | `0.0.0.0:9090` | gRPC server address |
| `--grpc-web.address` | string | `0.0.0.0:9091` | gRPC-web server address |
| `--api.address` | string | `0.0.0.0:1317` | REST API server address |
| `--minimum-gas-prices` | string | `0.0001uvibe` | Minimum gas price |
| `--rpc.laddr` | string | `0.0.0.0:26657` | RPC server address |

**Example:**
```bash
# Start with custom home
wevibed start --home /data/wevibed --pruning everything

# Start with custom RPC
wevibed start --rpc.laddr tcp://0.0.0.0:26657
```

### init

Initialize a new node.

```bash
wevibed init [moniker] [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--chain-id` | string | | Chain ID |
| `--home` | string | `~/.wevibed` | Node home directory |
| `--overwrite` | bool | false | Overwrite existing genesis |

**Example:**
```bash
wevibed init my-node --chain-id wevibe-local-1
```

### version

Print the daemon version.

```bash
wevibed version
```

**Example Output:**
```
WeVibe Chain v1.0.0
CometBFT v0.37.0
Cosmos SDK v0.47.0
```

### tendermint

Tendermint subcommands.

#### wevibed tendermint show-node-id

Show the node ID.

```bash
wevibed tendermint show-node-id --home <home-dir>
```

#### wevibed tendermint show-address

Show the node bech32 address.

```bash
wevibed tendermint show-address --home <home-dir>
```

#### wevibed tendermint unsafe-reset-all

Reset the blockchain and all data.

```bash
wevibed tendermint unsafe-reset-all --home <home-dir>
```

---

## Query Commands

Query commands use the pattern:
```bash
wevibed query <module> <command> [flags]
```

### Organization Module

#### wevibed query org org

Get organization details.

```bash
wevibed query org org <org_id> [flags]
```

**Example:**
```bash
wevibed query org org my-org
```

**Output:**
```json
{
  "org_id": "my-org",
  "leader": "wevibe1abc...",
  "created_at": "12345678",
  ...
}
```

#### wevibed query org members

List organization members.

```bash
wevibed query org members <org_id> [flags]
```

**Example:**
```bash
wevibed query org members my-org
```

#### wevibed query org is-member

Check if address is a member.

```bash
wevibed query org is-member <org_id> <pubkey> [flags]
```

**Example:**
```bash
wevibed query org is-member my-org wevibe1abc...
```

#### wevibed query org treasury

Get organization treasury balance.

```bash
wevibed query org treasury <org_id> [flags]
```

#### wevibed query org rep-tiers

Get reputation tiers.

```bash
wevibed query org rep-tiers <org_id> [flags]
```

#### wevibed query org config

Get organization configuration.

```bash
wevibed query org config <org_id> [flags]
```

#### wevibed query org params

Get module parameters.

```bash
wevibed query org params [flags]
```

---

### Memory Module

#### wevibed query memory memory

Get memory details.

```bash
wevibed query memory memory <org_id> <content_hash> [flags]
```

**Example:**
```bash
wevibed query memory memory my-org 0a1b2c...
```

#### wevibed query memory pending

List pending commitments.

```bash
wevibed query memory pending <org_id> [flags]
```

#### wevibed query memory count

Get approved memory count.

```bash
wevibed query memory count <org_id> [flags]
```

#### wevibed query memory merkle-root

Get epoch Merkle root.

```bash
wevibed query memory merkle-root <org_id> <epoch> [flags]
```

**Example:**
```bash
wevibed query memory merkle-root my-org 100
```

#### wevibed query memory relationships

List memory relationships.

```bash
wevibed query memory relationships <org_id> <cid> [flags]
```

#### wevibed query memory validity

Get validity metadata.

```bash
wevibed query memory validity <org_id> <cid> [flags]
```

#### wevibed query memory contest

Get contest details.

```bash
wevibed query memory contest <org_id> <contest_id> [flags]
```

#### wevibed query memory contests

List contests for a memory.

```bash
wevibed query memory contests <org_id> <cid> [flags]
```

#### wevibed query memory params

Get module parameters.

```bash
wevibed query memory params [flags]
```

---

### Serve Module

#### wevibed query serve stats

Get epoch serve statistics.

```bash
wevibed query serve stats <org_id> <epoch> [flags]
```

**Example:**
```bash
wevibed query serve stats my-org 100
```

#### wevibed query serve contributor

Get contributor serves.

```bash
wevibed query serve contributor <contributor_id> <epoch> [flags]
```

#### wevibed query serve memory-count

Get memory serve count.

```bash
wevibed query serve memory-count <org_id> <content_hash> <epoch> [flags]
```

#### wevibed query serve params

Get module parameters.

```bash
wevibed query serve params [flags]
```

---

### Attestation Module

#### wevibed query attestation get-session

Get a session attestation by org and session hash.

```bash
wevibed query attestation get-session <org_id> <session_hash> [flags]
```

**Example:**
```bash
wevibed query attestation get-session my-org 1a2b3c...
```

---

#### wevibed query attestation list-sessions

List session attestations for an org in a specific epoch.

```bash
wevibed query attestation list-sessions <org_id> <epoch> [flags]
```

**Example:**
```bash
wevibed query attestation list-sessions my-org 100
```

---

#### wevibed query attestation params

Get module parameters.

```bash
wevibed query attestation params [flags]
```

---

### Bandwidth Module

#### wevibed query bandwidth state

Get bandwidth state.

```bash
wevibed query bandwidth state <org_id> <epoch> [flags]
```

#### wevibed query bandwidth override

Get bandwidth override.

```bash
wevibed query bandwidth override <org_id> [flags]
```

#### wevibed query bandwidth remaining

Get remaining bandwidth.

```bash
wevibed query bandwidth remaining <org_id> <epoch> [flags]
```

#### wevibed query bandwidth params

Get module parameters.

```bash
wevibed query bandwidth params [flags]
```

---

### Reputation Module

#### wevibed query reputation reputation

Get contributor reputation.

```bash
wevibed query reputation reputation <developer> [flags]
```

#### wevibed query reputation xp

Get contributor XP.

```bash
wevibed query reputation xp <developer> [flags]
```

#### wevibed query reputation serve-stats

Get serve statistics.

```bash
wevibed query reputation serve-stats <developer> [flags]
```

#### wevibed query reputation org-set

Get contributor org set.

```bash
wevibed query reputation org-set <developer> [flags]
```

#### wevibed query reputation profile

Get cross-org profile.

```bash
wevibed query reputation profile <developer> [flags]
```

#### wevibed query reputation active

Check if module is active.

```bash
wevibed query reputation active [flags]
```

#### wevibed query reputation params

Get module parameters.

```bash
wevibed query reputation params [flags]
```

---

### Emissions Module

#### wevibed query emissions pool

Get emission pool.

```bash
wevibed query emissions pool [flags]
```

#### wevibed query emissions work-score

Get work score.

```bash
wevibed query emissions work-score <operator_id> <org_id> <epoch> [flags]
```

#### wevibed query emissions operator-reward

Get operator reward.

```bash
wevibed query emissions operator-reward <operator_id> [flags]
```

#### wevibed query emissions params

Get module parameters.

```bash
wevibed query emissions params [flags]
```

---

## Transaction Commands

Transaction commands use the pattern:
```bash
wevibed tx <module> <command> [args] [flags]
```

### Global Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--chain-id` | string | | Chain ID |
| `--home` | string | `~/.wevibed` | Node home |
| `--node` | string | `tcp://localhost:26657` | RPC endpoint |
| `--keyring-backend` | string | `os` | Keyring backend |
| `--broadcast-mode` | string | `sync` | Broadcast mode |
| `--fees` | string | | Fees |
| `--gas` | string | `auto` | Gas |

### Organization Module

#### wevibed tx org register-org

Register a new organization.

```bash
wevibed tx org register-org [org_id] [leader] [flags]
```

**Example:**
```bash
wevibed tx org register-org my-org wevibe1abc... \
  --from my-key \
  --chain-id wevibe-local-1 \
  --gas 200000
```

#### wevibed tx org add-member

Add a member.

```bash
wevibed tx org add-member [org_id] [pubkey] [role] [flags]
```

**Example:**
```bash
wevibed tx org add-member my-org wevibe1xyz... moderator --from my-key
```

#### wevibed tx org remove-member

Remove a member.

```bash
wevibed tx org remove-member [org_id] [pubkey] [flags]
```

#### wevibed tx org fund-treasury

Fund treasury.

```bash
wevibed tx org fund-treasury [org_id] [amount] [flags]
```

**Example:**
```bash
wevibed tx org fund-treasury my-org 1000000uvibe --from my-key
```

#### wevibed tx org withdraw-treasury

Withdraw from treasury.

```bash
wevibed tx org withdraw-treasury [org_id] [amount] [recipient] [flags]
```

#### wevibed tx org set-rep-tiers

Set reputation tiers.

```bash
wevibed tx org set-rep-tiers [org_id] [tier_json_file_or_json_string] [flags]
```

**Example JSON:**
```json
[
  {"min_reputation": "0", "max_reputation": "1000", "max_contributions_per_epoch": "100", "payout_per_serve": "1"},
  {"min_reputation": "1001", "max_reputation": "10000", "max_contributions_per_epoch": "500", "payout_per_serve": "2"}
]
```

#### wevibed tx org set-org-config

Set organization config.

```bash
wevibed tx org set-org-config [org_id] [flags] [flags]
```

**Flags:**
| Flag | Type | Description |
|------|------|-------------|
| `--serve-receipt-required` | bool | Require serve receipts |
| `--decay-rate-bps` | uint | Decay rate in basis points |
| `--contest-stake-vibe` | uint | Contest stake amount |

**Example:**
```bash
wevibed tx org set-org-config my-org \
  --serve-receipt-required=true \
  --decay-rate-bps=100 \
  --from my-key
```

#### wevibed tx org grant-trial-allowance

Grant trial allowance.

```bash
wevibed tx org grant-trial-allowance [org_id] [grantee] [daily_submissions] [trial_days] [flags]
```

---

### Memory Module

#### wevibed tx memory submit-commitment

Submit a memory commitment.

```bash
wevibed tx memory submit-commitment [org_id] [content_hash] [keywords] [contributor_id] [flags]
```

**Example:**
```bash
wevibed tx memory submit-commitment my-org 0a1b2c... "knowledge,guide" wevibe1contributor... --from my-key
```

#### wevibed tx memory approve-memory

Approve a memory.

```bash
wevibed tx memory approve-memory [org_id] [content_hash] [encrypted_blob] [flags]
```

**Example:**
```bash
wevibed tx memory approve-memory my-org 0a1b2c... "base64encodedblob..." --from my-key
```

#### wevibed tx memory reject-memory

Reject a memory.

```bash
wevibed tx memory reject-memory [org_id] [content_hash] [flags]
```

#### wevibed tx memory relate-memories

Relate two memories.

```bash
wevibed tx memory relate-memories [org_id] [source_cid] [target_cid] [relation_type] [flags]
```

**Relation Types:** `CONTRADICTS`, `REPLACES`, `DEPRECATES`, `SUPERSEDES`

**Example:**
```bash
wevibed tx memory relate-memories my-org cid1 cid2 DEPRECATES --from my-key
```

#### wevibed tx memory contest-memory

Contest a memory.

```bash
wevibed tx memory contest-memory [org_id] [memory_cid] [reason] [flags]
```

#### wevibed tx memory resolve-contest

Resolve a contest.

```bash
wevibed tx memory resolve-contest [org_id] [contest_id] [upheld] [flags]
```

**Example:**
```bash
# Uphold contest
wevibed tx memory resolve-contest my-org contest-123 true --from my-key

# Reject contest
wevibed tx memory resolve-contest my-org contest-123 false --from my-key
```

#### wevibed tx memory set-validity-bounds

Set validity bounds.

```bash
wevibed tx memory set-validity-bounds [org_id] [memory_cid] [valid_after_epoch] [valid_until_epoch] [flags]
```

#### wevibed tx memory archive-memory

Archive a memory.

```bash
wevibed tx memory archive-memory [org_id] [memory_cid] [flags]
```

#### wevibed tx memory purge-expired

Purge expired commitments.

```bash
wevibed tx memory purge-expired [org_id] [flags]
```

---

### Serve Module

#### wevibed tx serve submit-serve-batch

Submit serve receipts.

```bash
wevibed tx serve submit-serve-batch [org_id] [epoch] [serve_json_file_or_json_string] [flags]
```

**Serve JSON Format:**
```json
[
  {
    "memory_content_hash": "0a1b2c...",
    "serve_key": "serve-key-123",
    "contributor_id": "wevibe1contributor...",
    "nullifier": "nullifier123..."
  }
]
```

**Example:**
```bash
wevibed tx serve submit-serve-batch my-org 100 '[{"memory_content_hash":"0a1b2c...","serve_key":"key1","contributor_id":"wevibe1abc...","nullifier":"null1..."}]' --from my-key
```

---

### Attestation Module

#### wevibed tx attestation submit-session

Submit a session attestation.

```bash
wevibed tx attestation submit-session [org_id] [session_hash] [flags]
```

**Flags:**
| Flag | Type | Description |
|------|------|-------------|
| `--model-id` | string | Model ID (e.g. "qwen3:4b", "claude-sonnet-4-20250514") |
| `--turn-count` | uint32 | Number of turns in the session |
| `--token-count` | uint32 | Total tokens consumed |
| `--provider-type` | string | "local" or "cloud" |
| `--commitllm-receipt-hash` | string | Hex-encoded CommitLLM receipt hash (32 bytes, optional) |
| `--provider-signature-hash` | string | Hex-encoded cloud provider signature hash (32 bytes, optional) |
| `--contributor-id` | string | Contributor bech32 address |
| `--epoch` | uint64 | WeVibe epoch |

**Example:**
```bash
wevibed tx attestation submit-session my-org 1a2b3c... \
  --model-id="qwen3:4b" \
  --turn-count=15 \
  --token-count=3200 \
  --provider-type=local \
  --commitllm-receipt-hash=ab12cd... \
  --contributor-id="wevibe1abc..." \
  --epoch=100 \
  --from my-key
```

---

#### wevibed query attestation get-session

Get a session attestation by org and session hash.

```bash
wevibed query attestation get-session [org_id] [session_hash]
```

---

#### wevibed query attestation list-sessions

List session attestations for an org in a specific epoch.

```bash
wevibed query attestation list-sessions [org_id] [epoch]
```

---

#### wevibed query attestation params

Get attestation module parameters.

```bash
wevibed query attestation params
```

---

### Bandwidth Module

#### wevibed tx bandwidth set-override

Set bandwidth override.

```bash
wevibed tx bandwidth set-override [org_id] [memory_cap] [serve_cap] [flags]
```

**Example:**
```bash
wevibed tx bandwidth set-override my-org 2000 20000 --from my-key
```

---

### Reputation Module

#### wevibed tx reputation update-reputation

Update reputation.

```bash
wevibed tx reputation update-reputation [developer] [memory_cid] [difficulty] [quality] [domain_tags] [provenance] [flags]
```

**Example:**
```bash
wevibed tx reputation update-reputation wevibe1dev... cid123 5 8 "rust,cosmos" verified --from my-key
```

---

### Emissions Module

#### wevibed tx emissions mint-daily-emission

Mint daily emission (authority only).

```bash
wevibed tx emissions mint-daily-emission [epoch] [flags]
```

**Example:**
```bash
wevibed tx emissions mint-daily-emission 101 --from gov-key --authority wevibe1gov...
```

#### wevibed tx emissions distribute-operator-rewards

Distribute operator rewards.

```bash
wevibed tx emissions distribute-operator-rewards [epoch] [rewards_json] [flags]
```

**Rewards JSON:**
```json
[
  {"operator_id": "op-1", "amount": "50000"},
  {"operator_id": "op-2", "amount": "30000"}
]
```

---

## Key Management

### keys add

Add a key to the keyring.

```bash
wevibed keys add [name] [flags]
```

**Flags:**
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--keyring-backend` | string | `os` | Keyring backend |
| `--algo` | string | `secp256k1` | Key algorithm |
| `--hd-path` | string | | HD derivation path |

**Example:**
```bash
wevibed keys add my-key --keyring-backend test
```

### keys list

List all keys.

```bash
wevibed keys list [flags]
```

### keys show

Show key details.

```bash
wevibed keys show [name_or_address] [flags]
```

### keys delete

Delete a key.

```bash
wevibed keys delete [name] [flags]
```

### keys export

Export a key.

```bash
wevibed keys export [name] [flags]
```

### keys import

Import a key.

```bash
wevibed keys import [name] [keyfile] [flags]
```

---

## Genesis Commands

### genesis init-validate-genesis

Validate genesis file.

```bash
wevibed genesis validate-genesis [flags]
```

### genesis add-genesis-account

Add account to genesis.

```bash
wevibed genesis add-genesis-account [address_or_key_name] [coin denominations] [flags]
```

**Example:**
```bash
wevibed genesis add-genesis-account my-key 1000000000uvibe --home ~/.wevibed
```

### genesis gentx

Generate genesis transaction.

```bash
wevibed genesis gentx [key_name] [stake_amount] [flags]
```

**Example:**
```bash
wevibed genesis gentx validator 500000000uvibe --chain-id wevibe-local-1
```

### genesis collect-gentxs

Collect genesis transactions.

```bash
wevibed genesis collect-gentxs [flags]
```

---

## Governance Commands

### tx gov submit-proposal

Submit a governance proposal.

```bash
wevibed tx gov submit-proposal [proposal_type] [proposal_json_or_file] [flags]
```

### tx gov vote

Vote on a proposal.

```bash
wevibed tx gov vote [proposal_id] [option] [flags]
```

**Options:** `yes`, `no`, `abstain`, `no_with_veto`

### tx gov deposit

Deposit to a proposal.

```bash
wevibed tx gov deposit [proposal_id] [amount] [flags]
```

### query gov proposals

List proposals.

```bash
wevibed query gov proposals [flags]
```

### query gov proposal

Get proposal details.

```bash
wevibed query gov proposal [proposal_id] [flags]
```

### query gov deposits

Get proposal deposits.

```bash
wevibed query gov deposits [proposal_id] [flags]
```

### query gov votes

Get proposal votes.

```bash
wevibed query gov votes [proposal_id] [flags]
```

---

## Bank Commands

### tx bank send

Send tokens.

```bash
wevibed tx bank send [from_key_or_address] [to_address] [amount] [flags]
```

**Example:**
```bash
wevibed tx bank send my-key wevibe1abc... 100000uvibe --chain-id wevibe-local-1
```

### query bank balances

Get account balances.

```bash
wevibed query bank balances [address] [flags]
```

### query bank supply

Get token supply.

```bash
wevibed query bank supply [denom] [flags]
```

---

## Debug Commands

### debug addr

Decode an address.

```bash
wevibed debug addr [address]
```

### debug pubkey

Decode a public key.

```bash
wevibed debug pubkey [pubkey]
```

### debug raw-bytes

Convert raw bytes.

```bash
wevibed debug raw-bytes [bytes]
```

---

## Shell Completion

Generate shell completion scripts:

```bash
# Bash
wevibed completion bash > /etc/bash_completion.d/wevibed

# Zsh
wevibed completion zsh > "${fpath[1]}/_wevibed"

# Fish
wevibed completion fish > ~/.config/fish/completions/wevibed.fish
```
