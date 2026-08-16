#!/usr/bin/env bash 

echo -e "running the test suite\n"

clean=${1-default}
benchmark=${2-default}

set -euo pipefail

if [ $clean = "c" ]; then 
    go clean -testcache
fi

go test -v -race ./...

if [ $benchmark = "b" ]; then
    echo -e "running the benchmarks\n"
    go test -bench=. -benchmem
fi