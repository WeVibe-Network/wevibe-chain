#!/usr/bin/env bash
set -euo pipefail

CHAIN_ID="${CHAIN_ID:-wevibe-local-1}"
MONIKER="${MONIKER:-wevibe-validator}"
HOME_DIR="${WEVIBED_HOME:-/root/.wevibed}"
DENOM="${DENOM:-uvibe}"
KEYRING_BACKEND="test"
WEVIBED_BINARY="${WEVIBED_BINARY:-wevibed}"

if [ -f "$HOME_DIR/config/genesis.json" ]; then
    echo "Chain already initialized at $HOME_DIR"
    exec "$WEVIBED_BINARY" start --home "$HOME_DIR" "$@"
fi

echo "=== Initializing WeVibe chain ==="
echo "Chain ID:  $CHAIN_ID"
echo "Moniker:   $MONIKER"
echo "Home:      $HOME_DIR"
echo "Denom:     $DENOM"

"$WEVIBED_BINARY" init "$MONIKER" --chain-id "$CHAIN_ID" --home "$HOME_DIR" 2>&1 | tail -1

if [ "$(uname)" = "Darwin" ]; then
    sed -i '' 's|laddr = "tcp://127.0.0.1:26657"|laddr = "tcp://0.0.0.0:26657"|' "$HOME_DIR/config/config.toml"
    sed -i '' 's|cors_allowed_origins = \[\]|cors_allowed_origins = ["*"]|' "$HOME_DIR/config/config.toml"
else
    sed -i 's|laddr = "tcp://127.0.0.1:26657"|laddr = "tcp://0.0.0.0:26657"|' "$HOME_DIR/config/config.toml"
    sed -i 's|cors_allowed_origins = \[\]|cors_allowed_origins = ["*"]|' "$HOME_DIR/config/config.toml"
fi

# Ensure gRPC is enabled (required for hub chain client).
# Use production-oriented pruning so nodes keep recent state for operational
# queries while controlling disk growth, and keep IAVL fastnode enabled for
# better query/index performance.
if [ "$(uname)" = "Darwin" ]; then
    sed -i '' 's|^enable = false.*# grpc|enable = true|' "$HOME_DIR/config/app.toml"
    sed -i '' 's|^address = "localhost:9090"|address = "0.0.0.0:9090"|' "$HOME_DIR/config/app.toml"
    sed -i '' '/^\[api\]/,/^\[grpc\]/ s/^enable = false/enable = true/' "$HOME_DIR/config/app.toml"
    sed -i '' 's|^address = "tcp://localhost:1317"|address = "tcp://0.0.0.0:1317"|' "$HOME_DIR/config/app.toml"
    sed -i '' 's|^pruning = "default"|pruning = "custom"|' "$HOME_DIR/config/app.toml"
    sed -i '' 's|^pruning-keep-recent = ".*"|pruning-keep-recent = "100"|' "$HOME_DIR/config/app.toml"
    sed -i '' 's|^pruning-interval = ".*"|pruning-interval = "10"|' "$HOME_DIR/config/app.toml"
    sed -i '' "s|^minimum-gas-prices = \".*\"|minimum-gas-prices = \"0.025${DENOM}\"|" "$HOME_DIR/config/app.toml"
else
    sed -i 's|^enable = false.*# grpc|enable = true|' "$HOME_DIR/config/app.toml"
    sed -i 's|^address = "localhost:9090"|address = "0.0.0.0:9090"|' "$HOME_DIR/config/app.toml"
    sed -i '/^\[api\]/,/^\[grpc\]/ s/^enable = false/enable = true/' "$HOME_DIR/config/app.toml"
    sed -i 's|^address = "tcp://localhost:1317"|address = "tcp://0.0.0.0:1317"|' "$HOME_DIR/config/app.toml"
    sed -i 's|^pruning = "default"|pruning = "custom"|' "$HOME_DIR/config/app.toml"
    sed -i 's|^pruning-keep-recent = ".*"|pruning-keep-recent = "100"|' "$HOME_DIR/config/app.toml"
    sed -i 's|^pruning-interval = ".*"|pruning-interval = "10"|' "$HOME_DIR/config/app.toml"
    sed -i "s|^minimum-gas-prices = \".*\"|minimum-gas-prices = \"0.025${DENOM}\"|" "$HOME_DIR/config/app.toml"
fi

"$WEVIBED_BINARY" keys add validator \
    --keyring-backend "$KEYRING_BACKEND" \
    --home "$HOME_DIR" \
    --output json 2>/dev/null | jq -r '.address' > /tmp/validator_addr.txt

VALIDATOR_ADDR=$(cat /tmp/validator_addr.txt)
echo "Validator address: $VALIDATOR_ADDR"

"$WEVIBED_BINARY" genesis add-genesis-account "$VALIDATOR_ADDR" "10000000000000${DENOM}" \
    --home "$HOME_DIR" \
    --keyring-backend "$KEYRING_BACKEND"

# Hub submitter account — deterministic dev mnemonic (LOCAL DEV ONLY)
# This mnemonic MUST match the hub's ChainSubmitterMnemonic config
# Derived with HD path m/44'/118'/0'/0/0 (Cosmos standard), key name "submitter"
SUBMITTER_MNEMONIC="abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
echo "$SUBMITTER_MNEMONIC" | "$WEVIBED_BINARY" keys add hub-submitter --recover --keyring-backend test --home "$HOME_DIR" 2>/dev/null
SUBMITTER_ADDR=$("$WEVIBED_BINARY" keys show hub-submitter -a --keyring-backend test --home "$HOME_DIR")
echo "Hub submitter address: $SUBMITTER_ADDR"
"$WEVIBED_BINARY" genesis add-genesis-account "$SUBMITTER_ADDR" "1000000000${DENOM}" \
    --home "$HOME_DIR" \
    --keyring-backend "$KEYRING_BACKEND"

# Foundation account — deterministic dev mnemonic (LOCAL DEV ONLY)
# CO-041 locked allocation: 100,000,000 VIBE = 100000000000000 uvibe (10%, unlocked at genesis)
FOUNDATION_MNEMONIC="test test test test test test test test test test test junk"
echo "$FOUNDATION_MNEMONIC" | "$WEVIBED_BINARY" keys add foundation --recover --keyring-backend test --home "$HOME_DIR" 2>/dev/null
FOUNDATION_ADDR=$("$WEVIBED_BINARY" keys show foundation -a --keyring-backend test --home "$HOME_DIR")
echo "Foundation address: $FOUNDATION_ADDR"
"$WEVIBED_BINARY" genesis add-genesis-account "$FOUNDATION_ADDR" "100000000000000${DENOM}" \
    --home "$HOME_DIR" \
    --keyring-backend "$KEYRING_BACKEND"

# Faucet account — deterministic dev mnemonic (LOCAL DEV ONLY)
# Must match FAUCET_MNEMONIC in compose/.env templates.
# Allocation: 1,000,000 VIBE = 1000000000000 uvibe
FAUCET_MNEMONIC="legal winner thank year wave sausage worth useful legal winner thank yellow"
echo "$FAUCET_MNEMONIC" | "$WEVIBED_BINARY" keys add faucet --recover --keyring-backend test --home "$HOME_DIR" 2>/dev/null
FAUCET_ADDR=$("$WEVIBED_BINARY" keys show faucet -a --keyring-backend test --home "$HOME_DIR")
echo "Faucet address: $FAUCET_ADDR"
"$WEVIBED_BINARY" genesis add-genesis-account "$FAUCET_ADDR" "1000000000000${DENOM}" \
    --home "$HOME_DIR" \
    --keyring-backend "$KEYRING_BACKEND"

"$WEVIBED_BINARY" genesis gentx validator "1000000000000${DENOM}" \
    --chain-id "$CHAIN_ID" \
    --home "$HOME_DIR" \
    --keyring-backend "$KEYRING_BACKEND" \
    --keyring-dir "$HOME_DIR" \
    --yes 2>&1 | tail -1

"$WEVIBED_BINARY" genesis collect-gentxs --home "$HOME_DIR" 2>&1 | tail -1

# Configure wevibe_epoch duration.
# WEVIBE_EPOCH_DURATION_SECONDS is the single duration knob for local replay
# and deployment environments that need non-default epoch cadence.
EPOCH_DURATION_SECONDS="${WEVIBE_EPOCH_DURATION_SECONDS:-}"
if [ -n "$EPOCH_DURATION_SECONDS" ]; then
    if ! [[ "$EPOCH_DURATION_SECONDS" =~ ^[0-9]+$ ]] || [ "$EPOCH_DURATION_SECONDS" -lt 1 ]; then
        echo "invalid WEVIBE_EPOCH_DURATION_SECONDS: $EPOCH_DURATION_SECONDS" >&2
        exit 1
    fi
    EPOCH_DURATION="${EPOCH_DURATION_SECONDS}s"
else
    EPOCH_DURATION="60s"
fi

jq --arg dur "$EPOCH_DURATION" '.app_state.epochs.epochs += [{
  "identifier": "wevibe_epoch",
  "duration": $dur,
  "current_epoch": "0",
  "current_epoch_start_height": "0",
  "current_epoch_start_time": "0001-01-01T00:00:00Z",
  "epoch_counting_started": false
}]' "$HOME_DIR/config/genesis.json" > /tmp/genesis_epoch.json
mv /tmp/genesis_epoch.json "$HOME_DIR/config/genesis.json"
echo "Configured wevibe_epoch with duration=$EPOCH_DURATION"

# Configure governance params for local dev (shorter periods)
jq '.app_state.gov.params.min_deposit[0].denom = "uvibe" |
    .app_state.gov.params.min_deposit[0].amount = "10000000" |
    .app_state.gov.params.max_deposit_period = "172800s" |
    .app_state.gov.params.voting_period = "172800s"' \
    "$HOME_DIR/config/genesis.json" > /tmp/genesis_gov.json
mv /tmp/genesis_gov.json "$HOME_DIR/config/genesis.json"

# Disable x/mint inflation — x/emissions handles the supply schedule
jq '.app_state.mint.minter.inflation = "0.000000000000000000" |
    .app_state.mint.minter.annual_provisions = "0.000000000000000000" |
    .app_state.mint.params.inflation_rate_change = "0.000000000000000000" |
    .app_state.mint.params.inflation_max = "0.000000000000000000" |
    .app_state.mint.params.inflation_min = "0.000000000000000000"' \
    "$HOME_DIR/config/genesis.json" > /tmp/genesis_mint.json
mv /tmp/genesis_mint.json "$HOME_DIR/config/genesis.json"
echo "Disabled x/mint inflation (x/emissions handles supply)"

# Seed emissions + reputation genesis state.
#
# `wevibed init` builds genesis.json from app.ModuleBasics, which only contains
# the SDK modules — the custom WeVibe modules (emissions, reputation, ...) are
# not in it, so their app_state keys are absent and the SDK ModuleManager skips
# InitGenesis for any module whose genesis data is nil. Seeding these keys makes
# the data present; the modules' module.HasGenesis InitGenesis then runs.
#
# emissions: seed the full 32-year emission pool (CO-041 locked allocations).
# The emissions genesis is JSON-marshaled from a Go struct; the field names below
# match that struct exactly. The pool carries the validator 32yr pool (570M VIBE =
# 570000000000000 uvibe) and contributor 32yr pool (320M VIBE = 320000000000000
# uvibe). operator_share=80 + validator_share=20 must sum to 100 for the pool's
# Validate(). This is the single source of truth for the local genesis pool.
jq '.app_state.emissions = {
  "emission_pool": {
    "total_supply": 0,
    "daily_mint": 1000000000,
    "operator_share": 80,
    "validator_share": 20,
    "epoch": 0,
    "validator_pool_remaining_uvibe": 570000000000000,
    "contributor_pool_remaining_uvibe": 320000000000000,
    "contributor_rollover_uvibe": 0,
    "start_epoch": 0,
    "total_epochs_elapsed": 0
  }
}' "$HOME_DIR/config/genesis.json" > /tmp/genesis_emissions.json
mv /tmp/genesis_emissions.json "$HOME_DIR/config/genesis.json"
echo "Seeded app_state.emissions (full 32-year pool: validator 570M VIBE pool, contributor 320M VIBE pool)"

# reputation: activate the module at genesis (D-13.5). Active is an explicit
# genesis decision, so it must be set here rather than defaulted.
jq '.app_state.reputation = {"active": true}' \
    "$HOME_DIR/config/genesis.json" > /tmp/genesis_reputation.json
mv /tmp/genesis_reputation.json "$HOME_DIR/config/genesis.json"
echo "Seeded app_state.reputation (active=true)"

"$WEVIBED_BINARY" genesis validate-genesis --home "$HOME_DIR" 2>&1 | tail -1

echo "=== Chain initialized successfully ==="
echo "Start with: $WEVIBED_BINARY start --home $HOME_DIR"

if [ "${1:-}" = "start" ] || [ "${1:-}" = "--start" ]; then
    shift || true
    exec "$WEVIBED_BINARY" start --home "$HOME_DIR" "$@"
fi
