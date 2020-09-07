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

cd $SCRIPT_DIR/omniparser/schemavalidate || panic_fail_op
green_printf "go:generate in 'omniparser/schemavalidate'..."
go generate || panic_fail_op

cd $SCRIPT_DIR

echo
green_printf "go generate completed!\n"
