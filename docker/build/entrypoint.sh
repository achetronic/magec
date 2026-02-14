#!/bin/sh
set -e

STORE_PATH="/app/data/store.json"
SEEDS_DIR="/app/data/seeds"

if [ ! -f "$STORE_PATH" ] && [ -n "$MAGEC_SEED" ]; then
  SEED_FILE="$SEEDS_DIR/$MAGEC_SEED.json"
  if [ -f "$SEED_FILE" ]; then
    mkdir -p "$(dirname "$STORE_PATH")"
    cp "$SEED_FILE" "$STORE_PATH"
    echo "Store seeded from preset: $MAGEC_SEED"
  else
    echo "WARNING: Seed '$MAGEC_SEED' not found at $SEED_FILE, starting with empty store"
  fi
fi

exec ./magec "$@"
