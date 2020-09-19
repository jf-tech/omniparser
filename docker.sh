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

cd $(git rev-parse --show-toplevel)

function build() {
  green_printf "Before building, first cleaning... \n"
  clean || panic_fail_op
  green_printf "Building omniparser image ... \n"
  docker build -t omniparser-cli-server . || panic_fail_op
  green_printf "Launching omniparser ... \n"
  docker run -d -p 8080:8080 omniparser-cli-server || panic_fail_op
  green_printf "Showing omniparser logs ... \n"
  docker ps --filter="ancestor=omniparser-cli-server" --format="{{.ID}}" | xargs docker logs -f
}

function clean() {
  green_printf "Stopping/removing all omniparser containers  ... \n"
  docker ps --filter="ancestor=omniparser-cli-server" --format="{{.ID}}" | \
      xargs docker rm --force || true
  green_printf "Removing all omniparser images  ... \n"
  docker images --filter "reference=omniparser-cli-server:*" --format "{{.ID}}" | \
      xargs docker rmi --force || true
  green_printf "Cleaning complete.\n"
}

if [ "$1" = "build" ]; then
  CMD="build"
elif [ "$1" = "clean" ]; then
  CMD="clean"
elif [ -z "$1" ]; then
  CMD="clean"
else
  panic "Error: unknown arg '$1'" >&2
fi

if [ $CMD = 'build' ]; then
  build
else
  clean
fi
