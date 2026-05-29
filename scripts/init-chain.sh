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
else
    sed -i 's|laddr = "tcp://127.0.0.1:26657"|laddr = "tcp://0.0.0.0:26657"|' "$HOME_DIR/config/config.toml"
fi

# Ensure gRPC is enabled (required for hub chain client).
# Use production-oriented pruning so nodes keep recent state for operational
# queries while controlling disk growth, and keep IAVL fastnode enabled for
# better query/index performance.
if [ "$(uname)" = "Darwin" ]; then
    sed -i '' 's|^enable = false.*# grpc|enable = true|' "$HOME_DIR/config/app.toml"
    sed -i '' 's|^address = "localhost:9090"|address = "0.0.0.0:9090"|' "$HOME_DIR/config/app.toml"
    sed -i '' 's|^pruning = "default"|pruning = "custom"|' "$HOME_DIR/config/app.toml"
    sed -i '' 's|^pruning-keep-recent = ".*"|pruning-keep-recent = "100"|' "$HOME_DIR/config/app.toml"
    sed -i '' 's|^pruning-interval = ".*"|pruning-interval = "10"|' "$HOME_DIR/config/app.toml"
    sed -i '' "s|^minimum-gas-prices = \".*\"|minimum-gas-prices = \"0.025${DENOM}\"|" "$HOME_DIR/config/app.toml"
else
    sed -i 's|^enable = false.*# grpc|enable = true|' "$HOME_DIR/config/app.toml"
    sed -i 's|^address = "localhost:9090"|address = "0.0.0.0:9090"|' "$HOME_DIR/config/app.toml"
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

"$WEVIBED_BINARY" genesis add-genesis-account "$VALIDATOR_ADDR" "1000000000${DENOM}" \
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

"$WEVIBED_BINARY" genesis gentx validator "500000000${DENOM}" \
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

"$WEVIBED_BINARY" genesis validate-genesis --home "$HOME_DIR" 2>&1 | tail -1

echo "=== Chain initialized successfully ==="
echo "Start with: $WEVIBED_BINARY start --home $HOME_DIR"

if [ "${1:-}" = "start" ] || [ "${1:-}" = "--start" ]; then
    shift || true
    exec "$WEVIBED_BINARY" start --home "$HOME_DIR" "$@"
fi
