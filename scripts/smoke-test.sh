#!/usr/bin/env bash
set -euo pipefail

RPC_URL="${RPC_URL:-http://localhost:26657}"  # Docker: port 26657 is mapped to host in docker-compose.yml
EXPECTED_CHAIN_ID="${CHAIN_ID:-wevibe-local-1}"
WEVIBED_BINARY="${WEVIBED_BINARY:-wevibed}"
MAX_WAIT=30
INTERVAL=2

echo "=== WeVibe Chain Smoke Test ==="
echo "RPC: $RPC_URL"
echo "Binary: $WEVIBED_BINARY"
echo "Expected chain id: $EXPECTED_CHAIN_ID"

echo -n "Waiting for RPC..."
elapsed=0
while ! curl -sf "$RPC_URL/health" > /dev/null 2>&1; do
    sleep $INTERVAL
    elapsed=$((elapsed + INTERVAL))
    if [ $elapsed -ge $MAX_WAIT ]; then
        echo " TIMEOUT (${MAX_WAIT}s)"
        echo "FAIL: Node did not become healthy"
        exit 1
    fi
    echo -n "."
done
echo " OK"

echo -n "Node status... "
STATUS=$(curl -sf "$RPC_URL/status")
CHAIN_ID=$(echo "$STATUS" | jq -r '.result.node_info.network')
HEIGHT=$(echo "$STATUS" | jq -r '.result.sync_info.latest_block_height')
echo "chain=$CHAIN_ID height=$HEIGHT"

if [ "$HEIGHT" = "0" ]; then
    echo "FAIL: Block height is 0"
    exit 1
fi

echo -n "Net info... "
NET=$(curl -sf "$RPC_URL/net_info")
LISTENING=$(echo "$NET" | jq -r '.result.listening')
echo "listening=$LISTENING"

echo -n "Genesis... "
GENESIS=$(curl -sf "$RPC_URL/genesis")
GEN_CHAIN=$(echo "$GENESIS" | jq -r '.result.genesis.chain_id')
echo "genesis_chain=$GEN_CHAIN"

if [ "$CHAIN_ID" != "$EXPECTED_CHAIN_ID" ]; then
    echo "FAIL: Unexpected chain id from status: $CHAIN_ID (expected $EXPECTED_CHAIN_ID)"
    exit 1
fi

if [ "$GEN_CHAIN" != "$EXPECTED_CHAIN_ID" ]; then
    echo "FAIL: Unexpected chain id from genesis: $GEN_CHAIN (expected $EXPECTED_CHAIN_ID)"
    exit 1
fi

echo -n "Waiting for block height > 1..."
elapsed=0
while true; do
    H=$(curl -sf "$RPC_URL/status" | jq -r '.result.sync_info.latest_block_height')
    if [ "$H" -gt 1 ] 2>/dev/null; then
        echo " height=$H OK"
        break
    fi
    sleep $INTERVAL
    elapsed=$((elapsed + INTERVAL))
    if [ $elapsed -ge $MAX_WAIT ]; then
        echo " TIMEOUT"
        echo "FAIL: Chain stuck at height $H"
        exit 1
    fi
    echo -n "."
done

echo ""
echo "=== SMOKE TEST PASSED ==="
echo "Chain ID: $CHAIN_ID"
echo "Height:   $(curl -sf "$RPC_URL/status" | jq -r '.result.sync_info.latest_block_height')"
