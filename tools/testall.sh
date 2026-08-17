#!/usr/bin/env bash

clean=${1-default}
benchmark=${2-default}

set -euo pipefail

if [ "$clean" = "c" ]; then # Clean test cache and run tests
    echo -e "running the test suite\n"
    go clean -testcache
    go test -v -race ./...
fi

if [ "$clean" = "x" ]; then # just eXecute the tests
    echo -e "running the test suite\n"
    go test -v -race ./...
fi

if [ "$benchmark" = "b" ]; then  # run Benchmarks
    echo -e "running the benchmarks\n"
    go test -bench=. -benchmem
fi
