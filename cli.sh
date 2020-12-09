#!/bin/bash
CUR_DIR=$(pwd)
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd -P)"
cd "$SCRIPT_DIR" && \
  go build \
    -o "$CUR_DIR/op" \
    -ldflags "-X main.gitCommit=$(git rev-parse HEAD) -X main.buildEpochSec=$(date +%s)" \
    "$SCRIPT_DIR/cli/op.go"
cd "$CUR_DIR" && ./op "$@"
