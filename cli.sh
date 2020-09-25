#!/bin/bash
SCRIPT_DIR=$(pwd `dirname "$0"`)
go run $SCRIPT_DIR/cli/op.go "$@"
