#!/bin/bash

function green_printf () {
  printf "\e[32m$@\e[m\n"
}

function red_printf () {
  printf "\e[31m$@\e[m\n"
}

function panic () {
  echo
  red_printf "$@"
  echo
  exit 1
}

function panic_fail_op () {
  panic "Operation failed! Exit."
}

SCRIPT_DIR=$(pwd `dirname "$0"`)

cd $SCRIPT_DIR/validation || panic_fail_op
green_printf "go:generate in 'validation'...\n"
go generate || panic_fail_op

cd $SCRIPT_DIR/
green_printf "Remove all existing test '.snapshots' directories...\n"
find . -type d | grep -e "\.snapshots$" | xargs rm -rf || panic_fail_op

green_printf "Regenerating all snapshots...\n"
go clean -testcache ./... || panic_fail_op
go test ./...

green_printf "\nVerifying snapshots generation...\n"
go clean -testcache ./... || panic_fail_op
go test ./... || panic_fail_op

cd $SCRIPT_DIR/
green_printf "All done!\n"
